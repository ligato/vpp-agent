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
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l4plugin/bin_api/session"
)

// AddAppNamespace calls respective VPP binary api to configure AppNamespace
func AddAppNamespace(secret uint64, swIfIdx, ip4FibID, ip6FibID uint32, id []byte, log logging.Logger, vppChan *govppapi.Channel) error {
	log.Debugf("Adding App Namespace %v to interface %v", string(id), swIfIdx)

	req := &session.AppNamespaceAddDel{
		SwIfIndex:      swIfIdx,
		Secret:         secret,
		IP4FibID:       ip4FibID,
		IP6FibID:       ip6FibID,
		NamespaceID:    id,
		NamespaceIDLen: uint8(len(id)),
	}

	reply := &session.AppNamespaceAddDelReply{}

	err := vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		log.WithFields(logging.Fields{"Error": err, "AppNamespace": string(id)}).Error("Error while configuring AppNamespace")
		return err
	}
	if reply.Retval != 0 {
		log.WithField("Return value", reply.Retval).Error("Unexpected return value")
		return err
	}
	log.WithField("AppNamespaceID", string(id)).Debug("AppNamespace added.")

	return nil
}
