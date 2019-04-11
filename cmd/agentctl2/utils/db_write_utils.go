package utils

import (
	"context"

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
	_, dataType, _ := ParseKey(key)

	switch dataType {
	case acl.ModelACL.KeyPrefix():
		writeACLConfigToDb(db, key, json)
	case interfaces.ModelInterface.KeyPrefix():
		writeInterfaceConfigToDb(db, key, json)
	case l2.ModelBridgeDomain.KeyPrefix():
		writeBridgeDomainConfigToDb(db, key, json)
	case l2.ModelFIBEntry.KeyPrefix():
		writeFibTableConfigToDb(db, key, json)
	case l2.ModelXConnectPair.KeyPrefix():
		writeXConnectConfigToDb(db, key, json)
	case l3.ModelARPEntry.KeyPrefix():
		writeARPConfigToDb(db, key, json)
	case l3.ModelRoute.KeyPrefix():
		writeRouteConfigToDb(db, key, json)
	case l3.ModelProxyARP.KeyPrefix():
		writeProxyConfigToDb(db, key, json)
	case l3.ModelIPScanNeighbor.KeyPrefix():
		writeIPScanneConfigToDb(db, key, json)
	case nat.ModelNat44Global.KeyPrefix():
		writeNATConfigToDb(db, key, json)
	case nat.ModelDNat44.KeyPrefix():
		writeDNATConfigToDb(db, key, json)
	case ipsec.ModelSecurityPolicyDatabase.KeyPrefix():
		writeIPSecPolicyConfigToDb(db, key, json)
	case ipsec.ModelSecurityAssociation.KeyPrefix():
		writeIPSecAssociateConfigToDb(db, key, json)
	case linterface.ModelInterface.KeyPrefix():
		writelInterfaceConfigToDb(db, key, json)
	case ll3.ModelARPEntry.KeyPrefix():
		writelARPConfigToDb(db, key, json)
	case ll3.ModelRoute.KeyPrefix():
		writelRouteConfigToDb(db, key, json)
	}
}

func writeACLConfigToDb(db keyval.ProtoTxn, key string, json string) {
	var err error

	vacl := &acl.ACL{}

	err = vacl.UnmarshalJSON([]byte(json))

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed convert acl json format to protobuf - "+err.Error()))
	}

	err = writeDataToDb(db, key, vacl)

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed write interface configuration to Etcd - "+err.Error()))
	}
}

func writeInterfaceConfigToDb(db keyval.ProtoTxn, key string, json string) {
	var err error

	intr := &interfaces.Interface{}

	err = intr.UnmarshalJSON([]byte(json))

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed convert interface json format to protobuf - "+err.Error()))
	}

	err = writeDataToDb(db, key, intr)

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed write interface configuration to Etcd - "+err.Error()))
	}
}

func writeBridgeDomainConfigToDb(db keyval.ProtoTxn, key string, json string) {
	var err error

	bridge := &l2.BridgeDomain{}

	err = bridge.UnmarshalJSON([]byte(json))

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed convert interface json format to protobuf - "+err.Error()))
	}

	err = writeDataToDb(db, key, bridge)

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed write interface configuration to Etcd - "+err.Error()))
	}
}

func writeFibTableConfigToDb(db keyval.ProtoTxn, key string, json string) {
	var err error

	fib := &l2.FIBEntry{}

	err = fib.UnmarshalJSON([]byte(json))

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed convert fib json format to protobuf - "+err.Error()))
	}

	err = writeDataToDb(db, key, fib)

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed write interface configuration to Etcd - "+err.Error()))
	}
}

func writeXConnectConfigToDb(db keyval.ProtoTxn, key string, json string) {
	var err error

	xconnect := &l2.XConnectPair{}

	err = xconnect.UnmarshalJSON([]byte(json))

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed convert xconnect json format to protobuf - "+err.Error()))
	}

	err = writeDataToDb(db, key, xconnect)

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed write interface configuration to Etcd - "+err.Error()))
	}
}

