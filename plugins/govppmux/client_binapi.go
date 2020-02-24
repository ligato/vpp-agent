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

package govppmux

import (
	"context"
	"runtime/trace"
	"time"

	govppapi "git.fd.io/govpp.git/api"
	"git.fd.io/govpp.git/core"
	"go.ligato.io/cn-infra/v2/logging"
)

// NewAPIChannel returns a new API channel for communication with VPP via govpp core.
// It uses default buffer sizes for the request and reply Go channels.
func (p *Plugin) NewAPIChannel() (govppapi.Channel, error) {
	ch, err := p.vppConn.NewAPIChannel()
	if err != nil {
		return nil, err
	}
	retryCfg := retryConfig{
		p.config.RetryRequestCount,
		p.config.RetryRequestTimeout,
	}
	return newGovppChan(ch, retryCfg), nil
}

// NewAPIChannelBuffered returns a new API channel for communication with VPP via govpp core.
// It allows to specify custom buffer sizes for the request and reply Go channels.
func (p *Plugin) NewAPIChannelBuffered(reqChanBufSize, replyChanBufSize int) (govppapi.Channel, error) {
	ch, err := p.vppConn.NewAPIChannelBuffered(reqChanBufSize, replyChanBufSize)
	if err != nil {
		return nil, err
	}
	retryCfg := retryConfig{
		p.config.RetryRequestCount,
		p.config.RetryRequestTimeout,
	}
	return newGovppChan(ch, retryCfg), nil
}

// goVppChan implements govpp channel interface. Instance is returned by NewAPIChannel() or NewAPIChannelBuffered(),
// and contains *govpp.channel dynamic type (vppChan field). Implemented methods allow custom handling of low-level
// govpp.
type goVppChan struct {
	govppapi.Channel
	// Retry data
	retry retryConfig
}

func newGovppChan(ch govppapi.Channel, retryCfg retryConfig) *goVppChan {
	govppChan := &goVppChan{
		Channel: ch,
		retry:   retryCfg,
	}
	reportChannelsOpened()
	return govppChan
}

func (c *goVppChan) Close() {
	c.Channel.Close()
	reportChannelsClosed()
}

// helper struct holding info about retry configuration
type retryConfig struct {
	attempts int
	timeout  time.Duration
}

// govppRequestCtx is custom govpp RequestCtx.
type govppRequestCtx struct {
	ctx  context.Context
	task *trace.Task

	// Original request context
	requestCtx govppapi.RequestCtx
	// Function allowing to re-send request in case it's granted by the config file
	sendRequest func(govppapi.Message) govppapi.RequestCtx
	// Parameter for sendRequest
	requestMsg govppapi.Message

	retry retryConfig
	start time.Time
}

// govppMultirequestCtx is custom govpp MultiRequestCtx.
type govppMultirequestCtx struct {
	ctx  context.Context
	task *trace.Task

	// Original multi request context
	requestCtx govppapi.MultiRequestCtx
	// Parameter for sendRequest
	requestMsg govppapi.Message

	start time.Time
}

// SendRequest sends asynchronous request to the vpp and receives context used to receive reply.
// Plugin govppmux allows to re-send retry which failed because of disconnected vpp, if enabled.
func (c *goVppChan) SendRequest(request govppapi.Message) govppapi.RequestCtx {
	ctx, task := trace.NewTask(context.Background(), "govpp.SendRequest")
	trace.Log(ctx, "messageName", request.GetMessageName())

	start := time.Now()
	// Send request now and wait for context
	requestCtx := c.Channel.SendRequest(request)

	reportRequestSent(request)

	// Return context with value and function which allows to send request again if needed
	return &govppRequestCtx{
		ctx:         ctx,
		task:        task,
		requestCtx:  requestCtx,
		sendRequest: c.Channel.SendRequest,
		requestMsg:  request,
		retry:       c.retry,
		start:       start,
	}
}

// ReceiveReply handles request and returns error if occurred. Also does retry if this option is available.
func (r *govppRequestCtx) ReceiveReply(reply govppapi.Message) error {
	defer r.task.End()

	var timeout time.Duration
	attempts := r.retry.attempts
	if r.retry.timeout > 0 { // Default value is 500ms
		timeout = r.retry.timeout
	}
	// Receive reply from original send
	err := r.requestCtx.ReceiveReply(reply)
	for retry := 1; err == core.ErrNotConnected; retry++ {
		if retry > attempts {
			break // max attempts exceeded
		}
		logging.Warnf("govppmux: request retry (%d/%d), message %s in %v",
			retry, attempts, r.requestMsg.GetMessageName(), timeout)
		// Wait before next attempt
		time.Sleep(timeout)
		// Retry request
		trace.Logf(r.ctx, "requestRetry", "%d/%d", retry, attempts)
		err = r.sendRequest(r.requestMsg).ReceiveReply(reply)
	}
	if err != nil {
		reportRequestFailed(reply, err)
	} else {
		reportRequestSuccess(r.requestMsg, r.start)
		reportRepliesReceived(reply)
	}
	return err
}

// SendMultiRequest sends asynchronous request to the vpp and receives context used to receive reply.
func (c *goVppChan) SendMultiRequest(request govppapi.Message) govppapi.MultiRequestCtx {
	ctx, task := trace.NewTask(context.Background(), "govpp.SendMultiRequest")
	trace.Log(ctx, "msgName", request.GetMessageName())

	start := time.Now()
	// Send request now and wait for context
	requestCtx := c.Channel.SendMultiRequest(request)

	reportRequestSent(request)

	// Return context with value and function which allows to send request again if needed
	return &govppMultirequestCtx{
		ctx:        ctx,
		task:       task,
		requestCtx: requestCtx,
		requestMsg: request,
		start:      start,
	}
}

// ReceiveReply handles request and returns error if occurred.
func (r *govppMultirequestCtx) ReceiveReply(reply govppapi.Message) (bool, error) {
	// Receive reply from original send
	last, err := r.requestCtx.ReceiveReply(reply)
	if last || err != nil {
		defer r.task.End()
		if err != nil {
			reportRequestFailed(r.requestMsg, err)
		} else {
			reportRequestSuccess(r.requestMsg, r.start)
		}
	} else {
		reportRepliesReceived(reply)
	}
	return last, err
}
