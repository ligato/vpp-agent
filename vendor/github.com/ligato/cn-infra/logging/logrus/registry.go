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
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/ligato/cn-infra/logging"
	"sync"
)

// LoggerRegistry holds all created loggers
var LoggerRegistry *LogRegistry

// LogRegistry contains logger map and rwlock guarding access to it
type LogRegistry struct {
	// mapping holds logger instances indexed by their names
	mapping map[string]*Logger
	rwmutex sync.RWMutex
}

// ListLoggers returns a map (loggerName => log level)
func (lr *LogRegistry) ListLoggers() map[string]string {
	lr.rwmutex.RLock()
	defer lr.rwmutex.RUnlock()
	list := map[string]string{}
	for k, v := range lr.mapping {
		list[k] = v.GetLevel().String()
	}
	return list
}

// SetLevel modifies log level of selected logger in the registry
func (lr *LogRegistry) SetLevel(logger, level string) error {
	lr.rwmutex.RLock()
	defer lr.rwmutex.RUnlock()
	lg, ok := lr.mapping[logger]
	if !ok {
		return fmt.Errorf("Logger %s not found", logger)
	}
	lvl, err := logrus.ParseLevel(level)
	if err == nil {
		switch lvl {
		case logrus.DebugLevel:
			lg.SetLevel(logging.DebugLevel)
		case logrus.InfoLevel:
			lg.SetLevel(logging.InfoLevel)
		case logrus.WarnLevel:
			lg.SetLevel(logging.WarnLevel)
		case logrus.ErrorLevel:
			lg.SetLevel(logging.ErrorLevel)
		case logrus.PanicLevel:
			lg.SetLevel(logging.PanicLevel)
		case logrus.FatalLevel:
			lg.SetLevel(logging.FatalLevel)
		}

	}
	return nil
}

// GetLevel returns the currently set log level of the logger
func (lr *LogRegistry) GetLevel(logger string) (string, error) {
	lr.rwmutex.RLock()
	defer lr.rwmutex.RUnlock()
	lg, ok := lr.mapping[logger]
	if !ok {
		return "", fmt.Errorf("Logger %s not found", logger)
	}
	return lg.GetLevel().String(), nil
}

// Lookup returns a logger instance identified by name from registry
func (lr *LogRegistry) Lookup(loggerName string) (logger logging.Logger, found bool) {
	lr.rwmutex.RLock()
	defer lr.rwmutex.RUnlock()
	logger, found = lr.mapping[loggerName]
	return
}

// ClearRegistry removes all loggers except the default one from registry
func (lr *LogRegistry) ClearRegistry() {
	lr.rwmutex.Lock()
	defer lr.rwmutex.Unlock()

	for k := range lr.mapping {
		if k != defaultLoggerName {
			delete(lr.mapping, k)
		}
	}
}

// addLogger inserts logger into map
func (lr *LogRegistry) addLogger(name string, logger *Logger) {
	lr.rwmutex.Lock()
	defer lr.rwmutex.Unlock()

	lr.mapping[name] = logger
}
