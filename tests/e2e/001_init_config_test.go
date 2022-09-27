//  Copyright (c) 2022 Cisco and/or its affiliates.
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

package e2e

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"

	"go.ligato.io/vpp-agent/v3/proto/ligato/kvscheduler"
)

// TestInitFromFile tests configuring initial state of NB from file
func TestInitFromFile(t *testing.T) {
	ctx := Setup(t, WithoutVPPAgent())
	defer ctx.Teardown() // will teardown also VPP-Agent created later

	// create init file content
	initialConfig := `
netallocConfig: {}
linuxConfig: {}
vppConfig:
  interfaces:
    - name: loop-test-from-init-file
      type: SOFTWARE_LOOPBACK
      enabled: true
      ipAddresses:
        - 10.10.1.1/24
      mtu: 9000
`
	initialConfigFileName := CreateFileOnSharedVolume(ctx, "initial-config.yaml", initialConfig)

	// create config content for init file usage
	initFileRegistryConfig := `
disable-initial-configuration: false
initial-configuration-file-path: %v
`
	initFileRegistryConfig = fmt.Sprintf(initFileRegistryConfig, initialConfigFileName)

	// create VPP-Agent
	ctx.StartAgent(
		mainAgentName,
		WithAdditionalAgentCmdParams(WithPluginConfigArg(ctx, "initfileregistry", initFileRegistryConfig)),
		WithoutManualInitialAgentResync(),
	)

	// check whether initial configuration inside file is correctly loaded in running VPP-Agent
	initInterfaceConfigState := func() kvscheduler.ValueState {
		return ctx.GetValueStateByKey("vpp/interface/loop-test-from-init-file/address/static/10.10.1.1/24")
	}
	ctx.Eventually(initInterfaceConfigState).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"loopback from init file was not properly created")
}

// TestInitFromEtcd tests configuring initial state of NB from Etcd
func TestInitFromEtcd(t *testing.T) {
	ctx := Setup(t,
		WithEtcd(),
		WithoutVPPAgent(),
	)
	defer ctx.Teardown() // will teardown also VPP-Agent created later

	// put NB config into Etcd
	ctx.Expect(ctx.Etcd.Put(
		fmt.Sprintf("/vnf-agent/%v/config/vpp/v2/interfaces/loop-test-from-etcd", AgentInstanceName(ctx)),
		`{"name":"loop-test-from-etcd","type":"SOFTWARE_LOOPBACK","enabled":true,"ip_addresses":["10.10.1.2/24"], "mtu":9000}`)).
		To(Succeed(), "can't insert data into ETCD")

	// prepare Etcd config for VPP-Agent
	etcdConfig := `insecure-transport: true
dial-timeout: 1s
endpoints:
    - "%v:2379"
`
	etcdIPAddress := ctx.Etcd.IPAddress()
	ctx.Expect(etcdIPAddress).ShouldNot(BeNil())
	etcdConfig = fmt.Sprintf(etcdConfig, etcdIPAddress)

	// create VPP-Agent
	ctx.StartAgent(
		mainAgentName,
		WithAdditionalAgentCmdParams(WithPluginConfigArg(ctx, "etcd", etcdConfig)),
		WithoutManualInitialAgentResync(),
	)

	// check whether NB initial configuration is correctly loaded from Etcd in running VPP-Agent
	initInterfaceConfigState := func() kvscheduler.ValueState {
		return ctx.GetValueStateByKey("vpp/interface/loop-test-from-etcd/address/static/10.10.1.2/24")
	}
	ctx.Eventually(initInterfaceConfigState).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"loopback from etcd was not properly created")
}

