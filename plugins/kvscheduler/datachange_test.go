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

	"github.com/golang/protobuf/proto"
	. "github.com/onsi/gomega"

	. "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler/internal/test"
	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler/internal/utils"
	. "go.ligato.io/vpp-agent/v3/proto/ligato/kvscheduler"
)

var testCtx = WithSimulation(context.Background())

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
	descriptor1 := test.NewMockDescriptor(&KVDescriptor{
		Name:          descriptor1Name,
		NBKeyPrefix:   prefixA,
		KeySelector:   prefixSelector(prefixA),
		ValueTypeName: proto.MessageName(test.NewArrayValue()),
		DerivedValues: test.ArrayValueDerBuilder,
		WithMetadata:  true,
	}, mockSB, 0)
	// -> descriptor2:
	descriptor2 := test.NewMockDescriptor(&KVDescriptor{
		Name:          descriptor2Name,
		NBKeyPrefix:   prefixB,
		KeySelector:   prefixSelector(prefixB),
		ValueTypeName: proto.MessageName(test.NewArrayValue()),
		DerivedValues: test.ArrayValueDerBuilder,
		Dependencies: func(key string, value proto.Message) []Dependency {
			if key == prefixB+baseValue2+"/item1" {
				depKey := prefixA + baseValue1
				return []Dependency{
					{Label: depKey, Key: depKey},
				}
			}
			if key == prefixB+baseValue2+"/item2" {
				depKey := prefixA + baseValue1 + "/item1"
				return []Dependency{
					{Label: depKey, Key: depKey},
				}
			}
			return nil
		},
		WithMetadata:         true,
		RetrieveDependencies: []string{descriptor1Name},
	}, mockSB, 0)
	// -> descriptor3:
	descriptor3 := test.NewMockDescriptor(&KVDescriptor{
		Name:          descriptor3Name,
		NBKeyPrefix:   prefixC,
		KeySelector:   prefixSelector(prefixC),
		ValueTypeName: proto.MessageName(test.NewArrayValue()),
		DerivedValues: test.ArrayValueDerBuilder,
		UpdateWithRecreate: func(key string, oldValue, newValue proto.Message, metadata Metadata) bool {
			return key == prefixC+baseValue3
		},
		WithMetadata:         true,
		RetrieveDependencies: []string{descriptor2Name},
	}, mockSB, 0)

	// register all 3 descriptors with the scheduler
	scheduler.RegisterKVDescriptor(descriptor1)
	scheduler.RegisterKVDescriptor(descriptor2)
	scheduler.RegisterKVDescriptor(descriptor3)

	// get metadata map created for each descriptor
	metadataMap := scheduler.GetMetadataMap(descriptor1.Name)
	nameToInteger1, withMetadataMap := metadataMap.(test.NameToInteger)
	Expect(withMetadataMap).To(BeTrue())
	metadataMap = scheduler.GetMetadataMap(descriptor2.Name)
	nameToInteger2, withMetadataMap := metadataMap.(test.NameToInteger)
	Expect(withMetadataMap).To(BeTrue())
	metadataMap = scheduler.GetMetadataMap(descriptor3.Name)
	nameToInteger3, withMetadataMap := metadataMap.(test.NameToInteger)
	Expect(withMetadataMap).To(BeTrue())

	// run non-resync transaction against empty SB
	startTime := time.Now()
	schedulerTxn := scheduler.StartNBTransaction()
	schedulerTxn.SetValue(prefixB+baseValue2, test.NewArrayValue("item1", "item2"))
	schedulerTxn.SetValue(prefixA+baseValue1, test.NewArrayValue("item2"))
	schedulerTxn.SetValue(prefixC+baseValue3, test.NewArrayValue("item1", "item2"))
	description := "testing data change"
	seqNum, err := schedulerTxn.Commit(WithDescription(testCtx, description))
	stopTime := time.Now()
	Expect(seqNum).To(BeEquivalentTo(0))
	Expect(err).ShouldNot(HaveOccurred())

	// check the state of SB
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	// -> base value 1
	value := mockSB.GetValue(prefixA + baseValue1)
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewArrayValue("item2"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(0))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item1 derived from base value was not added
	value = mockSB.GetValue(prefixA + baseValue1 + "/item1")
	Expect(value).To(BeNil())
	// -> item2 derived from base value 1
	value = mockSB.GetValue(prefixA + baseValue1 + "/item2")
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("item2"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> base value 2
	value = mockSB.GetValue(prefixB + baseValue2)
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewArrayValue("item1", "item2"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(0))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item1 derived from base value 2
	value = mockSB.GetValue(prefixB + baseValue2 + "/item1")
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("item1"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item2 derived from base value 2 is pending
	value = mockSB.GetValue(prefixB + baseValue2 + "/item2")
	Expect(value).To(BeNil())
	// -> base value 3
	value = mockSB.GetValue(prefixC + baseValue3)
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewArrayValue("item1", "item2"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(0))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item1 derived from base value 3
	value = mockSB.GetValue(prefixC + baseValue3 + "/item1")
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("item1"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item2 derived from base value 3
	value = mockSB.GetValue(prefixC + baseValue3 + "/item2")
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("item2"))).To(BeTrue())
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
	Expect(metadata.GetInteger()).To(BeEquivalentTo(0))

	// check operations executed in SB
	opHistory := mockSB.PopHistoryOfOps()
	Expect(opHistory).To(HaveLen(7))
	operation := opHistory[0]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[1]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item2"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[2]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixB + baseValue2))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[3]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixB + baseValue2 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[4]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[5]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[6]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3 + "/item2"))
	Expect(operation.Err).To(BeNil())

	// check transaction operations
	txnHistory := scheduler.GetTransactionHistory(time.Time{}, time.Now())
	Expect(txnHistory).To(HaveLen(1))
	txn := txnHistory[0]
	Expect(txn.PreRecord).To(BeFalse())
	Expect(txn.Start.After(startTime)).To(BeTrue())
	Expect(txn.Start.Before(txn.Stop)).To(BeTrue())
	Expect(txn.Stop.Before(stopTime)).To(BeTrue())
	Expect(txn.SeqNum).To(BeEquivalentTo(0))
	Expect(txn.TxnType).To(BeEquivalentTo(NBTransaction))
	Expect(txn.ResyncType).To(BeEquivalentTo(NotResync))
	Expect(txn.Description).To(Equal(description))
	checkRecordedValues(txn.Values, []RecordedKVPair{
		{Key: prefixA + baseValue1, Value: utils.RecordProtoMessage(test.NewArrayValue("item2")), Origin: FromNB},
		{Key: prefixB + baseValue2, Value: utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")), Origin: FromNB},
		{Key: prefixC + baseValue3, Value: utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")), Origin: FromNB},
	})

	txnOps := RecordedTxnOps{
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixA + baseValue1,
			NewValue:  utils.RecordProtoMessage(test.NewArrayValue("item2")),
			PrevState: ValueState_NONEXISTENT,
			NewState:  ValueState_CONFIGURED,
		},
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixA + baseValue1 + "/item2",
			IsDerived: true,
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("item2")),
			PrevState: ValueState_NONEXISTENT,
			NewState:  ValueState_CONFIGURED,
		},
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixB + baseValue2,
			NewValue:  utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")),
			PrevState: ValueState_NONEXISTENT,
			NewState:  ValueState_CONFIGURED,
		},
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixB + baseValue2 + "/item1",
			IsDerived: true,
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("item1")),
			PrevState: ValueState_NONEXISTENT,
			NewState:  ValueState_CONFIGURED,
		},
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixB + baseValue2 + "/item2",
			IsDerived: true,
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("item2")),
			PrevState: ValueState_NONEXISTENT,
			NewState:  ValueState_PENDING,
			NOOP:      true,
		},
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixC + baseValue3,
			NewValue:  utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")),
			PrevState: ValueState_NONEXISTENT,
			NewState:  ValueState_CONFIGURED,
		},
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixC + baseValue3 + "/item1",
			IsDerived: true,
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("item1")),
			PrevState: ValueState_NONEXISTENT,
			NewState:  ValueState_CONFIGURED,
		},
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixC + baseValue3 + "/item2",
			IsDerived: true,
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("item2")),
			PrevState: ValueState_NONEXISTENT,
			NewState:  ValueState_CONFIGURED,
		},
	}
	checkTxnOperations(txn.Planned, txnOps)
	checkTxnOperations(txn.Executed, txnOps)

	// check flag stats
	graphR := scheduler.graph.Read()
	errorStats := graphR.GetFlagStats(ErrorFlagIndex, nil)
	Expect(errorStats.TotalCount).To(BeEquivalentTo(0))
	pendingStats := graphR.GetFlagStats(UnavailValueFlagIndex, nil)
	Expect(pendingStats.TotalCount).To(BeEquivalentTo(1))
	derivedStats := graphR.GetFlagStats(DerivedFlagIndex, nil)
	Expect(derivedStats.TotalCount).To(BeEquivalentTo(5))
	lastUpdateStats := graphR.GetFlagStats(LastUpdateFlagIndex, nil)
	Expect(lastUpdateStats.TotalCount).To(BeEquivalentTo(8))
	descriptorStats := graphR.GetFlagStats(DescriptorFlagIndex, nil)
	Expect(descriptorStats.TotalCount).To(BeEquivalentTo(8))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor1Name))
	Expect(descriptorStats.PerValueCount[descriptor1Name]).To(BeEquivalentTo(2))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor2Name))
	Expect(descriptorStats.PerValueCount[descriptor2Name]).To(BeEquivalentTo(3))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor3Name))
	Expect(descriptorStats.PerValueCount[descriptor3Name]).To(BeEquivalentTo(3))
	valueStateStats := graphR.GetFlagStats(ValueStateFlagIndex, nil)
	Expect(valueStateStats.TotalCount).To(BeEquivalentTo(8))
	Expect(valueStateStats.PerValueCount).To(HaveKey(ValueState_CONFIGURED.String()))
	Expect(valueStateStats.PerValueCount[ValueState_CONFIGURED.String()]).To(BeEquivalentTo(7))
	Expect(valueStateStats.PerValueCount).To(HaveKey(ValueState_PENDING.String()))
	Expect(valueStateStats.PerValueCount[ValueState_PENDING.String()]).To(BeEquivalentTo(1))
	graphR.Release()

	// check value dumps
	views := []View{NBView, SBView, CachedView}
	for _, view := range views {
		// descriptor1
		expValues := []KVWithMetadata{
			{Key: prefixA + baseValue1, Value: test.NewArrayValue("item2"), Origin: FromNB, Metadata: &test.OnlyInteger{Integer: 0}},
		}
		dumpedValues, err := scheduler.DumpValuesByKeyPrefix(prefixA, view)
		Expect(err).To(BeNil())
		checkValues(dumpedValues, expValues)
		dumpedValues, err = scheduler.DumpValuesByDescriptor(descriptor1Name, view)
		Expect(err).To(BeNil())
		checkValues(dumpedValues, expValues)
		// descriptor2
		expValues = []KVWithMetadata{
			{Key: prefixB + baseValue2, Value: test.NewArrayValue("item1", "item2"), Origin: FromNB, Metadata: &test.OnlyInteger{Integer: 0}},
		}
		dumpedValues, err = scheduler.DumpValuesByKeyPrefix(prefixB, view)
		Expect(err).To(BeNil())
		checkValues(dumpedValues, expValues)
		dumpedValues, err = scheduler.DumpValuesByDescriptor(descriptor2Name, view)
		Expect(err).To(BeNil())
		checkValues(dumpedValues, expValues)
		// descriptor3
		expValues = []KVWithMetadata{
			{Key: prefixC + baseValue3, Value: test.NewArrayValue("item1", "item2"), Origin: FromNB, Metadata: &test.OnlyInteger{Integer: 0}},
		}
		dumpedValues, err = scheduler.DumpValuesByKeyPrefix(prefixC, view)
		Expect(err).To(BeNil())
		checkValues(dumpedValues, expValues)
		dumpedValues, err = scheduler.DumpValuesByDescriptor(descriptor3Name, view)
		Expect(err).To(BeNil())
		checkValues(dumpedValues, expValues)
	}
	mockSB.PopHistoryOfOps() // remove Retrieve-s from the history

	// check value states
	status := scheduler.GetValueStatus(prefixA + baseValue1)
	Expect(status).ToNot(BeNil())
	checkBaseValueStatus(status, &BaseValueStatus{
		Value: &ValueStatus{
			Key:           prefixA + baseValue1,
			State:         ValueState_CONFIGURED,
			LastOperation: TxnOperation_CREATE,
		},
		DerivedValues: []*ValueStatus{
			{
				Key:           prefixA + baseValue1 + "/item2",
				State:         ValueState_CONFIGURED,
				LastOperation: TxnOperation_CREATE,
			},
		},
	})
	status = scheduler.GetValueStatus(prefixB + baseValue2)
	Expect(status).ToNot(BeNil())
	checkBaseValueStatus(status, &BaseValueStatus{
		Value: &ValueStatus{
			Key:           prefixB + baseValue2,
			State:         ValueState_CONFIGURED,
			LastOperation: TxnOperation_CREATE,
		},
		DerivedValues: []*ValueStatus{
			{
				Key:           prefixB + baseValue2 + "/item1",
				State:         ValueState_CONFIGURED,
				LastOperation: TxnOperation_CREATE,
			},
			{
				Key:           prefixB + baseValue2 + "/item2",
				State:         ValueState_PENDING,
				LastOperation: TxnOperation_CREATE,
				Details:       []string{prefixA + baseValue1 + "/item1"},
			},
		},
	})
	status = scheduler.GetValueStatus(prefixC + baseValue3)
	Expect(status).ToNot(BeNil())
	checkBaseValueStatus(status, &BaseValueStatus{
		Value: &ValueStatus{
			Key:           prefixC + baseValue3,
			State:         ValueState_CONFIGURED,
			LastOperation: TxnOperation_CREATE,
		},
		DerivedValues: []*ValueStatus{
			{
				Key:           prefixC + baseValue3 + "/item1",
				State:         ValueState_CONFIGURED,
				LastOperation: TxnOperation_CREATE,
			},
			{
				Key:           prefixC + baseValue3 + "/item2",
				State:         ValueState_CONFIGURED,
				LastOperation: TxnOperation_CREATE,
			},
		},
	})

	// run 2nd non-resync transaction
	startTime = time.Now()
	schedulerTxn2 := scheduler.StartNBTransaction()
	schedulerTxn2.SetValue(prefixC+baseValue3, test.NewArrayValue("item1"))
	schedulerTxn2.SetValue(prefixA+baseValue1, test.NewArrayValue("item1"))
	seqNum, err = schedulerTxn2.Commit(testCtx)
	stopTime = time.Now()
	Expect(seqNum).To(BeEquivalentTo(1))
	Expect(err).ShouldNot(HaveOccurred())

	// check the state of SB
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	// -> base value 1
	value = mockSB.GetValue(prefixA + baseValue1)
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewArrayValue("item1"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(0))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item1 derived from base value was added
	value = mockSB.GetValue(prefixA + baseValue1 + "/item1")
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("item1"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item2 derived from base value 1 was deleted
	value = mockSB.GetValue(prefixA + baseValue1 + "/item2")
	Expect(value).To(BeNil())
	// -> base value 2
	value = mockSB.GetValue(prefixB + baseValue2)
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewArrayValue("item1", "item2"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(0))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item1 derived from base value 2
	value = mockSB.GetValue(prefixB + baseValue2 + "/item1")
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("item1"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item2 derived from base value 2 is no longer pending
	value = mockSB.GetValue(prefixB + baseValue2 + "/item2")
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("item2"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> base value 3
	value = mockSB.GetValue(prefixC + baseValue3)
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewArrayValue("item1"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(1))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item1 derived from base value 3
	value = mockSB.GetValue(prefixC + baseValue3 + "/item1")
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("item1"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item2 derived from base value 3 was deleted
	value = mockSB.GetValue(prefixC + baseValue3 + "/item2")
	Expect(value).To(BeNil())

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
	Expect(opHistory).To(HaveLen(9))
	operation = opHistory[0]
	Expect(operation.OpType).To(Equal(test.MockDelete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[1]
	Expect(operation.OpType).To(Equal(test.MockDelete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3 + "/item2"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[2]
	Expect(operation.OpType).To(Equal(test.MockDelete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[3]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[4]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[5]
	Expect(operation.OpType).To(Equal(test.MockDelete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item2"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[6]
	Expect(operation.OpType).To(Equal(test.MockUpdate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[7]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[8]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixB + baseValue2 + "/item2"))
	Expect(operation.Err).To(BeNil())

	// check transaction operations
	txnHistory = scheduler.GetTransactionHistory(startTime, stopTime) // first txn not included
	Expect(txnHistory).To(HaveLen(1))
	txn = txnHistory[0]
	Expect(txn.PreRecord).To(BeFalse())
	Expect(txn.Start.After(startTime)).To(BeTrue())
	Expect(txn.Start.Before(txn.Stop)).To(BeTrue())
	Expect(txn.Stop.Before(stopTime)).To(BeTrue())
	Expect(txn.SeqNum).To(BeEquivalentTo(1))
	Expect(txn.TxnType).To(BeEquivalentTo(NBTransaction))
	Expect(txn.ResyncType).To(BeEquivalentTo(NotResync))
	Expect(txn.Description).To(BeEmpty())
	checkRecordedValues(txn.Values, []RecordedKVPair{
		{Key: prefixA + baseValue1, Value: utils.RecordProtoMessage(test.NewArrayValue("item1")), Origin: FromNB},
		{Key: prefixC + baseValue3, Value: utils.RecordProtoMessage(test.NewArrayValue("item1")), Origin: FromNB},
	})

	txnOps = RecordedTxnOps{
		{
			Operation:  TxnOperation_DELETE,
			Key:        prefixC + baseValue3 + "/item1",
			IsDerived:  true,
			PrevValue:  utils.RecordProtoMessage(test.NewStringValue("item1")),
			PrevState:  ValueState_CONFIGURED,
			NewState:   ValueState_REMOVED,
			IsRecreate: true,
		},
		{
			Operation:  TxnOperation_DELETE,
			Key:        prefixC + baseValue3 + "/item2",
			IsDerived:  true,
			PrevValue:  utils.RecordProtoMessage(test.NewStringValue("item2")),
			PrevState:  ValueState_CONFIGURED,
			NewState:   ValueState_REMOVED,
			IsRecreate: true,
		},
		{
			Operation:  TxnOperation_DELETE,
			Key:        prefixC + baseValue3,
			PrevValue:  utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")),
			PrevState:  ValueState_CONFIGURED,
			NewState:   ValueState_REMOVED,
			IsRecreate: true,
		},
		{
			Operation:  TxnOperation_CREATE,
			Key:        prefixC + baseValue3,
			NewValue:   utils.RecordProtoMessage(test.NewArrayValue("item1")),
			PrevState:  ValueState_REMOVED,
			NewState:   ValueState_CONFIGURED,
			IsRecreate: true,
		},
		{
			Operation:  TxnOperation_CREATE,
			Key:        prefixC + baseValue3 + "/item1",
			IsDerived:  true,
			NewValue:   utils.RecordProtoMessage(test.NewStringValue("item1")),
			PrevState:  ValueState_NONEXISTENT, // TODO: derived value removed from the graph, ok?
			NewState:   ValueState_CONFIGURED,
			IsRecreate: true,
		},
		{
			Operation: TxnOperation_DELETE,
			Key:       prefixA + baseValue1 + "/item2",
			IsDerived: true,
			PrevValue: utils.RecordProtoMessage(test.NewStringValue("item2")),
			PrevState: ValueState_CONFIGURED,
			NewState:  ValueState_REMOVED,
		},
		{
			Operation: TxnOperation_UPDATE,
			Key:       prefixA + baseValue1,
			PrevValue: utils.RecordProtoMessage(test.NewArrayValue("item2")),
			NewValue:  utils.RecordProtoMessage(test.NewArrayValue("item1")),
			PrevState: ValueState_CONFIGURED,
			NewState:  ValueState_CONFIGURED,
		},
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixA + baseValue1 + "/item1",
			IsDerived: true,
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("item1")),
			PrevState: ValueState_NONEXISTENT,
			NewState:  ValueState_CONFIGURED,
		},
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixB + baseValue2 + "/item2",
			IsDerived: true,
			PrevValue: utils.RecordProtoMessage(test.NewStringValue("item2")),
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("item2")),
			PrevState: ValueState_PENDING,
			NewState:  ValueState_CONFIGURED,
		},
	}
	checkTxnOperations(txn.Planned, txnOps)
	checkTxnOperations(txn.Executed, txnOps)

	// check value states
	status = scheduler.GetValueStatus(prefixA + baseValue1)
	Expect(status).ToNot(BeNil())
	checkBaseValueStatus(status, &BaseValueStatus{
		Value: &ValueStatus{
			Key:           prefixA + baseValue1,
			State:         ValueState_CONFIGURED,
			LastOperation: TxnOperation_UPDATE,
		},
		DerivedValues: []*ValueStatus{
			{
				Key:           prefixA + baseValue1 + "/item1",
				State:         ValueState_CONFIGURED,
				LastOperation: TxnOperation_CREATE,
			},
		},
	})
	status = scheduler.GetValueStatus(prefixB + baseValue2)
	Expect(status).ToNot(BeNil())
	checkBaseValueStatus(status, &BaseValueStatus{
		Value: &ValueStatus{
			Key:           prefixB + baseValue2,
			State:         ValueState_CONFIGURED,
			LastOperation: TxnOperation_CREATE,
		},
		DerivedValues: []*ValueStatus{
			{
				Key:           prefixB + baseValue2 + "/item1",
				State:         ValueState_CONFIGURED,
				LastOperation: TxnOperation_CREATE,
			},
			{
				Key:           prefixB + baseValue2 + "/item2",
				State:         ValueState_CONFIGURED,
				LastOperation: TxnOperation_CREATE,
			},
		},
	})
	status = scheduler.GetValueStatus(prefixC + baseValue3)
	Expect(status).ToNot(BeNil())
	checkBaseValueStatus(status, &BaseValueStatus{
		Value: &ValueStatus{
			Key:           prefixC + baseValue3,
			State:         ValueState_CONFIGURED,
			LastOperation: TxnOperation_UPDATE,
		},
		DerivedValues: []*ValueStatus{
			{
				Key:           prefixC + baseValue3 + "/item1",
				State:         ValueState_CONFIGURED,
				LastOperation: TxnOperation_CREATE,
			},
		},
	})

	// check flag stats
	graphR = scheduler.graph.Read()
	errorStats = graphR.GetFlagStats(ErrorFlagIndex, nil)
	Expect(errorStats.TotalCount).To(BeEquivalentTo(0))
	pendingStats = graphR.GetFlagStats(UnavailValueFlagIndex, nil)
	Expect(pendingStats.TotalCount).To(BeEquivalentTo(1))
	derivedStats = graphR.GetFlagStats(DerivedFlagIndex, nil)
	Expect(derivedStats.TotalCount).To(BeEquivalentTo(8))
	lastUpdateStats = graphR.GetFlagStats(LastUpdateFlagIndex, nil)
	Expect(lastUpdateStats.TotalCount).To(BeEquivalentTo(13))
	descriptorStats = graphR.GetFlagStats(DescriptorFlagIndex, nil)
	Expect(descriptorStats.TotalCount).To(BeEquivalentTo(13))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor1Name))
	Expect(descriptorStats.PerValueCount[descriptor1Name]).To(BeEquivalentTo(4))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor2Name))
	Expect(descriptorStats.PerValueCount[descriptor2Name]).To(BeEquivalentTo(4))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor3Name))
	Expect(descriptorStats.PerValueCount[descriptor3Name]).To(BeEquivalentTo(5))
	valueStateStats = graphR.GetFlagStats(ValueStateFlagIndex, nil)
	Expect(valueStateStats.TotalCount).To(BeEquivalentTo(13))
	Expect(valueStateStats.PerValueCount).To(HaveKey(ValueState_CONFIGURED.String()))
	Expect(valueStateStats.PerValueCount[ValueState_CONFIGURED.String()]).To(BeEquivalentTo(12))
	Expect(valueStateStats.PerValueCount).To(HaveKey(ValueState_PENDING.String()))
	Expect(valueStateStats.PerValueCount[ValueState_PENDING.String()]).To(BeEquivalentTo(1))
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
	descriptor1 := test.NewMockDescriptor(&KVDescriptor{
		Name:          descriptor1Name,
		NBKeyPrefix:   prefixA,
		KeySelector:   prefixSelector(prefixA),
		ValueTypeName: proto.MessageName(test.NewArrayValue()),
		DerivedValues: test.ArrayValueDerBuilder,
		WithMetadata:  true,
	}, mockSB, 0)
	// -> descriptor2:
	descriptor2 := test.NewMockDescriptor(&KVDescriptor{
		Name:          descriptor2Name,
		NBKeyPrefix:   prefixB,
		KeySelector:   prefixSelector(prefixB),
		ValueTypeName: proto.MessageName(test.NewArrayValue()),
		DerivedValues: test.ArrayValueDerBuilder,
		Dependencies: func(key string, value proto.Message) []Dependency {
			if key == prefixB+baseValue2+"/item1" {
				depKey := prefixA + baseValue1
				return []Dependency{
					{Label: depKey, Key: depKey},
				}
			}
			if key == prefixB+baseValue2+"/item2" {
				depKey := prefixA + baseValue1 + "/item1"
				return []Dependency{
					{Label: depKey, Key: depKey},
				}
			}
			return nil
		},
		WithMetadata:         true,
		RetrieveDependencies: []string{descriptor1Name},
	}, mockSB, 0)
	// -> descriptor3:
	descriptor3 := test.NewMockDescriptor(&KVDescriptor{
		Name:          descriptor3Name,
		NBKeyPrefix:   prefixC,
		KeySelector:   prefixSelector(prefixC),
		ValueTypeName: proto.MessageName(test.NewArrayValue()),
		DerivedValues: test.ArrayValueDerBuilder,
		UpdateWithRecreate: func(key string, oldValue, newValue proto.Message, metadata Metadata) bool {
			return key == prefixC+baseValue3
		},
		WithMetadata:         true,
		RetrieveDependencies: []string{descriptor2Name},
	}, mockSB, 0)

	// register all 3 descriptors with the scheduler
	scheduler.RegisterKVDescriptor(descriptor1)
	scheduler.RegisterKVDescriptor(descriptor2)
	scheduler.RegisterKVDescriptor(descriptor3)

	// get metadata map created for each descriptor
	metadataMap := scheduler.GetMetadataMap(descriptor1.Name)
	nameToInteger1, withMetadataMap := metadataMap.(test.NameToInteger)
	Expect(withMetadataMap).To(BeTrue())
	metadataMap = scheduler.GetMetadataMap(descriptor2.Name)
	nameToInteger2, withMetadataMap := metadataMap.(test.NameToInteger)
	Expect(withMetadataMap).To(BeTrue())
	metadataMap = scheduler.GetMetadataMap(descriptor3.Name)
	nameToInteger3, withMetadataMap := metadataMap.(test.NameToInteger)
	Expect(withMetadataMap).To(BeTrue())

	// run 1st non-resync transaction against empty SB
	schedulerTxn := scheduler.StartNBTransaction()
	schedulerTxn.SetValue(prefixB+baseValue2, test.NewArrayValue("item1", "item2"))
	schedulerTxn.SetValue(prefixA+baseValue1, test.NewArrayValue("item2"))
	schedulerTxn.SetValue(prefixC+baseValue3, test.NewArrayValue("item1", "item2"))
	seqNum, err := schedulerTxn.Commit(testCtx)
	Expect(seqNum).To(BeEquivalentTo(0))
	Expect(err).ShouldNot(HaveOccurred())
	mockSB.PopHistoryOfOps()

	// plan error before 2nd txn
	failedModifyClb := func() {
		mockSB.SetValue(prefixA+baseValue1, test.NewArrayValue(),
			&test.OnlyInteger{Integer: 0}, FromNB, false)
	}
	mockSB.PlanError(prefixA+baseValue1, errors.New("failed to modify value"), failedModifyClb)

	// subscribe to receive notifications about value state changes for prefixA
	statusChan := make(chan *BaseValueStatus, 5)
	scheduler.WatchValueStatus(statusChan, prefixSelector(prefixA))

	// run 2nd non-resync transaction against empty SB that will fail and will be reverted
	startTime := time.Now()
	schedulerTxn2 := scheduler.StartNBTransaction()
	schedulerTxn2.SetValue(prefixC+baseValue3, test.NewArrayValue("item1"))
	schedulerTxn2.SetValue(prefixA+baseValue1, test.NewArrayValue("item1"))
	seqNum, err = schedulerTxn2.Commit(WithRevert(testCtx))
	stopTime := time.Now()
	Expect(seqNum).To(BeEquivalentTo(1))
	Expect(err).ToNot(BeNil())
	txnErr := err.(*TransactionError)
	Expect(txnErr.GetTxnInitError()).ShouldNot(HaveOccurred())
	kvErrors := txnErr.GetKVErrors()
	Expect(kvErrors).To(HaveLen(1))
	Expect(kvErrors[0].Key).To(BeEquivalentTo(prefixA + baseValue1))
	Expect(kvErrors[0].TxnOperation).To(BeEquivalentTo(TxnOperation_UPDATE))
	Expect(kvErrors[0].Error.Error()).To(BeEquivalentTo("failed to modify value"))

	// receive the status update with the value reverted back to the original
	var valueStatus *BaseValueStatus
	Eventually(statusChan, time.Second).Should(Receive(&valueStatus))
	checkBaseValueStatus(valueStatus, &BaseValueStatus{
		Value: &ValueStatus{
			Key:           prefixA + baseValue1,
			State:         ValueState_CONFIGURED,
			LastOperation: TxnOperation_UPDATE,
		},
		DerivedValues: []*ValueStatus{
			{
				Key:           prefixA + baseValue1 + "/item2",
				State:         ValueState_CONFIGURED,
				LastOperation: TxnOperation_CREATE,
			},
		},
	})

	// check the state of SB
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	// -> base value 1
	value := mockSB.GetValue(prefixA + baseValue1)
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewArrayValue("item2"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(0))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item1 derived from base value was NOT added
	value = mockSB.GetValue(prefixA + baseValue1 + "/item1")
	Expect(value).To(BeNil())
	// -> item2 derived from base value 1 was first deleted by then added back
	value = mockSB.GetValue(prefixA + baseValue1 + "/item2")
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("item2"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> base value 2
	value = mockSB.GetValue(prefixB + baseValue2)
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewArrayValue("item1", "item2"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(0))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item1 derived from base value 2
	value = mockSB.GetValue(prefixB + baseValue2 + "/item1")
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("item1"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item2 derived from base value 2 is still pending
	value = mockSB.GetValue(prefixB + baseValue2 + "/item2")
	Expect(value).To(BeNil())
	// -> base value 3 was reverted back to state after 1st txn
	value = mockSB.GetValue(prefixC + baseValue3)
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewArrayValue("item1", "item2"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(2))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item1 derived from base value 3
	value = mockSB.GetValue(prefixC + baseValue3 + "/item1")
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("item1"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item2 derived from base value 3
	value = mockSB.GetValue(prefixC + baseValue3 + "/item2")
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("item2"))).To(BeTrue())
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
	Expect(opHistory).To(HaveLen(15))
	operation := opHistory[0]
	Expect(operation.OpType).To(Equal(test.MockDelete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[1]
	Expect(operation.OpType).To(Equal(test.MockDelete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3 + "/item2"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[2]
	Expect(operation.OpType).To(Equal(test.MockDelete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[3]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[4]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[5]
	Expect(operation.OpType).To(Equal(test.MockDelete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item2"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[6]
	Expect(operation.OpType).To(Equal(test.MockUpdate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1))
	Expect(operation.Err).ToNot(BeNil())
	Expect(operation.Err.Error()).To(BeEquivalentTo("failed to modify value"))
	// reverting:
	operation = opHistory[7] // refresh failed value
	Expect(operation.OpType).To(Equal(test.MockRetrieve))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	checkValues(operation.CorrelateRetrieve, []KVWithMetadata{
		{
			Key:      prefixA + baseValue1,
			Value:    test.NewArrayValue("item1"),
			Metadata: &test.OnlyInteger{Integer: 0},
			Origin:   FromNB,
		},
	})
	operation = opHistory[8]
	Expect(operation.OpType).To(Equal(test.MockUpdate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[9]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item2"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[10]
	Expect(operation.OpType).To(Equal(test.MockDelete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[11]
	Expect(operation.OpType).To(Equal(test.MockDelete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[12]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[13]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[14]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3 + "/item2"))
	Expect(operation.Err).To(BeNil())

	// check transaction operations
	txnHistory := scheduler.GetTransactionHistory(startTime, time.Now())
	Expect(txnHistory).To(HaveLen(1))
	txn := txnHistory[0]
	Expect(txn.PreRecord).To(BeFalse())
	Expect(txn.Start.After(startTime)).To(BeTrue())
	Expect(txn.Start.Before(txn.Stop)).To(BeTrue())
	Expect(txn.Stop.Before(stopTime)).To(BeTrue())
	Expect(txn.SeqNum).To(BeEquivalentTo(1))
	Expect(txn.TxnType).To(BeEquivalentTo(NBTransaction))
	Expect(txn.ResyncType).To(BeEquivalentTo(NotResync))
	Expect(txn.Description).To(BeEmpty())
	checkRecordedValues(txn.Values, []RecordedKVPair{
		{Key: prefixA + baseValue1, Value: utils.RecordProtoMessage(test.NewArrayValue("item1")), Origin: FromNB},
		{Key: prefixC + baseValue3, Value: utils.RecordProtoMessage(test.NewArrayValue("item1")), Origin: FromNB},
	})

	// planned operations
	txnOps := RecordedTxnOps{
		{
			Operation:  TxnOperation_DELETE,
			Key:        prefixC + baseValue3 + "/item1",
			IsDerived:  true,
			PrevValue:  utils.RecordProtoMessage(test.NewStringValue("item1")),
			PrevState:  ValueState_CONFIGURED,
			NewState:   ValueState_REMOVED,
			IsRecreate: true,
		},
		{
			Operation:  TxnOperation_DELETE,
			Key:        prefixC + baseValue3 + "/item2",
			IsDerived:  true,
			PrevValue:  utils.RecordProtoMessage(test.NewStringValue("item2")),
			PrevState:  ValueState_CONFIGURED,
			NewState:   ValueState_REMOVED,
			IsRecreate: true,
		},
		{
			Operation:  TxnOperation_DELETE,
			Key:        prefixC + baseValue3,
			PrevValue:  utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")),
			PrevState:  ValueState_CONFIGURED,
			NewState:   ValueState_REMOVED,
			IsRecreate: true,
		},
		{
			Operation:  TxnOperation_CREATE,
			Key:        prefixC + baseValue3,
			NewValue:   utils.RecordProtoMessage(test.NewArrayValue("item1")),
			PrevState:  ValueState_REMOVED,
			NewState:   ValueState_CONFIGURED,
			IsRecreate: true,
		},
		{
			Operation:  TxnOperation_CREATE,
			Key:        prefixC + baseValue3 + "/item1",
			IsDerived:  true,
			NewValue:   utils.RecordProtoMessage(test.NewStringValue("item1")),
			PrevState:  ValueState_NONEXISTENT,
			NewState:   ValueState_CONFIGURED,
			IsRecreate: true,
		},
		{
			Operation: TxnOperation_DELETE,
			Key:       prefixA + baseValue1 + "/item2",
			IsDerived: true,
			PrevValue: utils.RecordProtoMessage(test.NewStringValue("item2")),
			PrevState: ValueState_CONFIGURED,
			NewState:  ValueState_REMOVED,
		},
		{
			Operation: TxnOperation_UPDATE,
			Key:       prefixA + baseValue1,
			PrevValue: utils.RecordProtoMessage(test.NewArrayValue("item2")),
			NewValue:  utils.RecordProtoMessage(test.NewArrayValue("item1")),
			PrevState: ValueState_CONFIGURED,
			NewState:  ValueState_CONFIGURED,
		},
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixA + baseValue1 + "/item1",
			IsDerived: true,
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("item1")),
			PrevState: ValueState_NONEXISTENT,
			NewState:  ValueState_CONFIGURED,
		},
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixB + baseValue2 + "/item2",
			IsDerived: true,
			PrevValue: utils.RecordProtoMessage(test.NewStringValue("item2")),
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("item2")),
			PrevState: ValueState_PENDING,
			NewState:  ValueState_CONFIGURED,
		},
	}
	checkTxnOperations(txn.Planned, txnOps)

	// executed operations
	txnOps = RecordedTxnOps{
		{
			Operation:  TxnOperation_DELETE,
			Key:        prefixC + baseValue3 + "/item1",
			IsDerived:  true,
			PrevValue:  utils.RecordProtoMessage(test.NewStringValue("item1")),
			PrevState:  ValueState_CONFIGURED,
			NewState:   ValueState_REMOVED,
			IsRecreate: true,
		},
		{
			Operation:  TxnOperation_DELETE,
			Key:        prefixC + baseValue3 + "/item2",
			IsDerived:  true,
			PrevValue:  utils.RecordProtoMessage(test.NewStringValue("item2")),
			PrevState:  ValueState_CONFIGURED,
			NewState:   ValueState_REMOVED,
			IsRecreate: true,
		},
		{
			Operation:  TxnOperation_DELETE,
			Key:        prefixC + baseValue3,
			PrevValue:  utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")),
			PrevState:  ValueState_CONFIGURED,
			NewState:   ValueState_REMOVED,
			IsRecreate: true,
		},
		{
			Operation:  TxnOperation_CREATE,
			Key:        prefixC + baseValue3,
			NewValue:   utils.RecordProtoMessage(test.NewArrayValue("item1")),
			PrevState:  ValueState_REMOVED,
			NewState:   ValueState_CONFIGURED,
			IsRecreate: true,
		},
		{
			Operation:  TxnOperation_CREATE,
			Key:        prefixC + baseValue3 + "/item1",
			IsDerived:  true,
			NewValue:   utils.RecordProtoMessage(test.NewStringValue("item1")),
			PrevState:  ValueState_NONEXISTENT,
			NewState:   ValueState_CONFIGURED,
			IsRecreate: true,
		},
		{
			Operation: TxnOperation_DELETE,
			Key:       prefixA + baseValue1 + "/item2",
			IsDerived: true,
			PrevValue: utils.RecordProtoMessage(test.NewStringValue("item2")),
			PrevState: ValueState_CONFIGURED,
			NewState:  ValueState_REMOVED,
		},
		{
			Operation: TxnOperation_UPDATE,
			Key:       prefixA + baseValue1,
			PrevValue: utils.RecordProtoMessage(test.NewArrayValue("item2")),
			NewValue:  utils.RecordProtoMessage(test.NewArrayValue("item1")),
			PrevState: ValueState_CONFIGURED,
			NewState:  ValueState_FAILED,
			NewErr:    errors.New("failed to modify value"),
		},
		// reverting:
		{
			Operation: TxnOperation_UPDATE,
			Key:       prefixA + baseValue1,
			PrevValue: utils.RecordProtoMessage(test.NewArrayValue()),
			NewValue:  utils.RecordProtoMessage(test.NewArrayValue("item2")),
			PrevState: ValueState_FAILED,
			NewState:  ValueState_CONFIGURED,
			PrevErr:   errors.New("failed to modify value"),
			IsRevert:  true,
		},
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixA + baseValue1 + "/item2",
			IsDerived: true,
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("item2")),
			PrevState: ValueState_NONEXISTENT,
			NewState:  ValueState_CONFIGURED,
			IsRevert:  true,
		},
		{
			Operation:  TxnOperation_DELETE,
			Key:        prefixC + baseValue3 + "/item1",
			IsDerived:  true,
			PrevValue:  utils.RecordProtoMessage(test.NewStringValue("item1")),
			PrevState:  ValueState_CONFIGURED,
			NewState:   ValueState_REMOVED,
			IsRevert:   true,
			IsRecreate: true,
		},
		{
			Operation:  TxnOperation_DELETE,
			Key:        prefixC + baseValue3,
			PrevValue:  utils.RecordProtoMessage(test.NewArrayValue("item1")),
			PrevState:  ValueState_CONFIGURED,
			NewState:   ValueState_REMOVED,
			IsRevert:   true,
			IsRecreate: true,
		},
		{
			Operation:  TxnOperation_CREATE,
			Key:        prefixC + baseValue3,
			NewValue:   utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")),
			PrevState:  ValueState_REMOVED,
			NewState:   ValueState_CONFIGURED,
			IsRevert:   true,
			IsRecreate: true,
		},
		{
			Operation:  TxnOperation_CREATE,
			Key:        prefixC + baseValue3 + "/item1",
			IsDerived:  true,
			NewValue:   utils.RecordProtoMessage(test.NewStringValue("item1")),
			PrevState:  ValueState_NONEXISTENT,
			NewState:   ValueState_CONFIGURED,
			IsRevert:   true,
			IsRecreate: true,
		},
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixC + baseValue3 + "/item2",
			IsDerived: true,
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("item2")),
			PrevState: ValueState_NONEXISTENT,
			NewState:  ValueState_CONFIGURED,
			IsRevert:  true,
		},
	}
	checkTxnOperations(txn.Executed, txnOps)

	// check flag stats
	graphR := scheduler.graph.Read()
	errorStats := graphR.GetFlagStats(ErrorFlagIndex, nil)
	Expect(errorStats.TotalCount).To(BeEquivalentTo(1))
	pendingStats := graphR.GetFlagStats(UnavailValueFlagIndex, nil)
	Expect(pendingStats.TotalCount).To(BeEquivalentTo(1))
	derivedStats := graphR.GetFlagStats(DerivedFlagIndex, nil)
	Expect(derivedStats.TotalCount).To(BeEquivalentTo(9))
	lastUpdateStats := graphR.GetFlagStats(LastUpdateFlagIndex, nil)
	Expect(lastUpdateStats.TotalCount).To(BeEquivalentTo(16))
	descriptorStats := graphR.GetFlagStats(DescriptorFlagIndex, nil)
	Expect(descriptorStats.TotalCount).To(BeEquivalentTo(16))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor1Name))
	Expect(descriptorStats.PerValueCount[descriptor1Name]).To(BeEquivalentTo(5))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor2Name))
	Expect(descriptorStats.PerValueCount[descriptor2Name]).To(BeEquivalentTo(3))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor3Name))
	Expect(descriptorStats.PerValueCount[descriptor3Name]).To(BeEquivalentTo(8))
	valueStateStats := graphR.GetFlagStats(ValueStateFlagIndex, nil)
	Expect(valueStateStats.TotalCount).To(BeEquivalentTo(16))
	Expect(valueStateStats.PerValueCount).To(HaveKey(ValueState_CONFIGURED.String()))
	Expect(valueStateStats.PerValueCount[ValueState_CONFIGURED.String()]).To(BeEquivalentTo(14))
	Expect(valueStateStats.PerValueCount).To(HaveKey(ValueState_FAILED.String()))
	Expect(valueStateStats.PerValueCount[ValueState_FAILED.String()]).To(BeEquivalentTo(1))
	Expect(valueStateStats.PerValueCount).To(HaveKey(ValueState_PENDING.String()))
	Expect(valueStateStats.PerValueCount[ValueState_PENDING.String()]).To(BeEquivalentTo(1))
	graphR.Release()

	// close scheduler
	err = scheduler.Close()
	Expect(err).To(BeNil())
}

