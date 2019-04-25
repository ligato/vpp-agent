package utils

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ligato/vpp-agent/api/models/linux"
	"github.com/ligato/vpp-agent/api/models/vpp"

	"github.com/ligato/vpp-agent/api/configurator"

	"github.com/ligato/cn-infra/health/statuscheck/model/status"

	"github.com/gogo/protobuf/proto"

	"github.com/ligato/cn-infra/db/keyval"
	acl "github.com/ligato/vpp-agent/api/models/vpp/acl"
	interfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	ipsec "github.com/ligato/vpp-agent/api/models/vpp/ipsec"
	l2 "github.com/ligato/vpp-agent/api/models/vpp/l2"
	l3 "github.com/ligato/vpp-agent/api/models/vpp/l3"
	nat "github.com/ligato/vpp-agent/api/models/vpp/nat"

	linterface "github.com/ligato/vpp-agent/api/models/linux/interfaces"
	ll3 "github.com/ligato/vpp-agent/api/models/linux/l3"

	"errors"
)

// VppMetaData defines the etcd metadata.
type VppMetaData struct {
	Rev int64
	Key string
}

// VppStatusWithMD contains a VPP Status data record and its etcd
// metadata.
type VppStatusWithMD struct {
	VppMetaData
	status.AgentStatus
}

// VppData defines a structure to hold all etcd data records (of all
// types) for one VPP.
type VppData struct {
	Status   map[string]VppStatusWithMD
	Config   configurator.Config
	ShowEtcd bool
	ShowConf bool
}

// EtcdDump is a map of VppData records. It constitutes a temporary
// storage for data retrieved from etcd. "Temporary" means during
// the execution of an agentctl command. Every command reads
// data from etcd first, then processes it, and finally either outputs
// the processed data to the user or updates one or more data records
// in etcd.
type EtcdDump map[string]*VppData

const (
	stsIDAgent = "Agent"
)

// NewEtcdDump returns a new instance of the temporary storage
// that will hold data retrieved from etcd.
func NewEtcdDump() EtcdDump {
	return make(EtcdDump)
}

func (ed EtcdDump) ReadStatusDataFromDb(db keyval.ProtoBroker, key string,
	agent string) (found bool, err error) {
	vd, ok := ed[agent]
	if !ok {
		vd = newVppDataRecord()
	}

	ps := strings.Split(key, "/")
	len := len(ps)

	ed[agent], err = readStatusFromDb(db, vd, key, ps[len-1])
	return true, err
}

// ReadDataFromDb reads a data record from etcd, parses it according to
// the expected record type and stores it in the EtcdDump temporary
// storage. A record is identified by a Key.
//
// The function returns an error if the etcd client encountered an
// error. The function returns true if the specified item has been
// found.
func (ed EtcdDump) ReadDataFromDb(db keyval.ProtoBroker, key string,
	agent string) (found bool, err error) {
	vd, ok := ed[agent]
	if !ok {
		vd = newVppDataRecord()
	}

	switch {
	case strings.HasPrefix(key, acl.ModelACL.KeyPrefix()):
		ed[agent], err = readACLConfigFromDb(db, vd, key)
	case strings.HasPrefix(key, interfaces.ModelInterface.KeyPrefix()):
		ed[agent], err = readInterfaceConfigFromDb(db, vd, key)
	case strings.HasPrefix(key, l2.ModelBridgeDomain.KeyPrefix()):
		ed[agent], err = readBridgeConfigFromDb(db, vd, key)
	case strings.HasPrefix(key, l2.ModelFIBEntry.KeyPrefix()):
		ed[agent], err = readFibTableConfigFromDb(db, vd, key)
	case strings.HasPrefix(key, l2.ModelXConnectPair.KeyPrefix()):
		ed[agent], err = readXConnectConfigFromDb(db, vd, key)
	case strings.HasPrefix(key, l3.ModelARPEntry.KeyPrefix()):
		ed[agent], err = readARPConfigFromDb(db, vd, key)
	case strings.HasPrefix(key, l3.ModelRoute.KeyPrefix()):
		ed[agent], err = readStatiRouteConfigFromDb(db, vd, key)
	case strings.HasPrefix(key, l3.ModelProxyARP.KeyPrefix()):
		ed[agent], err = readProxyARPConfigFromDb(db, vd, key)
	case strings.HasPrefix(key, l3.ModelIPScanNeighbor.KeyPrefix()):
		ed[agent], err = readIPScanNeightConfigFromDb(db, vd, key)
		//FIXME: Error in key
	//case NATPath:
	//	ed[label], err = readNATConfigFromDb(db, vd, key)
	//case DNATPath:
	//	ed[label], err = readDNATConfigFromDb(db, vd, key, params)
	case strings.HasPrefix(key, ipsec.ModelSecurityPolicyDatabase.KeyPrefix()):
		ed[agent], err = readIPSecPolicyConfigFromDb(db, vd, key)
	case strings.HasPrefix(key, ipsec.ModelSecurityAssociation.KeyPrefix()):
		ed[agent], err = readIPSecAssociateConfigFromDb(db, vd, key)
	case strings.HasPrefix(key, linterface.ModelInterface.KeyPrefix()):
		ed[agent], err = readLinuxInterfaceConfigFromDb(db, vd, key)
	case strings.HasPrefix(key, ll3.ModelARPEntry.KeyPrefix()):
		ed[agent], err = readLinuxARPConfigFromDb(db, vd, key)
	case strings.HasPrefix(key, ll3.ModelRoute.KeyPrefix()):
		ed[agent], err = readLinuxRouteConfigFromDb(db, vd, key)
	}

	return true, err
}

