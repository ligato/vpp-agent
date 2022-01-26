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

package descriptor

import (
	"errors"
	"net"

	"go.ligato.io/cn-infra/v2/logging"
	"google.golang.org/protobuf/proto"

	vpp_syslog "go.ligato.io/vpp-agent/v3/examples/customize/custom_vpp_plugin/proto/custom/vpp/syslog"
	"go.ligato.io/vpp-agent/v3/examples/customize/custom_vpp_plugin/syslog/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/examples/customize/custom_vpp_plugin/syslog/vppcalls"
	"go.ligato.io/vpp-agent/v3/pkg/models"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
)

const (
	SyslogSenderDescriptorName = "vpp-syslog-sender"
)

var (
	ErrSyslogSenderWithoutCollector = errors.New("VPP syslog sender defined without collector IP")
)

type SenderDescriptor struct {
	log     logging.Logger
	handler vppcalls.SyslogVppAPI
}

func NewSenderDescriptor(handler vppcalls.SyslogVppAPI, log logging.LoggerFactory) *SenderDescriptor {
	return &SenderDescriptor{
		handler: handler,
		log:     log.NewLogger("syslog-descriptor"),
	}
}

func (d *SenderDescriptor) GetDescriptor() *adapter.SyslogSenderDescriptor {
	return &adapter.SyslogSenderDescriptor{
		Name:            SyslogSenderDescriptorName,
		NBKeyPrefix:     vpp_syslog.ModelSyslogSender.KeyPrefix(),
		ValueTypeName:   vpp_syslog.ModelSyslogSender.ProtoName(),
		KeySelector:     vpp_syslog.ModelSyslogSender.IsKeyValid,
		KeyLabel:        vpp_syslog.ModelSyslogSender.StripKeyPrefix,
		ValueComparator: d.EquivalentIPRedirect,
		Validate:        d.Validate,
		Create:          d.Create,
		Delete:          d.Delete,
		Retrieve:        d.Retrieve,
	}
}

// EquivalentIPRedirect is case-insensitive comparison function for punt.IpRedirect.
func (d *SenderDescriptor) EquivalentIPRedirect(key string, oldSender, newSender *vpp_syslog.Sender) bool {
	// parameters compared by proto equal
	return proto.Equal(oldSender, newSender)
}

func (d *SenderDescriptor) Validate(key string, sender *vpp_syslog.Sender) error {

	// validate collector IP
	if sender.Collector == "" || net.ParseIP(sender.Collector).IsUnspecified() {
		return kvs.NewInvalidValueError(ErrSyslogSenderWithoutCollector, "collector")
	}

	return nil
}

func (d *SenderDescriptor) Create(key string, sender *vpp_syslog.Sender) (metadata interface{}, err error) {
	// add Punt to host/socket
	err = d.handler.SetSender(vppcalls.SenderConfig{
		Source:    net.ParseIP(sender.Source),
		Collector: net.ParseIP(sender.Collector),
		Port:      int(sender.Port),
	})
	if err != nil {
		d.log.Error(err)
		return nil, err
	}
	return nil, nil
}

func (d *SenderDescriptor) Delete(key string, sender *vpp_syslog.Sender, metadata interface{}) error {
	err := d.handler.DisableSender()
	if err != nil {
		d.log.Error(err)
		return err
	}
	return nil
}

func (d *SenderDescriptor) Retrieve(correlate []adapter.SyslogSenderKVWithMetadata) (dump []adapter.SyslogSenderKVWithMetadata, err error) {
	cfg, err := d.handler.GetSender()
	if err != nil {
		d.log.Error(err)
		return nil, err
	}

	if cfg.Source != nil && cfg.Collector != nil {
		sender := &vpp_syslog.Sender{
			Source:    cfg.Source.String(),
			Collector: cfg.Collector.String(),
			Port:      int32(cfg.Port),
		}
		dump = append(dump, adapter.SyslogSenderKVWithMetadata{
			Key:    models.Key(sender),
			Value:  sender,
			Origin: kvs.FromNB,
		})
	}

	return dump, nil
}
