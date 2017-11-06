package stn

// KeyPrefix returns the prefix used in ETCD to store vpp STN config
func KeyPrefix() string {
	return "vpp/config/v1/stn/rules/"
}

// Key returns the prefix used in ETCD to store vpp STN config
// of a particular rule in selected vpp instance
func Key(ruleName string) string {
	return "vpp/config/v1/stn/rules/" + ruleName
}
