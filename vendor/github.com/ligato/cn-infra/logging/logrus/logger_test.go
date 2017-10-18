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
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"testing"

	lg "github.com/Sirupsen/logrus"
	"github.com/ligato/cn-infra/logging"
	"github.com/onsi/gomega"
)

func LogAndAssertJSON(t *testing.T, log func(*Logger), assertions func(fields lg.Fields)) {
	var buffer bytes.Buffer
	var fields lg.Fields

	gomega.RegisterTestingT(t)

	logger := NewLogger("testLogger")
	logger.SetOutput(&buffer)
	logger.SetFormatter(new(lg.JSONFormatter))

	log(logger)

	err := json.Unmarshal(buffer.Bytes(), &fields)
	gomega.Expect(err).To(gomega.BeNil())

	assertions(fields)
}

func LogAndAssertText(t *testing.T, log func(*Logger), assertions func(fields map[string]string)) {
	var buffer bytes.Buffer
	gomega.RegisterTestingT(t)

	logger := NewLogger("testLogger")
	logger.SetOutput(&buffer)
	logger.SetFormatter(&lg.TextFormatter{
		DisableColors: true,
	})

	log(logger)

	fields := make(map[string]string)
	for _, kv := range strings.Split(buffer.String(), " ") {
		if !strings.Contains(kv, "=") {
			continue
		}
		kvArr := strings.Split(kv, "=")
		key := strings.TrimSpace(kvArr[0])
		val := kvArr[1]
		if kvArr[1][0] == '"' {
			var err error
			val, err = strconv.Unquote(val)
			gomega.Expect(err).To(gomega.BeNil())
		}
		fields[key] = val
	}
	assertions(fields)
}

func TestPrint(t *testing.T) {
	LogAndAssertJSON(t, func(log *Logger) {
		log.Print("test")
	}, func(fields lg.Fields) {
		gomega.Expect(fields["msg"]).To(gomega.BeEquivalentTo("test"))
		gomega.Expect(fields["level"]).To(gomega.BeEquivalentTo("info"))
	})
}

func TestInfo(t *testing.T) {
	LogAndAssertJSON(t, func(log *Logger) {
		log.Info("test")
	}, func(fields lg.Fields) {
		gomega.Expect(fields["msg"]).To(gomega.BeEquivalentTo("test"))
		gomega.Expect(fields["level"]).To(gomega.BeEquivalentTo("info"))
	})
}

func TestWarn(t *testing.T) {
	LogAndAssertJSON(t, func(log *Logger) {
		log.Warn("test")
	}, func(fields lg.Fields) {
		gomega.Expect(fields["msg"]).To(gomega.BeEquivalentTo("test"))
		gomega.Expect(fields["level"]).To(gomega.BeEquivalentTo("warning"))
	})
}

func TestInfolnShouldAddSpacesBetweenStrings(t *testing.T) {
	LogAndAssertJSON(t, func(log *Logger) {
		log.Infoln("test", "test")
	}, func(fields lg.Fields) {
		gomega.Expect(fields["msg"]).To(gomega.BeEquivalentTo("test test"))
	})
}

func TestInfolnShouldAddSpacesBetweenStringAndNonstring(t *testing.T) {
	LogAndAssertJSON(t, func(log *Logger) {
		log.Infoln("test", 10)
	}, func(fields lg.Fields) {
		gomega.Expect(fields["msg"]).To(gomega.BeEquivalentTo("test 10"))
	})
}

func TestInfolnShouldAddSpacesBetweenTwoNonStrings(t *testing.T) {
	LogAndAssertJSON(t, func(log *Logger) {
		log.Infoln(10, 10)
	}, func(fields lg.Fields) {
		gomega.Expect(fields["msg"]).To(gomega.BeEquivalentTo("10 10"))
	})
}

func TestInfoShouldAddSpacesBetweenTwoNonStrings(t *testing.T) {
	LogAndAssertJSON(t, func(log *Logger) {
		log.Infoln(10, 10)
	}, func(fields lg.Fields) {
		gomega.Expect(fields["msg"]).To(gomega.BeEquivalentTo("10 10"))
	})
}