func writeARPConfigToDb(db keyval.ProtoTxn, key string, json string) {
	var err error

	arp := &l3.ARPEntry{}

	err = arp.UnmarshalJSON([]byte(json))

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed convert arp json format to protobuf - "+err.Error()))
	}

	err = writeDataToDb(db, key, arp)

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed write interface configuration to Etcd - "+err.Error()))
	}
}

func writeRouteConfigToDb(db keyval.ProtoTxn, key string, json string) {
	var err error

	route := &l3.Route{}

	err = route.UnmarshalJSON([]byte(json))

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed convert route json format to protobuf - "+err.Error()))
	}

	err = writeDataToDb(db, key, route)

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed write interface configuration to Etcd - "+err.Error()))
	}
}

func writeProxyConfigToDb(db keyval.ProtoTxn, key string, json string) {
	var err error

	proxy := &l3.ProxyARP{}

	err = proxy.UnmarshalJSON([]byte(json))

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed convert proxy json format to protobuf - "+err.Error()))
	}

	err = writeDataToDb(db, key, proxy)

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed write interface configuration to Etcd - "+err.Error()))
	}
}

func writeIPScanneConfigToDb(db keyval.ProtoTxn, key string, json string) {
	var err error

	ipscanner := &l3.IPScanNeighbor{}

	err = ipscanner.UnmarshalJSON([]byte(json))

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed convert ipscanner json format to protobuf - "+err.Error()))
	}

	err = writeDataToDb(db, key, ipscanner)

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed write interface configuration to Etcd - "+err.Error()))
	}
}

func writeNATConfigToDb(db keyval.ProtoTxn, key string, json string) {
	var err error

	natg := &nat.Nat44Global{}

	err = natg.UnmarshalJSON([]byte(json))

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed convert nat json format to protobuf - "+err.Error()))
	}

	err = writeDataToDb(db, key, natg)

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed write interface configuration to Etcd - "+err.Error()))
	}
}

func writeDNATConfigToDb(db keyval.ProtoTxn, key string, json string) {
	var err error

	dnat := &nat.DNat44{}

	err = dnat.UnmarshalJSON([]byte(json))

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed convert dnat json format to protobuf - "+err.Error()))
	}

	err = writeDataToDb(db, key, dnat)

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed write interface configuration to Etcd - "+err.Error()))
	}
}

func writeIPSecPolicyConfigToDb(db keyval.ProtoTxn, key string, json string) {
	var err error

	policy := &ipsec.SecurityPolicyDatabase{}

	err = policy.UnmarshalJSON([]byte(json))

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed convert policy json format to protobuf - "+err.Error()))
	}

	err = writeDataToDb(db, key, policy)

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed write policy configuration to Etcd - "+err.Error()))
	}
}

func writeIPSecAssociateConfigToDb(db keyval.ProtoTxn, key string, json string) {
	var err error

	ipa := &ipsec.SecurityAssociation{}

	err = ipa.UnmarshalJSON([]byte(json))

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed convert ipsec association json format to protobuf - "+err.Error()))
	}

	err = writeDataToDb(db, key, ipa)

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed write ipsec association configuration to Etcd - "+err.Error()))
	}
}

func writelInterfaceConfigToDb(db keyval.ProtoTxn, key string, json string) {
	var err error

	lint := &linterface.Interface{}

	err = lint.UnmarshalJSON([]byte(json))

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed convert linux interface json format to protobuf - "+err.Error()))
	}

	err = writeDataToDb(db, key, lint)

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed write linux interface configuration to Etcd - "+err.Error()))
	}
}

func writelARPConfigToDb(db keyval.ProtoTxn, key string, json string) {
	var err error

	arp := &ll3.ARPEntry{}

	err = arp.UnmarshalJSON([]byte(json))

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed convert linux arp json format to protobuf - "+err.Error()))
	}

	err = writeDataToDb(db, key, arp)

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed write linux arp configuration to Etcd - "+err.Error()))
	}
}

func writelRouteConfigToDb(db keyval.ProtoTxn, key string, json string) {
	var err error

	route := &ll3.Route{}

	err = route.UnmarshalJSON([]byte(json))

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed convert linux route json format to protobuf - "+err.Error()))
	}

	err = writeDataToDb(db, key, route)

	if nil != err {
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
