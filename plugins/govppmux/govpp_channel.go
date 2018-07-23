package govppmux

import (
	"time"

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
}

// helper struct holding info about retry configuration
type retryConfig struct {
	attempts int
	timeout  time.Duration
}

// ReceiveReply handles request and returns error if occurred. Also does retry if this option is available.
func (r *govppRequestCtx) ReceiveReply(reply govppapi.Message) error {
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

// SendRequest sends asynchronous request to the vpp and receives context used to receive reply.
// Plugin govppmux allows to re-send retry which failed because of disconnected vpp, if enabled.
func (c *goVppChan) SendRequest(request govppapi.Message) govppapi.RequestCtx {
	sendRequest := c.Channel.SendRequest
	// Send request now and wait for context
	requestCtx := sendRequest(request)

	// Return context with value and function which allows to send request again if needed
	return &govppRequestCtx{requestCtx, sendRequest, request, c.retry}
}
