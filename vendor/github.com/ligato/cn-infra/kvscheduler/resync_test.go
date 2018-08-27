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
)

func TestEmptyResync(t *testing.T) {
	RegisterTestingT(t)

	// prepare KV Scheduler
	scheduler := NewPlugin(UseDeps(func(deps *Deps) {
		deps.HTTPHandlers = nil
	}))
	err := scheduler.Init()
	Expect(err).To(BeNil())

	// prepare mocks
	mockSB := test.NewMockSouthbound()
	descriptor1 := test.NewMockDescriptor(&test.MockDescriptorArgs{
		Name:            descriptor1Name,
		KeySelector:     prefixSelector(prefixA),
		NBKeyPrefixes:   []string{prefixA},
		WithMetadata:    true,
		DumpIsSupported: true,
	}, mockSB, 0)

	// register descriptor with the scheduler
	scheduler.RegisterKVDescriptor(descriptor1)
	nbPrefixes := scheduler.GetRegisteredNBKeyPrefixes()
	Expect(nbPrefixes).To(HaveLen(1))
	Expect(nbPrefixes).To(ContainElement(prefixA))

	// get metadata map created for the descriptor
	metadataMap := scheduler.GetMetadataMap(descriptor1.GetName())
	_, withMetadataMap := metadataMap.(test.NameToInteger)
	Expect(withMetadataMap).To(BeTrue())

	// transaction history should be initially empty
	Expect(scheduler.getTransactionHistory(time.Time{}, time.Time{})).To(BeEmpty())

	// run transaction with empty resync
	startTime := time.Now()
	kvErrors, txnError := scheduler.StartNBTransaction().Resync([]KeyValueDataPair{}).Commit(context.Background())
	stopTime := time.Now()
	Expect(txnError).ShouldNot(HaveOccurred())
	Expect(kvErrors).To(BeEmpty())

	// check the state of SB
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	Expect(mockSB.GetValues(nil)).To(BeEmpty())

	// check metadata
	Expect(metadataMap.ListAllNames()).To(BeEmpty())

	// check executed operations
	opHistory := mockSB.PopHistoryOfOps()
	Expect(opHistory).To(HaveLen(1))
	Expect(opHistory[0].OpType).To(Equal(test.Dump))
	Expect(opHistory[0].CorrelateDump).To(BeEmpty())
	Expect(opHistory[0].Descriptor).To(BeEquivalentTo(descriptor1Name))

	// single transaction consisted of zero operations
	txnHistory := scheduler.getTransactionHistory(time.Time{}, time.Time{})
	Expect(txnHistory).To(HaveLen(1))
	txn := txnHistory[0]
	Expect(txn.preRecord).To(BeFalse())
	Expect(txn.start.After(startTime)).To(BeTrue())
	Expect(txn.start.Before(txn.stop)).To(BeTrue())
	Expect(txn.stop.Before(stopTime)).To(BeTrue())
	Expect(txn.seqNum).To(BeEquivalentTo(0))
	Expect(txn.txnType).To(BeEquivalentTo(nbTransaction))
	Expect(txn.isResync).To(BeTrue())
	Expect(txn.values).To(BeEmpty())
	Expect(txn.preErrors).To(BeEmpty())
	Expect(txn.planned).To(BeEmpty())
	Expect(txn.executed).To(BeEmpty())

	// check flag stats
	graphR := scheduler.graph.Read()
	errorStats := graphR.GetFlagStats(ErrorFlagName, nil)
	Expect(errorStats.TotalCount).To(BeEquivalentTo(0))
	pendingStats := graphR.GetFlagStats(PendingFlagName, nil)
	Expect(pendingStats.TotalCount).To(BeEquivalentTo(0))
	derivedStats := graphR.GetFlagStats(DerivedFlagName, nil)
	Expect(derivedStats.TotalCount).To(BeEquivalentTo(0))
	lastUpdateStats := graphR.GetFlagStats(LastUpdateFlagName, nil)
	Expect(lastUpdateStats.TotalCount).To(BeEquivalentTo(0))
	lastChangeStats := graphR.GetFlagStats(LastChangeFlagName, nil)
	Expect(lastChangeStats.TotalCount).To(BeEquivalentTo(0))
	descriptorStats := graphR.GetFlagStats(DescriptorFlagName, nil)
	Expect(descriptorStats.TotalCount).To(BeEquivalentTo(0))
	originStats := graphR.GetFlagStats(OriginFlagName, nil)
	Expect(originStats.TotalCount).To(BeEquivalentTo(0))
	graphR.Release()

	// close scheduler
	err = scheduler.Close()
	Expect(err).To(BeNil())
}

