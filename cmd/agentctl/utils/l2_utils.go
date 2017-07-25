package utils

import (
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/vpp-agent/defaultplugins/l2plugin/model/l2"
	"errors"
	"fmt"
)

// Bridge domain flag names
const (
	BDName = "bridge-domain-name"
	IfName = "interface-name"
	BVI = "bvi"
	SHZ = "split-horizon-group"
	IPAddress = "ip-address"
	PhysAddress = "physical-address"
	StaticConfig = "static-config"
	IsDrop = "is-drop"
	IsDelete = "is-delete"
)

// GetBridgeDomainKeyAndValue returns true if a bridge domain with the specified name was found together with the BD key,
// data and data broker
func GetBridgeDomainKeyAndValue(endpoints []string, label string, bdName string) (bool, string, *l2.BridgeDomains_BridgeDomain, keyval.ProtoBroker) {
	validateBdIdentifiers(label, bdName)

	db, err := GetDbForOneAgent(endpoints, label)
	if err != nil {
		ExitWithError(ExitBadConnection, err)
	}

	key := l2.BridgeDomainKey(bdName)
	bd := &l2.BridgeDomains_BridgeDomain{}

	found, _, err := db.GetValue(key, bd)
	if err != nil {
		ExitWithError(ExitError, errors.New("Error getting existing config - " + err.Error()))
	}

	return found, key, bd, db
}

// GetFibEntry returns the FIB entry if exists
func GetFibEntry(endpoints []string, label string, fibMac string) (bool, string, *l2.FibTableEntries_FibTableEntry) {
	db, err := GetDbForOneAgent(endpoints, label)
	if err != nil {
		ExitWithError(ExitBadConnection, err)
	}

	key := l2.FibKeyPrefix() + fibMac
	fibEntry := &l2.FibTableEntries_FibTableEntry{}

	found, _, err := db.GetValue(key, fibEntry)
	if err != nil {
		ExitWithError(ExitError, errors.New("Error getting existing config - " + err.Error()))
	}

	return found, key, fibEntry
}

// WriteBridgeDomainToDb writes bridge domain to the ETCD
func WriteBridgeDomainToDb(db keyval.ProtoBroker, key string, bd *l2.BridgeDomains_BridgeDomain) {
	validateBridgeDomain(bd)
	db.Put(key, bd)
}

// WriteFibDataToDb writes FIB entry to the ETCD
func WriteFibDataToDb(db keyval.ProtoBroker, key string, fib *l2.FibTableEntries_FibTableEntry) {
	db.Put(key, fib)
}

// DeleteFibDataFromDb removes FIB entry from the ETCD
func DeleteFibDataFromDb(db keyval.ProtoBroker, key string) {
	db.Delete(key)
}

func validateBridgeDomain( bd *l2.BridgeDomains_BridgeDomain) {
	fmt.Printf("Validating bridge domain\n bd: %+v\n", bd)
	// todo implement
}

func validateBdIdentifiers(label string, name string) {
	if label == "" {
		ExitWithError(ExitInvalidInput, errors.New("Missing microservice label"))
	}
	if name == "" {
		ExitWithError(ExitInvalidInput, errors.New("Missing bridge domain name"))
	}
}