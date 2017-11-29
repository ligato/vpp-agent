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

//go:generate binapi-generator --input-dir=bin_api --output-dir=bin_api

package core

import (
	"errors"
	"os"
	"sync"
	"sync/atomic"
	"time"

	logger "github.com/sirupsen/logrus"

	"git.fd.io/govpp.git/adapter"
	"git.fd.io/govpp.git/api"
	"git.fd.io/govpp.git/core/bin_api/vpe"
	"github.com/fsnotify/fsnotify"
)

const (
	requestChannelBufSize      = 100 // default size of the request channel buffers
	replyChannelBufSize        = 100 // default size of the reply channel buffers
	notificationChannelBufSize = 100 // default size of the notification channel buffers
)

var (
	healthCheckProbeInterval = time.Second * 1        // default health check probe interval
	healthCheckReplyTimeout  = time.Millisecond * 100 // timeout for reply to a health check probe
	healthCheckThreshold     = 1                      // number of failed healthProbe until the error is reported
)

// ConnectionState holds the current state of the connection to VPP.
type ConnectionState int

const (
	// Connected connection state means that the connection to VPP has been successfully established.
	Connected ConnectionState = iota

	// Disconnected connection state means that the connection to VPP has been lost.
	Disconnected = iota
)

const (
	// watchedFolder is a folder where vpp's shared memory is supposed to be created.
	// File system events are monitored in this folder.
	watchedFolder = "/dev/shm/"
	// watchedFile is a name of the file in the watchedFolder. Once the file is present
	// the vpp is ready to accept a new connection.
	watchedFile = watchedFolder + "vpe-api"
)

// ConnectionEvent is a notification about change in the VPP connection state.
type ConnectionEvent struct {
	// Timestamp holds the time when the event has been generated.
	Timestamp time.Time

	// State holds the new state of the connection to VPP at the time when the event has been generated.
	State ConnectionState
}

// Connection represents a shared memory connection to VPP via vppAdapter.
type Connection struct {
	vpp       adapter.VppAdapter // VPP adapter
	connected uint32             // non-zero if the adapter is connected to VPP
	codec     *MsgCodec          // message codec

	msgIDs     map[string]uint16 // map of message IDs indexed by message name + CRC
	msgIDsLock sync.RWMutex      // lock for the message IDs map

	channels     map[uint32]*api.Channel // map of all API channels indexed by the channel ID
	channelsLock sync.RWMutex            // lock for the channels map

	notifSubscriptions     map[uint16][]*api.NotifSubscription // map od all notification subscriptions indexed by message ID
	notifSubscriptionsLock sync.RWMutex                        // lock for the subscriptions map

	maxChannelID uint32 // maximum used client ID
	pingReqID    uint16 // ID if the ControlPing message
	pingReplyID  uint16 // ID of the ControlPingReply message
}

// channelMetadata contains core-local metadata of an API channel.
type channelMetadata struct {
	id        uint32 // channel ID
	multipart uint32 // 1 if multipart request is being processed, 0 otherwise
}

var (
	log      *logger.Logger // global logger
	conn     *Connection    // global handle to the Connection (used in the message receive callback)
	connLock sync.RWMutex   // lock for the global connection
)

// init initializes global logger, which logs debug level messages to stdout.
func init() {
	log = logger.New()
	log.Out = os.Stdout
	log.Level = logger.DebugLevel
}

// SetLogger sets global logger to provided one.
func SetLogger(l *logger.Logger) {
	log = l
}

// SetHealthCheckProbeInterval sets health check probe interval.
// Beware: Function is not thread-safe. It is recommended to setup this parameter
// before connecting to vpp.
func SetHealthCheckProbeInterval(interval time.Duration) {
	healthCheckProbeInterval = interval
}

// SetHealthCheckReplyTimeout sets timeout for reply to a health check probe.
// If reply arrives after the timeout, check is considered as failed.
// Beware: Function is not thread-safe. It is recommended to setup this parameter
// before connecting to vpp.
func SetHealthCheckReplyTimeout(timeout time.Duration) {
	healthCheckReplyTimeout = timeout
}

