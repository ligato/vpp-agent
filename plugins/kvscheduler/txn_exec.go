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
	"fmt"
	"runtime/trace"
	"sort"
	"strings"

	"github.com/golang/protobuf/proto"

	"go.ligato.io/cn-infra/v2/logging"

	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler/internal/graph"
	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler/internal/utils"
	"go.ligato.io/vpp-agent/v3/proto/ligato/kvscheduler"
)

// applyValueArgs collects all arguments to applyValue method.
type applyValueArgs struct {
	graphW  graph.RWAccess
	txn     *transaction
	kv      kvForTxn
	baseKey string

	applied    utils.KeySet // set of values already(+being) applied
	recreating utils.KeySet // set of values currently being re-created

	isRetry bool
	dryRun  bool

	// set inside of the recursive chain of applyValue-s
	isDepUpdate bool
	isDerived   bool

	// handling of dependency cycles
	depth  int
	branch utils.KeySet
}

// executeTransaction executes pre-processed transaction.
// If <dry-run> is enabled, Validate/Create/Delete/Update operations will not be executed
// and the graph will be returned to its original state at the end.
func (s *Scheduler) executeTransaction(txn *transaction, graphW graph.RWAccess, dryRun bool) (executed kvs.RecordedTxnOps) {
	op := "execute transaction"
	if dryRun {
		op = "simulate transaction"
	}
	defer trace.StartRegion(txn.ctx, op).End()
	if dryRun {
		defer trackTransactionMethod("simulateTransaction")()
	} else {
		defer trackTransactionMethod("executeTransaction")()
	}

	if s.logGraphWalk {
		msg := fmt.Sprintf("%s (seqNum=%d)", op, txn.seqNum)
		fmt.Printf("%s %s\n", nodeVisitBeginMark, msg)
		defer fmt.Printf("%s %s\n", nodeVisitEndMark, msg)
	}

	branch := utils.NewMapBasedKeySet() // branch of current recursive calls to applyValue used to handle cycles
	applied := utils.NewMapBasedKeySet()

	prevValues := make([]kvs.KeyValuePair, 0, len(txn.values))

	// execute transaction either in best-effort mode or with revert on the first failure
	var revert bool
	for _, kv := range txn.values {
		applied.Add(kv.key)
		ops, prevValue, err := s.applyValue(&applyValueArgs{
			graphW:  graphW,
			txn:     txn,
			kv:      kv,
			baseKey: kv.key,
			applied: applied,
			dryRun:  dryRun,
			isRetry: txn.txnType == kvs.RetryFailedOps,
			branch:  branch,
		})
		executed = append(executed, ops...)
		prevValues = append(prevValues, kvs.KeyValuePair{})
		copy(prevValues[1:], prevValues)
		prevValues[0] = prevValue
		if err != nil {
			if txn.txnType == kvs.NBTransaction && txn.nb.revertOnFailure {
				// refresh failed value and trigger reverting
				// (not dry-run)
				failedKey := utils.NewSingletonKeySet(kv.key)
				s.refreshGraph(graphW, failedKey, nil, true)
				revert = true
				break
			}
		}
	}

	if revert {
		// record graph state in-between failure and revert
		graphW.Release()
		graphW = s.graph.Write(!dryRun, true)

		// revert back to previous values
		for _, kvPair := range prevValues {
			ops, _, _ := s.applyValue(&applyValueArgs{
				graphW: graphW,
				txn:    txn,
				kv: kvForTxn{
					key:      kvPair.Key,
					value:    kvPair.Value,
					origin:   kvs.FromNB,
					isRevert: true,
				},
				baseKey: kvPair.Key,
				applied: applied,
				dryRun:  dryRun,
				branch:  branch,
			})
			executed = append(executed, ops...)
		}
	}

	// get rid of uninteresting intermediate pending Create/Delete operations
	executed = s.compressTxnOps(executed)
	return executed
}