func TestInfoShouldNotAddSpacesBetweenStringAndNonstring(t *testing.T) {
	LogAndAssertJSON(t, func(log *Logger) {
		log.Info("test", 10)
	}, func(fields lg.Fields) {
		gomega.Expect(fields["msg"]).To(gomega.BeEquivalentTo("test10"))
	})
}

func TestInfoShouldNotAddSpacesBetweenStrings(t *testing.T) {
	LogAndAssertJSON(t, func(log *Logger) {
		log.Info("test", "test")
	}, func(fields lg.Fields) {
		gomega.Expect(fields["msg"]).To(gomega.BeEquivalentTo("testtest"))
	})
}

func TestWithFieldsShouldAllowAssignments(t *testing.T) {
	var buffer bytes.Buffer
	var fields Fields

	logger := NewLogger("testLogger")
	logger.SetOutput(&buffer)
	logger.SetFormatter(new(lg.JSONFormatter))
	entry := NewEntry(logger)

	entry2 := entry.WithFields(lg.Fields{
		"key1": "value1",
	})

	entry2.WithField("key2", "value2").Info("test")
	err := json.Unmarshal(buffer.Bytes(), &fields)
	gomega.Expect(err).To(gomega.BeNil())

	gomega.Expect("value2").To(gomega.BeEquivalentTo(fields["key2"]))
	gomega.Expect("value1").To(gomega.BeEquivalentTo(fields["key1"]))

	buffer = bytes.Buffer{}
	fields = Fields{}
	entry2.Info("test")
	err = json.Unmarshal(buffer.Bytes(), &fields)
	gomega.Expect(err).To(gomega.BeNil())

	_, ok := fields["key2"]
	gomega.Expect(ok).To(gomega.BeFalse())
	gomega.Expect(fields["key1"]).To(gomega.BeEquivalentTo("value1"))
}

func TestUserSuppliedFieldDoesNotOverwriteDefaults(t *testing.T) {
	LogAndAssertJSON(t, func(log *Logger) {
		log.WithField("msg", "hello").Info("test")
	}, func(fields lg.Fields) {
		gomega.Expect(fields["msg"]).To(gomega.BeEquivalentTo("test"))
	})
}

func TestUserSuppliedMsgFieldHasPrefix(t *testing.T) {
	LogAndAssertJSON(t, func(log *Logger) {
		log.WithField("msg", "hello").Info("test")
	}, func(fields lg.Fields) {
		gomega.Expect(fields["msg"]).To(gomega.BeEquivalentTo("test"))
		gomega.Expect(fields["fields.msg"]).To(gomega.BeEquivalentTo("hello"))
	})
}

func TestUserSuppliedTimeFieldHasPrefix(t *testing.T) {
	LogAndAssertJSON(t, func(log *Logger) {
		log.WithField("time", "hello").Info("test")
	}, func(fields lg.Fields) {
		gomega.Expect(fields["fields.time"]).To(gomega.BeEquivalentTo("hello"))
	})
}

func TestUserSuppliedLevelFieldHasPrefix(t *testing.T) {
	LogAndAssertJSON(t, func(log *Logger) {
		log.WithField("level", 1).Info("test")
	}, func(fields lg.Fields) {
		gomega.Expect(fields["level"]).To(gomega.BeEquivalentTo("info"))
		gomega.Expect(fields["fields.level"]).To(gomega.BeEquivalentTo(1.0)) // JSON has floats only
	})
}

func TestDefaultFieldsAreNotPrefixed(t *testing.T) {
	LogAndAssertText(t, func(log *Logger) {
		ll := log.WithField("herp", "derp")
		ll.Info("hello")
		ll.Info("bye")
	}, func(fields map[string]string) {
		for _, fieldName := range []string{"fields.level", "fields.time", "fields.msg"} {
			if _, ok := fields[fieldName]; ok {
				t.Fatalf("should not have prefixed %q: %v", fieldName, fields)
			}
		}
	})
}

