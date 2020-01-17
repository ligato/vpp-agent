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
	"errors"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"github.com/golang/protobuf/proto"
	. "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler/internal/test"
	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler/internal/utils"
	. "go.ligato.io/vpp-agent/v3/proto/ligato/kvscheduler"
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
	descriptor1 := test.NewMockDescriptor(&KVDescriptor{
		Name:         descriptor1Name,
		NBKeyPrefix:  prefixA,
		KeySelector:  prefixSelector(prefixA),
		WithMetadata: true,
	}, mockSB, 0)

	// register descriptor with the scheduler
	scheduler.RegisterKVDescriptor(descriptor1)
	nbPrefixes := scheduler.GetRegisteredNBKeyPrefixes()
	Expect(nbPrefixes).To(HaveLen(1))
	Expect(nbPrefixes).To(ContainElement(prefixA))

	// get metadata map created for the descriptor
	metadataMap := scheduler.GetMetadataMap(descriptor1.Name)
	_, withMetadataMap := metadataMap.(test.NameToInteger)
	Expect(withMetadataMap).To(BeTrue())

	// transaction history should be initially empty
	Expect(scheduler.GetTransactionHistory(time.Time{}, time.Time{})).To(BeEmpty())

	// run transaction with empty resync
	startTime := time.Now()
	ctx := WithResync(testCtx, FullResync, true)
	description := "testing empty resync"
	ctx = WithDescription(ctx, description)
	seqNum, err := scheduler.StartNBTransaction().Commit(ctx)
	stopTime := time.Now()
	Expect(seqNum).To(BeEquivalentTo(0))
	Expect(err).ShouldNot(HaveOccurred())

	// check the state of SB
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	Expect(mockSB.GetValues(nil)).To(BeEmpty())

	// check metadata
	Expect(metadataMap.ListAllNames()).To(BeEmpty())

	// check executed operations
	opHistory := mockSB.PopHistoryOfOps()
	Expect(opHistory).To(HaveLen(1))
	Expect(opHistory[0].OpType).To(Equal(test.MockRetrieve))
	Expect(opHistory[0].CorrelateRetrieve).To(BeEmpty())
	Expect(opHistory[0].Descriptor).To(BeEquivalentTo(descriptor1Name))

	// single transaction consisted of zero operations
	txnHistory := scheduler.GetTransactionHistory(time.Time{}, time.Time{})
	Expect(txnHistory).To(HaveLen(1))
	txn := txnHistory[0]
	Expect(txn.PreRecord).To(BeFalse())
	Expect(txn.Start.After(startTime)).To(BeTrue())
	Expect(txn.Start.Before(txn.Stop)).To(BeTrue())
	Expect(txn.Stop.Before(stopTime)).To(BeTrue())
	Expect(txn.SeqNum).To(BeEquivalentTo(0))
	Expect(txn.TxnType).To(BeEquivalentTo(NBTransaction))
	Expect(txn.RetryAttempt).To(BeEquivalentTo(0))
	Expect(txn.RetryForTxn).To(BeEquivalentTo(0))
	Expect(txn.ResyncType).To(BeEquivalentTo(FullResync))
	Expect(txn.Description).To(Equal(description))
	Expect(txn.Values).To(BeEmpty())
	Expect(txn.Planned).To(BeEmpty())
	Expect(txn.Executed).To(BeEmpty())

	// check flag stats
	graphR := scheduler.graph.Read()
	errorStats := graphR.GetFlagStats(ErrorFlagIndex, nil)
	Expect(errorStats.TotalCount).To(BeEquivalentTo(0))
	pendingStats := graphR.GetFlagStats(UnavailValueFlagIndex, nil)
	Expect(pendingStats.TotalCount).To(BeEquivalentTo(0))
	derivedStats := graphR.GetFlagStats(DerivedFlagIndex, nil)
	Expect(derivedStats.TotalCount).To(BeEquivalentTo(0))
	lastUpdateStats := graphR.GetFlagStats(LastUpdateFlagIndex, nil)
	Expect(lastUpdateStats.TotalCount).To(BeEquivalentTo(0))
	descriptorStats := graphR.GetFlagStats(DescriptorFlagIndex, nil)
	Expect(descriptorStats.TotalCount).To(BeEquivalentTo(0))
	valueStateStats := graphR.GetFlagStats(ValueStateFlagIndex, nil)
	Expect(valueStateStats.TotalCount).To(BeEquivalentTo(0))

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
	descriptor1 := test.NewMockDescriptor(&KVDescriptor{
		Name:          descriptor1Name,
		NBKeyPrefix:   prefixA,
		KeySelector:   prefixSelector(prefixA),
		ValueTypeName: proto.MessageName(test.NewArrayValue()),
		DerivedValues: test.ArrayValueDerBuilder,
		Dependencies: func(key string, value proto.Message) []Dependency {
			if key == prefixA+baseValue2 {
				depKey := prefixA + baseValue1 + "/item1" // base value depends on a derived value
				return []Dependency{
					{Label: depKey, Key: depKey},
				}
			}
			if key == prefixA+baseValue1+"/item2" {
				depKey := prefixA + baseValue2 + "/item1" // derived value depends on another derived value
				return []Dependency{
					{Label: depKey, Key: depKey},
				}
			}
			return nil
		},
		WithMetadata: true,
	}, mockSB, 0)

	// register descriptor with the scheduler
	scheduler.RegisterKVDescriptor(descriptor1)

	// get metadata map created for the descriptor
	metadataMap := scheduler.GetMetadataMap(descriptor1.Name)
	nameToInteger, withMetadataMap := metadataMap.(test.NameToInteger)
	Expect(withMetadataMap).To(BeTrue())

	// run resync transaction with empty SB
	startTime := time.Now()
	schedulerTxn := scheduler.StartNBTransaction()
	schedulerTxn.SetValue(prefixA+baseValue2, test.NewArrayValue("item1"))
	schedulerTxn.SetValue(prefixA+baseValue1, test.NewArrayValue("item1", "item2"))
	ctx := WithResync(testCtx, FullResync, true)
	description := "testing resync against empty SB"
	ctx = WithDescription(ctx, description)
	seqNum, err := schedulerTxn.Commit(ctx)
	stopTime := time.Now()
	Expect(seqNum).To(BeEquivalentTo(0))
	Expect(err).ShouldNot(HaveOccurred())

	// check the state of SB
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	// -> base value 1
	value := mockSB.GetValue(prefixA + baseValue1)
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewArrayValue("item1", "item2"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(0))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item1 derived from base value 1
	value = mockSB.GetValue(prefixA + baseValue1 + "/item1")
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("item1"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item2 derived from base value 1
	value = mockSB.GetValue(prefixA + baseValue1 + "/item2")
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("item2"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> base value 2
	value = mockSB.GetValue(prefixA + baseValue2)
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewArrayValue("item1"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(1))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item1 derived from base value 2
	value = mockSB.GetValue(prefixA + baseValue2 + "/item1")
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("item1"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	Expect(mockSB.GetValues(nil)).To(HaveLen(5))

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
	Expect(operation.OpType).To(Equal(test.MockRetrieve))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	checkValues(operation.CorrelateRetrieve, []KVWithMetadata{
		{
			Key:      prefixA + baseValue1,
			Value:    test.NewArrayValue("item1", "item2"),
			Metadata: nil,
			Origin:   FromNB,
		},
		{
			Key:      prefixA + baseValue2,
			Value:    test.NewArrayValue("item1"),
			Metadata: nil,
			Origin:   FromNB,
		},
	})
	operation = opHistory[1]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[2]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[3]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue2))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[4]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue2 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[5]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item2"))
	Expect(operation.Err).To(BeNil())

	// check value dumps
	expValues := []KVWithMetadata{
		{Key: prefixA + baseValue1, Value: test.NewArrayValue("item1", "item2"), Origin: FromNB, Metadata: &test.OnlyInteger{Integer: 0}},
		{Key: prefixA + baseValue2, Value: test.NewArrayValue("item1"), Origin: FromNB, Metadata: &test.OnlyInteger{Integer: 1}},
	}
	views := []View{NBView, SBView, CachedView}
	for _, view := range views {
		dumpedValues, err := scheduler.DumpValuesByKeyPrefix(prefixA, view)
		Expect(err).To(BeNil())
		checkValues(dumpedValues, expValues)
	}
	for _, view := range views {
		dumpedValues, err := scheduler.DumpValuesByDescriptor(descriptor1Name, view)
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
				Key:           prefixA + baseValue1 + "/item1",
				State:         ValueState_CONFIGURED,
				LastOperation: TxnOperation_CREATE,
			},
			{
				Key:           prefixA + baseValue1 + "/item2",
				State:         ValueState_CONFIGURED,
				LastOperation: TxnOperation_CREATE,
			},
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
		DerivedValues: []*ValueStatus{
			{
				Key:           prefixA + baseValue2 + "/item1",
				State:         ValueState_CONFIGURED,
				LastOperation: TxnOperation_CREATE,
			},
		},
	})

	// single transaction consisted of 6 operations
	txnHistory := scheduler.GetTransactionHistory(time.Time{}, time.Now())
	Expect(txnHistory).To(HaveLen(1))
	txn := txnHistory[0]
	Expect(txn.PreRecord).To(BeFalse())
	Expect(txn.Start.After(startTime)).To(BeTrue())
	Expect(txn.Start.Before(txn.Stop)).To(BeTrue())
	Expect(txn.Stop.Before(stopTime)).To(BeTrue())
	Expect(txn.SeqNum).To(BeEquivalentTo(0))
	Expect(txn.TxnType).To(BeEquivalentTo(NBTransaction))
	Expect(txn.ResyncType).To(BeEquivalentTo(FullResync))
	Expect(txn.Description).To(Equal(description))
	checkRecordedValues(txn.Values, []RecordedKVPair{
		{Key: prefixA + baseValue1, Value: utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")), Origin: FromNB},
		{Key: prefixA + baseValue2, Value: utils.RecordProtoMessage(test.NewArrayValue("item1")), Origin: FromNB},
	})

	txnOps := RecordedTxnOps{
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixA + baseValue1,
			NewValue:  utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")),
			PrevState: ValueState_NONEXISTENT,
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
			Key:       prefixA + baseValue2,
			NewValue:  utils.RecordProtoMessage(test.NewArrayValue("item1")),
			PrevState: ValueState_NONEXISTENT,
			NewState:  ValueState_CONFIGURED,
		},
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixA + baseValue2 + "/item1",
			IsDerived: true,
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("item1")),
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
	}
	checkTxnOperations(txn.Planned, txnOps)
	checkTxnOperations(txn.Executed, txnOps)

	// now remove everything using resync with empty data
	startTime = time.Now()
	seqNum, err = scheduler.StartNBTransaction().Commit(WithResync(testCtx, FullResync, true))
	stopTime = time.Now()
	Expect(seqNum).To(BeEquivalentTo(1))
	Expect(err).ShouldNot(HaveOccurred())

	// check the state of SB
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	Expect(mockSB.GetValues(nil)).To(BeEmpty())

	// check metadata
	Expect(metadataMap.ListAllNames()).To(BeEmpty())

	// check executed operations
	opHistory = mockSB.PopHistoryOfOps()
	Expect(opHistory).To(HaveLen(6))
	operation = opHistory[0]
	Expect(operation.OpType).To(Equal(test.MockRetrieve))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	checkValues(operation.CorrelateRetrieve, []KVWithMetadata{
		{
			Key:      prefixA + baseValue1,
			Value:    test.NewArrayValue("item1", "item2"),
			Metadata: &test.OnlyInteger{Integer: 0},
			Origin:   FromNB,
		},
		{
			Key:      prefixA + baseValue2,
			Value:    test.NewArrayValue("item1"),
			Metadata: &test.OnlyInteger{Integer: 1},
			Origin:   FromNB,
		},
	})
	operation = opHistory[1]
	Expect(operation.OpType).To(Equal(test.MockDelete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item2"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[2]
	Expect(operation.OpType).To(Equal(test.MockDelete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue2 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[3]
	Expect(operation.OpType).To(Equal(test.MockDelete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue2))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[4]
	Expect(operation.OpType).To(Equal(test.MockDelete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[5]
	Expect(operation.OpType).To(Equal(test.MockDelete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1))
	Expect(operation.Err).To(BeNil())

	// this second transaction consisted of 6 operations
	txnHistory = scheduler.GetTransactionHistory(time.Time{}, time.Now())
	Expect(txnHistory).To(HaveLen(2))
	txn = txnHistory[1]
	Expect(txn.PreRecord).To(BeFalse())
	Expect(txn.Start.After(startTime)).To(BeTrue())
	Expect(txn.Start.Before(txn.Stop)).To(BeTrue())
	Expect(txn.Stop.Before(stopTime)).To(BeTrue())
	Expect(txn.SeqNum).To(BeEquivalentTo(1))
	Expect(txn.TxnType).To(BeEquivalentTo(NBTransaction))
	Expect(txn.ResyncType).To(BeEquivalentTo(FullResync))
	Expect(txn.Description).To(BeEmpty())
	checkRecordedValues(txn.Values, []RecordedKVPair{
		{Key: prefixA + baseValue1, Value: utils.RecordProtoMessage(nil), Origin: FromNB},
		{Key: prefixA + baseValue2, Value: utils.RecordProtoMessage(nil), Origin: FromNB},
	})

	txnOps = RecordedTxnOps{
		{
			Operation: TxnOperation_DELETE,
			Key:       prefixA + baseValue1 + "/item2",
			IsDerived: true,
			PrevValue: utils.RecordProtoMessage(test.NewStringValue("item2")),
			PrevState: ValueState_CONFIGURED,
			NewState:  ValueState_REMOVED,
		},
		{
			Operation: TxnOperation_DELETE,
			Key:       prefixA + baseValue2 + "/item1",
			IsDerived: true,
			PrevValue: utils.RecordProtoMessage(test.NewStringValue("item1")),
			PrevState: ValueState_CONFIGURED,
			NewState:  ValueState_REMOVED,
		},
		{
			Operation: TxnOperation_DELETE,
			Key:       prefixA + baseValue2,
			PrevValue: utils.RecordProtoMessage(test.NewArrayValue("item1")),
			PrevState: ValueState_CONFIGURED,
			NewState:  ValueState_REMOVED,
		},
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
			PrevValue: utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")),
			PrevState: ValueState_CONFIGURED,
			NewState:  ValueState_REMOVED,
		},
	}
	checkTxnOperations(txn.Planned, txnOps)
	checkTxnOperations(txn.Executed, txnOps)

	// check flag stats
	// Note: removed derived values are not kept in the graph
	graphR := scheduler.graph.Read()
	errorStats := graphR.GetFlagStats(ErrorFlagIndex, nil)
	Expect(errorStats.TotalCount).To(BeEquivalentTo(0))
	pendingStats := graphR.GetFlagStats(UnavailValueFlagIndex, nil)
	Expect(pendingStats.TotalCount).To(BeEquivalentTo(2))
	derivedStats := graphR.GetFlagStats(DerivedFlagIndex, nil)
	Expect(derivedStats.TotalCount).To(BeEquivalentTo(3))
	lastUpdateStats := graphR.GetFlagStats(LastUpdateFlagIndex, nil)
	Expect(lastUpdateStats.TotalCount).To(BeEquivalentTo(7))
	descriptorStats := graphR.GetFlagStats(DescriptorFlagIndex, nil)
	Expect(descriptorStats.TotalCount).To(BeEquivalentTo(7))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor1Name))
	Expect(descriptorStats.PerValueCount[descriptor1Name]).To(BeEquivalentTo(7))
	valueStateStats := graphR.GetFlagStats(ValueStateFlagIndex, nil)
	Expect(valueStateStats.TotalCount).To(BeEquivalentTo(7))
	Expect(valueStateStats.PerValueCount).To(HaveKey(ValueState_CONFIGURED.String()))
	Expect(valueStateStats.PerValueCount[ValueState_CONFIGURED.String()]).To(BeEquivalentTo(5))
	Expect(valueStateStats.PerValueCount).To(HaveKey(ValueState_REMOVED.String()))
	Expect(valueStateStats.PerValueCount[ValueState_REMOVED.String()]).To(BeEquivalentTo(2))
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
	mockSB.SetValue(prefixA+baseValue1, test.NewArrayValue("item1"),
		&test.OnlyInteger{Integer: 0}, FromNB, false)
	mockSB.SetValue(prefixA+baseValue1+"/item1", test.NewStringValue("item1"),
		nil, FromNB, true)
	mockSB.SetValue(prefixA+baseValue2, test.NewArrayValue("item1"),
		&test.OnlyInteger{Integer: 1}, FromNB, false)
	mockSB.SetValue(prefixA+baseValue2+"/item1", test.NewStringValue("item1"),
		nil, FromNB, true)
	mockSB.SetValue(prefixA+baseValue3, test.NewArrayValue("item1"),
		&test.OnlyInteger{Integer: 2}, FromNB, false)
	mockSB.SetValue(prefixA+baseValue3+"/item1", test.NewStringValue("item1"),
		nil, FromNB, true)
	// -> descriptor1:
	descriptor1 := test.NewMockDescriptor(&KVDescriptor{
		Name:          descriptor1Name,
		NBKeyPrefix:   prefixA,
		KeySelector:   prefixSelector(prefixA),
		ValueTypeName: proto.MessageName(test.NewArrayValue()),
		DerivedValues: test.ArrayValueDerBuilder,
		Dependencies: func(key string, value proto.Message) []Dependency {
			if key == prefixA+baseValue2+"/item1" {
				depKey := prefixA + baseValue1
				return []Dependency{
					{Label: depKey, Key: depKey},
				}
			}
			if key == prefixA+baseValue2+"/item2" {
				depKey := prefixA + baseValue1 + "/item1"
				return []Dependency{
					{Label: depKey, Key: depKey},
				}
			}
			return nil
		},
		UpdateWithRecreate: func(key string, oldValue, newValue proto.Message, metadata Metadata) bool {
			return key == prefixA+baseValue3
		},
		WithMetadata: true,
	}, mockSB, 3)

	// register descriptor with the scheduler
	scheduler.RegisterKVDescriptor(descriptor1)

	// get metadata map created for the descriptor
	metadataMap := scheduler.GetMetadataMap(descriptor1.Name)
	nameToInteger, withMetadataMap := metadataMap.(test.NameToInteger)
	Expect(withMetadataMap).To(BeTrue())

	// run resync transaction with SB that already has some values added
	startTime := time.Now()
	schedulerTxn := scheduler.StartNBTransaction()
	schedulerTxn.SetValue(prefixA+baseValue1, test.NewArrayValue("item2"))
	schedulerTxn.SetValue(prefixA+baseValue2, test.NewArrayValue("item1", "item2"))
	schedulerTxn.SetValue(prefixA+baseValue3, test.NewArrayValue("item1", "item2"))
	seqNum, err := schedulerTxn.Commit(WithResync(testCtx, FullResync, true))
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
	// -> item1 derived from base value 1 was removed
	value = mockSB.GetValue(prefixA + baseValue1 + "/item1")
	Expect(value).To(BeNil())
	// -> item2 derived from base value 1
	value = mockSB.GetValue(prefixA + baseValue1 + "/item2")
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("item2"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> base value 2
	value = mockSB.GetValue(prefixA + baseValue2)
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewArrayValue("item1", "item2"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(1))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item1 derived from base value 2
	value = mockSB.GetValue(prefixA + baseValue2 + "/item1")
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("item1"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item2 derived from base value 2 is pending
	value = mockSB.GetValue(prefixA + baseValue2 + "/item2")
	Expect(value).To(BeNil())
	// -> base value 3
	value = mockSB.GetValue(prefixA + baseValue3)
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewArrayValue("item1", "item2"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(3))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item1 derived from base value 3
	value = mockSB.GetValue(prefixA + baseValue3 + "/item1")
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("item1"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item2 derived from base value 3
	value = mockSB.GetValue(prefixA + baseValue3 + "/item2")
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("item2"))).To(BeTrue())
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
	Expect(opHistory).To(HaveLen(10))
	operation := opHistory[0]
	Expect(operation.OpType).To(Equal(test.MockRetrieve))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	checkValues(operation.CorrelateRetrieve, []KVWithMetadata{
		{
			Key:      prefixA + baseValue1,
			Value:    test.NewArrayValue("item2"),
			Metadata: nil,
			Origin:   FromNB,
		},
		{
			Key:      prefixA + baseValue2,
			Value:    test.NewArrayValue("item1", "item2"),
			Metadata: nil,
			Origin:   FromNB,
		},
		{
			Key:      prefixA + baseValue3,
			Value:    test.NewArrayValue("item1", "item2"),
			Metadata: nil,
			Origin:   FromNB,
		},
	})
	operation = opHistory[1]
	Expect(operation.OpType).To(Equal(test.MockDelete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue3 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[2]
	Expect(operation.OpType).To(Equal(test.MockDelete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue3))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[3]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue3))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[4]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue3 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[5]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue3 + "/item2"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[6]
	Expect(operation.OpType).To(Equal(test.MockDelete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[7]
	Expect(operation.OpType).To(Equal(test.MockUpdate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[8]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item2"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[9]
	Expect(operation.OpType).To(Equal(test.MockUpdate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue2))
	Expect(operation.Err).To(BeNil())

	// check value dumps
	// TODO: actual configuration (when kept in-graph separately) will have to not include
	//       baseValue2/item2, but that means we will have to refresh the graph
	//       always after something (base or derived value) is found to be pending
	expValues := []KVWithMetadata{
		{Key: prefixA + baseValue1, Value: test.NewArrayValue("item2"), Origin: FromNB, Metadata: &test.OnlyInteger{Integer: 0}},
		{Key: prefixA + baseValue2, Value: test.NewArrayValue("item1", "item2"), Origin: FromNB, Metadata: &test.OnlyInteger{Integer: 1}},
		{Key: prefixA + baseValue3, Value: test.NewArrayValue("item1", "item2"), Origin: FromNB, Metadata: &test.OnlyInteger{Integer: 3}},
	}
	views := []View{NBView, SBView, CachedView}
	for _, view := range views {
		dumpedValues, err := scheduler.DumpValuesByKeyPrefix(prefixA, view)
		Expect(err).To(BeNil())
		checkValues(dumpedValues, expValues)
		dumpedValues, err = scheduler.DumpValuesByDescriptor(descriptor1Name, view)
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
	status = scheduler.GetValueStatus(prefixA + baseValue2)
	Expect(status).ToNot(BeNil())
	checkBaseValueStatus(status, &BaseValueStatus{
		Value: &ValueStatus{
			Key:           prefixA + baseValue2,
			State:         ValueState_CONFIGURED,
			LastOperation: TxnOperation_UPDATE,
		},
		DerivedValues: []*ValueStatus{
			{
				Key:           prefixA + baseValue2 + "/item1",
				State:         ValueState_CONFIGURED,
				LastOperation: TxnOperation_UPDATE,
			},
			{
				Key:           prefixA + baseValue2 + "/item2",
				State:         ValueState_PENDING,
				LastOperation: TxnOperation_CREATE,
				Details:       []string{prefixA + baseValue1 + "/item1"},
			},
		},
	})
	status = scheduler.GetValueStatus(prefixA + baseValue3)
	Expect(status).ToNot(BeNil())
	checkBaseValueStatus(status, &BaseValueStatus{
		Value: &ValueStatus{
			Key:           prefixA + baseValue3,
			State:         ValueState_CONFIGURED,
			LastOperation: TxnOperation_UPDATE,
		},
		DerivedValues: []*ValueStatus{
			{
				Key:           prefixA + baseValue3 + "/item1",
				State:         ValueState_CONFIGURED,
				LastOperation: TxnOperation_CREATE,
			},
			{
				Key:           prefixA + baseValue3 + "/item2",
				State:         ValueState_CONFIGURED,
				LastOperation: TxnOperation_CREATE,
			},
		},
	})

	// check transaction operations
	txnHistory := scheduler.GetTransactionHistory(time.Time{}, time.Time{})
	Expect(txnHistory).To(HaveLen(1))
	txn := txnHistory[0]
	Expect(txn.PreRecord).To(BeFalse())
	Expect(txn.Start.After(startTime)).To(BeTrue())
	Expect(txn.Start.Before(txn.Stop)).To(BeTrue())
	Expect(txn.Stop.Before(stopTime)).To(BeTrue())
	Expect(txn.SeqNum).To(BeEquivalentTo(0))
	Expect(txn.TxnType).To(BeEquivalentTo(NBTransaction))
	Expect(txn.ResyncType).To(BeEquivalentTo(FullResync))
	Expect(txn.Description).To(BeEmpty())
	checkRecordedValues(txn.Values, []RecordedKVPair{
		{Key: prefixA + baseValue1, Value: utils.RecordProtoMessage(test.NewArrayValue("item2")), Origin: FromNB},
		{Key: prefixA + baseValue2, Value: utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")), Origin: FromNB},
		{Key: prefixA + baseValue3, Value: utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")), Origin: FromNB},
	})

	txnOps := RecordedTxnOps{
		{
			Operation:  TxnOperation_DELETE,
			Key:        prefixA + baseValue3 + "/item1",
			IsDerived:  true,
			PrevValue:  utils.RecordProtoMessage(test.NewStringValue("item1")),
			PrevState:  ValueState_DISCOVERED,
			NewState:   ValueState_REMOVED,
			IsRecreate: true,
		},
		{
			Operation:  TxnOperation_DELETE,
			Key:        prefixA + baseValue3,
			PrevValue:  utils.RecordProtoMessage(test.NewArrayValue("item1")),
			PrevState:  ValueState_DISCOVERED,
			NewState:   ValueState_REMOVED,
			IsRecreate: true,
		},
		{
			Operation:  TxnOperation_CREATE,
			Key:        prefixA + baseValue3,
			NewValue:   utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")),
			PrevState:  ValueState_REMOVED,
			NewState:   ValueState_CONFIGURED,
			IsRecreate: true,
		},
		{
			Operation:  TxnOperation_CREATE,
			Key:        prefixA + baseValue3 + "/item1",
			IsDerived:  true,
			NewValue:   utils.RecordProtoMessage(test.NewStringValue("item1")),
			PrevState:  ValueState_NONEXISTENT, // TODO: derived value removed from the graph, ok?
			NewState:   ValueState_CONFIGURED,
			IsRecreate: true,
		},
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixA + baseValue3 + "/item2",
			IsDerived: true,
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("item2")),
			PrevState: ValueState_NONEXISTENT,
			NewState:  ValueState_CONFIGURED,
		},
		{
			Operation: TxnOperation_DELETE,
			Key:       prefixA + baseValue1 + "/item1",
			IsDerived: true,
			PrevValue: utils.RecordProtoMessage(test.NewStringValue("item1")),
			PrevState: ValueState_DISCOVERED,
			NewState:  ValueState_REMOVED,
		},
		{
			Operation: TxnOperation_UPDATE,
			Key:       prefixA + baseValue1,
			PrevValue: utils.RecordProtoMessage(test.NewArrayValue("item1")),
			NewValue:  utils.RecordProtoMessage(test.NewArrayValue("item2")),
			PrevState: ValueState_DISCOVERED,
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
			Operation: TxnOperation_UPDATE,
			Key:       prefixA + baseValue2,
			PrevValue: utils.RecordProtoMessage(test.NewArrayValue("item1")),
			NewValue:  utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")),
			PrevState: ValueState_DISCOVERED,
			NewState:  ValueState_CONFIGURED,
		},
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixA + baseValue2 + "/item2",
			IsDerived: true,
			NOOP:      true,
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("item2")),
			PrevState: ValueState_NONEXISTENT,
			NewState:  ValueState_PENDING,
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
	Expect(descriptorStats.PerValueCount[descriptor1Name]).To(BeEquivalentTo(8))
	valueStateStats := graphR.GetFlagStats(ValueStateFlagIndex, nil)
	Expect(valueStateStats.TotalCount).To(BeEquivalentTo(8))
	Expect(valueStateStats.PerValueCount).To(HaveKey(ValueState_CONFIGURED.String()))
	Expect(valueStateStats.PerValueCount[ValueState_CONFIGURED.String()]).To(BeEquivalentTo(7))
	Expect(valueStateStats.PerValueCount).To(HaveKey(ValueState_PENDING.String()))
	Expect(valueStateStats.PerValueCount[ValueState_PENDING.String()]).To(BeEquivalentTo(1))
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
	mockSB.SetValue(prefixA+baseValue1, test.NewStringValue(baseValue1),
		nil, FromSB, false)
	// -> descriptor1:
	descriptor1 := test.NewMockDescriptor(&KVDescriptor{
		Name:          descriptor1Name,
		KeySelector:   prefixSelector(prefixA),
		NBKeyPrefix:   prefixA,
		ValueTypeName: proto.MessageName(test.NewArrayValue()),
		DerivedValues: test.ArrayValueDerBuilder,
		Dependencies: func(key string, value proto.Message) []Dependency {
			if key == prefixA+baseValue2 {
				depKey := prefixA + baseValue1
				return []Dependency{
					{Label: depKey, Key: depKey},
				}
			}
			return nil
		},
		WithMetadata: true,
	}, mockSB, 0)

	// register descriptor with the scheduler
	scheduler.RegisterKVDescriptor(descriptor1)

	// get metadata map created for the descriptor
	metadataMap := scheduler.GetMetadataMap(descriptor1.Name)
	nameToInteger, withMetadataMap := metadataMap.(test.NameToInteger)
	Expect(withMetadataMap).To(BeTrue())

	// run resync transaction that should keep values not managed by NB untouched
	startTime := time.Now()
	schedulerTxn := scheduler.StartNBTransaction()
	schedulerTxn.SetValue(prefixA+baseValue2, test.NewArrayValue("item1"))
	seqNum, err := schedulerTxn.Commit(WithResync(testCtx, FullResync, true))
	stopTime := time.Now()
	Expect(seqNum).To(BeEquivalentTo(0))
	Expect(err).ShouldNot(HaveOccurred())

	// check the state of SB
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	// -> base value 1
	value := mockSB.GetValue(prefixA + baseValue1)
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue(baseValue1))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromSB))
	// -> base value 2
	value = mockSB.GetValue(prefixA + baseValue2)
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewArrayValue("item1"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(0))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item1 derived from base value 2
	value = mockSB.GetValue(prefixA + baseValue2 + "/item1")
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("item1"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	Expect(mockSB.GetValues(nil)).To(HaveLen(3))

	// check metadata
	metadata, exists := nameToInteger.LookupByName(baseValue1)
	Expect(exists).To(BeFalse())
	Expect(metadata).To(BeNil())
	metadata, exists = nameToInteger.LookupByName(baseValue2)
	Expect(exists).To(BeTrue())
	Expect(metadata.GetInteger()).To(BeEquivalentTo(0))

	// check operations executed in SB
	opHistory := mockSB.PopHistoryOfOps()
	Expect(opHistory).To(HaveLen(3))
	operation := opHistory[0]
	Expect(operation.OpType).To(Equal(test.MockRetrieve))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	checkValues(operation.CorrelateRetrieve, []KVWithMetadata{
		{
			Key:      prefixA + baseValue2,
			Value:    test.NewArrayValue("item1"),
			Metadata: nil,
			Origin:   FromNB,
		},
	})
	operation = opHistory[1]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue2))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[2]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue2 + "/item1"))
	Expect(operation.Err).To(BeNil())

	// check value dumps
	nbConfig := []KVWithMetadata{
		{Key: prefixA + baseValue2, Value: test.NewArrayValue("item1"), Origin: FromNB, Metadata: &test.OnlyInteger{Integer: 0}},
	}
	sbState := []KVWithMetadata{
		{Key: prefixA + baseValue1, Value: test.NewStringValue(baseValue1), Origin: FromSB, Metadata: nil},
		{Key: prefixA + baseValue2, Value: test.NewArrayValue("item1"), Origin: FromNB, Metadata: &test.OnlyInteger{Integer: 0}},
	}
	views := []View{NBView, SBView, CachedView}
	for _, view := range views {
		var expValues []KVWithMetadata
		if view == NBView {
			expValues = nbConfig
		} else {
			expValues = sbState
		}
		dumpedValues, err := scheduler.DumpValuesByKeyPrefix(prefixA, view)
		Expect(err).To(BeNil())
		checkValues(dumpedValues, expValues)
		dumpedValues, err = scheduler.DumpValuesByDescriptor(descriptor1Name, view)
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
			State:         ValueState_OBTAINED,
			LastOperation: TxnOperation_UNDEFINED,
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
		DerivedValues: []*ValueStatus{
			{
				Key:           prefixA + baseValue2 + "/item1",
				State:         ValueState_CONFIGURED,
				LastOperation: TxnOperation_CREATE,
			},
		},
	})

	// check transaction operations
	txnHistory := scheduler.GetTransactionHistory(startTime, time.Now())
	Expect(txnHistory).To(HaveLen(1))
	txn := txnHistory[0]
	Expect(txn.PreRecord).To(BeFalse())
	Expect(txn.Start.After(startTime)).To(BeTrue())
	Expect(txn.Start.Before(txn.Stop)).To(BeTrue())
	Expect(txn.Stop.Before(stopTime)).To(BeTrue())
	Expect(txn.SeqNum).To(BeEquivalentTo(0))
	Expect(txn.TxnType).To(BeEquivalentTo(NBTransaction))
	Expect(txn.ResyncType).To(BeEquivalentTo(FullResync))
	Expect(txn.Description).To(BeEmpty())
	checkRecordedValues(txn.Values, []RecordedKVPair{
		{Key: prefixA + baseValue1, Value: utils.RecordProtoMessage(test.NewStringValue(baseValue1)), Origin: FromSB},
		{Key: prefixA + baseValue2, Value: utils.RecordProtoMessage(test.NewArrayValue("item1")), Origin: FromNB},
	})

	txnOps := RecordedTxnOps{
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixA + baseValue2,
			NewValue:  utils.RecordProtoMessage(test.NewArrayValue("item1")),
			PrevState: ValueState_NONEXISTENT,
			NewState:  ValueState_CONFIGURED,
		},
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixA + baseValue2 + "/item1",
			IsDerived: true,
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("item1")),
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
	Expect(pendingStats.TotalCount).To(BeEquivalentTo(0))
	derivedStats := graphR.GetFlagStats(DerivedFlagIndex, nil)
	Expect(derivedStats.TotalCount).To(BeEquivalentTo(1))
	lastUpdateStats := graphR.GetFlagStats(LastUpdateFlagIndex, nil)
	Expect(lastUpdateStats.TotalCount).To(BeEquivalentTo(3))
	descriptorStats := graphR.GetFlagStats(DescriptorFlagIndex, nil)
	Expect(descriptorStats.TotalCount).To(BeEquivalentTo(3))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor1Name))
	Expect(descriptorStats.PerValueCount[descriptor1Name]).To(BeEquivalentTo(3))
	valueStateStats := graphR.GetFlagStats(ValueStateFlagIndex, nil)
	Expect(valueStateStats.TotalCount).To(BeEquivalentTo(3))
	Expect(valueStateStats.PerValueCount).To(HaveKey(ValueState_CONFIGURED.String()))
	Expect(valueStateStats.PerValueCount[ValueState_CONFIGURED.String()]).To(BeEquivalentTo(2))
	Expect(valueStateStats.PerValueCount).To(HaveKey(ValueState_OBTAINED.String()))
	Expect(valueStateStats.PerValueCount[ValueState_OBTAINED.String()]).To(BeEquivalentTo(1))
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
	mockSB.SetValue(prefixA+baseValue1, test.NewArrayValue("item1"),
		&test.OnlyInteger{Integer: 0}, FromNB, false)
	mockSB.SetValue(prefixA+baseValue1+"/item1", test.NewStringValue("item1"),
		nil, FromNB, true)
	mockSB.SetValue(prefixB+baseValue2, test.NewArrayValue("item1"),
		&test.OnlyInteger{Integer: 0}, FromNB, false)
	mockSB.SetValue(prefixB+baseValue2+"/item1", test.NewStringValue("item1"),
		nil, FromNB, true)
	mockSB.SetValue(prefixC+baseValue3, test.NewArrayValue("item1"),
		&test.OnlyInteger{Integer: 0}, FromNB, false)
	mockSB.SetValue(prefixC+baseValue3+"/item1", test.NewStringValue("item1"),
		nil, FromNB, true)
	// -> descriptor1:
	descriptor1 := test.NewMockDescriptor(&KVDescriptor{
		Name:          descriptor1Name,
		NBKeyPrefix:   prefixA,
		KeySelector:   prefixSelector(prefixA),
		ValueTypeName: proto.MessageName(test.NewArrayValue()),
		DerivedValues: test.ArrayValueDerBuilder,
		WithMetadata:  true,
	}, mockSB, 1)
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
	}, mockSB, 1)
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
	metadataMap := scheduler.GetMetadataMap(descriptor1.Name)
	nameToInteger1, withMetadataMap := metadataMap.(test.NameToInteger)
	Expect(withMetadataMap).To(BeTrue())
	metadataMap = scheduler.GetMetadataMap(descriptor2.Name)
	nameToInteger2, withMetadataMap := metadataMap.(test.NameToInteger)
	Expect(withMetadataMap).To(BeTrue())
	metadataMap = scheduler.GetMetadataMap(descriptor3.Name)
	nameToInteger3, withMetadataMap := metadataMap.(test.NameToInteger)
	Expect(withMetadataMap).To(BeTrue())

	// run resync transaction with SB that already has some values added
	startTime := time.Now()
	schedulerTxn := scheduler.StartNBTransaction()
	schedulerTxn.SetValue(prefixB+baseValue2, test.NewArrayValue("item1", "item2"))
	schedulerTxn.SetValue(prefixA+baseValue1, test.NewArrayValue("item2"))
	schedulerTxn.SetValue(prefixC+baseValue3, test.NewArrayValue("item1", "item2"))
	seqNum, err := schedulerTxn.Commit(WithResync(testCtx, FullResync, true))
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
	// -> item1 derived from base value 1 was removed
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
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(1))
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
	Expect(metadata.GetInteger()).To(BeEquivalentTo(1))

	// check operations executed in SB
	opHistory := mockSB.PopHistoryOfOps()
	Expect(opHistory).To(HaveLen(12))
	operation := opHistory[0]
	Expect(operation.OpType).To(Equal(test.MockRetrieve))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	checkValues(operation.CorrelateRetrieve, []KVWithMetadata{
		{
			Key:      prefixA + baseValue1,
			Value:    test.NewArrayValue("item2"),
			Metadata: nil,
			Origin:   FromNB,
		},
	})
	operation = opHistory[1]
	Expect(operation.OpType).To(Equal(test.MockRetrieve))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	checkValues(operation.CorrelateRetrieve, []KVWithMetadata{
		{
			Key:      prefixB + baseValue2,
			Value:    test.NewArrayValue("item1", "item2"),
			Metadata: nil,
			Origin:   FromNB,
		},
	})
	operation = opHistory[2]
	Expect(operation.OpType).To(Equal(test.MockRetrieve))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	checkValues(operation.CorrelateRetrieve, []KVWithMetadata{
		{
			Key:      prefixC + baseValue3,
			Value:    test.NewArrayValue("item1", "item2"),
			Metadata: nil,
			Origin:   FromNB,
		},
	})
	operation = opHistory[3]
	Expect(operation.OpType).To(Equal(test.MockDelete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[4]
	Expect(operation.OpType).To(Equal(test.MockDelete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[5]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[6]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[7]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3 + "/item2"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[8]
	Expect(operation.OpType).To(Equal(test.MockDelete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[9]
	Expect(operation.OpType).To(Equal(test.MockUpdate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[10]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item2"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[11]
	Expect(operation.OpType).To(Equal(test.MockUpdate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixB + baseValue2))
	Expect(operation.Err).To(BeNil())

	// check transaction operations
	txnHistory := scheduler.GetTransactionHistory(time.Time{}, time.Time{})
	Expect(txnHistory).To(HaveLen(1))
	txn := txnHistory[0]
	Expect(txn.PreRecord).To(BeFalse())
	Expect(txn.Start.After(startTime)).To(BeTrue())
	Expect(txn.Start.Before(txn.Stop)).To(BeTrue())
	Expect(txn.Stop.Before(stopTime)).To(BeTrue())
	Expect(txn.SeqNum).To(BeEquivalentTo(0))
	Expect(txn.TxnType).To(BeEquivalentTo(NBTransaction))
	Expect(txn.ResyncType).To(BeEquivalentTo(FullResync))
	Expect(txn.Description).To(BeEmpty())
	checkRecordedValues(txn.Values, []RecordedKVPair{
		{Key: prefixA + baseValue1, Value: utils.RecordProtoMessage(test.NewArrayValue("item2")), Origin: FromNB},
		{Key: prefixB + baseValue2, Value: utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")), Origin: FromNB},
		{Key: prefixC + baseValue3, Value: utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")), Origin: FromNB},
	})

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
			{Key: prefixC + baseValue3, Value: test.NewArrayValue("item1", "item2"), Origin: FromNB, Metadata: &test.OnlyInteger{Integer: 1}},
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
	status = scheduler.GetValueStatus(prefixB + baseValue2)
	Expect(status).ToNot(BeNil())
	checkBaseValueStatus(status, &BaseValueStatus{
		Value: &ValueStatus{
			Key:           prefixB + baseValue2,
			State:         ValueState_CONFIGURED,
			LastOperation: TxnOperation_UPDATE,
		},
		DerivedValues: []*ValueStatus{
			{
				Key:           prefixB + baseValue2 + "/item1",
				State:         ValueState_CONFIGURED,
				LastOperation: TxnOperation_UPDATE,
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
			LastOperation: TxnOperation_UPDATE,
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

	txnOps := RecordedTxnOps{
		{
			Operation:  TxnOperation_DELETE,
			Key:        prefixC + baseValue3 + "/item1",
			IsDerived:  true,
			PrevValue:  utils.RecordProtoMessage(test.NewStringValue("item1")),
			PrevState:  ValueState_DISCOVERED,
			NewState:   ValueState_REMOVED,
			IsRecreate: true,
		},
		{
			Operation:  TxnOperation_DELETE,
			Key:        prefixC + baseValue3,
			PrevValue:  utils.RecordProtoMessage(test.NewArrayValue("item1")),
			PrevState:  ValueState_DISCOVERED,
			NewState:   ValueState_REMOVED,
			IsRecreate: true,
		},
		{
			Operation:  TxnOperation_CREATE,
			Key:        prefixC + baseValue3,
			NewValue:   utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")),
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
			Operation: TxnOperation_CREATE,
			Key:       prefixC + baseValue3 + "/item2",
			IsDerived: true,
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("item2")),
			PrevState: ValueState_NONEXISTENT,
			NewState:  ValueState_CONFIGURED,
		},
		{
			Operation: TxnOperation_DELETE,
			Key:       prefixA + baseValue1 + "/item1",
			IsDerived: true,
			PrevValue: utils.RecordProtoMessage(test.NewStringValue("item1")),
			PrevState: ValueState_DISCOVERED,
			NewState:  ValueState_REMOVED,
		},
		{
			Operation: TxnOperation_UPDATE,
			Key:       prefixA + baseValue1,
			PrevValue: utils.RecordProtoMessage(test.NewArrayValue("item1")),
			NewValue:  utils.RecordProtoMessage(test.NewArrayValue("item2")),
			PrevState: ValueState_DISCOVERED,
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
			Operation: TxnOperation_UPDATE,
			Key:       prefixB + baseValue2,
			PrevValue: utils.RecordProtoMessage(test.NewArrayValue("item1")),
			NewValue:  utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")),
			PrevState: ValueState_DISCOVERED,
			NewState:  ValueState_CONFIGURED,
		},
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixB + baseValue2 + "/item2",
			IsDerived: true,
			NOOP:      true,
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("item2")),
			PrevState: ValueState_NONEXISTENT,
			NewState:  ValueState_PENDING,
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
	mockSB.SetValue(prefixA+baseValue1, test.NewArrayValue(),
		&test.OnlyInteger{Integer: 0}, FromNB, false)
	// -> descriptor1:
	descriptor1 := test.NewMockDescriptor(&KVDescriptor{
		Name:          descriptor1Name,
		NBKeyPrefix:   prefixA,
		KeySelector:   prefixSelector(prefixA),
		ValueTypeName: proto.MessageName(test.NewArrayValue()),
		DerivedValues: test.ArrayValueDerBuilder,
		WithMetadata:  true,
	}, mockSB, 1)
	// -> planned error
	mockSB.PlanError(prefixA+baseValue1+"/item2", errors.New("failed to add value"),
		func() {
			mockSB.SetValue(prefixA+baseValue1, test.NewArrayValue("item1"),
				&test.OnlyInteger{Integer: 0}, FromNB, false)
		})

	// register descriptor with the scheduler
	scheduler.RegisterKVDescriptor(descriptor1)

	// subscribe to receive notifications about value state changes
	errorChan := make(chan *BaseValueStatus, 5)
	scheduler.WatchValueStatus(errorChan, prefixSelector(prefixA))

	// get metadata map created for the descriptor
	metadataMap := scheduler.GetMetadataMap(descriptor1.Name)
	nameToInteger, withMetadataMap := metadataMap.(test.NameToInteger)
	Expect(withMetadataMap).To(BeTrue())

	// run resync transaction that will fail for one value
	startTime := time.Now()
	resyncTxn := scheduler.StartNBTransaction()
	resyncTxn.SetValue(prefixA+baseValue1, test.NewArrayValue("item1", "item2"))
	description := "testing resync with retry"
	ctx := testCtx
	ctx = WithRetry(ctx, 3*time.Second, 3, false)
	ctx = WithResync(ctx, FullResync, true)
	ctx = WithDescription(ctx, description)
	seqNum, err := resyncTxn.Commit(ctx)
	stopTime := time.Now()
	Expect(seqNum).To(BeEquivalentTo(0))
	Expect(err).ToNot(BeNil())
	txnErr := err.(*TransactionError)
	Expect(txnErr.GetTxnInitError()).ShouldNot(HaveOccurred())
	kvErrors := txnErr.GetKVErrors()
	Expect(kvErrors).To(HaveLen(1))
	Expect(kvErrors[0].TxnOperation).To(BeEquivalentTo(TxnOperation_CREATE))
	Expect(kvErrors[0].Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item2"))
	Expect(kvErrors[0].Error.Error()).To(BeEquivalentTo("failed to add value"))

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
	Expect(operation.OpType).To(Equal(test.MockRetrieve))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	checkValues(operation.CorrelateRetrieve, []KVWithMetadata{
		{
			Key:      prefixA + baseValue1,
			Value:    test.NewArrayValue("item1", "item2"),
			Metadata: nil,
			Origin:   FromNB,
		},
	})
	operation = opHistory[1]
	Expect(operation.OpType).To(Equal(test.MockUpdate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[2]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[3]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item2"))
	Expect(operation.Err).ToNot(BeNil())
	Expect(operation.Err.Error()).To(BeEquivalentTo("failed to add value"))
	operation = opHistory[4] // refresh failed value
	Expect(operation.OpType).To(Equal(test.MockRetrieve))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	checkValues(operation.CorrelateRetrieve, []KVWithMetadata{
		{
			Key:      prefixA + baseValue1,
			Value:    test.NewArrayValue("item1", "item2"),
			Metadata: &test.OnlyInteger{Integer: 0},
			Origin:   FromNB,
		},
	})

	// check transaction operations
	txnHistory := scheduler.GetTransactionHistory(time.Time{}, time.Time{})
	Expect(txnHistory).To(HaveLen(1))
	txn := txnHistory[0]
	Expect(txn.PreRecord).To(BeFalse())
	Expect(txn.Start.After(startTime)).To(BeTrue())
	Expect(txn.Start.Before(txn.Stop)).To(BeTrue())
	Expect(txn.Stop.Before(stopTime)).To(BeTrue())
	Expect(txn.SeqNum).To(BeEquivalentTo(0))
	Expect(txn.TxnType).To(BeEquivalentTo(NBTransaction))
	Expect(txn.ResyncType).To(BeEquivalentTo(FullResync))
	Expect(txn.Description).To(Equal(description))
	checkRecordedValues(txn.Values, []RecordedKVPair{
		{Key: prefixA + baseValue1, Value: utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")), Origin: FromNB},
	})

	txnOps := RecordedTxnOps{
		{
			Operation: TxnOperation_UPDATE,
			Key:       prefixA + baseValue1,
			PrevValue: utils.RecordProtoMessage(test.NewArrayValue()),
			NewValue:  utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")),
			PrevState: ValueState_DISCOVERED,
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
			Key:       prefixA + baseValue1 + "/item2",
			IsDerived: true,
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("item2")),
			PrevState: ValueState_NONEXISTENT,
			NewState:  ValueState_CONFIGURED,
		},
	}
	checkTxnOperations(txn.Planned, txnOps)
	txnOps[2].NewState = ValueState_RETRYING
	txnOps[2].NewErr = errors.New("failed to add value")
	checkTxnOperations(txn.Executed, txnOps)

	// check flag stats
	graphR := scheduler.graph.Read()
	errorStats := graphR.GetFlagStats(ErrorFlagIndex, nil)
	Expect(errorStats.TotalCount).To(BeEquivalentTo(1))
	pendingStats := graphR.GetFlagStats(UnavailValueFlagIndex, nil)
	Expect(pendingStats.TotalCount).To(BeEquivalentTo(1))
	derivedStats := graphR.GetFlagStats(DerivedFlagIndex, nil)
	Expect(derivedStats.TotalCount).To(BeEquivalentTo(2))
	lastUpdateStats := graphR.GetFlagStats(LastUpdateFlagIndex, nil)
	Expect(lastUpdateStats.TotalCount).To(BeEquivalentTo(3))
	descriptorStats := graphR.GetFlagStats(DescriptorFlagIndex, nil)
	Expect(descriptorStats.TotalCount).To(BeEquivalentTo(3))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor1Name))
	Expect(descriptorStats.PerValueCount[descriptor1Name]).To(BeEquivalentTo(3))
	valueStateStats := graphR.GetFlagStats(ValueStateFlagIndex, nil)
	Expect(valueStateStats.TotalCount).To(BeEquivalentTo(3))
	Expect(valueStateStats.PerValueCount).To(HaveKey(ValueState_CONFIGURED.String()))
	Expect(valueStateStats.PerValueCount[ValueState_CONFIGURED.String()]).To(BeEquivalentTo(2))
	Expect(valueStateStats.PerValueCount).To(HaveKey(ValueState_RETRYING.String()))
	Expect(valueStateStats.PerValueCount[ValueState_RETRYING.String()]).To(BeEquivalentTo(1))
	graphR.Release()

	// check value state updates received through the channel
	var valueStatus *BaseValueStatus
	Eventually(errorChan, time.Second).Should(Receive(&valueStatus))
	checkBaseValueStatus(valueStatus, &BaseValueStatus{
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
			{
				Key:           prefixA + baseValue1 + "/item2",
				State:         ValueState_RETRYING,
				LastOperation: TxnOperation_CREATE,
				Error:         "failed to add value",
			},
		},
	})

	// eventually the value should get "fixed"
	Eventually(errorChan, 5*time.Second).Should(Receive(&valueStatus))
	checkBaseValueStatus(valueStatus, &BaseValueStatus{
		Value: &ValueStatus{
			Key:           prefixA + baseValue1,
			State:         ValueState_CONFIGURED,
			LastOperation: TxnOperation_UPDATE,
		},
		DerivedValues: []*ValueStatus{
			{
				Key:           prefixA + baseValue1 + "/item1",
				State:         ValueState_CONFIGURED,
				LastOperation: TxnOperation_UPDATE,
			},
			{
				Key:           prefixA + baseValue1 + "/item2",
				State:         ValueState_CONFIGURED,
				LastOperation: TxnOperation_CREATE,
			},
		},
	})

	// check the state of SB after retry
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	// -> base value 1
	value = mockSB.GetValue(prefixA + baseValue1)
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewArrayValue("item1", "item2"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(0))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item1 derived from base value 1
	value = mockSB.GetValue(prefixA + baseValue1 + "/item1")
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("item1"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	Expect(mockSB.GetValues(nil)).To(HaveLen(3))
	// -> item2 derived from base value 1 was re-added
	value = mockSB.GetValue(prefixA + baseValue1 + "/item2")
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("item2"))).To(BeTrue())
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
	Expect(operation.OpType).To(Equal(test.MockUpdate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[1]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item2"))
	Expect(operation.Err).To(BeNil())

	// check retry transaction operations
	txnHistory = scheduler.GetTransactionHistory(time.Time{}, time.Now())
	Expect(txnHistory).To(HaveLen(2))
	txn = txnHistory[1]
	Expect(txn.PreRecord).To(BeFalse())
	Expect(txn.Start.After(stopTime)).To(BeTrue())
	Expect(txn.Start.Before(txn.Stop)).To(BeTrue())
	Expect(txn.Stop.Before(time.Now())).To(BeTrue())
	Expect(txn.SeqNum).To(BeEquivalentTo(1))
	Expect(txn.TxnType).To(BeEquivalentTo(RetryFailedOps))
	Expect(txn.ResyncType).To(BeEquivalentTo(NotResync))
	Expect(txn.Description).To(BeEmpty())
	checkRecordedValues(txn.Values, []RecordedKVPair{
		{Key: prefixA + baseValue1, Value: utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")), Origin: FromNB},
	})

	txnOps = RecordedTxnOps{
		{
			Operation: TxnOperation_UPDATE,
			Key:       prefixA + baseValue1,
			PrevValue: utils.RecordProtoMessage(test.NewArrayValue("item1")),
			NewValue:  utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")),
			PrevState: ValueState_CONFIGURED,
			NewState:  ValueState_CONFIGURED,
			IsRetry:   true,
		},
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixA + baseValue1 + "/item2",
			IsDerived: true,
			PrevValue: utils.RecordProtoMessage(test.NewStringValue("item2")), // TODO: shouldn't be nil?
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("item2")),
			PrevState: ValueState_RETRYING,
			NewState:  ValueState_CONFIGURED,
			PrevErr:   errors.New("failed to add value"),
			IsRetry:   true,
		},
	}
	checkTxnOperations(txn.Planned, txnOps)
	checkTxnOperations(txn.Executed, txnOps)

	// check flag stats
	graphR = scheduler.graph.Read()
	errorStats = graphR.GetFlagStats(ErrorFlagIndex, nil)
	Expect(errorStats.TotalCount).To(BeEquivalentTo(1))
	pendingStats = graphR.GetFlagStats(UnavailValueFlagIndex, nil)
	Expect(pendingStats.TotalCount).To(BeEquivalentTo(1))
	derivedStats = graphR.GetFlagStats(DerivedFlagIndex, nil)
	Expect(derivedStats.TotalCount).To(BeEquivalentTo(4))
	lastUpdateStats = graphR.GetFlagStats(LastUpdateFlagIndex, nil)
	Expect(lastUpdateStats.TotalCount).To(BeEquivalentTo(6))
	descriptorStats = graphR.GetFlagStats(DescriptorFlagIndex, nil)
	Expect(descriptorStats.TotalCount).To(BeEquivalentTo(6))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor1Name))
	Expect(descriptorStats.PerValueCount[descriptor1Name]).To(BeEquivalentTo(6))
	valueStateStats = graphR.GetFlagStats(ValueStateFlagIndex, nil)
	Expect(valueStateStats.TotalCount).To(BeEquivalentTo(6))
	Expect(valueStateStats.PerValueCount).To(HaveKey(ValueState_CONFIGURED.String()))
	Expect(valueStateStats.PerValueCount[ValueState_CONFIGURED.String()]).To(BeEquivalentTo(5))
	Expect(valueStateStats.PerValueCount).To(HaveKey(ValueState_RETRYING.String()))
	Expect(valueStateStats.PerValueCount[ValueState_RETRYING.String()]).To(BeEquivalentTo(1))
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
