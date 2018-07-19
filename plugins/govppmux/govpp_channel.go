package govppmux

import (
	"time"

	govppapi "git.fd.io/govpp.git/api"
	"git.fd.io/govpp.git/core"
	"github.com/ligato/cn-infra/logging/logrus"
)

const defaultRetryRequestTimeout = 500 * time.Millisecond

// goVppChan implements govpp channel interface. Instance is returned by NewAPIChannel() or NewAPIChannelBuffered(),
// and contains *govpp.channel dynamic type (vppChan field). Implemented methods allow custom handling of low-level
// govpp.
type goVppChan struct {
	vppChan govppapi.Channel
	// Retry data
	retry *retryConfig
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
	retry *retryConfig
}

// helper struct holding info about retry configuration
type retryConfig struct {
	attempts int
	timeout  time.Duration
}

// ReceiveReply handles request and returns error if occurred. Also does retry if this option is available.
func (r *govppRequestCtx) ReceiveReply(reply govppapi.Message) error {
	var err error
	// Receive reply from original send
	if err = r.requestCtx.ReceiveReply(reply); err != nil && err == core.ErrNotConnected {
		if r.retry != nil && r.retry.attempts > 0 {
			// Set default timeout between retries if not set
			if r.retry.timeout == 0 {
				r.retry.timeout = defaultRetryRequestTimeout
			}
			// Try to re-sent requests
			for i := 1; i <= r.retry.attempts; i++ {
				logrus.DefaultLogger().Warnf("Retrying message %v: %d", r.requestMsg.GetMessageName(), i)
				ctx := r.sendRequest(r.requestMsg)
				if err = ctx.ReceiveReply(reply); err == nil {
					return nil
				}
				time.Sleep(r.retry.timeout)
			}
		}
	}

	return err
}

// SendRequest sends asynchronous request to the vpp and receives context used to receive reply.
// Plugin govppmux allows to re-send retry which failed because of disconnected vpp, if enabled.
func (c *goVppChan) SendRequest(request govppapi.Message) govppapi.RequestCtx {
	sendRequest := c.vppChan.SendRequest
	// Send request now and wait for context
	requestCtx := sendRequest(request)

	// Return context with value and function which allows to send request again if needed
	return &govppRequestCtx{requestCtx, sendRequest, request, c.retry}
}

func (c *goVppChan) SendMultiRequest(request govppapi.Message) govppapi.MultiRequestCtx {
	return c.vppChan.SendMultiRequest(request)
}

func (c *goVppChan) SubscribeNotification(notifChan chan govppapi.Message, msgFactory func() govppapi.Message) (*govppapi.NotifSubscription, error) {
	return c.vppChan.SubscribeNotification(notifChan, msgFactory)
}

func (c *goVppChan) UnsubscribeNotification(subscription *govppapi.NotifSubscription) error {
	return c.vppChan.UnsubscribeNotification(subscription)
}

func (c *goVppChan) CheckMessageCompatibility(messages ...govppapi.Message) error {
	return c.vppChan.CheckMessageCompatibility(messages...)
}

func (c *goVppChan) SetReplyTimeout(timeout time.Duration) {
	c.vppChan.SetReplyTimeout(timeout)
}

func (c *goVppChan) GetRequestChannel() chan<- *govppapi.VppRequest {
	return c.vppChan.GetRequestChannel()
}

func (c *goVppChan) GetReplyChannel() <-chan *govppapi.VppReply {
	return c.vppChan.GetReplyChannel()
}

func (c *goVppChan) GetNotificationChannel() chan<- *govppapi.NotifSubscribeRequest {
	return c.vppChan.GetNotificationChannel()
}

func (c *goVppChan) GetNotificationReplyChannel() <-chan error {
	return c.vppChan.GetNotificationReplyChannel()
}

func (c *goVppChan) GetMessageDecoder() govppapi.MessageDecoder {
	return c.vppChan.GetMessageDecoder()
}

func (c *goVppChan) GetID() uint16 {
	return c.vppChan.GetID()
}

func (c *goVppChan) Close() {
	c.vppChan.Close()
}
