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

package utils_test

import (
	"testing"
	"github.com/onsi/gomega"
	"github.com/ligato/vpp-agent/cmd/agentctl/utils"
	"github.com/ligato/vpp-agent/defaultplugins/ifplugin/model/interfaces"
)

func TestPrintDataAsJson(t *testing.T) {
	gomega.RegisterTestingT(t)
	etcdDump := getEtcdDataMap()

	buffer, err := etcdDump.PrintDataAsJSON(nil)
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(buffer).ToNot(gomega.BeNil())

	output := buffer.String()
	// Test whether both label from json data are present in output
	gomega.Expect(output).To(gomega.ContainSubstring(utils.IfConfig))
	gomega.Expect(output).To(gomega.ContainSubstring(utils.IfState))
}

func TestPrintDataAsText(t *testing.T) {
	gomega.RegisterTestingT(t)
	etcdDump := getEtcdDataMap()

	buffer := etcdDump.PrintDataAsText(false, false)
	gomega.Expect(buffer).ToNot(gomega.BeNil())

	output := buffer.String()
	// Test several key flags from text output
	gomega.Expect(output).To(gomega.ContainSubstring("vpp1"))
	gomega.Expect(output).To(gomega.ContainSubstring("iface1"))
	gomega.Expect(output).To(gomega.ContainSubstring("Stats"))
	gomega.Expect(output).To(gomega.ContainSubstring("IpAddr"))
}

func TestPrintDataAsTextWithEtcd(t *testing.T) {
	gomega.RegisterTestingT(t)
	etcdDump := getEtcdDataMap()

	buffer := etcdDump.PrintDataAsText(true, false)
	gomega.Expect(buffer).ToNot(gomega.BeNil())

	output := buffer.String()
	// Test ETCD output if 'showEtcd' is true
	gomega.Expect(output).To(gomega.ContainSubstring("ETCD"))
	gomega.Expect(output).To(gomega.ContainSubstring("Cfg"))
	gomega.Expect(output).To(gomega.ContainSubstring("Sts"))
}

func getEtcdDataMap() utils.EtcdDump{
	// Vpp metadata (the same for every entity)
	vppMetaData := utils.VppMetaData{
		Rev: 1,
		Key: "test-key",
	}

	// Interface config/status data container
	ifaceWithMD := utils.InterfaceWithMD{
		Config: &utils.IfconfigWithMD{ Metadata: vppMetaData, Interface: getInterface() },
		State: &utils.IfstateWithMD{ Metadata: vppMetaData, InterfaceState: getInterfaceStatus() },
	}

	// Interface label/data map
	interfaceMap := make(map[string]utils.InterfaceWithMD)
	interfaceMap["iface1"] = ifaceWithMD

	vppData := utils.VppData{
		Interfaces: interfaceMap,
	}

	// Vpp data map
	dataMap := make(map[string]*utils.VppData)
	dataMap["vpp1"] = &vppData

	// Return as etcd dump
	var etcdDump utils.EtcdDump = dataMap
	return etcdDump
}

func getInterface() *interfaces.Interfaces_Interface {
	return &interfaces.Interfaces_Interface{
		Name: "interface",
		IpAddresses: []string{"192.168.1.10"},
	}
}

func getInterfaceStatus() *interfaces.InterfacesState_Interface {
	return &interfaces.InterfacesState_Interface{
		Name: "interface",
		InternalName: "internal-interface",
		Statistics: &interfaces.InterfacesState_Interface_Statistics{
			InPackets: 10,
		},
	}
}