func TestDoubleLoggingDoesntPrefixPreviousFields(t *testing.T) {

	var buffer bytes.Buffer
	var fields Fields

	logger := NewLogger("testLogger")
	logger.SetOutput(&buffer)
	logger.SetFormatter(new(lg.JSONFormatter))

	llog := logger.WithField("context", "eating raw fish")

	llog.Info("looks delicious")

	err := json.Unmarshal(buffer.Bytes(), &fields)
	gomega.Expect(err).To(gomega.BeNil(), "should have decoded first message")
	gomega.Expect(len(fields)).To(gomega.BeEquivalentTo(7), "should only have msg/time/level/context/loc/tag/logger fields")
	gomega.Expect(fields["msg"]).To(gomega.BeEquivalentTo("looks delicious"))
	gomega.Expect(fields["context"]).To(gomega.BeEquivalentTo("eating raw fish"))

	buffer.Reset()

	llog.Warn("omg it is!")

	err = json.Unmarshal(buffer.Bytes(), &fields)
	gomega.Expect(err).To(gomega.BeNil(), "should have decoded second message")
	gomega.Expect(len(fields)).To(gomega.BeEquivalentTo(7), "should only have msg/time/level/context/logger fields")
	gomega.Expect(fields["msg"]).To(gomega.BeEquivalentTo("omg it is!"))
	gomega.Expect(fields["context"]).To(gomega.BeEquivalentTo("eating raw fish"))
	gomega.Expect(fields["fields.msg"]).To(gomega.BeNil(), "should not have prefixed previous `msg` entry")
}

func TestGetSetLevelRace(t *testing.T) {
	logger := NewLogger("testLogger")

	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				logger.SetLevel(logging.InfoLevel)
			} else {
				logger.GetLevel()
			}
		}(i)

	}
	wg.Wait()
}

func TestLoggingRace(t *testing.T) {
	logger := NewLogger("testLogger")

	var wg sync.WaitGroup
	wg.Add(100)

	for i := 0; i < 100; i++ {
		go func() {
			logger.Info("info")
			wg.Done()
		}()
	}
	wg.Wait()
}

// Compile test
func TestLogInterface(t *testing.T) {
	var buffer bytes.Buffer
	fn := func(l *Logger) {
		b := l.WithField("key", "value")
		b.Debug("Test")
	}
	// test logger
	logger := NewLogger("testLogger")
	logger.SetOutput(&buffer)

	fn(logger)

	// test Entry
	e := logger.withField("another", "value")
	fn(e.logger)
}

func TestSetTag(t *testing.T) {
	LogAndAssertJSON(t, func(log *Logger) {
		log.SetTag("testtag")
		log.Info("hello")
	}, func(fields lg.Fields) {
		gomega.Expect(fields["tag"]).To(gomega.BeEquivalentTo("testtag"))
	})
}

func TestClearTag(t *testing.T) {
	LogAndAssertJSON(t, func(log *Logger) {
		log.SetTag("testtag")
		log.ClearTag()
		log.Info("hello")
	}, func(fields lg.Fields) {
		gomega.Expect(fields["tag"]).To(gomega.BeEquivalentTo("00000000"))
	})
}

func TestInitTag(t *testing.T) {
	LogAndAssertJSON(t, func(log *Logger) {
		log.InitTag("testtag")
		log.Info("hello")
	}, func(fields lg.Fields) {
		gomega.Expect(fields["tag"]).To(gomega.BeEquivalentTo("testtag"))
	})
}

func TestGetTag(t *testing.T) {
	LogAndAssertJSON(t, func(log *Logger) {
		log.SetTag("testtag")
		tag := log.GetTag()
		log.Info(tag)
	}, func(fields lg.Fields) {
		gomega.Expect(fields["msg"]).To(gomega.BeEquivalentTo("testtag"))
	})
}
