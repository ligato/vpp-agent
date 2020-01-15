// Copyright (c) 2018 Cisco and/or its affiliates.
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

package kvscheduler

import (
	"strings"

	"github.com/golang/protobuf/proto"
	. "github.com/onsi/gomega"

	. "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler/internal/test"
	. "go.ligato.io/vpp-agent/v3/proto/ligato/kvscheduler"
)

const (
	descriptor1Name = "descriptor1"
	descriptor2Name = "descriptor2"
	descriptor3Name = "descriptor3"

	prefixA = "/prefixA/"
	prefixB = "/prefixB/"
	prefixC = "/prefixC/"

	baseValue1 = "base-value1"
	baseValue2 = "base-value2"
	baseValue3 = "base-value3"
	baseValue4 = "base-value4"
)

type DumpFnc func(string, View) ([]KVWithMetadata, error)

func prefixSelector(prefix string) func(key string) bool {
	return func(key string) bool {
		return strings.HasPrefix(key, prefix)
	}
}

func checkRecordedValues(recorded, expected []RecordedKVPair) {
	Expect(len(recorded)).To(Equal(len(expected)))
	for _, kv := range expected {
		found := false
		for _, kv2 := range recorded {
			if kv2.Key == kv.Key {
				found = true
				Expect(proto.Equal(kv2.Value, kv.Value)).To(BeTrue())
				Expect(kv2.Origin).To(Equal(kv.Origin))
			}
		}
		Expect(found).To(BeTrue())
	}
}

func checkTxnOperation(recorded, expected *RecordedTxnOp) {
	Expect(recorded.Operation).To(Equal(expected.Operation))
	Expect(recorded.Key).To(Equal(expected.Key))
	Expect(proto.Equal(recorded.PrevValue, expected.PrevValue)).To(BeTrue())
	Expect(proto.Equal(recorded.NewValue, expected.NewValue)).To(BeTrue())
	Expect(recorded.PrevState).To(Equal(expected.PrevState))
	Expect(recorded.NewState).To(Equal(expected.NewState))
	if expected.PrevErr == nil {
		Expect(recorded.PrevErr).To(BeNil())
	} else {
		Expect(recorded.PrevErr).ToNot(BeNil())
		Expect(recorded.PrevErr.Error()).To(BeEquivalentTo(expected.PrevErr.Error()))
	}
	if expected.NewErr == nil {
		Expect(recorded.NewErr).To(BeNil())
	} else {
		Expect(recorded.NewErr).ToNot(BeNil())
		Expect(recorded.NewErr.Error()).To(BeEquivalentTo(expected.NewErr.Error()))
	}
	Expect(recorded.NOOP).To(Equal(expected.NOOP))
	Expect(recorded.IsDerived).To(Equal(expected.IsDerived))
	Expect(recorded.IsProperty).To(Equal(expected.IsProperty))
	Expect(recorded.IsRevert).To(Equal(expected.IsRevert))
	Expect(recorded.IsRetry).To(Equal(expected.IsRetry))
	Expect(recorded.IsRecreate).To(Equal(expected.IsRecreate))
}

func checkTxnOperations(recorded, expected RecordedTxnOps) {
	Expect(recorded).To(HaveLen(len(expected)))
	for idx, recordedOp := range recorded {
		checkTxnOperation(recordedOp, expected[idx])
	}
}

func checkValues(received, expected []KVWithMetadata) {
	Expect(received).To(HaveLen(len(expected)))
	for _, kv := range expected {
		found := false
		for _, kv2 := range received {
			if kv2.Key == kv.Key {
				found = true
				Expect(kv2.Origin).To(BeEquivalentTo(kv.Origin))
				Expect(proto.Equal(kv2.Value, kv.Value)).To(BeTrue())
				if kv.Metadata == nil {
					Expect(kv2.Metadata).To(BeNil())
				} else {
					Expect(kv2.Metadata).ToNot(BeNil())
					expIntMeta := kv.Metadata.(*test.OnlyInteger)
					receivedMeta := kv2.Metadata.(*test.OnlyInteger)
					Expect(receivedMeta.GetInteger()).To(BeEquivalentTo(expIntMeta.GetInteger()))
				}
			}
		}
		Expect(found).To(BeTrue())
	}
}

func checkBaseValueStatus(received, expected *BaseValueStatus) {
	checkValueStatus(received.Value, expected.Value)
	Expect(received.DerivedValues).To(HaveLen(len(expected.DerivedValues)))
	for _, expDer := range expected.DerivedValues {
		found := false
		for _, recvDer := range received.DerivedValues {
			if expDer.Key == recvDer.Key {
				checkValueStatus(recvDer, expDer)
				found = true
				break
			}
		}
		Expect(found).To(BeTrue())
	}
}

func checkValueStatus(received, expected *ValueStatus) {
	Expect(received.Error).To(BeEquivalentTo(expected.Error))
	Expect(received.State).To(BeEquivalentTo(expected.State))
	Expect(received.LastOperation).To(BeEquivalentTo(expected.LastOperation))
	Expect(equalStringArrays(received.Details, expected.Details)).To(BeTrue())
}

func equalStringArrays(sa1, sa2 []string) bool {
	if len(sa1) != len(sa2) {
		return false
	}
	for _, s1 := range sa1 {
		found := false
		for _, s2 := range sa2 {
			if s1 == s2 {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
