// Copyright (c) 2020 Pantheon.tech
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
	"context"
	. "github.com/onsi/gomega"
	"testing"
	"time"

	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/plugins/linux/ifplugin/linuxcalls"
	"go.ligato.io/vpp-agent/v3/plugins/netalloc/utils"
	"go.ligato.io/vpp-agent/v3/proto/ligato/kvscheduler"
	linux_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/linux/interfaces"
	linux_namespace "go.ligato.io/vpp-agent/v3/proto/ligato/linux/namespace"
	netalloc_api "go.ligato.io/vpp-agent/v3/proto/ligato/netalloc"
)

// Test dummy interfaces (additional loopbacks).
func TestDummyInterface(t *testing.T) {
	ctx := Setup(t)
	defer ctx.Teardown()

	const (
		dummy1Hostname = "lo1"
		dummy2Hostname = "lo2"
		ipAddr1        = "192.168.7.7"
		ipAddr2        = "10.7.7.7"
		ipAddr3        = "10.8.8.8"
		netMask        = "/24"
		msName         = "microservice1"
	)

	dummyIf1 := &linux_interfaces.Interface{
		Name:        "dummy1",
		Type:        linux_interfaces.Interface_DUMMY,
		Enabled:     true,
		IpAddresses: []string{ipAddr1 + netMask, ipAddr2 + netMask},
		HostIfName:  dummy1Hostname,
		Namespace: &linux_namespace.NetNamespace{
			Type:      linux_namespace.NetNamespace_MICROSERVICE,
			Reference: msNamePrefix + msName,
		},
	}
	dummyIf2 := &linux_interfaces.Interface{
		Name:        "dummy2",
		Type:        linux_interfaces.Interface_DUMMY,
		Enabled:     true,
		IpAddresses: []string{ipAddr3 + netMask},
		HostIfName:  dummy2Hostname,
		Namespace: &linux_namespace.NetNamespace{
			Type:      linux_namespace.NetNamespace_MICROSERVICE,
			Reference: msNamePrefix + msName,
		},
	}

	ctx.StartMicroservice(msName)
	req := ctx.GenericClient().ChangeRequest()
	err := req.Update(
		dummyIf1,
		dummyIf2,
	).Send(context.Background())
	Expect(err).ToNot(HaveOccurred())

	Eventually(ctx.GetValueStateClb(dummyIf1)).Should(Equal(kvscheduler.ValueState_CONFIGURED))
	Expect(ctx.GetValueState(dummyIf2)).To(Equal(kvscheduler.ValueState_CONFIGURED))
	Expect(ctx.PingFromMs(msName, ipAddr1)).To(Succeed())
	Expect(ctx.PingFromMs(msName, ipAddr2)).To(Succeed())
	Expect(ctx.PingFromMs(msName, ipAddr3)).To(Succeed())
	Expect(ctx.AgentInSync()).To(BeTrue())

	// Delete dummy2
	req = ctx.GenericClient().ChangeRequest()
	err = req.Delete(
		dummyIf2,
	).Send(context.Background())
	Expect(err).ToNot(HaveOccurred())

	Expect(ctx.GetValueState(dummyIf1)).To(Equal(kvscheduler.ValueState_CONFIGURED))
	Expect(ctx.GetValueState(dummyIf2)).ToNot(Equal(kvscheduler.ValueState_CONFIGURED))
	Expect(ctx.PingFromMs(msName, ipAddr1)).To(Succeed())
	Expect(ctx.PingFromMs(msName, ipAddr2)).To(Succeed())
	Expect(ctx.PingFromMs(msName, ipAddr3)).ToNot(Succeed())
	Expect(ctx.AgentInSync()).To(BeTrue())

	// restart microservice
	ctx.StopMicroservice(msName)
	Eventually(ctx.GetValueStateClb(dummyIf1)).Should(Equal(kvscheduler.ValueState_PENDING))
	Expect(ctx.AgentInSync()).To(BeTrue())
	ctx.StartMicroservice(msName)
	Eventually(ctx.GetValueStateClb(dummyIf1)).Should(Equal(kvscheduler.ValueState_CONFIGURED))
	Expect(ctx.PingFromMs(msName, ipAddr1)).To(Succeed())
	Expect(ctx.PingFromMs(msName, ipAddr2)).To(Succeed())
	Expect(ctx.AgentInSync()).To(BeTrue())
}

