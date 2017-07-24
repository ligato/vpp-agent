package l2

import (
	"strings"
	"fmt"
)

// Prefixes
const (
	// BdPrefix is the relative key prefix for bridge domains.
	BdPrefix = "vpp/config/v1/bd/"
	// BdErrPrefix is the relative key prefix for the bridge domain error
	BdErrPrefix = "vpp/config/v1/bd/error/"
	// FIBPrefix is the relative key prefix for FIB table entries.
	//TODO FIBPrefix = "vpp/config/v1/bd/<bd>/fib/"
	FIBPrefix = "vpp/config/v1/bd/fib/"
	// XconnectPrefix is the relative key prefix for xconnects.
	XconnectPrefix = "vpp/config/v1/xconnect/"
)

// BridgeDomainKeyPrefix returns the prefix used in ETCD to store vpp bridge domain config
func BridgeDomainKeyPrefix() string {
	return BdPrefix
}

// BridgeDomainKey returns the prefix used in ETCD to store vpp bridge domain config
// of particular bridge domain in selected vpp instance
func BridgeDomainKey(bdName string) string {
	return BdPrefix + bdName
}

// BridgeDomainErrorPrefix returns the prefix used in ETCD to store bridge domain errors
func BridgeDomainErrorPrefix() string {
	return BdErrPrefix
}

// BridgeDomainErrorKey returns the key used in ETCD to store bridge domain errors
func BridgeDomainErrorKey(bdLabel string) string {
	return BdErrPrefix + bdLabel
}

// ParseBDNameFromKey returns suffix of the ky
func ParseBDNameFromKey(key string) (name string, err error) {
	lastSlashPos := strings.LastIndex(key, "/")
	if lastSlashPos > 0 && lastSlashPos < len(key)-1 {
		return key[lastSlashPos+1:], nil
	}

	return key, fmt.Errorf("wrong format of the key %s", key)
}

// FibKeyPrefix returns the prefix used in ETCD to store vpp fib table entry config
func FibKeyPrefix() string {
	return FIBPrefix
}

// FibKey returns the prefix used in ETCD to store vpp fib table entry config
// of particular fib in selected vpp instance
func FibKey(fibMac string) string {
	return FIBPrefix + fibMac
}

// XConnectKeyPrefix returns the prefix used in ETCD to store vpp xConnect pair config
func XConnectKeyPrefix() string {
	return XconnectPrefix
}

// XConnectKey returns the prefix used in ETCD to store vpp xConnect pair config
// of particular xConnect pair in selected vpp instance
func XConnectKey(rxIface string) string {
	return XconnectPrefix + rxIface
}
