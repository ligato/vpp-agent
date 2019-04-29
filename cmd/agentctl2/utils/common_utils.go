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

package utils

import (
	"os"
	"strings"

	"github.com/go-errors/errors"

	"github.com/gogo/protobuf/proto"

	"fmt"

	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/db/keyval/etcd"
	"github.com/ligato/cn-infra/db/keyval/kvproto"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/servicelabel"
)

// Common exit flags
const (
	ExitSuccess = iota
	ExitError
	ExitBadConnection
	ExitInvalidInput
	ExitBadFeature
	ExitInterrupted
	ExitIO
	ExitBadArgs = 128
)

func GetAgentLabel(key string) (agentLabel string) {
	ps := strings.Split(strings.TrimPrefix(key, servicelabel.GetAllAgentsPrefix()), "/")
	if 1 > len(key) {
		ExitWithError(ExitInvalidInput, errors.New("Wrong key, key: "+key))
	}

	agentLabel = ps[0]

	return agentLabel
}

// GetDbForAllAgents opens a connection to etcd, specified in the command line
// or the "ETCD_ENDPOINTS" environment variable.
func GetDbForAllAgents(endpoints []string) (*kvproto.ProtoWrapper, error) {
	if len(endpoints) > 0 {
		ep := strings.Join(endpoints, ",")
		os.Setenv("ETCD_ENDPOINTS", ep)
	}

	cfg := &etcd.Config{}
	etcdConfig, err := etcd.ConfigToClient(cfg)

	// Log warnings and errors only.
	log := logrus.DefaultLogger()
	log.SetLevel(logging.WarnLevel)
	etcdBroker, err := etcd.NewEtcdConnectionWithBytes(*etcdConfig, log)
	if err != nil {
		return nil, err
	}

	return kvproto.NewProtoWrapperWithSerializer(etcdBroker, &keyval.SerializerJSON{}), nil
}

// GetDbForOneAgent opens a connection to etcd, specified in the command line
// or the "ETCD_ENDPOINTS" environment variable.
func GetDbForOneAgent(endpoints []string, agentLabel string) (keyval.ProtoBroker, error) {
	if len(endpoints) > 0 {
		ep := strings.Join(endpoints, ",")
		os.Setenv("ETCD_ENDPOINTS", ep)
	}

	cfg := &etcd.Config{}
	etcdConfig, err := etcd.ConfigToClient(cfg)

	// Log warnings and errors only.
	log := logrus.DefaultLogger()
	log.SetLevel(logging.WarnLevel)
	etcdBroker, err := etcd.NewEtcdConnectionWithBytes(*etcdConfig, log)
	if err != nil {
		return nil, err
	}

	return kvproto.NewProtoWrapperWithSerializer(etcdBroker, &keyval.SerializerJSON{}).
		NewBroker(servicelabel.GetAllAgentsPrefix() + agentLabel + "/"), nil
}

func GetModuleName(module proto.Message) string {
	str := proto.MessageName(module)

	tmp := strings.Split(str, ".")

	outstr := tmp[len(tmp)-1]

	if "linux" == tmp[0] {
		outstr = "Linux" + outstr
	}

	return outstr
}

// ExitWithError is used by all commands to print out an error
// and exit.
func ExitWithError(code int, err error) {
	fmt.Fprintln(os.Stderr, "Error: ", err)
	os.Exit(code)
}