// applyValue applies new value received from NB or SB.
// It returns the list of executed operations.
func (s *Scheduler) applyValue(args *applyValueArgs) (executed kvs.RecordedTxnOps, prevValue kvs.KeyValuePair, err error) {
	// dependency cycle detection
	if cycle := args.branch.Has(args.kv.key); cycle {
		return executed, prevValue, err
	}
	args.branch.Add(args.kv.key)
	defer args.branch.Del(args.kv.key)

	// verbose logging
	if s.logGraphWalk {
		endLog := s.logNodeVisit("applyValue", args)
		defer endLog()
	}

	// create new revision of the node for the given key-value pair
	node := args.graphW.SetNode(args.kv.key)

	// remember previous value for a potential revert
	prevValue.Key = node.GetKey()
	prevValue.Value = node.GetValue()

	// remember previous value status to detect and notify about changes
	prevState := getNodeState(node)
	prevOp := getNodeLastOperation(node)
	prevErr := getNodeErrorString(node)
	prevDetails := getValueDetails(node)

	// prepare operation description - fill attributes that we can even before executing the operation
	txnOp := s.preRecordTxnOp(args, node)

	// determine the operation type
	if args.isDepUpdate {
		s.determineDepUpdateOperation(node, txnOp)
		if txnOp.Operation == kvscheduler.TxnOperation_UNDEFINED {
			// nothing needs to be updated
			if node.GetValue() == nil {
				// this value was already deleted (unsatisfied, derived) within
				// the same cycle of runDepUpdates(), and we do not want to leak
				// node with nil value
				args.graphW.DeleteNode(args.kv.key)
			}
			return
		}
	} else if args.kv.value == nil {
		txnOp.Operation = kvscheduler.TxnOperation_DELETE
	} else if node.GetValue() == nil || !isNodeAvailable(node) {
		txnOp.Operation = kvscheduler.TxnOperation_CREATE
	} else {
		txnOp.Operation = kvscheduler.TxnOperation_UPDATE
	}

	// remaining txnOp attributes to fill:
	//		NewState   bool
	//		NewErr     error
	//      NOOP       bool
	//      IsRecreate bool

	// update node flags
	prevUpdate := getNodeLastUpdate(node)
	lastUpdateFlag := &LastUpdateFlag{
		txnSeqNum: args.txn.seqNum,
		txnOp:     txnOp.Operation,
		value:     args.kv.value,
		revert:    args.kv.isRevert,
	}
	if args.txn.txnType == kvs.NBTransaction {
		lastUpdateFlag.retryEnabled = args.txn.nb.retryEnabled
		lastUpdateFlag.retryArgs = args.txn.nb.retryArgs
	} else if prevUpdate != nil {
		// inherit retry arguments from the last NB txn for this value
		lastUpdateFlag.retryEnabled = prevUpdate.retryEnabled
		lastUpdateFlag.retryArgs = prevUpdate.retryArgs
	} else if args.isDerived {
		// inherit from the parent value
		parentNode := args.graphW.GetNode(args.baseKey)
		prevParentUpdate := getNodeLastUpdate(parentNode)
		if prevParentUpdate != nil {
			lastUpdateFlag.retryEnabled = prevParentUpdate.retryEnabled
			lastUpdateFlag.retryArgs = prevParentUpdate.retryArgs
		}

	}
	node.SetFlags(lastUpdateFlag)

	// if the value is already "broken" by this transaction, do not try to update
	// anymore, unless this is a revert
	// (needs to be refreshed first in the post-processing stage)
	if (prevState == kvscheduler.ValueState_FAILED || prevState == kvscheduler.ValueState_RETRYING) &&
		!args.kv.isRevert && prevUpdate != nil && prevUpdate.txnSeqNum == args.txn.seqNum {
		_, prevErr := getNodeError(node)
		return executed, prevValue, prevErr
	}

	// run selected operation
	switch txnOp.Operation {
	case kvscheduler.TxnOperation_DELETE:
		executed, err = s.applyDelete(node, txnOp, args, args.isDepUpdate, false)
	case kvscheduler.TxnOperation_CREATE:
		executed, err = s.applyCreate(node, txnOp, args)
	case kvscheduler.TxnOperation_UPDATE:
		executed, err = s.applyUpdate(node, txnOp, args)
	}

	// detect value state changes
	if !args.dryRun {
		nodeR := args.graphW.GetNode(args.kv.key)
		if prevUpdate == nil || prevState != getNodeState(nodeR) || prevOp != getNodeLastOperation(nodeR) ||
			prevErr != getNodeErrorString(nodeR) || !equalValueDetails(prevDetails, getValueDetails(nodeR)) {
			s.updatedStates.Add(args.baseKey)
		}
	}

	return executed, prevValue, err
}