func TestResyncWithEmptySB(t *testing.T) {
	RegisterTestingT(t)

	// prepare KV Scheduler
	scheduler := NewPlugin(UseDeps(func(deps *Deps) {
		deps.HTTPHandlers = nil
	}))
	err := scheduler.Init()
	Expect(err).To(BeNil())

	// prepare mocks
	mockSB := test.NewMockSouthbound()
	descriptor1 := test.NewMockDescriptor(&test.MockDescriptorArgs{
		Name:            descriptor1Name,
		KeySelector:     prefixSelector(prefixA),
		NBKeyPrefixes:   []string{prefixA},
		ValueBuilder:    test.ArrayValueBuilder(prefixA),
		DependencyBuilder: func(key string, value Value) []Dependency {
			if key == prefixA + baseValue2 {
				depKey := prefixA + baseValue1 + "/item1" // base value depends on a derived value
				return []Dependency{
					{Label: depKey, Key: depKey},
				}
			}
			if key == prefixA + baseValue1 + "/item2" {
				depKey := prefixA + baseValue2 + "/item1" // derived value depends on another derived value
				return []Dependency{
					{Label: depKey, Key: depKey},
				}
			}
			return nil
		},
		DerValuesBuilder: test.ArrayValueDerBuilder(Object),
		WithMetadata:    true,
		DumpIsSupported: true,
	}, mockSB, 0)

	// register descriptor with the scheduler
	scheduler.RegisterKVDescriptor(descriptor1)

	// get metadata map created for the descriptor
	metadataMap := scheduler.GetMetadataMap(descriptor1.GetName())
	nameToInteger, withMetadataMap := metadataMap.(test.NameToInteger)
	Expect(withMetadataMap).To(BeTrue())

	// run resync transaction with empty SB
	startTime := time.Now()
	values := []KeyValueDataPair{
		{Key: prefixA + baseValue2, ValueData: []string{"item1"}},
		{Key: prefixA + baseValue1, ValueData: []string{"item1", "item2"}},
	}
	kvErrors, txnError := scheduler.StartNBTransaction().Resync(values).Commit(context.Background())
	stopTime := time.Now()
	Expect(txnError).ShouldNot(HaveOccurred())
	Expect(kvErrors).To(BeEmpty())

	// check the state of SB
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	// -> base value 1
	value := mockSB.GetValue(prefixA + baseValue1)
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewArrayValue(Object, baseValue1, "item1", "item2"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(0))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item1 derived from base value 1
	value = mockSB.GetValue(prefixA + baseValue1 + "/item1")
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewStringValue(Object, "item1", "item1"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item2 derived from base value 1
	value = mockSB.GetValue(prefixA + baseValue1 + "/item2")
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewStringValue(Object, "item2", "item2"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> base value 2
	value = mockSB.GetValue(prefixA + baseValue2)
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewArrayValue(Object, baseValue2, "item1"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(1))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item1 derived from base value 2
	value = mockSB.GetValue(prefixA + baseValue2 + "/item1")
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewStringValue(Object, "item1", "item1"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	Expect(mockSB.GetValues(nil)).To(HaveLen(5))

	// check scheduler API
	prefixAValues := scheduler.GetValues(prefixSelector(prefixA))
	checkValues(prefixAValues, []KeyValuePair{
		{Key: prefixA + baseValue1, Value: test.NewArrayValue(Object, baseValue1, "item1", "item2")},
		{Key: prefixA + baseValue1 + "/item1", Value: test.NewStringValue(Object, "item1", "item1")},
		{Key: prefixA + baseValue1 + "/item2", Value: test.NewStringValue(Object, "item2", "item2")},
		{Key: prefixA + baseValue2, Value: test.NewArrayValue(Object, baseValue2, "item1")},
		{Key: prefixA + baseValue2 + "/item1", Value: test.NewStringValue(Object, "item1", "item1")},
	})
	Expect(scheduler.GetValue(prefixA + baseValue1).Equivalent(test.NewArrayValue(Object, baseValue1, "item1", "item2"))).To(BeTrue())
	Expect(scheduler.GetValue(prefixA + baseValue1 + "/item1").Equivalent(test.NewStringValue(Object, "item1", "item1"))).To(BeTrue())
	Expect(scheduler.GetFailedValues(nil)).To(BeEmpty())
	Expect(scheduler.GetPendingValues(nil)).To(BeEmpty())

	// check metadata
	metadata, exists := nameToInteger.LookupByName(baseValue1)
	Expect(exists).To(BeTrue())
	Expect(metadata.GetInteger()).To(BeEquivalentTo(0))
	metadata, exists = nameToInteger.LookupByName(baseValue2)
	Expect(exists).To(BeTrue())
	Expect(metadata.GetInteger()).To(BeEquivalentTo(1))

	// check executed operations
	opHistory := mockSB.PopHistoryOfOps()
	Expect(opHistory).To(HaveLen(6))
	operation := opHistory[0]
	Expect(operation.OpType).To(Equal(test.Dump))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	checkValuesForCorrelation(operation.CorrelateDump, []KVWithMetadata{
		{
			Key: prefixA + baseValue1,
			Value: test.NewArrayValue(Object, baseValue1, "item1", "item2"),
			Metadata: nil,
			Origin: FromNB,
		},
		{
			Key: prefixA + baseValue2,
			Value: test.NewArrayValue(Object, baseValue2, "item1"),
			Metadata: nil,
			Origin: FromNB,
		},
	})
	operation = opHistory[1]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[2]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[3]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue2))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[4]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue2 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[5]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item2"))
	Expect(operation.Err).To(BeNil())

	// single transaction consisted of 6 operations
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
		{key: prefixA + baseValue1, value: &recordedValue{valueType: Object, label: baseValue1, string: "[item1,item2]"}, origin: FromNB},
		{key: prefixA + baseValue2, value: &recordedValue{valueType: Object, label: baseValue2, string: "[item1]"}, origin: FromNB},
	})
	Expect(txn.preErrors).To(BeEmpty())

	txnOps := recordedTxnOps{
		{
			operation:  add,
			key:        prefixA + baseValue1,
			newValue:   &recordedValue{valueType: Object, label: baseValue1, string: "[item1,item2]"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  add,
			key:        prefixA + baseValue1 + "/item1",
			newValue:   &recordedValue{valueType: Object, label: "item1", string: "item1"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  add,
			key:        prefixA + baseValue1 + "/item2",
			newValue:   &recordedValue{valueType: Object, label: "item2", string: "item2"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			isPending:  true,
		},
		{
			operation:  add,
			key:        prefixA + baseValue2,
			newValue:   &recordedValue{valueType: Object, label: baseValue2, string: "[item1]"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  add,
			key:        prefixA + baseValue2 + "/item1",
			newValue:   &recordedValue{valueType: Object, label: "item1", string: "item1"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  add,
			key:        prefixA + baseValue1 + "/item2",
			prevValue:  &recordedValue{valueType: Object, label: "item2", string: "item2"},
			newValue:   &recordedValue{valueType: Object, label: "item2", string: "item2"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			wasPending: true,
		},
	}
	checkTxnOperations(txn.planned, txnOps)
	checkTxnOperations(txn.executed, txnOps)

	// now remove everything using resync with empty data
	startTime = time.Now()
	kvErrors, txnError = scheduler.StartNBTransaction().Resync([]KeyValueDataPair{}).Commit(context.Background())
	stopTime = time.Now()
	Expect(txnError).ShouldNot(HaveOccurred())
	Expect(kvErrors).To(BeEmpty())

	// check the state of SB
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	Expect(mockSB.GetValues(nil)).To(BeEmpty())

	// check metadata
	Expect(metadataMap.ListAllNames()).To(BeEmpty())

	// check executed operations
	opHistory = mockSB.PopHistoryOfOps()
	Expect(opHistory).To(HaveLen(6))
	operation = opHistory[0]
	Expect(operation.OpType).To(Equal(test.Dump))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	checkValuesForCorrelation(operation.CorrelateDump, []KVWithMetadata{
		{
			Key: prefixA + baseValue1,
			Value: test.NewArrayValue(Object, baseValue1, "item1", "item2"),
			Metadata: &test.OnlyInteger{Integer: 0},
			Origin: FromNB,
		},
		{
			Key: prefixA + baseValue2,
			Value: test.NewArrayValue(Object, baseValue2, "item1"),
			Metadata: &test.OnlyInteger{Integer: 1},
			Origin: FromNB,
		},
	})
	operation = opHistory[1]
	Expect(operation.OpType).To(Equal(test.Delete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item2"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[2]
	Expect(operation.OpType).To(Equal(test.Delete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue2 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[3]
	Expect(operation.OpType).To(Equal(test.Delete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue2))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[4]
	Expect(operation.OpType).To(Equal(test.Delete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[5]
	Expect(operation.OpType).To(Equal(test.Delete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1))
	Expect(operation.Err).To(BeNil())

	// this second transaction consisted of 7 operations
	txnHistory = scheduler.getTransactionHistory(time.Time{}, time.Now())
	Expect(txnHistory).To(HaveLen(2))
	txn = txnHistory[1]
	Expect(txn.preRecord).To(BeFalse())
	Expect(txn.start.After(startTime)).To(BeTrue())
	Expect(txn.start.Before(txn.stop)).To(BeTrue())
	Expect(txn.stop.Before(stopTime)).To(BeTrue())
	Expect(txn.seqNum).To(BeEquivalentTo(1))
	Expect(txn.txnType).To(BeEquivalentTo(nbTransaction))
	Expect(txn.isResync).To(BeTrue())
	checkRecordedValues(txn.values, []recordedKVPair{
		{key: prefixA + baseValue1, value: nil, origin: FromNB},
		{key: prefixA + baseValue2, value: nil, origin: FromNB},
	})
	Expect(txn.preErrors).To(BeEmpty())

	txnOps = recordedTxnOps{
		{
			operation:  del,
			key:        prefixA + baseValue1 + "/item2",
			prevValue:  &recordedValue{valueType: Object, label: "item2", string: "item2"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			isPending:  true,
		},
		{
			operation:  del,
			key:        prefixA + baseValue2 + "/item1",
			prevValue:  &recordedValue{valueType: Object, label: "item1", string: "item1"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  del,
			key:        prefixA + baseValue2,
			prevValue:  &recordedValue{valueType: Object, label: baseValue2, string: "[item1]"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			isPending:  true,
		},
		{
			operation:  del,
			key:        prefixA + baseValue1 + "/item1",
			prevValue:  &recordedValue{valueType: Object, label: "item1", string: "item1"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  del,
			key:        prefixA + baseValue1 + "/item2",
			prevValue:  &recordedValue{valueType: Object, label: "item2", string: "item2"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			wasPending: true,
		},
		{
			operation:  del,
			key:        prefixA + baseValue1,
			prevValue:  &recordedValue{valueType: Object, label: baseValue1, string: "[item1,item2]"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  del,
			key:        prefixA + baseValue2,
			prevValue:  &recordedValue{valueType: Object, label: baseValue2, string: "[item1]"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			wasPending: true,
		},
	}
	checkTxnOperations(txn.planned, txnOps)
	checkTxnOperations(txn.executed, txnOps)

	// check flag stats
	graphR := scheduler.graph.Read()
	errorStats := graphR.GetFlagStats(ErrorFlagName, nil)
	Expect(errorStats.TotalCount).To(BeEquivalentTo(0))
	pendingStats := graphR.GetFlagStats(PendingFlagName, nil)
	Expect(pendingStats.TotalCount).To(BeEquivalentTo(0))
	derivedStats := graphR.GetFlagStats(DerivedFlagName, nil)
	Expect(derivedStats.TotalCount).To(BeEquivalentTo(3))
	lastUpdateStats := graphR.GetFlagStats(LastUpdateFlagName, nil)
	Expect(lastUpdateStats.TotalCount).To(BeEquivalentTo(5))
	lastChangeStats := graphR.GetFlagStats(LastChangeFlagName, nil)
	Expect(lastChangeStats.TotalCount).To(BeEquivalentTo(2))
	descriptorStats := graphR.GetFlagStats(DescriptorFlagName, nil)
	Expect(descriptorStats.TotalCount).To(BeEquivalentTo(5))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor1Name))
	Expect(descriptorStats.PerValueCount[descriptor1Name]).To(BeEquivalentTo(5))
	originStats := graphR.GetFlagStats(OriginFlagName, nil)
	Expect(originStats.TotalCount).To(BeEquivalentTo(5))
	Expect(originStats.PerValueCount).To(HaveKey(FromNB.String()))
	Expect(originStats.PerValueCount[FromNB.String()]).To(BeEquivalentTo(5))
	graphR.Release()

	// close scheduler
	err = scheduler.Close()
	Expect(err).To(BeNil())
}

func TestResyncWithNonEmptySB(t *testing.T) {
	RegisterTestingT(t)

	// prepare KV Scheduler
	scheduler := NewPlugin(UseDeps(func(deps *Deps) {
		deps.HTTPHandlers = nil
	}))
	err := scheduler.Init()
	Expect(err).To(BeNil())

	// prepare mocks
	mockSB := test.NewMockSouthbound()
	// -> initial content:
	mockSB.SetValue(prefixA + baseValue1, test.NewArrayValue(Object, baseValue1, "item1"),
		&test.OnlyInteger{Integer: 0}, FromNB, false)
	mockSB.SetValue(prefixA + baseValue1 + "/item1", test.NewStringValue(Object, "item1", "item1"),
		nil, FromNB, true)
	mockSB.SetValue(prefixA + baseValue2, test.NewArrayValue(Object, baseValue2, "item1"),
		&test.OnlyInteger{Integer:1}, FromNB, false)
	mockSB.SetValue(prefixA + baseValue2 + "/item1", test.NewStringValue(Object, "item1", "item1"),
		nil, FromNB, true)
	mockSB.SetValue(prefixA + baseValue3, test.NewArrayValue(Object, baseValue3, "item1"),
		&test.OnlyInteger{Integer:2}, FromNB, false)
	mockSB.SetValue(prefixA + baseValue3 + "/item1", test.NewStringValue(Object, "item1", "item1"),
		nil, FromNB, true)
	// -> descriptor1:
	descriptor1 := test.NewMockDescriptor(&test.MockDescriptorArgs{
		Name:            descriptor1Name,
		KeySelector:     prefixSelector(prefixA),
		NBKeyPrefixes:   []string{prefixA},
		ValueBuilder:    test.ArrayValueBuilder(prefixA),
		DependencyBuilder: func(key string, value Value) []Dependency {
			if key == prefixA + baseValue2 + "/item1" {
				depKey := prefixA + baseValue1
				return []Dependency{
					{Label: depKey, Key: depKey},
				}
			}
			if key == prefixA + baseValue2 + "/item2" {
				depKey := prefixA + baseValue1 + "/item1"
				return []Dependency{
					{Label: depKey, Key: depKey},
				}
			}
			return nil
		},
		DerValuesBuilder: test.ArrayValueDerBuilder(Object),
		RecreateChecker: func(key string, oldValue, newValue Value, metadata Metadata) bool {
			if key == prefixA + baseValue3 {
				return true
			}
			return false
		},
		WithMetadata:    true,
		DumpIsSupported: true,
	}, mockSB, 3)

	// register descriptor with the scheduler
	scheduler.RegisterKVDescriptor(descriptor1)

	// get metadata map created for the descriptor
	metadataMap := scheduler.GetMetadataMap(descriptor1.GetName())
	nameToInteger, withMetadataMap := metadataMap.(test.NameToInteger)
	Expect(withMetadataMap).To(BeTrue())

	// run resync transaction with SB that already has some values added
	startTime := time.Now()
	values := []KeyValueDataPair{
		{Key: prefixA + baseValue2, ValueData: []string{"item1", "item2"}},
		{Key: prefixA + baseValue1, ValueData: []string{"item2"}},
		{Key: prefixA + baseValue3, ValueData: []string{"item1", "item2"}},
	}
	kvErrors, txnError := scheduler.StartNBTransaction().Resync(values).Commit(context.Background())
	stopTime := time.Now()
	Expect(txnError).ShouldNot(HaveOccurred())
	Expect(kvErrors).To(BeEmpty())

	// check the state of SB
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	// -> base value 1
	value := mockSB.GetValue(prefixA + baseValue1)
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewArrayValue(Object, baseValue1, "item2"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(0))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item1 derived from base value 1 was removed
	value = mockSB.GetValue(prefixA + baseValue1 + "/item1")
	Expect(value).To(BeNil())
	// -> item2 derived from base value 1
	value = mockSB.GetValue(prefixA + baseValue1 + "/item2")
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewStringValue(Object, "item2", "item2"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> base value 2
	value = mockSB.GetValue(prefixA + baseValue2)
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewArrayValue(Object, baseValue2, "item1", "item2"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(1))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item1 derived from base value 2
	value = mockSB.GetValue(prefixA + baseValue2 + "/item1")
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewStringValue(Object, "item1", "item1"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item2 derived from base value 2 is pending
	value = mockSB.GetValue(prefixA + baseValue2 + "/item2")
	Expect(value).To(BeNil())
	// -> base value 3
	value = mockSB.GetValue(prefixA + baseValue3)
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewArrayValue(Object, baseValue3, "item1", "item2"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(3))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item1 derived from base value 3
	value = mockSB.GetValue(prefixA + baseValue3 + "/item1")
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewStringValue(Object, "item1", "item1"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item2 derived from base value 3
	value = mockSB.GetValue(prefixA + baseValue3 + "/item2")
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewStringValue(Object, "item2", "item2"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	Expect(mockSB.GetValues(nil)).To(HaveLen(7))

	// check metadata
	metadata, exists := nameToInteger.LookupByName(baseValue1)
	Expect(exists).To(BeTrue())
	Expect(metadata.GetInteger()).To(BeEquivalentTo(0))
	metadata, exists = nameToInteger.LookupByName(baseValue2)
	Expect(exists).To(BeTrue())
	Expect(metadata.GetInteger()).To(BeEquivalentTo(1))
	metadata, exists = nameToInteger.LookupByName(baseValue3)
	Expect(exists).To(BeTrue())
	Expect(metadata.GetInteger()).To(BeEquivalentTo(3))

	// check operations executed in SB
	opHistory := mockSB.PopHistoryOfOps()
	Expect(opHistory).To(HaveLen(11))
	operation := opHistory[0]
	Expect(operation.OpType).To(Equal(test.Dump))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	checkValuesForCorrelation(operation.CorrelateDump, []KVWithMetadata{
		{
			Key: prefixA + baseValue1,
			Value: test.NewArrayValue(Object, baseValue1, "item2"),
			Metadata: nil,
			Origin: FromNB,
		},
		{
			Key: prefixA + baseValue2,
			Value: test.NewArrayValue(Object, baseValue2, "item1", "item2"),
			Metadata: nil,
			Origin: FromNB,
		},
		{
			Key: prefixA + baseValue3,
			Value: test.NewArrayValue(Object, baseValue3, "item1", "item2"),
			Metadata: nil,
			Origin: FromNB,
		},
	})
	operation = opHistory[1]
	Expect(operation.OpType).To(Equal(test.Delete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue3 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[2]
	Expect(operation.OpType).To(Equal(test.Delete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue3))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[3]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue3))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[4]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue3 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[5]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue3 + "/item2"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[6]
	Expect(operation.OpType).To(Equal(test.Delete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[7]
	Expect(operation.OpType).To(Equal(test.Modify))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[8]
	Expect(operation.OpType).To(Equal(test.Update))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue2 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[9]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item2"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[10]
	Expect(operation.OpType).To(Equal(test.Modify))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue2))
	Expect(operation.Err).To(BeNil())

	// check transaction operations
	txnHistory := scheduler.getTransactionHistory(time.Time{}, time.Time{})
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
		{key: prefixA + baseValue1, value: &recordedValue{valueType: Object, label: baseValue1, string: "[item2]"}, origin: FromNB},
		{key: prefixA + baseValue2, value: &recordedValue{valueType: Object, label: baseValue2, string: "[item1,item2]"}, origin: FromNB},
		{key: prefixA + baseValue3, value: &recordedValue{valueType: Object, label: baseValue3, string: "[item1,item2]"}, origin: FromNB},
	})
	Expect(txn.preErrors).To(BeEmpty())

	txnOps := recordedTxnOps{
		{
			operation:  del,
			key:        prefixA + baseValue3 + "/item1",
			prevValue:  &recordedValue{valueType: Object, label: "item1", string: "item1"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  del,
			key:        prefixA + baseValue3,
			prevValue:  &recordedValue{valueType: Object, label: baseValue3, string: "[item1]"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			isPending:  true,
		},
		{
			operation:  add,
			key:        prefixA + baseValue3,
			newValue:   &recordedValue{valueType: Object, label: baseValue3, string: "[item1,item2]"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			wasPending: true,
		},
		{
			operation:  add,
			key:        prefixA + baseValue3 + "/item1",
			newValue:   &recordedValue{valueType: Object, label: "item1", string: "item1"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  add,
			key:        prefixA + baseValue3 + "/item2",
			newValue:   &recordedValue{valueType: Object, label: "item2", string: "item2"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  del,
			key:        prefixA + baseValue1 + "/item1",
			prevValue:  &recordedValue{valueType: Object, label: "item1", string: "item1"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  modify,
			key:        prefixA + baseValue1,
			prevValue:  &recordedValue{valueType: Object, label: baseValue1, string: "[item1]"},
			newValue:   &recordedValue{valueType: Object, label: baseValue1, string: "[item2]"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  update,
			key:        prefixA + baseValue2 + "/item1",
			prevValue:  &recordedValue{valueType: Object, label: "item1", string: "item1"},
			newValue:   &recordedValue{valueType: Object, label: "item1", string: "item1"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  add,
			key:        prefixA + baseValue1 + "/item2",
			newValue:   &recordedValue{valueType: Object, label: "item2", string: "item2"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  modify,
			key:        prefixA + baseValue2,
			prevValue:  &recordedValue{valueType: Object, label: baseValue2, string: "[item1]"},
			newValue:   &recordedValue{valueType: Object, label: baseValue2, string: "[item1,item2]"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  add,
			key:        prefixA + baseValue2 + "/item2",
			newValue:   &recordedValue{valueType: Object, label: "item2", string: "item2"},
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
	Expect(derivedStats.TotalCount).To(BeEquivalentTo(5))
	lastUpdateStats := graphR.GetFlagStats(LastUpdateFlagName, nil)
	Expect(lastUpdateStats.TotalCount).To(BeEquivalentTo(8))
	lastChangeStats := graphR.GetFlagStats(LastChangeFlagName, nil)
	Expect(lastChangeStats.TotalCount).To(BeEquivalentTo(3))
	descriptorStats := graphR.GetFlagStats(DescriptorFlagName, nil)
	Expect(descriptorStats.TotalCount).To(BeEquivalentTo(8))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor1Name))
	Expect(descriptorStats.PerValueCount[descriptor1Name]).To(BeEquivalentTo(8))
	originStats := graphR.GetFlagStats(OriginFlagName, nil)
	Expect(originStats.TotalCount).To(BeEquivalentTo(8))
	Expect(originStats.PerValueCount).To(HaveKey(FromNB.String()))
	Expect(originStats.PerValueCount[FromNB.String()]).To(BeEquivalentTo(8))
	graphR.Release()

	// close scheduler
	err = scheduler.Close()
	Expect(err).To(BeNil())
}

func TestResyncNotRemovingSBValues(t *testing.T) {
	RegisterTestingT(t)

	// prepare KV Scheduler
	scheduler := NewPlugin(UseDeps(func(deps *Deps) {
		deps.HTTPHandlers = nil
	}))
	err := scheduler.Init()
	Expect(err).To(BeNil())

	// prepare mocks
	mockSB := test.NewMockSouthbound()
	// -> initial content:
	mockSB.SetValue(prefixA + baseValue1, test.NewStringValue(Action, baseValue1, baseValue1),
		nil, FromSB, false)
	// -> descriptor1:
	descriptor1 := test.NewMockDescriptor(&test.MockDescriptorArgs{
		Name:            descriptor1Name,
		KeySelector:     prefixSelector(prefixA),
		NBKeyPrefixes:   []string{prefixA},
		ValueBuilder:    test.ArrayValueBuilder(prefixA),
		DependencyBuilder: func(key string, value Value) []Dependency {
			if key == prefixA + baseValue2  {
				depKey := prefixA + baseValue1
				return []Dependency{
					{Label: depKey, Key: depKey},
				}
			}
			return nil
		},
		DerValuesBuilder: test.ArrayValueDerBuilder(Object),
		WithMetadata:    true,
		DumpIsSupported: true,
	}, mockSB, 0)

	// register descriptor with the scheduler
	scheduler.RegisterKVDescriptor(descriptor1)

	// get metadata map created for the descriptor
	metadataMap := scheduler.GetMetadataMap(descriptor1.GetName())
	nameToInteger, withMetadataMap := metadataMap.(test.NameToInteger)
	Expect(withMetadataMap).To(BeTrue())

	// run resync transaction that should keep values not managed by NB untouched
	startTime := time.Now()
	values := []KeyValueDataPair{
		{Key: prefixA + baseValue2, ValueData: []string{"item1"}},
	}
	kvErrors, txnError := scheduler.StartNBTransaction().Resync(values).Commit(context.Background())
	stopTime := time.Now()
	Expect(txnError).ShouldNot(HaveOccurred())
	Expect(kvErrors).To(BeEmpty())

	// check the state of SB
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	// -> base value 1
	value := mockSB.GetValue(prefixA + baseValue1)
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewStringValue(Action, baseValue1, baseValue1))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromSB))
	// -> base value 2
	value = mockSB.GetValue(prefixA + baseValue2)
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewArrayValue(Object, baseValue2, "item1"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(0))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item1 derived from base value 2
	value = mockSB.GetValue(prefixA + baseValue2 + "/item1")
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewStringValue(Object, "item1", "item1"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	Expect(mockSB.GetValues(nil)).To(HaveLen(3))

	// check metadata
	metadata, exists := nameToInteger.LookupByName(baseValue1)
	Expect(exists).To(BeFalse())
	metadata, exists = nameToInteger.LookupByName(baseValue2)
	Expect(exists).To(BeTrue())
	Expect(metadata.GetInteger()).To(BeEquivalentTo(0))

	// check operations executed in SB
	opHistory := mockSB.PopHistoryOfOps()
	Expect(opHistory).To(HaveLen(3))
	operation := opHistory[0]
	Expect(operation.OpType).To(Equal(test.Dump))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	checkValuesForCorrelation(operation.CorrelateDump, []KVWithMetadata{
		{
			Key: prefixA + baseValue2,
			Value: test.NewArrayValue(Object, baseValue2, "item1"),
			Metadata: nil,
			Origin: FromNB,
		},
	})
	operation = opHistory[1]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue2))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[2]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue2 + "/item1"))
	Expect(operation.Err).To(BeNil())

	// check transaction operations
	txnHistory := scheduler.getTransactionHistory(startTime, time.Now())
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
		{key: prefixA + baseValue1, value: &recordedValue{valueType: Action, label: baseValue1, string: baseValue1}, origin: FromSB},
		{key: prefixA + baseValue2, value: &recordedValue{valueType: Object, label: baseValue2, string: "[item1]"}, origin: FromNB},
	})
	Expect(txn.preErrors).To(BeEmpty())

	txnOps := recordedTxnOps{
		{
			operation:  add,
			key:        prefixA + baseValue2,
			newValue:   &recordedValue{valueType: Object, label: baseValue2, string: "[item1]"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  add,
			key:        prefixA + baseValue2 + "/item1",
			newValue:   &recordedValue{valueType: Object, label: "item1", string: "item1"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
	}
	checkTxnOperations(txn.planned, txnOps)
	checkTxnOperations(txn.executed, txnOps)

	// check flag stats
	graphR := scheduler.graph.Read()
	errorStats := graphR.GetFlagStats(ErrorFlagName, nil)
	Expect(errorStats.TotalCount).To(BeEquivalentTo(0))
	pendingStats := graphR.GetFlagStats(PendingFlagName, nil)
	Expect(pendingStats.TotalCount).To(BeEquivalentTo(0))
	derivedStats := graphR.GetFlagStats(DerivedFlagName, nil)
	Expect(derivedStats.TotalCount).To(BeEquivalentTo(1))
	lastUpdateStats := graphR.GetFlagStats(LastUpdateFlagName, nil)
	Expect(lastUpdateStats.TotalCount).To(BeEquivalentTo(3))
	lastChangeStats := graphR.GetFlagStats(LastChangeFlagName, nil)
	Expect(lastChangeStats.TotalCount).To(BeEquivalentTo(2))
	descriptorStats := graphR.GetFlagStats(DescriptorFlagName, nil)
	Expect(descriptorStats.TotalCount).To(BeEquivalentTo(3))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor1Name))
	Expect(descriptorStats.PerValueCount[descriptor1Name]).To(BeEquivalentTo(3))
	originStats := graphR.GetFlagStats(OriginFlagName, nil)
	Expect(originStats.TotalCount).To(BeEquivalentTo(3))
	Expect(originStats.PerValueCount).To(HaveKey(FromNB.String()))
	Expect(originStats.PerValueCount[FromNB.String()]).To(BeEquivalentTo(2))
	Expect(originStats.PerValueCount).To(HaveKey(FromSB.String()))
	Expect(originStats.PerValueCount[FromSB.String()]).To(BeEquivalentTo(1))
	graphR.Release()

	// close scheduler
	err = scheduler.Close()
	Expect(err).To(BeNil())
}

func TestResyncWithMultipleDescriptors(t *testing.T) {
	RegisterTestingT(t)

	// prepare KV Scheduler
	scheduler := NewPlugin(UseDeps(func(deps *Deps) {
		deps.HTTPHandlers = nil
	}))
	err := scheduler.Init()
	Expect(err).To(BeNil())

	// prepare mocks
	mockSB := test.NewMockSouthbound()
	// -> initial content:
	mockSB.SetValue(prefixA + baseValue1, test.NewArrayValue(Object, baseValue1, "item1"),
		&test.OnlyInteger{Integer: 0}, FromNB, false)
	mockSB.SetValue(prefixA + baseValue1 + "/item1", test.NewStringValue(Object, "item1", "item1"),
		nil, FromNB, true)
	mockSB.SetValue(prefixB + baseValue2, test.NewArrayValue(Object, baseValue2, "item1"),
		&test.OnlyInteger{Integer:0}, FromNB, false)
	mockSB.SetValue(prefixB + baseValue2 + "/item1", test.NewStringValue(Object, "item1", "item1"),
		nil, FromNB, true)
	mockSB.SetValue(prefixC + baseValue3, test.NewArrayValue(Object, baseValue3, "item1"),
		&test.OnlyInteger{Integer:0}, FromNB, false)
	mockSB.SetValue(prefixC + baseValue3 + "/item1", test.NewStringValue(Object, "item1", "item1"),
		nil, FromNB, true)
	// -> descriptor1:
	descriptor1 := test.NewMockDescriptor(&test.MockDescriptorArgs{
		Name:             descriptor1Name,
		KeySelector:      prefixSelector(prefixA),
		NBKeyPrefixes:    []string{prefixA},
		ValueBuilder:     test.ArrayValueBuilder(prefixA),
		DerValuesBuilder: test.ArrayValueDerBuilder(Object),
		WithMetadata:     true,
		DumpIsSupported:  true,
	}, mockSB, 1)
	// -> descriptor2:
	descriptor2 := test.NewMockDescriptor(&test.MockDescriptorArgs{
		Name:            descriptor2Name,
		KeySelector:     prefixSelector(prefixB),
		NBKeyPrefixes:   []string{prefixB},
		ValueBuilder:    test.ArrayValueBuilder(prefixB),
		DependencyBuilder: func(key string, value Value) []Dependency {
			if key == prefixB + baseValue2 + "/item1" {
				depKey := prefixA + baseValue1
				return []Dependency{
					{Label: depKey, Key: depKey},
				}
			}
			if key == prefixB + baseValue2 + "/item2" {
				depKey := prefixA + baseValue1 + "/item1"
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
	}, mockSB, 1)
	// -> descriptor3:
	descriptor3 := test.NewMockDescriptor(&test.MockDescriptorArgs{
		Name:             descriptor3Name,
		KeySelector:      prefixSelector(prefixC),
		NBKeyPrefixes:    []string{prefixC},
		ValueBuilder:     test.ArrayValueBuilder(prefixC),
		DerValuesBuilder: test.ArrayValueDerBuilder(Object),
		RecreateChecker:  func(key string, oldValue, newValue Value, metadata Metadata) bool {
			if key == prefixC + baseValue3 {
				return true
			}
			return false
		},
		WithMetadata:     true,
		DumpIsSupported:  true,
		DumpDependencies: []string{descriptor2Name},
	}, mockSB, 1)

	// register all 3 descriptors with the scheduler
	scheduler.RegisterKVDescriptor(descriptor1)
	scheduler.RegisterKVDescriptor(descriptor2)
	scheduler.RegisterKVDescriptor(descriptor3)
	nbPrefixes := scheduler.GetRegisteredNBKeyPrefixes()
	Expect(nbPrefixes).To(HaveLen(3))
	Expect(nbPrefixes).To(ContainElement(prefixA))
	Expect(nbPrefixes).To(ContainElement(prefixB))
	Expect(nbPrefixes).To(ContainElement(prefixC))

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

	// run resync transaction with SB that already has some values added
	startTime := time.Now()
	values := []KeyValueDataPair{
		{Key: prefixB + baseValue2, ValueData: []string{"item1", "item2"}},
		{Key: prefixA + baseValue1, ValueData: []string{"item2"}},
		{Key: prefixC + baseValue3, ValueData: []string{"item1", "item2"}},
	}
	kvErrors, txnError := scheduler.StartNBTransaction().Resync(values).Commit(context.Background())
	stopTime := time.Now()
	Expect(txnError).ShouldNot(HaveOccurred())
	Expect(kvErrors).To(BeEmpty())

	// check the state of SB
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	// -> base value 1
	value := mockSB.GetValue(prefixA + baseValue1)
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewArrayValue(Object, baseValue1, "item2"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(0))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item1 derived from base value 1 was removed
	value = mockSB.GetValue(prefixA + baseValue1 + "/item1")
	Expect(value).To(BeNil())
	// -> item2 derived from base value 1
	value = mockSB.GetValue(prefixA + baseValue1 + "/item2")
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewStringValue(Object, "item2", "item2"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
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
	// -> base value 3
	value = mockSB.GetValue(prefixC + baseValue3)
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewArrayValue(Object, baseValue3, "item1", "item2"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(1))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item1 derived from base value 3
	value = mockSB.GetValue(prefixC + baseValue3 + "/item1")
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewStringValue(Object, "item1", "item1"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item2 derived from base value 3
	value = mockSB.GetValue(prefixC + baseValue3 + "/item2")
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewStringValue(Object, "item2", "item2"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	Expect(mockSB.GetValues(nil)).To(HaveLen(7))

	// check metadata
	metadata, exists := nameToInteger1.LookupByName(baseValue1)
	Expect(exists).To(BeTrue())
	Expect(metadata.GetInteger()).To(BeEquivalentTo(0))
	metadata, exists = nameToInteger2.LookupByName(baseValue2)
	Expect(exists).To(BeTrue())
	Expect(metadata.GetInteger()).To(BeEquivalentTo(0))
	metadata, exists = nameToInteger3.LookupByName(baseValue3)
	Expect(exists).To(BeTrue())
	Expect(metadata.GetInteger()).To(BeEquivalentTo(1))

	// check operations executed in SB
	opHistory := mockSB.PopHistoryOfOps()
	Expect(opHistory).To(HaveLen(13))
	operation := opHistory[0]
	Expect(operation.OpType).To(Equal(test.Dump))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	checkValuesForCorrelation(operation.CorrelateDump, []KVWithMetadata{
		{
			Key: prefixA + baseValue1,
			Value: test.NewArrayValue(Object, baseValue1, "item2"),
			Metadata: nil,
			Origin: FromNB,
		},
	})
	operation = opHistory[1]
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
	operation = opHistory[2]
	Expect(operation.OpType).To(Equal(test.Dump))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	checkValuesForCorrelation(operation.CorrelateDump, []KVWithMetadata{
		{
			Key: prefixC + baseValue3,
			Value: test.NewArrayValue(Object, baseValue3, "item1", "item2"),
			Metadata: nil,
			Origin: FromNB,
		},
	})
	operation = opHistory[3]
	Expect(operation.OpType).To(Equal(test.Delete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[4]
	Expect(operation.OpType).To(Equal(test.Delete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[5]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[6]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[7]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3 + "/item2"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[8]
	Expect(operation.OpType).To(Equal(test.Delete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[9]
	Expect(operation.OpType).To(Equal(test.Modify))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[10]
	Expect(operation.OpType).To(Equal(test.Update))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixB + baseValue2 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[11]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item2"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[12]
	Expect(operation.OpType).To(Equal(test.Modify))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixB + baseValue2))
	Expect(operation.Err).To(BeNil())

	// check transaction operations
	txnHistory := scheduler.getTransactionHistory(time.Time{}, time.Time{})
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
		{key: prefixA + baseValue1, value: &recordedValue{valueType: Object, label: baseValue1, string: "[item2]"}, origin: FromNB},
		{key: prefixB + baseValue2, value: &recordedValue{valueType: Object, label: baseValue2, string: "[item1,item2]"}, origin: FromNB},
		{key: prefixC + baseValue3, value: &recordedValue{valueType: Object, label: baseValue3, string: "[item1,item2]"}, origin: FromNB},
	})
	Expect(txn.preErrors).To(BeEmpty())

	txnOps := recordedTxnOps{
		{
			operation:  del,
			key:        prefixC + baseValue3 + "/item1",
			prevValue:  &recordedValue{valueType: Object, label: "item1", string: "item1"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  del,
			key:        prefixC + baseValue3,
			prevValue:  &recordedValue{valueType: Object, label: baseValue3, string: "[item1]"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			isPending:  true,
		},
		{
			operation:  add,
			key:        prefixC + baseValue3,
			newValue:   &recordedValue{valueType: Object, label: baseValue3, string: "[item1,item2]"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			wasPending: true,
		},
		{
			operation:  add,
			key:        prefixC + baseValue3 + "/item1",
			newValue:   &recordedValue{valueType: Object, label: "item1", string: "item1"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  add,
			key:        prefixC + baseValue3 + "/item2",
			newValue:   &recordedValue{valueType: Object, label: "item2", string: "item2"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  del,
			key:        prefixA + baseValue1 + "/item1",
			prevValue:  &recordedValue{valueType: Object, label: "item1", string: "item1"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  modify,
			key:        prefixA + baseValue1,
			prevValue:  &recordedValue{valueType: Object, label: baseValue1, string: "[item1]"},
			newValue:   &recordedValue{valueType: Object, label: baseValue1, string: "[item2]"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  update,
			key:        prefixB + baseValue2 + "/item1",
			prevValue:  &recordedValue{valueType: Object, label: "item1", string: "item1"},
			newValue:   &recordedValue{valueType: Object, label: "item1", string: "item1"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  add,
			key:        prefixA + baseValue1 + "/item2",
			newValue:   &recordedValue{valueType: Object, label: "item2", string: "item2"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  modify,
			key:        prefixB + baseValue2,
			prevValue:  &recordedValue{valueType: Object, label: baseValue2, string: "[item1]"},
			newValue:   &recordedValue{valueType: Object, label: baseValue2, string: "[item1,item2]"},
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
	Expect(derivedStats.TotalCount).To(BeEquivalentTo(5))
	lastUpdateStats := graphR.GetFlagStats(LastUpdateFlagName, nil)
	Expect(lastUpdateStats.TotalCount).To(BeEquivalentTo(8))
	lastChangeStats := graphR.GetFlagStats(LastChangeFlagName, nil)
	Expect(lastChangeStats.TotalCount).To(BeEquivalentTo(3))
	descriptorStats := graphR.GetFlagStats(DescriptorFlagName, nil)
	Expect(descriptorStats.TotalCount).To(BeEquivalentTo(8))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor1Name))
	Expect(descriptorStats.PerValueCount[descriptor1Name]).To(BeEquivalentTo(2))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor2Name))
	Expect(descriptorStats.PerValueCount[descriptor2Name]).To(BeEquivalentTo(3))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor3Name))
	Expect(descriptorStats.PerValueCount[descriptor3Name]).To(BeEquivalentTo(3))
	originStats := graphR.GetFlagStats(OriginFlagName, nil)
	Expect(originStats.TotalCount).To(BeEquivalentTo(8))
	Expect(originStats.PerValueCount).To(HaveKey(FromNB.String()))
	Expect(originStats.PerValueCount[FromNB.String()]).To(BeEquivalentTo(8))
	graphR.Release()

	// close scheduler
	err = scheduler.Close()
	Expect(err).To(BeNil())
}

func TestResyncWithRetry(t *testing.T) {
	RegisterTestingT(t)

	// prepare KV Scheduler
	scheduler := NewPlugin(UseDeps(func(deps *Deps) {
		deps.HTTPHandlers = nil
	}))
	err := scheduler.Init()
	Expect(err).To(BeNil())

	// prepare mocks
	mockSB := test.NewMockSouthbound()
	// -> initial content:
	mockSB.SetValue(prefixA + baseValue1, test.NewArrayValue(Object, baseValue1),
		&test.OnlyInteger{Integer:0}, FromNB, false)
	// -> descriptor1:
	descriptor1 := test.NewMockDescriptor(&test.MockDescriptorArgs{
		Name:             descriptor1Name,
		KeySelector:      prefixSelector(prefixA),
		NBKeyPrefixes:    []string{prefixA},
		ValueBuilder:     test.ArrayValueBuilder(prefixA),
		DerValuesBuilder: test.ArrayValueDerBuilder(Object),
		WithMetadata:     true,
		DumpIsSupported:  true,
	}, mockSB, 1)
	// -> planned error
	mockSB.PlanError(prefixA + baseValue1 + "/item2", errors.New("failed to add value"),
		func() {
			mockSB.SetValue(prefixA + baseValue1, test.NewArrayValue(Object, baseValue1, "item1"),
				&test.OnlyInteger{Integer:0}, FromNB, false)
			})

	// register descriptor with the scheduler
	scheduler.RegisterKVDescriptor(descriptor1)

	// subscribe to receive notifications about errors
	errorChan := make(chan KeyWithError, 5)
	scheduler.SubscribeForErrors(errorChan, prefixSelector(prefixA))

	// get metadata map created for the descriptor
	metadataMap := scheduler.GetMetadataMap(descriptor1.GetName())
	nameToInteger, withMetadataMap := metadataMap.(test.NameToInteger)
	Expect(withMetadataMap).To(BeTrue())

	// run resync transaction that will fail for one value
	startTime := time.Now()
	values := []KeyValueDataPair{
		{Key: prefixA + baseValue1, ValueData: []string{"item1", "item2"}},
	}
	resyncTxn := scheduler.StartNBTransaction(WithRetry(3 * time.Second, false))
	kvErrors, txnError := resyncTxn.Resync(values).Commit(context.Background())
	stopTime := time.Now()
	Expect(txnError).ShouldNot(HaveOccurred())
	Expect(kvErrors).To(HaveLen(1))
	Expect(kvErrors[0].Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item2"))
	Expect(kvErrors[0].Error.Error()).To(BeEquivalentTo("failed to add value"))

	// check the state of SB
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	// -> base value 1
	value := mockSB.GetValue(prefixA + baseValue1)
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewArrayValue(Object, baseValue1, "item1"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(0))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item1 derived from base value 1
	value = mockSB.GetValue(prefixA + baseValue1 + "/item1")
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewStringValue(Object, "item1", "item1"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item2 derived from base value 1 failed to get added
	value = mockSB.GetValue(prefixA + baseValue1 + "/item2")
	Expect(value).To(BeNil())
	Expect(mockSB.GetValues(nil)).To(HaveLen(2))

	// check metadata
	metadata, exists := nameToInteger.LookupByName(baseValue1)
	Expect(exists).To(BeTrue())
	Expect(metadata.GetInteger()).To(BeEquivalentTo(0))

	// check operations executed in SB
	opHistory := mockSB.PopHistoryOfOps()
	Expect(opHistory).To(HaveLen(5))
	operation := opHistory[0]
	Expect(operation.OpType).To(Equal(test.Dump))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	checkValuesForCorrelation(operation.CorrelateDump, []KVWithMetadata{
		{
			Key: prefixA + baseValue1,
			Value: test.NewArrayValue(Object, baseValue1, "item1", "item2"),
			Metadata: nil,
			Origin: FromNB,
		},
	})
	operation = opHistory[1]
	Expect(operation.OpType).To(Equal(test.Modify))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[2]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[3]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item2"))
	Expect(operation.Err).ToNot(BeNil())
	Expect(operation.Err.Error()).To(BeEquivalentTo("failed to add value"))
	operation = opHistory[4] // refresh failed value
	Expect(operation.OpType).To(Equal(test.Dump))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	checkValuesForCorrelation(operation.CorrelateDump, []KVWithMetadata{
		{
			Key: prefixA + baseValue1,
			Value: test.NewArrayValue(Object, baseValue1, "item1", "item2"),
			Metadata: &test.OnlyInteger{Integer:0},
			Origin: FromNB,
		},
	})

	// check transaction operations
	txnHistory := scheduler.getTransactionHistory(time.Time{}, time.Time{})
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
		{key: prefixA + baseValue1, value: &recordedValue{valueType: Object, label: baseValue1, string: "[item1,item2]"}, origin: FromNB},
	})
	Expect(txn.preErrors).To(BeEmpty())

	txnOps := recordedTxnOps{
		{
			operation:  modify,
			key:        prefixA + baseValue1,
			prevValue:  &recordedValue{valueType: Object, label: baseValue1, string: "[]"},
			newValue:   &recordedValue{valueType: Object, label: baseValue1, string: "[item1,item2]"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  add,
			key:        prefixA + baseValue1 + "/item1",
			newValue:   &recordedValue{valueType: Object, label: "item1", string: "item1"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  add,
			key:        prefixA + baseValue1 + "/item2",
			newValue:   &recordedValue{valueType: Object, label: "item2", string: "item2"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
	}
	checkTxnOperations(txn.planned, txnOps)
	txnOps[2].isPending = true
	txnOps[2].newErr = errors.New("failed to add value")
	checkTxnOperations(txn.executed, txnOps)

	// check flag stats
	graphR := scheduler.graph.Read()
	errorStats := graphR.GetFlagStats(ErrorFlagName, nil)
	Expect(errorStats.TotalCount).To(BeEquivalentTo(1))
	pendingStats := graphR.GetFlagStats(PendingFlagName, nil)
	Expect(pendingStats.TotalCount).To(BeEquivalentTo(1))
	derivedStats := graphR.GetFlagStats(DerivedFlagName, nil)
	Expect(derivedStats.TotalCount).To(BeEquivalentTo(2))
	lastUpdateStats := graphR.GetFlagStats(LastUpdateFlagName, nil)
	Expect(lastUpdateStats.TotalCount).To(BeEquivalentTo(3))
	lastChangeStats := graphR.GetFlagStats(LastChangeFlagName, nil)
	Expect(lastChangeStats.TotalCount).To(BeEquivalentTo(1))
	descriptorStats := graphR.GetFlagStats(DescriptorFlagName, nil)
	Expect(descriptorStats.TotalCount).To(BeEquivalentTo(3))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor1Name))
	Expect(descriptorStats.PerValueCount[descriptor1Name]).To(BeEquivalentTo(3))
	originStats := graphR.GetFlagStats(OriginFlagName, nil)
	Expect(originStats.TotalCount).To(BeEquivalentTo(3))
	Expect(originStats.PerValueCount).To(HaveKey(FromNB.String()))
	Expect(originStats.PerValueCount[FromNB.String()]).To(BeEquivalentTo(3))
	graphR.Release()

	// check error updates received through the channel
	var errorNotif KeyWithError
	Eventually(errorChan, time.Second).Should(Receive(&errorNotif))
	Expect(errorNotif.Key).To(Equal(prefixA + baseValue1 + "/item2"))
	Expect(errorNotif.Error).ToNot(BeNil())
	Expect(errorNotif.Error.Error()).To(BeEquivalentTo("failed to add value"))

	// eventually the value should get "fixed"
	Eventually(errorChan, 5*time.Second).Should(Receive(&errorNotif))
	Expect(errorNotif.Key).To(Equal(prefixA + baseValue1 + "/item2"))
	Expect(errorNotif.Error).To(BeNil())

	// check the state of SB after retry
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	// -> base value 1
	value = mockSB.GetValue(prefixA + baseValue1)
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewArrayValue(Object, baseValue1, "item1", "item2"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(0))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item1 derived from base value 1
	value = mockSB.GetValue(prefixA + baseValue1 + "/item1")
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewStringValue(Object, "item1", "item1"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	Expect(mockSB.GetValues(nil)).To(HaveLen(3))
	// -> item2 derived from base value 1 was re-added
	value = mockSB.GetValue(prefixA + baseValue1 + "/item2")
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewStringValue(Object, "item2", "item2"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	Expect(mockSB.GetValues(nil)).To(HaveLen(3))

	// check metadata
	metadata, exists = nameToInteger.LookupByName(baseValue1)
	Expect(exists).To(BeTrue())
	Expect(metadata.GetInteger()).To(BeEquivalentTo(0))

	// check operations executed in SB during retry
	opHistory = mockSB.PopHistoryOfOps()
	Expect(opHistory).To(HaveLen(2))
	operation = opHistory[0]
	Expect(operation.OpType).To(Equal(test.Modify))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[1]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item2"))
	Expect(operation.Err).To(BeNil())

	// check retry transaction operations
	txnHistory = scheduler.getTransactionHistory(time.Time{}, time.Now())
	Expect(txnHistory).To(HaveLen(2))
	txn = txnHistory[1]
	Expect(txn.preRecord).To(BeFalse())
	Expect(txn.start.After(stopTime)).To(BeTrue())
	Expect(txn.start.Before(txn.stop)).To(BeTrue())
	Expect(txn.stop.Before(time.Now())).To(BeTrue())
	Expect(txn.seqNum).To(BeEquivalentTo(1))
	Expect(txn.txnType).To(BeEquivalentTo(retryFailedOps))
	Expect(txn.isResync).To(BeFalse())
	checkRecordedValues(txn.values, []recordedKVPair{
		{key: prefixA + baseValue1, value: &recordedValue{valueType: Object, label: baseValue1, string: "[item1,item2]"}, origin: FromNB},
	})
	Expect(txn.preErrors).To(BeEmpty())

	txnOps = recordedTxnOps{
		{
			operation:  modify,
			key:        prefixA + baseValue1,
			prevValue:  &recordedValue{valueType: Object, label: baseValue1, string: "[item1]"},
			newValue:   &recordedValue{valueType: Object, label: baseValue1, string: "[item1,item2]"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			isRetry:    true,
		},
		{
			operation:  add,
			key:        prefixA + baseValue1 + "/item2",
			prevValue:  &recordedValue{valueType: Object, label: "item2", string: "item2"},
			newValue:   &recordedValue{valueType: Object, label: "item2", string: "item2"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			prevErr:    errors.New("failed to add value"),
			wasPending: true,
			isRetry:    true,
		},
	}
	checkTxnOperations(txn.planned, txnOps)
	checkTxnOperations(txn.executed, txnOps)

	// check flag stats
	graphR = scheduler.graph.Read()
	errorStats = graphR.GetFlagStats(ErrorFlagName, nil)
	Expect(errorStats.TotalCount).To(BeEquivalentTo(1))
	pendingStats = graphR.GetFlagStats(PendingFlagName, nil)
	Expect(pendingStats.TotalCount).To(BeEquivalentTo(1))
	derivedStats = graphR.GetFlagStats(DerivedFlagName, nil)
	Expect(derivedStats.TotalCount).To(BeEquivalentTo(4))
	lastUpdateStats = graphR.GetFlagStats(LastUpdateFlagName, nil)
	Expect(lastUpdateStats.TotalCount).To(BeEquivalentTo(6))
	lastChangeStats = graphR.GetFlagStats(LastChangeFlagName, nil)
	Expect(lastChangeStats.TotalCount).To(BeEquivalentTo(2))
	descriptorStats = graphR.GetFlagStats(DescriptorFlagName, nil)
	Expect(descriptorStats.TotalCount).To(BeEquivalentTo(6))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor1Name))
	Expect(descriptorStats.PerValueCount[descriptor1Name]).To(BeEquivalentTo(6))
	originStats = graphR.GetFlagStats(OriginFlagName, nil)
	Expect(originStats.TotalCount).To(BeEquivalentTo(6))
	Expect(originStats.PerValueCount).To(HaveKey(FromNB.String()))
	Expect(originStats.PerValueCount[FromNB.String()]).To(BeEquivalentTo(6))
	graphR.Release()

	// close scheduler
	err = scheduler.Close()
	Expect(err).To(BeNil())
}

/* when graph dump is needed:
	graphR := scheduler.graph.Read()
	graphDump := graphR.Dump()
	fmt.Print(graphDump)
	graphR.Release()
 */