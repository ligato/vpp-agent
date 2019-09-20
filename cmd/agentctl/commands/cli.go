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
	"path"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/db/keyval/etcd"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/servicelabel"
	"google.golang.org/grpc"

	"github.com/ligato/vpp-agent/cmd/agentctl/utils"
	"github.com/ligato/vpp-agent/pkg/models"
)

type AgentCli struct {
	host string

	httpClient *utils.HTTPClient
}

func NewAgentCli() *AgentCli {
	return &AgentCli{}
}

func (cli *AgentCli) Init() {
	Debugf("init cli - global flags: %+v\n", global)

	httpAddr := net.JoinHostPort(global.AgentHost, global.PortHTTP)

	cli.httpClient = utils.NewHTTPClient(httpAddr)

	log := logrus.NewLogger("http-client")
	if global.Debug {
		log.SetLevel(logging.DebugLevel)
	}
	cli.httpClient.Log = log
}

func (cli *AgentCli) NewGRPCClient() *grpc.ClientConn {
	addr := net.JoinHostPort(global.AgentHost, global.PortGRPC)

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		ExitWithError(fmt.Errorf("connecting to gRPC failed: %v", err))
	}

	return conn
}

func (cli *AgentCli) NewKVDBClient() keyval.BytesBroker {
	etcdCfg := getEtcdConfig(global.Endpoints)

	log := logrus.NewLogger("kvdb-client")
	if global.Debug {
		log.SetLevel(logging.DebugLevel)
	} else {
		log.SetLevel(logging.WarnLevel)
	}

	kvdb, err := etcd.NewEtcdConnectionWithBytes(etcdCfg, log)
	if err != nil {
		ExitWithError(fmt.Errorf("connecting to Etcd failed: %v", err))
	}

	return &kvdbClient{kvdb}
}

func getEtcdConfig(endpoints []string) etcd.ClientConfig {
	cfg := etcd.ClientConfig{
		Config: &clientv3.Config{
			Endpoints:   endpoints,
			DialTimeout: time.Second * 3,
		},
		OpTimeout: time.Second * 10,
	}
	return cfg
}

func ensureAllAgentsPrefix(key string) string {
	if strings.HasPrefix(key, servicelabel.GetAllAgentsPrefix()) {
		return key
	}
	return path.Join(servicelabel.GetAllAgentsPrefix(), key)
}

func completeFullKey(key string) string {
	if strings.HasPrefix(key, servicelabel.GetAllAgentsPrefix()) {
		return key
	}
	if global.ServiceLabel == "" {
		ExitWithError(fmt.Errorf("service label is not defined, cannot get complete key"))
	}
	key = path.Join(servicelabel.GetAllAgentsPrefix(), global.ServiceLabel, key)
	return key
}

func stripAgentPrefix(key string) string {
	if !strings.HasPrefix(key, servicelabel.GetAllAgentsPrefix()) {
		return key
	}
	keyParts := strings.Split(key, "/")
	if len(keyParts) < 4 || keyParts[0] != "" {
		return path.Join(keyParts[2:]...)
	}
	return path.Join(keyParts[3:]...)
}

// kvdbClient provides client access to the KVDB server.
type kvdbClient struct {
	keyval.BytesBroker
}

func (k *kvdbClient) Put(key string, data []byte, opts ...datasync.PutOption) (err error) {
	key = completeFullKey(key)
	Debugf("kvdbClient.Put: %s", key)

	return k.BytesBroker.Put(key, data, opts...)
}

func (k *kvdbClient) GetValue(key string) (data []byte, found bool, revision int64, err error) {
	key = completeFullKey(key)
	Debugf("kvdbClient.GetValue: %s", key)

	return k.BytesBroker.GetValue(key)
}

func (k *kvdbClient) ListValues(prefix string) (keyval.BytesKeyValIterator, error) {
	prefix = ensureAllAgentsPrefix(prefix)
	Debugf("kvdbClient.ListValues: %s", prefix)

	return k.BytesBroker.ListValues(prefix)
}

func (k *kvdbClient) ListKeys(prefix string) (keyval.BytesKeyIterator, error) {
	prefix = ensureAllAgentsPrefix(prefix)
	Debugf("kvdbClient.ListKeys: %s", prefix)

	return k.BytesBroker.ListKeys(prefix)
}

func (k *kvdbClient) Delete(key string, opts ...datasync.DelOption) (existed bool, err error) {
	key = completeFullKey(key)
	Debugf("kvdbClient.Delete: %s", key)

	return k.BytesBroker.Delete(key, opts...)
}

type ModelDetail struct {
	Name            string
	Module          string
	Type            string
	Version         string
	Alias           string `json:",omitempty"`
	KeyPrefix       string
	NameTemplate    string `json:",omitempty"`
	ProtoName       string
	ProtoDescriptor *descriptor.DescriptorProto     `json:",omitempty"`
	ProtoFile       *descriptor.FileDescriptorProto `json:",omitempty"`
	Fields          protoFields                     `json:",omitempty"`
	Proto           string                          `json:",omitempty"`
	Location        string                          `json:",omitempty"`
}

type protoFields []*descriptor.FieldDescriptorProto

func (cli *AgentCli) AllModels() []ModelDetail {
	var list []ModelDetail

	registeredModels := models.RegisteredModels()
	Debugf("found %d registered models", len(registeredModels))

	for _, m := range registeredModels {
		module := strings.Split(m.Model.Module, ".")
		typ := m.Model.Type
		version := m.Model.Version

		name := fmt.Sprintf("%s.%s", m.Model.Module, typ)
		alias := fmt.Sprintf("%s.%s", module[0], typ)
		if alias == name {
			alias = ""
		}

		protoName := m.Info["protoName"]
		keyPrefix := m.Info["keyPrefix"]
		nameTemplate := m.Info["nameTemplate"]

		p := reflect.New(proto.MessageType(protoName)).Elem().Interface().(descriptor.Message)
		fd, _ := descriptor.ForMessage(p)

		detail := ModelDetail{
			Name:         name,
			Module:       m.Model.Module,
			Version:      version,
			Type:         typ,
			Alias:        alias,
			KeyPrefix:    keyPrefix,
			ProtoName:    protoName,
			NameTemplate: nameTemplate,
			//Fields:       dp.GetField(),
			//ProtoDescriptor: dp,
			//ProtoFile:       fd,
			//Proto:    proto.MarshalTextString(fd),
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
