//  Copyright (c) 2019 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package binapi

import (
	"errors"
	"fmt"

	govppapi "go.fd.io/govpp/api"
	"go.ligato.io/cn-infra/v2/logging"
)

// Versions is a map of all binapi messages for each supported VPP versions.
var Versions = map[Version]VersionMsgs{}

type CompatibilityChecker interface {
	// CheckCompatiblity checks compatibility with given binapi messages.
	CheckCompatiblity(...govppapi.Message) error
}

func CompatibleVersion(ch CompatibilityChecker) (Version, error) {
	if len(Versions) == 0 {
		return "", fmt.Errorf("no binapi versions loaded")
	}
	var picked = struct {
		version      Version
		incompatible int
	}{}
	for version, check := range Versions {
		var compErr *govppapi.CompatibilityError

		// check core compatibility
		coreMsgs := check.Core.AllMessages()
		if err := ch.CheckCompatiblity(coreMsgs...); errors.As(err, &compErr) {
			logging.Debugf("binapi version %v core incompatible (%d/%d messages)", version, len(compErr.IncompatibleMessages), len(coreMsgs))
			continue
		} else if err != nil {
			logging.Warnf("binapi version %v check failed: %v", version, err)
			continue
		}

		// check plugins compatibility
		pluginMsgs := check.Plugins.AllMessages()
		if err := ch.CheckCompatiblity(pluginMsgs...); errors.As(err, &compErr) {
			// some plugins might be disabled
			logging.Debugf("binapi version %v partly incompatible: (%d/%d messages)", version, len(compErr.IncompatibleMessages), len(pluginMsgs))
			if picked.version == "" || picked.incompatible > len(compErr.IncompatibleMessages) {
				picked.version = version
				picked.incompatible = len(compErr.IncompatibleMessages)
			}
		} else if err != nil {
			logging.Warnf("binapi version %v check failed: %v", version, err)
		} else {
			logging.Debugf("binapi version %v fully COMPATIBLE (%d messages)", version, len(coreMsgs)+len(pluginMsgs))
			return version, nil
		}
	}
	if picked.version != "" {
		logging.Debugf("choosing the most compatible binapi version: %v", picked.version)
		return picked.version, nil
	}
	return "", fmt.Errorf("no compatible binapi version found")
}

// VersionMsgs contains list of messages in version.
type VersionMsgs struct {
	Core    MessagesList
	Plugins MessagesList
}

// AllMessages returns messages from message list funcs combined.
func (vc VersionMsgs) AllMessages() []govppapi.Message {
	var msgs []govppapi.Message
	msgs = append(msgs, vc.Core.AllMessages()...)
	msgs = append(msgs, vc.Plugins.AllMessages()...)
	return msgs
}

// Version represents VPP version for generated binapi.
type Version string

// MessagesList aggregates multiple funcs that return messages.
type MessagesList []func() []govppapi.Message

// Messages is used to initialize message list.
func Messages(funcs ...func() []govppapi.Message) MessagesList {
	var list MessagesList
	list.Add(funcs...)
	return list
}

// Add adds funcs to message list.
func (list *MessagesList) Add(funcs ...func() []govppapi.Message) {
	for _, msgFunc := range funcs {
		*list = append(*list, msgFunc)
	}
}

// AllMessages returns messages from message list funcs combined.
func (list *MessagesList) AllMessages() []govppapi.Message {
	var msgs []govppapi.Message
	for _, msgFunc := range *list {
		msgs = append(msgs, msgFunc()...)
	}
	return msgs
}
