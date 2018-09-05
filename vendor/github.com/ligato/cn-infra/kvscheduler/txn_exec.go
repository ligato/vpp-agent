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
	"sort"

	. "github.com/ligato/cn-infra/kvscheduler/api"
	"github.com/ligato/cn-infra/kvscheduler/graph"
	"github.com/ligato/cn-infra/logging"
)

// applyValueArgs collects all arguments to applyValue method.
type applyValueArgs struct {
	graphW graph.RWAccess
	txn    *preProcessedTxn
	kv     kvForTxn

	isRetry bool
	dryRun  bool

	// set inside of the recursive chain of applyValue-s
	isUpdate  bool
	isDerived bool

	// failed base values for potential retry
	failed keySet

	// handling of dependency cycles
	branch keySet
}

// executeTransaction executes pre-processed transaction.
// If <dry-run> is enabled, Add/Delete/Update/Modify operations will not be executed
// and the graph will be returned to its original state at the end.
func (scheduler *Scheduler) executeTransaction(txn *preProcessedTxn, dryRun bool) (executed recordedTxnOps, failed keySet) {
	graphW := scheduler.graph.Write(true)
	failed = make(keySet)  // non-derived values in a failed state
	branch := make(keySet) // branch of current recursive calls to applyValue used to handle cycles

	// order to achieve the shortest sequence of operations in average
	orderedVals := scheduler.orderValuesByOp(graphW, txn.values)

	var (
		prevValues []KeyValuePair
		revert     bool
	)
	// execute transaction either in best-effort mode or with revert on the first failure
	for _, kv := range orderedVals {
		ops, prevValue, err := scheduler.applyValue(
			&applyValueArgs{
				graphW:  graphW,
				txn:     txn,
				kv:      kv,
				dryRun:  dryRun,
				isRetry: txn.args.txnType == retryFailedOps,
				failed:  failed,
				branch:  branch,
			})
		executed = append(executed, ops...)
		prevValues = append([]KeyValuePair{prevValue}, prevValues...)
		if err != nil {
			if txn.args.txnType == nbTransaction && txn.args.nb.revertOnFailure {
				// refresh failed value and trigger reverting
				delete(failed, kv.key) // do not retry unless reverting fails
				scheduler.refreshGraph(graphW, keySet{}.add(kv.key), nil)
				revert = true
				break
			}
		}
	}

	if revert {
		// record graph state in-between failure and revert
		graphW.Release()
		graphW = scheduler.graph.Write(true)

		// revert back to previous values
		for _, kvPair := range prevValues {
			ops, _, _ := scheduler.applyValue(
				&applyValueArgs{
					graphW: graphW,
					txn:    txn,
					kv: kvForTxn{
						key:      kvPair.Key,
						value:    kvPair.Value,
						origin:   FromNB,
						isRevert: true,
					},
					dryRun: dryRun,
					failed: failed,
					branch: branch,
				})
			executed = append(executed, ops...)
		}
	}

	graphW.Release()
	return executed, failed
}

