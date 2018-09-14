// Copyright (c) 2018 Cisco and/or its affiliates.
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
//
// DESCRIPTION:
//		GoVPP connection stub for unit tests
//		Required if tested entity uses measured channel
//

package govppmux

import (
	"git.fd.io/govpp.git/adapter"
	"git.fd.io/govpp.git/api"
	"git.fd.io/govpp.git/core"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/logging/measure"
)

// Connection represents govpp connection stub which allows to connect, disconnect and call extended govpp plugin API
type Connection struct {
	connection *core.Connection
}

// Connect creates new connection using provided VPP adapter and wrapped to govppmux.Connection type
func Connect(vppAdapter adapter.VppAdapter) (*Connection, error) {
	connection, err := core.Connect(vppAdapter)
	return &Connection{connection}, err
}

// Disconnect ends original connection
func (c *Connection) Disconnect() {
	c.connection.Disconnect()
}

// NewAPIChannel calls the method on vpp connection
func (c *Connection) NewAPIChannel() (api.Channel, error) {
	return c.connection.NewAPIChannel()
}

// NewAPIChannelBuffered calls the method on vpp connection
func (c *Connection) NewAPIChannelBuffered(reqChanBufSize, replyChanBufSize int) (api.Channel, error) {
	return c.connection.NewAPIChannelBuffered(reqChanBufSize, replyChanBufSize)
}

// NewMeasuredAPIChannel does not actually use stopwatch parameter, since it is not needed in tests
func (c *Connection) NewMeasuredAPIChannel(s *measure.Stopwatch) (api.Channel, error) {
	logrus.DefaultLogger().Warnf("Measured VPP channel created using govpp connection stub, stopwatch is ignored")
	return c.connection.NewAPIChannel()
}

// NewMeasuredAPIChannelBuffered does not actually use stopwatch parameter, since it is not needed in tests
func (c *Connection) NewMeasuredAPIChannelBuffered(reqChanBufSize, replyChanBufSize int, s *measure.Stopwatch) (api.Channel, error) {
	logrus.DefaultLogger().Warnf("Measured buffered VPP channel created using govpp connection stub, stopwatch is ignored")
	return c.connection.NewAPIChannelBuffered(reqChanBufSize, replyChanBufSize)
}
