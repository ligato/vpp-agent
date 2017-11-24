// icmp-responder is a simple example showing how to answer APR and ICMP echo
// requests through a memif interface. Package "google/gopacket" is used to decode
// and construct packets.
//
// The appropriate VPP configuration for the opposite memif is:
//   vpp$ create memif id 1 socket /tmp/icmp-responder-example slave secret secret
//   vpp$ set int state memif0/1 up
//   vpp$ set int ip address memif0/1 192.168.1.2/24
//
// To start the example, simply type:
//   root$ ./icmp-responder
//
// icmp-responder needs to be run as root so that it can access the socket
// created by VPP.
//
// Normally, the memif interface is in the master mode. Pass CLI flag "--slave"
// to create memif in the slave mode:
//   root$ ./icmp-responder --slave
//
// Don't forget to put the opposite memif into the master mode in that case.
//
// To verify the connection, run:
//   vpp$ ping 192.168.1.1
//   64 bytes from 192.168.1.1: icmp_seq=2 ttl=255 time=.6974 ms
//   64 bytes from 192.168.1.1: icmp_seq=3 ttl=255 time=.6310 ms
//   64 bytes from 192.168.1.1: icmp_seq=4 ttl=255 time=1.0350 ms
//   64 bytes from 192.168.1.1: icmp_seq=5 ttl=255 time=.5359 ms
//
//   Statistics: 5 sent, 4 received, 20% packet loss
//   vpp$ sh ip arp
//   Time           IP4       Flags      Ethernet              Interface
//   68.5648   192.168.1.1     D    aa:aa:aa:aa:aa:aa memif0/1
//
// Note: it is expected that the first ping is shown as lost. It was actually
// converted to an ARP request. This is a VPP feature common to all interface
// types.
//
// Stop the example with an interrupt signal.
package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"git.fd.io/govpp.git/extras/libmemif"
)

const (
	// Socket through which the opposite memifs will establish the connection.
	Socket = "/tmp/icmp-responder-example"

	// Secret used to authenticate the memif connection.
	Secret = "secret"

	// ConnectionID is an identifier used to match opposite memifs.
	ConnectionID = 1

	// IPAddress assigned to the memif interface.
	IPAddress = "192.168.1.1"

	// MAC address assigned to the memif interface.
	MAC = "aa:aa:aa:aa:aa:aa"

	// NumQueues is the (configured!) number of queues for both Rx & Tx.
	// The actual number agreed during connection establishment may be smaller!
	NumQueues uint8 = 3
)

// For management of go routines.
var wg sync.WaitGroup
var stopCh chan struct{}

// Parsed addresses.
var hwAddr net.HardwareAddr
var ipAddr net.IP

// ErrUnhandledPacket is thrown and printed when an unexpected packet is received.
var ErrUnhandledPacket = errors.New("received an unhandled packet")

// OnConnect is called when a memif connection gets established.
func OnConnect(memif *libmemif.Memif) (err error) {
	details, err := memif.GetDetails()
	if err != nil {
		fmt.Printf("libmemif.GetDetails() error: %v\n", err)
	}
	fmt.Printf("memif %s has been connected: %+v\n", memif.IfName, details)

	stopCh = make(chan struct{})
	// Start a separate go routine for each RX queue.
	// (memif queue is a unit of parallelism for Rx/Tx).
	// Beware: the number of queues created may be lower than what was requested
	// in MemifConfiguration (the master makes the final decision).
	// Use Memif.GetDetails to get the number of queues.
	var i uint8
	for i = 0; i < uint8(len(details.RxQueues)); i++ {
		wg.Add(1)
		go IcmpResponder(memif, i)
	}
	return nil
}

// OnDisconnect is called when a memif connection is lost.
func OnDisconnect(memif *libmemif.Memif) (err error) {
	fmt.Printf("memif %s has been disconnected\n", memif.IfName)
	// Stop all packet producers and consumers.
	close(stopCh)
	wg.Wait()
	return nil
}

