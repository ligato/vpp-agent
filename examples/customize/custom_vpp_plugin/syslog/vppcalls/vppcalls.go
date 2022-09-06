//  Copyright (c) 2020 Cisco and/or its affiliates.
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

package vppcalls

import (
	"net"

	govppapi "go.fd.io/govpp/api"
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/plugins/vpp"
)

type SenderConfig struct {
	Source    net.IP
	Collector net.IP
	Port      int
}

// SyslogVppAPI defines VPP handler API in vpp-version agnostic way.
// It cannot not use any generated binary API code directly.
type SyslogVppAPI interface {
	SetSender(sender SenderConfig) error
	GetSender() (*SenderConfig, error)
	DisableSender() error
}

var handler = vpp.RegisterHandler(vpp.HandlerDesc{
	Name:       "syslog",
	HandlerAPI: (*SyslogVppAPI)(nil),
})

type NewHandlerFunc func(ch govppapi.Channel, log logging.Logger) SyslogVppAPI

// AddHandlerVersion is used to register implementations of VPP handler API
// interface for a specific VPP version.
func AddHandlerVersion(version vpp.Version, msgs []govppapi.Message, h NewHandlerFunc) {
	handler.AddVersion(vpp.HandlerVersion{
		Version: version,
		Check: func(c vpp.Client) error {
			ch, err := c.NewAPIChannel()
			if err != nil {
				return err
			}
			return ch.CheckCompatiblity(msgs...)
		},
		NewHandler: func(c vpp.Client, a ...interface{}) vpp.HandlerAPI {
			ch, err := c.NewAPIChannel()
			if err != nil {
				return err
			}
			return h(ch, a[0].(logging.Logger))
		},
	})
}

// CompatibleVppHandler checks all the registered implementations to find the
// compatible handler implementation.
func CompatibleVppHandler(c vpp.Client, log logging.Logger) SyslogVppAPI {
	if v := handler.FindCompatibleVersion(c); v != nil {
		return v.NewHandler(c, log).(SyslogVppAPI)
	}
	return nil
}
