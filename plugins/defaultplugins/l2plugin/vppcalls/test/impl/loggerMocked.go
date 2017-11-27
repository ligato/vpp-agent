package impl

import (
	"github.com/ligato/cn-infra/logging"
)

//MockedLogger implements LogWithLevel and Logger
type MockedLogger struct {

}

// Debug logs using Debug level
func (mockedLogger *MockedLogger) Debug(args ...interface{}) {

}

// Debugf prints formatted log using Debug level
func (mockedLogger *MockedLogger) Debugf(format string, args ...interface{}) {

}

// Info logs using Info level
func (mockedLogger *MockedLogger) Info(args ...interface{}) {

}

// Infof prints formatted log using Info level
func (mockedLogger *MockedLogger) Infof(format string, args ...interface{}) {

}
// Warn logs using Warning level
func (mockedLogger *MockedLogger) Warn(args ...interface{}) {

}
// Warnf prints formatted log using Warn level
func (mockedLogger *MockedLogger) Warnf(format string, args ...interface{}) {

}
// Error logs using Error level
func (mockedLogger *MockedLogger) Error(args ...interface{}) {

}
// Errorf prints formatted log using Error level
func (mockedLogger *MockedLogger) Errorf(format string, args ...interface{}) {

}
// Panic logs using Panic level and panics
func (mockedLogger *MockedLogger) Panic(args ...interface{}) {

}
// Panicf prints formatted log using Panic level and panic
func (mockedLogger *MockedLogger) Panicf(format string, args ...interface{}) {

}
// Fatal logs using Fatal level and calls os.Exit(1)
func (mockedLogger *MockedLogger) Fatal(args ...interface{}) {

}
// Fatalf prints formatted log using Fatal level and calls os.Exit(1)
func (mockedLogger *MockedLogger) Fatalf(format string, args ...interface{}) {

}

// Fatalln  is used
func (mockedLogger *MockedLogger) Fatalln(args ...interface{}) {

}

// Print  is used
func (mockedLogger *MockedLogger) Print(v ...interface{}) {

}

//Printf is used
func (mockedLogger *MockedLogger) Printf(format string, v ...interface{}) {

}

//Println  is used
func (mockedLogger *MockedLogger) Println(v ...interface{}) {

}

// SetLevel is used
func (mockedLogger *MockedLogger) SetLevel(level logging.LogLevel) {

}

// GetLevel returns currently set logLevel
func (mockedLogger *MockedLogger) GetLevel() logging.LogLevel {
	return 0
}
// WithField creates one structured field
func (mockedLogger *MockedLogger) WithField(key string, value interface{}) logging.LogWithLevel {
	return nil
}
// WithFields creates multiple structured fields
func (mockedLogger *MockedLogger) WithFields(fields map[string]interface{}) logging.LogWithLevel {
	return nil
}

// GetName return the logger name
func (mockedLogger *MockedLogger) GetName() string {
	return ""
}