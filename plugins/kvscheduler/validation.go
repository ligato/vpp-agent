// Copyright (c) 2020 Pantheon.tech
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
	"github.com/go-errors/errors"
	"github.com/golang/protobuf/proto"
	"go.ligato.io/vpp-agent/v3/pkg/models"
	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"google.golang.org/protobuf/types/dynamicpb"
)

// ValidateSemantically validates given proto messages according to semantic validation(KVDescriptor.Validate)
// from registered KVDescriptors. If all messages are valid, nil is returned. If all message could be
// validated, kvscheduler.MessageValidationErrors is returned. In any other case, error is returned.
//
// Usage of dynamic proto messages (dynamicpb.Message) described by remotely known models is not supported.
// The reason for this is that the KVDescriptors can validate only statically generated proto messages and
// remotely retrieved dynamic proto messages can't be converted to such proto messages (there are
// no locally available statically generated proto models).
func (s *Scheduler) ValidateSemantically(messages []proto.Message) error {
	s.txnLock.Lock()
	defer s.txnLock.Unlock()

	invalidMessageErrors := make([]*api.InvalidMessageError, 0)
	for _, message := range messages {
		originalMessage := message

		// if needed, convert dynamic proto message to statically generated proto message
		// (validators in descriptors can validate only statically generated proto messages)
		if dynamicMessage, isDyn := message.(*dynamicpb.Message); isDyn {
			model, err := models.GetModelFor(message)
			if err != nil {
				return errors.Errorf("can't get model for message due to: %v (message=%v)", err, message)
			}
			goType := model.LocalGoType() // only for locally known models will return meaningful go type
			if goType == nil {
				return errors.Errorf("dynamic messages for remote models are not supported due to "+
					"not available go type of statically generated proto message (dynamic message=%v)", message)
			}
			message, err = models.DynamicMessageToGeneratedMessage(dynamicMessage, goType)
			if err != nil {
				return errors.Errorf("can't convert dynamic message to statically generated message "+
					"due to: %v (dynamic message=%v)", err, dynamicMessage)
			}
		}

		// get descriptor and key for given message
		key, err := models.GetKey(message)
		if err != nil {
			return errors.Errorf("can't get message key due to: %v (message=%v)", err, message)
		}
		descriptor := s.registry.GetDescriptorForKey(key)
		if descriptor == nil {
			s.Log.Debug("Skipping validation for proto message key %s "+
				"due to missing descriptor (proto message: %v)", key, message)
			continue
		}
		descHandler := newDescriptorHandler(descriptor)

		// validate and collect validation errors
		if err = descHandler.validate(key, message); err != nil {
			if ivError, ok := err.(*api.InvalidValueError); ok {
				// only InvalidValueErrors are supposed to describe data invalidity
				invalidMessageErrors = append(invalidMessageErrors,
					api.NewInvalidMessageError(originalMessage, ivError, nil))
			} else {
				return errors.Errorf("can't validate message due to: %v (message=%v)", err, message)
			}
		}

		// validate also derived values
		for _, derivedValue := range descHandler.derivedValues(key, message) {
			descriptor = s.registry.GetDescriptorForKey(derivedValue.Key)
			if descriptor == nil {
				s.Log.Debug("Skipping validation for proto message's derived value key %s "+
					"due to missing descriptor (proto message: %v, derived value proto message: %v)",
					derivedValue.Key, message, derivedValue.Value)
				continue
			}
			descHandler = newDescriptorHandler(descriptor)
			if err = descHandler.validate(derivedValue.Key, derivedValue.Value); err != nil {
				if ivError, ok := err.(*api.InvalidValueError); ok {
					// only InvalidValueErrors are supposed to describe data invalidity
					invalidMessageErrors = append(invalidMessageErrors,
						api.NewInvalidMessageError(derivedValue.Value, ivError, originalMessage))
				} else {
					return errors.Errorf("can't validate message due to: "+
						"%v (message=%v)", err, derivedValue.Value)
				}
			}
		}
	}

	if len(invalidMessageErrors) > 0 {
		return api.NewInvalidMessagesError(invalidMessageErrors)
	}
	return nil
}