// applyDelete removes value.
func (s *Scheduler) applyDelete(node graph.NodeRW, txnOp *kvs.RecordedTxnOp, args *applyValueArgs,
	pending, recreate bool) (executed kvs.RecordedTxnOps, err error) {

	if s.logGraphWalk {
		endLog := s.logNodeVisit("applyDelete", args)
		defer endLog()
	}

	if node.GetValue() == nil {
		// remove value that does not exist => noop (do not even record)
		args.graphW.DeleteNode(args.kv.key)
		return executed, nil
	}

	// reflect removal in the graph at the return
	var (
		inheritedErr error
		retriableErr bool
	)
	prevState := getNodeState(node)
	defer func() {
		if inheritedErr != nil {
			// revert back to available, derived value failed instead
			node.DelFlags(UnavailValueFlagIndex)
			s.updateNodeState(node, prevState, args)
			return
		}
		if err == nil {
			node.DelFlags(ErrorFlagIndex)
			if pending {
				// deleted due to missing dependencies
				txnOp.NewState = kvscheduler.ValueState_PENDING
				s.updateNodeState(node, txnOp.NewState, args)
			} else {
				// removed by request
				txnOp.NewState = kvscheduler.ValueState_REMOVED
				if args.isDerived && !recreate {
					args.graphW.DeleteNode(args.kv.key)
				} else {
					s.updateNodeState(node, txnOp.NewState, args)
				}
			}
		} else {
			txnOp.NewErr = err
			txnOp.NewState = s.markFailedValue(node, args, err, retriableErr)
			if !args.applied.Has(getNodeBaseKey(node)) {
				// value removal not originating from this transaction
				err = nil
			}
		}
		executed = append(executed, txnOp)
	}()

	if !isNodeAvailable(node) {
		// removing value that was pending => just update the state in the graph
		txnOp.NOOP = true
		return
	}

	// already mark as unavailable so that other nodes will not view it as satisfied
	// dependency during removal
	node.SetFlags(&UnavailValueFlag{})
	if !pending {
		// state may still change if delete fails
		s.updateNodeState(node, kvscheduler.ValueState_REMOVED, args)
	}

	// remove derived values
	if !args.isDerived {
		var derivedVals []kvForTxn
		for _, derivedNode := range getDerivedNodes(node) {
			derivedVals = append(derivedVals, kvForTxn{
				key:      derivedNode.GetKey(),
				value:    nil, // delete
				origin:   args.kv.origin,
				isRevert: args.kv.isRevert,
			})
		}
		var derExecs kvs.RecordedTxnOps
		derExecs, inheritedErr = s.applyDerived(derivedVals, args, false)
		executed = append(executed, derExecs...)
		if inheritedErr != nil {
			err = inheritedErr
			return
		}
	}

	// update values that depend on this kv-pair
	depExecs, inheritedErr := s.runDepUpdates(node, args, false)
	executed = append(executed, depExecs...)
	if inheritedErr != nil {
		err = inheritedErr
		return
	}

	// execute delete operation
	descriptor := s.registry.GetDescriptorForKey(node.GetKey())
	handler := newDescriptorHandler(descriptor)
	if !args.dryRun && descriptor != nil {
		if args.kv.origin != kvs.FromSB {
			err = handler.delete(node.GetKey(), node.GetValue(), node.GetMetadata())
		}
		if err != nil {
			retriableErr = handler.isRetriableFailure(err)
		} else if canNodeHaveMetadata(node) && descriptor.WithMetadata {
			node.SetMetadata(nil)
		}
	}
	return
}

