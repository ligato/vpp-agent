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
	"github.com/onsi/gomega"
	"testing"
)

func TestListLoggers(t *testing.T) {
	gomega.RegisterTestingT(t)
	loggers := LoggerRegistry.ListLoggers()
	gomega.Expect(loggers).NotTo(gomega.BeNil())

	lg, found := loggers[defaultLoggerName]
	gomega.Expect(found).To(gomega.BeTrue())
	gomega.Expect(lg).NotTo(gomega.BeNil())
}

func TestNewLogger(t *testing.T) {
	const loggerName = "myLogger"
	gomega.RegisterTestingT(t)
	lg, err := NewNamed(loggerName)
	gomega.Expect(lg).NotTo(gomega.BeNil())
	gomega.Expect(err).To(gomega.BeNil())

	loggers := LoggerRegistry.ListLoggers()
	gomega.Expect(loggers).NotTo(gomega.BeNil())

	fromRegistry, found := loggers[loggerName]
	gomega.Expect(found).To(gomega.BeTrue())
	gomega.Expect(fromRegistry).NotTo(gomega.BeNil())
}

func TestGetSetLevel(t *testing.T) {
	gomega.RegisterTestingT(t)
	const level = "error"
	//existing logger
	err := LoggerRegistry.SetLevel(defaultLoggerName, level)
	gomega.Expect(err).To(gomega.BeNil())

	loggers := LoggerRegistry.ListLoggers()
	gomega.Expect(loggers).NotTo(gomega.BeNil())

	logger, found := loggers[defaultLoggerName]
	gomega.Expect(found).To(gomega.BeTrue())
	gomega.Expect(logger).NotTo(gomega.BeNil())
	gomega.Expect(loggers[defaultLoggerName]).To(gomega.BeEquivalentTo(level))

	currentLevel, err := LoggerRegistry.GetLevel(defaultLoggerName)
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(level).To(gomega.BeEquivalentTo(currentLevel))

	//non-existing logger
	err = LoggerRegistry.SetLevel("unknown", level)
	gomega.Expect(err).NotTo(gomega.BeNil())

	_, err = LoggerRegistry.GetLevel("unknown")
	gomega.Expect(err).NotTo(gomega.BeNil())
}

func TestGetLoggerByName(t *testing.T) {
	const (
		loggerA = "myLoggerA"
		loggerB = "myLoggerB"
	)
	lgA, err := NewNamed(loggerA)
	gomega.Expect(lgA).NotTo(gomega.BeNil())
	gomega.Expect(err).To(gomega.BeNil())

	lgB, err := NewNamed(loggerB)
	gomega.Expect(lgB).NotTo(gomega.BeNil())
	gomega.Expect(err).To(gomega.BeNil())

	returnedA, found := LoggerRegistry.Lookup(loggerA)
	gomega.Expect(found).To(gomega.BeTrue())
	gomega.Expect(returnedA).To(gomega.BeEquivalentTo(lgA))

	returnedB, found := LoggerRegistry.Lookup(loggerB)
	gomega.Expect(found).To(gomega.BeTrue())
	gomega.Expect(returnedB).To(gomega.BeEquivalentTo(lgB))

	unknown, found := LoggerRegistry.Lookup("unknown")
	gomega.Expect(found).To(gomega.BeFalse())
	gomega.Expect(unknown).To(gomega.BeNil())
}

func TestClearRegistry(t *testing.T) {
	const (
		loggerA = "loggerA"
		loggerB = "loggerB"
	)
	lgA, err := NewNamed(loggerA)
	gomega.Expect(lgA).NotTo(gomega.BeNil())
	gomega.Expect(err).To(gomega.BeNil())

	lgB, err := NewNamed(loggerB)
	gomega.Expect(lgB).NotTo(gomega.BeNil())
	gomega.Expect(err).To(gomega.BeNil())

	LoggerRegistry.ClearRegistry()

	_, found := LoggerRegistry.Lookup(loggerA)
	gomega.Expect(found).To(gomega.BeFalse())

	_, found = LoggerRegistry.Lookup(loggerB)
	gomega.Expect(found).To(gomega.BeFalse())

	_, found = LoggerRegistry.Lookup(defaultLoggerName)
	gomega.Expect(found).To(gomega.BeTrue())
}