// SetHealthCheckThreshold sets the number of failed healthProbe checks until the error is reported.
// Beware: Function is not thread-safe. It is recommended to setup this parameter
// before connecting to vpp.
func SetHealthCheckThreshold(threshold int) {
	healthCheckThreshold = threshold
}

// Connect connects to VPP using specified VPP adapter and returns the connection handle.
// This call blocks until VPP is connected, or an error occurs. Only one connection attempt will be performed.
func Connect(vppAdapter adapter.VppAdapter) (*Connection, error) {
	// create new connection handle
	c, err := newConnection(vppAdapter)
	if err != nil {
		return nil, err
	}

	// blocking attempt to connect to VPP
	err = c.connectVPP()
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// AsyncConnect asynchronously connects to VPP using specified VPP adapter and returns the connection handle
// and ConnectionState channel. This call does not block until connection is established, it
// returns immediately. The caller is supposed to watch the returned ConnectionState channel for
// Connected/Disconnected events. In case of disconnect, the library will asynchronously try to reconnect.
func AsyncConnect(vppAdapter adapter.VppAdapter) (*Connection, chan ConnectionEvent, error) {
	// create new connection handle
	c, err := newConnection(vppAdapter)
	if err != nil {
		return nil, nil, err
	}

	// asynchronously attempt to connect to VPP
	connChan := make(chan ConnectionEvent, notificationChannelBufSize)
	go c.connectLoop(connChan)

	return conn, connChan, nil
}

// Disconnect disconnects from VPP and releases all connection-related resources.
func (c *Connection) Disconnect() {
	if c == nil {
		return
	}
	connLock.Lock()
	defer connLock.Unlock()

	if c != nil && c.vpp != nil {
		c.disconnectVPP()
	}
	conn = nil
}

// newConnection returns new connection handle.
func newConnection(vppAdapter adapter.VppAdapter) (*Connection, error) {
	connLock.Lock()
	defer connLock.Unlock()

	if conn != nil {
		return nil, errors.New("only one connection per process is supported")
	}

	conn = &Connection{vpp: vppAdapter, codec: &MsgCodec{}}
	conn.channels = make(map[uint32]*api.Channel)
	conn.msgIDs = make(map[string]uint16)
	conn.notifSubscriptions = make(map[uint16][]*api.NotifSubscription)

	conn.vpp.SetMsgCallback(msgCallback)
	return conn, nil
}

// connectVPP performs one blocking attempt to connect to VPP.
func (c *Connection) connectVPP() error {
	log.Debug("Connecting to VPP...")

	// blocking connect
	err := c.vpp.Connect()
	if err != nil {
		log.Warn(err)
		return err
	}

	// store connected state
	atomic.StoreUint32(&c.connected, 1)

	// store control ping IDs
	c.pingReqID, _ = c.GetMessageID(&vpe.ControlPing{})
	c.pingReplyID, _ = c.GetMessageID(&vpe.ControlPingReply{})

	log.Info("Connected to VPP.")
	return nil
}

// disconnectVPP disconnects from VPP in case it is connected.
func (c *Connection) disconnectVPP() {
	if atomic.CompareAndSwapUint32(&c.connected, 1, 0) {
		c.vpp.Disconnect()
	}
}

func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// waitForVpp blocks until shared memory for sending bin api calls
// is present on the file system.
func waitForVpp() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	err = watcher.Add(watchedFolder)
	if err != nil {
		return err
	}

	if fileExists(watchedFile) {
		return nil
	}

	for {
		ev := <-watcher.Events
		if ev.Name == watchedFile && (ev.Op&fsnotify.Create) == fsnotify.Create {
			break
		}
	}
	return nil
}

