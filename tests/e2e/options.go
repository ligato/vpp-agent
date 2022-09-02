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

// SetupOpt is options data holder for customizing setup of tests
type SetupOpt struct {
	AgentOptMods   []AgentOptModifier
	EtcdOptMods    []EtcdOptModifier
	DNSOptMods     []DNSOptModifier
	SetupAgent     bool
	SetupEtcd      bool
	SetupDNSServer bool

	ctx *TestCtx
}

// AgentOpt is options data holder for customizing setup of agent
type AgentOpt struct {
	Runtime             ComponentRuntime
	RuntimeStartOptions RuntimeStartOptionsFunc
	Name                string
	Image               string
	Env                 []string
	UseEtcd             bool
	InitialResync       bool
	ContainerOptsHook   func(*docker.CreateContainerOptions)
}

// MicroserviceOpt is options data holder for customizing setup of microservice
type MicroserviceOpt struct {
	Runtime             ComponentRuntime
	RuntimeStartOptions RuntimeStartOptionsFunc
	Name                string
	ContainerOptsHook   func(*docker.CreateContainerOptions)
}

// EtcdOpt is options data holder for customizing setup of ETCD
type EtcdOpt struct {
	Runtime                       ComponentRuntime
	RuntimeStartOptions           RuntimeStartOptionsFunc
	UseHTTPS                      bool
	UseTestContainerForNetworking bool
}

// DNSOpt is options data holder for customizing setup of DNS server
type DNSOpt struct {
	Runtime             ComponentRuntime
	RuntimeStartOptions RuntimeStartOptionsFunc

	// DomainNameSuffix is common suffix of all static dns entries configured in hostsConfig
	DomainNameSuffix string
	// HostsConfig is content of configuration of static DNS entries in hosts file format
	HostsConfig string
}

// RuntimeStartOptionsFunc is function that provides component runtime start options
type RuntimeStartOptionsFunc func(ctx *TestCtx, options interface{}) (interface{}, error)

// PingOpt are options for pinging command.
type PingOpt struct {
	AllowedLoss int    // percentage of allowed loss for success
	SourceIface string // outgoing interface name
	MaxTimeout  int    // timeout in seconds before ping exits
	Count       int    // number of pings
}

// SetupOptModifier is function customizing general setup options
type SetupOptModifier func(*SetupOpt)

// AgentOptModifier is function customizing Agent setup options
type AgentOptModifier func(*AgentOpt)

// MicroserviceOptModifier is function customizing Microservice setup options
type MicroserviceOptModifier func(*MicroserviceOpt)

// EtcdOptModifier is function customizing ETCD setup options
type EtcdOptModifier func(*EtcdOpt)

// DNSOptModifier is function customizing DNS server setup options
type DNSOptModifier func(*DNSOpt)

// PingOptModifier is modifiers of pinging options
type PingOptModifier func(*PingOpt)

// DefaultSetupOpt creates default values for SetupOpt
func DefaultSetupOpt(testCtx *TestCtx) *SetupOpt {
	opt := &SetupOpt{
		AgentOptMods:   nil,
		EtcdOptMods:    nil,
		DNSOptMods:     nil,
		SetupAgent:     true,
		SetupEtcd:      false,
		SetupDNSServer: false,
		ctx:            testCtx,
	}
	return opt
}

// DefaultEtcdOpt creates default values for EtcdOpt
func DefaultEtcdOpt(ctx *TestCtx) *EtcdOpt {
	return &EtcdOpt{
		Runtime: &ContainerRuntime{
			ctx:         ctx,
			logIdentity: "ETCD",
			stopTimeout: etcdStopTimeout,
		},
		RuntimeStartOptions:           ETCDStartOptionsForContainerRuntime,
		UseHTTPS:                      false,
		UseTestContainerForNetworking: false,
	}
}

// DefaultDNSOpt creates default values for DNSOpt
func DefaultDNSOpt(testCtx *TestCtx) *DNSOpt {
	return &DNSOpt{
		Runtime: &ContainerRuntime{
			ctx:         testCtx,
			logIdentity: "DNS server",
			stopTimeout: dnsStopTimeout,
		},
		RuntimeStartOptions: DNSServerStartOptionsForContainerRuntime,
		DomainNameSuffix:    "", // no DNS entries => no common domain name suffix
		HostsConfig:         "", // no DNS entries
	}
}

// DefaultMicroserviceiOpt creates default values for MicroserviceOpt
func DefaultMicroserviceOpt(testCtx *TestCtx, msName string) *MicroserviceOpt {
	return &MicroserviceOpt{
		Runtime: &ContainerRuntime{
			ctx:         testCtx,
			logIdentity: "Microservice " + msName,
			stopTimeout: msStopTimeout,
		},
		RuntimeStartOptions: MicroserviceStartOptionsForContainerRuntime,
		Name:                msName,
	}
}

// DefaultAgentOpt creates default values for AgentOpt
func DefaultAgentOpt(testCtx *TestCtx, agentName string) *AgentOpt {
	agentImg := agentImage
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
		Runtime: &ContainerRuntime{
			ctx:         testCtx,
			logIdentity: "Agent " + agentName,
			stopTimeout: agentStopTimeout,
		},
		RuntimeStartOptions: AgentStartOptionsForContainerRuntime,
		Name:                agentName,
		Image:               agentImg,
		UseEtcd:             false,
		InitialResync:       true,
		Env: []string{
			"INITIAL_LOGLVL=" + logging.DefaultLogger.GetLevel().String(),
			"ETCD_CONFIG=" + etcdConfig,
			"GRPC_CONFIG=" + grpcConfig,
			"DEBUG=" + os.Getenv("DEBUG"),
			"MICROSERVICE_LABEL=" + agentName,
		},
	}
	return opt
}