func TestDependencyCycles(t *testing.T) {
	RegisterTestingT(t)

	// prepare KV Scheduler
	scheduler := NewPlugin(UseDeps(func(deps *Deps) {
		deps.HTTPHandlers = nil
	}))
	err := scheduler.Init()
	Expect(err).To(BeNil())

	// prepare mocks
	mockSB := test.NewMockSouthbound()
	// -> descriptor:
	descriptor := test.NewMockDescriptor(&KVDescriptor{
		Name:            descriptor1Name,
		KeySelector:     prefixSelector(prefixA),
		NBKeyPrefix:     prefixA,
		ValueTypeName:   proto.MessageName(test.NewStringValue("")),
		ValueComparator: test.StringValueComparator,
		Dependencies: func(key string, value proto.Message) []Dependency {
			if key == prefixA+baseValue1 {
				depKey := prefixA + baseValue2
				return []Dependency{
					{Label: depKey, Key: depKey},
				}
			}
			if key == prefixA+baseValue2 {
				depKey := prefixA + baseValue3
				return []Dependency{
					{Label: depKey, Key: depKey},
				}
			}
			if key == prefixA+baseValue3 {
				depKey1 := prefixA + baseValue1
				depKey2 := prefixA + baseValue4
				return []Dependency{
					{Label: depKey1, Key: depKey1},
					{Label: depKey2, Key: depKey2},
				}
			}
			return nil
		},
		WithMetadata: false,
	}, mockSB, 0, test.WithoutRetrieve)

	// register the descriptor
	scheduler.RegisterKVDescriptor(descriptor)

	// run non-resync transaction against empty SB
	startTime := time.Now()
	schedulerTxn := scheduler.StartNBTransaction()
	schedulerTxn.SetValue(prefixA+baseValue1, test.NewStringValue("base-value1-data"))
	schedulerTxn.SetValue(prefixA+baseValue2, test.NewStringValue("base-value2-data"))
	schedulerTxn.SetValue(prefixA+baseValue3, test.NewStringValue("base-value3-data"))
	description := "testing dependency cycles"
	seqNum, err := schedulerTxn.Commit(WithDescription(testCtx, description))
	stopTime := time.Now()
	Expect(seqNum).To(BeEquivalentTo(0))
	Expect(err).ShouldNot(HaveOccurred())

	// check the state of SB
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	Expect(mockSB.GetValues(nil)).To(HaveLen(0))

	// check value states
	status := scheduler.GetValueStatus(prefixA + baseValue1)
	Expect(status).ToNot(BeNil())
	checkBaseValueStatus(status, &BaseValueStatus{
		Value: &ValueStatus{
			Key:           prefixA + baseValue1,
			State:         ValueState_PENDING,
			LastOperation: TxnOperation_CREATE,
			Details:       []string{prefixA + baseValue2},
		},
	})
	status = scheduler.GetValueStatus(prefixA + baseValue2)
	Expect(status).ToNot(BeNil())
	checkBaseValueStatus(status, &BaseValueStatus{
		Value: &ValueStatus{
			Key:           prefixA + baseValue2,
			State:         ValueState_PENDING,
			LastOperation: TxnOperation_CREATE,
			Details:       []string{prefixA + baseValue3},
		},
	})
	status = scheduler.GetValueStatus(prefixA + baseValue3)
	Expect(status).ToNot(BeNil())
	checkBaseValueStatus(status, &BaseValueStatus{
		Value: &ValueStatus{
			Key:           prefixA + baseValue3,
			State:         ValueState_PENDING,
			LastOperation: TxnOperation_CREATE,
			Details:       []string{prefixA + baseValue4, prefixA + baseValue1},
		},
	})

	// check operations executed in SB
	opHistory := mockSB.PopHistoryOfOps()
	Expect(opHistory).To(HaveLen(0))

	// check transaction operations
	txnHistory := scheduler.GetTransactionHistory(time.Time{}, time.Now())
	Expect(txnHistory).To(HaveLen(1))
	txn := txnHistory[0]
	Expect(txn.PreRecord).To(BeFalse())
	Expect(txn.Start.After(startTime)).To(BeTrue())
	Expect(txn.Start.Before(txn.Stop)).To(BeTrue())
	Expect(txn.Stop.Before(stopTime)).To(BeTrue())
	Expect(txn.SeqNum).To(BeEquivalentTo(0))
	Expect(txn.TxnType).To(BeEquivalentTo(NBTransaction))
	Expect(txn.ResyncType).To(BeEquivalentTo(NotResync))
	Expect(txn.Description).To(Equal(description))
	checkRecordedValues(txn.Values, []RecordedKVPair{
		{Key: prefixA + baseValue1, Value: utils.RecordProtoMessage(test.NewStringValue("base-value1-data")), Origin: FromNB},
		{Key: prefixA + baseValue2, Value: utils.RecordProtoMessage(test.NewStringValue("base-value2-data")), Origin: FromNB},
		{Key: prefixA + baseValue3, Value: utils.RecordProtoMessage(test.NewStringValue("base-value3-data")), Origin: FromNB},
	})

	txnOps := RecordedTxnOps{
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixA + baseValue1,
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("base-value1-data")),
			PrevState: ValueState_NONEXISTENT,
			NewState:  ValueState_PENDING,
			NOOP:      true,
		},
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixA + baseValue2,
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("base-value2-data")),
			PrevState: ValueState_NONEXISTENT,
			NewState:  ValueState_PENDING,
			NOOP:      true,
		},
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixA + baseValue3,
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("base-value3-data")),
			PrevState: ValueState_NONEXISTENT,
			NewState:  ValueState_PENDING,
			NOOP:      true,
		},
	}
	checkTxnOperations(txn.Planned, txnOps)
	checkTxnOperations(txn.Executed, txnOps)

	// check flag stats
	graphR := scheduler.graph.Read()
	errorStats := graphR.GetFlagStats(ErrorFlagIndex, nil)
	Expect(errorStats.TotalCount).To(BeEquivalentTo(0))
	pendingStats := graphR.GetFlagStats(UnavailValueFlagIndex, nil)
	Expect(pendingStats.TotalCount).To(BeEquivalentTo(3))
	derivedStats := graphR.GetFlagStats(DerivedFlagIndex, nil)
	Expect(derivedStats.TotalCount).To(BeEquivalentTo(0))
	lastUpdateStats := graphR.GetFlagStats(LastUpdateFlagIndex, nil)
	Expect(lastUpdateStats.TotalCount).To(BeEquivalentTo(3))
	descriptorStats := graphR.GetFlagStats(DescriptorFlagIndex, nil)
	Expect(descriptorStats.TotalCount).To(BeEquivalentTo(3))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor1Name))
	Expect(descriptorStats.PerValueCount[descriptor1Name]).To(BeEquivalentTo(3))
	valueStateStats := graphR.GetFlagStats(ValueStateFlagIndex, nil)
	Expect(valueStateStats.TotalCount).To(BeEquivalentTo(3))
	Expect(valueStateStats.PerValueCount).To(HaveKey(ValueState_PENDING.String()))
	Expect(valueStateStats.PerValueCount[ValueState_PENDING.String()]).To(BeEquivalentTo(3))
	graphR.Release()

	// run second transaction that will make the cycle of values ready to be added
	startTime = time.Now()
	schedulerTxn = scheduler.StartNBTransaction()
	schedulerTxn.SetValue(prefixA+baseValue4, test.NewStringValue("base-value4-data"))
	seqNum, err = schedulerTxn.Commit(testCtx)
	stopTime = time.Now()
	Expect(seqNum).To(BeEquivalentTo(1))
	Expect(err).ShouldNot(HaveOccurred())

	// check the state of SB
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	Expect(mockSB.GetValues(nil)).To(HaveLen(4))
	// -> base value 1
	value := mockSB.GetValue(prefixA + baseValue1)
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("base-value1-data"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> base value 2
	value = mockSB.GetValue(prefixA + baseValue2)
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("base-value2-data"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> base value 3
	value = mockSB.GetValue(prefixA + baseValue3)
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("base-value3-data"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> base value 4
	value = mockSB.GetValue(prefixA + baseValue4)
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("base-value4-data"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))

	// check value states
	status = scheduler.GetValueStatus(prefixA + baseValue1)
	Expect(status).ToNot(BeNil())
	checkBaseValueStatus(status, &BaseValueStatus{
		Value: &ValueStatus{
			Key:           prefixA + baseValue1,
			State:         ValueState_CONFIGURED,
			LastOperation: TxnOperation_CREATE,
		},
	})
	status = scheduler.GetValueStatus(prefixA + baseValue2)
	Expect(status).ToNot(BeNil())
	checkBaseValueStatus(status, &BaseValueStatus{
		Value: &ValueStatus{
			Key:           prefixA + baseValue2,
			State:         ValueState_CONFIGURED,
			LastOperation: TxnOperation_CREATE,
		},
	})
	status = scheduler.GetValueStatus(prefixA + baseValue3)
	Expect(status).ToNot(BeNil())
	checkBaseValueStatus(status, &BaseValueStatus{
		Value: &ValueStatus{
			Key:           prefixA + baseValue3,
			State:         ValueState_CONFIGURED,
			LastOperation: TxnOperation_CREATE,
		},
	})
	status = scheduler.GetValueStatus(prefixA + baseValue4)
	Expect(status).ToNot(BeNil())
	checkBaseValueStatus(status, &BaseValueStatus{
		Value: &ValueStatus{
			Key:           prefixA + baseValue4,
			State:         ValueState_CONFIGURED,
			LastOperation: TxnOperation_CREATE,
		},
	})

	// check operations executed in SB
	opHistory = mockSB.PopHistoryOfOps()
	Expect(opHistory).To(HaveLen(4))
	operation := opHistory[0]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue4))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[1]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue3))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[2]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue2))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[3]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1))
	Expect(operation.Err).To(BeNil())

	// check transaction operations
	txnHistory = scheduler.GetTransactionHistory(time.Time{}, time.Now())
	Expect(txnHistory).To(HaveLen(2))
	txn = txnHistory[1]
	Expect(txn.PreRecord).To(BeFalse())
	Expect(txn.Start.After(startTime)).To(BeTrue())
	Expect(txn.Start.Before(txn.Stop)).To(BeTrue())
	Expect(txn.Stop.Before(stopTime)).To(BeTrue())
	Expect(txn.SeqNum).To(BeEquivalentTo(1))
	Expect(txn.TxnType).To(BeEquivalentTo(NBTransaction))
	Expect(txn.ResyncType).To(BeEquivalentTo(NotResync))
	Expect(txn.Description).To(BeEmpty())
	checkRecordedValues(txn.Values, []RecordedKVPair{
		{Key: prefixA + baseValue4, Value: utils.RecordProtoMessage(test.NewStringValue("base-value4-data")), Origin: FromNB},
	})

	txnOps = RecordedTxnOps{
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixA + baseValue4,
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("base-value4-data")),
			PrevState: ValueState_NONEXISTENT,
			NewState:  ValueState_CONFIGURED,
		},
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixA + baseValue3,
			PrevValue: utils.RecordProtoMessage(test.NewStringValue("base-value3-data")),
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("base-value3-data")),
			PrevState: ValueState_PENDING,
			NewState:  ValueState_CONFIGURED,
		},
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixA + baseValue2,
			PrevValue: utils.RecordProtoMessage(test.NewStringValue("base-value2-data")),
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("base-value2-data")),
			PrevState: ValueState_PENDING,
			NewState:  ValueState_CONFIGURED,
		},
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixA + baseValue1,
			PrevValue: utils.RecordProtoMessage(test.NewStringValue("base-value1-data")),
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("base-value1-data")),
			PrevState: ValueState_PENDING,
			NewState:  ValueState_CONFIGURED,
		},
	}
	checkTxnOperations(txn.Planned, txnOps)
	checkTxnOperations(txn.Executed, txnOps)

	// check flag stats
	graphR = scheduler.graph.Read()
	errorStats = graphR.GetFlagStats(ErrorFlagIndex, nil)
	Expect(errorStats.TotalCount).To(BeEquivalentTo(0))
	pendingStats = graphR.GetFlagStats(UnavailValueFlagIndex, nil)
	Expect(pendingStats.TotalCount).To(BeEquivalentTo(3))
	derivedStats = graphR.GetFlagStats(DerivedFlagIndex, nil)
	Expect(derivedStats.TotalCount).To(BeEquivalentTo(0))
	lastUpdateStats = graphR.GetFlagStats(LastUpdateFlagIndex, nil)
	Expect(lastUpdateStats.TotalCount).To(BeEquivalentTo(7))
	descriptorStats = graphR.GetFlagStats(DescriptorFlagIndex, nil)
	Expect(descriptorStats.TotalCount).To(BeEquivalentTo(7))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor1Name))
	Expect(descriptorStats.PerValueCount[descriptor1Name]).To(BeEquivalentTo(7))
	valueStateStats = graphR.GetFlagStats(ValueStateFlagIndex, nil)
	Expect(valueStateStats.TotalCount).To(BeEquivalentTo(7))
	Expect(valueStateStats.PerValueCount).To(HaveKey(ValueState_PENDING.String()))
	Expect(valueStateStats.PerValueCount[ValueState_PENDING.String()]).To(BeEquivalentTo(3))
	Expect(valueStateStats.PerValueCount).To(HaveKey(ValueState_CONFIGURED.String()))
	Expect(valueStateStats.PerValueCount[ValueState_CONFIGURED.String()]).To(BeEquivalentTo(4))
	graphR.Release()

	// plan error before 3rd txn
	mockSB.PlanError(prefixA+baseValue2, errors.New("failed to remove the value"), nil)

	// run third transaction that will break the cycle even though the delete operation will fail
	startTime = time.Now()
	schedulerTxn = scheduler.StartNBTransaction()
	schedulerTxn.SetValue(prefixA+baseValue2, nil)
	seqNum, err = schedulerTxn.Commit(testCtx)
	stopTime = time.Now()
	Expect(seqNum).To(BeEquivalentTo(2))
	Expect(err).ToNot(BeNil())
	txnErr := err.(*TransactionError)
	Expect(txnErr.GetTxnInitError()).ShouldNot(HaveOccurred())
	kvErrors := txnErr.GetKVErrors()
	Expect(kvErrors).To(HaveLen(1))
	Expect(kvErrors[0].Key).To(BeEquivalentTo(prefixA + baseValue2))
	Expect(kvErrors[0].TxnOperation).To(BeEquivalentTo(TxnOperation_DELETE))
	Expect(kvErrors[0].Error.Error()).To(BeEquivalentTo("failed to remove the value"))

	// check the state of SB
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	Expect(mockSB.GetValues(nil)).To(HaveLen(2))
	// -> base value 1 - pending
	value = mockSB.GetValue(prefixA + baseValue1)
	Expect(value).To(BeNil())
	// -> base value 2 - failed to remove
	value = mockSB.GetValue(prefixA + baseValue2)
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("base-value2-data"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> base value 3 - pending
	value = mockSB.GetValue(prefixA + baseValue3)
	Expect(value).To(BeNil())
	// -> base value 4
	value = mockSB.GetValue(prefixA + baseValue4)
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("base-value4-data"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))

	// check value states
	status = scheduler.GetValueStatus(prefixA + baseValue1)
	Expect(status).ToNot(BeNil())
	checkBaseValueStatus(status, &BaseValueStatus{
		Value: &ValueStatus{
			Key:           prefixA + baseValue1,
			State:         ValueState_PENDING,
			LastOperation: TxnOperation_DELETE,
			Details:       []string{prefixA + baseValue2},
		},
	})
	status = scheduler.GetValueStatus(prefixA + baseValue2)
	Expect(status).ToNot(BeNil())
	checkBaseValueStatus(status, &BaseValueStatus{
		Value: &ValueStatus{
			Key:           prefixA + baseValue2,
			State:         ValueState_FAILED,
			LastOperation: TxnOperation_DELETE,
			Error:         "failed to remove the value",
		},
	})
	status = scheduler.GetValueStatus(prefixA + baseValue3)
	Expect(status).ToNot(BeNil())
	checkBaseValueStatus(status, &BaseValueStatus{
		Value: &ValueStatus{
			Key:           prefixA + baseValue3,
			State:         ValueState_PENDING,
			LastOperation: TxnOperation_DELETE,
			Details:       []string{prefixA + baseValue1},
		},
	})
	status = scheduler.GetValueStatus(prefixA + baseValue4)
	Expect(status).ToNot(BeNil())
	checkBaseValueStatus(status, &BaseValueStatus{
		Value: &ValueStatus{
			Key:           prefixA + baseValue4,
			State:         ValueState_CONFIGURED,
			LastOperation: TxnOperation_CREATE,
		},
	})

	// check operations executed in SB
	opHistory = mockSB.PopHistoryOfOps()
	Expect(opHistory).To(HaveLen(3))
	operation = opHistory[0]
	Expect(operation.OpType).To(Equal(test.MockDelete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue3))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[1]
	Expect(operation.OpType).To(Equal(test.MockDelete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[2]
	Expect(operation.OpType).To(Equal(test.MockDelete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue2))
	Expect(operation.Err.Error()).To(BeEquivalentTo("failed to remove the value"))

	// check transaction operations
	txnHistory = scheduler.GetTransactionHistory(time.Time{}, time.Now())
	Expect(txnHistory).To(HaveLen(3))
	txn = txnHistory[2]
	Expect(txn.PreRecord).To(BeFalse())
	Expect(txn.Start.After(startTime)).To(BeTrue())
	Expect(txn.Start.Before(txn.Stop)).To(BeTrue())
	Expect(txn.Stop.Before(stopTime)).To(BeTrue())
	Expect(txn.SeqNum).To(BeEquivalentTo(2))
	Expect(txn.TxnType).To(BeEquivalentTo(NBTransaction))
	Expect(txn.ResyncType).To(BeEquivalentTo(NotResync))
	Expect(txn.Description).To(BeEmpty())
	checkRecordedValues(txn.Values, []RecordedKVPair{
		{Key: prefixA + baseValue2, Value: utils.RecordProtoMessage(nil), Origin: FromNB},
	})

	// -> planned
	txnOps = RecordedTxnOps{
		{
			Operation: TxnOperation_DELETE,
			Key:       prefixA + baseValue3,
			PrevValue: utils.RecordProtoMessage(test.NewStringValue("base-value3-data")),
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("base-value3-data")),
			PrevState: ValueState_CONFIGURED,
			NewState:  ValueState_PENDING,
		},
		{
			Operation: TxnOperation_DELETE,
			Key:       prefixA + baseValue1,
			PrevValue: utils.RecordProtoMessage(test.NewStringValue("base-value1-data")),
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("base-value1-data")),
			PrevState: ValueState_CONFIGURED,
			NewState:  ValueState_PENDING,
		},
		{
			Operation: TxnOperation_DELETE,
			Key:       prefixA + baseValue2,
			PrevValue: utils.RecordProtoMessage(test.NewStringValue("base-value2-data")),
			PrevState: ValueState_CONFIGURED,
			NewState:  ValueState_REMOVED,
		},
	}
	checkTxnOperations(txn.Planned, txnOps)

	// -> executed
	txnOps = RecordedTxnOps{
		{
			Operation: TxnOperation_DELETE,
			Key:       prefixA + baseValue3,
			PrevValue: utils.RecordProtoMessage(test.NewStringValue("base-value3-data")),
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("base-value3-data")),
			PrevState: ValueState_CONFIGURED,
			NewState:  ValueState_PENDING,
		},
		{
			Operation: TxnOperation_DELETE,
			Key:       prefixA + baseValue1,
			PrevValue: utils.RecordProtoMessage(test.NewStringValue("base-value1-data")),
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("base-value1-data")),
			PrevState: ValueState_CONFIGURED,
			NewState:  ValueState_PENDING,
		},
		{
			Operation: TxnOperation_DELETE,
			Key:       prefixA + baseValue2,
			PrevValue: utils.RecordProtoMessage(test.NewStringValue("base-value2-data")),
			PrevState: ValueState_CONFIGURED,
			NewState:  ValueState_FAILED,
			NewErr:    errors.New("failed to remove the value"),
		},
	}
	checkTxnOperations(txn.Executed, txnOps)

	// check flag stats
	graphR = scheduler.graph.Read()
	errorStats = graphR.GetFlagStats(ErrorFlagIndex, nil)
	Expect(errorStats.TotalCount).To(BeEquivalentTo(1))
	pendingStats = graphR.GetFlagStats(UnavailValueFlagIndex, nil)
	Expect(pendingStats.TotalCount).To(BeEquivalentTo(6))
	derivedStats = graphR.GetFlagStats(DerivedFlagIndex, nil)
	Expect(derivedStats.TotalCount).To(BeEquivalentTo(0))
	lastUpdateStats = graphR.GetFlagStats(LastUpdateFlagIndex, nil)
	Expect(lastUpdateStats.TotalCount).To(BeEquivalentTo(10))
	descriptorStats = graphR.GetFlagStats(DescriptorFlagIndex, nil)
	Expect(descriptorStats.TotalCount).To(BeEquivalentTo(10))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor1Name))
	Expect(descriptorStats.PerValueCount[descriptor1Name]).To(BeEquivalentTo(10))
	valueStateStats = graphR.GetFlagStats(ValueStateFlagIndex, nil)
	Expect(valueStateStats.TotalCount).To(BeEquivalentTo(10))
	Expect(valueStateStats.PerValueCount).To(HaveKey(ValueState_PENDING.String()))
	Expect(valueStateStats.PerValueCount[ValueState_PENDING.String()]).To(BeEquivalentTo(5))
	Expect(valueStateStats.PerValueCount).To(HaveKey(ValueState_CONFIGURED.String()))
	Expect(valueStateStats.PerValueCount[ValueState_CONFIGURED.String()]).To(BeEquivalentTo(4))
	Expect(valueStateStats.PerValueCount).To(HaveKey(ValueState_FAILED.String()))
	Expect(valueStateStats.PerValueCount[ValueState_FAILED.String()]).To(BeEquivalentTo(1))
	graphR.Release()

	// finally, run 4th txn to get back the removed value
	schedulerTxn = scheduler.StartNBTransaction()
	schedulerTxn.SetValue(prefixA+baseValue2, test.NewStringValue("base-value2-data-new"))
	seqNum, err = schedulerTxn.Commit(testCtx)
	Expect(seqNum).To(BeEquivalentTo(3))
	Expect(err).ShouldNot(HaveOccurred())

	// check the state of SB
	//Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty()) <- there is a validation error, but that's OK since descriptor does not define Retrieve
	Expect(mockSB.GetValues(nil)).To(HaveLen(4))
	// -> base value 1
	value = mockSB.GetValue(prefixA + baseValue1)
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("base-value1-data"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> base value 2
	value = mockSB.GetValue(prefixA + baseValue2)
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("base-value2-data-new"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> base value 3
	value = mockSB.GetValue(prefixA + baseValue3)
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("base-value3-data"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> base value 4
	value = mockSB.GetValue(prefixA + baseValue4)
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("base-value4-data"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))

	// check value states
	status = scheduler.GetValueStatus(prefixA + baseValue1)
	Expect(status).ToNot(BeNil())
	checkBaseValueStatus(status, &BaseValueStatus{
		Value: &ValueStatus{
			Key:           prefixA + baseValue1,
			State:         ValueState_CONFIGURED,
			LastOperation: TxnOperation_CREATE,
		},
	})
	status = scheduler.GetValueStatus(prefixA + baseValue2)
	Expect(status).ToNot(BeNil())
	checkBaseValueStatus(status, &BaseValueStatus{
		Value: &ValueStatus{
			Key:           prefixA + baseValue2,
			State:         ValueState_CONFIGURED,
			LastOperation: TxnOperation_CREATE,
		},
	})
	status = scheduler.GetValueStatus(prefixA + baseValue3)
	Expect(status).ToNot(BeNil())
	checkBaseValueStatus(status, &BaseValueStatus{
		Value: &ValueStatus{
			Key:           prefixA + baseValue3,
			State:         ValueState_CONFIGURED,
			LastOperation: TxnOperation_CREATE,
		},
	})
	status = scheduler.GetValueStatus(prefixA + baseValue4)
	Expect(status).ToNot(BeNil())
	checkBaseValueStatus(status, &BaseValueStatus{
		Value: &ValueStatus{
			Key:           prefixA + baseValue4,
			State:         ValueState_CONFIGURED,
			LastOperation: TxnOperation_CREATE,
		},
	})
}

func TestFailedDeleteOfDerivedValue(t *testing.T) {
	RegisterTestingT(t)

	// prepare KV Scheduler
	scheduler := NewPlugin(UseDeps(func(deps *Deps) {
		deps.HTTPHandlers = nil
	}))
	err := scheduler.Init()
	Expect(err).To(BeNil())

	// prepare mocks
	mockSB := test.NewMockSouthbound()
	// descriptor:
	descriptor := test.NewMockDescriptor(&KVDescriptor{
		Name:          descriptor1Name,
		NBKeyPrefix:   prefixA,
		KeySelector:   prefixSelector(prefixA),
		ValueTypeName: proto.MessageName(test.NewArrayValue()),
		DerivedValues: test.ArrayValueDerBuilder,
		WithMetadata:  true,
	}, mockSB, 0)
	scheduler.RegisterKVDescriptor(descriptor)

	// run non-resync transaction against empty SB
	schedulerTxn := scheduler.StartNBTransaction()
	schedulerTxn.SetValue(prefixA+baseValue1, test.NewArrayValue("item1"))
	seqNum, err := schedulerTxn.Commit(testCtx)
	Expect(seqNum).To(BeEquivalentTo(0))
	Expect(err).ShouldNot(HaveOccurred())

	// check the state of SB
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	// -> base value 1
	value := mockSB.GetValue(prefixA + baseValue1)
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewArrayValue("item1"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(0))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item1 derived from base value 1
	value = mockSB.GetValue(prefixA + baseValue1 + "/item1")
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("item1"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))

	// plan error before 2nd txn
	failedDeleteClb := func() {
		mockSB.SetValue(prefixA+baseValue1, test.NewArrayValue("item1"),
			&test.OnlyInteger{Integer: 0}, FromNB, false)
	}
	mockSB.PlanError(prefixA+baseValue1+"/item1", errors.New("failed to delete value"), failedDeleteClb)

	// run 2nd non-resync transaction that will have errors
	startTime := time.Now()
	schedulerTxn2 := scheduler.StartNBTransaction()
	schedulerTxn2.SetValue(prefixA+baseValue1, nil)
	seqNum, err = schedulerTxn2.Commit(testCtx)
	stopTime := time.Now()
	Expect(seqNum).To(BeEquivalentTo(1))
	Expect(err).ToNot(BeNil())
	txnErr := err.(*TransactionError)
	Expect(txnErr.GetTxnInitError()).ShouldNot(HaveOccurred())
	kvErrors := txnErr.GetKVErrors()
	Expect(kvErrors).To(HaveLen(1))
	Expect(kvErrors[0].Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item1"))
	Expect(kvErrors[0].TxnOperation).To(BeEquivalentTo(TxnOperation_DELETE))
	Expect(kvErrors[0].Error.Error()).To(BeEquivalentTo("failed to delete value"))

	// check transaction operations
	txnHistory := scheduler.GetTransactionHistory(time.Time{}, time.Now())
	Expect(txnHistory).To(HaveLen(2))
	txn := txnHistory[1]
	Expect(txn.PreRecord).To(BeFalse())
	Expect(txn.Start.After(startTime)).To(BeTrue())
	Expect(txn.Start.Before(txn.Stop)).To(BeTrue())
	Expect(txn.Stop.Before(stopTime)).To(BeTrue())
	Expect(txn.SeqNum).To(BeEquivalentTo(1))
	Expect(txn.TxnType).To(BeEquivalentTo(NBTransaction))
	Expect(txn.ResyncType).To(BeEquivalentTo(NotResync))
	Expect(txn.Description).To(BeEmpty())
	checkRecordedValues(txn.Values, []RecordedKVPair{
		{Key: prefixA + baseValue1, Value: nil, Origin: FromNB},
	})

	// -> planned
	txnOps := RecordedTxnOps{
		{
			Operation: TxnOperation_DELETE,
			Key:       prefixA + baseValue1 + "/item1",
			IsDerived: true,
			PrevValue: utils.RecordProtoMessage(test.NewStringValue("item1")),
			PrevState: ValueState_CONFIGURED,
			NewState:  ValueState_REMOVED,
		},
		{
			Operation: TxnOperation_DELETE,
			Key:       prefixA + baseValue1,
			PrevValue: utils.RecordProtoMessage(test.NewArrayValue("item1")),
			PrevState: ValueState_CONFIGURED,
			NewState:  ValueState_REMOVED,
		},
	}
	checkTxnOperations(txn.Planned, txnOps)

	// -> executed
	txnOps = RecordedTxnOps{
		{
			Operation: TxnOperation_DELETE,
			Key:       prefixA + baseValue1 + "/item1",
			IsDerived: true,
			PrevValue: utils.RecordProtoMessage(test.NewStringValue("item1")),
			PrevState: ValueState_CONFIGURED,
			NewState:  ValueState_FAILED,
			NewErr:    errors.New("failed to delete value"),
		},
	}
	checkTxnOperations(txn.Executed, txnOps)

	// check value status
	status := scheduler.GetValueStatus(prefixA + baseValue1)
	Expect(status).ToNot(BeNil())
	checkBaseValueStatus(status, &BaseValueStatus{
		Value: &ValueStatus{
			Key:           prefixA + baseValue1,
			State:         ValueState_CONFIGURED,
			LastOperation: TxnOperation_DELETE,
		},
		DerivedValues: []*ValueStatus{
			{
				Key:           prefixA + baseValue1 + "/item1",
				State:         ValueState_FAILED,
				LastOperation: TxnOperation_DELETE,
				Error:         "failed to delete value",
			},
		},
	})

	// close scheduler
	err = scheduler.Close()
	Expect(err).To(BeNil())
}

func TestFailedRecreateOfDerivedValue(t *testing.T) {
	RegisterTestingT(t)

	// prepare KV Scheduler
	scheduler := NewPlugin(UseDeps(func(deps *Deps) {
		deps.HTTPHandlers = nil
	}))
	err := scheduler.Init()
	Expect(err).To(BeNil())

	// prepare mocks
	mockSB := test.NewMockSouthbound()
	// descriptor:
	descriptor := test.NewMockDescriptor(&KVDescriptor{
		Name:          descriptor1Name,
		NBKeyPrefix:   prefixA,
		KeySelector:   prefixSelector(prefixA),
		ValueTypeName: proto.MessageName(test.NewArrayValue()),
		DerivedValues: test.ArrayValueDerBuilder,
		WithMetadata:  true,
		UpdateWithRecreate: func(key string, oldValue, newValue proto.Message, metadata Metadata) bool {
			return key == prefixA+baseValue1+"/item1"
		},
	}, mockSB, 0)
	scheduler.RegisterKVDescriptor(descriptor)

	// run non-resync transaction against empty SB
	arrayVal1 := test.NewArrayValueWithSuffix("-v1", "item1")
	schedulerTxn := scheduler.StartNBTransaction()
	schedulerTxn.SetValue(prefixA+baseValue1, arrayVal1)
	seqNum, err := schedulerTxn.Commit(testCtx)
	Expect(seqNum).To(BeEquivalentTo(0))
	Expect(err).ShouldNot(HaveOccurred())

	// check the state of SB
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	// -> base value 1
	value := mockSB.GetValue(prefixA + baseValue1)
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, arrayVal1)).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(0))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item1 derived from base value 1
	value = mockSB.GetValue(prefixA + baseValue1 + "/item1")
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("item1-v1"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))

	// plan error before 2nd txn
	failedCreateClb := func() {
		mockSB.SetValue(prefixA+baseValue1, test.NewArrayValue(),
			&test.OnlyInteger{Integer: 0}, FromNB, false)
	}
	mockSB.PlanError(prefixA+baseValue1+"/item1", nil, nil)                                              // Delete
	mockSB.PlanError(prefixA+baseValue1+"/item1", errors.New("failed to create value"), failedCreateClb) // (Re)Create

	// run 2nd non-resync transaction that will have errors
	startTime := time.Now()
	schedulerTxn2 := scheduler.StartNBTransaction()
	arrayVal2 := test.NewArrayValueWithSuffix("-v2", "item1")
	schedulerTxn2.SetValue(prefixA+baseValue1, arrayVal2)
	seqNum, err = schedulerTxn2.Commit(testCtx)
	stopTime := time.Now()
	Expect(seqNum).To(BeEquivalentTo(1))
	Expect(err).ToNot(BeNil())
	txnErr := err.(*TransactionError)
	Expect(txnErr.GetTxnInitError()).ShouldNot(HaveOccurred())
	kvErrors := txnErr.GetKVErrors()
	Expect(kvErrors).To(HaveLen(1))
	Expect(kvErrors[0].Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item1"))
	Expect(kvErrors[0].TxnOperation).To(BeEquivalentTo(TxnOperation_CREATE))
	Expect(kvErrors[0].Error.Error()).To(BeEquivalentTo("failed to create value"))

	// check transaction operations
	txnHistory := scheduler.GetTransactionHistory(time.Time{}, time.Now())
	Expect(txnHistory).To(HaveLen(2))
	txn := txnHistory[1]
	Expect(txn.PreRecord).To(BeFalse())
	Expect(txn.Start.After(startTime)).To(BeTrue())
	Expect(txn.Start.Before(txn.Stop)).To(BeTrue())
	Expect(txn.Stop.Before(stopTime)).To(BeTrue())
	Expect(txn.SeqNum).To(BeEquivalentTo(1))
	Expect(txn.TxnType).To(BeEquivalentTo(NBTransaction))
	Expect(txn.ResyncType).To(BeEquivalentTo(NotResync))
	Expect(txn.Description).To(BeEmpty())
	checkRecordedValues(txn.Values, []RecordedKVPair{
		{Key: prefixA + baseValue1, Value: utils.RecordProtoMessage(arrayVal2), Origin: FromNB},
	})

	// -> planned
	txnOps := RecordedTxnOps{
		{
			Operation: TxnOperation_UPDATE,
			Key:       prefixA + baseValue1,
			PrevValue: utils.RecordProtoMessage(arrayVal1),
			NewValue:  utils.RecordProtoMessage(arrayVal2),
			PrevState: ValueState_CONFIGURED,
			NewState:  ValueState_CONFIGURED,
		},
		{
			Operation:  TxnOperation_DELETE,
			Key:        prefixA + baseValue1 + "/item1",
			IsDerived:  true,
			PrevValue:  utils.RecordProtoMessage(test.NewStringValue("item1-v1")),
			PrevState:  ValueState_CONFIGURED,
			NewState:   ValueState_REMOVED,
			IsRecreate: true,
		},
		{
			Operation:  TxnOperation_CREATE,
			Key:        prefixA + baseValue1 + "/item1",
			IsDerived:  true,
			NewValue:   utils.RecordProtoMessage(test.NewStringValue("item1-v2")),
			PrevState:  ValueState_REMOVED,
			NewState:   ValueState_CONFIGURED,
			IsRecreate: true,
		},
	}
	checkTxnOperations(txn.Planned, txnOps)

	// -> executed
	txnOps = RecordedTxnOps{
		{
			Operation: TxnOperation_UPDATE,
			Key:       prefixA + baseValue1,
			PrevValue: utils.RecordProtoMessage(arrayVal1),
			NewValue:  utils.RecordProtoMessage(arrayVal2),
			PrevState: ValueState_CONFIGURED,
			NewState:  ValueState_CONFIGURED,
		},
		{
			Operation:  TxnOperation_DELETE,
			Key:        prefixA + baseValue1 + "/item1",
			IsDerived:  true,
			PrevValue:  utils.RecordProtoMessage(test.NewStringValue("item1-v1")),
			PrevState:  ValueState_CONFIGURED,
			NewState:   ValueState_REMOVED,
			IsRecreate: true,
		},
		{
			Operation:  TxnOperation_CREATE,
			Key:        prefixA + baseValue1 + "/item1",
			IsDerived:  true,
			NewValue:   utils.RecordProtoMessage(test.NewStringValue("item1-v2")),
			PrevState:  ValueState_REMOVED,
			NewState:   ValueState_FAILED,
			NewErr:     errors.New("failed to create value"),
			IsRecreate: true,
		},
	}
	checkTxnOperations(txn.Executed, txnOps)

	// close scheduler
	err = scheduler.Close()
	Expect(err).To(BeNil())
}
