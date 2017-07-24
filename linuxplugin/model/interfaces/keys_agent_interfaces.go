package interfaces

const (
	// InterfacePrefix is a prefix used in ETCD to store configuration for Linux interfaces.
	InterfacePrefix = "linux/config/v1/interface/"
)

// InterfaceKeyPrefix returns the prefix used in ETCD to store config for Linux interfaces
func InterfaceKeyPrefix() string {
	return InterfacePrefix
}

// InterfaceKey returns the prefix used in ETCD to store configuration of a particular Linux interface.
func InterfaceKey(ifaceLabel string) string {
	return InterfacePrefix + ifaceLabel
}
