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
	"testing"
	"time"

	. "github.com/onsi/gomega"

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
			Reference: MsNamePrefix + msName,
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
			Reference: MsNamePrefix + msName,
		},
	}

	ctx.StartMicroservice(msName)
	req := ctx.GenericClient().ChangeRequest()
	err := req.Update(
		dummyIf1,
		dummyIf2,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred())

	ctx.Eventually(ctx.GetValueStateClb(dummyIf1)).Should(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Expect(ctx.GetValueState(dummyIf2)).To(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Expect(ctx.PingFromMs(msName, ipAddr1)).To(Succeed())
	ctx.Expect(ctx.PingFromMs(msName, ipAddr2)).To(Succeed())
	ctx.Expect(ctx.PingFromMs(msName, ipAddr3)).To(Succeed())
	ctx.Expect(ctx.AgentInSync()).To(BeTrue())

	// Delete dummy2
	req = ctx.GenericClient().ChangeRequest()
	err = req.Delete(
		dummyIf2,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred())

	ctx.Expect(ctx.GetValueState(dummyIf1)).To(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Expect(ctx.GetValueState(dummyIf2)).ToNot(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Expect(ctx.PingFromMs(msName, ipAddr1)).To(Succeed())
	ctx.Expect(ctx.PingFromMs(msName, ipAddr2)).To(Succeed())
	ctx.Expect(ctx.PingFromMs(msName, ipAddr3)).ToNot(Succeed())
	ctx.Expect(ctx.AgentInSync()).To(BeTrue())

	// restart microservice
	ctx.StopMicroservice(msName)
	ctx.Eventually(ctx.GetValueStateClb(dummyIf1)).Should(Equal(kvscheduler.ValueState_PENDING))
	ctx.Expect(ctx.AgentInSync()).To(BeTrue())
	ctx.StartMicroservice(msName)
	ctx.Eventually(ctx.GetValueStateClb(dummyIf1)).Should(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Expect(ctx.PingFromMs(msName, ipAddr1)).To(Succeed())
	ctx.Expect(ctx.PingFromMs(msName, ipAddr2)).To(Succeed())
	ctx.Expect(ctx.AgentInSync()).To(BeTrue())

	// Disable dummy1
	dummyIf1.Enabled = false
	req = ctx.GenericClient().ChangeRequest()
	err = req.Update(
		dummyIf1,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred())
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

	ifHandler := ctx.Agent.LinuxInterfaceHandler()

	hasIP := func(ifName, ipAddr string) bool {
		addrs, err := ifHandler.GetAddressList(ifName)
		ctx.Expect(err).ToNot(HaveOccurred())
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
	ctx.Expect(err).ToNot(HaveOccurred())

	// referenced interface does not exist yet
	ctx.Expect(ctx.GetValueState(existingIface)).To(Equal(kvscheduler.ValueState_PENDING))

	// create referenced host interface using linuxcalls
	err = ifHandler.AddDummyInterface(ifaceHostName)
	ctx.Expect(err).ToNot(HaveOccurred())
	err = ifHandler.SetInterfaceUp(ifaceHostName)
	ctx.Expect(err).ToNot(HaveOccurred())

	ctx.Eventually(ctx.GetValueStateClb(existingIface)).Should(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Eventually(ctx.GetDerivedValueStateClb(existingIface, addrKey(ipAddr1+netMask))).
		Should(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Eventually(ctx.GetDerivedValueStateClb(existingIface, addrKey(ipAddr2+netMask))).
		Should(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Expect(ctx.AgentInSync()).To(BeTrue())

	// check that the IP addresses have been configured
	ctx.Expect(hasIP(ifaceHostName, ipAddr1)).To(BeTrue())
	ctx.Expect(hasIP(ifaceHostName, ipAddr2)).To(BeTrue())

	// add third IP address externally, it should get removed by resync
	ipAddr, _, err := utils.ParseIPAddr(ipAddr3+netMask, nil)
	ctx.Expect(err).ToNot(HaveOccurred())
	err = ifHandler.AddInterfaceIP(ifaceHostName, ipAddr)
	ctx.Expect(err).ToNot(HaveOccurred())

	// resync should remove the address that was added externally
	ctx.Expect(ctx.AgentInSync()).To(BeFalse())
	ctx.Expect(hasIP(ifaceHostName, ipAddr1)).To(BeTrue())
	ctx.Expect(hasIP(ifaceHostName, ipAddr2)).To(BeTrue())
	ctx.Expect(hasIP(ifaceHostName, ipAddr3)).To(BeFalse())

	// remove the EXISTING interface (IP addresses should be unassigned)
	req = ctx.GenericClient().ChangeRequest()
	err = req.Delete(
		existingIface,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred())
	ctx.Expect(ctx.GetValueState(existingIface)).ToNot(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Expect(hasIP(ifaceHostName, ipAddr1)).To(BeFalse())
	ctx.Expect(hasIP(ifaceHostName, ipAddr2)).To(BeFalse())
	ctx.Expect(hasIP(ifaceHostName, ipAddr3)).To(BeFalse())

	// cleanup
	err = ifHandler.DeleteInterface(ifaceHostName)
	ctx.Expect(err).ToNot(HaveOccurred())
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

	ifHandler := ctx.Agent.LinuxInterfaceHandler()

	hasIP := func(ifName, ipAddr string) bool {
		addrs, err := ifHandler.GetAddressList(ifName)
		ctx.Expect(err).ToNot(HaveOccurred())
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
	ctx.Expect(err).ToNot(HaveOccurred())

	// the referenced interface does not exist yet
	ctx.Expect(ctx.GetValueState(existingIface)).To(Equal(kvscheduler.ValueState_PENDING))

	// create referenced host interface using linuxcalls (without IPs for now)
	err = ifHandler.AddDummyInterface(ifaceHostName)
	ctx.Expect(err).ToNot(HaveOccurred())
	err = ifHandler.SetInterfaceUp(ifaceHostName)
	ctx.Expect(err).ToNot(HaveOccurred())

	ctx.Eventually(ctx.GetValueStateClb(existingIface)).Should(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Consistently(ctx.GetDerivedValueStateClb(existingIface, addrKey(ipAddr1+netMask))).
		Should(Equal(kvscheduler.ValueState_PENDING))
	ctx.Consistently(ctx.GetDerivedValueStateClb(existingIface, addrKey(ipAddr2+netMask))).
		Should(Equal(kvscheduler.ValueState_PENDING))
	ctx.Consistently(ctx.GetDerivedValueStateClb(existingIface, addrKey(ipAddr3+netMask))).
		Should(Equal(kvscheduler.ValueState_PENDING))
	ctx.Expect(ctx.AgentInSync()).To(BeTrue())

	// add IP addresses using linuxcalls (except ipAddr3)
	ctx.Expect(hasIP(ifaceHostName, ipAddr1)).To(BeFalse())
	ctx.Expect(hasIP(ifaceHostName, ipAddr2)).To(BeFalse())
	ctx.Expect(hasIP(ifaceHostName, ipAddr3)).To(BeFalse())
	ipAddr, _, err := utils.ParseIPAddr(ipAddr1+netMask, nil)
	ctx.Expect(err).ToNot(HaveOccurred())
	err = ifHandler.AddInterfaceIP(ifaceHostName, ipAddr)
	ctx.Expect(err).ToNot(HaveOccurred())
	ipAddr, _, err = utils.ParseIPAddr(ipAddr2+netMask, nil)
	ctx.Expect(err).ToNot(HaveOccurred())
	err = ifHandler.AddInterfaceIP(ifaceHostName, ipAddr)
	ctx.Expect(err).ToNot(HaveOccurred())

	// ipAddr1 and ipAddr2 should be eventually marked as configured
	ctx.Eventually(ctx.GetDerivedValueStateClb(existingIface, addrKey(ipAddr1+netMask))).
		Should(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Eventually(ctx.GetDerivedValueStateClb(existingIface, addrKey(ipAddr2+netMask))).
		Should(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Consistently(ctx.GetDerivedValueStateClb(existingIface, addrKey(ipAddr3+netMask))).
		Should(Equal(kvscheduler.ValueState_PENDING))
	ctx.Expect(ctx.AgentInSync()).To(BeTrue())

	// remove one IP address
	ipAddr, _, err = utils.ParseIPAddr(ipAddr1+netMask, nil)
	ctx.Expect(err).ToNot(HaveOccurred())
	err = ifHandler.DelInterfaceIP(ifaceHostName, ipAddr)
	ctx.Expect(err).ToNot(HaveOccurred())
	ctx.Eventually(ctx.GetDerivedValueStateClb(existingIface, addrKey(ipAddr1+netMask))).
		Should(Equal(kvscheduler.ValueState_PENDING))
	ctx.Consistently(ctx.GetDerivedValueStateClb(existingIface, addrKey(ipAddr2+netMask))).
		Should(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Consistently(ctx.GetDerivedValueStateClb(existingIface, addrKey(ipAddr3+netMask))).
		Should(Equal(kvscheduler.ValueState_PENDING))

	// remove the EXISTING interface (the actual interface should be left untouched including IPs)
	req = ctx.GenericClient().ChangeRequest()
	err = req.Delete(
		existingIface,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred())
	ctx.Expect(ctx.GetValueState(existingIface)).ToNot(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Expect(hasIP(ifaceHostName, ipAddr1)).To(BeFalse())
	ctx.Expect(hasIP(ifaceHostName, ipAddr2)).To(BeTrue())
	ctx.Expect(hasIP(ifaceHostName, ipAddr3)).To(BeFalse())

	// cleanup
	err = ifHandler.DeleteInterface(ifaceHostName)
	ctx.Expect(err).ToNot(HaveOccurred())
}
