// Copyright (c) 2017 Cisco and/or its affiliates.
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

package logrus

import (
	lg "github.com/Sirupsen/logrus"
	"github.com/ligato/cn-infra/logging"
)

// Tag names for structured fields of log message
const (
	locKey    = "loc"
	tagKey    = "tag"
	loggerKey = "logger"
)

// LogMsg represent an item to be logged
type LogMsg struct {
	logger *Logger
	*lg.Entry
}

// NewEntry creates a new LogMsg instance.
func NewEntry(logger *Logger) *LogMsg {
	return &LogMsg{
		logger: logger,
		Entry:  lg.NewEntry(logger.std),
	}
}

// GetTag returns a tag identifying go routine where the entry was created
func (entry *LogMsg) GetTag() string {
	return entry.logger.GetTag()
}

// GetLineInfo returns the information of line and file that is associated with frame at the given depth on stack.
func (entry *LogMsg) GetLineInfo(depth int) string {
	return entry.logger.GetLineInfo(depth)
}

func (entry *LogMsg) header() *LogMsg {
	return entry.logger.header(1)
}

func (entry *LogMsg) withField(key string, value interface{}, depth ...int) *LogMsg {
	d := 1
	if depth != nil && len(depth) > 0 {
		d += depth[0]
	}

	return entry.withFields(Fields{key: value}, d)
}

// WithField creates an entry with a single field.
func (entry *LogMsg) WithField(key string, value interface{}) logging.LogWithLevel {
	return entry.withField(key, value)
}

// Add a map of fields to the LogMsg.
func (entry *LogMsg) withFields(fields Fields, depth ...int) *LogMsg {
	d := entry.logger.depth + 1
	if depth != nil && len(depth) > 0 {
		d += depth[0]
	}
	f := make(lg.Fields, len(fields))
	for k, v := range fields {
		f[k] = v
	}
	if _, ok := f[tagKey]; !ok {
		f[tagKey] = entry.GetTag()
	}
	if _, ok := f[locKey]; !ok {
		f[locKey] = entry.GetLineInfo(d)
	}
	f[loggerKey] = entry.logger.name
	e := entry.Entry.WithFields(f)
	return &LogMsg{
		logger: entry.logger,
		Entry:  e,
	}
}

// WithFields creates an entry with multiple fields.
func (entry *LogMsg) WithFields(fields map[string]interface{}) logging.LogWithLevel {
	return entry.withFields(Fields(fields))
}
