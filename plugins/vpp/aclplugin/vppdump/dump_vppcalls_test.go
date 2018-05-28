package vppdump

import (
	"testing"

	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	acl_api "github.com/ligato/vpp-agent/plugins/vpp/binapi/acl"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpe"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	. "github.com/onsi/gomega"
)

// Test translation of IP rule into ACL Plugin's format
func TestGetIPRuleMatch(t *testing.T) {
	icmpV4Rule := getIPRuleMatches(acl_api.ACLRule{
		SrcIPAddr:      []byte{10, 0, 0, 1},
		SrcIPPrefixLen: 24,
		DstIPAddr:      []byte{20, 0, 0, 1},
		DstIPPrefixLen: 24,
		Proto:          ICMPv4Proto,
	})
	if icmpV4Rule.GetIcmp() == nil {
		t.Fatal("should have icmp match")
	}

	icmpV6Rule := getIPRuleMatches(acl_api.ACLRule{
		IsIpv6:			1,
		SrcIPAddr:      []byte{'d', 'e', 'd', 'd', 1},
		SrcIPPrefixLen: 64,
		DstIPAddr:      []byte{'d', 'e', 'd', 'd', 2},
		DstIPPrefixLen: 32,
		Proto:          ICMPv6Proto,
	})
	if icmpV6Rule.GetIcmp() == nil {
		t.Fatal("should have icmpv6 match")
	}

	tcpRule := getIPRuleMatches(acl_api.ACLRule{
		SrcIPAddr:      []byte{10, 0, 0, 1},
		SrcIPPrefixLen: 24,
		DstIPAddr:      []byte{20, 0, 0, 1},
		DstIPPrefixLen: 24,
		Proto:          TCPProto,
	})
	if tcpRule.GetTcp() == nil {
		t.Fatal("should have tcp match")
	}

	udpRule := getIPRuleMatches(acl_api.ACLRule{
		SrcIPAddr:      []byte{10, 0, 0, 1},
		SrcIPPrefixLen: 24,
		DstIPAddr:      []byte{20, 0, 0, 1},
		DstIPPrefixLen: 24,
		Proto:          UDPProto,
	})
	if udpRule.GetUdp() == nil {
		t.Fatal("should have udp match")
	}
}

// Test translation of MACIP rule into ACL Plugin's format
func TestGetMACIPRuleMatches(t *testing.T) {
	macipV4Rule := getMACIPRuleMatches(acl_api.MacipACLRule{
		IsPermit:		1,
		SrcMac: 		[]byte{2, 'd', 'e', 'a', 'd', 2},
		SrcMacMask: 	[]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		SrcIPAddr:  	[]byte{10, 0, 0, 1},
		SrcIPPrefixLen: 32,
	})
	if macipV4Rule.GetSourceMacAddress() == "" {
		t.Fatal("should have mac match")
	}
	macipV6Rule := getMACIPRuleMatches(acl_api.MacipACLRule{
		IsPermit:		0,
		IsIpv6:         1,
		SrcMac: 		[]byte{2, 'd', 'e', 'a', 'd', 2},
		SrcMacMask: 	[]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		SrcIPAddr:  	[]byte{'d', 'e', 'a', 'd', 1},
		SrcIPPrefixLen: 64,
	})
	if macipV6Rule.GetSourceMacAddress() == "" {
		t.Fatal("should have mac match")
	}
}

// Test dumping of IP rules
func TestDumpIPACL(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&acl_api.ACLDetails{
		ACLIndex: 0,
		Tag:      []byte{'a','c','l','1'},
		Count:    1,
		R:        []acl_api.ACLRule{ {IsPermit:1}},
	})
	ctx.MockVpp.MockReply(&acl_api.ACLDetails{
		ACLIndex: 1,
		Tag:      []byte{'a','c','l','2'},
		Count:    2,
		R:        []acl_api.ACLRule{ {IsPermit:0}, {IsPermit:2}},
	})
	ctx.MockVpp.MockReply(&acl_api.ACLDetails{
		ACLIndex: 2,
		Tag:      []byte{'a','c','l','3'},
		Count:    3,
		R:        []acl_api.ACLRule{ {IsPermit:0}, {IsPermit:1}, {IsPermit:2}},
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})
	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{
		SwIfIndex: 1,
		Count:     2,
		NInput:    1,
		Acls:      []uint32{0, 2},
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	swIfIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test", nil))
	swIfIndexes.RegisterName("if0", 1, nil)

	ifaces, err := DumpIPACL(swIfIndexes, logrus.DefaultLogger(), ctx.MockChannel, nil)
	Expect(err).To(Succeed())
	Expect(ifaces).To(HaveLen(3))
	//Expect(ifaces[0].Identifier.ACLIndex).To(Equal(uint32(0)))
	//Expect(ifaces[0].ACLDetails.Rules[0].AclAction).To(Equal(uint32(1)))
	//Expect(ifaces[1].Identifier.ACLIndex).To(Equal(uint32(1)))
	//Expect(ifaces[2].Identifier.ACLIndex).To(Equal(uint32(2)))
}

