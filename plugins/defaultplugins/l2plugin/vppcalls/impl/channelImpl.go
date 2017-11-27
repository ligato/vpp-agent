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

package impl

import (
	govppapi "git.fd.io/govpp.git/api"
)
//ChannelWrapperImpl is structure which wrap vpp Channel structure and implements ChannelIntf
type ChannelWrapperImpl struct {
	vppChan *govppapi.Channel
}

//SendRequest is method from ChannelIntf which is implemented as call namesake method of vpp Channel structure
func (ch *ChannelWrapperImpl) SendRequest(msg govppapi.Message) *govppapi.RequestCtx {
	return ch.vppChan.SendRequest(msg)
}