package acl

// KeyPrefix returns the prefix used in ETCD to store vpp ACLs config
func KeyPrefix() string {
	return "vpp/config/v1/acl/"
}

// Key returns the prefix used in ETCD to store vpp ACL config
// of a particular ACL in selected vpp instance
func Key(aclName string) string {
	return "vpp/config/v1/acl/" + aclName
}
