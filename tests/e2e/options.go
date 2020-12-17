// Copyright (c) 2020 Pantheon.tech
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

package e2e

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	docker "github.com/fsouza/go-dockerclient"
	. "github.com/onsi/gomega"
	"go.ligato.io/cn-infra/v2/logging"
)

const vppAgentDefaultImg = "ligato/vpp-agent:latest"

// SetupOpt is options data holder for customizing setup of tests
type SetupOpt struct {
	*AgentOpt
	*EtcdOpt
	SetupAgent bool
	SetupEtcd  bool
}

// AgentOpt is options data holder for customizing setup of agent
type AgentOpt struct {
	Image                 string
	Env                   []string
	UseEtcd               bool
	NoManualInitialResync bool
	ContainerOptsHook     func(*docker.CreateContainerOptions)
}

// EtcdOpt is options data holder for customizing setup of ETCD
type EtcdOpt struct {
	UseHTTPS                      bool
	UseTestContainerForNetworking bool
}

// SetupOptModifier is function customizing general setup options
type SetupOptModifier func(*SetupOpt)

// AgentOptModifier is function customizing Agent setup options
type AgentOptModifier func(*AgentOpt)

// EtcdOptModifier is function customizing ETCD seup options
type EtcdOptModifier func(*EtcdOpt)

// DefaultSetupOpt creates default values for SetupOpt
func DefaultSetupOpt() *SetupOpt {
	opt := &SetupOpt{
		AgentOpt:   DefaultAgentOpt(),
		EtcdOpt:    DefaultEtcdOpt(),
		SetupAgent: true,
		SetupEtcd:  false,
	}
	return opt
}

// DefaultEtcdOpt creates default values for EtcdOpt
func DefaultEtcdOpt() *EtcdOpt {
	return &EtcdOpt{
		UseHTTPS:                      false,
		UseTestContainerForNetworking: false,
	}
}

// DefaultAgentOpt creates default values for AgentOpt
func DefaultAgentOpt() *AgentOpt {
	agentImg := vppAgentDefaultImg
	if img := os.Getenv("VPP_AGENT"); img != "" {
		agentImg = img
	}
	grpcConfig := "grpc.conf"
	if val := os.Getenv("GRPC_CONFIG"); val != "" {
		grpcConfig = val
	}
	etcdConfig := "DISABLED"
	if val := os.Getenv("ETCD_CONFIG"); val != "" {
		etcdConfig = val
	}
	opt := &AgentOpt{
		Image:                 agentImg,
		UseEtcd:               false,
		NoManualInitialResync: false,
		Env: []string{
			"INITIAL_LOGLVL=" + logging.DefaultLogger.GetLevel().String(),
			"ETCD_CONFIG=" + etcdConfig,
			"GRPC_CONFIG=" + grpcConfig,
		},
	}
	return opt
}

// WithoutVPPAgent is test setup option disabling vpp-agent setup
func WithoutVPPAgent() SetupOptModifier {
	return func(o *SetupOpt) {
		o.SetupAgent = false
	}
}

// WithEtcd is test setup option enabling vpp-agent setup
func WithEtcd(etcdOpts ...EtcdOptModifier) SetupOptModifier {
	return func(o *SetupOpt) {
		o.SetupEtcd = true
		if o.EtcdOpt == nil {
			o.EtcdOpt = DefaultEtcdOpt()
		}
		for _, etcdOptModifier := range etcdOpts {
			etcdOptModifier(o.EtcdOpt)
		}
	}
}

// WithoutManualInitialAgentResync is test setup option disabling manual agent resync just after agent setup
func WithoutManualInitialAgentResync() AgentOptModifier {
	return func(o *AgentOpt) {
		o.NoManualInitialResync = true
	}
}

// WithAdditionalAgentCmdParams is test setup option adding additional command line parameters to executing vpp-agent
func WithAdditionalAgentCmdParams(params ...string) AgentOptModifier {
	return func(o *AgentOpt) {
		o.Env = append(o.Env, params...)
	}
}

// WithPluginConfigArg persists configContent for give VPP-Agent plugin (expecting generic plugin config name)
// and returns argument for VPP-Agent executable to use this plugin configuration file.
func WithPluginConfigArg(ctx *TestCtx, pluginName string, configContent string) string {
	configFilePath := CreateFileOnSharedVolume(ctx, fmt.Sprintf("%v.config", pluginName), configContent)
	return fmt.Sprintf("%v_CONFIG=%v", strings.ToUpper(pluginName), configFilePath)
}

// FIXME container that will use it can have it mounted in different location as seen by the container where
//  it is created (this works now due to the same mountpoint of shared volume in every container)

// CreateFileOnSharedVolume persists fileContent to file in mounted shared volume used for sharing file
// between containers. It returns the absolute path to the newly created file as seen by the container
// that creates it.
func CreateFileOnSharedVolume(ctx *TestCtx, simpleFileName string, fileContent string) string {
	filePath, err := filepath.Abs(filepath.Join(ctx.testShareDir,
		fmt.Sprintf("e2e-test-%v-%v", ctx.t.Name(), simpleFileName)))
	Expect(err).To(Not(HaveOccurred()))
	Expect(ioutil.WriteFile(filePath, []byte(fileContent), 0777)).To(Succeed())

	// TODO register in context and delete in teardown? this doesn't matter
	//  that much because file names contain unique test names so no file collision can happen
	return filePath
}

// WithEtcdHTTPsConnection is ETCD test setup option that will use HTTPS connection to ETCD (by default it is used
// unsecure HTTP connection)
func WithEtcdHTTPsConnection() EtcdOptModifier {
	return func(o *EtcdOpt) {
		o.UseHTTPS = true
	}
}

// WithEtcdTestContainerNetworking is ETCD test setup option that will use main Test container for
// networking (by default the ETCD has separate networking)
func WithEtcdTestContainerNetworking() EtcdOptModifier {
	return func(o *EtcdOpt) {
		o.UseTestContainerForNetworking = true
	}
}

func extractEtcdOptions(opt *SetupOpt) EtcdOptModifier {
	return func(etcdOpt *EtcdOpt) {
		copyOptions(etcdOpt, opt.EtcdOpt)
	}
}

func extractAgentOptions(opt *SetupOpt) AgentOptModifier {
	return func(agentOpt *AgentOpt) {
		copyOptions(agentOpt, opt.AgentOpt)
	}
}

func copyOptions(to interface{}, from interface{}) {
	fromVal := reflect.ValueOf(from).Elem()
	toVal := reflect.ValueOf(to).Elem()
	for i := 0; i < fromVal.NumField(); i++ {
		toVal.Field(i).Set(fromVal.Field(i))
	}
}