func readStatusFromDb(db keyval.ProtoBroker, vd *VppData, key string, name string) (*VppData, error) {
	id := stsIDAgent
	if name != "" {
		id = name
	}
	sts := status.AgentStatus{}
	found, rev, err := readDataFromDb(db, key, &sts)
	if found && err == nil {
		vd.Status[id] =
			VppStatusWithMD{VppMetaData{rev, key}, sts}
	}
	return vd, err
}

func readACLConfigFromDb(db keyval.ProtoBroker, vd *VppData, key string) (*VppData, error) {
	acl := &acl.ACL{}

	found, _, err := readDataFromDb(db, key, acl)
	if found && err == nil {
		vd.Config.VppConfig.Acls = append(vd.Config.VppConfig.Acls, acl)
	}
	return vd, err
}

func readInterfaceConfigFromDb(db keyval.ProtoBroker, vd *VppData, key string) (*VppData, error) {
	int := &interfaces.Interface{}

	found, _, err := readDataFromDb(db, key, int)
	if found && err == nil {
		vd.Config.VppConfig.Interfaces = append(vd.Config.VppConfig.Interfaces, int)
	}
	return vd, err
}

func readBridgeConfigFromDb(db keyval.ProtoBroker, vd *VppData, key string) (*VppData, error) {
	br := &l2.BridgeDomain{}

	found, _, err := readDataFromDb(db, key, br)
	if found && err == nil {
		vd.Config.VppConfig.BridgeDomains = append(vd.Config.VppConfig.BridgeDomains, br)
	}
	return vd, err
}

func readFibTableConfigFromDb(db keyval.ProtoBroker, vd *VppData, key string) (*VppData, error) {
	fib := &l2.FIBEntry{}

	found, _, err := readDataFromDb(db, key, fib)
	if found && err == nil {
		vd.Config.VppConfig.Fibs = append(vd.Config.VppConfig.Fibs, fib)
	}
	return vd, err
}

func readXConnectConfigFromDb(db keyval.ProtoBroker, vd *VppData, key string) (*VppData, error) {
	xconnect := &l2.XConnectPair{}

	found, _, err := readDataFromDb(db, key, xconnect)
	if found && err == nil {
		vd.Config.VppConfig.XconnectPairs = append(vd.Config.VppConfig.XconnectPairs, xconnect)
	}
	return vd, err
}

func readARPConfigFromDb(db keyval.ProtoBroker, vd *VppData, key string) (*VppData, error) {
	arp := &l3.ARPEntry{}

	found, _, err := readDataFromDb(db, key, arp)
	if found && err == nil {
		vd.Config.VppConfig.Arps = append(vd.Config.VppConfig.Arps, arp)
	}
	return vd, err
}

func readStatiRouteConfigFromDb(db keyval.ProtoBroker, vd *VppData, key string) (*VppData, error) {
	route := &l3.Route{}

	found, _, err := readDataFromDb(db, key, route)
	if found && err == nil {
		vd.Config.VppConfig.Routes = append(vd.Config.VppConfig.Routes, route)
	}
	return vd, err
}

