// Copyright (c) 2017 Cisco and/or its affiliates.
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

package adapter

import (
	"errors"
)

const (
	// DefaultBinapiSocket defines a default socket file path for VPP binary API.
	DefaultBinapiSocket = "/run/vpp-api.sock"
)

var (
	// ErrNotImplemented is an error returned when missing implementation.
	ErrNotImplemented = errors.New("not implemented for this OS")
)

// MsgCallback defines func signature for message callback.
type MsgCallback func(msgID uint16, data []byte)

// VppAPI provides connection to VPP binary API.
// It is responsible for sending and receiving of binary-encoded messages to/from VPP.
type VppAPI interface {
	// Connect connects the process to VPP.
	Connect() error

	// Disconnect disconnects the process from VPP.
	Disconnect() error

	// GetMsgID returns a runtime message ID for the given message name and CRC.
	GetMsgID(msgName string, msgCrc string) (msgID uint16, err error)

	// SendMsg sends a binary-encoded message to VPP.
	SendMsg(context uint32, data []byte) error

	// SetMsgCallback sets a callback function that will be called by the adapter whenever a message comes from VPP.
	SetMsgCallback(cb MsgCallback)

	// WaitReady waits until adapter is ready.
	WaitReady() error
}
