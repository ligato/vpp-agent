// Copyright (c) 2018 Cisco and/or its affiliates.
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

package vppdump

import (
	"time"

	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/stn"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
)

// DumpStnRules returns a list of all STN rules configured on the VPP
func DumpStnRules(vppChan vppcalls.VPPChannel, stopwatch *measure.Stopwatch) (rules []*stn.StnRulesDetails, err error) {
	defer func(t time.Time) {
		stopwatch.TimeLog(stn.StnRulesDump{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &stn.StnRulesDump{}
	reqCtx := vppChan.SendMultiRequest(req)
	for {
		msg := &stn.StnRulesDetails{}
		stop, err := reqCtx.ReceiveReply(msg)
		if stop {
			break
		}
		if err != nil {
			return nil, err
		}
		rules = append(rules, msg)
	}

	return rules, nil
}
