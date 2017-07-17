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

package etcdv3

import (
	"github.com/coreos/etcd/clientv3"
	"github.com/ligato/cn-infra/db/keyval/plugin"
	"github.com/ligato/cn-infra/logging"
)

// ProtoPluginEtcd implements Plugin interface therefore can be loaded with other plugins
type ProtoPluginEtcd struct {
	*plugin.Skeleton
	/*TODO
	Config *clientv3.Config //TODO `inject:""`
	client *clientv3.Client
	*/
}

// NewEtcdPlugin creates a new instance of ProtoPluginEtcd. Configuration of etcd connection is loaded from file.
func NewEtcdPlugin(cfg *Config) *ProtoPluginEtcd {

	skeleton := plugin.NewSkeleton(
		func(log logging.Logger) (plugin.Connection, error) {
			etcdConfig, err := ConfigToClientv3(cfg)
			if err != nil {
				return nil, err
			}
			return NewEtcdConnectionWithBytes(*etcdConfig, log)
		},
	)
	return &ProtoPluginEtcd{Skeleton: skeleton}
}

// NewEtcdPluginUsingClient creates a new instance of ProtoPluginEtcd using given etcd client
func NewEtcdPluginUsingClient(client *clientv3.Client) *ProtoPluginEtcd {
	skeleton := plugin.NewSkeleton(
		func(log logging.Logger) (plugin.Connection, error) {
			return NewEtcdConnectionUsingClient(client, log)
		},
	)
	return &ProtoPluginEtcd{Skeleton: skeleton}
}
