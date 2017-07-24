package bfd

// BfdSessionPrefix bfd-session/
const BfdSessionPrefix = "vpp/config/v1/bfd/session/"

// BfdAuthKeysPrefix bfd-key/
const BfdAuthKeysPrefix = "vpp/config/v1/bfd/auth-key/"

// BfdEchoFunctionPrefix bfd-echo-function/
const BfdEchoFunctionPrefix = "vpp/config/v1/bfd/echo-function"

// SessionKeyPrefix returns the prefix used in ETCD to store vpp bfd config
func SessionKeyPrefix() string {
	return BfdSessionPrefix
}

// AuthKeysKeyPrefix returns the prefix used in ETCD to store vpp bfd config
func AuthKeysKeyPrefix() string {
	return BfdAuthKeysPrefix
}

// EchoFunctionKeyPrefix returns the prefix used in ETCD to store vpp bfd config
func EchoFunctionKeyPrefix() string {
	return BfdEchoFunctionPrefix
}

// SessionKey returns the prefix used in ETCD to store vpp bfd config
// of particular bfd session in selected vpp instance
func SessionKey(bfdSessionIfaceLabel string) string {
	return BfdSessionPrefix + bfdSessionIfaceLabel
}

// AuthKeysKey returns the prefix used in ETCD to store vpp bfd config
// of particular bfd key in selected vpp instance
func AuthKeysKey(bfdKeyIDLabel string) string {
	return BfdAuthKeysPrefix + bfdKeyIDLabel
}

// EchoFunctionKey returns the prefix used in ETCD to store vpp bfd config
// of particular bfd echo function in selected vpp instance
func EchoFunctionKey(bfdEchoIfaceLabel string) string {
	return BfdEchoFunctionPrefix + bfdEchoIfaceLabel
}
