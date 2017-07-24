package vppcalls

import (
	govppapi "git.fd.io/govpp.git/api"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/defaultplugins/aclplugin/bin_api/acl"
)

// CheckMsgCompatibilityForACL checks if CRSs are compatible with VPP in runtime
func CheckMsgCompatibilityForACL(vppChannel *govppapi.Channel) error {
	msgs := []govppapi.Message{
		&acl.ACLAddReplace{},
		&acl.ACLAddReplaceReply{},
		&acl.ACLDel{},
		&acl.ACLDelReply{},
		&acl.MacipACLAdd{},
		&acl.MacipACLAddReply{},
		&acl.MacipACLDel{},
		&acl.MacipACLDelReply{},
		&acl.ACLDump{},
		&acl.ACLDetails{},
		&acl.MacipACLDump{},
		&acl.MacipACLDetails{},
		&acl.ACLInterfaceListDump{},
		&acl.ACLInterfaceListDetails{},
		&acl.ACLInterfaceSetACLList{},
		&acl.ACLInterfaceSetACLListReply{},
		&acl.MacipACLInterfaceAddDel{},
		&acl.MacipACLInterfaceAddDelReply{},
	}
	err := vppChannel.CheckMessageCompatibility(msgs...)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}
