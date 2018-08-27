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
	"strings"
	. "github.com/onsi/gomega"

	"github.com/ligato/cn-infra/kvscheduler/test"
	. "github.com/ligato/cn-infra/kvscheduler/api"
)

func TestDataChangeTransactions(t *testing.T) {
	RegisterTestingT(t)

	// prepare KV Scheduler
	scheduler := NewPlugin(UseDeps(func(deps *Deps) {
		deps.HTTPHandlers = nil
		}))
	err := scheduler.Init()
	Expect(err).To(BeNil())

	// prepare mocks
	mockSB := test.NewMockSouthbound()
	// -> descriptor1:
	descriptor1 := test.NewMockDescriptor(&test.MockDescriptorArgs{
		Name:             descriptor1Name,
		KeySelector:      prefixSelector(prefixA),
		NBKeyPrefixes:    []string{prefixA},
		ValueBuilder:     test.ArrayValueBuilder(prefixA),
		DerValuesBuilder: test.ArrayValueDerBuilder(Object),
		WithMetadata:     true,
		DumpIsSupported:  true,
	}, mockSB, 0)
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
	}, mockSB, 0)
	// -> descriptor3:
	descriptor3 := test.NewMockDescriptor(&test.MockDescriptorArgs{
		Name:            descriptor3Name,
		KeySelector:     prefixSelector(prefixC),
		NBKeyPrefixes:   []string{prefixC},
		ValueBuilder:    func(key string, valueData interface{}) (value Value, err error) {
			label := strings.TrimPrefix(key, prefixC)
			items, ok := valueData.([]string)
			if !ok {
				return nil, ErrInvalidValueDataType(key)
			}
			return test.NewArrayValue(Object, label, items...), nil
		},
		DerValuesBuilder: test.ArrayValueDerBuilder(Object),
		RecreateChecker: func(key string, oldValue, newValue Value, metadata Metadata) bool {
			if key == prefixC + baseValue3 {
				return true
			}
			return false
		},
		WithMetadata:     true,
		DumpIsSupported:  true,
		DumpDependencies: []string{descriptor2Name},
	}, mockSB, 0)

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

	// run non-resync transaction against empty SB
	startTime := time.Now()
	schedulerTxn := scheduler.StartNBTransaction()
	schedulerTxn.SetValueData(prefixB + baseValue2, []string{"item1", "item2"})
	schedulerTxn.SetValueData(prefixA + baseValue1, []string{"item2"})
	schedulerTxn.SetValueData(prefixC + baseValue3, []string{"item1", "item2"})
	kvErrors, txnError := schedulerTxn.Commit(context.Background())
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
	// -> item1 derived from base value was not added
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
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(0))
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

	// check pending values
	pendingValues := scheduler.GetPendingValues(nil)
	checkValues(pendingValues, []KeyValuePair{
		{Key: prefixB + baseValue2 + "/item2", Value: test.NewStringValue(Object, "item2", "item2")},
	})

	// check metadata
	metadata, exists := nameToInteger1.LookupByName(baseValue1)
	Expect(exists).To(BeTrue())
	Expect(metadata.GetInteger()).To(BeEquivalentTo(0))
	metadata, exists = nameToInteger2.LookupByName(baseValue2)
	Expect(exists).To(BeTrue())
	Expect(metadata.GetInteger()).To(BeEquivalentTo(0))
	metadata, exists = nameToInteger3.LookupByName(baseValue3)
	Expect(exists).To(BeTrue())
	Expect(metadata.GetInteger()).To(BeEquivalentTo(0))

	// check operations executed in SB
	opHistory := mockSB.PopHistoryOfOps()
	Expect(opHistory).To(HaveLen(7))
	operation := opHistory[0]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[1]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item2"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[2]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixB + baseValue2))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[3]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixB + baseValue2 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[4]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[5]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[6]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3 + "/item2"))
	Expect(operation.Err).To(BeNil())

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
	Expect(txn.isResync).To(BeFalse())
	checkRecordedValues(txn.values, []recordedKVPair{
		{key: prefixA + baseValue1, value: &recordedValue{valueType: Object, label: baseValue1, string: "[item2]"}, origin: FromNB},
		{key: prefixB + baseValue2, value: &recordedValue{valueType: Object, label: baseValue2, string: "[item1,item2]"}, origin: FromNB},
		{key: prefixC + baseValue3, value: &recordedValue{valueType: Object, label: baseValue3, string: "[item1,item2]"}, origin: FromNB},
	})
	Expect(txn.preErrors).To(BeEmpty())

	txnOps := recordedTxnOps{
		{
			operation:  add,
			key:        prefixA + baseValue1,
			newValue:   &recordedValue{valueType: Object, label: baseValue1, string: "[item2]"},
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
			operation:  add,
			key:        prefixB + baseValue2,
			newValue:   &recordedValue{valueType: Object, label: baseValue2, string: "[item1,item2]"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
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
			key:        prefixC + baseValue3,
			newValue:   &recordedValue{valueType: Object, label: baseValue3, string: "[item1,item2]"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
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

	// run 2nd non-resync transaction against empty SB
	startTime = time.Now()
	schedulerTxn2 := scheduler.StartNBTransaction()
	schedulerTxn2.SetValueData(prefixC + baseValue3, []string{"item1"})
	schedulerTxn2.SetValueData(prefixA + baseValue1, []string{"item1"})
	kvErrors, txnError = schedulerTxn2.Commit(context.Background())
	stopTime = time.Now()
	Expect(txnError).ShouldNot(HaveOccurred())
	Expect(kvErrors).To(BeEmpty())

	// check the state of SB
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	// -> base value 1
	value = mockSB.GetValue(prefixA + baseValue1)
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewArrayValue(Object, baseValue1, "item1"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(0))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item1 derived from base value was added
	value = mockSB.GetValue(prefixA + baseValue1 + "/item1")
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewStringValue(Object, "item1", "item1"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item2 derived from base value 1 was deleted
	value = mockSB.GetValue(prefixA + baseValue1 + "/item2")
	Expect(value).To(BeNil())
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
	// -> item2 derived from base value 2 is no longer pending
	value = mockSB.GetValue(prefixB + baseValue2 + "/item2")
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewStringValue(Object, "item2", "item2"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> base value 3
	value = mockSB.GetValue(prefixC + baseValue3)
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewArrayValue(Object, baseValue3, "item1"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(1))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item1 derived from base value 3
	value = mockSB.GetValue(prefixC + baseValue3 + "/item1")
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewStringValue(Object, "item1", "item1"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item2 derived from base value 3 was deleted
	value = mockSB.GetValue(prefixC + baseValue3 + "/item2")
	Expect(value).To(BeNil())

	// check pending values
	Expect(scheduler.GetPendingValues(nil)).To(BeEmpty())

	// check metadata
	metadata, exists = nameToInteger1.LookupByName(baseValue1)
	Expect(exists).To(BeTrue())
	Expect(metadata.GetInteger()).To(BeEquivalentTo(0))
	metadata, exists = nameToInteger2.LookupByName(baseValue2)
	Expect(exists).To(BeTrue())
	Expect(metadata.GetInteger()).To(BeEquivalentTo(0))
	metadata, exists = nameToInteger3.LookupByName(baseValue3)
	Expect(exists).To(BeTrue())
	Expect(metadata.GetInteger()).To(BeEquivalentTo(1)) // re-created

	// check operations executed in SB
	opHistory = mockSB.PopHistoryOfOps()
	Expect(opHistory).To(HaveLen(10))
	operation = opHistory[0]
	Expect(operation.OpType).To(Equal(test.Delete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[1]
	Expect(operation.OpType).To(Equal(test.Delete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3 + "/item2"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[2]
	Expect(operation.OpType).To(Equal(test.Delete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[3]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[4]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[5]
	Expect(operation.OpType).To(Equal(test.Delete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item2"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[6]
	Expect(operation.OpType).To(Equal(test.Modify))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[7]
	Expect(operation.OpType).To(Equal(test.Update))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixB + baseValue2 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[8]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[9]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixB + baseValue2 + "/item2"))
	Expect(operation.Err).To(BeNil())

	// check transaction operations
	txnHistory = scheduler.getTransactionHistory(startTime, stopTime) // first txn not included
	Expect(txnHistory).To(HaveLen(1))
	txn = txnHistory[0]
	Expect(txn.preRecord).To(BeFalse())
	Expect(txn.start.After(startTime)).To(BeTrue())
	Expect(txn.start.Before(txn.stop)).To(BeTrue())
	Expect(txn.stop.Before(stopTime)).To(BeTrue())
	Expect(txn.seqNum).To(BeEquivalentTo(1))
	Expect(txn.txnType).To(BeEquivalentTo(nbTransaction))
	Expect(txn.isResync).To(BeFalse())
	checkRecordedValues(txn.values, []recordedKVPair{
		{key: prefixA + baseValue1, value: &recordedValue{valueType: Object, label: baseValue1, string: "[item1]"}, origin: FromNB},
		{key: prefixC + baseValue3, value: &recordedValue{valueType: Object, label: baseValue3, string: "[item1]"}, origin: FromNB},
	})
	Expect(txn.preErrors).To(BeEmpty())

	txnOps = recordedTxnOps{
		{
			operation:  del,
			key:        prefixC + baseValue3 + "/item1",
			prevValue:  &recordedValue{valueType: Object, label: "item1", string: "item1"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  del,
			key:        prefixC + baseValue3 + "/item2",
			prevValue:  &recordedValue{valueType: Object, label: "item2", string: "item2"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  del,
			key:        prefixC + baseValue3,
			prevValue:	&recordedValue{valueType: Object, label: baseValue3, string: "[item1,item2]"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			isPending:  true,
		},
		{
			operation:  add,
			key:        prefixC + baseValue3,
			newValue:   &recordedValue{valueType: Object, label: baseValue3, string: "[item1]"},
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
			operation:  del,
			key:        prefixA + baseValue1 + "/item2",
			prevValue:  &recordedValue{valueType: Object, label: "item2", string: "item2"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  modify,
			key:        prefixA + baseValue1,
			prevValue:  &recordedValue{valueType: Object, label: baseValue1, string: "[item2]"},
			newValue:   &recordedValue{valueType: Object, label: baseValue1, string: "[item1]"},
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
			key:        prefixA + baseValue1 + "/item1",
			newValue:   &recordedValue{valueType: Object, label: "item1", string: "item1"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
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
	Expect(pendingStats.TotalCount).To(BeEquivalentTo(1))
	derivedStats = graphR.GetFlagStats(DerivedFlagName, nil)
	Expect(derivedStats.TotalCount).To(BeEquivalentTo(9))
	lastUpdateStats = graphR.GetFlagStats(LastUpdateFlagName, nil)
	Expect(lastUpdateStats.TotalCount).To(BeEquivalentTo(14))
	lastChangeStats = graphR.GetFlagStats(LastChangeFlagName, nil)
	Expect(lastChangeStats.TotalCount).To(BeEquivalentTo(5))
	descriptorStats = graphR.GetFlagStats(DescriptorFlagName, nil)
	Expect(descriptorStats.TotalCount).To(BeEquivalentTo(14))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor1Name))
	Expect(descriptorStats.PerValueCount[descriptor1Name]).To(BeEquivalentTo(4))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor2Name))
	Expect(descriptorStats.PerValueCount[descriptor2Name]).To(BeEquivalentTo(5))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor3Name))
	Expect(descriptorStats.PerValueCount[descriptor3Name]).To(BeEquivalentTo(5))
	originStats = graphR.GetFlagStats(OriginFlagName, nil)
	Expect(originStats.TotalCount).To(BeEquivalentTo(14))
	Expect(originStats.PerValueCount).To(HaveKey(FromNB.String()))
	Expect(originStats.PerValueCount[FromNB.String()]).To(BeEquivalentTo(14))
	graphR.Release()

	// close scheduler
	err = scheduler.Close()
	Expect(err).To(BeNil())
}

func TestDataChangeTransactionWithRevert(t *testing.T) {
	RegisterTestingT(t)

	// prepare KV Scheduler
	scheduler := NewPlugin(UseDeps(func(deps *Deps) {
		deps.HTTPHandlers = nil
	}))
	err := scheduler.Init()
	Expect(err).To(BeNil())

	// prepare mocks
	mockSB := test.NewMockSouthbound()
	// -> descriptor1:
	descriptor1 := test.NewMockDescriptor(&test.MockDescriptorArgs{
		Name:             descriptor1Name,
		KeySelector:      prefixSelector(prefixA),
		NBKeyPrefixes:    []string{prefixA},
		ValueBuilder:     test.ArrayValueBuilder(prefixA),
		DerValuesBuilder: test.ArrayValueDerBuilder(Object),
		WithMetadata:     true,
		DumpIsSupported:  true,
	}, mockSB, 0)
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
	}, mockSB, 0)
	// -> descriptor3:
	descriptor3 := test.NewMockDescriptor(&test.MockDescriptorArgs{
		Name:            descriptor3Name,
		KeySelector:     prefixSelector(prefixC),
		NBKeyPrefixes:   []string{prefixC},
		ValueBuilder:    test.ArrayValueBuilder(prefixC),
		DerValuesBuilder: test.ArrayValueDerBuilder(Object),
		RecreateChecker: func(key string, oldValue, newValue Value, metadata Metadata) bool {
			if key == prefixC + baseValue3 {
				return true
			}
			return false
		},
		WithMetadata:     true,
		DumpIsSupported:  true,
		DumpDependencies: []string{descriptor2Name},
	}, mockSB, 0)

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

	// run 1st non-resync transaction against empty SB
	schedulerTxn := scheduler.StartNBTransaction()
	schedulerTxn.SetValueData(prefixB + baseValue2, []string{"item1", "item2"})
	schedulerTxn.SetValueData(prefixA + baseValue1, []string{"item2"})
	schedulerTxn.SetValueData(prefixC + baseValue3, []string{"item1", "item2"})
	kvErrors, txnError := schedulerTxn.Commit(context.Background())
	Expect(txnError).ShouldNot(HaveOccurred())
	Expect(kvErrors).To(BeEmpty())
	mockSB.PopHistoryOfOps()

	// plan error before 2nd txn
	failedModifyClb := func() {
		mockSB.SetValue(prefixA + baseValue1, test.NewArrayValue(Object, baseValue1),
			&test.OnlyInteger{Integer:0}, FromNB, false)
	}
	mockSB.PlanError(prefixA + baseValue1, errors.New("failed to modify value"), failedModifyClb)
	mockSB.PlanError(prefixA + baseValue1, errors.New("failed to modify value, again"), failedModifyClb) // the error will repeat one more time

	// subscribe to receive notifications about errors
	errorChan := make(chan KeyWithError, 5)
	scheduler.SubscribeForErrors(errorChan, prefixSelector(prefixA))

	// run 2nd non-resync transaction against empty SB that will fail and will be reverted
	startTime := time.Now()
	schedulerTxn2 := scheduler.StartNBTransaction(WithRevert(), WithRetry(3 * time.Second, true))
	schedulerTxn2.SetValueData(prefixC + baseValue3, []string{"item1"})
	schedulerTxn2.SetValueData(prefixA + baseValue1, []string{"item1"})
	kvErrors, txnError = schedulerTxn2.Commit(context.Background())
	stopTime := time.Now()
	Expect(txnError).ShouldNot(HaveOccurred())
	Expect(kvErrors).To(HaveLen(1))
	Expect(kvErrors[0].Key).To(BeEquivalentTo(prefixA + baseValue1))
	Expect(kvErrors[0].Error.Error()).To(BeEquivalentTo("failed to modify value"))

	// receive the error notification
	var errorNotif KeyWithError
	Eventually(errorChan, time.Second).Should(Receive(&errorNotif))
	Expect(errorNotif.Key).To(Equal(prefixA + baseValue1))
	Expect(errorNotif.Error).ToNot(BeNil())
	Expect(errorNotif.Error.Error()).To(BeEquivalentTo("failed to modify value"))

	// check the state of SB
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	// -> base value 1
	value := mockSB.GetValue(prefixA + baseValue1)
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewArrayValue(Object, baseValue1))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(0))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item1 derived from base value was NOT added
	value = mockSB.GetValue(prefixA + baseValue1 + "/item1")
	Expect(value).To(BeNil())
	// -> item2 derived from base value 1 was deleted
	value = mockSB.GetValue(prefixA + baseValue1 + "/item2")
	Expect(value).To(BeNil())
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
	// -> item2 derived from base value 2 is still pending
	value = mockSB.GetValue(prefixB + baseValue2 + "/item2")
	Expect(value).To(BeNil())
	// -> base value 3 was re-verted back to state after 1st txn
	value = mockSB.GetValue(prefixC + baseValue3)
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewArrayValue(Object, baseValue3, "item1", "item2"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(2))
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

	// check metadata
	metadata, exists := nameToInteger1.LookupByName(baseValue1)
	Expect(exists).To(BeTrue())
	Expect(metadata.GetInteger()).To(BeEquivalentTo(0))
	metadata, exists = nameToInteger2.LookupByName(baseValue2)
	Expect(exists).To(BeTrue())
	Expect(metadata.GetInteger()).To(BeEquivalentTo(0))
	metadata, exists = nameToInteger3.LookupByName(baseValue3)
	Expect(exists).To(BeTrue())
	Expect(metadata.GetInteger()).To(BeEquivalentTo(2)) // re-created twice

	// check operations executed in SB during 2nd txn
	opHistory := mockSB.PopHistoryOfOps()
	Expect(opHistory).To(HaveLen(13))
	operation := opHistory[0]
	Expect(operation.OpType).To(Equal(test.Delete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[1]
	Expect(operation.OpType).To(Equal(test.Delete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3 + "/item2"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[2]
	Expect(operation.OpType).To(Equal(test.Delete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[3]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[4]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[5]
	Expect(operation.OpType).To(Equal(test.Delete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item2"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[6]
	Expect(operation.OpType).To(Equal(test.Modify))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1))
	Expect(operation.Err).ToNot(BeNil())
	Expect(operation.Err.Error()).To(BeEquivalentTo("failed to modify value"))
	// reverting:
	operation = opHistory[7]
	Expect(operation.OpType).To(Equal(test.Delete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[8]
	Expect(operation.OpType).To(Equal(test.Delete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[9]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[10]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[11]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3 + "/item2"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[12] // refresh failed value
	Expect(operation.OpType).To(Equal(test.Dump))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	checkValuesForCorrelation(operation.CorrelateDump, []KVWithMetadata{
		{
			Key: prefixA + baseValue1,
			Value: test.NewArrayValue(Object, baseValue1, "item1"),
			Metadata: &test.OnlyInteger{Integer:0},
			Origin: FromNB,
		},
	})

	// check transaction operations
	txnHistory := scheduler.getTransactionHistory(startTime, time.Now())
	Expect(txnHistory).To(HaveLen(1))
	txn := txnHistory[0]
	Expect(txn.preRecord).To(BeFalse())
	Expect(txn.start.After(startTime)).To(BeTrue())
	Expect(txn.start.Before(txn.stop)).To(BeTrue())
	Expect(txn.stop.Before(stopTime)).To(BeTrue())
	Expect(txn.seqNum).To(BeEquivalentTo(1))
	Expect(txn.txnType).To(BeEquivalentTo(nbTransaction))
	Expect(txn.isResync).To(BeFalse())
	checkRecordedValues(txn.values, []recordedKVPair{
		{key: prefixA + baseValue1, value: &recordedValue{valueType: Object, label: baseValue1, string: "[item1]"}, origin: FromNB},
		{key: prefixC + baseValue3, value: &recordedValue{valueType: Object, label: baseValue3, string: "[item1]"}, origin: FromNB},
	})
	Expect(txn.preErrors).To(BeEmpty())

	// planned operations
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
			key:        prefixC + baseValue3 + "/item2",
			prevValue:  &recordedValue{valueType: Object, label: "item2", string: "item2"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  del,
			key:        prefixC + baseValue3,
			prevValue:	&recordedValue{valueType: Object, label: baseValue3, string: "[item1,item2]"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			isPending:  true,
		},
		{
			operation:  add,
			key:        prefixC + baseValue3,
			newValue:   &recordedValue{valueType: Object, label: baseValue3, string: "[item1]"},
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
			operation:  del,
			key:        prefixA + baseValue1 + "/item2",
			prevValue:  &recordedValue{valueType: Object, label: "item2", string: "item2"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  modify,
			key:        prefixA + baseValue1,
			prevValue:  &recordedValue{valueType: Object, label: baseValue1, string: "[item2]"},
			newValue:   &recordedValue{valueType: Object, label: baseValue1, string: "[item1]"},
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
			key:        prefixA + baseValue1 + "/item1",
			newValue:   &recordedValue{valueType: Object, label: "item1", string: "item1"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
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

	// executed operations
	txnOps = recordedTxnOps{
		{
			operation:  del,
			key:        prefixC + baseValue3 + "/item1",
			prevValue:  &recordedValue{valueType: Object, label: "item1", string: "item1"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  del,
			key:        prefixC + baseValue3 + "/item2",
			prevValue:  &recordedValue{valueType: Object, label: "item2", string: "item2"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  del,
			key:        prefixC + baseValue3,
			prevValue:	&recordedValue{valueType: Object, label: baseValue3, string: "[item1,item2]"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			isPending:  true,
		},
		{
			operation:  add,
			key:        prefixC + baseValue3,
			newValue:   &recordedValue{valueType: Object, label: baseValue3, string: "[item1]"},
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
			operation:  del,
			key:        prefixA + baseValue1 + "/item2",
			prevValue:  &recordedValue{valueType: Object, label: "item2", string: "item2"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  modify,
			key:        prefixA + baseValue1,
			prevValue:  &recordedValue{valueType: Object, label: baseValue1, string: "[item2]"},
			newValue:   &recordedValue{valueType: Object, label: baseValue1, string: "[item1]"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			newErr:     errors.New("failed to modify value"),
		},
		// reverting:
		{
			operation:  del,
			key:        prefixC + baseValue3 + "/item1",
			prevValue:  &recordedValue{valueType: Object, label: "item1", string: "item1"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			isRevert:   true,
		},
		{
			operation:  del,
			key:        prefixC + baseValue3,
			prevValue:	&recordedValue{valueType: Object, label: baseValue3, string: "[item1]"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			isPending:  true,
			isRevert:   true,
		},
		{
			operation:  add,
			key:        prefixC + baseValue3,
			newValue:   &recordedValue{valueType: Object, label: baseValue3, string: "[item1,item2]"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			isRevert:   true,
			wasPending: true,
		},
		{
			operation:  add,
			key:        prefixC + baseValue3 + "/item1",
			newValue:   &recordedValue{valueType: Object, label: "item1", string: "item1"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			isRevert:   true,
		},
		{
			operation:  add,
			key:        prefixC + baseValue3 + "/item2",
			newValue:   &recordedValue{valueType: Object, label: "item2", string: "item2"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			isRevert:   true,
		},
	}
	checkTxnOperations(txn.executed, txnOps)

	// check flag stats
	graphR := scheduler.graph.Read()
	errorStats := graphR.GetFlagStats(ErrorFlagName, nil)
	Expect(errorStats.TotalCount).To(BeEquivalentTo(1))
	pendingStats := graphR.GetFlagStats(PendingFlagName, nil)
	Expect(pendingStats.TotalCount).To(BeEquivalentTo(1))
	derivedStats := graphR.GetFlagStats(DerivedFlagName, nil)
	Expect(derivedStats.TotalCount).To(BeEquivalentTo(7))
	lastUpdateStats := graphR.GetFlagStats(LastUpdateFlagName, nil)
	Expect(lastUpdateStats.TotalCount).To(BeEquivalentTo(12))
	lastChangeStats := graphR.GetFlagStats(LastChangeFlagName, nil)
	Expect(lastChangeStats.TotalCount).To(BeEquivalentTo(5))
	descriptorStats := graphR.GetFlagStats(DescriptorFlagName, nil)
	Expect(descriptorStats.TotalCount).To(BeEquivalentTo(12))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor1Name))
	Expect(descriptorStats.PerValueCount[descriptor1Name]).To(BeEquivalentTo(3))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor2Name))
	Expect(descriptorStats.PerValueCount[descriptor2Name]).To(BeEquivalentTo(3))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor3Name))
	Expect(descriptorStats.PerValueCount[descriptor3Name]).To(BeEquivalentTo(6))
	originStats := graphR.GetFlagStats(OriginFlagName, nil)
	Expect(originStats.TotalCount).To(BeEquivalentTo(12))
	Expect(originStats.PerValueCount).To(HaveKey(FromNB.String()))
	Expect(originStats.PerValueCount[FromNB.String()]).To(BeEquivalentTo(12))
	graphR.Release()

	// first attempt to revert the baseValue1 to pre-txn2 state will fail
	Eventually(errorChan, 5*time.Second).Should(Receive(&errorNotif))
	Expect(errorNotif.Key).To(Equal(prefixA + baseValue1))
	Expect(errorNotif.Error).ToNot(BeNil())
	Expect(errorNotif.Error.Error()).To(BeEquivalentTo("failed to modify value, again"))

	// check operations executed in SB during 1st revert attempt
	opHistory = mockSB.PopHistoryOfOps()
	Expect(opHistory).To(HaveLen(2))
	operation = opHistory[0]
	Expect(operation.OpType).To(Equal(test.Modify))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1))
	Expect(operation.Err).ToNot(BeNil())
	Expect(operation.Err.Error()).To(BeEquivalentTo("failed to modify value, again"))
	operation = opHistory[1] // refresh failed revert
	Expect(operation.OpType).To(Equal(test.Dump))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	checkValuesForCorrelation(operation.CorrelateDump, []KVWithMetadata{
		{
			Key: prefixA + baseValue1,
			Value: test.NewArrayValue(Object, baseValue1, "item2"),
			Metadata: &test.OnlyInteger{Integer:0},
			Origin: FromNB,
		},
	})

	// check transaction operations
	txnHistory = scheduler.getTransactionHistory(startTime, time.Now())
	Expect(txnHistory).To(HaveLen(2))
	txn = txnHistory[1]
	Expect(txn.preRecord).To(BeFalse())
	Expect(txn.start.After(stopTime)).To(BeTrue())
	Expect(txn.start.Before(txn.stop)).To(BeTrue())
	Expect(txn.stop.Before(time.Now())).To(BeTrue())
	Expect(txn.seqNum).To(BeEquivalentTo(2))
	Expect(txn.txnType).To(BeEquivalentTo(retryFailedOps))
	Expect(txn.isResync).To(BeFalse())
	checkRecordedValues(txn.values, []recordedKVPair{
		{key: prefixA + baseValue1, value: &recordedValue{valueType: Object, label: baseValue1, string: "[item2]"}, origin: FromNB},
	})
	Expect(txn.preErrors).To(BeEmpty())

	// planned operations
	txnOps = recordedTxnOps{
		{
			operation:  modify,
			key:        prefixA + baseValue1,
			prevValue:  &recordedValue{valueType: Object, label: baseValue1, string: "[]"},
			newValue:   &recordedValue{valueType: Object, label: baseValue1, string: "[item2]"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			prevErr:    errors.New("failed to modify value"),
			isRevert:   true,
			isRetry:    true,
		},
		{
			operation:  update,
			key:        prefixB + baseValue2 + "/item1",
			prevValue:  &recordedValue{valueType: Object, label: "item1", string: "item1"},
			newValue:   &recordedValue{valueType: Object, label: "item1", string: "item1"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			isRevert:   true,
			isRetry:    true,
		},
		{
			operation:  add,
			key:        prefixA + baseValue1 + "/item2",
			newValue:   &recordedValue{valueType: Object, label: "item2", string: "item2"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			isRevert:   true,
			isRetry:    true,
		},
	}
	checkTxnOperations(txn.planned, txnOps)

	// executed operations
	txnOps = recordedTxnOps{
		{
			operation:  modify,
			key:        prefixA + baseValue1,
			prevValue:  &recordedValue{valueType: Object, label: baseValue1, string: "[]"},
			newValue:   &recordedValue{valueType: Object, label: baseValue1, string: "[item2]"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			prevErr:    errors.New("failed to modify value"),
			newErr:     errors.New("failed to modify value, again"),
			isRevert:   true,
			isRetry:    true,
		},
	}
	checkTxnOperations(txn.executed, txnOps)

	// second attempt to revert the baseValue1 to pre-txn2 should succeed
	Eventually(errorChan, 10*time.Second).Should(Receive(&errorNotif))
	Expect(errorNotif.Key).To(Equal(prefixA + baseValue1))
	Expect(errorNotif.Error).To(BeNil())

	// check operations executed in SB during 2nd revert attempt
	opHistory = mockSB.PopHistoryOfOps()
	Expect(opHistory).To(HaveLen(3))
	operation = opHistory[0]
	Expect(operation.OpType).To(Equal(test.Modify))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[1]
	Expect(operation.OpType).To(Equal(test.Update))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixB + baseValue2 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[2]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item2"))
	Expect(operation.Err).To(BeNil())

	// check transaction operations
	txnHistory = scheduler.getTransactionHistory(startTime, time.Now())
	Expect(txnHistory).To(HaveLen(3))
	txn = txnHistory[2]
	Expect(txn.preRecord).To(BeFalse())
	Expect(txn.start.After(stopTime)).To(BeTrue())
	Expect(txn.start.Before(txn.stop)).To(BeTrue())
	Expect(txn.stop.Before(time.Now())).To(BeTrue())
	Expect(txn.seqNum).To(BeEquivalentTo(3))
	Expect(txn.txnType).To(BeEquivalentTo(retryFailedOps))
	Expect(txn.isResync).To(BeFalse())
	checkRecordedValues(txn.values, []recordedKVPair{
		{key: prefixA + baseValue1, value: &recordedValue{valueType: Object, label: baseValue1, string: "[item2]"}, origin: FromNB},
	})
	Expect(txn.preErrors).To(BeEmpty())
	txnOps = recordedTxnOps{
		{
			operation:  modify,
			key:        prefixA + baseValue1,
			prevValue:  &recordedValue{valueType: Object, label: baseValue1, string: "[]"},
			newValue:   &recordedValue{valueType: Object, label: baseValue1, string: "[item2]"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			prevErr:    errors.New("failed to modify value, again"),
			isRevert:   true,
			isRetry:    true,
		},
		{
			operation:  update,
			key:        prefixB + baseValue2 + "/item1",
			prevValue:  &recordedValue{valueType: Object, label: "item1", string: "item1"},
			newValue:   &recordedValue{valueType: Object, label: "item1", string: "item1"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			isRevert:   true,
			isRetry:    true,
		},
		{
			operation:  add,
			key:        prefixA + baseValue1 + "/item2",
			newValue:   &recordedValue{valueType: Object, label: "item2", string: "item2"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			isRevert:   true,
			isRetry:    true,
		},
	}
	checkTxnOperations(txn.planned, txnOps)
	checkTxnOperations(txn.executed, txnOps)

	// close scheduler
	err = scheduler.Close()
	Expect(err).To(BeNil())
}

