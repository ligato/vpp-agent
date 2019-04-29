package utils

import (
	"context"
	json2 "encoding/json"
	"strings"

	vpp_punt "github.com/ligato/vpp-agent/api/models/vpp/punt"

	"github.com/gogo/protobuf/proto"
	"github.com/ligato/cn-infra/db/keyval"
	acl "github.com/ligato/vpp-agent/api/models/vpp/acl"
	interfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	ipsec "github.com/ligato/vpp-agent/api/models/vpp/ipsec"
	l2 "github.com/ligato/vpp-agent/api/models/vpp/l2"
	l3 "github.com/ligato/vpp-agent/api/models/vpp/l3"
	nat "github.com/ligato/vpp-agent/api/models/vpp/nat"
	"github.com/ligato/vpp-agent/cmd/agentctl/utils"

	linterface "github.com/ligato/vpp-agent/api/models/linux/interfaces"
	ll3 "github.com/ligato/vpp-agent/api/models/linux/l3"

	"errors"
)

func WriteData(db keyval.ProtoTxn, key string, json string) {

	switch {
	case strings.HasPrefix(key, acl.ModelACL.KeyPrefix()):
		writeACLConfigToDb(db, key, json)
	case strings.HasPrefix(key, interfaces.ModelInterface.KeyPrefix()):
		writeInterfaceConfigToDb(db, key, json)
	case strings.HasPrefix(key, l2.ModelBridgeDomain.KeyPrefix()):
		writeBridgeDomainConfigToDb(db, key, json)
	case strings.HasPrefix(key, l2.ModelFIBEntry.KeyPrefix()):
		writeFibTableConfigToDb(db, key, json)
	case strings.HasPrefix(key, l2.ModelXConnectPair.KeyPrefix()):
		writeXConnectConfigToDb(db, key, json)
	case strings.HasPrefix(key, l3.ModelARPEntry.KeyPrefix()):
		writeARPConfigToDb(db, key, json)
	case strings.HasPrefix(key, l3.ModelRoute.KeyPrefix()):
		writeRouteConfigToDb(db, key, json)
	case strings.HasPrefix(key, l3.ModelProxyARP.KeyPrefix()):
		writeProxyConfigToDb(db, key, json)
	case strings.HasPrefix(key, l3.ModelIPScanNeighbor.KeyPrefix()):
		writeIPScanneConfigToDb(db, key, json)
	case strings.HasPrefix(key, nat.ModelNat44Global.KeyPrefix()):
		writeNATConfigToDb(db, key, json)
	case strings.HasPrefix(key, nat.ModelDNat44.KeyPrefix()):
		writeDNATConfigToDb(db, key, json)
	case strings.HasPrefix(key, ipsec.ModelSecurityPolicyDatabase.KeyPrefix()):
		writeIPSecPolicyConfigToDb(db, key, json)
	case strings.HasPrefix(key, ipsec.ModelSecurityAssociation.KeyPrefix()):
		writeIPSecAssociateConfigToDb(db, key, json)
	case strings.HasPrefix(key, vpp_punt.ModelIPRedirect.KeyPrefix()):
		writeIPRedirectConfigToDb(db, key, json)
	case strings.HasPrefix(key, vpp_punt.ModelToHost.KeyPrefix()):
		writeToHostConfigToDb(db, key, json)
	case strings.HasPrefix(key, linterface.ModelInterface.KeyPrefix()):
		writelInterfaceConfigToDb(db, key, json)
	case strings.HasPrefix(key, ll3.ModelARPEntry.KeyPrefix()):
		writelARPConfigToDb(db, key, json)
	case strings.HasPrefix(key, ll3.ModelRoute.KeyPrefix()):
		writelRouteConfigToDb(db, key, json)
	default:
		utils.ExitWithError(utils.ExitInvalidInput,
			errors.New("Unknown input key"))
	}
}

func writeACLConfigToDb(db keyval.ProtoTxn, key string, json string) {
	var err error

	vacl := &acl.ACL{}
	err = proto.Unmarshal([]byte(json), vacl)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed convert acl json format to protobuf - "+err.Error()))
	}

	err = writeDataToDb(db, key, vacl)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed write interface configuration to Etcd - "+err.Error()))
	}
}

func writeInterfaceConfigToDb(db keyval.ProtoTxn, key string, json string) {
	var err error

	intr := &interfaces.Interface{}

	err = json2.Unmarshal([]byte(json), intr)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed convert interface json format to protobuf - "+err.Error()))
	}

	err = writeDataToDb(db, key, intr)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed write interface configuration to Etcd - "+err.Error()))
	}
}

func writeBridgeDomainConfigToDb(db keyval.ProtoTxn, key string, json string) {
	var err error

	bridge := &l2.BridgeDomain{}

	err = json2.Unmarshal([]byte(json), bridge)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed convert interface json format to protobuf - "+err.Error()))
	}

	err = writeDataToDb(db, key, bridge)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed write interface configuration to Etcd - "+err.Error()))
	}
}

func writeFibTableConfigToDb(db keyval.ProtoTxn, key string, json string) {
	var err error

	fib := &l2.FIBEntry{}

	err = json2.Unmarshal([]byte(json), fib)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed convert fib json format to protobuf - "+err.Error()))
	}

	err = writeDataToDb(db, key, fib)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed write interface configuration to Etcd - "+err.Error()))
	}
}

func writeXConnectConfigToDb(db keyval.ProtoTxn, key string, json string) {
	var err error

	xconnect := &l2.XConnectPair{}

	err = json2.Unmarshal([]byte(json), xconnect)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed convert xconnect json format to protobuf - "+err.Error()))
	}

	err = writeDataToDb(db, key, xconnect)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed write interface configuration to Etcd - "+err.Error()))
	}
}

