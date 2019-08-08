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
	"fmt"
	"net"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/db/keyval/etcd"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"

	"github.com/ligato/vpp-agent/cmd/agentctl/utils"
	"github.com/ligato/vpp-agent/pkg/models"
)

type Config struct {
}

type AgentCli struct {
	*utils.RestClient
}

func NewAgentCli() *AgentCli {
	return &AgentCli{}
}

func (cli *AgentCli) Initialize() {
	Debugf("[DEBUG] Initialize - globalsFlags: %+v\n\n", globalFlags)

	host := globalFlags.AgentHost
	if host == "" {
		host = "127.0.0.1"
	}
	httpAddr := net.JoinHostPort(host, globalFlags.HttpPort)
	cli.RestClient = utils.NewRestClient(httpAddr)
}

func (cli *AgentCli) KVDBClient() keyval.BytesBroker {
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

	return kvdb.NewBroker(globalFlags.ServiceLabel)
}

type ModelDetail struct {
	Module          []string
	Type            string
	Version         string
	Name            string
	Alias           string
	ProtoName       string
	KeyPrefix       string
	NameTemplate    string
	ProtoDescriptor *descriptor.DescriptorProto
	ProtoFile       *descriptor.FileDescriptorProto `json:",omitempty"`
	Location        string
}

func (cli *AgentCli) AllModels() []ModelDetail {
	var list []ModelDetail
	for _, m := range models.RegisteredModels() {
		module := strings.Split(m.Model.Module, ".")
		typ := m.Model.Type
		version := m.Model.Version

		name := fmt.Sprintf("%s.%s", m.Model.Module, typ)
		alias := fmt.Sprintf("%s.%s", module[0], typ)

		protoName := m.Info["protoName"]
		keyPrefix := m.Info["keyPrefix"]
		nameTemplate := m.Info["nameTemplate"]

		p := reflect.New(proto.MessageType(protoName)).Elem().Interface().(descriptor.Message)
		fd, dp := descriptor.ForMessage(p)

		detail := ModelDetail{
			Module:          module,
			Type:            typ,
			Version:         version,
			Name:            name,
			Alias:           alias,
			ProtoName:       protoName,
			KeyPrefix:       keyPrefix,
			NameTemplate:    nameTemplate,
			ProtoDescriptor: dp,
			//ProtoFile:       fd,
			Location: fd.GetName(),
		}

		Debugf(" - model detail: %+v", detail)

		list = append(list, detail)
	}
	sort.Sort(modelsByName(list))
	return list
}

type modelsByName []ModelDetail

func (s modelsByName) Len() int {
	return len(s)
}

func (s modelsByName) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

func (s modelsByName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
