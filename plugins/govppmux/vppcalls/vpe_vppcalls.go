package vppcalls

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/vpe"
)

// VersionInfo contains values returned from ShowVersion
type VersionInfo struct {
	Program        string
	Version        string
	BuildDate      string
	BuildDirectory string
}

// GetVersionInfo retrieves version info
func GetVersionInfo(vppChan *govppapi.Channel) (*VersionInfo, error) {
	req := &vpe.ShowVersion{}
	reply := &vpe.ShowVersionReply{}

	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return nil, err
	} else if reply.Retval != 0 {
		return nil, fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	info := &VersionInfo{
		Program:        string(cleanBytes(reply.Program)),
		Version:        string(cleanBytes(reply.Version)),
		BuildDate:      string(cleanBytes(reply.BuildDate)),
		BuildDirectory: string(cleanBytes(reply.BuildDirectory)),
	}

	return info, nil
}

type NodeCounter struct {
}

// GetNodeCounters retrieves node counters info
func GetNodeCounters(vppChan *govppapi.Channel) (*NodeCounter, error) {
	const cmd = "show node counters"
	req := &vpe.CliInband{
		Cmd:    []byte(cmd),
		Length: uint32(len(cmd)),
	}
	reply := &vpe.CliInbandReply{}

	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return nil, err
	} else if reply.Retval != 0 {
		return nil, fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	data := reply.Reply[:reply.Length]
	fmt.Printf("%q\n", string(data))
	fmt.Printf("%v\n", strings.Fields(string(data)))

	info := &NodeCounter{}

	return info, nil
}

type RuntimeInfo struct {
	Items []RuntimeItem
}

type RuntimeItem struct {
	Name        string
	State       string
	Calls       uint
	Vendors     uint
	Suspends    uint
	Clocks      float64
	VectorsCall float64
}

// GetNodeCounters retrieves node counters info
func GetRuntimeInfo(vppChan *govppapi.Channel) (*RuntimeInfo, error) {
	const cmd = "show runtime"
	req := &vpe.CliInband{
		Cmd:    []byte(cmd),
		Length: uint32(len(cmd)),
	}
	reply := &vpe.CliInbandReply{}

	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return nil, err
	} else if reply.Retval != 0 {
		return nil, fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	data := reply.Reply[:reply.Length]
	fmt.Printf("%q\n", string(data))
	fmt.Printf("%v\n", strings.Fields(string(data)))

	var items []RuntimeItem

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.Replace(line, "event wait", "event-wait", -1)
		line = strings.Replace(line, "any wait", "any-wait", -1)
		fields := strings.Fields(line)
		if len(fields) == 7 {
			if fields[0] == "Name" {
				items = []RuntimeItem{}
				continue
			}
			if items != nil {
				calls, err := strconv.ParseUint(fields[2], 10, 32)
				if err != nil {
					return nil, err
				}
				vendors, err := strconv.ParseUint(fields[3], 10, 32)
				if err != nil {
					return nil, err
				}
				suspends, err := strconv.ParseUint(fields[4], 10, 32)
				if err != nil {
					return nil, err
				}
				clocks, err := strconv.ParseFloat(fields[5], 10)
				if err != nil {
					return nil, err
				}
				vectorsCall, err := strconv.ParseFloat(fields[6], 10)
				if err != nil {
					return nil, err
				}
				items = append(items, RuntimeItem{
					Name:        fields[0],
					State:       fields[1],
					Calls:       uint(calls),
					Vendors:     uint(vendors),
					Suspends:    uint(suspends),
					Clocks:      clocks,
					VectorsCall: vectorsCall,
				})
			}
		} else {
			fmt.Printf("fields: %+v\n", fields)
		}
	}

	info := &RuntimeInfo{
		Items: items,
	}

	return info, nil
}

func cleanBytes(b []byte) []byte {
	return bytes.SplitN(b, []byte{0x00}, 2)[0]
}