func writeARPConfigToDb(db keyval.ProtoTxn, key string, json string) {
	var err error

	arp := &l3.ARPEntry{}

	err = json2.Unmarshal([]byte(json), arp)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed convert arp json format to protobuf - "+err.Error()))
	}

	err = writeDataToDb(db, key, arp)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed write interface configuration to Etcd - "+err.Error()))
	}
}

func writeRouteConfigToDb(db keyval.ProtoTxn, key string, json string) {
	var err error

	route := &l3.Route{}

	err = json2.Unmarshal([]byte(json), route)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed convert route json format to protobuf - "+err.Error()))
	}

	err = writeDataToDb(db, key, route)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed write interface configuration to Etcd - "+err.Error()))
	}
}

func writeProxyConfigToDb(db keyval.ProtoTxn, key string, json string) {
	var err error

	proxy := &l3.ProxyARP{}

	err = json2.Unmarshal([]byte(json), proxy)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed convert proxy json format to protobuf - "+err.Error()))
	}

	err = writeDataToDb(db, key, proxy)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed write interface configuration to Etcd - "+err.Error()))
	}
}

func writeIPScanneConfigToDb(db keyval.ProtoTxn, key string, json string) {
	var err error

	ipscanner := &l3.IPScanNeighbor{}

	err = json2.Unmarshal([]byte(json), ipscanner)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed convert ipscanner json format to protobuf - "+err.Error()))
	}

	err = writeDataToDb(db, key, ipscanner)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed write interface configuration to Etcd - "+err.Error()))
	}
}

func writeNATConfigToDb(db keyval.ProtoTxn, key string, json string) {
	var err error

	natg := &nat.Nat44Global{}

	err = json2.Unmarshal([]byte(json), natg)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed convert nat json format to protobuf - "+err.Error()))
	}

	err = writeDataToDb(db, key, natg)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed write interface configuration to Etcd - "+err.Error()))
	}
}

func writeDNATConfigToDb(db keyval.ProtoTxn, key string, json string) {
	var err error

	dnat := &nat.DNat44{}

	err = json2.Unmarshal([]byte(json), dnat)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed convert dnat json format to protobuf - "+err.Error()))
	}

	err = writeDataToDb(db, key, dnat)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed write interface configuration to Etcd - "+err.Error()))
	}
}

func writeIPSecPolicyConfigToDb(db keyval.ProtoTxn, key string, json string) {
	var err error

	policy := &ipsec.SecurityPolicyDatabase{}

	err = json2.Unmarshal([]byte(json), policy)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed convert policy json format to protobuf - "+err.Error()))
	}

	err = writeDataToDb(db, key, policy)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed write policy configuration to Etcd - "+err.Error()))
	}
}

func writeIPSecAssociateConfigToDb(db keyval.ProtoTxn, key string, json string) {
	var err error

	ipa := &ipsec.SecurityAssociation{}

	err = json2.Unmarshal([]byte(json), ipa)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed convert ipsec association json format to protobuf - "+err.Error()))
	}

	err = writeDataToDb(db, key, ipa)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed write ipsec association configuration to Etcd - "+err.Error()))
	}
}

func writeIPRedirectConfigToDb(db keyval.ProtoTxn, key string, json string) {
	var err error

	ipa := &vpp_punt.IPRedirect{}

	err = json2.Unmarshal([]byte(json), ipa)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed convert ip redirect json format to protobuf - "+err.Error()))
	}

	err = writeDataToDb(db, key, ipa)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed write ip redirect configuration to Etcd - "+err.Error()))
	}
}

func writeToHostConfigToDb(db keyval.ProtoTxn, key string, json string) {
	var err error

	ipa := &vpp_punt.ToHost{}

	err = json2.Unmarshal([]byte(json), ipa)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed convert tohost json format to protobuf - "+err.Error()))
	}

	err = writeDataToDb(db, key, ipa)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed write tohost configuration to Etcd - "+err.Error()))
	}
}

func writelInterfaceConfigToDb(db keyval.ProtoTxn, key string, json string) {
	var err error

	lint := &linterface.Interface{}

	err = json2.Unmarshal([]byte(json), lint)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed convert linux interface json format to protobuf - "+err.Error()))
	}

	err = writeDataToDb(db, key, lint)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed write linux interface configuration to Etcd - "+err.Error()))
	}
}

func writelARPConfigToDb(db keyval.ProtoTxn, key string, json string) {
	var err error

	arp := &ll3.ARPEntry{}

	err = json2.Unmarshal([]byte(json), arp)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed convert linux arp json format to protobuf - "+err.Error()))
	}

	err = writeDataToDb(db, key, arp)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed write linux arp configuration to Etcd - "+err.Error()))
	}
}

func writelRouteConfigToDb(db keyval.ProtoTxn, key string, json string) {
	var err error

	route := &ll3.Route{}

	err = json2.Unmarshal([]byte(json), route)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed convert linux route json format to protobuf - "+err.Error()))
	}

	err = writeDataToDb(db, key, route)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed write linux route configuration to Etcd - "+err.Error()))
	}
}

func writeDataToDb(db keyval.ProtoTxn, key string, obj proto.Message) error {
	tx := db.Put(key, obj)

	err := tx.Commit(context.Background())

	return err
}

func DelDataFromDb(db keyval.ProtoTxn, key string) error {
	tx := db.Delete(key)

	err := tx.Commit(context.Background())

	return err
}
