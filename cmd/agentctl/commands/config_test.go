//  Copyright (c) 2020 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package commands

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"go.ligato.io/vpp-agent/v3/proto/ligato/configurator"
	"go.ligato.io/vpp-agent/v3/proto/ligato/vpp"
)

func Test_prepareNotifyFilters(t *testing.T) {
	type args struct {
		filters []string
	}
	tests := []struct {
		name string
		args args
		want []*configurator.Notification
	}{
		{
			name: "vpp-notification",
			args: args{
				filters: []string{`{"vpp_notification":{}}`},
			},
			want: []*configurator.Notification{
				{Notification: &configurator.Notification_VppNotification{VppNotification: &vpp.Notification{}}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := prepareNotifyFilters(tt.args.filters)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("prepareNotifyFilters() = %v, want %v", got, tt.want)
			}
			for i, n := range got {
				if !proto.Equal(n, tt.want[i]) {
					t.Errorf("prepareNotifyFilters()[%d] = %v, want %v", i, n, tt.want[i])
				}
			}

		})
	}
}