// applyCreate creates new value which previously didn't exist or was unavailable.
func (s *Scheduler) applyCreate(node graph.NodeRW, txnOp *kvs.RecordedTxnOp, args *applyValueArgs) (executed kvs.RecordedTxnOps, err error) {
	if s.logGraphWalk {
		endLog := s.logNodeVisit("applyCreate", args)
		defer endLog()
	}
	node.SetValue(args.kv.value)

	// get descriptor
	descriptor := s.registry.GetDescriptorForKey(args.kv.key)
	handler := newDescriptorHandler(descriptor)
	if descriptor != nil {
		node.SetFlags(&DescriptorFlag{descriptor.Name})
		node.SetLabel(handler.keyLabel(args.kv.key))
	}

	// handle unimplemented value
	unimplemented := args.kv.origin == kvs.FromNB && !args.isDerived && descriptor == nil
	if unimplemented {
		if getNodeState(node) == kvscheduler.ValueState_UNIMPLEMENTED {
			// already known
			return
		}
		node.SetFlags(&UnavailValueFlag{})
		node.DelFlags(ErrorFlagIndex)
		txnOp.NOOP = true
		txnOp.NewState = kvscheduler.ValueState_UNIMPLEMENTED
		s.updateNodeState(node, txnOp.NewState, args)
		return kvs.RecordedTxnOps{txnOp}, nil
	}

	// mark derived value
	if args.isDerived {
		node.SetFlags(&DerivedFlag{baseKey: args.baseKey})
	}

	// validate value
	if !args.dryRun && args.kv.origin == kvs.FromNB {
		err = handler.validate(node.GetKey(), node.GetValue())
		if err != nil {
			node.SetFlags(&UnavailValueFlag{})
			txnOp.NewErr = err
			txnOp.NewState = kvscheduler.ValueState_INVALID
			txnOp.NOOP = true
			s.updateNodeState(node, txnOp.NewState, args)
			node.SetFlags(&ErrorFlag{err: err, retriable: false})
			if !args.applied.Has(getNodeBaseKey(node)) {
				// invalid value not originating from this transaction
				err = nil
			}
			return kvs.RecordedTxnOps{txnOp}, err
		}
	}

	// apply new relations
	derives, updateExecs, inheritedErr := s.applyNewRelations(node, handler, nil, true, args)
	executed = append(executed, updateExecs...)
	if inheritedErr != nil {
		// error is not expected here, executed operations should be NOOPs
		err = inheritedErr
		return
	}

	if !isNodeReady(node) {
		// if not ready, nothing to do
		node.SetFlags(&UnavailValueFlag{})
		node.DelFlags(ErrorFlagIndex)
		txnOp.NewState = kvscheduler.ValueState_PENDING
		txnOp.NOOP = true
		s.updateNodeState(node, txnOp.NewState, args)
		return kvs.RecordedTxnOps{txnOp}, nil
	}

	// execute Create operation
	if !args.dryRun && descriptor != nil {
		var metadata interface{}

		if args.kv.origin != kvs.FromSB {
			metadata, err = handler.create(node.GetKey(), node.GetValue())
		} else {
			// already created in SB
			metadata = args.kv.metadata
		}

		if err != nil {
			// create failed => assume the value is unavailable
			node.SetFlags(&UnavailValueFlag{})
			retriableErr := handler.isRetriableFailure(err)
			txnOp.NewErr = err
			txnOp.NewState = s.markFailedValue(node, args, err, retriableErr)
			if !args.applied.Has(getNodeBaseKey(node)) {
				// value not originating from this transaction
				err = nil
			}
			return kvs.RecordedTxnOps{txnOp}, err
		}

		// add metadata to the map
		if canNodeHaveMetadata(node) && descriptor.WithMetadata {
			node.SetMetadataMap(descriptor.Name)
			node.SetMetadata(metadata)
		}
	}

	// finalize node and save before going to derived values + dependencies
	node.DelFlags(ErrorFlagIndex, UnavailValueFlagIndex)
	if args.kv.origin == kvs.FromSB {
		txnOp.NewState = kvscheduler.ValueState_OBTAINED
	} else {
		txnOp.NewState = kvscheduler.ValueState_CONFIGURED
	}
	s.updateNodeState(node, txnOp.NewState, args)
	executed = append(executed, txnOp)

	// update values that depend on this kv-pair
	depExecs, inheritedErr := s.runDepUpdates(node, args, true)
	executed = append(executed, depExecs...)
	if inheritedErr != nil {
		err = inheritedErr
		return
	}

	// created derived values
	if !args.isDerived {
		var derivedVals []kvForTxn
		for _, derivedVal := range derives {
			derivedVals = append(derivedVals, kvForTxn{
				key:      derivedVal.Key,
				value:    derivedVal.Value,
				origin:   args.kv.origin,
				isRevert: args.kv.isRevert,
			})
		}
		derExecs, inheritedErr := s.applyDerived(derivedVals, args, true)
		executed = append(executed, derExecs...)
		if inheritedErr != nil {
			err = inheritedErr
		}
	}
	return
}

