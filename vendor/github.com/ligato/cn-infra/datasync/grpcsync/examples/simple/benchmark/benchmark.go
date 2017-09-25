package benchmark

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/prometheus/common/log"
	"gitlab.cisco.com/ctao/vnf-agent/agent/cmd/vpp-agent-ctl/commands/rpccmd"
	"gitlab.cisco.com/ctao/vnf-agent/plugins/vpp/defaultplugins/ifplugin/testing"
	"gitlab.cisco.com/ctao/vnf-agent/plugins/vpp/defaultplugins/l2plugin/model/l2"
	testing2 "gitlab.cisco.com/ctao/vnf-agent/plugins/vpp/defaultplugins/l2plugin/testing"
	//"gitlab.cisco.com/ctao/vnf-agent/plugins/vpp/defaultplugins/ifplugin/model/interfaces"
	"gitlab.cisco.com/ctao/vnf-agent/plugins/vpp/defaultplugins/ifplugin/model/interfaces"
)

func Run() error {
	sender := &rpccmd.GrpcCommands{}
	err := sender.Init()
	if err != nil {
		return err
	}
	defer sender.Close()

	startSending, endSending, err := sendData(sender)
	if err != nil {
		return err
	}

	sendingDuration := endSending.Sub(startSending)

	fmt.Println("Sending duration: ", userFriendlyDuration(sendingDuration))

	return nil
}

func userFriendlyDuration(duration time.Duration) string {
	left := duration
	ret := ""

	if left > time.Second {
		sec := time.Duration(left / time.Second)
		left -= sec * time.Second

		ret += fmt.Sprintf("%d s ", sec)
	}

	if left > time.Millisecond {
		ms := time.Duration(left / time.Millisecond)
		left -= ms * time.Millisecond

		ret += fmt.Sprintf("%d ms ", ms)
	}

	if left > time.Microsecond {
		mics := time.Duration(left / time.Microsecond)
		left -= mics * time.Microsecond

		ret += fmt.Sprintf("%d micrs ", mics)
	}

	if left > time.Nanosecond {
		ns := time.Duration(left / time.Nanosecond)
		left -= ns * time.Nanosecond

		ret += fmt.Sprintf("%d ns ", ns)
	}

	return ret
}

var nilTime = time.Unix(0, 0)

func sendData(sender *rpccmd.GrpcCommands) (startTime time.Time, endTime time.Time, err error) {
	reqs := map[string] /*key*/ []byte{}

	fmt.Println("Preparing data")
	var if0Name, if1Name string
	for i := 0; i < 2; i++ {
		ifName := fmt.Sprintf("bench_if%d", i)
		if i == 0 {
			if0Name = ifName
		} else if i == 1 {
			if1Name = ifName
		}

		ipAddr := fmt.Sprintf("10.0.0.%d", i)
		//iface := testing.LoopbackBuilder(if1Name, ipAddr)
		iface := testing.MemifBuilder(ifName, ipAddr, true, uint32(i))
		if req, err := json.Marshal(iface); err != nil {
			return nilTime, nilTime, err
		} else {
			reqs[interfaces.InterfaceKey(ifName)] = req
		}
		log.Debug("iface", i, " ", iface)
	}
	log.Debug("if0Name ", if0Name)
	log.Debug("if1Name ", if1Name)
	var bd0Name string
	for i := 0; i < 1; i++ {
		bdName := fmt.Sprintf("bench_bd%d", i)
		if i == 0 {
			bd0Name = bdName
		}
		bd := testing2.SimpleBridgeDomain2XIfaceBuilder(bdName,
			if0Name, if1Name, false, false)
		log.Debug("bd", i, " ", bd)

		if req, err := json.Marshal(bd); err != nil {
			return nilTime, nilTime, err
		} else {
			reqs[l2.BridgeDomainKey(bdName)] = req
		}
	}
	for i := 0; i < 100; i++ {
		it := fmt.Sprintf("%d", i)
		mac := ""

		a := false
		b := len(it) - 1
		t := 0
		//fill with digits
		for ; b >= 0 && t < 12; b-- {
			if a && (b < 11 || b > 0) {
				mac = ":" + string(it[b]) + mac
			} else {
				mac = string(it[b]) + mac
			}
			a = !a
			t++
		}
		//padd with zeros
		for ; t < 12; t++ {
			if a && t < 12-1 {
				mac = ":0" + mac
			} else {
				mac = "0" + mac
			}
			a = !a
		}

		l2fib := l2.FibTableEntries_FibTableEntry{
			PhysAddress:             mac,
			BridgeDomain:            bd0Name,
			OutgoingInterface:       if1Name,
			StaticConfig:            true,
			BridgedVirtualInterface: false,
		}

		if req, err := json.Marshal(l2fib); err != nil {
			return nilTime, nilTime, err
		} else {
			reqs[l2.FibKey(mac)] = req
		}
	}
	fmt.Println("Preparing data - finished successfuly")

	fmt.Println("Sending data")
	startTime = time.Now()
	for key, val := range reqs {
		if val == nil {
			return nilTime, nilTime, errors.New("FUCK")
		}
		sender.Put(key, val)
	}
	endTime = time.Now()
	fmt.Println("Sending data - finished successfuly")

	return startTime, endTime, nil
}