// applyValue applies new value received from NB or SB.
// It returns the list of executed operations.
func (scheduler *Scheduler) applyValue(args *applyValueArgs) (executed recordedTxnOps, prevValue KeyValuePair, err error) {
	// dependency cycle detection
	if _, cycle := args.branch[args.kv.key]; cycle {
		return executed, prevValue, err
	}
	args.branch.add(args.kv.key)
	defer delete(args.branch, args.kv.key)

	// create new revision of the node for the given key-value pair
	node := args.graphW.SetNode(args.kv.key)

	// remember previous value for a potential revert
	prevValue = KeyValuePair{Key: node.GetKey(), Value: node.GetValue()}

	// update node flags
	prevUpdate := getNodeLastUpdate(node)
	prevErr := getNodeError(node)
	if !args.isUpdate {
		// with update it is not certain if any update is actually needed,
		// so let applyUpdate() to refresh LastUpdateFlag
		node.SetFlags(&LastUpdateFlag{args.txn.seqNum})
	}
	if !args.isUpdate {
		if !args.isDerived {
			lastChangeFlag := &LastChangeFlag{
				txnSeqNum: args.txn.seqNum,
				value:     args.kv.value,
				origin:    args.kv.origin,
				revert:    args.kv.isRevert,
			}
			switch args.txn.args.txnType {
			case nbTransaction:
				lastChangeFlag.retryEnabled = args.txn.args.nb.retryFailed
				lastChangeFlag.retryPeriod = args.txn.args.nb.retryPeriod
				lastChangeFlag.retryExpBackoff = args.txn.args.nb.expBackoffRetry
			case retryFailedOps:
				prevLastChange := getNodeLastChange(node)
				lastChangeFlag.retryEnabled = prevLastChange.retryEnabled
				lastChangeFlag.retryPeriod = prevLastChange.retryPeriod
				lastChangeFlag.retryExpBackoff = prevLastChange.retryExpBackoff
			}
			node.SetFlags(lastChangeFlag)
		} else {
			node.SetFlags(&DerivedFlag{})
		}
		node.SetFlags(&OriginFlag{args.kv.origin})
	}

	// if the value is already "broken" by this transaction, do not try to update
	// anymore, unless this is a revert (needs to be refreshed first in the post-processing
	// stage)
	if !args.kv.isRevert && prevErr != nil &&
		prevUpdate != nil && prevUpdate.txnSeqNum == args.txn.seqNum {
		return executed, prevValue, prevErr
	}

	// prepare operation description - fill attributes that we can even before executing the operation
	txnOp := scheduler.preRecordTxnOp(args, node)

	// determine the operation type
	if args.isUpdate {
		txnOp.operation = update // triggered from within recursive applyValue-s
	} else if args.kv.value == nil {
		txnOp.operation = del
	} else if node.GetValue() == nil || isNodePending(node) {
		txnOp.operation = add
	} else {
		txnOp.operation = modify
	}

	// remaining txnOp attributes to fill:
	//		isPending  bool
	//		newErr     error

	switch txnOp.operation {
	case del:
		executed, err = scheduler.applyDelete(node, txnOp, args, false)
	case add:
		executed, err = scheduler.applyAdd(node, txnOp, args)
	case modify:
		executed, err = scheduler.applyModify(node, txnOp, args)
	case update:
		executed, err = scheduler.applyUpdate(node, txnOp, args)
	}

	return executed, prevValue, err
}

// applyDelete either deletes value or moves it to the pending state.
func (scheduler *Scheduler) applyDelete(node graph.NodeRW, txnOp *recordedTxnOp, args *applyValueArgs, pending bool) (executed recordedTxnOps, err error) {
	if !args.dryRun {
		defer args.graphW.Save()
	}

	if node.GetValue() == nil {
		// remove value that does not exist => noop
		args.graphW.DeleteNode(args.kv.key)
		return executed, nil
	}
	if isNodePending(node) {
		// removing value that was pending => just remove from the in-memory graph
		args.graphW.DeleteNode(args.kv.key)
		return recordedTxnOps{txnOp}, nil
	}

	if pending {
		// already mark as pending so that other nodes will not view it as satisfied
		// dependency during removal
		node.SetFlags(&PendingFlag{})
	}

	// remove derived values
	var derivedVals []kvForTxn
	for _, derivedNode := range getDerivedNodes(node) {
		derivedVals = append(derivedVals, kvForTxn{
			key:      derivedNode.GetKey(),
			value:    nil, // delete
			origin:   args.kv.origin,
			isRevert: args.kv.isRevert,
		})
	}
	derExecs, wasErr := scheduler.applyDerived(derivedVals, args, false)
	executed = append(executed, derExecs...)

	// continue even if removal of a derived value has failed ...

	// update values that depend on this kv-pair
	executed = append(executed, scheduler.runUpdates(node, args)...)

	// execute delete operation
	descriptor := scheduler.registry.GetDescriptorForKey(node.GetKey())
	if !args.dryRun && descriptor != nil {
		if args.kv.origin != FromSB {
			err = descriptor.Delete(node.GetKey(), node.GetValue(), node.GetMetadata())
		}
		if err != nil {
			wasErr = err
			node.SetFlags(&ErrorFlag{err})
		}
		if withMeta, _ := descriptor.WithMetadata(); canNodeHaveMetadata(node) && withMeta {
			node.SetMetadata(nil)
		}
	}

	// remove node or mark as failed
	if wasErr != nil {
		if !isNodeDerived(node) {
			args.failed.add(node.GetKey())
		}
	} else if !pending {
		args.graphW.DeleteNode(args.kv.key)
	}

	txnOp.newErr = err
	txnOp.isPending = pending
	executed = append(executed, txnOp)
	return executed, wasErr
}

