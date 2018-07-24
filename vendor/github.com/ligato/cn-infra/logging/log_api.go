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

package logging

import "fmt"

var (
	// DefaultLogger is the default logger
	DefaultLogger Logger

	// DefaultRegistry is the default logging registry
	DefaultRegistry Registry
)

// LogLevel represents severity of log record
type LogLevel uint32

const (
	// PanicLevel - highest level of severity. Logs and then calls panic with the message passed in.
	PanicLevel LogLevel = iota
	// FatalLevel - logs and then calls `os.Exit(1)`.
	FatalLevel
	// ErrorLevel - used for errors that should definitely be noted.
	ErrorLevel
	// WarnLevel - non-critical entries that deserve eyes.
	WarnLevel
	// InfoLevel - general operational entries about what's going on inside the application.
	InfoLevel
	// DebugLevel - enabled for debugging, very verbose logging.
	DebugLevel
)

// String converts the LogLevel to a string. E.g. PanicLevel becomes "panic".
func (level LogLevel) String() string {
	switch level {
	case PanicLevel:
		return "panic"
	case FatalLevel:
		return "fatal"
	case ErrorLevel:
		return "error"
	case WarnLevel:
		return "warning"
	case InfoLevel:
		return "info"
	case DebugLevel:
		return "debug"
	}

	return fmt.Sprintf("unknown(%d)", level)
}

// Fields is a type accepted by WithFields method. It can be used to instantiate map using shorter notation.
type Fields map[string]interface{}

// LogWithLevel allows to log with different log levels
type LogWithLevel interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Fatalln(args ...interface{})
	Panic(args ...interface{})
	Panicf(format string, args ...interface{})

	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

// Logger provides logging capabilities
type Logger interface {
	// GetName return the logger name
	GetName() string
	// SetLevel modifies the LogLevel
	SetLevel(level LogLevel)
	// GetLevel returns currently set logLevel
	GetLevel() LogLevel
	// WithField creates one structured field
	WithField(key string, value interface{}) LogWithLevel
	// WithFields creates multiple structured fields
	WithFields(fields Fields) LogWithLevel

	LogWithLevel
}

// LoggerFactory is API for the plugins that want to create their own loggers.
type LoggerFactory interface {
	NewLogger(name string) Logger
}

// Registry groups multiple Logger instances and allows to mange their log levels.
type Registry interface {
	// LoggerFactory allow to create new loggers
	LoggerFactory
	// List Loggers returns a map (loggerName => log level)
	ListLoggers() map[string]string
	// SetLevel modifies log level of selected logger in the registry
	SetLevel(logger, level string) error
	// GetLevel returns the currently set log level of the logger from registry
	GetLevel(logger string) (string, error)
	// Lookup returns a logger instance identified by name from registry
	Lookup(loggerName string) (logger Logger, found bool)
	// ClearRegistry removes all loggers except the default one from registry
	ClearRegistry()
}

// PluginLogger is intended for:
// 1. small plugins (that just need one logger; name corresponds to plugin name)
// 2. large plugins that need multiple loggers (all loggers share same name prefix)
type PluginLogger interface {
	// Plugin has by default possibility to log
	// Logger name is initialized with plugin name
	Logger
	// LoggerFactory can be optionally used by large plugins
	// to create child loggers (their names are prefixed by plugin logger name)
	LoggerFactory
}

// ForPlugin is used to initialize plugin logger by name
// and optionally created children (their name prefixed by plugin logger name)
func ForPlugin(name string) PluginLogger {
	if logger, found := DefaultRegistry.Lookup(name); found {
		DefaultLogger.Debugf("using plugin logger for %q that was already initialized", name)
		return &pluginLogger{
			Logger:        logger,
			LoggerFactory: &prefixedLoggerFactory{name, DefaultRegistry},
		}
	}
	return NewPluginLogger(name, DefaultRegistry)
}

// NewPluginLogger creates new logger with given LoggerFactory.
func NewPluginLogger(name string, factory LoggerFactory) PluginLogger {
	return &pluginLogger{
		Logger:        factory.NewLogger(name),
		LoggerFactory: &prefixedLoggerFactory{name, factory},
	}
}

type pluginLogger struct {
	Logger
	LoggerFactory
}

type prefixedLoggerFactory struct {
	prefix  string
	factory LoggerFactory
}

func (p *prefixedLoggerFactory) NewLogger(name string) Logger {
	return p.factory.NewLogger(p.prefix + name)
}