// applyUpdate applies new value to existing non-pending value.
func (s *Scheduler) applyUpdate(node graph.NodeRW, txnOp *kvs.RecordedTxnOp, args *applyValueArgs) (executed kvs.RecordedTxnOps, err error) {
	if s.logGraphWalk {
		endLog := s.logNodeVisit("applyUpdate", args)
		defer endLog()
	}

	// validate new value
	descriptor := s.registry.GetDescriptorForKey(args.kv.key)
	handler := newDescriptorHandler(descriptor)
	if !args.dryRun && args.kv.origin == kvs.FromNB {
		err = handler.validate(node.GetKey(), args.kv.value)
		if err != nil {
			node.SetValue(args.kv.value) // save the invalid value
			node.SetFlags(&UnavailValueFlag{})
			txnOp.NewErr = err
			txnOp.NewState = kvscheduler.ValueState_INVALID
			txnOp.NOOP = true
			s.updateNodeState(node, txnOp.NewState, args)
			node.SetFlags(&ErrorFlag{err: err, retriable: false})
			if !args.applied.Has(getNodeBaseKey(node)) {
				// invalid value not originating from this transaction
				err = nil
			}
			return kvs.RecordedTxnOps{txnOp}, err
		}
	}

	// compare new value with the old one
	equivalent := handler.equivalentValues(node.GetKey(), node.GetValue(), args.kv.value)

	// re-create the value if required by the descriptor
	recreate := !equivalent &&
		args.kv.origin != kvs.FromSB &&
		handler.updateWithRecreate(args.kv.key, node.GetValue(), args.kv.value, node.GetMetadata())

	if recreate {
		// mark keys which are being re-created for preRecordTxnOp
		args.recreating = getDerivedKeys(node)
		args.recreating.Add(node.GetKey())
		defer func() { args.recreating = nil }()
		// remove the obsolete revision of the value
		delOp := s.preRecordTxnOp(args, node)
		delOp.Operation = kvscheduler.TxnOperation_DELETE
		delOp.NewValue = nil
		delExec, inheritedErr := s.applyDelete(node, delOp, args, false, true)
		executed = append(executed, delExec...)
		if inheritedErr != nil {
			err = inheritedErr
			return
		}
		// create the new revision of the value
		node = args.graphW.SetNode(args.kv.key)
		createOp := s.preRecordTxnOp(args, node)
		createOp.Operation = kvscheduler.TxnOperation_CREATE
		createOp.PrevValue = nil
		createExec, inheritedErr := s.applyCreate(node, createOp, args)
		executed = append(executed, createExec...)
		err = inheritedErr
		return
	}

	// save the new value
	prevValue := node.GetValue()
	node.SetValue(args.kv.value)

	// apply new relations
	derives, updateExecs, inheritedErr := s.applyNewRelations(node, handler, prevValue, !equivalent, args)
	executed = append(executed, updateExecs...)
	if inheritedErr != nil {
		node.SetValue(prevValue) // revert back the original value
		err = inheritedErr
		return
	}

	// if the new dependencies are not satisfied => delete and set as pending with the new value
	if !equivalent && !isNodeReady(node) {
		node.SetValue(prevValue) // apply delete on the original value
		delExec, inheritedErr := s.applyDelete(node, txnOp, args, true, false)
		executed = append(executed, delExec...)
		if inheritedErr != nil {
			err = inheritedErr
		}
		node.SetValue(args.kv.value)
		return
	}

	// execute update operation
	if !args.dryRun && !equivalent && descriptor != nil {
		var newMetadata interface{}

		// call Update handler
		if args.kv.origin != kvs.FromSB {
			newMetadata, err = handler.update(node.GetKey(), prevValue, node.GetValue(), node.GetMetadata())
		} else {
			// already modified in SB
			newMetadata = args.kv.metadata
		}

		if err != nil {
			retriableErr := handler.isRetriableFailure(err)
			txnOp.NewErr = err
			txnOp.NewState = s.markFailedValue(node, args, err, retriableErr)
			executed = append(executed, txnOp)
			if !args.applied.Has(getNodeBaseKey(node)) {
				// update not originating from this transaction
				err = nil
			}
			return
		}

		// update metadata
		if canNodeHaveMetadata(node) && descriptor.WithMetadata {
			node.SetMetadata(newMetadata)
		}
	}

	// finalize node and save before going to new/modified derived values + dependencies
	node.DelFlags(ErrorFlagIndex, UnavailValueFlagIndex)
	if args.kv.origin == kvs.FromSB {
		txnOp.NewState = kvscheduler.ValueState_OBTAINED
	} else {
		txnOp.NewState = kvscheduler.ValueState_CONFIGURED
	}
	s.updateNodeState(node, txnOp.NewState, args)

	// if the value was modified or the state changed, record operation
	if !equivalent || txnOp.PrevState != txnOp.NewState {
		// do not record transition if it only confirms that the value is in sync
		confirmsInSync := equivalent &&
			txnOp.PrevState == kvscheduler.ValueState_DISCOVERED &&
			txnOp.NewState == kvscheduler.ValueState_CONFIGURED
		if !confirmsInSync {
			txnOp.NOOP = equivalent
			executed = append(executed, txnOp)
		}
	}

	if !args.isDerived {
		// update/create derived values
		var derivedVals []kvForTxn
		for _, derivedVal := range derives {
			derivedVals = append(derivedVals, kvForTxn{
				key:      derivedVal.Key,
				value:    derivedVal.Value,
				origin:   args.kv.origin,
				isRevert: args.kv.isRevert,
			})
		}
		derExecs, inheritedErr := s.applyDerived(derivedVals, args, true)
		executed = append(executed, derExecs...)
		if inheritedErr != nil {
			err = inheritedErr
		}
	}
	return
}