// Test interfaces created externally but with IP addresses assigned by the agent.
func TestExistingInterface(t *testing.T) {
	ctx := Setup(t)
	defer ctx.Teardown()

	const (
		ifaceHostName = "loop1"
		ifaceName     = "existing-loop1"
		ipAddr1       = "192.168.7.7"
		ipAddr2       = "10.7.7.7"
		ipAddr3       = "172.16.7.7"
		netMask       = "/24"
	)

	existingIface := &linux_interfaces.Interface{
		Name:        ifaceName,
		Type:        linux_interfaces.Interface_EXISTING,
		Enabled:     true,
		IpAddresses: []string{ipAddr1 + netMask, ipAddr2 + netMask},
		HostIfName:  ifaceHostName,
	}

	ifHandler := linuxcalls.NewNetLinkHandler(
		nil, nil, "", 0, logging.DefaultLogger)

	hasIP := func(ifName, ipAddr string) bool {
		addrs, err := ifHandler.GetAddressList(ifName)
		Expect(err).ToNot(HaveOccurred())
		for _, addr := range addrs {
			if addr.IP.String() == ipAddr {
				return true
			}
		}
		return false
	}

	addrKey := func(addr string) string {
		return linux_interfaces.InterfaceAddressKey(
			ifaceName, addr, netalloc_api.IPAddressSource_STATIC)
	}

	req := ctx.GenericClient().ChangeRequest()
	err := req.Update(
		existingIface,
	).Send(context.Background())
	Expect(err).ToNot(HaveOccurred())

	// referenced interface does not exist yet
	Expect(ctx.GetValueState(existingIface)).To(Equal(kvscheduler.ValueState_PENDING))

	// create referenced host interface using linuxcalls
	err = ifHandler.AddDummyInterface(ifaceHostName)
	Expect(err).ToNot(HaveOccurred())
	err = ifHandler.SetInterfaceUp(ifaceHostName)
	Expect(err).ToNot(HaveOccurred())

	Eventually(ctx.GetValueStateClb(existingIface)).Should(Equal(kvscheduler.ValueState_CONFIGURED))
	Eventually(ctx.GetDerivedValueStateClb(existingIface, addrKey(ipAddr1+netMask))).
		Should(Equal(kvscheduler.ValueState_CONFIGURED))
	Eventually(ctx.GetDerivedValueStateClb(existingIface, addrKey(ipAddr2+netMask))).
		Should(Equal(kvscheduler.ValueState_CONFIGURED))
	Expect(ctx.AgentInSync()).To(BeTrue())

	// check that the IP addresses have been configured
	Expect(hasIP(ifaceHostName, ipAddr1)).To(BeTrue())
	Expect(hasIP(ifaceHostName, ipAddr2)).To(BeTrue())

	// add third IP address externally, it should get removed by resync
	ipAddr, _, err := utils.ParseIPAddr(ipAddr3+netMask, nil)
	Expect(err).ToNot(HaveOccurred())
	err = ifHandler.AddInterfaceIP(ifaceHostName, ipAddr)
	Expect(err).ToNot(HaveOccurred())

	// resync should remove the address that was added externally
	Expect(ctx.AgentInSync()).To(BeFalse())
	Expect(hasIP(ifaceHostName, ipAddr1)).To(BeTrue())
	Expect(hasIP(ifaceHostName, ipAddr2)).To(BeTrue())
	Expect(hasIP(ifaceHostName, ipAddr3)).To(BeFalse())

	// remove the EXISTING interface (IP addresses should be unassigned)
	req = ctx.GenericClient().ChangeRequest()
	err = req.Delete(
		existingIface,
	).Send(context.Background())
	Expect(err).ToNot(HaveOccurred())
	Expect(ctx.GetValueState(existingIface)).ToNot(Equal(kvscheduler.ValueState_CONFIGURED))
	Expect(hasIP(ifaceHostName, ipAddr1)).To(BeFalse())
	Expect(hasIP(ifaceHostName, ipAddr2)).To(BeFalse())
	Expect(hasIP(ifaceHostName, ipAddr3)).To(BeFalse())

	// cleanup
	err = ifHandler.DeleteInterface(ifaceHostName)
	Expect(err).ToNot(HaveOccurred())
}

