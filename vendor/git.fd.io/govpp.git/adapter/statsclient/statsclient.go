// Copyright (c) 2019 Cisco and/or its affiliates.
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

// Package statsclient is pure Go implementation of VPP stats API client.
package statsclient

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"unsafe"

	logger "github.com/sirupsen/logrus"

	"git.fd.io/govpp.git/adapter"
)

var (
	// Debug is global variable that determines debug mode
	Debug = os.Getenv("DEBUG_GOVPP_STATS") != ""

	// Log is global logger
	Log = logger.New()
)

// init initializes global logger, which logs debug level messages to stdout.
func init() {
	Log.Out = os.Stdout
	if Debug {
		Log.Level = logger.DebugLevel
		Log.Debug("govpp/statsclient: enabled debug mode")
	}
}

// StatsClient is the pure Go implementation for VPP stats API.
type StatsClient struct {
	sockAddr string

	currentEpoch int64
	statSegment
}

// NewStatsClient returns new VPP stats API client.
func NewStatsClient(sockAddr string) *StatsClient {
	if sockAddr == "" {
		sockAddr = adapter.DefaultStatsSocket
	}
	return &StatsClient{
		sockAddr: sockAddr,
	}
}

const sockNotFoundWarn = `stats socket not found at: %s
------------------------------------------------------------
 VPP stats socket is missing!
 Is VPP running with stats segment enabled?

 To enable it add following section to startup config:
   statseg {
     default
   }
------------------------------------------------------------
`

func (c *StatsClient) Connect() error {
	// check if socket exists
	if _, err := os.Stat(c.sockAddr); os.IsNotExist(err) {
		Log.Warnf(sockNotFoundWarn, c.sockAddr)
		return fmt.Errorf("stats socket file %s does not exist", c.sockAddr)
	} else if err != nil {
		return fmt.Errorf("stats socket error: %v", err)
	}

	if err := c.statSegment.connect(c.sockAddr); err != nil {
		return err
	}

	ver := c.readVersion()
	Log.Debugf("stat segment version: %v", ver)

	if err := checkVersion(ver); err != nil {
		return err
	}

	return nil
}

func (c *StatsClient) Disconnect() error {
	if err := c.statSegment.disconnect(); err != nil {
		return err
	}

	return nil
}

func (c *StatsClient) ListStats(patterns ...string) (statNames []string, err error) {
	sa := c.accessStart()
	if sa == nil {
		return nil, fmt.Errorf("access failed")
	}

	dirOffset, _, _ := c.readOffsets()
	Log.Debugf("dirOffset: %v", dirOffset)

	vecLen := vectorLen(unsafe.Pointer(&c.sharedHeader[dirOffset]))
	Log.Debugf("vecLen: %v", vecLen)
	Log.Debugf("unsafe.Sizeof(statSegDirectoryEntry{}): %v", unsafe.Sizeof(statSegDirectoryEntry{}))

	for i := uint64(0); i < vecLen; i++ {
		offset := uintptr(i) * unsafe.Sizeof(statSegDirectoryEntry{})
		dirEntry := (*statSegDirectoryEntry)(add(unsafe.Pointer(&c.sharedHeader[dirOffset]), offset))

		nul := bytes.IndexByte(dirEntry.name[:], '\x00')
		if nul < 0 {
			Log.Debugf("no zero byte found for: %q", dirEntry.name[:])
			continue
		}
		name := string(dirEntry.name[:nul])
		if name == "" {
			Log.Debugf("entry with empty name found (%d)", i)
			continue
		}

		Log.Debugf(" %80q (type: %v, data: %d, offset: %d) ", name, dirEntry.directoryType, dirEntry.unionData, dirEntry.offsetVector)

		if nameMatches(name, patterns) {
			statNames = append(statNames, name)
		}

		// TODO: copy the listed entries elsewhere
	}

	if !c.accessEnd(sa) {
		return nil, adapter.ErrStatDirBusy
	}

	c.currentEpoch = sa.epoch

	return statNames, nil
}

func (c *StatsClient) DumpStats(patterns ...string) (entries []*adapter.StatEntry, err error) {
	epoch, _ := c.readEpoch()
	if c.currentEpoch > 0 && c.currentEpoch != epoch { // TODO: do list stats before dump
		return nil, fmt.Errorf("old data")
	}

	sa := c.accessStart()
	if sa == nil {
		return nil, fmt.Errorf("access failed")
	}

	dirOffset, _, _ := c.readOffsets()
	vecLen := vectorLen(unsafe.Pointer(&c.sharedHeader[dirOffset]))

	for i := uint64(0); i < vecLen; i++ {
		offset := uintptr(i) * unsafe.Sizeof(statSegDirectoryEntry{})
		dirEntry := (*statSegDirectoryEntry)(add(unsafe.Pointer(&c.sharedHeader[dirOffset]), offset))

		nul := bytes.IndexByte(dirEntry.name[:], '\x00')
		if nul < 0 {
			Log.Debugf("no zero byte found for: %q", dirEntry.name[:])
			continue
		}
		name := string(dirEntry.name[:nul])
		if name == "" {
			Log.Debugf("entry with empty name found (%d)", i)
			continue
		}

		Log.Debugf(" - %s (type: %v, data: %v, offset: %v) ", name, dirEntry.directoryType, dirEntry.unionData, dirEntry.offsetVector)

		entry := adapter.StatEntry{
			Name: name,
			Type: adapter.StatType(dirEntry.directoryType),
			Data: c.copyData(dirEntry),
		}

		Log.Debugf("\tentry data: %+v %#v (%T)", entry.Data, entry.Data, entry.Data)

		if nameMatches(entry.Name, patterns) {
			entries = append(entries, &entry)
		}
	}

	if !c.accessEnd(sa) {
		return nil, adapter.ErrStatDumpBusy
	}

	return entries, nil
}

func nameMatches(name string, patterns []string) bool {
	if len(patterns) == 0 {
		return true
	}
	for _, pattern := range patterns {
		matched, err := regexp.MatchString(pattern, name)
		if err == nil && matched {
			return true
		}
	}
	return false
}