// applyNewRelations updates relation definitions and removes obsolete derived
// values.
func (s *Scheduler) applyNewRelations(node graph.NodeRW, handler *descriptorHandler,
	prevValue proto.Message, updateDeps bool,
	args *applyValueArgs) (derivedVals []kvs.KeyValuePair, executed kvs.RecordedTxnOps, err error) {

	if args.isDerived && !updateDeps {
		// nothing to update
		return
	}

	// get the set of derived keys before update
	prevDerivedKeys := utils.NewSliceBasedKeySet()
	if !args.isDerived && prevValue != nil {
		for _, kv := range handler.derivedValues(node.GetKey(), prevValue) {
			prevDerivedKeys.Add(kv.Key)
		}
	}

	// get the set of derived keys after update
	newDerivedKeys := utils.NewSliceBasedKeySet()
	if !args.isDerived {
		derivedVals = handler.derivedValues(node.GetKey(), node.GetValue())
		for _, kv := range derivedVals {
			newDerivedKeys.Add(kv.Key)
		}
	}
	updateDerived := !prevDerivedKeys.Equals(newDerivedKeys)
	if updateDeps || updateDerived {
		dependencies := handler.dependencies(node.GetKey(), node.GetValue())
		node.SetTargets(constructTargets(dependencies, derivedVals))
	}

	// remove obsolete derived values
	if updateDerived {
		var obsoleteDerVals []kvForTxn
		prevDerivedKeys.Subtract(newDerivedKeys)
		for _, obsolete := range prevDerivedKeys.Iterate() {
			obsoleteDerVals = append(obsoleteDerVals, kvForTxn{
				key:      obsolete,
				value:    nil, // delete
				origin:   args.kv.origin,
				isRevert: args.kv.isRevert,
			})
		}
		if len(obsoleteDerVals) > 0 {
			executed, err = s.applyDerived(obsoleteDerVals, args, false)
		}
	}
	return
}

