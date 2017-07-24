package interfaces

import (
	"strings"
	"fmt"
)

const (
	// InterfacePrefix vpp/config/v1/interface/
	InterfacePrefix = "vpp/config/v1/interface/"
	// IfStatePrefix vpp/status/v1/interface/
	IfStatePrefix = "vpp/status/v1/interface/"
	// IfStateErrorPrefix vpp/status/v1/interface/error
	IfStateErrorPrefix = "vpp/status/v1/interface/error/"
)

// InterfaceKeyPrefix returns the prefix used in ETCD to store vpp interfaces config
func InterfaceKeyPrefix() string {
	return InterfacePrefix
}

// ParseNameFromKey returns suffix of the ky
func ParseNameFromKey(key string) (name string, err error) {
	lastSlashPos := strings.LastIndex(key, "/")
	if lastSlashPos > 0 && lastSlashPos < len(key)-1 {
		return key[lastSlashPos+1:], nil
	}

	return key, fmt.Errorf("wrong format of the key %s", key)
}

// InterfaceKey returns the prefix used in ETCD to store vpp interface config
// of particular interface in selected vpp instance
func InterfaceKey(ifaceLabel string) string {
	return InterfacePrefix + ifaceLabel
}

// InterfaceErrorPrefix returns the prefix used in ETCD to store interface errors
func InterfaceErrorPrefix() string {
	return IfStateErrorPrefix
}

// InterfaceErrorKey returns the key used in ETCD to store interface errors
func InterfaceErrorKey(ifaceLabel string) string {
	return IfStateErrorPrefix + ifaceLabel
}

// InterfaceStateKeyPrefix returns the prefix used in ETCD to store vpp interfaces state data
func InterfaceStateKeyPrefix() string {
	return IfStatePrefix
}

// InterfaceStateKey returns the prefix used in ETCD to store vpp interface state data
// of particular interface in selected vpp instance
func InterfaceStateKey(ifaceLabel string) string {
	return IfStatePrefix + ifaceLabel
}