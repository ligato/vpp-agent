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

// Entry represent an item to be logged
type Entry struct {
	logger *Logger
	*lg.Entry
}

// NewEntry creates a new Entry instance.
func NewEntry(logger *Logger) *Entry {
	return &Entry{
		logger: logger,
		Entry:  lg.NewEntry(logger.std),
	}
}

// GetTag returns a tag identifying go routine where the entry was created
func (entry *Entry) GetTag() string {
	return entry.logger.GetTag()
}

// GetLineInfo returns the information of line and file that is associated with frame at the given depth on stack.
func (entry *Entry) GetLineInfo(depth int) string {
	return entry.logger.GetLineInfo(depth)
}

func (entry *Entry) header() *Entry {
	return entry.logger.header(1)
}

// WithError adds an error as single field (using the key defined in ErrorKey) to the Entry.
func (entry *Entry) WithError(err error) *Entry {
	return entry.withField(ErrorKey, err, 1)
}

func (entry *Entry) withField(key string, value interface{}, depth ...int) *Entry {
	d := 1
	if depth != nil && len(depth) > 0 {
		d += depth[0]
	}

	return entry.withFields(Fields{key: value}, d)
}

// WithField creates an entry with a single field.
func (entry *Entry) WithField(key string, value interface{}) logging.LogWithLevel {
	return entry.withField(key, value)
}

// Add a map of fields to the Entry.
func (entry *Entry) withFields(fields Fields, depth ...int) *Entry {
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
	e := entry.Entry.WithFields(f)
	return &Entry{
		logger: entry.logger,
		Entry:  e,
	}
}

// WithFields creates an entry with multiple fields.
func (entry *Entry) WithFields(fields map[string]interface{}) logging.LogWithLevel {
	return entry.withFields(Fields(fields))
}