// TestInitFromFileAndEtcd tests configuring initial state of NB from Etcd and from file
func TestInitFromFileAndEtcd(t *testing.T) {
	ctx := Setup(t,
		WithEtcd(),
		WithoutVPPAgent(),
	)
	defer ctx.Teardown() // will teardown also VPP-Agent created later

	// put NB config into Etcd
	ctx.Expect(ctx.Etcd.Put(
		fmt.Sprintf("/vnf-agent/%v/config/vpp/v2/interfaces/memif-from-etcd", AgentInstanceName(ctx)),
		`{
"name":"memif-from-etcd",
"type":"MEMIF",
"enabled":true,
"ip_addresses":["10.10.1.1/32"], 
"mtu":1500,
"memif": {
		"master": false,
		"id": 1,
		"socket_filename": "/run/vpp/default.sock"
	}
}`)).To(Succeed(), "can't insert data1 into ETCD")
	ctx.Expect(ctx.Etcd.Put(
		fmt.Sprintf("/vnf-agent/%v/config/vpp/v2/interfaces/memif-from-both-sources", AgentInstanceName(ctx)),
		`{
"name":"memif-from-both-sources",
"type":"MEMIF",
"enabled":true,
"ip_addresses":["10.10.1.3/32"], 
"mtu":1500,
"memif": {
		"master": false,
		"id": 3,
		"socket_filename": "/run/vpp/default.sock"
	}
}`)).To(Succeed(), "can't insert data2 into ETCD")

	// create init file content
	initialConfig := `
netallocConfig: {}
linuxConfig: {}
vppConfig:
 interfaces:
   - name: memif-from-init-file
     type: MEMIF
     enabled: true
     ipAddresses:
       - 10.10.1.2/32
     mtu: 1500
     memif:
         master: false
         id: 2
         socketFilename: /run/vpp/default.sock
   - name: memif-from-both-sources
     type: MEMIF
     enabled: true
     ipAddresses:
       - 10.10.1.4/32
     mtu: 1500
     memif:
         master: false
         id: 4
         socketFilename: /run/vpp/default.sock
`
	initialConfigFileName := CreateFileOnSharedVolume(ctx, "initial-config.yaml", initialConfig)

	// create config content for NB init file usage
	initFileRegistryConfig := `
disable-initial-configuration: false
initial-configuration-file-path: %v
`
	initFileRegistryConfig = fmt.Sprintf(initFileRegistryConfig, initialConfigFileName)

	// create config content for etcd connection
	etcdConfig := `insecure-transport: true
dial-timeout: 1s
endpoints:
    - "%v:2379"
`
	etcdIPAddress := ctx.Etcd.IPAddress()
	ctx.Expect(etcdIPAddress).ShouldNot(BeNil())
	etcdConfig = fmt.Sprintf(etcdConfig, etcdIPAddress)

	// create VPP-Agent
	ctx.StartAgent(
		mainAgentName,
		WithAdditionalAgentCmdParams(WithPluginConfigArg(ctx, "etcd", etcdConfig),
			WithPluginConfigArg(ctx, "initfileregistry", initFileRegistryConfig)),
		WithoutManualInitialAgentResync(),
	)

	// check whether initial configuration is correctly loaded from Etcd and file in running VPP-Agent
	initInterfaceConfigState := func(interfaceName string, ipAddress string) kvscheduler.ValueState {
		return ctx.GetValueStateByKey(
			fmt.Sprintf("vpp/interface/%v/address/static/%v/32", interfaceName, ipAddress))
	}
	ctx.Eventually(initInterfaceConfigState("memif-from-etcd", "10.10.1.1")).
		Should(Equal(kvscheduler.ValueState_CONFIGURED),
			"unique memif from etcd was not properly created")
	ctx.Eventually(initInterfaceConfigState("memif-from-init-file", "10.10.1.2")).
		Should(Equal(kvscheduler.ValueState_CONFIGURED),
			"unique memif from init file was not properly created")
	ctx.Eventually(initInterfaceConfigState("memif-from-both-sources", "10.10.1.3")).
		Should(Equal(kvscheduler.ValueState_CONFIGURED),
			"conflicting memif (defined in init file and etcd) was either not correctly "+
				"merged (etcd data should have priority) or other things prevented its proper creation")
}
