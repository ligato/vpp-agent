package stn

// StnRulesPrefix stn key/
const StnRulesPrefix = "vpp/config/v1/stn/rules/"

// KeyPrefix returns the prefix used in ETCD to store vpp STN config
func KeyPrefix() string {
	return StnRulesPrefix
}

// Key returns the prefix used in ETCD to store vpp STN config
// of a particular rule in selected vpp instance.
func Key(ruleName string) string {
	return StnRulesPrefix + ruleName
}