// DefaultPingOpts creates default values for PingOpt
func DefaultPingOpts() *PingOpt {
	return &PingOpt{
		AllowedLoss: 49, // by default at least half of the packets should get through
		MaxTimeout:  4,
	}
}

// WithoutVPPAgent is test setup option disabling vpp-agent setup
func WithoutVPPAgent() SetupOptModifier {
	return func(o *SetupOpt) {
		o.SetupAgent = false
	}
}

// WithCustomVPPAgent is test setup option using alternative vpp-agent image (customized original vpp-agent)
func WithCustomVPPAgent() SetupOptModifier {
	return func(o *SetupOpt) {
		o.AgentOptMods = append(o.AgentOptMods, func(ao *AgentOpt) {
			ao.Image = "vppagent.test.ligato.io:custom"
		})
	}
}

// WithEtcd is test setup option enabling etcd setup
func WithEtcd(etcdOptMods ...EtcdOptModifier) SetupOptModifier {
	return func(o *SetupOpt) {
		o.SetupEtcd = true
		o.EtcdOptMods = append(o.EtcdOptMods, etcdOptMods...)
	}
}

// WithDNSServer is test setup option enabling setup of container serving as dns server
func WithDNSServer(dnsOpts ...DNSOptModifier) SetupOptModifier {
	return func(o *SetupOpt) {
		o.SetupDNSServer = true
	}
}

// WithoutManualInitialAgentResync is test setup option disabling manual agent resync just after agent setup
func WithoutManualInitialAgentResync() AgentOptModifier {
	return func(o *AgentOpt) {
		o.InitialResync = false
	}
}

// WithAdditionalAgentCmdParams is test setup option adding additional command line parameters to executing vpp-agent
func WithAdditionalAgentCmdParams(params ...string) AgentOptModifier {
	return func(o *AgentOpt) {
		o.Env = append(o.Env, params...)
	}
}

// WithZonedStaticEntries is test setup option configuring group of static dns cache entries that belong
// to the same zone (have the same domain name suffix). The static dns cache entries are lines of config file
// in linux /etc/hosts file format.
// Currently supporting only one domain name suffix with static entries (even when DNS server solution supports
// multiple "zones" that each of them can be configured by one file in hosts file format)
func WithZonedStaticEntries(zoneDomainNameSuffix string, staticEntries ...string) DNSOptModifier {
	return func(o *DNSOpt) {
		o.DomainNameSuffix = zoneDomainNameSuffix
		o.HostsConfig = strings.Join(staticEntries, "\n")
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
	// subtest test names can container filepath.Separator
	testName := strings.ReplaceAll(ctx.t.Name(), string(filepath.Separator), "-")
	filePath, err := filepath.Abs(filepath.Join(ctx.testShareDir,
		fmt.Sprintf("e2e-test-%v-%v", testName, simpleFileName)))
	ctx.Expect(err).To(Not(HaveOccurred()))
	ctx.Expect(ioutil.WriteFile(filePath, []byte(fileContent), 0777)).To(Succeed())

	// TODO register in context and delete in teardown? this doesn't matter
	//  that much because file names contain unique test names so no file collision can happen
	return filePath
}

// WithMSContainerStartHook is microservice test setup option that will set the microservice container start
// hook that will modify the microservice start options.
func WithMSContainerStartHook(hook func(*docker.CreateContainerOptions)) MicroserviceOptModifier {
	return func(opt *MicroserviceOpt) {
		opt.ContainerOptsHook = hook
	}
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

// NewPingOpts create new PingOpt
func NewPingOpts(opts ...PingOptModifier) *PingOpt {
	options := DefaultPingOpts()
	for _, o := range opts {
		o(options)
	}
	return options
}

func (ping *PingOpt) args() []string {
	var args []string
	if ping.MaxTimeout > 0 {
		args = append(args, "-w", fmt.Sprint(ping.MaxTimeout))
	}
	if ping.Count > 0 {
		args = append(args, "-c", fmt.Sprint(ping.Count))
	}
	if ping.SourceIface != "" {
		args = append(args, "-I", ping.SourceIface)
	}
	return args
}

// PingWithAllowedLoss sets max allowed packet loss for pinging to be considered successful.
func PingWithAllowedLoss(maxLoss int) PingOptModifier {
	return func(opts *PingOpt) {
		opts.AllowedLoss = maxLoss
	}
}

// PingWithSourceInterface set source interface for ping packets.
func PingWithSourceInterface(iface string) PingOptModifier {
	return func(opts *PingOpt) {
		opts.SourceIface = iface
	}
}

func copyOptions(to interface{}, from interface{}) {
	fromVal := reflect.ValueOf(from).Elem()
	toVal := reflect.ValueOf(to).Elem()
	for i := 0; i < fromVal.NumField(); i++ {
		toVal.Field(i).Set(fromVal.Field(i))
	}
}
