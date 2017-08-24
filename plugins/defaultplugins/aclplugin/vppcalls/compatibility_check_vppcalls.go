// Copyright (c) 2017 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vppcalls

import (
	govppapi "git.fd.io/govpp.git/api"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/bin_api/acl"
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
		log.DefaultLogger().Error(err)
		return err
	}
	return nil
}
