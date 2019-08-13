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
	"strings"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/gomega"

	. "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/internal/test"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/internal/utils"
)

func TestNotifications(t *testing.T) {
	RegisterTestingT(t)

	// prepare KV Scheduler
	scheduler := NewPlugin(UseDeps(func(deps *Deps) {
		deps.HTTPHandlers = nil
	}))
	err := scheduler.Init()
	Expect(err).To(BeNil())
	scheduler.config.EnableTxnSimulation = true

	// prepare mocks
	mockSB := test.NewMockSouthbound()
	// -> descriptor1 (notifications):
	descriptor1 := test.NewMockDescriptor(&KVDescriptor{
		Name:        descriptor1Name,
		NBKeyPrefix: prefixA,
		KeySelector: func(key string) bool {
			if !strings.HasPrefix(key, prefixA) {
				return false
			}
			if strings.Contains(strings.TrimPrefix(key, prefixA), "/") {
				return false // exclude derived values
			}
			return true
		},
		ValueTypeName: proto.MessageName(test.NewArrayValue()),
		DerivedValues: test.ArrayValueDerBuilder,
		WithMetadata:  true,
	}, mockSB, 0, test.WithoutRetrieve)
	// -> descriptor2:
	descriptor2 := test.NewMockDescriptor(&KVDescriptor{
		Name:          descriptor2Name,
		NBKeyPrefix:   prefixB,
		KeySelector:   prefixSelector(prefixB),
		ValueTypeName: proto.MessageName(test.NewArrayValue()),
		DerivedValues: test.ArrayValueDerBuilder,
		Dependencies: func(key string, value proto.Message) []Dependency {
			if key == prefixB+baseValue2 {
				depKey := prefixA + baseValue1
				return []Dependency{
					{Label: depKey, Key: depKey},
				}
			}
			if key == prefixB+baseValue2+"/item2" {
				depKey := prefixA + baseValue1 + "/item2"
				return []Dependency{
					{Label: depKey, Key: depKey},
				}
			}
			return nil
		},
		WithMetadata:         true,
		RetrieveDependencies: []string{descriptor1Name},
	}, mockSB, 0)

	// register both descriptors with the scheduler
	scheduler.RegisterKVDescriptor(descriptor1)
	scheduler.RegisterKVDescriptor(descriptor2)

	// get metadata map created for each descriptor
	metadataMap := scheduler.GetMetadataMap(descriptor1.Name)
	nameToInteger1, withMetadataMap := metadataMap.(test.NameToInteger)
	Expect(withMetadataMap).To(BeTrue())
	metadataMap = scheduler.GetMetadataMap(descriptor2.Name)
	nameToInteger2, withMetadataMap := metadataMap.(test.NameToInteger)
	Expect(withMetadataMap).To(BeTrue())

	// run resync transaction against empty SB
	startTime := time.Now()
	schedulerTxn := scheduler.StartNBTransaction()
	schedulerTxn.SetValue(prefixB+baseValue2, test.NewArrayValue("item1", "item2"))
	seqNum, err := schedulerTxn.Commit(WithResync(testCtx, FullResync, true))
	stopTime := time.Now()
	Expect(seqNum).To(BeEquivalentTo(0))
	Expect(err).ShouldNot(HaveOccurred())

	// check the state of SB
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	Expect(mockSB.GetValues(nil)).To(BeEmpty())

	// check metadata
	Expect(metadataMap.ListAllNames()).To(BeEmpty())

	// check operations executed in SB
	opHistory := mockSB.PopHistoryOfOps()
	Expect(opHistory).To(HaveLen(1))
	operation := opHistory[0]
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
	Expect(txn.ResyncType).To(BeEquivalentTo(FullResync))
	Expect(txn.Description).To(BeEmpty())
	checkRecordedValues(txn.Values, []RecordedKVPair{
		{Key: prefixB + baseValue2, Value: utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")), Origin: FromNB},
	})

	txnOps := RecordedTxnOps{
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixB + baseValue2,
			NewValue:  utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")),
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
	Expect(pendingStats.TotalCount).To(BeEquivalentTo(1))
	derivedStats := graphR.GetFlagStats(DerivedFlagIndex, nil)
	Expect(derivedStats.TotalCount).To(BeEquivalentTo(0))
	lastUpdateStats := graphR.GetFlagStats(LastUpdateFlagIndex, nil)
	Expect(lastUpdateStats.TotalCount).To(BeEquivalentTo(1))
	descriptorStats := graphR.GetFlagStats(DescriptorFlagIndex, nil)
	Expect(descriptorStats.TotalCount).To(BeEquivalentTo(1))
	Expect(descriptorStats.PerValueCount).ToNot(HaveKey(descriptor1Name))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor2Name))
	Expect(descriptorStats.PerValueCount[descriptor2Name]).To(BeEquivalentTo(1))
	valueStateStats := graphR.GetFlagStats(ValueStateFlagIndex, nil)
	Expect(valueStateStats.TotalCount).To(BeEquivalentTo(1))
	Expect(valueStateStats.PerValueCount).To(HaveKey(ValueState_PENDING.String()))
	Expect(valueStateStats.PerValueCount[ValueState_PENDING.String()]).To(BeEquivalentTo(1))
	graphR.Release()

	// check value dumps for prefix B
	nbConfig := []KVWithMetadata{
		{Key: prefixB + baseValue2, Value: test.NewArrayValue("item1", "item2"), Origin: FromNB},
	}
	views := []View{NBView, SBView, CachedView}
	for _, view := range views {
		var expValues []KVWithMetadata
		if view == NBView {
			expValues = nbConfig
		} // else empty expected set of values
		dumpedValues, err := scheduler.DumpValuesByKeyPrefix(prefixB, view)
		Expect(err).To(BeNil())
		checkValues(dumpedValues, expValues)
		dumpedValues, err = scheduler.DumpValuesByDescriptor(descriptor2Name, view)
		Expect(err).To(BeNil())
		checkValues(dumpedValues, expValues)
	}
	mockSB.PopHistoryOfOps() // remove Retrieve-s from the history

	// check value status
	status := scheduler.GetValueStatus(prefixB + baseValue2)
	Expect(status).ToNot(BeNil())
	checkBaseValueStatus(status, &BaseValueStatus{
		Value: &ValueStatus{
			Key:           prefixB + baseValue2,
			State:         ValueState_PENDING,
			LastOperation: TxnOperation_CREATE,
			Details:       []string{prefixA + baseValue1},
		},
	})

	// subscribe to receive notifications about value state changes for prefixA
	statusChan := make(chan *BaseValueStatus, 5)
	scheduler.WatchValueStatus(statusChan, prefixSelector(prefixA))

	// send notification
	startTime = time.Now()
	mockSB.SetValue(prefixA+baseValue1, test.NewArrayValue("item1"), &test.OnlyInteger{Integer: 10}, FromSB, false)
	notifError := scheduler.PushSBNotification(prefixA+baseValue1, test.NewArrayValue("item1"),
		&test.OnlyInteger{Integer: 10})
	Expect(notifError).ShouldNot(HaveOccurred())

	// wait until the notification is processed
	Eventually(func() []*KVWithMetadata {
		return mockSB.GetValues(nil)
	}, 2*time.Second).Should(HaveLen(3))
	stopTime = time.Now()

	// check the state of SB
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	// -> base value 1
	value := mockSB.GetValue(prefixA + baseValue1)
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewArrayValue("item1"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(10))
	Expect(value.Origin).To(BeEquivalentTo(FromSB))
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
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixB + baseValue2))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[1]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixB + baseValue2 + "/item1"))
	Expect(operation.Err).To(BeNil())

	// check value dumps for prefix A
	sbState := []KVWithMetadata{
		{Key: prefixA + baseValue1, Value: test.NewArrayValue("item1"), Origin: FromSB, Metadata: &test.OnlyInteger{Integer: 10}},
	}
	for _, view := range views {
		// prefix
		var expValues []KVWithMetadata
		expErrMsg := "descriptor does not support Retrieve operation" // Retrieve not supported
		if view != NBView {
			expValues = sbState
		} // else empty set of dumped values
		dumpedValues, err := scheduler.DumpValuesByKeyPrefix(prefixA, view)
		if view == SBView {
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(BeEquivalentTo(expErrMsg))
		} else {
			Expect(err).To(BeNil())
			checkValues(dumpedValues, expValues)
		}
		dumpedValues, err = scheduler.DumpValuesByDescriptor(descriptor1Name, view)
		if view == SBView {
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(BeEquivalentTo(expErrMsg))
		} else {
			Expect(err).To(BeNil())
			checkValues(dumpedValues, expValues)
		}
	}
	mockSB.PopHistoryOfOps() // remove Retrieve-s from the history

	// check value status
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
				Details:       []string{prefixA + baseValue1 + "/item2"},
			},
		},
	})
	status = scheduler.GetValueStatus(prefixA + baseValue1)
	Expect(status).ToNot(BeNil())
	checkBaseValueStatus(status, &BaseValueStatus{
		Value: &ValueStatus{
			Key:           prefixA + baseValue1,
			State:         ValueState_OBTAINED,
			LastOperation: TxnOperation_UNDEFINED,
		},
		DerivedValues: []*ValueStatus{
			{
				Key:           prefixA + baseValue1 + "/item1",
				State:         ValueState_OBTAINED,
				LastOperation: TxnOperation_UNDEFINED,
			},
		},
	})

	// check transaction operations
	txnHistory = scheduler.GetTransactionHistory(startTime, time.Now())
	Expect(txnHistory).To(HaveLen(1))
	txn = txnHistory[0]
	Expect(txn.PreRecord).To(BeFalse())
	Expect(txn.Start.After(startTime)).To(BeTrue())
	Expect(txn.Start.Before(txn.Stop)).To(BeTrue())
	Expect(txn.Stop.Before(stopTime)).To(BeTrue())
	Expect(txn.SeqNum).To(BeEquivalentTo(1))
	Expect(txn.TxnType).To(BeEquivalentTo(SBNotification))
	Expect(txn.ResyncType).To(BeEquivalentTo(NotResync))
	Expect(txn.Description).To(BeEmpty())
	checkRecordedValues(txn.Values, []RecordedKVPair{
		{Key: prefixA + baseValue1, Value: utils.RecordProtoMessage(test.NewArrayValue("item1")), Origin: FromSB},
	})

	txnOps = RecordedTxnOps{
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixA + baseValue1,
			NewValue:  utils.RecordProtoMessage(test.NewArrayValue("item1")),
			PrevState: ValueState_NONEXISTENT,
			NewState:  ValueState_OBTAINED,
		},
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixB + baseValue2,
			PrevValue: utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")),
			NewValue:  utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")),
			PrevState: ValueState_PENDING,
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
			Operation:  TxnOperation_CREATE,
			Key:        prefixA + baseValue1 + "/item1",
			IsDerived:  true,
			IsProperty: true,
			NewValue:   utils.RecordProtoMessage(test.NewStringValue("item1")),
			PrevState:  ValueState_NONEXISTENT,
			NewState:   ValueState_OBTAINED,
		},
	}
	checkTxnOperations(txn.Planned, txnOps)
	checkTxnOperations(txn.Executed, txnOps)

	// check flag stats
	graphR = scheduler.graph.Read()
	errorStats = graphR.GetFlagStats(ErrorFlagIndex, nil)
	Expect(errorStats.TotalCount).To(BeEquivalentTo(0))
	pendingStats = graphR.GetFlagStats(UnavailValueFlagIndex, nil)
	Expect(pendingStats.TotalCount).To(BeEquivalentTo(2))
	derivedStats = graphR.GetFlagStats(DerivedFlagIndex, nil)
	Expect(derivedStats.TotalCount).To(BeEquivalentTo(3))
	lastUpdateStats = graphR.GetFlagStats(LastUpdateFlagIndex, nil)
	Expect(lastUpdateStats.TotalCount).To(BeEquivalentTo(6))
	descriptorStats = graphR.GetFlagStats(DescriptorFlagIndex, nil)
	Expect(descriptorStats.TotalCount).To(BeEquivalentTo(5))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor1Name))
	Expect(descriptorStats.PerValueCount[descriptor1Name]).To(BeEquivalentTo(1))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor2Name))
	Expect(descriptorStats.PerValueCount[descriptor2Name]).To(BeEquivalentTo(4))
	valueStateStats = graphR.GetFlagStats(ValueStateFlagIndex, nil)
	Expect(valueStateStats.TotalCount).To(BeEquivalentTo(6))
	Expect(valueStateStats.PerValueCount).To(HaveKey(ValueState_PENDING.String()))
	Expect(valueStateStats.PerValueCount[ValueState_PENDING.String()]).To(BeEquivalentTo(2))
	Expect(valueStateStats.PerValueCount).To(HaveKey(ValueState_CONFIGURED.String()))
	Expect(valueStateStats.PerValueCount[ValueState_CONFIGURED.String()]).To(BeEquivalentTo(2))
	Expect(valueStateStats.PerValueCount).To(HaveKey(ValueState_OBTAINED.String()))
	Expect(valueStateStats.PerValueCount[ValueState_OBTAINED.String()]).To(BeEquivalentTo(2))
	graphR.Release()

	// send 2nd notification
	startTime = time.Now()
	mockSB.SetValue(prefixA+baseValue1, test.NewArrayValue("item1", "item2"), &test.OnlyInteger{Integer: 11}, FromSB, false)
	notifError = scheduler.PushSBNotification(prefixA+baseValue1, test.NewArrayValue("item1", "item2"),
		&test.OnlyInteger{Integer: 11})
	Expect(notifError).ShouldNot(HaveOccurred())

	// wait until the notification is processed
	Eventually(func() []*KVWithMetadata {
		return mockSB.GetValues(nil)
	}, 2*time.Second).Should(HaveLen(4))
	stopTime = time.Now()

	// check the state of SB
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	// -> base value 1
	value = mockSB.GetValue(prefixA + baseValue1)
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewArrayValue("item1", "item2"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(11))
	Expect(value.Origin).To(BeEquivalentTo(FromSB))
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
	// -> item2 derived from base value 2
	value = mockSB.GetValue(prefixB + baseValue2 + "/item2")
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("item2"))).To(BeTrue())
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
	Expect(opHistory).To(HaveLen(1))
	operation = opHistory[0]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixB + baseValue2 + "/item2"))
	Expect(operation.Err).To(BeNil())

	// check value status
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

	// check transaction operations
	txnHistory = scheduler.GetTransactionHistory(startTime, time.Now())
	Expect(txnHistory).To(HaveLen(1))
	txn = txnHistory[0]
	Expect(txn.PreRecord).To(BeFalse())
	Expect(txn.Start.After(startTime)).To(BeTrue())
	Expect(txn.Start.Before(txn.Stop)).To(BeTrue())
	Expect(txn.Stop.Before(stopTime)).To(BeTrue())
	Expect(txn.SeqNum).To(BeEquivalentTo(2))
	Expect(txn.TxnType).To(BeEquivalentTo(SBNotification))
	Expect(txn.ResyncType).To(BeEquivalentTo(NotResync))
	Expect(txn.Description).To(BeEmpty())
	checkRecordedValues(txn.Values, []RecordedKVPair{
		{Key: prefixA + baseValue1, Value: utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")), Origin: FromSB},
	})

	txnOps = RecordedTxnOps{
		{
			Operation: TxnOperation_UPDATE,
			Key:       prefixA + baseValue1,
			PrevValue: utils.RecordProtoMessage(test.NewArrayValue("item1")),
			NewValue:  utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")),
			PrevState: ValueState_OBTAINED,
			NewState:  ValueState_OBTAINED,
		},
		{
			Operation:  TxnOperation_CREATE,
			Key:        prefixA + baseValue1 + "/item2",
			IsDerived:  true,
			IsProperty: true,
			NewValue:   utils.RecordProtoMessage(test.NewStringValue("item2")),
			PrevState:  ValueState_NONEXISTENT,
			NewState:   ValueState_OBTAINED,
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

	// check flag stats
	graphR = scheduler.graph.Read()
	errorStats = graphR.GetFlagStats(ErrorFlagIndex, nil)
	Expect(errorStats.TotalCount).To(BeEquivalentTo(0))
	pendingStats = graphR.GetFlagStats(UnavailValueFlagIndex, nil)
	Expect(pendingStats.TotalCount).To(BeEquivalentTo(2))
	derivedStats = graphR.GetFlagStats(DerivedFlagIndex, nil)
	Expect(derivedStats.TotalCount).To(BeEquivalentTo(6))
	lastUpdateStats = graphR.GetFlagStats(LastUpdateFlagIndex, nil)
	Expect(lastUpdateStats.TotalCount).To(BeEquivalentTo(10))
	descriptorStats = graphR.GetFlagStats(DescriptorFlagIndex, nil)
	Expect(descriptorStats.TotalCount).To(BeEquivalentTo(7))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor1Name))
	Expect(descriptorStats.PerValueCount[descriptor1Name]).To(BeEquivalentTo(2))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor2Name))
	Expect(descriptorStats.PerValueCount[descriptor2Name]).To(BeEquivalentTo(5))
	valueStateStats = graphR.GetFlagStats(ValueStateFlagIndex, nil)
	Expect(valueStateStats.TotalCount).To(BeEquivalentTo(10))
	Expect(valueStateStats.PerValueCount).To(HaveKey(ValueState_PENDING.String()))
	Expect(valueStateStats.PerValueCount[ValueState_PENDING.String()]).To(BeEquivalentTo(2))
	Expect(valueStateStats.PerValueCount).To(HaveKey(ValueState_CONFIGURED.String()))
	Expect(valueStateStats.PerValueCount[ValueState_CONFIGURED.String()]).To(BeEquivalentTo(3))
	Expect(valueStateStats.PerValueCount).To(HaveKey(ValueState_OBTAINED.String()))
	Expect(valueStateStats.PerValueCount[ValueState_OBTAINED.String()]).To(BeEquivalentTo(5))
	graphR.Release()

	// send 3rd notification
	startTime = time.Now()
	mockSB.SetValue(prefixA+baseValue1, nil, nil, FromSB, false)
	notifError = scheduler.PushSBNotification(prefixA+baseValue1, nil, nil)
	Expect(notifError).ShouldNot(HaveOccurred())

	// wait until the notification is processed
	Eventually(func() bool {
		return len(mockSB.GetValues(nil)) == 0 && len(metadataMap.ListAllNames()) == 0
	}, 2*time.Second).Should(BeTrue())
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
	Expect(operation.OpType).To(Equal(test.MockDelete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixB + baseValue2 + "/item2"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[1]
	Expect(operation.OpType).To(Equal(test.MockDelete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixB + baseValue2 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[2]
	Expect(operation.OpType).To(Equal(test.MockDelete))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixB + baseValue2))
	Expect(operation.Err).To(BeNil())

	// check value status
	status = scheduler.GetValueStatus(prefixB + baseValue2)
	Expect(status).ToNot(BeNil())
	checkBaseValueStatus(status, &BaseValueStatus{
		Value: &ValueStatus{
			Key:           prefixB + baseValue2,
			State:         ValueState_PENDING,
			LastOperation: TxnOperation_DELETE,
			Details:       []string{prefixA + baseValue1},
		},
	})
	status = scheduler.GetValueStatus(prefixA + baseValue1)
	Expect(status).ToNot(BeNil())
	checkBaseValueStatus(status, &BaseValueStatus{
		Value: &ValueStatus{
			Key:           prefixA + baseValue1,
			State:         ValueState_NONEXISTENT,
			LastOperation: TxnOperation_UNDEFINED,
		},
	})

	// check transaction operations
	txnHistory = scheduler.GetTransactionHistory(startTime, time.Now())
	Expect(txnHistory).To(HaveLen(1))
	txn = txnHistory[0]
	Expect(txn.PreRecord).To(BeFalse())
	Expect(txn.Start.After(startTime)).To(BeTrue())
	Expect(txn.Start.Before(txn.Stop)).To(BeTrue())
	Expect(txn.Stop.Before(stopTime)).To(BeTrue())
	Expect(txn.SeqNum).To(BeEquivalentTo(3))
	Expect(txn.TxnType).To(BeEquivalentTo(SBNotification))
	Expect(txn.ResyncType).To(BeEquivalentTo(NotResync))
	Expect(txn.Description).To(BeEmpty())
	checkRecordedValues(txn.Values, []RecordedKVPair{
		{Key: prefixA + baseValue1, Value: utils.RecordProtoMessage(nil), Origin: FromSB},
	})

	txnOps = RecordedTxnOps{
		{
			Operation:  TxnOperation_DELETE,
			Key:        prefixA + baseValue1 + "/item1",
			IsDerived:  true,
			IsProperty: true,
			PrevValue:  utils.RecordProtoMessage(test.NewStringValue("item1")),
			PrevState:  ValueState_OBTAINED,
			NewState:   ValueState_REMOVED,
		},
		{
			Operation: TxnOperation_DELETE,
			Key:       prefixB + baseValue2 + "/item2",
			IsDerived: true,
			PrevValue: utils.RecordProtoMessage(test.NewStringValue("item2")),
			PrevState: ValueState_CONFIGURED,
			NewState:  ValueState_REMOVED,
		},
		{
			Operation:  TxnOperation_DELETE,
			Key:        prefixA + baseValue1 + "/item2",
			IsDerived:  true,
			IsProperty: true,
			PrevValue:  utils.RecordProtoMessage(test.NewStringValue("item2")),
			PrevState:  ValueState_OBTAINED,
			NewState:   ValueState_REMOVED,
		},
		{
			Operation: TxnOperation_DELETE,
			Key:       prefixB + baseValue2 + "/item1",
			IsDerived: true,
			PrevValue: utils.RecordProtoMessage(test.NewStringValue("item1")),
			PrevState: ValueState_CONFIGURED,
			NewState:  ValueState_REMOVED,
		},
		{
			Operation: TxnOperation_DELETE,
			Key:       prefixB + baseValue2,
			PrevValue: utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")),
			NewValue:  utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")), // TODO: do we want nil instead?
			PrevState: ValueState_CONFIGURED,
			NewState:  ValueState_PENDING,
		},
		{
			Operation: TxnOperation_DELETE,
			Key:       prefixA + baseValue1,
			PrevValue: utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")),
			PrevState: ValueState_OBTAINED,
			NewState:  ValueState_REMOVED,
		},
	}
	checkTxnOperations(txn.Planned, txnOps)
	checkTxnOperations(txn.Executed, txnOps)

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
	scheduler.config.EnableTxnSimulation = true

	// prepare mocks
	mockSB := test.NewMockSouthbound()
	// -> descriptor1 (notifications):
	descriptor1 := test.NewMockDescriptor(&KVDescriptor{
		Name:          descriptor1Name,
		NBKeyPrefix:   prefixA,
		KeySelector:   prefixSelector(prefixA),
		ValueTypeName: proto.MessageName(test.NewArrayValue()),
		DerivedValues: test.ArrayValueDerBuilder,
		WithMetadata:  true,
	}, mockSB, 0, test.WithoutRetrieve)
	// -> descriptor2:
	descriptor2 := test.NewMockDescriptor(&KVDescriptor{
		Name:          descriptor2Name,
		NBKeyPrefix:   prefixB,
		KeySelector:   prefixSelector(prefixB),
		ValueTypeName: proto.MessageName(test.NewArrayValue()),
		Dependencies: func(key string, value proto.Message) []Dependency {
			if key == prefixB+baseValue2 {
				depKey := prefixA + baseValue1
				return []Dependency{
					{Label: depKey, Key: depKey},
				}
			}
			if key == prefixB+baseValue2+"/item2" {
				depKey := prefixA + baseValue1 + "/item2"
				return []Dependency{
					{Label: depKey, Key: depKey},
				}
			}
			return nil
		},
		DerivedValues: test.ArrayValueDerBuilder,
		WithMetadata:  true,
	}, mockSB, 0)
	// -> descriptor3:
	descriptor3 := test.NewMockDescriptor(&KVDescriptor{
		Name:            descriptor3Name,
		NBKeyPrefix:     prefixC,
		KeySelector:     prefixSelector(prefixC),
		ValueTypeName:   proto.MessageName(test.NewStringValue("")),
		ValueComparator: test.StringValueComparator,
		Dependencies: func(key string, value proto.Message) []Dependency {
			if key == prefixC+baseValue3 {
				return []Dependency{
					{
						Label: prefixA,
						AnyOf: AnyOfDependency{
							KeyPrefixes: []string{prefixA},
						},
					},
				}
			}
			return nil
		},
		WithMetadata:         true,
		RetrieveDependencies: []string{descriptor2Name},
	}, mockSB, 0)

	// -> planned errors
	mockSB.PlanError(prefixB+baseValue2+"/item2", errors.New("failed to add derived value"),
		func() {
			mockSB.SetValue(prefixB+baseValue2, test.NewArrayValue("item1"),
				&test.OnlyInteger{Integer: 0}, FromNB, false)
		})
	mockSB.PlanError(prefixC+baseValue3, errors.New("failed to add value"),
		func() {
			mockSB.SetValue(prefixC+baseValue3, nil, nil, FromNB, false)
		})

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

	// run 1st data-change transaction with retry against empty SB
	schedulerTxn1 := scheduler.StartNBTransaction()
	schedulerTxn1.SetValue(prefixB+baseValue2, test.NewArrayValue("item1", "item2"))
	seqNum, err := schedulerTxn1.Commit(WithRetryDefault(testCtx))
	Expect(seqNum).To(BeEquivalentTo(0))
	Expect(err).ShouldNot(HaveOccurred())

	// run 2nd data-change transaction with retry
	schedulerTxn2 := scheduler.StartNBTransaction()
	schedulerTxn2.SetValue(prefixC+baseValue3, test.NewStringValue("base-value3-data"))
	seqNum, err = schedulerTxn2.Commit(WithRetry(testCtx, 3*time.Second, 3, true))
	Expect(seqNum).To(BeEquivalentTo(1))
	Expect(err).ShouldNot(HaveOccurred())

	// check the state of SB - empty since dependencies are not met
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	Expect(mockSB.GetValues(nil)).To(BeEmpty())
	Expect(mockSB.PopHistoryOfOps()).To(HaveLen(0))

	// check metadata
	Expect(metadataMap.ListAllNames()).To(BeEmpty())

	// subscribe to receive notifications about values which are going to fail
	prefBStatusChan := make(chan *BaseValueStatus, 5)
	scheduler.WatchValueStatus(prefBStatusChan, prefixSelector(prefixB))
	prefCStatusChan := make(chan *BaseValueStatus, 5)
	scheduler.WatchValueStatus(prefCStatusChan, prefixSelector(prefixC))

	// send notification
	startTime := time.Now()
	notifError := scheduler.PushSBNotification(prefixA+baseValue1, test.NewArrayValue("item1", "item2"),
		&test.OnlyInteger{Integer: 10})
	Expect(notifError).ShouldNot(HaveOccurred())

	// wait until the notification is processed
	Eventually(func() []*KVWithMetadata {
		return mockSB.GetValues(nil)
	}, 2*time.Second).Should(HaveLen(2))
	stopTime := time.Now()

	// check value state updates received through the channels
	var valueStatus *BaseValueStatus
	Eventually(prefBStatusChan, time.Second).Should(Receive(&valueStatus))
	checkBaseValueStatus(valueStatus, &BaseValueStatus{
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
				State:         ValueState_RETRYING,
				LastOperation: TxnOperation_CREATE,
				Error:         "failed to add derived value",
			},
		},
	})
	Eventually(prefCStatusChan, time.Second).Should(Receive(&valueStatus))
	checkBaseValueStatus(valueStatus, &BaseValueStatus{
		Value: &ValueStatus{
			Key:           prefixC + baseValue3,
			State:         ValueState_RETRYING,
			LastOperation: TxnOperation_CREATE,
			Error:         "failed to add value",
		},
	})

	// check the state of SB
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	// -> base value 2
	value := mockSB.GetValue(prefixB + baseValue2)
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewArrayValue("item1"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(0))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item1 derived from base value 2
	value = mockSB.GetValue(prefixB + baseValue2 + "/item1")
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("item1"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item2 derived from base value 2 failed to get created
	value = mockSB.GetValue(prefixB + baseValue2 + "/item2")
	Expect(value).To(BeNil())
	// -> base value 3 failed to get created
	value = mockSB.GetValue(prefixC + baseValue3)
	Expect(value).To(BeNil())
	Expect(mockSB.GetValues(nil)).To(HaveLen(2))

	// check metadata
	metadata, exists := nameToInteger1.LookupByName(baseValue1)
	Expect(exists).To(BeTrue())
	Expect(metadata.GetInteger()).To(BeEquivalentTo(10))
	metadata, exists = nameToInteger2.LookupByName(baseValue2)
	Expect(exists).To(BeTrue())
	Expect(metadata.GetInteger()).To(BeEquivalentTo(0))
	metadata, exists = nameToInteger3.LookupByName(baseValue3)
	Expect(exists).To(BeFalse())
	Expect(metadata).To(BeNil())

	// check operations executed in SB
	opHistory := mockSB.PopHistoryOfOps()
	Expect(opHistory).To(HaveLen(6))
	operation := opHistory[0]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixB + baseValue2))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[1]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixB + baseValue2 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[2]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3))
	Expect(operation.Err).ToNot(BeNil())
	Expect(operation.Err.Error()).To(BeEquivalentTo("failed to add value"))
	operation = opHistory[3]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixB + baseValue2 + "/item2"))
	Expect(operation.Err).ToNot(BeNil())
	Expect(operation.Err.Error()).To(BeEquivalentTo("failed to add derived value"))
	operation = opHistory[4] // refresh failed value
	Expect(operation.OpType).To(Equal(test.MockRetrieve))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	checkValues(operation.CorrelateRetrieve, []KVWithMetadata{
		{
			Key:      prefixB + baseValue2,
			Value:    test.NewArrayValue("item1", "item2"),
			Metadata: &test.OnlyInteger{Integer: 0},
			Origin:   FromNB,
		},
	})
	operation = opHistory[5] // refresh failed value
	Expect(operation.OpType).To(Equal(test.MockRetrieve))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	checkValues(operation.CorrelateRetrieve, []KVWithMetadata{})

	// check last transaction
	txnHistory := scheduler.GetTransactionHistory(time.Time{}, time.Now())
	Expect(txnHistory).To(HaveLen(3))
	txn := txnHistory[2]
	Expect(txn.PreRecord).To(BeFalse())
	Expect(txn.Start.After(startTime)).To(BeTrue())
	Expect(txn.Start.Before(txn.Stop)).To(BeTrue())
	Expect(txn.Stop.Before(stopTime)).To(BeTrue())
	Expect(txn.SeqNum).To(BeEquivalentTo(2))
	Expect(txn.TxnType).To(BeEquivalentTo(SBNotification))
	Expect(txn.ResyncType).To(BeEquivalentTo(NotResync))
	Expect(txn.Description).To(BeEmpty())
	checkRecordedValues(txn.Values, []RecordedKVPair{
		{Key: prefixA + baseValue1, Value: utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")), Origin: FromSB},
	})

	// -> planned operations
	txnOps := RecordedTxnOps{
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixA + baseValue1,
			NewValue:  utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")),
			PrevState: ValueState_NONEXISTENT,
			NewState:  ValueState_OBTAINED,
		},
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixB + baseValue2,
			PrevValue: utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")),
			NewValue:  utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")),
			PrevState: ValueState_PENDING,
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
			Key:       prefixC + baseValue3,
			PrevValue: utils.RecordProtoMessage(test.NewStringValue("base-value3-data")),
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("base-value3-data")),
			PrevState: ValueState_PENDING,
			NewState:  ValueState_CONFIGURED,
		},
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixA + baseValue1 + "/item1",
			IsDerived: true,
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("item1")),
			PrevState: ValueState_NONEXISTENT,
			NewState:  ValueState_OBTAINED,
		},
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixA + baseValue1 + "/item2",
			IsDerived: true,
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("item2")),
			PrevState: ValueState_NONEXISTENT,
			NewState:  ValueState_OBTAINED,
		},
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixB + baseValue2 + "/item2",
			IsDerived: true,
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("item2")),
			PrevState: ValueState_NONEXISTENT,
			NewState:  ValueState_CONFIGURED,
		},
	}
	checkTxnOperations(txn.Planned, txnOps)

	// -> executed operations
	txnOps = RecordedTxnOps{
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixA + baseValue1,
			NewValue:  utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")),
			PrevState: ValueState_NONEXISTENT,
			NewState:  ValueState_OBTAINED,
		},
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixB + baseValue2,
			PrevValue: utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")),
			NewValue:  utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")),
			PrevState: ValueState_PENDING,
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
			Key:       prefixC + baseValue3,
			PrevValue: utils.RecordProtoMessage(test.NewStringValue("base-value3-data")),
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("base-value3-data")),
			PrevState: ValueState_PENDING,
			NewState:  ValueState_RETRYING,
			NewErr:    errors.New("failed to add value"),
		},
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixA + baseValue1 + "/item1",
			IsDerived: true,
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("item1")),
			PrevState: ValueState_NONEXISTENT,
			NewState:  ValueState_OBTAINED,
		},
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixA + baseValue1 + "/item2",
			IsDerived: true,
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("item2")),
			PrevState: ValueState_NONEXISTENT,
			NewState:  ValueState_OBTAINED,
		},
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixB + baseValue2 + "/item2",
			IsDerived: true,
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("item2")),
			PrevState: ValueState_NONEXISTENT,
			NewState:  ValueState_RETRYING,
			NewErr:    errors.New("failed to add derived value"),
		},
	}
	checkTxnOperations(txn.Executed, txnOps)

	// check flag stats
	graphR := scheduler.graph.Read()
	errorStats := graphR.GetFlagStats(ErrorFlagIndex, nil)
	Expect(errorStats.TotalCount).To(BeEquivalentTo(2))
	pendingStats := graphR.GetFlagStats(UnavailValueFlagIndex, nil)
	Expect(pendingStats.TotalCount).To(BeEquivalentTo(4))
	derivedStats := graphR.GetFlagStats(DerivedFlagIndex, nil)
	Expect(derivedStats.TotalCount).To(BeEquivalentTo(4))
	lastUpdateStats := graphR.GetFlagStats(LastUpdateFlagIndex, nil)
	Expect(lastUpdateStats.TotalCount).To(BeEquivalentTo(9))
	descriptorStats := graphR.GetFlagStats(DescriptorFlagIndex, nil)
	Expect(descriptorStats.TotalCount).To(BeEquivalentTo(9))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor1Name))
	Expect(descriptorStats.PerValueCount[descriptor1Name]).To(BeEquivalentTo(3))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor2Name))
	Expect(descriptorStats.PerValueCount[descriptor2Name]).To(BeEquivalentTo(4))
	Expect(descriptorStats.PerValueCount).To(HaveKey(descriptor3Name))
	Expect(descriptorStats.PerValueCount[descriptor3Name]).To(BeEquivalentTo(2))
	valueStateStats := graphR.GetFlagStats(ValueStateFlagIndex, nil)
	Expect(valueStateStats.TotalCount).To(BeEquivalentTo(9))
	Expect(valueStateStats.PerValueCount).To(HaveKey(ValueState_OBTAINED.String()))
	Expect(valueStateStats.PerValueCount[ValueState_OBTAINED.String()]).To(BeEquivalentTo(3))
	Expect(valueStateStats.PerValueCount).To(HaveKey(ValueState_CONFIGURED.String()))
	Expect(valueStateStats.PerValueCount[ValueState_CONFIGURED.String()]).To(BeEquivalentTo(2))
	Expect(valueStateStats.PerValueCount).To(HaveKey(ValueState_RETRYING.String()))
	Expect(valueStateStats.PerValueCount[ValueState_RETRYING.String()]).To(BeEquivalentTo(2))
	graphR.Release()

	// item2 derived from baseValue2 should get fixed first
	startTime = time.Now()
	Eventually(prefBStatusChan, 3*time.Second).Should(Receive(&valueStatus))
	// TODO: do we want UPDATEs here? (or just CREATE since nothing has changed)
	checkBaseValueStatus(valueStatus, &BaseValueStatus{
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
				State:         ValueState_CONFIGURED,
				LastOperation: TxnOperation_CREATE,
			},
		},
	})
	stopTime = time.Now()

	// check the state of SB
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	// -> item2 derived from base value 2 is now created
	value = mockSB.GetValue(prefixB + baseValue2 + "/item2")
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("item2"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	Expect(mockSB.GetValues(nil)).To(HaveLen(3))

	// check operations executed in SB
	opHistory = mockSB.PopHistoryOfOps()
	Expect(opHistory).To(HaveLen(2))
	operation = opHistory[0]
	Expect(operation.OpType).To(Equal(test.MockUpdate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixB + baseValue2))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[1]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor2Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixB + baseValue2 + "/item2"))
	Expect(operation.Err).To(BeNil())

	// check last transaction
	txnHistory = scheduler.GetTransactionHistory(startTime, time.Now())
	Expect(txnHistory).To(HaveLen(1))
	txn = txnHistory[0]
	Expect(txn.PreRecord).To(BeFalse())
	Expect(txn.Start.After(startTime)).To(BeTrue())
	Expect(txn.Start.Before(txn.Stop)).To(BeTrue())
	Expect(txn.Stop.Before(stopTime)).To(BeTrue())
	Expect(txn.SeqNum).To(BeEquivalentTo(3))
	Expect(txn.TxnType).To(BeEquivalentTo(RetryFailedOps))
	Expect(txn.ResyncType).To(BeEquivalentTo(NotResync))
	Expect(txn.Description).To(BeEmpty())
	checkRecordedValues(txn.Values, []RecordedKVPair{
		{Key: prefixB + baseValue2, Value: utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")), Origin: FromNB},
	})
	txnOps = RecordedTxnOps{
		{
			Operation: TxnOperation_UPDATE,
			Key:       prefixB + baseValue2,
			PrevValue: utils.RecordProtoMessage(test.NewArrayValue("item1")),
			NewValue:  utils.RecordProtoMessage(test.NewArrayValue("item1", "item2")),
			PrevState: ValueState_CONFIGURED,
			NewState:  ValueState_CONFIGURED,
			IsRetry:   true,
		},
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixB + baseValue2 + "/item2",
			IsDerived: true,
			PrevValue: utils.RecordProtoMessage(test.NewStringValue("item2")),
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("item2")),
			PrevState: ValueState_RETRYING,
			NewState:  ValueState_CONFIGURED,
			PrevErr:   errors.New("failed to add derived value"),
			IsRetry:   true,
		},
	}
	checkTxnOperations(txn.Planned, txnOps)
	checkTxnOperations(txn.Executed, txnOps)

	// base-value3 should get fixed eventually as well
	startTime = time.Now()
	Eventually(prefCStatusChan, 5*time.Second).Should(Receive(&valueStatus))
	checkBaseValueStatus(valueStatus, &BaseValueStatus{
		Value: &ValueStatus{
			Key:           prefixC + baseValue3,
			State:         ValueState_CONFIGURED,
			LastOperation: TxnOperation_CREATE,
		},
	})
	stopTime = time.Now()

	// check the state of SB
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	// -> base value 3 is now created
	value = mockSB.GetValue(prefixC + baseValue3)
	Expect(value).ToNot(BeNil())
	Expect(proto.Equal(value.Value, test.NewStringValue("base-value3-data"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(test.MetaWithInteger).GetInteger()).To(BeEquivalentTo(0))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	Expect(mockSB.GetValues(nil)).To(HaveLen(4))

	// check operations executed in SB
	opHistory = mockSB.PopHistoryOfOps()
	Expect(opHistory).To(HaveLen(1))
	operation = opHistory[0]
	Expect(operation.OpType).To(Equal(test.MockCreate))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor3Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixC + baseValue3))
	Expect(operation.Err).To(BeNil())

	// check last transaction
	txnHistory = scheduler.GetTransactionHistory(startTime, time.Time{})
	Expect(txnHistory).To(HaveLen(1))
	txn = txnHistory[0]
	Expect(txn.PreRecord).To(BeFalse())
	Expect(txn.Start.After(startTime)).To(BeTrue())
	Expect(txn.Start.Before(txn.Stop)).To(BeTrue())
	Expect(txn.Stop.Before(stopTime)).To(BeTrue())
	Expect(txn.SeqNum).To(BeEquivalentTo(4))
	Expect(txn.TxnType).To(BeEquivalentTo(RetryFailedOps))
	Expect(txn.ResyncType).To(BeEquivalentTo(NotResync))
	Expect(txn.Description).To(BeEmpty())
	checkRecordedValues(txn.Values, []RecordedKVPair{
		{Key: prefixC + baseValue3, Value: utils.RecordProtoMessage(test.NewStringValue("base-value3-data")), Origin: FromNB},
	})
	txnOps = RecordedTxnOps{
		{
			Operation: TxnOperation_CREATE,
			Key:       prefixC + baseValue3,
			PrevValue: utils.RecordProtoMessage(test.NewStringValue("base-value3-data")),
			NewValue:  utils.RecordProtoMessage(test.NewStringValue("base-value3-data")),
			PrevState: ValueState_RETRYING,
			NewState:  ValueState_CONFIGURED,
			PrevErr:   errors.New("failed to add value"),
			IsRetry:   true,
		},
	}
	checkTxnOperations(txn.Planned, txnOps)
	checkTxnOperations(txn.Executed, txnOps)

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
