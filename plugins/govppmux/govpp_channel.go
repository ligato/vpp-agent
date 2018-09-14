package govppmux

import (
	"time"

	"github.com/ligato/cn-infra/logging/measure"

	govppapi "git.fd.io/govpp.git/api"
	"git.fd.io/govpp.git/core"
	"github.com/ligato/cn-infra/logging/logrus"
)

// goVppChan implements govpp channel interface. Instance is returned by NewAPIChannel() or NewAPIChannelBuffered(),
// and contains *govpp.channel dynamic type (vppChan field). Implemented methods allow custom handling of low-level
// govpp.
type goVppChan struct {
	govppapi.Channel
	// Retry data
	retry retryConfig
	// Stopwatch used to measure binary api call duration. Can be nil, in that case time is not measured (stopwatch
	// is disabled)
	stopwatch *measure.Stopwatch
}

// helper struct holding info about retry configuration
type retryConfig struct {
	attempts int
	timeout  time.Duration
}

// govppRequestCtx is custom govpp RequestCtx.
type govppRequestCtx struct {
	// Original request context
	requestCtx govppapi.RequestCtx
	// Function allowing to re-send request in case it's granted by the config file
	sendRequest func(govppapi.Message) govppapi.RequestCtx
	// Parameter for sendRequest
	requestMsg govppapi.Message
	// Retry data
	retry retryConfig
	// Stopwatch object
	stopwatch *measure.Stopwatch
	// Current duration
	started time.Time
}

// govppMultirequestCtx is custom govpp MultiRequestCtx.
type govppMultirequestCtx struct {
	// Original multi request context
	requestCtx govppapi.MultiRequestCtx
	// Function allowing to re-send request in case it's granted by the config file
	sendRequest func(govppapi.Message) govppapi.MultiRequestCtx
	// Parameter for sendRequest
	requestMsg govppapi.Message
	// Stopwatch object
	stopwatch *measure.Stopwatch
	// Current duration
	started time.Time
}

// SendRequest sends asynchronous request to the vpp and receives context used to receive reply.
// Plugin govppmux allows to re-send retry which failed because of disconnected vpp, if enabled.
func (c *goVppChan) SendRequest(request govppapi.Message) govppapi.RequestCtx {
	startTime := time.Now()

	logrus.DefaultLogger().Warnf("request sent %v", request.GetMessageName())
	sendRequest := c.Channel.SendRequest
	// Send request now and wait for context
	requestCtx := sendRequest(request)

	// Return context with value and function which allows to send request again if needed
	return &govppRequestCtx{
		requestCtx:  requestCtx,
		sendRequest: sendRequest,
		requestMsg:  request,
		retry:       c.retry,
		stopwatch:   c.stopwatch,
		started:     startTime,
	}
}

// ReceiveReply handles request and returns error if occurred. Also does retry if this option is available.
func (r *govppRequestCtx) ReceiveReply(reply govppapi.Message) error {
	defer func(t time.Time) {
		if r.stopwatch != nil {
			r.stopwatch.TimeLog(r.requestMsg.GetMessageName()).LogTimeEntry(time.Since(r.started))
			r.stopwatch.PrintLog()
		}
	}(time.Now())

	var timeout time.Duration
	maxAttempts := r.retry.attempts
	if r.retry.timeout > 0 { // Default value is 500ms
		timeout = r.retry.timeout
	}

	var err error
	// Receive reply from original send
	if err = r.requestCtx.ReceiveReply(reply); err == core.ErrNotConnected && maxAttempts > 0 {
		// Try to re-sent requests
		for attemptIdx := 1; attemptIdx <= maxAttempts; attemptIdx++ {
			// Wait, then try again
			time.Sleep(timeout)
			logrus.DefaultLogger().Warnf("Govppmux: retrying binary API message %v, attempt: %d",
				r.requestMsg.GetMessageName(), attemptIdx)
			if err = r.sendRequest(r.requestMsg).ReceiveReply(reply); err != core.ErrNotConnected {
				return err
			}
		}
	}

	return err
}

// SendMultiRequest sends asynchronous request to the vpp and receives context used to receive reply.
func (c *goVppChan) SendMultiRequest(request govppapi.Message) govppapi.MultiRequestCtx {
	startTime := time.Now()

	sendMultiRequest := c.Channel.SendMultiRequest
	// Send request now and wait for context
	requestCtx := sendMultiRequest(request)

	// Return context with value and function which allows to send request again if needed
	return &govppMultirequestCtx{
		requestCtx:  requestCtx,
		sendRequest: sendMultiRequest,
		requestMsg:  request,
		stopwatch:   c.stopwatch,
		started:     startTime,
	}
}

// ReceiveReply handles request and returns error if occurred.
func (r *govppMultirequestCtx) ReceiveReply(reply govppapi.Message) (bool, error) {
	// Receive reply from original send
	last, err := r.requestCtx.ReceiveReply(reply)
	r.stopwatch.TimeLog(r.requestMsg.GetMessageName()).LogTimeEntry(time.Since(r.started))

	if last && r.stopwatch != nil {
		r.stopwatch.PrintLog()
	}
	return last, err
}