func readProxyARPConfigFromDb(db keyval.ProtoBroker, vd *VppData, key string) (*VppData, error) {
	parp := &l3.ProxyARP{}

	found, _, err := readDataFromDb(db, key, parp)
	if found && err == nil {
		vd.Config.VppConfig.ProxyArp = parp
	}
	return vd, err
}

func readIPScanNeightConfigFromDb(db keyval.ProtoBroker, vd *VppData, key string) (*VppData, error) {
	scan := &l3.IPScanNeighbor{}

	found, _, err := readDataFromDb(db, key, scan)
	if found && err == nil {
		vd.Config.VppConfig.IpscanNeighbor = scan
	}
	return vd, err
}

func readNATConfigFromDb(db keyval.ProtoBroker, vd *VppData, key string) (*VppData, error) {
	nat := &nat.Nat44Global{}

	found, _, err := readDataFromDb(db, key, nat)
	if found && err == nil {
		vd.Config.VppConfig.Nat44Global = nat
	}
	return vd, err
}

func readDNATConfigFromDb(db keyval.ProtoBroker, vd *VppData, key string) (*VppData, error) {
	dnat := &nat.DNat44{}

	found, _, err := readDataFromDb(db, key, dnat)
	if found && err == nil {
		vd.Config.VppConfig.Dnat44S = append(vd.Config.VppConfig.Dnat44S, dnat)
	}
	return vd, err
}

func readIPSecPolicyConfigFromDb(db keyval.ProtoBroker, vd *VppData, key string) (*VppData, error) {
	policy := &ipsec.SecurityPolicyDatabase{}

	found, _, err := readDataFromDb(db, key, policy)
	if found && err == nil {
		vd.Config.VppConfig.IpsecSpds = append(vd.Config.VppConfig.IpsecSpds, policy)
	}
	return vd, err
}

func readIPSecAssociateConfigFromDb(db keyval.ProtoBroker, vd *VppData, key string) (*VppData, error) {
	ipsec := &ipsec.SecurityAssociation{}

	found, _, err := readDataFromDb(db, key, ipsec)
	if found && err == nil {
		vd.Config.VppConfig.IpsecSas = append(vd.Config.VppConfig.IpsecSas, ipsec)
	}
	return vd, err
}

func readLinuxInterfaceConfigFromDb(db keyval.ProtoBroker, vd *VppData, key string) (*VppData, error) {
	int := &linterface.Interface{}

	found, _, err := readDataFromDb(db, key, int)
	if found && err == nil {
		vd.Config.LinuxConfig.Interfaces = append(vd.Config.LinuxConfig.Interfaces, int)
	}
	return vd, err
}

func readLinuxARPConfigFromDb(db keyval.ProtoBroker, vd *VppData, key string) (*VppData, error) {
	arp := &ll3.ARPEntry{}

	found, _, err := readDataFromDb(db, key, arp)
	if found && err == nil {
		vd.Config.LinuxConfig.ArpEntries = append(vd.Config.LinuxConfig.ArpEntries, arp)
	}
	return vd, err
}

func readLinuxRouteConfigFromDb(db keyval.ProtoBroker, vd *VppData, key string) (*VppData, error) {
	route := &ll3.Route{}

	found, _, err := readDataFromDb(db, key, route)
	if found && err == nil {
		vd.Config.LinuxConfig.Routes = append(vd.Config.LinuxConfig.Routes, route)
	}
	return vd, err
}

func readDataFromDb(db keyval.ProtoBroker, key string, obj proto.Message) (bool, int64, error) {
	found, rev, err := db.GetValue(key, obj)
	if err != nil {
		return false, rev, errors.New("Could not read from database, Key:" + key + ", error" + err.Error())
	}
	if !found {
		fmt.Printf("WARNING: data for Key '%s' not found\n", key)
	}
	return found, rev, nil
}

func newVppDataRecord() *VppData {
	return &VppData{
		Status: make(map[string]VppStatusWithMD),
		Config: configurator.Config{
			VppConfig:   &vpp.ConfigData{},
			LinuxConfig: &linux.ConfigData{},
		},
		ShowEtcd: false,
		ShowConf: false,
	}
}

func (ed EtcdDump) getSortedKeys() []string {
	keys := []string{}
	for k := range ed {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
