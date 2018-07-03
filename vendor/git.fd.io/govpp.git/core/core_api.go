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

package core

import (
	"time"

	"git.fd.io/govpp.git/api"
)

// ChannelProvider provides the communication channel with govpp core.
type ChannelProvider interface {
	// NewAPIChannel returns a new channel for communication with VPP via govpp core.
	// It uses default buffer sizes for the request and reply Go channels.
	NewAPIChannel() (Channel, error)

	// NewAPIChannelBuffered returns a new channel for communication with VPP via govpp core.
	// It allows to specify custom buffer sizes for the request and reply Go channels.
	NewAPIChannelBuffered(reqChanBufSize, replyChanBufSize int) (Channel, error)
}

// MessageDecoder provides functionality for decoding binary data to generated API messages.
type MessageDecoder interface {
	// DecodeMsg decodes binary-encoded data of a message into provided Message structure.
	DecodeMsg(data []byte, msg api.Message) error
}

// MessageIdentifier provides identification of generated API messages.
type MessageIdentifier interface {
	// GetMessageID returns message identifier of given API message.
	GetMessageID(msg api.Message) (uint16, error)
	// LookupByID looks up message name and crc by ID
	LookupByID(ID uint16) (string, error)
}

// Channel provides methods for direct communication with VPP channel.
type Channel interface {
	// SendRequest asynchronously sends a request to VPP. Returns a request context, that can be used to call ReceiveReply.
	// In case of any errors by sending, the error will be delivered to ReplyChan (and returned by ReceiveReply).
	SendRequest(msg api.Message) *RequestCtx
	// SendMultiRequest asynchronously sends a multipart request (request to which multiple responses are expected) to VPP.
	// Returns a multipart request context, that can be used to call ReceiveReply.
	// In case of any errors by sending, the error will be delivered to ReplyChan (and returned by ReceiveReply).
	SendMultiRequest(msg api.Message) *MultiRequestCtx
	// SubscribeNotification subscribes for receiving of the specified notification messages via provided Go channel.
	// Note that the caller is responsible for creating the Go channel with preferred buffer size. If the channel's
	// buffer is full, the notifications will not be delivered into it.
	SubscribeNotification(notifChan chan api.Message, msgFactory func() api.Message) (*NotifSubscription, error)
	// UnsubscribeNotification unsubscribes from receiving the notifications tied to the provided notification subscription.
	UnsubscribeNotification(subscription *NotifSubscription) error
	// CheckMessageCompatibility checks whether provided messages are compatible with the version of VPP
	// which the library is connected to.
	CheckMessageCompatibility(messages ...api.Message) error
	// SetReplyTimeout sets the timeout for replies from VPP. It represents the maximum time the API waits for a reply
	// from VPP before returning an error.
	SetReplyTimeout(timeout time.Duration)
	// GetRequestChannel returns request go channel of the VPP channel
	GetRequestChannel() chan<- *VppRequest
	// GetReplyChannel returns reply go channel of the VPP channel
	GetReplyChannel() <-chan *VppReply
	// GetNotificationChannel returns notification go channel of the VPP channel
	GetNotificationChannel() chan<- *NotifSubscribeRequest
	// GetNotificationReplyChannel returns notification reply go channel of the VPP channel
	GetNotificationReplyChannel() <-chan error
	// GetMessageDecoder returns message decoder instance
	GetMessageDecoder() MessageDecoder
	// GetID returns channel's ID
	GetID() uint16
	// Close closes the API channel and releases all API channel-related resources in the ChannelProvider.
	Close()
}
