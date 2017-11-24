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

package testutil

import (
	"fmt"
	"sync"

	"regexp"
	"testing"

	"github.com/ligato/cn-infra/logging"
)

// Logger is a wrapper of Logrus logger. In addition to Logrus functionality,
// it allows to define static log fields that are added to all subsequent log entries. It also automatically
// appends file name and line from which the log comes. In order to distinguish logs from different
// go routines, a tag (number that is based on the stack address) is computed. To achieve better readability,
// numeric value of a tag can be replaced with a string using SetTag function.
type Logger struct {
	access sync.RWMutex
	name   string
	std    *testing.T
	level  logging.LogLevel
}

// NewLogger is a constructor which creates instances of named logger.
// This constructor is called from logRegistry which is useful
// when log levels need to be changed by management API (such as REST).
//
// Example:
//
//    logger := NewLogger("loggerXY")
//    logger.Info()
//
func NewLogger(name string, t *testing.T) *Logger {
	logger := &Logger{
		std:  t,
		name: name,
	}

	return logger
}

var validLoggerName = regexp.MustCompile(`^[a-zA-Z0-9.-]+$`).MatchString

func checkLoggerName(name string) error {
	if !validLoggerName(name) {
		return fmt.Errorf("logger name can contain only alphanum characters, dash and comma")
	}
	return nil
}

// GetName returns the logger name.
func (logger *Logger) GetName() string {
	return logger.name
}

// SetLevel sets the standard logger level.
func (logger *Logger) SetLevel(level logging.LogLevel) {
	logger.access.Lock()
	defer logger.access.Unlock()
	logger.level = level
}

// GetLevel returns the standard logger level.
func (logger *Logger) GetLevel() logging.LogLevel {
	return logger.level
}

func (logger *Logger) withField(key string, value interface{}) *logMsg {
	return logger.withFields(logging.Fields{key: value})
}

// logMsg represent an item to be logged.
type logMsg struct {
	*Logger
	fields logging.Fields
}

// WithField creates an entry from the standard logger and adds a field to
// it. If you want multiple fields, use `WithFields`.
//
// Note that it doesn't log until you call Debug, Print, Info, Warn, Fatal
// or Panic on the logMsg it returns.
func (logger *Logger) WithField(key string, value interface{}) logging.LogWithLevel {
	return logger.withFields(logging.Fields{key: value})
}

func (logger *Logger) withFields(fields logging.Fields) *logMsg {
	f := make(logging.Fields, len(fields))

	for k, v := range fields {
		f[k] = v
	}

	return &logMsg{
		Logger: logger,
		fields: f,
	}
}

// WithFields creates an entry from the standard logger and adds multiple
// fields to it. This is simply a helper for `WithField`, invoking it
// once for each field.
//
// Note that it doesn't log until you call Debug, Print, Info, Warn, Fatal
// or Panic on the logMsg it returns.
func (logger *Logger) WithFields(fields map[string]interface{}) logging.LogWithLevel {
	return logger.withFields(logging.Fields(fields))
}

// Debug logs a message at level Debug on the standard logger.
func (logger *Logger) Debug(args ...interface{}) {
	logger.std.Log(append([]interface{}{"Debug"}, args...)...)
}

// Print logs a message at level Info on the standard logger.
func (logger *Logger) Print(args ...interface{}) {
	logger.std.Log(args...)
}

// Info logs a message at level Info on the standard logger.
func (logger *Logger) Info(args ...interface{}) {
	logger.std.Log(append([]interface{}{"Info"}, args...)...)
}

// Warn logs a message at level Warn on the standard logger.
func (logger *Logger) Warn(args ...interface{}) {
	logger.std.Log(append([]interface{}{"Warn"}, args...)...)
}

// Warning logs a message at level Warn on the standard logger.
func (logger *Logger) Warning(args ...interface{}) {
	logger.std.Log(append([]interface{}{"Warning"}, args...)...)
}

// Error logs a message at level Error on the standard logger.
func (logger *Logger) Error(args ...interface{}) {
	logger.std.Error(append([]interface{}{"Error"}, args...)...)
}

// Panic logs a message at level Panic on the standard logger.
func (logger *Logger) Panic(args ...interface{}) {
	logger.std.Log(append([]interface{}{"Panic"}, args)...)
	panic(args[0])
}

// Fatal logs a message at level Fatal on the standard logger.
func (logger *Logger) Fatal(args ...interface{}) {
	logger.std.Log(append([]interface{}{"Fatal"}, args...)...)
	panic(args[0])
}

// Debugf logs a message at level Debug on the standard logger.
func (logger *Logger) Debugf(format string, args ...interface{}) {
	logger.std.Logf("Debug "+format, args...)
}

// Printf logs a message at level Info on the standard logger.
func (logger *Logger) Printf(format string, args ...interface{}) {
	logger.std.Logf(format, args...)
}

// Infof logs a message at level Info on the standard logger.
func (logger *Logger) Infof(format string, args ...interface{}) {
	logger.std.Logf("Info"+format, args...)
}

// Warnf logs a message at level Warn on the standard logger.
func (logger *Logger) Warnf(format string, args ...interface{}) {
	logger.std.Logf("Warn"+format, args...)
}

// Warningf logs a message at level Warn on the standard logger.
func (logger *Logger) Warningf(format string, args ...interface{}) {
	logger.std.Logf("Warning"+format, args...)
}

// Errorf logs a message at level Error on the standard logger.
func (logger *Logger) Errorf(format string, args ...interface{}) {
	logger.std.Errorf(format, args...)
}

// Panicf logs a message at level Panic on the standard logger.
func (logger *Logger) Panicf(format string, args ...interface{}) {
	logger.std.Logf("Panic"+format, args...)
}

// Fatalf logs a message at level Fatal on the standard logger.
func (logger *Logger) Fatalf(format string, args ...interface{}) {
	logger.std.Logf("Fatal"+format, args...)
}

// Debugln logs a message at level Debug on the standard logger.
func (logger *Logger) Debugln(args ...interface{}) {
	logger.std.Log(append([]interface{}{"Debug"}, args...)...)
}

// Println logs a message at level Info on the standard logger.
func (logger *Logger) Println(args ...interface{}) {
	logger.std.Log(args...)
}

// Infoln logs a message at level Info on the standard logger.
func (logger *Logger) Infoln(args ...interface{}) {
	logger.std.Log(append([]interface{}{"Info"}, args)...)
}

// Warnln logs a message at level Warn on the standard logger.
func (logger *Logger) Warnln(args ...interface{}) {
	logger.std.Log(append([]interface{}{"Warn"}, args...)...)
}

// Warningln logs a message at level Warn on the standard logger.
func (logger *Logger) Warningln(args ...interface{}) {
	logger.std.Log(append([]interface{}{"Warning"}, args...)...)
}

// Errorln logs a message at level Error on the standard logger.
func (logger *Logger) Errorln(args ...interface{}) {
	logger.std.Log(append([]interface{}{"Error"}, args...)...)
}

// Panicln logs a message at level Panic on the standard logger.
func (logger *Logger) Panicln(args ...interface{}) {
	logger.std.Log(append([]interface{}{"Panic"}, args...)...)
}

// Fatalln logs a message at level Fatal on the standard logger.
func (logger *Logger) Fatalln(args ...interface{}) {
	logger.std.Log(append([]interface{}{"Fatall"}, args...)...)
}