// applyAdd adds new value which previously didn't exist or was pending.
func (scheduler *Scheduler) applyAdd(node graph.NodeRW, txnOp *recordedTxnOp, args *applyValueArgs) (executed recordedTxnOps, err error) {
	if !args.dryRun {
		defer args.graphW.Save()
	}
	node.SetValue(args.kv.value)

	// get descriptor
	descriptor := scheduler.registry.GetDescriptorForKey(args.kv.key)
	if descriptor != nil {
		node.SetFlags(&DescriptorFlag{descriptor.GetName()})
	}

	// build relations with other targets
	var (
		derives      []KeyValuePair
		dependencies []Dependency
	)
	if descriptor != nil {
		derives = descriptor.DerivedValues(node.GetKey(), node.GetValue())
		dependencies = descriptor.Dependencies(node.GetKey(), node.GetValue())
	}
	node.SetTargets(constructTargets(dependencies, derives))

	if !isNodeReady(node) {
		// if not ready, nothing to do
		node.SetFlags(&PendingFlag{})
		node.DelFlags(ErrorFlagName)
		txnOp.isPending = true
		return recordedTxnOps{txnOp}, nil
	}

	// execute add operation
	if !args.dryRun && descriptor != nil {
		var (
			err      error
			metadata interface{}
		)

		if args.kv.origin != FromSB {
			metadata, err = descriptor.Add(node.GetKey(), node.GetValue())
		} else {
			// already added in SB
			metadata = args.kv.metadata
		}

		if err != nil {
			// add failed => keep value pending
			node.SetFlags(&PendingFlag{})
			node.SetFlags(&ErrorFlag{err})
			if !isNodeDerived(node) {
				args.failed.add(node.GetKey())
			}
			txnOp.isPending = true
			txnOp.newErr = err
			return recordedTxnOps{txnOp}, err
		}

		// add metadata to the map
		if withMeta, _ := descriptor.WithMetadata(); canNodeHaveMetadata(node) && withMeta {
			node.SetMetadataMap(descriptor.GetName())
			node.SetMetadata(metadata)
		}
	}

	// finalize node and save before going to derived values + dependencies
	node.DelFlags(ErrorFlagName, PendingFlagName)
	executed = append(executed, txnOp)
	if !args.dryRun {
		args.graphW.Save()
	}

	// update values that depend on this kv-pair
	executed = append(executed, scheduler.runUpdates(node, args)...)

	// created derived values
	var derivedVals []kvForTxn
	for _, derivedVal := range derives {
		derivedVals = append(derivedVals, kvForTxn{
			key:      derivedVal.Key,
			value:    derivedVal.Value,
			origin:   args.kv.origin,
			isRevert: args.kv.isRevert,
		})
	}
	derExecs, err := scheduler.applyDerived(derivedVals, args, true)
	executed = append(executed, derExecs...)

	if err != nil && !isNodeDerived(node) {
		args.failed.add(node.GetKey())
	}

	return executed, err
}