// IcmpResponder answers to ICMP pings with ICMP pongs.
func IcmpResponder(memif *libmemif.Memif, queueID uint8) {
	defer wg.Done()

	// Get channel which fires every time there are packets to read on the queue.
	interruptCh, err := memif.GetQueueInterruptChan(queueID)
	if err != nil {
		// Example of libmemif error handling code:
		switch err {
		case libmemif.ErrQueueID:
			fmt.Printf("libmemif.Memif.GetQueueInterruptChan() complains about invalid queue id!?")
		// Here you would put all the errors that need to be handled individually...
		default:
			fmt.Printf("libmemif.Memif.GetQueueInterruptChan() error: %v\n", err)
		}
		return
	}

	for {
		select {
		case <-interruptCh:
			// Read all packets from the queue but at most 10 at once.
			// Since there is only one interrupt signal sent for an entire burst
			// of packets, an interrupt handling routine should repeatedly call
			// RxBurst() until the function returns an empty slice of packets.
			// This way it is ensured that there are no packets left
			// on the queue unread when the interrupt signal is cleared.
			for {
				packets, err := memif.RxBurst(queueID, 10)
				if err != nil {
					fmt.Printf("libmemif.Memif.RxBurst() error: %v\n", err)
					// Skip this burst, continue with the next one 3secs later...
					break
				}
				if len(packets) == 0 {
					// No more packets to read until the next interrupt.
					break
				}
				// Generate response for each supported request.
				responses := []libmemif.RawPacketData{}
				for _, packet := range packets {
					fmt.Println("Received new packet:")
					DumpPacket(packet)
					response, err := GeneratePacketResponse(packet)
					if err == nil {
						fmt.Println("Sending response:")
						DumpPacket(response)
						responses = append(responses, response)
					} else {
						fmt.Printf("Failed to generate response: %v\n", err)
					}
				}
				// Send pongs / ARP responses. We may not be able to do it in one
				// burst if the ring is (almost) full or the internal buffer cannot
				// contain it.
				sent := 0
				for {
					count, err := memif.TxBurst(queueID, responses[sent:])
					if err != nil {
						fmt.Printf("libmemif.Memif.TxBurst() error: %v\n", err)
						break
					} else {
						fmt.Printf("libmemif.Memif.TxBurst() has sent %d packets.\n", count)
						sent += int(count)
						if sent == len(responses) {
							break
						}
					}
				}
			}
		case <-stopCh:
			return
		}
	}
}

// DumpPacket prints a human-readable description of the packet.
func DumpPacket(packetData libmemif.RawPacketData) {
	packet := gopacket.NewPacket(packetData, layers.LayerTypeEthernet, gopacket.Default)
	fmt.Println(packet.Dump())
}

// GeneratePacketResponse returns an appropriate answer to an ARP request
// or an ICMP echo request.
func GeneratePacketResponse(packetData libmemif.RawPacketData) (response libmemif.RawPacketData, err error) {
	packet := gopacket.NewPacket(packetData, layers.LayerTypeEthernet, gopacket.Default)

	ethLayer := packet.Layer(layers.LayerTypeEthernet)
	if ethLayer == nil {
		fmt.Println("Missing ETH layer.")
		return nil, ErrUnhandledPacket
	}
	eth, _ := ethLayer.(*layers.Ethernet)

	if eth.EthernetType == layers.EthernetTypeARP {
		// Handle ARP request.
		arpLayer := packet.Layer(layers.LayerTypeARP)
		if arpLayer == nil {
			fmt.Println("Missing ARP layer.")
			return nil, ErrUnhandledPacket
		}
		arp, _ := arpLayer.(*layers.ARP)
		if arp.Operation != layers.ARPRequest {
			fmt.Println("Not ARP request.")
			return nil, ErrUnhandledPacket
		}
		fmt.Println("Received an ARP request.")

		// Build packet layers.
		ethResp := layers.Ethernet{
			SrcMAC:       hwAddr,
			DstMAC:       eth.SrcMAC,
			EthernetType: layers.EthernetTypeARP,
		}
		arpResp := layers.ARP{
			AddrType:          layers.LinkTypeEthernet,
			Protocol:          layers.EthernetTypeIPv4,
			HwAddressSize:     6,
			ProtAddressSize:   4,
			Operation:         layers.ARPReply,
			SourceHwAddress:   []byte(hwAddr),
			SourceProtAddress: []byte(ipAddr),
			DstHwAddress:      arp.SourceHwAddress,
			DstProtAddress:    arp.SourceProtAddress,
		}
		// Set up buffer and options for serialization.
		buf := gopacket.NewSerializeBuffer()
		opts := gopacket.SerializeOptions{
			FixLengths:       true,
			ComputeChecksums: true,
		}
		err := gopacket.SerializeLayers(buf, opts, &ethResp, &arpResp)
		if err != nil {
			fmt.Println("SerializeLayers error: ", err)
		}
		return buf.Bytes(), nil
	}

	if eth.EthernetType == layers.EthernetTypeIPv4 {
		// Respond to ICMP request.
		ipLayer := packet.Layer(layers.LayerTypeIPv4)
		if ipLayer == nil {
			fmt.Println("Missing IPv4 layer.")
			return nil, ErrUnhandledPacket
		}
		ipv4, _ := ipLayer.(*layers.IPv4)
		if ipv4.Protocol != layers.IPProtocolICMPv4 {
			fmt.Println("Not ICMPv4 protocol.")
			return nil, ErrUnhandledPacket
		}
		icmpLayer := packet.Layer(layers.LayerTypeICMPv4)
		if icmpLayer == nil {
			fmt.Println("Missing ICMPv4 layer.")
			return nil, ErrUnhandledPacket
		}
		icmp, _ := icmpLayer.(*layers.ICMPv4)
		if icmp.TypeCode.Type() != layers.ICMPv4TypeEchoRequest {
			fmt.Println("Not ICMPv4 echo request.")
			return nil, ErrUnhandledPacket
		}
		fmt.Println("Received an ICMPv4 echo request.")

		// Build packet layers.
		ethResp := layers.Ethernet{
			SrcMAC:       hwAddr,
			DstMAC:       eth.SrcMAC,
			EthernetType: layers.EthernetTypeIPv4,
		}
		ipv4Resp := layers.IPv4{
			Version:    4,
			IHL:        5,
			TOS:        0,
			Id:         0,
			Flags:      0,
			FragOffset: 0,
			TTL:        255,
			Protocol:   layers.IPProtocolICMPv4,
			SrcIP:      ipAddr,
			DstIP:      ipv4.SrcIP,
		}
		icmpResp := layers.ICMPv4{
			TypeCode: layers.CreateICMPv4TypeCode(layers.ICMPv4TypeEchoReply, 0),
			Id:       icmp.Id,
			Seq:      icmp.Seq,
		}

		// Set up buffer and options for serialization.
		buf := gopacket.NewSerializeBuffer()
		opts := gopacket.SerializeOptions{
			FixLengths:       true,
			ComputeChecksums: true,
		}
		err := gopacket.SerializeLayers(buf, opts, &ethResp, &ipv4Resp, &icmpResp,
			gopacket.Payload(icmp.Payload))
		if err != nil {
			fmt.Println("SerializeLayers error: ", err)
		}
		return buf.Bytes(), nil
	}

	return nil, ErrUnhandledPacket
}

