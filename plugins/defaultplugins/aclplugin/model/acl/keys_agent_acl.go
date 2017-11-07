package acl

const aclPrefix = "vpp/config/v1/acl/"

// KeyPrefix returns the prefix used in ETCD to store vpp ACLs config
func KeyPrefix() string {
	return aclPrefix
}

// Key returns the prefix used in ETCD to store vpp ACL config
// of a particular ACL in selected vpp instance
func Key(aclName string) string {
	return aclPrefix + aclName
}
