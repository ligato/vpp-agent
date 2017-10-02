package acl

import (
	"errors"
	govpp "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
)

// ConfigACL executes VPP binary API "acl_add_replace".
func ConfigACL(req *ACLAddReplace, ch *govpp.Channel, log logging.Logger) error {
	reply := &ACLAddReplaceReply{}
	err := ch.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}

	log.Debugf("ACL/REPLACE VPP response: %+v\n", reply)
	if reply.Retval != 0 {
		return errors.New("VPP returned not success")
	}

	return nil
}

// DumpACL executes VPP binary API "acl_details".
func DumpACL(ch *govpp.Channel) (*ACLDetails, error) {
	reply := &ACLDetails{}
	err := ch.SendRequest(&ACLDump{}).ReceiveReply(reply)
	if err != nil {
		return nil, err
	}

	return reply, nil
}

// DelACL executes VPP binary API "acl_del".
func DelACL(req *ACLDel, ch *govpp.Channel, log logging.Logger) error {
	reply := &ACLDelReply{}
	err := ch.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}

	log.Debugf("ACL DEL VPP response: %+v\n", reply)
	if reply.Retval != 0 {
		return errors.New("VPP returned not success")
	}

	return nil
}

// ConfigIfACL executes VPP binary API "acl_interface_set_acl_list".
func ConfigIfACL(ifACL *ACLInterfaceSetACLList, ch *govpp.Channel, log logging.Logger) error {
	reply := &ACLInterfaceSetACLListReply{}
	err := ch.SendRequest(ifACL).ReceiveReply(reply)
	if err != nil {
		return err
	}

	log.Debugf("ACL Interface Set VPP response: %+v\n", reply)
	if reply.Retval != 0 {
		return errors.New("VPP returned not success")
	}

	return nil
}

// DelIfACL executes VPP binary API "acl_interface_set_acl_list" to clear
// the list of ACLs associated with a given interface.
func DelIfACL(req *ACLInterfaceSetACLList, ch *govpp.Channel, log logging.Logger) error {
	reply := &ACLInterfaceSetACLListReply{}
	err := ch.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}

	log.Debugf("ACL DEL VPP response: %+v\n", reply)
	if reply.Retval != 0 {
		return errors.New("VPP returned not success")
	}

	return nil
}