func main() {
	var err error
	fmt.Println("Starting 'icmp-responder' example...")

	hwAddr, err = net.ParseMAC(MAC)
	if err != nil {
		fmt.Println("Failed to parse the MAC address: %v", err)
		return
	}

	ip := net.ParseIP(IPAddress)
	if ip != nil {
		ipAddr = ip.To4()
	}
	if ipAddr == nil {
		fmt.Println("Failed to parse the IP address: %v", err)
		return
	}

	// If run with the "--slave" option, create memif in the slave mode.
	var isMaster = true
	var appSuffix string
	if len(os.Args) > 1 && (os.Args[1] == "--slave" || os.Args[1] == "-slave") {
		isMaster = false
		appSuffix = "-slave"
	}

	// Initialize libmemif first.
	appName := "ICMP-Responder" + appSuffix
	fmt.Println("Initializing libmemif as ", appName)
	err = libmemif.Init(appName)
	if err != nil {
		fmt.Printf("libmemif.Init() error: %v\n", err)
		return
	}
	// Schedule automatic cleanup.
	defer libmemif.Cleanup()

	// Prepare callbacks to use with the memif.
	// The same callbacks could be used with multiple memifs.
	// The first input argument (*libmemif.Memif) can be used to tell which
	// memif the callback was triggered for.
	memifCallbacks := &libmemif.MemifCallbacks{
		OnConnect:    OnConnect,
		OnDisconnect: OnDisconnect,
	}

	// Prepare memif1 configuration.
	memifConfig := &libmemif.MemifConfig{
		MemifMeta: libmemif.MemifMeta{
			IfName:         "memif1",
			ConnID:         ConnectionID,
			SocketFilename: Socket,
			Secret:         Secret,
			IsMaster:       isMaster,
			Mode:           libmemif.IfModeEthernet,
		},
		MemifShmSpecs: libmemif.MemifShmSpecs{
			NumRxQueues:  NumQueues,
			NumTxQueues:  NumQueues,
			BufferSize:   2048,
			Log2RingSize: 10,
		},
	}

	fmt.Printf("Callbacks: %+v\n", memifCallbacks)
	fmt.Printf("Config: %+v\n", memifConfig)

	// Create memif1 interface.
	memif, err := libmemif.CreateInterface(memifConfig, memifCallbacks)
	if err != nil {
		fmt.Printf("libmemif.CreateInterface() error: %v\n", err)
		return
	}
	// Schedule automatic cleanup of the interface.
	defer memif.Close()

	// Wait until an interrupt signal is received.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
}