// Test interfaces created externally including the IP address assignments.
func TestExistingLinkOnlyInterface(t *testing.T) {
	ctx := Setup(t)
	defer ctx.Teardown()

	SetDefaultConsistentlyDuration(3 * time.Second)
	SetDefaultConsistentlyPollingInterval(time.Second)

	const (
		ifaceHostName = "loop1"
		ifaceName     = "existing-loop1"
		ipAddr1       = "192.168.7.7"
		ipAddr2       = "10.7.7.7"
		ipAddr3       = "172.16.7.7"
		netMask       = "/24"
	)

	existingIface := &linux_interfaces.Interface{
		Name:        ifaceName,
		Type:        linux_interfaces.Interface_EXISTING,
		Enabled:     true,
		IpAddresses: []string{ipAddr1 + netMask, ipAddr2 + netMask, ipAddr3 + netMask},
		HostIfName:  ifaceHostName,
		LinkOnly:    true, // <- agent does not configure IP addresses (they are also "existing")
	}

	ifHandler := linuxcalls.NewNetLinkHandler(
		nil, nil, "", 0, logging.DefaultLogger)

	hasIP := func(ifName, ipAddr string) bool {
		addrs, err := ifHandler.GetAddressList(ifName)
		Expect(err).ToNot(HaveOccurred())
		for _, addr := range addrs {
			if addr.IP.String() == ipAddr {
				return true
			}
		}
		return false
	}

	addrKey := func(addr string) string {
		return linux_interfaces.InterfaceAddressKey(
			ifaceName, addr, netalloc_api.IPAddressSource_EXISTING)
	}

	req := ctx.GenericClient().ChangeRequest()
	err := req.Update(
		existingIface,
	).Send(context.Background())
	Expect(err).ToNot(HaveOccurred())

	// the referenced interface does not exist yet
	Expect(ctx.GetValueState(existingIface)).To(Equal(kvscheduler.ValueState_PENDING))

	// create referenced host interface using linuxcalls (without IPs for now)
	err = ifHandler.AddDummyInterface(ifaceHostName)
	Expect(err).ToNot(HaveOccurred())
	err = ifHandler.SetInterfaceUp(ifaceHostName)
	Expect(err).ToNot(HaveOccurred())

	Eventually(ctx.GetValueStateClb(existingIface)).Should(Equal(kvscheduler.ValueState_CONFIGURED))
	Consistently(ctx.GetDerivedValueStateClb(existingIface, addrKey(ipAddr1+netMask))).
		Should(Equal(kvscheduler.ValueState_PENDING))
	Consistently(ctx.GetDerivedValueStateClb(existingIface, addrKey(ipAddr2+netMask))).
		Should(Equal(kvscheduler.ValueState_PENDING))
	Consistently(ctx.GetDerivedValueStateClb(existingIface, addrKey(ipAddr3+netMask))).
		Should(Equal(kvscheduler.ValueState_PENDING))
	Expect(ctx.AgentInSync()).To(BeTrue())

	// add IP addresses using linuxcalls (except ipAddr3)
	Expect(hasIP(ifaceHostName, ipAddr1)).To(BeFalse())
	Expect(hasIP(ifaceHostName, ipAddr2)).To(BeFalse())
	Expect(hasIP(ifaceHostName, ipAddr3)).To(BeFalse())
	ipAddr, _, err := utils.ParseIPAddr(ipAddr1+netMask, nil)
	Expect(err).ToNot(HaveOccurred())
	err = ifHandler.AddInterfaceIP(ifaceHostName, ipAddr)
	Expect(err).ToNot(HaveOccurred())
	ipAddr, _, err = utils.ParseIPAddr(ipAddr2+netMask, nil)
	Expect(err).ToNot(HaveOccurred())
	err = ifHandler.AddInterfaceIP(ifaceHostName, ipAddr)
	Expect(err).ToNot(HaveOccurred())

	// ipAddr1 and ipAddr2 should be eventually marked as configured
	Eventually(ctx.GetDerivedValueStateClb(existingIface, addrKey(ipAddr1+netMask))).
		Should(Equal(kvscheduler.ValueState_CONFIGURED))
	Eventually(ctx.GetDerivedValueStateClb(existingIface, addrKey(ipAddr2+netMask))).
		Should(Equal(kvscheduler.ValueState_CONFIGURED))
	Consistently(ctx.GetDerivedValueStateClb(existingIface, addrKey(ipAddr3+netMask))).
		Should(Equal(kvscheduler.ValueState_PENDING))
	Expect(ctx.AgentInSync()).To(BeTrue())

	// remove one IP address
	ipAddr, _, err = utils.ParseIPAddr(ipAddr1+netMask, nil)
	Expect(err).ToNot(HaveOccurred())
	err = ifHandler.DelInterfaceIP(ifaceHostName, ipAddr)
	Expect(err).ToNot(HaveOccurred())
	Eventually(ctx.GetDerivedValueStateClb(existingIface, addrKey(ipAddr1+netMask))).
		Should(Equal(kvscheduler.ValueState_PENDING))
	Consistently(ctx.GetDerivedValueStateClb(existingIface, addrKey(ipAddr2+netMask))).
		Should(Equal(kvscheduler.ValueState_CONFIGURED))
	Consistently(ctx.GetDerivedValueStateClb(existingIface, addrKey(ipAddr3+netMask))).
		Should(Equal(kvscheduler.ValueState_PENDING))

	// remove the EXISTING interface (the actual interface should be left untouched including IPs)
	req = ctx.GenericClient().ChangeRequest()
	err = req.Delete(
		existingIface,
	).Send(context.Background())
	Expect(err).ToNot(HaveOccurred())
	Expect(ctx.GetValueState(existingIface)).ToNot(Equal(kvscheduler.ValueState_CONFIGURED))
	Expect(hasIP(ifaceHostName, ipAddr1)).To(BeFalse())
	Expect(hasIP(ifaceHostName, ipAddr2)).To(BeTrue())
	Expect(hasIP(ifaceHostName, ipAddr3)).To(BeFalse())

	// cleanup
	err = ifHandler.DeleteInterface(ifaceHostName)
	Expect(err).ToNot(HaveOccurred())
}
