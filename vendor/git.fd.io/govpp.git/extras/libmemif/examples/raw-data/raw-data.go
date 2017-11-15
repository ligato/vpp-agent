// raw-data is a basic example showing how to create a memif interface, handle
// events through callbacks and perform Rx/Tx of raw data. Before handling
// an actual packets it is important to understand the skeleton of libmemif-based
// applications.
//
// Since VPP expects proper packet data, it is not very useful to connect
// raw-data example with VPP, even though it will work, since all the received
// data will get dropped on the VPP side.
//
// To create a connection of two raw-data instances, run two processes
// concurrently:
//  - master memif:
//     $ ./raw-data
//  - slave memif:
//     $ ./raw-data --slave
//
// Every 3 seconds both sides send 3 raw-data packets to the opposite end through
// each queue. The received packets are printed to stdout.
//
// Stop an instance of raw-data with an interrupt signal.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	"git.fd.io/govpp.git/extras/libmemif"
)

const (
	// Socket through which the opposite memifs will establish the connection.
	Socket = "/tmp/raw-data-example"

	// Secret used to authenticate the memif connection.
	Secret = "secret"

	// ConnectionID is an identifier used to match opposite memifs.
	ConnectionID = 1

	// NumQueues is the (configured!) number of queues for both Rx & Tx.
	// The actual number agreed during connection establishment may be smaller!
	NumQueues uint8 = 3
)

// For management of go routines.
var wg sync.WaitGroup
var stopCh chan struct{}

// OnConnect is called when a memif connection gets established.
func OnConnect(memif *libmemif.Memif) (err error) {
	details, err := memif.GetDetails()
	if err != nil {
		fmt.Printf("libmemif.GetDetails() error: %v\n", err)
	}
	fmt.Printf("memif %s has been connected: %+v\n", memif.IfName, details)

	stopCh = make(chan struct{})
	// Start a separate go routine for each queue.
	// (memif queue is a unit of parallelism for Rx/Tx)
	// Beware: the number of queues created may be lower than what was requested
	// in MemifConfiguration (the master makes the final decision).
	// Use Memif.GetDetails to get the number of queues.
	var i uint8
	for i = 0; i < uint8(len(details.RxQueues)); i++ {
		wg.Add(1)
		go ReadAndPrintPackets(memif, i)
	}
	for i = 0; i < uint8(len(details.TxQueues)); i++ {
		wg.Add(1)
		go SendPackets(memif, i)
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

// ReadAndPrintPackets keeps receiving raw packet data from a selected queue
// and prints them to stdout.
func ReadAndPrintPackets(memif *libmemif.Memif, queueID uint8) {
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
				} else {
					if len(packets) == 0 {
						// No more packets to read until the next interrupt.
						break
					}
					for _, packet := range packets {
						fmt.Printf("Received packet queue=%d: %v\n", queueID, string(packet[:]))
					}
				}
			}
		case <-stopCh:
			return
		}
	}
}

// SendPackets keeps sending bursts of 3 raw-data packets every 3 seconds into
// the selected queue.
func SendPackets(memif *libmemif.Memif, queueID uint8) {
	defer wg.Done()

	counter := 0
	for {
		select {
		case <-time.After(3 * time.Second):
			counter++
			// Prepare fake packets.
			packets := []libmemif.RawPacketData{
				libmemif.RawPacketData("Packet #1 in burst number " + strconv.Itoa(counter)),
				libmemif.RawPacketData("Packet #2 in burst number " + strconv.Itoa(counter)),
				libmemif.RawPacketData("Packet #3 in burst number " + strconv.Itoa(counter)),
			}
			// Send the packets. We may not be able to do it in one burst if the ring
			// is (almost) full or the internal buffer cannot contain it.
			sent := 0
			for {
				count, err := memif.TxBurst(queueID, packets[sent:])
				if err != nil {
					fmt.Printf("libmemif.Memif.TxBurst() error: %v\n", err)
					break
				} else {
					fmt.Printf("libmemif.Memif.TxBurst() has sent %d packets.\n", count)
					sent += int(count)
					if sent == len(packets) {
						break
					}
				}
			}
		case <-stopCh:
			return
		}
	}
}

func main() {
	fmt.Println("Starting 'raw-data' example...")

	// If run with the "--slave" option, create memif in the slave mode.
	var isMaster = true
	var appSuffix string
	if len(os.Args) > 1 && (os.Args[1] == "--slave" || os.Args[1] == "-slave") {
		isMaster = false
		appSuffix = "-slave"
	}

	// Initialize libmemif first.
	appName := "Raw-Data" + appSuffix
	fmt.Println("Initializing libmemif as ", appName)
	err := libmemif.Init(appName)
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
