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

package vpp

import govppapi "git.fd.io/govpp.git/api"

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
