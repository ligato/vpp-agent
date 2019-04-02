
package utils

import (
	"fmt"
	"sort"

	"github.com/gogo/protobuf/proto"

	"github.com/ligato/cn-infra/db/keyval"
	acl "github.com/ligato/vpp-agent/api/models/vpp/acl"
	interfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	l2 "github.com/ligato/vpp-agent/api/models/vpp/l2"
	l3 "github.com/ligato/vpp-agent/api/models/vpp/l3"

	"errors"
)

const (
	ACLPath          = "config/vpp/acls/v2/acl/"
	InterfacePath    = "config/vpp/v2/interfaces/"
	BridgeDomainPath = "config/vpp/l2/v2/bridge-domain/"
	FibTablePath     = "config/vpp/l2/v2/fib/"
	XConnectPath     = "config/vpp/l2/v2/xconnect/"
	ARPPath          = "config/vpp/v2/arp/"
	RoutePath        = "config/vpp/v2/route/"
	ProxyARPPath 	 = "config/vpp/v2/proxyarp-global"
)

// VppMetaData defines the etcd metadata.
type VppMetaData struct {
	Rev int64
	Key string
}

// ACLConfigWithMD contains a data record for interface configuration
// and its etcd metadata.
type ACLConfigWithMD struct {
	Metadata  VppMetaData
	ACL 	  *acl.ACL
}

// InterfaceWithMD contains a data record for interface and its
// etcd metadata.
type ACLWithMD struct {
	Config *ACLConfigWithMD
}

// IfConfigWithMD contains a data record for interface configuration
// and its etcd metadata.
type IfConfigWithMD struct {
	Metadata  VppMetaData
	Interface *interfaces.Interface
}

// InterfaceWithMD contains a data record for interface and its
// etcd metadata.
type InterfaceWithMD struct {
	Config *IfConfigWithMD
//	State  *IfStateWithMD
}

// BdConfigWithMD contains a data record for interface configuration
// and its etcd metadata.
type BdConfigWithMD struct {
	Metadata  VppMetaData
	BridgeDomain *l2.BridgeDomain
}

// BdWithMD contains a data record for interface and its
// etcd metadata.
type BdWithMD struct {
	Config *BdConfigWithMD
}

// FibTableConfigWithMD contains a data record for interface configuration
// and its etcd metadata.
type FibTableConfigWithMD struct {
	Metadata  VppMetaData
	FIBEntry *l2.FIBEntry
}

// FibTableWithMD contains a data record for interface and its
// etcd metadata.
type FibTableWithMD struct {
	Config *FibTableConfigWithMD
}

// XconnectConfigWithMD contains a data record for interface configuration
// and its etcd metadata.
type XconnectConfigWithMD struct {
	Metadata  VppMetaData
	Xconnect *l2.XConnectPair
}

// XconnectWithMD contains a data record for interface and its
// etcd metadata.
type XconnectWithMD struct {
	Config *XconnectConfigWithMD
}

// ARPConfigWithMD contains a data record for interface configuration
// and its etcd metadata.
type ARPConfigWithMD struct {
	Metadata  VppMetaData
	ARPEntry *l3.ARPEntry
}

// ARPWithMD contains a data record for interface and its
// etcd metadata.
type ARPWithMD struct {
	Config *ARPConfigWithMD
}

// StaticRouterConfigWithMD contains a data record for interface configuration
// and its etcd metadata.
type StaticRoutesConfigWithMD struct {
	Metadata  VppMetaData
	Route *l3.Route
}

// StaticRouterWithMD contains a data record for interface and its
// etcd metadata.
type StaticRoutesWithMD struct {
	Config *StaticRoutesConfigWithMD
}

// ProxyARPConfigWithMD contains a data record for interface configuration
// and its etcd metadata.
type ProxyARPConfigWithMD struct {
	Metadata  VppMetaData
	ProxyARP *l3.ProxyARP
}

// ProxyARPWithMD contains a data record for interface and its
// etcd metadata.
type ProxyARPWithMD struct {
	Config *ProxyARPConfigWithMD
}

// VppData defines a structure to hold all etcd data records (of all
// types) for one VPP.
type VppData struct {
	ACL 			   map[string]ACLWithMD
	Interfaces         map[string]InterfaceWithMD
//	InterfaceErrors    map[string]InterfaceErrorWithMD
	BridgeDomains      map[string]BdWithMD
//	BridgeDomainErrors map[string]BridgeDomainErrorWithMD
	FibTableEntries    FibTableWithMD
	XConnectPairs      map[string]XconnectWithMD
	ARP 			   ARPWithMD
	StaticRoutes       StaticRoutesWithMD
	ProxyARP 		   ProxyARPWithMD
//	Status             map[string]VppStatusWithMD
	ShowEtcd           bool
	ShowConf		   bool
}

// EtcdDump is a map of VppData records. It constitutes a temporary
// storage for data retrieved from etcd. "Temporary" means during
// the execution of an agentctl command. Every command reads
// data from etcd first, then processes it, and finally either outputs
// the processed data to the user or updates one or more data records
// in etcd.
type EtcdDump map[string]*VppData

// NewEtcdDump returns a new instance of the temporary storage
// that will hold data retrieved from etcd.
func NewEtcdDump() EtcdDump {
	return make(EtcdDump)
}