// applyModify applies new value to existing non-pending value.
func (scheduler *Scheduler) applyModify(node graph.NodeRW, txnOp *recordedTxnOp, args *applyValueArgs) (executed recordedTxnOps, err error) {
	if !args.dryRun {
		defer args.graphW.Save()
	}

	equivalent := node.GetValue().Equivalent(args.kv.value)

	// re-create the value if required by the descriptor
	var recreate bool
	descriptor := scheduler.registry.GetDescriptorForKey(args.kv.key)
	if descriptor != nil && args.kv.origin != FromSB {
		recreate = descriptor.ModifyHasToRecreate(args.kv.key, node.GetValue(), args.kv.value, node.GetMetadata())
	}
	if !equivalent && recreate {
		// record operation as two - delete followed by add
		delOp := scheduler.preRecordTxnOp(args, node)
		delOp.operation = del
		delOp.newValue = nil
		addOp := scheduler.preRecordTxnOp(args, node)
		addOp.operation = add
		addOp.prevValue = nil
		addOp.wasPending = true
		// remove obsolete value
		delExec, err := scheduler.applyDelete(node, delOp, args, true)
		executed = append(executed, delExec...)
		if err != nil {
			return executed, err
		}
		// add the new revision of the value
		addExec, err := scheduler.applyAdd(node, addOp, args)
		executed = append(executed, addExec...)
		return executed, err
	}

	// save the new value
	prevValue := node.GetValue()
	node.SetValue(args.kv.value)

	// get the set of derived keys before modification
	prevDerived := getDerivedKeys(node)

	// set new targets
	var (
		derives      []KeyValuePair
		dependencies []Dependency
	)
	if descriptor != nil {
		derives = descriptor.DerivedValues(node.GetKey(), node.GetValue())
		dependencies = descriptor.Dependencies(node.GetKey(), node.GetValue())
	}
	node.SetTargets(constructTargets(dependencies, derives))

	// remove obsolete derived values
	var obsoleteDerVals []kvForTxn
	for obsolete := range prevDerived.subtract(getDerivedKeys(node)) {
		obsoleteDerVals = append(obsoleteDerVals, kvForTxn{
			key:      obsolete,
			value:    nil, // delete
			origin:   args.kv.origin,
			isRevert: args.kv.isRevert,
		})
	}
	derExecs, wasErr := scheduler.applyDerived(obsoleteDerVals, args, false)
	executed = append(executed, derExecs...)
	if wasErr != nil {
		node.SetFlags(&ErrorFlag{err: wasErr})
		txnOp.newErr = wasErr
		if !isNodeDerived(node) {
			args.failed.add(node.GetKey())
		}
	}

	// if the new dependencies are not satisfied => delete and set as pending with the new value
	if !isNodeReady(node) {
		delExec, err := scheduler.applyDelete(node, txnOp, args, true)
		executed = append(executed, delExec...)
		if err != nil {
			wasErr = err
		}
		return executed, wasErr
	}

	// execute modify operation
	if !args.dryRun && !equivalent && descriptor != nil {
		var (
			err         error
			newMetadata interface{}
		)

		if args.kv.origin != FromSB {
			newMetadata, err = descriptor.Modify(node.GetKey(), prevValue, node.GetValue(), node.GetMetadata())
		} else {
			// already modified in SB
			newMetadata = args.kv.metadata
		}

		if err != nil {
			node.SetFlags(&ErrorFlag{err})
			if !isNodeDerived(node) {
				args.failed.add(node.GetKey())
			}
			txnOp.newErr = err
			executed = append(executed, txnOp)
			return executed, err
		}

		// update metadata
		if withMeta, _ := descriptor.WithMetadata(); canNodeHaveMetadata(node) && withMeta {
			node.SetMetadata(newMetadata)
		}
	}

	// if new value is equivalent, but the value is in failed state from previous txn => run update
	if equivalent && wasErr == nil && getNodeError(node) != nil && descriptor != nil {
		txnOp.operation = update
		err := descriptor.Update(node.GetKey(), node.GetValue(), node.GetMetadata())
		if err != nil {
			node.SetFlags(&ErrorFlag{err})
			if !isNodeDerived(node) {
				args.failed.add(node.GetKey())
			}
			txnOp.newErr = err
			executed = append(executed, txnOp)
			return executed, err
		}
	}

	// finalize node and save before going to new/modified derived values + dependencies
	if wasErr == nil {
		node.DelFlags(ErrorFlagName)
	}
	if !equivalent || txnOp.operation == update || txnOp.newErr != nil {
		// if the value was modified, or update was executed (to clear error)
		// or removal of obsolete derived value has failed => record operation
		executed = append(executed, txnOp)
	}
	if !args.dryRun {
		args.graphW.Save()
	}

	// update values that depend on this kv-pair
	if !equivalent {
		executed = append(executed, scheduler.runUpdates(node, args)...)
	}

	// modify/add derived values
	var derivedVals []kvForTxn
	for _, derivedVal := range derives {
		derivedVals = append(derivedVals, kvForTxn{
			key:      derivedVal.Key,
			value:    derivedVal.Value,
			origin:   args.kv.origin,
			isRevert: args.kv.isRevert,
		})
	}
	derExecs, err = scheduler.applyDerived(derivedVals, args, true)
	executed = append(executed, derExecs...)
	if err != nil {
		wasErr = err
	}

	if wasErr != nil && !isNodeDerived(node) {
		args.failed.add(node.GetKey())
	}

	return executed, wasErr
}

