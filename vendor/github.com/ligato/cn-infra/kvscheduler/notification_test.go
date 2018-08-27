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
	"context"
	"errors"
	"testing"
	"time"
	. "github.com/onsi/gomega"

	"github.com/ligato/cn-infra/kvscheduler/test"
	. "github.com/ligato/cn-infra/kvscheduler/api"
	"strings"
)

func TestNotifications(t *testing.T) {
	RegisterTestingT(t)

	// prepare KV Scheduler
	scheduler := NewPlugin(UseDeps(func(deps *Deps) {
		deps.HTTPHandlers = nil
	}))
	err := scheduler.Init()
	Expect(err).To(BeNil())

	// prepare mocks
	mockSB := test.NewMockSouthbound()
	// -> descriptor1 (notifications):
	descriptor1 := test.NewMockDescriptor(&test.MockDescriptorArgs{
		Name:             descriptor1Name,
		KeySelector:      func(key string) bool {
			if !strings.HasPrefix(key, prefixA) {
				return false
			}
			if strings.Contains(strings.TrimPrefix(key, prefixA), "/") {
				return false // exclude derived values
			}
			return true
		},
		NBKeyPrefixes:    []string{prefixA},
		ValueBuilder:     test.ArrayValueBuilder(prefixA),
		DerValuesBuilder: test.ArrayValueDerBuilder(Property),
		WithMetadata:     true,
		DumpIsSupported:  false,
	}, mockSB, 0)
	// -> descriptor2:
	descriptor2 := test.NewMockDescriptor(&test.MockDescriptorArgs{
		Name:            descriptor2Name,
		KeySelector:     prefixSelector(prefixB),
		NBKeyPrefixes:   []string{prefixB},
		ValueBuilder:    test.ArrayValueBuilder(prefixB),
		DependencyBuilder: func(key string, value Value) []Dependency {
			if key == prefixB + baseValue2 {
				depKey := prefixA + baseValue1
				return []Dependency{
					{Label: depKey, Key: depKey},
				}
			}
			if key == prefixB + baseValue2 + "/item2" {
				depKey := prefixA + baseValue1 + "/item2"
				return []Dependency{
					{Label: depKey, Key: depKey},
				}
			}
			return nil
		},
		DerValuesBuilder: test.ArrayValueDerBuilder(Object),
		WithMetadata:     true,
		DumpIsSupported:  true,
		DumpDependencies: []string{descriptor1Name},
	}, mockSB, 0)

	// register both descriptors with the scheduler
	scheduler.RegisterKVDescriptor(descriptor1)
	scheduler.RegisterKVDescriptor(descriptor2)

	// get metadata map created for each descriptor
	metadataMap := scheduler.GetMetadataMap(descriptor1.GetName())
	nameToInteger1, withMetadataMap := metadataMap.(test.NameToInteger)
	Expect(withMetadataMap).To(BeTrue())
	metadataMap = scheduler.GetMetadataMap(descriptor2.GetName())
	nameToInteger2, withMetadataMap := metadataMap.(test.NameToInteger)
	Expect(withMetadataMap).To(BeTrue())

	// run resync transaction against empty SB
	startTime := time.Now()
	values := []KeyValueDataPair{
		{Key: prefixB + baseValue2, ValueData: []string{"item1", "item2"}},
	}
	kvErrors, txnError := scheduler.StartNBTransaction().Resync(values).Commit(context.Background())
	stopTime := time.Now()
	Expect(txnError).ShouldNot(HaveOccurred())
	Expect(kvErrors).To(BeEmpty())

	// check the state of SB
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	Expect(mockSB.GetValues(nil)).To(BeEmpty())

	// check metadata
	Expect(metadataMap.ListAllNames()).To(BeEmpty())

	// check operations executed in SB
	opHistory := mockSB.PopHistoryOfOps()
	Expect(opHistory).To(HaveLen(1))
	operation := opHistory[0]
	Expect(operation.OpType).To(Equal(test.Dump))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	checkValuesForCorrelation(operation.CorrelateDump, []KVWithMetadata{
		{
			Key: prefixB + baseValue2,
			Value: test.NewArrayValue(Object, baseValue2, "item1", "item2"),
			Metadata: nil,
			Origin: FromNB,
		},
	})

	// check transaction operations
	txnHistory := scheduler.getTransactionHistory(time.Time{}, time.Now())
	Expect(txnHistory).To(HaveLen(1))
	txn := txnHistory[0]
	Expect(txn.preRecord).To(BeFalse())
	Expect(txn.start.After(startTime)).To(BeTrue())
	Expect(txn.start.Before(txn.stop)).To(BeTrue())
	Expect(txn.stop.Before(stopTime)).To(BeTrue())
	Expect(txn.seqNum).To(BeEquivalentTo(0))
	Expect(txn.txnType).To(BeEquivalentTo(nbTransaction))
	Expect(txn.isResync).To(BeTrue())
	checkRecordedValues(txn.values, []recordedKVPair{
		{key: prefixB + baseValue2, value: &recordedValue{valueType: Object, label: baseValue2, string: "[item1,item2]"}, origin: FromNB},
	})
	Expect(txn.preErrors).To(BeEmpty())

	txnOps := recordedTxnOps{
		{
			operation:  add,
			key:        prefixB + baseValue2,
			newValue:   &recordedValue{valueType: Object, label: baseValue2, string: "[item1,item2]"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			isPending:  true,
		},
	}
	checkTxnOperations(txn.planned, txnOps)
	checkTxnOperations(txn.executed, txnOps)

	// check flag stats
	graphR := scheduler.graph.Read()
	errorStats := graphR.GetFlagStats(ErrorFlagName, nil)
	Expect(errorStats.TotalCount).To(BeEquivalentTo(0))
	pendingStats := graphR.GetFlagStats(PendingFlagName, nil)
	Expect(pendingStats.TotalCount).To(BeEquivalentTo(1))
	derivedStats := graphR.GetFlagStats(DerivedFlagName, nil)
	Expect(derivedStats.TotalCount).To(BeEquivalentTo(0))
	lastUpdateStats := graphR.GetFlagStats(LastUpdateFlagName, nil)
	Expect(lastUpdateStats.TotalCount).To(BeEquivalentTo(1))
	lastChangeStats := graphR.GetFlagStats(LastChangeFlagName, nil)
	Expect(lastChangeStats.TotalCount).To(BeEquivalentTo(1))
	descriptorStats := graphR.GetFlagStats(DescriptorFlagName, nil)
	Expect(descriptorStats.TotalCount).To(BeEquivalentTo(1))
	Expect(descriptorStats.PerValueCount).ToNot(HaveKey(descriptor1Name))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor2Name))
	Expect(descriptorStats.PerValueCount[descriptor2Name]).To(BeEquivalentTo(1))
	originStats := graphR.GetFlagStats(OriginFlagName, nil)
	Expect(originStats.TotalCount).To(BeEquivalentTo(1))
	Expect(originStats.PerValueCount).To(HaveKey(FromNB.String()))
	Expect(originStats.PerValueCount[FromNB.String()]).To(BeEquivalentTo(1))
	graphR.Release()

	// send notification
	startTime = time.Now()
	mockSB.SetValue(prefixA + baseValue1, test.NewArrayValue(Object, baseValue1, "item1"), &test.OnlyInteger{Integer:10}, FromSB, false)
	notifError := scheduler.PushSBNotification(prefixA + baseValue1, test.NewArrayValue(Object, baseValue1, "item1"),
		&test.OnlyInteger{Integer:10})
	Expect(notifError).ShouldNot(HaveOccurred())

	// wait until the notification is processed
	Eventually(func() []*KVWithMetadata {
		return mockSB.GetValues(nil)
	}, 2 * time.Second).Should(HaveLen(3))
	stopTime = time.Now()

	// check the state of SB
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	// -> base value 1
	value := mockSB.GetValue(prefixA + baseValue1)
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewArrayValue(Object, baseValue1, "item1"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(10))
	Expect(value.Origin).To(BeEquivalentTo(FromSB))
	// -> base value 2
	value = mockSB.GetValue(prefixB + baseValue2)
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewArrayValue(Object, baseValue2, "item1", "item2"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(0))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item1 derived from base value 2
	value = mockSB.GetValue(prefixB + baseValue2 + "/item1")
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewStringValue(Object, "item1", "item1"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item2 derived from base value 2 is pending
	value = mockSB.GetValue(prefixB + baseValue2 + "/item2")
	Expect(value).To(BeNil())

	// check metadata
	metadata, exists := nameToInteger1.LookupByName(baseValue1)
	Expect(exists).To(BeTrue())
	Expect(metadata.GetInteger()).To(BeEquivalentTo(10))
	metadata, exists = nameToInteger2.LookupByName(baseValue2)
	Expect(exists).To(BeTrue())
	Expect(metadata.GetInteger()).To(BeEquivalentTo(0))

	// check operations executed in SB
	opHistory = mockSB.PopHistoryOfOps()
	Expect(opHistory).To(HaveLen(2))
	operation = opHistory[0]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixB + baseValue2))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[1]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixB + baseValue2 + "/item1"))
	Expect(operation.Err).To(BeNil())

	// check transaction operations
	txnHistory = scheduler.getTransactionHistory(startTime, time.Now())
	Expect(txnHistory).To(HaveLen(1))
	txn = txnHistory[0]
	Expect(txn.preRecord).To(BeFalse())
	Expect(txn.start.After(startTime)).To(BeTrue())
	Expect(txn.start.Before(txn.stop)).To(BeTrue())
	Expect(txn.stop.Before(stopTime)).To(BeTrue())
	Expect(txn.seqNum).To(BeEquivalentTo(1))
	Expect(txn.txnType).To(BeEquivalentTo(sbNotification))
	Expect(txn.isResync).To(BeFalse())
	checkRecordedValues(txn.values, []recordedKVPair{
		{key: prefixA + baseValue1, value: &recordedValue{valueType: Object, label: baseValue1, string: "[item1]"}, origin: FromSB},
	})
	Expect(txn.preErrors).To(BeEmpty())

	txnOps = recordedTxnOps{
		{
			operation:  add,
			key:        prefixA + baseValue1,
			newValue:   &recordedValue{valueType: Object, label: baseValue1, string: "[item1]"},
			prevOrigin: FromSB,
			newOrigin:  FromSB,
		},
		{
			operation:  add,
			key:        prefixB + baseValue2,
			prevValue:  &recordedValue{valueType: Object, label: baseValue2, string: "[item1,item2]"},
			newValue:   &recordedValue{valueType: Object, label: baseValue2, string: "[item1,item2]"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			wasPending: true,
		},
		{
			operation:  add,
			key:        prefixB + baseValue2 + "/item1",
			newValue:   &recordedValue{valueType: Object, label: "item1", string: "item1"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  add,
			key:        prefixB + baseValue2 + "/item2",
			newValue:   &recordedValue{valueType: Object, label: "item2", string: "item2"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			isPending:  true,
		},
		{
			operation:  add,
			key:        prefixA + baseValue1 + "/item1",
			newValue:   &recordedValue{valueType: Property, label: "item1", string: "item1"},
			prevOrigin: FromSB,
			newOrigin:  FromSB,
		},
	}
	checkTxnOperations(txn.planned, txnOps)
	checkTxnOperations(txn.executed, txnOps)

	// check flag stats
	graphR = scheduler.graph.Read()
	errorStats = graphR.GetFlagStats(ErrorFlagName, nil)
	Expect(errorStats.TotalCount).To(BeEquivalentTo(0))
	pendingStats = graphR.GetFlagStats(PendingFlagName, nil)
	Expect(pendingStats.TotalCount).To(BeEquivalentTo(2))
	derivedStats = graphR.GetFlagStats(DerivedFlagName, nil)
	Expect(derivedStats.TotalCount).To(BeEquivalentTo(3))
	lastUpdateStats = graphR.GetFlagStats(LastUpdateFlagName, nil)
	Expect(lastUpdateStats.TotalCount).To(BeEquivalentTo(6))
	lastChangeStats = graphR.GetFlagStats(LastChangeFlagName, nil)
	Expect(lastChangeStats.TotalCount).To(BeEquivalentTo(3))
	descriptorStats = graphR.GetFlagStats(DescriptorFlagName, nil)
	Expect(descriptorStats.TotalCount).To(BeEquivalentTo(5))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor1Name))
	Expect(descriptorStats.PerValueCount[descriptor1Name]).To(BeEquivalentTo(1))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor2Name))
	Expect(descriptorStats.PerValueCount[descriptor2Name]).To(BeEquivalentTo(4))
	originStats = graphR.GetFlagStats(OriginFlagName, nil)
	Expect(originStats.TotalCount).To(BeEquivalentTo(6))
	Expect(originStats.PerValueCount).To(HaveKey(FromNB.String()))
	Expect(originStats.PerValueCount[FromNB.String()]).To(BeEquivalentTo(4))
	Expect(originStats.PerValueCount).To(HaveKey(FromSB.String()))
	Expect(originStats.PerValueCount[FromSB.String()]).To(BeEquivalentTo(2))
	graphR.Release()

	// send 2nd notification
	startTime = time.Now()
	mockSB.SetValue(prefixA + baseValue1, test.NewArrayValue(Object, baseValue1, "item1", "item2"), &test.OnlyInteger{Integer:11}, FromSB, false)
	notifError = scheduler.PushSBNotification(prefixA + baseValue1, test.NewArrayValue(Object, baseValue1, "item1", "item2"),
		&test.OnlyInteger{Integer:11})
	Expect(notifError).ShouldNot(HaveOccurred())

	// wait until the notification is processed
	Eventually(func() []*KVWithMetadata {
		return mockSB.GetValues(nil)
	}, 2 * time.Second).Should(HaveLen(4))
	stopTime = time.Now()

	// check the state of SB
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	// -> base value 1
	value = mockSB.GetValue(prefixA + baseValue1)
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewArrayValue(Object, baseValue1, "item1", "item2"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(11))
	Expect(value.Origin).To(BeEquivalentTo(FromSB))
	// -> base value 2
	value = mockSB.GetValue(prefixB + baseValue2)
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewArrayValue(Object, baseValue2, "item1", "item2"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(0))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item1 derived from base value 2
	value = mockSB.GetValue(prefixB + baseValue2 + "/item1")
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewStringValue(Object, "item1", "item1"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item2 derived from base value 2
	value = mockSB.GetValue(prefixB + baseValue2 + "/item2")
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewStringValue(Object, "item2", "item2"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))

	// check metadata
	metadata, exists = nameToInteger1.LookupByName(baseValue1)
	Expect(exists).To(BeTrue())
	Expect(metadata.GetInteger()).To(BeEquivalentTo(11))
	metadata, exists = nameToInteger2.LookupByName(baseValue2)
	Expect(exists).To(BeTrue())
	Expect(metadata.GetInteger()).To(BeEquivalentTo(0))

	// check operations executed in SB
	opHistory = mockSB.PopHistoryOfOps()
	Expect(opHistory).To(HaveLen(2))
	operation = opHistory[0]
	Expect(operation.OpType).To(Equal(test.Update))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixB + baseValue2))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[1]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixB + baseValue2 + "/item2"))
	Expect(operation.Err).To(BeNil())

	// check transaction operations
	txnHistory = scheduler.getTransactionHistory(startTime, time.Now())
	Expect(txnHistory).To(HaveLen(1))
	txn = txnHistory[0]
	Expect(txn.preRecord).To(BeFalse())
	Expect(txn.start.After(startTime)).To(BeTrue())
	Expect(txn.start.Before(txn.stop)).To(BeTrue())
	Expect(txn.stop.Before(stopTime)).To(BeTrue())
	Expect(txn.seqNum).To(BeEquivalentTo(2))
	Expect(txn.txnType).To(BeEquivalentTo(sbNotification))
	Expect(txn.isResync).To(BeFalse())
	checkRecordedValues(txn.values, []recordedKVPair{
		{key: prefixA + baseValue1, value: &recordedValue{valueType: Object, label: baseValue1, string: "[item1,item2]"}, origin: FromSB},
	})
	Expect(txn.preErrors).To(BeEmpty())

	txnOps = recordedTxnOps{
		{
			operation:  modify,
			key:        prefixA + baseValue1,
			prevValue:  &recordedValue{valueType: Object, label: baseValue1, string: "[item1]"},
			newValue:   &recordedValue{valueType: Object, label: baseValue1, string: "[item1,item2]"},
			prevOrigin: FromSB,
			newOrigin:  FromSB,
		},
		{
			operation:  update,
			key:        prefixB + baseValue2,
			prevValue:  &recordedValue{valueType: Object, label: baseValue2, string: "[item1,item2]"},
			newValue:   &recordedValue{valueType: Object, label: baseValue2, string: "[item1,item2]"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  add,
			key:        prefixA + baseValue1 + "/item2",
			newValue:   &recordedValue{valueType: Property, label: "item2", string: "item2"},
			prevOrigin: FromSB,
			newOrigin:  FromSB,
		},
		{
			operation:  add,
			key:        prefixB + baseValue2 + "/item2",
			prevValue:  &recordedValue{valueType: Object, label: "item2", string: "item2"},
			newValue:   &recordedValue{valueType: Object, label: "item2", string: "item2"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			wasPending: true,
		},

	}
	checkTxnOperations(txn.planned, txnOps)
	checkTxnOperations(txn.executed, txnOps)

	// check flag stats
	graphR = scheduler.graph.Read()
	errorStats = graphR.GetFlagStats(ErrorFlagName, nil)
	Expect(errorStats.TotalCount).To(BeEquivalentTo(0))
	pendingStats = graphR.GetFlagStats(PendingFlagName, nil)
	Expect(pendingStats.TotalCount).To(BeEquivalentTo(2))
	derivedStats = graphR.GetFlagStats(DerivedFlagName, nil)
	Expect(derivedStats.TotalCount).To(BeEquivalentTo(6))
	lastUpdateStats = graphR.GetFlagStats(LastUpdateFlagName, nil)
	Expect(lastUpdateStats.TotalCount).To(BeEquivalentTo(11))
	lastChangeStats = graphR.GetFlagStats(LastChangeFlagName, nil)
	Expect(lastChangeStats.TotalCount).To(BeEquivalentTo(5))
	descriptorStats = graphR.GetFlagStats(DescriptorFlagName, nil)
	Expect(descriptorStats.TotalCount).To(BeEquivalentTo(8))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor1Name))
	Expect(descriptorStats.PerValueCount[descriptor1Name]).To(BeEquivalentTo(2))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor2Name))
	Expect(descriptorStats.PerValueCount[descriptor2Name]).To(BeEquivalentTo(6))
	originStats = graphR.GetFlagStats(OriginFlagName, nil)
	Expect(originStats.TotalCount).To(BeEquivalentTo(11))
	Expect(originStats.PerValueCount).To(HaveKey(FromNB.String()))
	Expect(originStats.PerValueCount[FromNB.String()]).To(BeEquivalentTo(6))
	Expect(originStats.PerValueCount).To(HaveKey(FromSB.String()))
	Expect(originStats.PerValueCount[FromSB.String()]).To(BeEquivalentTo(5))
	graphR.Release()

	// send 3rd notification
	startTime = time.Now()
	mockSB.SetValue(prefixA + baseValue1, nil, nil, FromSB, false)
	notifError = scheduler.PushSBNotification(prefixA + baseValue1, nil, nil)
	Expect(notifError).ShouldNot(HaveOccurred())

	// wait until the notification is processed
	Eventually(func() []*KVWithMetadata {
		return mockSB.GetValues(nil)
	}, 2 * time.Second).Should(HaveLen(0))
	stopTime = time.Now()

	// check the state of SB
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	Expect(mockSB.GetValues(nil)).To(BeEmpty())

	// check metadata
	Expect(metadataMap.ListAllNames()).To(BeEmpty())

	// check operations executed in SB
	opHistory = mockSB.PopHistoryOfOps()
	Expect(opHistory).To(HaveLen(3))
	operation = opHistory[0]
	Expect(operation.OpType).To(Equal(test.Delete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixB + baseValue2 + "/item2"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[1]
	Expect(operation.OpType).To(Equal(test.Delete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixB + baseValue2 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[2]
	Expect(operation.OpType).To(Equal(test.Delete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixB + baseValue2))
	Expect(operation.Err).To(BeNil())

	// check transaction operations
	txnHistory = scheduler.getTransactionHistory(startTime, time.Now())
	Expect(txnHistory).To(HaveLen(1))
	txn = txnHistory[0]
	Expect(txn.preRecord).To(BeFalse())
	Expect(txn.start.After(startTime)).To(BeTrue())
	Expect(txn.start.Before(txn.stop)).To(BeTrue())
	Expect(txn.stop.Before(stopTime)).To(BeTrue())
	Expect(txn.seqNum).To(BeEquivalentTo(3))
	Expect(txn.txnType).To(BeEquivalentTo(sbNotification))
	Expect(txn.isResync).To(BeFalse())
	checkRecordedValues(txn.values, []recordedKVPair{
		{key: prefixA + baseValue1, value: nil, origin: FromSB},
	})
	Expect(txn.preErrors).To(BeEmpty())

	txnOps = recordedTxnOps{
		{
			operation:  del,
			key:        prefixA + baseValue1 + "/item1",
			prevValue:  &recordedValue{valueType: Property, label: "item1", string: "item1"},
			prevOrigin: FromSB,
			newOrigin:  FromSB,
		},
		{
			operation:  del,
			key:        prefixB + baseValue2 + "/item2",
			prevValue:  &recordedValue{valueType: Object, label: "item2", string: "item2"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			isPending:  true,
		},
		{
			operation:  del,
			key:        prefixA + baseValue1 + "/item2",
			prevValue:  &recordedValue{valueType: Property, label: "item2", string: "item2"},
			prevOrigin: FromSB,
			newOrigin:  FromSB,
		},
		{
			operation:  del,
			key:        prefixB + baseValue2 + "/item1",
			prevValue:  &recordedValue{valueType: Object, label: "item1", string: "item1"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  del,
			key:        prefixB + baseValue2 + "/item2",
			prevValue:  &recordedValue{valueType: Object, label: "item2", string: "item2"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			wasPending: true,
		},
		{
			operation:  del,
			key:        prefixB + baseValue2,
			prevValue:  &recordedValue{valueType: Object, label: baseValue2, string: "[item1,item2]"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			isPending:  true,
		},
		{
			operation:  del,
			key:        prefixA + baseValue1,
			prevValue:  &recordedValue{valueType: Object, label: baseValue1, string: "[item1,item2]"},
			prevOrigin: FromSB,
			newOrigin:  FromSB,
		},
	}
	checkTxnOperations(txn.planned, txnOps)
	checkTxnOperations(txn.executed, txnOps)

	// close scheduler
	err = scheduler.Close()
	Expect(err).To(BeNil())
}

func TestNotificationsWithRetry(t *testing.T) {
	RegisterTestingT(t)

	// prepare KV Scheduler
	scheduler := NewPlugin(UseDeps(func(deps *Deps) {
		deps.HTTPHandlers = nil
	}))
	err := scheduler.Init()
	Expect(err).To(BeNil())

	// prepare mocks
	mockSB := test.NewMockSouthbound()
	// -> descriptor1 (notifications):
	descriptor1 := test.NewMockDescriptor(&test.MockDescriptorArgs{
		Name:             descriptor1Name,
		KeySelector:      prefixSelector(prefixA),
		NBKeyPrefixes:    []string{prefixA},
		ValueBuilder:     test.ArrayValueBuilder(prefixA),
		DerValuesBuilder: test.ArrayValueDerBuilder(Action),
		WithMetadata:     true,
		DumpIsSupported:  false,
	}, mockSB, 0)
	// -> descriptor2:
	descriptor2 := test.NewMockDescriptor(&test.MockDescriptorArgs{
		Name:            descriptor2Name,
		KeySelector:     prefixSelector(prefixB),
		NBKeyPrefixes:   []string{prefixB},
		ValueBuilder:    test.ArrayValueBuilder(prefixB),
		DependencyBuilder: func(key string, value Value) []Dependency {
			if key == prefixB + baseValue2 {
				depKey := prefixA + baseValue1
				return []Dependency{
					{Label: depKey, Key: depKey},
				}
			}
			if key == prefixB + baseValue2 + "/item2" {
				depKey := prefixA + baseValue1 + "/item2"
				return []Dependency{
					{Label: depKey, Key: depKey},
				}
			}
			return nil
		},
		DerValuesBuilder: test.ArrayValueDerBuilder(Action),
		WithMetadata:     true,
		DumpIsSupported:  true,
	}, mockSB, 0)
	// -> descriptor3:
	descriptor3 := test.NewMockDescriptor(&test.MockDescriptorArgs{
		Name:            descriptor3Name,
		KeySelector:     prefixSelector(prefixC),
		NBKeyPrefixes:   []string{prefixC},
		ValueBuilder:    test.StringValueBuilder(prefixC),
		DependencyBuilder: func(key string, value Value) []Dependency {
			if key == prefixC + baseValue3 {
				return []Dependency{
					{Label: prefixA, AnyOf: prefixSelector(prefixA)},
				}
			}
			return nil
		},
		WithMetadata:     true,
		DumpIsSupported:  true,
		DumpDependencies: []string{descriptor2Name},
	}, mockSB, 0)

	// -> planned errors
	mockSB.PlanError(prefixB + baseValue2 + "/item2", errors.New("failed to add derived value"),
		func() {
			mockSB.SetValue(prefixB + baseValue2, test.NewArrayValue(Object, baseValue2, "item1"),
				&test.OnlyInteger{Integer:0}, FromNB, false)
		})
	for i := 0; i < 3; i++ {
		mockSB.PlanError(prefixC+baseValue3, errors.New("failed to add value"),
			func() {
				mockSB.SetValue(prefixC+baseValue3, nil, nil, FromNB, false)
			})
	}

	// subscribe to receive notifications about errors
	errorChan := make(chan KeyWithError, 5)
	scheduler.SubscribeForErrors(errorChan, nil)

	// register all 3 descriptors with the scheduler
	scheduler.RegisterKVDescriptor(descriptor1)
	scheduler.RegisterKVDescriptor(descriptor2)
	scheduler.RegisterKVDescriptor(descriptor3)

	// get metadata map created for each descriptor
	metadataMap := scheduler.GetMetadataMap(descriptor1.GetName())
	nameToInteger1, withMetadataMap := metadataMap.(test.NameToInteger)
	Expect(withMetadataMap).To(BeTrue())
	metadataMap = scheduler.GetMetadataMap(descriptor2.GetName())
	nameToInteger2, withMetadataMap := metadataMap.(test.NameToInteger)
	Expect(withMetadataMap).To(BeTrue())
	metadataMap = scheduler.GetMetadataMap(descriptor3.GetName())
	nameToInteger3, withMetadataMap := metadataMap.(test.NameToInteger)
	Expect(withMetadataMap).To(BeTrue())

	// run 1st data-change transaction with retry against empty SB
	schedulerTxn1 := scheduler.StartNBTransaction(WithRetry(3 * time.Second, true))
	schedulerTxn1.SetValueData(prefixB + baseValue2, []string{"item1", "item2"})
	kvErrors, txnError := schedulerTxn1.Commit(context.Background())
	Expect(txnError).ShouldNot(HaveOccurred())
	Expect(kvErrors).To(BeEmpty())

	// run 2nd data-change transaction with retry
	schedulerTxn2 := scheduler.StartNBTransaction(WithRetry(6 * time.Second, true))
	schedulerTxn2.SetValueData(prefixC + baseValue3, "base-value3-data")
	kvErrors, txnError = schedulerTxn2.Commit(context.Background())
	Expect(txnError).ShouldNot(HaveOccurred())
	Expect(kvErrors).To(BeEmpty())

	// check the state of SB - empty since dependencies are not met
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	Expect(mockSB.GetValues(nil)).To(BeEmpty())
	Expect(mockSB.PopHistoryOfOps()).To(HaveLen(0))

	// check metadata
	Expect(metadataMap.ListAllNames()).To(BeEmpty())

	// send notification
	startTime := time.Now()
	notifError := scheduler.PushSBNotification(prefixA + baseValue1, test.NewArrayValue(Object, baseValue1, "item1", "item2"),
		&test.OnlyInteger{Integer:10})
	Expect(notifError).ShouldNot(HaveOccurred())

	// wait until the notification is processed
	Eventually(func() []*KVWithMetadata {
		return mockSB.GetValues(nil)
	}, 2 * time.Second).Should(HaveLen(2))
	stopTime := time.Now()

	// receive the error notifications
	var errorNotif KeyWithError
	Eventually(errorChan, time.Second).Should(Receive(&errorNotif))
	Expect(errorNotif.Key).To(Equal(prefixC + baseValue3))
	Expect(errorNotif.Error).ToNot(BeNil())
	Expect(errorNotif.Error.Error()).To(BeEquivalentTo("failed to add value"))
	Eventually(errorChan, time.Second).Should(Receive(&errorNotif))
	Expect(errorNotif.Key).To(Equal(prefixC + baseValue3))
	Expect(errorNotif.Error).ToNot(BeNil())
	Expect(errorNotif.Error.Error()).To(BeEquivalentTo("failed to add value"))
	Eventually(errorChan, time.Second).Should(Receive(&errorNotif))
	Expect(errorNotif.Key).To(Equal(prefixB + baseValue2 + "/item2"))
	Expect(errorNotif.Error).ToNot(BeNil())
	Expect(errorNotif.Error.Error()).To(BeEquivalentTo("failed to add derived value"))
	Eventually(errorChan, time.Second).Should(Receive(&errorNotif))
	Expect(errorNotif.Key).To(Equal(prefixC + baseValue3))
	Expect(errorNotif.Error).ToNot(BeNil())
	Expect(errorNotif.Error.Error()).To(BeEquivalentTo("failed to add value"))

	// check the state of SB
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	// -> base value 2
	value := mockSB.GetValue(prefixB + baseValue2)
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewArrayValue(Object, baseValue2, "item1"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(0))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item1 derived from base value 2
	value = mockSB.GetValue(prefixB + baseValue2 + "/item1")
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewStringValue(Action, "item1", "item1"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item2 derived from base value 2 failed to get created
	value = mockSB.GetValue(prefixB + baseValue2 + "/item2")
	Expect(value).To(BeNil())
	// -> base value 3 failed to get created
	value = mockSB.GetValue(prefixC + baseValue3)
	Expect(value).To(BeNil())
	Expect(mockSB.GetValues(nil)).To(HaveLen(2))

	// check failed values
	failedVals := scheduler.GetFailedValues(nil)
	Expect(failedVals).To(HaveLen(2))
	Expect(failedVals).To(ContainElement(KeyWithError{Key: prefixB + baseValue2 + "/item2", Error: errors.New("failed to add derived value")}))
	Expect(failedVals).To(ContainElement(KeyWithError{Key: prefixC + baseValue3, Error: errors.New("failed to add value")}))

	// check metadata
	metadata, exists := nameToInteger1.LookupByName(baseValue1)
	Expect(exists).To(BeTrue())
	Expect(metadata.GetInteger()).To(BeEquivalentTo(10))
	metadata, exists = nameToInteger2.LookupByName(baseValue2)
	Expect(exists).To(BeTrue())
	Expect(metadata.GetInteger()).To(BeEquivalentTo(0))
	metadata, exists = nameToInteger3.LookupByName(baseValue3)
	Expect(exists).To(BeFalse())

	// check operations executed in SB
	opHistory := mockSB.PopHistoryOfOps()
	Expect(opHistory).To(HaveLen(8))
	operation := opHistory[0]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixB + baseValue2))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[1]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixB + baseValue2 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[2]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3))
	Expect(operation.Err).ToNot(BeNil())
	Expect(operation.Err.Error()).To(BeEquivalentTo("failed to add value"))
	operation = opHistory[3]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3))
	Expect(operation.Err).ToNot(BeNil())
	Expect(operation.Err.Error()).To(BeEquivalentTo("failed to add value"))
	operation = opHistory[4]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixB + baseValue2 + "/item2"))
	Expect(operation.Err).ToNot(BeNil())
	Expect(operation.Err.Error()).To(BeEquivalentTo("failed to add derived value"))
	operation = opHistory[5]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3))
	Expect(operation.Err).ToNot(BeNil())
	Expect(operation.Err.Error()).To(BeEquivalentTo("failed to add value"))
	operation = opHistory[6] // refresh failed value
	Expect(operation.OpType).To(Equal(test.Dump))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	checkValuesForCorrelation(operation.CorrelateDump, []KVWithMetadata{
		{
			Key: prefixB + baseValue2,
			Value: test.NewArrayValue(Object, baseValue2, "item1", "item2"),
			Metadata: &test.OnlyInteger{Integer:0},
			Origin: FromNB,
		},
	})
	operation = opHistory[7] // refresh failed value
	Expect(operation.OpType).To(Equal(test.Dump))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	checkValuesForCorrelation(operation.CorrelateDump, []KVWithMetadata{})

	// check last transaction
	txnHistory := scheduler.getTransactionHistory(time.Time{}, time.Now())
	Expect(txnHistory).To(HaveLen(3))
	txn := txnHistory[2]
	Expect(txn.preRecord).To(BeFalse())
	Expect(txn.start.After(startTime)).To(BeTrue())
	Expect(txn.start.Before(txn.stop)).To(BeTrue())
	Expect(txn.stop.Before(stopTime)).To(BeTrue())
	Expect(txn.seqNum).To(BeEquivalentTo(2))
	Expect(txn.txnType).To(BeEquivalentTo(sbNotification))
	Expect(txn.isResync).To(BeFalse())
	checkRecordedValues(txn.values, []recordedKVPair{
		{key: prefixA + baseValue1, value: &recordedValue{valueType: Object, label: baseValue1, string: "[item1,item2]"}, origin: FromSB},
	})
	Expect(txn.preErrors).To(BeEmpty())

	// -> planned operations
	txnOps := recordedTxnOps{
		{
			operation:  add,
			key:        prefixA + baseValue1,
			newValue:   &recordedValue{valueType: Object, label: baseValue1, string: "[item1,item2]"},
			prevOrigin: FromSB,
			newOrigin:  FromSB,
		},
		{
			operation:  add,
			key:        prefixB + baseValue2,
			prevValue:  &recordedValue{valueType: Object, label: baseValue2, string: "[item1,item2]"},
			newValue:   &recordedValue{valueType: Object, label: baseValue2, string: "[item1,item2]"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			wasPending: true,
		},
		{
			operation:  add,
			key:        prefixB + baseValue2 + "/item1",
			newValue:   &recordedValue{valueType: Action, label: "item1", string: "item1"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  add,
			key:        prefixB + baseValue2 + "/item2",
			newValue:   &recordedValue{valueType: Action, label: "item2", string: "item2"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			isPending:  true,
		},
		{
			operation:  add,
			key:        prefixC + baseValue3,
			prevValue:  &recordedValue{valueType: Object, label: baseValue3, string: "base-value3-data"},
			newValue:   &recordedValue{valueType: Object, label: baseValue3, string: "base-value3-data"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			wasPending: true,
		},
		{
			operation:  add,
			key:        prefixA + baseValue1 + "/item1",
			newValue:   &recordedValue{valueType: Action, label: "item1", string: "item1"},
			prevOrigin: FromSB,
			newOrigin:  FromSB,
		},
		{
			operation:  update,
			key:        prefixC + baseValue3,
			prevValue:  &recordedValue{valueType: Object, label: baseValue3, string: "base-value3-data"},
			newValue:   &recordedValue{valueType: Object, label: baseValue3, string: "base-value3-data"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  add,
			key:        prefixA + baseValue1 + "/item2",
			newValue:   &recordedValue{valueType: Action, label: "item2", string: "item2"},
			prevOrigin: FromSB,
			newOrigin:  FromSB,
		},
		{
			operation:  add,
			key:        prefixB + baseValue2 + "/item2",
			prevValue:  &recordedValue{valueType: Action, label: "item2", string: "item2"},
			newValue:   &recordedValue{valueType: Action, label: "item2", string: "item2"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			wasPending: true,
		},
		{
			operation:  update,
			key:        prefixC + baseValue3,
			prevValue:  &recordedValue{valueType: Object, label: baseValue3, string: "base-value3-data"},
			newValue:   &recordedValue{valueType: Object, label: baseValue3, string: "base-value3-data"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
	}
	checkTxnOperations(txn.planned, txnOps)

	// -> executed operations
	txnOps = recordedTxnOps{
		{
			operation:  add,
			key:        prefixA + baseValue1,
			newValue:   &recordedValue{valueType: Object, label: baseValue1, string: "[item1,item2]"},
			prevOrigin: FromSB,
			newOrigin:  FromSB,
		},
		{
			operation:  add,
			key:        prefixB + baseValue2,
			prevValue:  &recordedValue{valueType: Object, label: baseValue2, string: "[item1,item2]"},
			newValue:   &recordedValue{valueType: Object, label: baseValue2, string: "[item1,item2]"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			wasPending: true,
		},
		{
			operation:  add,
			key:        prefixB + baseValue2 + "/item1",
			newValue:   &recordedValue{valueType: Action, label: "item1", string: "item1"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  add,
			key:        prefixB + baseValue2 + "/item2",
			newValue:   &recordedValue{valueType: Action, label: "item2", string: "item2"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			isPending:  true,
		},
		{
			operation:  add,
			key:        prefixC + baseValue3,
			prevValue:  &recordedValue{valueType: Object, label: baseValue3, string: "base-value3-data"},
			newValue:   &recordedValue{valueType: Object, label: baseValue3, string: "base-value3-data"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			wasPending: true,
			isPending:  true,
			newErr:     errors.New("failed to add value"),
		},
		{
			operation:  add,
			key:        prefixA + baseValue1 + "/item1",
			newValue:   &recordedValue{valueType: Action, label: "item1", string: "item1"},
			prevOrigin: FromSB,
			newOrigin:  FromSB,
		},
		{
			operation:  add,
			key:        prefixC + baseValue3,
			prevValue:  &recordedValue{valueType: Object, label: baseValue3, string: "base-value3-data"},
			newValue:   &recordedValue{valueType: Object, label: baseValue3, string: "base-value3-data"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			wasPending: true,
			isPending:  true,
			prevErr:    errors.New("failed to add value"),
			newErr:     errors.New("failed to add value"),
		},
		{
			operation:  add,
			key:        prefixA + baseValue1 + "/item2",
			newValue:   &recordedValue{valueType: Action, label: "item2", string: "item2"},
			prevOrigin: FromSB,
			newOrigin:  FromSB,
		},
		{
			operation:  add,
			key:        prefixB + baseValue2 + "/item2",
			prevValue:  &recordedValue{valueType: Action, label: "item2", string: "item2"},
			newValue:   &recordedValue{valueType: Action, label: "item2", string: "item2"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			wasPending: true,
			isPending:  true,
			newErr:     errors.New("failed to add derived value"),
		},
		{
			operation:  add,
			key:        prefixC + baseValue3,
			prevValue:  &recordedValue{valueType: Object, label: baseValue3, string: "base-value3-data"},
			newValue:   &recordedValue{valueType: Object, label: baseValue3, string: "base-value3-data"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			wasPending: true,
			isPending:  true,
			prevErr:    errors.New("failed to add value"),
			newErr:     errors.New("failed to add value"),
		},
	}
	checkTxnOperations(txn.executed, txnOps)

	// check flag stats
	graphR := scheduler.graph.Read()
	errorStats := graphR.GetFlagStats(ErrorFlagName, nil)
	Expect(errorStats.TotalCount).To(BeEquivalentTo(2))
	pendingStats := graphR.GetFlagStats(PendingFlagName, nil)
	Expect(pendingStats.TotalCount).To(BeEquivalentTo(4))
	derivedStats := graphR.GetFlagStats(DerivedFlagName, nil)
	Expect(derivedStats.TotalCount).To(BeEquivalentTo(4))
	lastUpdateStats := graphR.GetFlagStats(LastUpdateFlagName, nil)
	Expect(lastUpdateStats.TotalCount).To(BeEquivalentTo(9))
	lastChangeStats := graphR.GetFlagStats(LastChangeFlagName, nil)
	Expect(lastChangeStats.TotalCount).To(BeEquivalentTo(5))
	descriptorStats := graphR.GetFlagStats(DescriptorFlagName, nil)
	Expect(descriptorStats.TotalCount).To(BeEquivalentTo(9))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor1Name))
	Expect(descriptorStats.PerValueCount[descriptor1Name]).To(BeEquivalentTo(3))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor2Name))
	Expect(descriptorStats.PerValueCount[descriptor2Name]).To(BeEquivalentTo(4))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor3Name))
	Expect(descriptorStats.PerValueCount[descriptor3Name]).To(BeEquivalentTo(2))
	originStats := graphR.GetFlagStats(OriginFlagName, nil)
	Expect(originStats.TotalCount).To(BeEquivalentTo(9))
	Expect(originStats.PerValueCount).To(HaveKey(FromNB.String()))
	Expect(originStats.PerValueCount[FromNB.String()]).To(BeEquivalentTo(6))
	Expect(originStats.PerValueCount).To(HaveKey(FromSB.String()))
	Expect(originStats.PerValueCount[FromSB.String()]).To(BeEquivalentTo(3))
	graphR.Release()

	// item2 derived from baseValue2 should get fixed first
	startTime = time.Now()
	Eventually(errorChan, 5*time.Second).Should(Receive(&errorNotif))
	Expect(errorNotif.Key).To(Equal(prefixB + baseValue2 + "/item2"))
	Expect(errorNotif.Error).To(BeNil())
	stopTime = time.Now()

	// check the state of SB
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	// -> item2 derived from base value 2 is now created
	value = mockSB.GetValue(prefixB + baseValue2 + "/item2")
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewStringValue(Action, "item2", "item2"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	Expect(mockSB.GetValues(nil)).To(HaveLen(3))

	// check failed values
	failedVals = scheduler.GetFailedValues(nil)
	Expect(failedVals).To(HaveLen(1))
	Expect(failedVals).To(ContainElement(KeyWithError{Key: prefixC + baseValue3, Error: errors.New("failed to add value")}))

	// check operations executed in SB
	opHistory = mockSB.PopHistoryOfOps()
	Expect(opHistory).To(HaveLen(2))
	operation = opHistory[0]
	Expect(operation.OpType).To(Equal(test.Modify))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixB + baseValue2))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[1]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixB + baseValue2 + "/item2"))
	Expect(operation.Err).To(BeNil())

	// check last transaction
	txnHistory = scheduler.getTransactionHistory(startTime, time.Now())
	Expect(txnHistory).To(HaveLen(1))
	txn = txnHistory[0]
	Expect(txn.preRecord).To(BeFalse())
	Expect(txn.start.After(startTime)).To(BeTrue())
	Expect(txn.start.Before(txn.stop)).To(BeTrue())
	Expect(txn.stop.Before(stopTime)).To(BeTrue())
	Expect(txn.seqNum).To(BeEquivalentTo(3))
	Expect(txn.txnType).To(BeEquivalentTo(retryFailedOps))
	Expect(txn.isResync).To(BeFalse())
	checkRecordedValues(txn.values, []recordedKVPair{
		{key: prefixB + baseValue2, value: &recordedValue{valueType: Object, label: baseValue2, string: "[item1,item2]"}, origin: FromNB},
	})
	Expect(txn.preErrors).To(BeEmpty())
	txnOps = recordedTxnOps{
		{
			operation:  modify,
			key:        prefixB + baseValue2,
			prevValue:  &recordedValue{valueType: Object, label: baseValue2, string: "[item1]"},
			newValue:   &recordedValue{valueType: Object, label: baseValue2, string: "[item1,item2]"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			isRetry:    true,
		},
		{
			operation:  add,
			key:        prefixB + baseValue2 + "/item2",
			prevValue:  &recordedValue{valueType: Action, label: "item2", string: "item2"},
			newValue:   &recordedValue{valueType: Action, label: "item2", string: "item2"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			wasPending: true,
			prevErr:    errors.New("failed to add derived value"),
			isRetry:    true,
		},
	}
	checkTxnOperations(txn.planned, txnOps)
	checkTxnOperations(txn.executed, txnOps)

	// base-value3 should get fixed eventually as well
	startTime = time.Now()
	Eventually(errorChan, 5*time.Second).Should(Receive(&errorNotif))
	Expect(errorNotif.Key).To(Equal(prefixC + baseValue3))
	Expect(errorNotif.Error).To(BeNil())
	stopTime = time.Now()

	// check the state of SB
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	// -> base value 3 is now created
	value = mockSB.GetValue(prefixC + baseValue3)
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewStringValue(Object, baseValue3, "base-value3-data"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(0))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	Expect(mockSB.GetValues(nil)).To(HaveLen(4))

	// check failed values
	failedVals = scheduler.GetFailedValues(nil)
	Expect(failedVals).To(HaveLen(0))

	// check operations executed in SB
	opHistory = mockSB.PopHistoryOfOps()
	Expect(opHistory).To(HaveLen(1))
	operation = opHistory[0]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3))
	Expect(operation.Err).To(BeNil())

	// check last transaction
	txnHistory = scheduler.getTransactionHistory(startTime, time.Time{})
	Expect(txnHistory).To(HaveLen(1))
	txn = txnHistory[0]
	Expect(txn.preRecord).To(BeFalse())
	Expect(txn.start.After(startTime)).To(BeTrue())
	Expect(txn.start.Before(txn.stop)).To(BeTrue())
	Expect(txn.stop.Before(stopTime)).To(BeTrue())
	Expect(txn.seqNum).To(BeEquivalentTo(4))
	Expect(txn.txnType).To(BeEquivalentTo(retryFailedOps))
	Expect(txn.isResync).To(BeFalse())
	checkRecordedValues(txn.values, []recordedKVPair{
		{key: prefixC + baseValue3, value: &recordedValue{valueType: Object, label: baseValue3, string: "base-value3-data"}, origin: FromNB},
	})
	Expect(txn.preErrors).To(BeEmpty())
	txnOps = recordedTxnOps{
		{
			operation:  add,
			key:        prefixC + baseValue3,
			prevValue:  &recordedValue{valueType: Object, label: baseValue3, string: "base-value3-data"},
			newValue:   &recordedValue{valueType: Object, label: baseValue3, string: "base-value3-data"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			wasPending: true,
			prevErr:    errors.New("failed to add value"),
			isRetry:    true,
		},
	}
	checkTxnOperations(txn.planned, txnOps)
	checkTxnOperations(txn.executed, txnOps)

	// check metadata
	metadata, exists = nameToInteger1.LookupByName(baseValue1)
	Expect(exists).To(BeTrue())
	Expect(metadata.GetInteger()).To(BeEquivalentTo(10))
	metadata, exists = nameToInteger2.LookupByName(baseValue2)
	Expect(exists).To(BeTrue())
	Expect(metadata.GetInteger()).To(BeEquivalentTo(0))
	metadata, exists = nameToInteger3.LookupByName(baseValue3)
	Expect(exists).To(BeTrue())
	Expect(metadata.GetInteger()).To(BeEquivalentTo(0))

	// close scheduler
	err = scheduler.Close()
	Expect(err).To(BeNil())
}