// connectLoop attempts to connect to VPP until it succeeds.
// Then it continues with healthCheckLoop.
func (c *Connection) connectLoop(connChan chan ConnectionEvent) {
	// loop until connected
	for {
		waitForVpp()
		err := c.connectVPP()
		if err == nil {
			// signal connected event
			connChan <- ConnectionEvent{Timestamp: time.Now(), State: Connected}
			break
		}
	}

	// we are now connected, continue with health check loop
	c.healthCheckLoop(connChan)
}

// healthCheckLoop checks whether connection to VPP is alive. In case of disconnect,
// it continues with connectLoop and tries to reconnect.
func (c *Connection) healthCheckLoop(connChan chan ConnectionEvent) {
	// create a separate API channel for health check probes
	ch, err := conn.NewAPIChannel()
	if err != nil {
		log.Error("Error by creating health check API channel, health check will be disabled:", err)
		return
	}

	failedChecks := 0
	// send health check probes until an error occurs
	for {
		// wait for healthCheckProbeInterval
		<-time.After(healthCheckProbeInterval)

		if atomic.LoadUint32(&c.connected) == 0 {
			// Disconnect has been called in the meantime, return the healthcheck - reconnect loop
			log.Debug("Disconnected on request, exiting health check loop.")
			return
		}

		// send the control ping
		ch.ReqChan <- &api.VppRequest{Message: &vpe.ControlPing{}}

		// expect response within timeout period
		select {
		case vppReply := <-ch.ReplyChan:
			err = vppReply.Error
		case <-time.After(healthCheckReplyTimeout):
			err = errors.New("probe reply not received within the timeout period")
		}

		if err != nil {
			failedChecks++
		} else {
			failedChecks = 0
		}

		if failedChecks >= healthCheckThreshold {
			// in case of error, break & disconnect
			log.Errorf("VPP health check failed: %v", err)
			// signal disconnected event via channel
			connChan <- ConnectionEvent{Timestamp: time.Now(), State: Disconnected}
			break
		}
	}

	// cleanup
	ch.Close()
	c.disconnectVPP()

	// we are now disconnected, start connect loop
	c.connectLoop(connChan)
}

// NewAPIChannel returns a new API channel for communication with VPP via govpp core.
// It uses default buffer sizes for the request and reply Go channels.
func (c *Connection) NewAPIChannel() (*api.Channel, error) {
	if c == nil {
		return nil, errors.New("nil connection passed in")
	}
	return c.NewAPIChannelBuffered(requestChannelBufSize, replyChannelBufSize)
}

// NewAPIChannelBuffered returns a new API channel for communication with VPP via govpp core.
// It allows to specify custom buffer sizes for the request and reply Go channels.
func (c *Connection) NewAPIChannelBuffered(reqChanBufSize, replyChanBufSize int) (*api.Channel, error) {
	if c == nil {
		return nil, errors.New("nil connection passed in")
	}
	chID := atomic.AddUint32(&c.maxChannelID, 1)
	chMeta := &channelMetadata{id: chID}

	ch := api.NewChannelInternal(chMeta)
	ch.MsgDecoder = c.codec
	ch.MsgIdentifier = c

	// create the communication channels
	ch.ReqChan = make(chan *api.VppRequest, reqChanBufSize)
	ch.ReplyChan = make(chan *api.VppReply, replyChanBufSize)
	ch.NotifSubsChan = make(chan *api.NotifSubscribeRequest, reqChanBufSize)
	ch.NotifSubsReplyChan = make(chan error, replyChanBufSize)

	// store API channel within the client
	c.channelsLock.Lock()
	c.channels[chID] = ch
	c.channelsLock.Unlock()

	// start watching on the request channel
	go c.watchRequests(ch, chMeta)

	return ch, nil
}

// releaseAPIChannel releases API channel that needs to be closed.
func (c *Connection) releaseAPIChannel(ch *api.Channel, chMeta *channelMetadata) {
	log.WithFields(logger.Fields{
		"context": chMeta.id,
	}).Debug("API channel closed.")

	// delete the channel from channels map
	c.channelsLock.Lock()
	delete(c.channels, chMeta.id)
	c.channelsLock.Unlock()
}
