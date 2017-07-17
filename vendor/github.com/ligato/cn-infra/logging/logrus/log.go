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
	"io"

	lg "github.com/Sirupsen/logrus"
	"github.com/ligato/cn-infra/logging"
)

var (
	// logf is the default logger used by package global functions.
	logf  *Logger
	depth int
	// ErrorKey is the key used to log an error object in structured form. See WithFields function.
	ErrorKey string
)

// Fields is a type for structured log entries.
type Fields map[string]interface{}

const locKey = "loc"
const tagKey = "tag"

const defaultLoggerName = "defaultLogger"

func init() {
	LoggerRegistry = &LogRegistry{mapping: map[string]*Logger{}}
	logf, _ = NewNamed(defaultLoggerName)
	depth = 2
	ErrorKey = lg.ErrorKey
}

// StandardLogger default logger instance used by package level functions.
func StandardLogger() *Logger {
	return logf
}

// InitTag sets the tag for the main go routine in the standard logger.
func InitTag(tag ...string) {
	logf.InitTag(tag...)
}

// GetTag returns the tag set for the current go routine in the standard logger.
func GetTag() string {
	return logf.GetTag()
}

// SetTag sets a tag in the standard logger.
func SetTag(tag ...string) {
	logf.SetTag(tag...)
}

// ClearTag remove a previously set tag in the standard logger.
func ClearTag() {
	logf.ClearTag()
}

// SetOutput sets the standard logger output.
func SetOutput(out io.Writer) {
	logf.SetOutput(out)
}

// SetFormatter sets the standard logger formatter.
func SetFormatter(formatter lg.Formatter) {
	logf.SetFormatter(formatter)
}

// SetLevel sets the standard logger level.
func SetLevel(level logging.LogLevel) {
	logf.SetLevel(level)
}

// GetLevel returns the standard logger level.
func GetLevel() logging.LogLevel {
	return logf.GetLevel()
}

// AddHook adds a hook to the standard logger hooks.
func AddHook(hook lg.Hook) {
	logf.AddHook(hook)
}

// WithError creates an entry from the standard logger and adds an error to it, using the value defined in ErrorKey as key.
func WithError(err error) *Entry {
	entry := logf.withField(ErrorKey, err, 1)
	return entry
}

// WithField creates an entry from the standard logger and adds a field to
// it. If you want multiple fields, use `WithFields`.
//
// Note that it doesn't log until you call Debug, Print, Info, Warn, Fatal
// or Panic on the Entry it returns.
func WithField(key string, value interface{}) logging.LogWithLevel {
	entry := logf.withField(key, value, 1)
	return entry
}

// WithFields creates an entry from the standard logger and adds multiple
// fields to it. This is simply a helper for `WithField`, invoking it
// once for each field.
//
// Note that it doesn't log until you call Debug, Print, Info, Warn, Fatal
// or Panic on the Entry it returns.
func WithFields(fields map[string]interface{}) *Entry {
	entry := logf.withFields(Fields(fields), 1)
	return entry
}

func header(d int) *Entry {
	t := logf.GetTag()
	l := logf.GetLineInfo(depth + d)
	e := WithFields(Fields{
		tagKey: t,
		locKey: l,
	})
	return e
}

// Debug logs a message at level Debug on the standard logger.
func Debug(args ...interface{}) {
	header(1).Debug(args...)
}

// Print logs a message at level Info on the standard logger.
func Print(args ...interface{}) {
	header(1).Print(args...)
}

// Info logs a message at level Info on the standard logger.
func Info(args ...interface{}) {
	header(1).Info(args...)
}

// Warn logs a message at level Warn on the standard logger.
func Warn(args ...interface{}) {
	header(1).Warn(args...)
}

// Warning logs a message at level Warn on the standard logger.
func Warning(args ...interface{}) {
	header(1).Warning(args...)
}

// Error logs a message at level Error on the standard logger.
func Error(args ...interface{}) {
	header(1).Error(args...)
}

// Panic logs a message at level Panic on the standard logger.
func Panic(args ...interface{}) {
	header(1).Panic(args...)
}

// Fatal logs a message at level Fatal on the standard logger.
func Fatal(args ...interface{}) {
	header(1).Fatal(args...)
}

// Debugf logs a message at level Debug on the standard logger.
func Debugf(format string, args ...interface{}) {
	header(1).Debugf(format, args...)
}

// Printf logs a message at level Info on the standard logger.
func Printf(format string, args ...interface{}) {
	header(1).Printf(format, args...)
}

// Infof logs a message at level Info on the standard logger.
func Infof(format string, args ...interface{}) {
	header(1).Infof(format, args...)
}

// Warnf logs a message at level Warn on the standard logger.
func Warnf(format string, args ...interface{}) {
	header(1).Warnf(format, args...)
}

// Warningf logs a message at level Warn on the standard logger.
func Warningf(format string, args ...interface{}) {
	header(1).Warningf(format, args...)
}

// Errorf logs a message at level Error on the standard logger.
func Errorf(format string, args ...interface{}) {
	header(1).Errorf(format, args...)
}

// Panicf logs a message at level Panic on the standard logger.
func Panicf(format string, args ...interface{}) {
	header(1).Panicf(format, args...)
}

// Fatalf logs a message at level Fatal on the standard logger.
func Fatalf(format string, args ...interface{}) {
	header(1).Fatalf(format, args...)
}

// Debugln logs a message at level Debug on the standard logger.
func Debugln(args ...interface{}) {
	header(1).Debugln(args...)
}

// Println logs a message at level Info on the standard logger.
func Println(args ...interface{}) {
	header(1).Println(args...)
}

// Infoln logs a message at level Info on the standard logger.
func Infoln(args ...interface{}) {
	header(1).Infoln(args...)
}

// Warnln logs a message at level Warn on the standard logger.
func Warnln(args ...interface{}) {
	header(1).Warnln(args...)
}

// Warningln logs a message at level Warn on the standard logger.
func Warningln(args ...interface{}) {
	header(1).Warningln(args...)
}

// Errorln logs a message at level Error on the standard logger.
func Errorln(args ...interface{}) {
	header(1).Errorln(args...)
}

// Panicln logs a message at level Panic on the standard logger.
func Panicln(args ...interface{}) {
	header(1).Panicln(args...)
}

// Fatalln logs a message at level Fatal on the standard logger.
func Fatalln(args ...interface{}) {
	header(1).Fatalln(args...)
}
