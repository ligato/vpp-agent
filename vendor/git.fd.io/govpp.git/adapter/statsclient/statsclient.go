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

	logger "github.com/sirupsen/logrus"

	"git.fd.io/govpp.git/adapter"
)

const (
	// DefaultSocketName is default VPP stats socket file path.
	DefaultSocketName = adapter.DefaultStatsSocket
)

const socketMissing = `
------------------------------------------------------------
 VPP stats socket file %s is missing!

  - is VPP running with stats segment enabled?
  - is the correct socket name configured?

 To enable it add following section to your VPP config:
   statseg {
     default
   }
------------------------------------------------------------
`

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

func debugf(f string, a ...interface{}) {
	if Debug {
		Log.Debugf(f, a...)
	}
}

// implements StatsAPI
var _ adapter.StatsAPI = (*StatsClient)(nil)

// StatsClient is the pure Go implementation for VPP stats API.
type StatsClient struct {
	sockAddr string

	statSegment
}

// NewStatsClient returns new VPP stats API client.
func NewStatsClient(sockAddr string) *StatsClient {
	if sockAddr == "" {
		sockAddr = DefaultSocketName
	}
	return &StatsClient{
		sockAddr: sockAddr,
	}
}

func (c *StatsClient) Connect() error {
	// check if socket exists
	if _, err := os.Stat(c.sockAddr); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, socketMissing, c.sockAddr)
		return fmt.Errorf("stats socket file %s does not exist", c.sockAddr)
	} else if err != nil {
		return fmt.Errorf("stats socket error: %v", err)
	}

	if err := c.statSegment.connect(c.sockAddr); err != nil {
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

func (c *StatsClient) ListStats(patterns ...string) (names []string, err error) {
	sa := c.accessStart()
	if sa.epoch == 0 {
		return nil, adapter.ErrStatsAccessFailed
	}

	indexes, err := c.listIndexes(patterns...)
	if err != nil {
		return nil, err
	}

	dirVector := c.getStatDirVector()
	vecLen := uint32(vectorLen(dirVector))

	for _, index := range indexes {
		if index >= vecLen {
			return nil, fmt.Errorf("stat entry index %d out of dir vector len (%d)", index, vecLen)
		}

		dirEntry := c.getStatDirIndex(dirVector, index)
		var name []byte
		for n := 0; n < len(dirEntry.name); n++ {
			if dirEntry.name[n] == 0 {
				name = dirEntry.name[:n]
				break
			}
		}
		names = append(names, string(name))
	}

	if !c.accessEnd(&sa) {
		return nil, adapter.ErrStatsDataBusy
	}

	return names, nil
}

func (c *StatsClient) DumpStats(patterns ...string) (entries []adapter.StatEntry, err error) {
	sa := c.accessStart()
	if sa.epoch == 0 {
		return nil, adapter.ErrStatsAccessFailed
	}

	indexes, err := c.listIndexes(patterns...)
	if err != nil {
		return nil, err
	}
	if entries, err = c.dumpEntries(indexes); err != nil {
		return nil, err
	}

	if !c.accessEnd(&sa) {
		return nil, adapter.ErrStatsDataBusy
	}

	return entries, nil
}

func (c *StatsClient) PrepareDir(patterns ...string) (*adapter.StatDir, error) {
	dir := new(adapter.StatDir)

	sa := c.accessStart()
	if sa.epoch == 0 {
		return nil, adapter.ErrStatsAccessFailed
	}

	indexes, err := c.listIndexes(patterns...)
	if err != nil {
		return nil, err
	}
	dir.Indexes = indexes

	entries, err := c.dumpEntries(indexes)
	if err != nil {
		return nil, err
	}
	dir.Entries = entries

	if !c.accessEnd(&sa) {
		return nil, adapter.ErrStatsDataBusy
	}
	dir.Epoch = sa.epoch

	return dir, nil
}

func (c *StatsClient) UpdateDir(dir *adapter.StatDir) (err error) {
	epoch, _ := c.getEpoch()
	if dir.Epoch != epoch {
		return adapter.ErrStatsDirStale
	}

	sa := c.accessStart()
	if sa.epoch == 0 {
		return adapter.ErrStatsAccessFailed
	}

	dirVector := c.getStatDirVector()

	for i, index := range dir.Indexes {
		dirEntry := c.getStatDirIndex(dirVector, index)

		var name []byte
		for n := 0; n < len(dirEntry.name); n++ {
			if dirEntry.name[n] == 0 {
				name = dirEntry.name[:n]
				break
			}
		}
		if len(name) == 0 {
			continue
		}

		entry := &dir.Entries[i]
		if !bytes.Equal(name, entry.Name) {
			continue
		}
		if adapter.StatType(dirEntry.directoryType) != entry.Type {
			continue
		}
		if entry.Data == nil {
			continue
		}
		if err := c.updateEntryData(dirEntry, &entry.Data); err != nil {
			return fmt.Errorf("updating stat data for entry %s failed: %v", name, err)
		}

	}

	if !c.accessEnd(&sa) {
		return adapter.ErrStatsDataBusy
	}

	return nil
}

// listIndexes lists indexes for all stat entries that match any of the regex patterns.
func (c *StatsClient) listIndexes(patterns ...string) (indexes []uint32, err error) {
	if len(patterns) == 0 {
		return c.listIndexesFunc(nil)
	}
	var regexes = make([]*regexp.Regexp, len(patterns))
	for i, pattern := range patterns {
		r, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("compiling regexp failed: %v", err)
		}
		regexes[i] = r
	}
	nameMatches := func(name []byte) bool {
		for _, r := range regexes {
			if r.Match(name) {
				return true
			}
		}
		return false
	}
	return c.listIndexesFunc(nameMatches)
}

