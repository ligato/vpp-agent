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

// MemoryInfo contains values returned from 'show memory'
type MemoryInfo struct {
	Threads []MemoryThread `json:"threads"`
}

// MemoryThread represents single thread memory counters
type MemoryThread struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Objects   uint64 `json:"objects"`
	Used      uint64 `json:"used"`
	Total     uint64 `json:"total"`
	Free      uint64 `json:"free"`
	Reclaimed uint64 `json:"reclaimed"`
	Overhead  uint64 `json:"overhead"`
	Capacity  uint64 `json:"capacity"`
}

// GetNodeCounters retrieves node counters info
func GetMemory(vppChan *govppapi.Channel) (*MemoryInfo, error) {
	const cmd = "show memory"
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

	var threads []MemoryThread
	var thread *MemoryThread

	for _, line := range strings.Split(string(data), "\n") {
		if thread != nil {
			for _, part := range strings.Split(line, ",") {
				fields := strings.Fields(strings.TrimSpace(part))
				if len(fields) > 1 {
					switch fields[1] {
					case "objects":
						thread.Objects = strToUint64(fields[0])
					case "of":
						thread.Used = strToUint64(fields[0])
						thread.Total = strToUint64(fields[2])
					case "free":
						thread.Free = strToUint64(fields[0])
					case "reclaimed":
						thread.Reclaimed = strToUint64(fields[0])
					case "overhead":
						thread.Overhead = strToUint64(fields[0])
					case "capacity":
						thread.Capacity = strToUint64(fields[0])
					}
				}
			}
			threads = append(threads, *thread)
			thread = nil
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 3 {
			if fields[0] == "Thread" {
				id, err := strconv.ParseUint(fields[1], 10, 64)
				if err != nil {
					return nil, err
				}
				thread = &MemoryThread{
					ID:   uint(id),
					Name: strings.SplitN(fields[2], string(0x00), 2)[0],
				}
				continue
			}
		}
	}

	info := &MemoryInfo{
		Threads: threads,
	}

	return info, nil
}

// NodeCounterInfo contains values returned from 'show node counters'
type NodeCounterInfo struct {
	Counters []NodeCounter `json:"counters"`
}

// NodeCounter represents single node counter
type NodeCounter struct {
	Count  uint64 `json:"count"`
	Node   string `json:"node"`
	Reason string `json:"reason"`
}

// GetNodeCounters retrieves node counters info
func GetNodeCounters(vppChan *govppapi.Channel) (*NodeCounterInfo, error) {
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

	var counters []NodeCounter

	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 3 {
			if fields[0] == "Count" {
				counters = []NodeCounter{}
				continue
			}
			if counters != nil {
				count, err := strconv.ParseUint(fields[0], 10, 64)
				if err != nil {
					return nil, err
				}
				counters = append(counters, NodeCounter{
					Count:  count,
					Node:   fields[1],
					Reason: fields[2],
				})
			}
		}
	}

	info := &NodeCounterInfo{
		Counters: counters,
	}

	return info, nil
}

// RuntimeInfo contains values returned from 'show runtime'
type RuntimeInfo struct {
	Items []RuntimeItem `json:"items"`
}

// NodeCounter represents single runtime item
type RuntimeItem struct {
	Name        string  `json:"name"`
	State       string  `json:"state"`
	Calls       uint64  `json:"calls"`
	Vectors     uint64  `json:"vectors"`
	Suspends    uint64  `json:"suspends"`
	Clocks      float64 `json:"clocks"`
	VectorsCall float64 `json:"vectors_call"`
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
				calls, err := strconv.ParseUint(fields[2], 10, 64)
				if err != nil {
					return nil, err
				}
				vectors, err := strconv.ParseUint(fields[3], 10, 64)
				if err != nil {
					return nil, err
				}
				suspends, err := strconv.ParseUint(fields[4], 10, 64)
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
					Calls:       calls,
					Vectors:     vectors,
					Suspends:    suspends,
					Clocks:      clocks,
					VectorsCall: vectorsCall,
				})
			}
		}
	}

	info := &RuntimeInfo{
		Items: items,
	}

	return info, nil
}

func strToUint64(s string) uint64 {
	s = strings.Replace(s, "k", "000", 1)
	num, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}
	return num
}

func cleanBytes(b []byte) []byte {
	return bytes.SplitN(b, []byte{0x00}, 2)[0]
}
