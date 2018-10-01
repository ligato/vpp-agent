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

	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/gomega"

	. "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/internal/test"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/internal/utils"
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

func prefixSelector(prefix string) func(key string) bool {
	return func(key string) bool {
		return strings.HasPrefix(key, prefix)
	}
}

func checkValues(received, expected []KeyValuePair) {
	Expect(len(received)).To(Equal(len(expected)))
	for _, kv := range expected {
		found := false
		for _, kv2 := range received {
			if kv2.Key == kv.Key {
				Expect(proto.Equal(kv2.Value, kv.Value)).To(BeTrue())
				found = true
			}
		}
		Expect(found).To(BeTrue())
	}
}

func checkRecordedValues(recorded, expected []recordedKVPair) {
	Expect(len(recorded)).To(Equal(len(expected)))
	for _, kv := range expected {
		found := false
		for _, kv2 := range recorded {
			if kv2.key == kv.key {
				found = true
				Expect(kv2.value).To(Equal(kv.value))
				Expect(kv2.origin).To(Equal(kv.origin))
			}
		}
		Expect(found).To(BeTrue())
	}
}

func checkTxnOperation(recorded, expected *recordedTxnOp) {
	Expect(recorded.operation).To(Equal(expected.operation))
	Expect(recorded.key).To(Equal(expected.key))
	Expect(recorded.derived).To(Equal(expected.derived))
	if expected.prevValue == "" {
		Expect(recorded.prevValue).To(Equal(utils.ProtoToString(nil)))
	} else {
		Expect(recorded.prevValue).To(Equal(expected.prevValue))
	}
	if expected.newValue == "" {
		Expect(recorded.newValue).To(Equal(utils.ProtoToString(nil)))
	} else {
		Expect(recorded.newValue).To(Equal(expected.newValue))
	}
	Expect(recorded.prevOrigin).To(Equal(expected.prevOrigin))
	Expect(recorded.newOrigin).To(Equal(expected.newOrigin))
	Expect(recorded.wasPending).To(Equal(expected.wasPending))
	Expect(recorded.isPending).To(Equal(expected.isPending))
	if expected.prevErr == nil {
		Expect(recorded.prevErr).To(BeNil())
	} else {
		Expect(recorded.prevErr).ToNot(BeNil())
		Expect(recorded.prevErr.Error()).To(BeEquivalentTo(expected.prevErr.Error()))
	}
	if expected.newErr == nil {
		Expect(recorded.newErr).To(BeNil())
	} else {
		Expect(recorded.newErr).ToNot(BeNil())
		Expect(recorded.newErr.Error()).To(BeEquivalentTo(expected.newErr.Error()))
	}
	Expect(recorded.isRevert).To(Equal(expected.isRevert))
	Expect(recorded.isRetry).To(Equal(expected.isRetry))
}

func checkTxnOperations(recorded, expected recordedTxnOps) {
	Expect(recorded).To(HaveLen(len(expected)))
	for idx, recordedOp := range recorded {
		checkTxnOperation(recordedOp, expected[idx])
	}
}

func checkValuesForCorrelation(received, expected []KVWithMetadata) {
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
