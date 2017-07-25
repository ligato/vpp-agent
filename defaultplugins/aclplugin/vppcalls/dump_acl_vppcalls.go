package vppcalls

import (
	govppapi "git.fd.io/govpp.git/api"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/defaultplugins/aclplugin/bin_api/acl"
	"github.com/ligato/vpp-agent/idxvpp"
)

// DumpInterface finds interface in VPP and returns its ACL configuration
func DumpInterface(swIndex uint32, vppChannel *govppapi.Channel) (*acl.ACLInterfaceListDetails, error) {
	req := &acl.ACLInterfaceListDump{}
	req.SwIfIndex = swIndex

	msg := &acl.ACLInterfaceListDetails{}

	err := vppChannel.SendRequest(req).ReceiveReply(msg)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

// DumpIPAcl test function
func DumpIPAcl(vppChannel *govppapi.Channel) error {
	log.Print("List of ACLs:")
	req := &acl.ACLDump{}
	req.ACLIndex = 0xffffffff
	reqContext := vppChannel.SendMultiRequest(req)
	for {
		msg := &acl.ACLDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return err
		}
		if stop {
			break
		}
		log.Printf("ACL index: %v, rule count: %v, tag: %v", msg.ACLIndex, msg.Count, string(msg.Tag[:]))

	}
	return nil
}

// DumpMacIPAcl test function
func DumpMacIPAcl(vppChannel *govppapi.Channel) error {
	req := &acl.MacipACLDump{}
	req.ACLIndex = 0xffffffff
	reqContext := vppChannel.SendMultiRequest(req)
	for {
		msg := &acl.MacipACLDump{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return err
		}
		if stop {
			break
		}
		log.Print(msg.ACLIndex)
	}
	return nil
}

// DumpInterfaces test function
func DumpInterfaces(swIndexes idxvpp.NameToIdxRW, vppChannel *govppapi.Channel) error {
	req := &acl.ACLInterfaceListDump{}
	req.SwIfIndex = 0xffffffff
	reqContext := vppChannel.SendMultiRequest(req)
	for {
		msg := &acl.ACLInterfaceListDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return err
		}
		if stop {
			break
		}
		name, _, found := swIndexes.LookupName(msg.SwIfIndex)
		if !found {
			continue
		}
		log.Printf("Interface %v is in %v acl in direction %v and applied in %v",
			name, msg.Count, msg.NInput, msg.Acls)
	}
	return nil
}