// Test dumping of MACIP rules
func TestDumpMACIPACL(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&acl_api.MacipACLDetails{
		ACLIndex: 0,
		Tag:      []byte{'a','c','l','1'},
		Count:    1,
		R:        []acl_api.MacipACLRule{ {IsPermit:1}},
	})
	ctx.MockVpp.MockReply(&acl_api.MacipACLDetails{
		ACLIndex: 1,
		Tag:      []byte{'a','c','l','2'},
		Count:    2,
		R:        []acl_api.MacipACLRule{ {IsPermit:0}, {IsPermit:2}},
	})
	ctx.MockVpp.MockReply(&acl_api.MacipACLDetails{
		ACLIndex: 2,
		Tag:      []byte{'a','c','l','3'},
		Count:    3,
		R:        []acl_api.MacipACLRule{ {IsPermit:0}, {IsPermit:1}, {IsPermit:2}},
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})
	ctx.MockVpp.MockReply(&acl_api.MacipACLInterfaceListDetails{
		SwIfIndex: 1,
		Count:     2,
		Acls:      []uint32{0, 2},
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	swIfIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test", nil))
	swIfIndexes.RegisterName("if0", 1, nil)

	ifaces, err := DumpMACIPACL(swIfIndexes, logrus.DefaultLogger(), ctx.MockChannel, nil)
	Expect(err).To(Succeed())
	Expect(ifaces).To(HaveLen(3))
	//Expect(ifaces[0].Identifier.ACLIndex).To(Equal(uint32(0)))
	//Expect(ifaces[0].ACLDetails.Rules[0].AclAction).To(Equal(uint32(1)))
	//Expect(ifaces[1].Identifier.ACLIndex).To(Equal(uint32(1)))
	//Expect(ifaces[2].Identifier.ACLIndex).To(Equal(uint32(2)))
}

// Test dumping of interfaces with assigned IP rules
func TestDumpACLInterfaces(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{
		SwIfIndex: 1,
		Count:     2,
		NInput:    1,
		Acls:      []uint32{0, 2},
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	swIfIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test", nil))
	swIfIndexes.RegisterName("if0", 1, nil)

	indexes := []uint32{0, 2}
	ifaces, err := DumpIPACLInterfaces(indexes, swIfIndexes, logrus.DefaultLogger(), ctx.MockChannel, nil)
	Expect(err).To(Succeed())
	Expect(ifaces).To(HaveLen(2))
	Expect(ifaces[0].Ingress).To(Equal([]string{"if0"}))
	Expect(ifaces[2].Egress).To(Equal([]string{"if0"}))
}

// Test dumping of interfaces with assigned MACIP rules
func TestDumpMACIPACLInterfaces(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&acl_api.MacipACLInterfaceListDetails{
		SwIfIndex: 1,
		Count:     2,
		Acls:      []uint32{0, 1},
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	swIfIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-sw_if_indexes", ifaceidx.IndexMetadata))
	swIfIndexes.RegisterName("if0", 1, nil)

	indexes := []uint32{0, 1}
	ifaces, err := DumpMACIPACLInterfaces(indexes, swIfIndexes, logrus.DefaultLogger(), ctx.MockChannel, nil)
	Expect(err).To(Succeed())
	Expect(ifaces).To(HaveLen(2))
	Expect(ifaces[0].Ingress).To(Equal([]string{"if0"}))
	Expect(ifaces[0].Egress).To(BeNil())
	Expect(ifaces[1].Ingress).To(Equal([]string{"if0"}))
	Expect(ifaces[1].Egress).To(BeNil())
}

// Test dumping of all configured ACLs with IP-type ruleData
func TestDumpIPAcls(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&acl_api.ACLDetails{
		ACLIndex: 0,
		Count:    1,
		R:        []acl_api.ACLRule{ {IsPermit:1}},
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	IPRuleACLs, err := DumpIPAcls(logrus.DefaultLogger(), ctx.MockChannel, nil)
	Expect(err).To(Succeed())
	Expect(IPRuleACLs).To(HaveLen(1))
}

// Test dumping of all configured ACLs with MACIP-type ruleData
func TestDumpMacIPAcls(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&acl_api.MacipACLDetails{
		ACLIndex: 0,
		Count:    1,
		R:        []acl_api.MacipACLRule{ {IsPermit:1}},
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	MacIPRuleACLs, err := DumpMacIPAcls(logrus.DefaultLogger(), ctx.MockChannel, nil)
	Expect(err).To(Succeed())
	Expect(MacIPRuleACLs).To(HaveLen(1))
}

func TestDumpInterfaceIPAcls(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{
		SwIfIndex: 0,
		Count:     2,
		NInput:    1,
		Acls:      []uint32{0, 1},
	})
	ctx.MockVpp.MockReply(&acl_api.ACLDetails{
		ACLIndex: 0,
		Count:    1,
		R:        []acl_api.ACLRule{{IsPermit: 1}, {IsPermit: 0}},
	})
	ctx.MockVpp.MockReply(&acl_api.ACLDetails{
		ACLIndex: 1,
		Count:    1,
		R:        []acl_api.ACLRule{{IsPermit: 2}, {IsPermit: 0}},
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	ACLs, err := DumpInterfaceIPAcls(logrus.DefaultLogger(), 0, ctx.MockChannel, nil)
	Expect(err).To(Succeed())
	Expect(ACLs.Acls).To(HaveLen(2))
}

func TestDumpInterfaceMACIPAcls(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&acl_api.MacipACLInterfaceListDetails{
		SwIfIndex: 0,
		Count:     2,
		Acls:      []uint32{0, 1},
	})
	ctx.MockVpp.MockReply(&acl_api.MacipACLDetails{
		ACLIndex: 0,
		Count:    1,
		R:        []acl_api.MacipACLRule{ {IsPermit:1}, {IsPermit:0} },
	})
	ctx.MockVpp.MockReply(&acl_api.MacipACLDetails{
		ACLIndex: 1,
		Count:    1,
		R:        []acl_api.MacipACLRule{ {IsPermit:2}, {IsPermit:1} },
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	ACLs, err := DumpInterfaceMACIPAcls(logrus.DefaultLogger(),0, ctx.MockChannel, nil)
	Expect(err).To(Succeed())
	Expect(ACLs.Acls).To(HaveLen(2))
}

func TestDumpInterface(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{
		SwIfIndex: 0,
		Count:     2,
		NInput:    1,
		Acls:      []uint32{0, 1},
	})
	IPacls, err := DumpInterfaceIPACLs(0, ctx.MockChannel, nil)
	Expect(err).To(BeNil())
	Expect(IPacls.Acls).To(HaveLen(2))

	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{})
	IPacls, err = DumpInterfaceIPACLs(0, ctx.MockChannel, nil)
	Expect(err).To(BeNil())
	Expect(IPacls.Acls).To(HaveLen(0))

	ctx.MockVpp.MockReply(&acl_api.MacipACLInterfaceListDetails{
		SwIfIndex: 0,
		Count:     2,
		Acls:      []uint32{0, 1},
	})
	MACIPacls, err := DumpInterfaceMACIPACLs(0, ctx.MockChannel, nil)
	Expect(err).To(BeNil())
	Expect(MACIPacls.Acls).To(HaveLen(2))

	ctx.MockVpp.MockReply(&acl_api.MacipACLInterfaceListDetails{})
	MACIPacls, err = DumpInterfaceMACIPACLs(0, ctx.MockChannel, nil)
	Expect(err).To(BeNil())
	Expect(MACIPacls.Acls).To(HaveLen(0))
}

func TestDumpInterfaces(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{
		SwIfIndex: 0,
		Count:     2,
		NInput:    1,
		Acls:      []uint32{0, 1},
	})
	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{
		SwIfIndex: 1,
		Count:     1,
		NInput:    1,
		Acls:      []uint32{2},
	})
	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{
		SwIfIndex: 2,
		Count:     2,
		NInput:    1,
		Acls:      []uint32{3, 4},
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})
	ctx.MockVpp.MockReply(&acl_api.MacipACLInterfaceListDetails{
		SwIfIndex: 3,
		Count:     2,
		Acls:      []uint32{6, 7},
	})
	ctx.MockVpp.MockReply(&acl_api.MacipACLInterfaceListDetails{
		SwIfIndex: 4,
		Count:     1,
		Acls:      []uint32{5},
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	IPacls, MACIPacls, err := DumpInterfaces(ctx.MockChannel, nil)
	Expect(err).To(BeNil())
	Expect(IPacls).To(HaveLen(3))
	Expect(MACIPacls).To(HaveLen(2))
}