// ReadDataFromDb reads a data record from etcd, parses it according to
// the expected record type and stores it in the EtcdDump temporary
// storage. A record is identified by a Key.
//
// The function returns an error if the etcd client encountered an
// error. The function returns true if the specified item has been
// found.
func (ed EtcdDump) ReadDataFromDb(db keyval.ProtoBroker, key string) (found bool, err error) {
	label, dataType, params, _:= ParseKey(key)

	vd, ok := ed[label]
	if !ok {
		vd = newVppDataRecord()
	}

	switch dataType {
	case ACLPath:
		ed[label], err = readAclConfigFromDb(db, vd, key, params)
	case InterfacePath:
		ed[label], err = readInterfaceConfigFromDb(db, vd, key, params)
	//case BridgeDomainPath:
	//	ed[label], err = readBridgeConfigFromDb(db, vd, key, params)
	case FibTablePath:
		ed[label], err = readFibTableConfigFromDb(db, vd, key, params)
	case XConnectPath:
		ed[label], err = readXConnectConfigFromDb(db, vd, key, params)
	case ARPPath:
		ed[label], err = readARPConfigFromDb(db, vd, key, params)
	case RoutePath:
		ed[label], err = readStatiRouteConfigFromDb(db, vd, key, params)
	case ProxyARPPath:
		ed[label], err = readProxyARPConfigFromDb(db, vd, key)
	}

	return true, err
}

func readAclConfigFromDb(db keyval.ProtoBroker, vd *VppData, key string, name string) (*VppData, error) {
	if name == "" {
		fmt.Printf("WARNING: Invalid ACL config Key '%s'\n", key)
		return vd, nil
	}

	acl := &acl.ACL{}

	found, rev, err := readDataFromDb(db, key, acl)
	if found && err == nil {
		vd.ACL[name] = ACLWithMD{
			Config: &ACLConfigWithMD{VppMetaData{rev, key}, acl},
		}
	}
	return vd, err
}

func readInterfaceConfigFromDb(db keyval.ProtoBroker, vd *VppData, key string, name string) (*VppData, error) {
	if name == "" {
		fmt.Printf("WARNING: Invalid interface config Key '%s'\n", key)
		return vd, nil
	}

	int := &interfaces.Interface{}

	found, rev, err := readDataFromDb(db, key, int)
	if found && err == nil {
		vd.Interfaces[name] = InterfaceWithMD {
			Config: &IfConfigWithMD{VppMetaData{rev, key}, int},
		}
	}
	return vd, err
}

func readBridgeConfigFromDb(db keyval.ProtoBroker, vd *VppData, key string, name string) (*VppData, error) {
	if name == "" {
		fmt.Printf("WARNING: Invalid Bridge domain config Key '%s'\n", key)
		return vd, nil
	}

	br := &l2.BridgeDomain{}

	found, rev, err := readDataFromDb(db, key, br)
	if found && err == nil {
		vd.BridgeDomains[name] = BdWithMD{
			Config: &BdConfigWithMD{VppMetaData{rev, key}, br},
		}
	}
	return vd, err
}

func readFibTableConfigFromDb(db keyval.ProtoBroker, vd *VppData, key string, name string) (*VppData, error) {
	if name == "" {
		fmt.Printf("WARNING: Invalid Fib table config Key '%s'\n", key)
		return vd, nil
	}

	fib := &l2.FIBEntry{}

	found, rev, err := readDataFromDb(db, key, fib)
	if found && err == nil {
		vd.FibTableEntries = FibTableWithMD{
			Config: &FibTableConfigWithMD{VppMetaData{rev, key}, fib},
		}
	}
	return vd, err
}

func readXConnectConfigFromDb(db keyval.ProtoBroker, vd *VppData, key string, name string) (*VppData, error) {
	if name == "" {
		fmt.Printf("WARNING: Invalid xconnect config Key '%s'\n", key)
		return vd, nil
	}

	xconnect := &l2.XConnectPair{}

	found, rev, err := readDataFromDb(db, key, xconnect)
	if found && err == nil {
		vd.XConnectPairs[name] = XconnectWithMD{
			Config: &XconnectConfigWithMD{VppMetaData{rev, key}, xconnect},
		}
	}
	return vd, err
}

func readARPConfigFromDb(db keyval.ProtoBroker, vd *VppData, key string, name string) (*VppData, error) {
	if name == "" {
		fmt.Printf("WARNING: Invalid arp config Key '%s'\n", key)
		return vd, nil
	}

	arp := &l3.ARPEntry{}

	found, rev, err := readDataFromDb(db, key, arp)
	if found && err == nil {
		vd.ARP = ARPWithMD {
			Config: &ARPConfigWithMD{VppMetaData{rev, key}, arp},
		}
	}
	return vd, err
}

func readStatiRouteConfigFromDb(db keyval.ProtoBroker, vd *VppData, key string, name string) (*VppData, error) {
	if name == "" {
		fmt.Printf("WARNING: Invalid static route config Key '%s'\n", key)
		return vd, nil
	}

	route := &l3.Route{}

	found, rev, err := readDataFromDb(db, key, route)
	if found && err == nil {
		vd.StaticRoutes = StaticRoutesWithMD{
			Config: &StaticRoutesConfigWithMD{VppMetaData{rev, key}, route},
		}
	}
	return vd, err
}

func readProxyARPConfigFromDb(db keyval.ProtoBroker, vd *VppData, key string) (*VppData, error) {
	parp := &l3.ProxyARP{}

	found, rev, err := readDataFromDb(db, key, parp)
	if found && err == nil {
		vd.ProxyARP = ProxyARPWithMD{
			Config: &ProxyARPConfigWithMD{VppMetaData{rev, key}, parp},
		}
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
		ACL:				make(map[string]ACLWithMD),
		Interfaces:         make(map[string]InterfaceWithMD),
		BridgeDomains: 		make(map[string]BdWithMD),
		FibTableEntries: 	FibTableWithMD{},
		XConnectPairs:      make(map[string]XconnectWithMD),
		ARP:				ARPWithMD{},
		StaticRoutes:       StaticRoutesWithMD{},
		ShowEtcd:           false,
		ShowConf:			false,
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