// applyDerived (re-)applies the given list of derived values.
func (s *Scheduler) applyDerived(derivedVals []kvForTxn, args *applyValueArgs, check bool) (executed kvs.RecordedTxnOps, err error) {
	var wasErr error
	if s.logGraphWalk {
		endLog := s.logNodeVisit("applyDerived", args)
		defer endLog()
	}

	// order derivedVals by key (just for deterministic behaviour which simplifies testing)
	sort.Slice(derivedVals, func(i, j int) bool { return derivedVals[i].key < derivedVals[j].key })

	for _, derived := range derivedVals {
		if check && !s.validDerivedKV(args.graphW, derived, args.txn.seqNum) {
			continue
		}
		derArgs := *args
		derArgs.kv = derived
		derArgs.isDerived = true
		derArgs.isDepUpdate = false
		ops, _, err := s.applyValue(&derArgs)
		if err != nil {
			wasErr = err
		}
		executed = append(executed, ops...)
	}
	return executed, wasErr
}

// runDepUpdates triggers dependency updates on all nodes that depend on the given node.
func (s *Scheduler) runDepUpdates(node graph.Node, args *applyValueArgs, forUnavailable bool) (executed kvs.RecordedTxnOps, err error) {
	if s.logGraphWalk {
		endLog := s.logNodeVisit("runDepUpdates", args)
		defer endLog()
	}

	var wasErr error
	var depNodes []graph.Node
	for _, depPerLabel := range node.GetSources(DependencyRelation) {
		depNodes = append(depNodes, depPerLabel.Nodes...)
	}

	// order depNodes by key (just for deterministic behaviour which simplifies testing)
	sort.Slice(depNodes, func(i, j int) bool { return depNodes[i].GetKey() < depNodes[j].GetKey() })

	for _, depNode := range depNodes {
		if getNodeOrigin(depNode) != kvs.FromNB {
			continue
		}
		if !isNodeAvailable(depNode) != forUnavailable {
			continue
		}
		var value proto.Message
		if lastUpdate := getNodeLastUpdate(depNode); lastUpdate != nil {
			value = lastUpdate.value
		} else {
			// state=DISCOVERED
			value = depNode.GetValue()
		}
		depArgs := *args
		depArgs.kv = kvForTxn{
			key:      depNode.GetKey(),
			value:    value,
			origin:   getNodeOrigin(depNode),
			isRevert: args.kv.isRevert,
		}
		depArgs.baseKey = getNodeBaseKey(depNode)
		depArgs.isDerived = isNodeDerived(depNode)
		depArgs.isDepUpdate = true
		ops, _, err := s.applyValue(&depArgs)
		if err != nil {
			wasErr = err
		}
		executed = append(executed, ops...)
	}
	return executed, wasErr
}

// determineDepUpdateOperation determines if the value needs update wrt. dependencies
// and what operation to execute.
func (s *Scheduler) determineDepUpdateOperation(node graph.NodeRW, txnOp *kvs.RecordedTxnOp) {
	// create node if dependencies are now all met
	if !isNodeAvailable(node) {
		if !isNodeReady(node) {
			// nothing to do
			return
		}
		txnOp.Operation = kvscheduler.TxnOperation_CREATE
	} else if !isNodeReady(node) {
		// node should not be available anymore
		txnOp.Operation = kvscheduler.TxnOperation_DELETE
	}
}