func (c *StatsClient) listIndexesFunc(f func(name []byte) bool) (indexes []uint32, err error) {
	if f == nil {
		// there is around ~3157 stats, so to avoid too many allocations
		// we set capacity to 3200 when listing all stats
		indexes = make([]uint32, 0, 3200)
	}

	dirVector := c.getStatDirVector()
	vecLen := uint32(vectorLen(dirVector))

	for i := uint32(0); i < vecLen; i++ {
		dirEntry := c.getStatDirIndex(dirVector, i)

		if f != nil {
			var name []byte
			for n := 0; n < len(dirEntry.name); n++ {
				if dirEntry.name[n] == 0 {
					name = dirEntry.name[:n]
					break
				}
			}
			if len(name) == 0 || !f(name) {
				continue
			}
		}
		indexes = append(indexes, i)
	}

	return indexes, nil
}

func (c *StatsClient) dumpEntries(indexes []uint32) (entries []adapter.StatEntry, err error) {
	dirVector := c.getStatDirVector()
	dirLen := uint32(vectorLen(dirVector))

	debugf("dumping entres for %d indexes", len(indexes))

	entries = make([]adapter.StatEntry, 0, len(indexes))
	for _, index := range indexes {
		if index >= dirLen {
			return nil, fmt.Errorf("stat entry index %d out of dir vector length (%d)", index, dirLen)
		}

		dirEntry := c.getStatDirIndex(dirVector, index)

		var name []byte
		for n := 0; n < len(dirEntry.name); n++ {
			if dirEntry.name[n] == 0 {
				name = dirEntry.name[:n]
				break
			}
		}

		if Debug {
			debugf(" - %3d. dir: %q type: %v offset: %d union: %d", index, name,
				adapter.StatType(dirEntry.directoryType), dirEntry.offsetVector, dirEntry.unionData)
		}

		if len(name) == 0 {
			continue
		}

		entry := adapter.StatEntry{
			Name: append([]byte(nil), name...),
			Type: adapter.StatType(dirEntry.directoryType),
			Data: c.copyEntryData(dirEntry),
		}
		entries = append(entries, entry)
	}

	return entries, nil
}
