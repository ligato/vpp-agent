// Copyright (c) 2018 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build !windows,!darwin

package vppapiclient

/*
#cgo CFLAGS: -DPNG_DEBUG=1
#cgo LDFLAGS: -lvppapiclient

#include "stat_client_wrapper.h"
*/
import "C"

import (
	"fmt"
	"os"
	"unsafe"

	"git.fd.io/govpp.git/adapter"
)

// global VPP stats API client, library vppapiclient only supports
// single connection at a time
var globalStatClient *statClient

// statClient is the default implementation of StatsAPI.
type statClient struct {
	socketName string
}

// NewStatClient returns new VPP stats API client.
func NewStatClient(socketName string) adapter.StatsAPI {
	return &statClient{
		socketName: socketName,
	}
}

func (c *statClient) Connect() error {
	if globalStatClient != nil {
		return fmt.Errorf("already connected to stats API, disconnect first")
	}

	var sockName string
	if c.socketName == "" {
		sockName = adapter.DefaultStatsSocket
	} else {
		sockName = c.socketName
	}

	if _, err := os.Stat(sockName); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("stats socket file %q does not exists, ensure that VPP is running with `statseg { ... }` section in config", sockName)
		}
		return fmt.Errorf("stats socket file error: %v", err)
	}

	rc := C.govpp_stat_connect(C.CString(sockName))
	if rc != 0 {
		return fmt.Errorf("connecting to VPP stats API failed (rc=%v)", rc)
	}

	globalStatClient = c
	return nil
}

func (c *statClient) Disconnect() error {
	globalStatClient = nil

	C.govpp_stat_disconnect()
	return nil
}

func (c *statClient) ListStats(patterns ...string) (stats []string, err error) {
	dir := C.govpp_stat_segment_ls(convertStringSlice(patterns))
	if dir == nil {
		return nil, adapter.ErrStatDirBusy
	}
	defer C.govpp_stat_segment_vec_free(unsafe.Pointer(dir))

	l := C.govpp_stat_segment_vec_len(unsafe.Pointer(dir))
	for i := 0; i < int(l); i++ {
		nameChar := C.govpp_stat_segment_dir_index_to_name(dir, C.uint32_t(i))
		stats = append(stats, C.GoString(nameChar))
		C.free(unsafe.Pointer(nameChar))
	}

	return stats, nil
}

func (c *statClient) DumpStats(patterns ...string) (stats []*adapter.StatEntry, err error) {
	dir := C.govpp_stat_segment_ls(convertStringSlice(patterns))
	if dir == nil {
		return nil, adapter.ErrStatDirBusy
	}
	defer C.govpp_stat_segment_vec_free(unsafe.Pointer(dir))

	dump := C.govpp_stat_segment_dump(dir)
	if dump == nil {
		return nil, adapter.ErrStatDumpBusy
	}
	defer C.govpp_stat_segment_data_free(dump)

	l := C.govpp_stat_segment_vec_len(unsafe.Pointer(dump))
	for i := 0; i < int(l); i++ {
		v := C.govpp_stat_segment_dump_index(dump, C.int(i))
		nameChar := v.name
		name := C.GoString(nameChar)
		typ := adapter.StatType(C.govpp_stat_segment_data_type(&v))

		stat := &adapter.StatEntry{
			Name: name,
			Type: typ,
		}

		switch typ {
		case adapter.ScalarIndex:
			stat.Data = adapter.ScalarStat(C.govpp_stat_segment_data_get_scalar_value(&v))

		case adapter.ErrorIndex:
			stat.Data = adapter.ErrorStat(C.govpp_stat_segment_data_get_error_value(&v))

		case adapter.SimpleCounterVector:
			length := int(C.govpp_stat_segment_vec_len(unsafe.Pointer(C.govpp_stat_segment_data_get_simple_counter(&v))))
			vector := make([][]adapter.Counter, length)
			for k := 0; k < length; k++ {
				for j := 0; j < int(C.govpp_stat_segment_vec_len(unsafe.Pointer(C.govpp_stat_segment_data_get_simple_counter_index(&v, C.int(k))))); j++ {
					vector[k] = append(vector[k], adapter.Counter(C.govpp_stat_segment_data_get_simple_counter_index_value(&v, C.int(k), C.int(j))))
				}
			}
			stat.Data = adapter.SimpleCounterStat(vector)

		case adapter.CombinedCounterVector:
			length := int(C.govpp_stat_segment_vec_len(unsafe.Pointer(C.govpp_stat_segment_data_get_combined_counter(&v))))
			vector := make([][]adapter.CombinedCounter, length)
			for k := 0; k < length; k++ {
				for j := 0; j < int(C.govpp_stat_segment_vec_len(unsafe.Pointer(C.govpp_stat_segment_data_get_combined_counter_index(&v, C.int(k))))); j++ {
					vector[k] = append(vector[k], adapter.CombinedCounter{
						Packets: adapter.Counter(C.govpp_stat_segment_data_get_combined_counter_index_packets(&v, C.int(k), C.int(j))),
						Bytes:   adapter.Counter(C.govpp_stat_segment_data_get_combined_counter_index_bytes(&v, C.int(k), C.int(j))),
					})
				}
			}
			stat.Data = adapter.CombinedCounterStat(vector)

		case adapter.NameVector:
			length := int(C.govpp_stat_segment_vec_len(unsafe.Pointer(C.govpp_stat_segment_data_get_name_vector(&v))))
			var vector []adapter.Name
			for k := 0; k < length; k++ {
				s := C.govpp_stat_segment_data_get_name_vector_index(&v, C.int(k))
				var name adapter.Name
				if s != nil {
					name = adapter.Name(C.GoString(s))
				}
				vector = append(vector, name)
			}
			stat.Data = adapter.NameStat(vector)

		default:
			fmt.Fprintf(os.Stderr, "invalid stat type: %v (%v)\n", typ, name)
			continue

		}

		stats = append(stats, stat)
	}

	return stats, nil
}

func convertStringSlice(strs []string) **C.uint8_t {
	var arr **C.uint8_t
	for _, str := range strs {
		arr = C.govpp_stat_segment_string_vector(arr, C.CString(str))
	}
	return arr
}
