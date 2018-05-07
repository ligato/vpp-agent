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

package rpc

import (
	"context"

	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/interfaces"
)

// GRPCService allows to send VPP notifications to external GRPC endpoints
type GRPCService interface {

	// sendNotification allows to send VPP notifications/statistic data. All the logic about
	// endpoint registration and connection is done in grpc plugin using appropriate configuration
	// file
	// todo make type independent
	SendNotification(ctx context.Context, notification *interfaces.InterfaceNotification)
}