// compressTxnOps removes uninteresting intermediate pending Create/Delete operations.
func (s *Scheduler) compressTxnOps(executed kvs.RecordedTxnOps) kvs.RecordedTxnOps {
	// compress Create operations
	compressed := make(kvs.RecordedTxnOps, 0, len(executed))
	for i, op := range executed {
		compressedOp := false
		if op.Operation == kvscheduler.TxnOperation_CREATE && op.NewState == kvscheduler.ValueState_PENDING {
			for j := i + 1; j < len(executed); j++ {
				if executed[j].Key == op.Key {
					if executed[j].Operation == kvscheduler.TxnOperation_CREATE {
						// compress
						compressedOp = true
						executed[j].PrevValue = op.PrevValue
						executed[j].PrevErr = op.PrevErr
						executed[j].PrevState = op.PrevState
					}
					break
				}
			}
		}
		if !compressedOp {
			compressed = append(compressed, op)
		}
	}

	// compress Delete operations
	length := len(compressed)
	for i := length - 1; i >= 0; i-- {
		op := compressed[i]
		compressedOp := false
		if op.Operation == kvscheduler.TxnOperation_DELETE && op.PrevState == kvscheduler.ValueState_PENDING {
			for j := i - 1; j >= 0; j-- {
				if compressed[j].Key == op.Key {
					if compressed[j].Operation == kvscheduler.TxnOperation_DELETE {
						// compress
						compressedOp = true
						compressed[j].NewValue = op.NewValue
						compressed[j].NewErr = op.NewErr
						compressed[j].NewState = op.NewState
					}
					break
				}
			}
		}
		if compressedOp {
			copy(compressed[i:], compressed[i+1:])
			length--
		}
	}
	compressed = compressed[:length]
	return compressed
}

// updateNodeState updates node state if it is really necessary.
func (s *Scheduler) updateNodeState(node graph.NodeRW, newState kvscheduler.ValueState, args *applyValueArgs) {
	if getNodeState(node) != newState {
		if s.logGraphWalk {
			indent := strings.Repeat(" ", (args.depth+1)*2)
			fmt.Printf("%s-> change value state from %v to %v\n", indent, getNodeState(node), newState)
		}
		node.SetFlags(&ValueStateFlag{valueState: newState})
	}
}

func (s *Scheduler) markFailedValue(node graph.NodeRW, args *applyValueArgs, err error,
	retriableErr bool) (newState kvscheduler.ValueState) {

	// decide value state between FAILED and RETRYING
	newState = kvscheduler.ValueState_FAILED
	toBeReverted := args.txn.txnType == kvs.NBTransaction && args.txn.nb.revertOnFailure && !args.kv.isRevert
	if retriableErr && !toBeReverted {
		// consider operation retry
		var alreadyRetried bool
		if args.txn.txnType == kvs.RetryFailedOps {
			baseKey := getNodeBaseKey(node)
			_, alreadyRetried = args.txn.retry.keys[baseKey]
		}
		attempt := 1
		if alreadyRetried {
			attempt = args.txn.retry.attempt + 1
		}
		lastUpdate := getNodeLastUpdate(node)
		if lastUpdate.retryEnabled && lastUpdate.retryArgs != nil &&
			(lastUpdate.retryArgs.MaxCount == 0 || attempt <= lastUpdate.retryArgs.MaxCount) {
			// retry is allowed
			newState = kvscheduler.ValueState_RETRYING
		}
	}
	s.updateNodeState(node, newState, args)
	node.SetFlags(&ErrorFlag{err: err, retriable: retriableErr})
	return newState
}

func (s *Scheduler) logNodeVisit(operation string, args *applyValueArgs) func() {
	msg := fmt.Sprintf("%s (key = %s)", operation, args.kv.key)
	args.depth++
	indent := strings.Repeat(" ", args.depth*2)
	fmt.Printf("%s%s %s\n", indent, nodeVisitBeginMark, msg)
	return func() {
		args.depth--
		fmt.Printf("%s%s %s\n", indent, nodeVisitEndMark, msg)
	}
}

// validDerivedKV check validity of a derived KV pair.
func (s *Scheduler) validDerivedKV(graphR graph.ReadAccess, kv kvForTxn, txnSeqNum uint64) bool {
	node := graphR.GetNode(kv.key)
	if kv.value == nil {
		s.Log.WithFields(logging.Fields{
			"txnSeqNum": txnSeqNum,
			"key":       kv.key,
		}).Warn("Derived nil value")
		return false
	}
	if node != nil {
		if !isNodeDerived(node) {
			s.Log.WithFields(logging.Fields{
				"txnSeqNum": txnSeqNum,
				"value":     kv.value,
				"key":       kv.key,
			}).Warn("Skipping derived value colliding with a base value")
			return false
		}
	}
	return true
}