// applyUpdate updates given value since dependencies have changed.
func (scheduler *Scheduler) applyUpdate(node graph.NodeRW, txnOp *recordedTxnOp, args *applyValueArgs) (executed recordedTxnOps, err error) {
	descriptor := scheduler.registry.GetDescriptorForKey(args.kv.key)

	// add node if dependencies are now all met
	if isNodePending(node) {
		if !isNodeReady(node) {
			// nothing to do - do NOT refresh LastUpdateFlag
			return executed, nil
		}
		node.SetFlags(&LastUpdateFlag{args.txn.seqNum})
		addOp := scheduler.preRecordTxnOp(args, node)
		addOp.operation = add
		executed, err = scheduler.applyAdd(node, addOp, args)
	} else {
		node.SetFlags(&LastUpdateFlag{args.txn.seqNum})
		// node is not pending
		if !isNodeReady(node) {
			// delete value and flag node as pending if some dependency is no longer satisfied
			delOp := scheduler.preRecordTxnOp(args, node)
			delOp.operation = del
			delOp.newValue = nil
			executed, err = scheduler.applyDelete(node, delOp, args, true)
		} else {
			// execute Update operation
			if !args.dryRun {
				err = descriptor.Update(node.GetKey(), node.GetValue(), node.GetMetadata())
				txnOp.newErr = err
			}
			executed = append(executed, txnOp)
			if err != nil {
				node.SetFlags(&ErrorFlag{err})
			}
		}
	}

	if err != nil {
		args.failed.add(getNodeBase(node).GetKey())
	}
	return executed, err
}

// applyDerived (re-)applies the given list of derived values.
func (scheduler *Scheduler) applyDerived(derivedVals []kvForTxn, args *applyValueArgs, check bool) (executed recordedTxnOps, err error) {
	var wasErr error

	// order derivedVals by key (just for deterministic behaviour which simplifies testing)
	sort.Slice(derivedVals, func(i, j int) bool { return derivedVals[i].key < derivedVals[j].key })

	for _, derived := range derivedVals {
		if check && !scheduler.validDerivedKV(args.graphW, derived, args.txn.seqNum) {
			continue
		}
		ops, _, err := scheduler.applyValue(
			&applyValueArgs{
				graphW:    args.graphW,
				txn:       args.txn,
				kv:        derived,
				isRetry:   args.isRetry,
				dryRun:    args.dryRun,
				isDerived: true, // <- is derived
				failed:    args.failed,
				branch:    args.branch,
			})
		if err != nil {
			wasErr = err
		}
		executed = append(executed, ops...)
	}
	return executed, wasErr
}

// runUpdates triggers updates on all nodes that depend on the given node.
func (scheduler *Scheduler) runUpdates(node graph.Node, args *applyValueArgs) (executed recordedTxnOps) {
	depNodes := node.GetSources(DependencyRelation)

	// order depNodes by key (just for deterministic behaviour which simplifies testing)
	sort.Slice(depNodes, func(i, j int) bool { return depNodes[i].GetKey() < depNodes[j].GetKey() })

	for _, depNode := range depNodes {
		if getNodeOrigin(depNode) != FromNB {
			continue
		}
		ops, _, _ := scheduler.applyValue(
			&applyValueArgs{
				graphW: args.graphW,
				txn:    args.txn,
				kv: kvForTxn{
					key:      depNode.GetKey(),
					value:    depNode.GetValue(),
					origin:   getNodeOrigin(depNode),
					isRevert: args.kv.isRevert,
				},
				isRetry:  args.isRetry,
				dryRun:   args.dryRun,
				isUpdate: true, // <- update
				failed:   args.failed,
				branch:   args.branch,
			})
		executed = append(executed, ops...)
	}
	return executed
}

// validDerivedKV check validity of a derived KV pair.
func (scheduler *Scheduler) validDerivedKV(graphR graph.ReadAccess, kv kvForTxn, txnSeqNum uint) bool {
	node := graphR.GetNode(kv.key)
	if kv.value == nil {
		scheduler.Log.WithFields(logging.Fields{
			"txnSeqNum": txnSeqNum,
			"key":       kv.key,
		}).Warn("Derived nil value")
		return false
	}
	if node != nil {
		if !isNodeDerived(node) {
			scheduler.Log.WithFields(logging.Fields{
				"txnSeqNum": txnSeqNum,
				"value":     kv.value,
				"key":       kv.key,
			}).Warn("Skipping derived value colliding with a base value")
			return false
		}
	}
	return true
}
