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

package commands

import (
	"net"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/db/keyval/etcd"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"

	"github.com/ligato/vpp-agent/cmd/agentctl/utils"
)

type AgentCli struct {
	*utils.RestClient
	KVDB keyval.BytesBroker
}

func (cli *AgentCli) Initialize() {
	Debugf("[DEBUG] Initialize - globalsFlags: %+v\n\n", globalFlags)

	httpAddr := net.JoinHostPort(globalFlags.AgentAddr, globalFlags.HttpPort)
	cli.RestClient = utils.NewRestClient(httpAddr)

	etcdCfg := etcd.ClientConfig{
		Config: &clientv3.Config{
			Endpoints:   globalFlags.Endpoints,
			DialTimeout: time.Second * 3,
		},
		OpTimeout: time.Second * 10,
	}
	log := logrus.NewLogger("etcd")
	log.SetLevel(logging.WarnLevel)
	kvdb, err := etcd.NewEtcdConnectionWithBytes(etcdCfg, log)
	if err != nil {
		ExitWithError(err)
	}

	cli.KVDB = kvdb.NewBroker(globalFlags.ServiceLabel)
}
