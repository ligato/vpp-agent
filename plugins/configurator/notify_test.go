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

package configurator

import (
	"testing"

	pb "go.ligato.io/vpp-agent/v3/proto/ligato/configurator"
	"go.ligato.io/vpp-agent/v3/proto/ligato/vpp"
	vpp_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

func TestFilters(t *testing.T) {
	type args struct {
		n       *pb.Notification
		filters []*pb.Notification
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "interface name",
			args: args{
				n: &pb.Notification{
					Notification: &pb.Notification_VppNotification{
						VppNotification: &vpp.Notification{Interface: &vpp_interfaces.InterfaceNotification{
							Type: vpp_interfaces.InterfaceNotification_UPDOWN,
							State: &vpp_interfaces.InterfaceState{
								Name:         "LOOP1",
								InternalName: "loop1",
								Type:         vpp_interfaces.Interface_SOFTWARE_LOOPBACK,
								IfIndex:      1,
								AdminStatus:  vpp_interfaces.InterfaceState_UP,
								OperStatus:   vpp_interfaces.InterfaceState_UP,
							},
						}},
					},
				},
				filters: []*pb.Notification{
					{
						Notification: &pb.Notification_VppNotification{
							VppNotification: &vpp.Notification{Interface: &vpp_interfaces.InterfaceNotification{
								State: &vpp_interfaces.InterfaceState{
									Name: "LOOP1",
								},
							}},
						},
					},
				}},
			want: true,
		},
		{
			name: "interface notification type",
			args: args{
				n: &pb.Notification{
					Notification: &pb.Notification_VppNotification{
						VppNotification: &vpp.Notification{Interface: &vpp_interfaces.InterfaceNotification{
							Type: vpp_interfaces.InterfaceNotification_UPDOWN,
							State: &vpp_interfaces.InterfaceState{
								Name:         "LOOP1",
								InternalName: "loop1",
								Type:         vpp_interfaces.Interface_SOFTWARE_LOOPBACK,
								IfIndex:      1,
								AdminStatus:  vpp_interfaces.InterfaceState_UP,
								OperStatus:   vpp_interfaces.InterfaceState_UP,
							},
						}},
					},
				},
				filters: []*pb.Notification{
					{
						Notification: &pb.Notification_VppNotification{
							VppNotification: &vpp.Notification{Interface: &vpp_interfaces.InterfaceNotification{
								Type: vpp_interfaces.InterfaceNotification_UPDOWN,
							}},
						},
					},
				}},
			want: true,
		},
		{
			name: "interface name and notification type",
			args: args{
				n: &pb.Notification{
					Notification: &pb.Notification_VppNotification{
						VppNotification: &vpp.Notification{Interface: &vpp_interfaces.InterfaceNotification{
							Type: vpp_interfaces.InterfaceNotification_UPDOWN,
							State: &vpp_interfaces.InterfaceState{
								Name:         "LOOP1",
								InternalName: "loop1",
								Type:         vpp_interfaces.Interface_SOFTWARE_LOOPBACK,
								IfIndex:      1,
								AdminStatus:  vpp_interfaces.InterfaceState_UP,
								OperStatus:   vpp_interfaces.InterfaceState_UP,
							},
						}},
					},
				},
				filters: []*pb.Notification{
					{
						Notification: &pb.Notification_VppNotification{
							VppNotification: &vpp.Notification{Interface: &vpp_interfaces.InterfaceNotification{
								Type: vpp_interfaces.InterfaceNotification_UPDOWN,
								State: &vpp_interfaces.InterfaceState{
									Name: "LOOP1",
								},
							}},
						},
					},
				}},
			want: true,
		},

		{
			name: "wrong notification type",
			args: args{
				n: &pb.Notification{
					Notification: &pb.Notification_VppNotification{
						VppNotification: &vpp.Notification{Interface: &vpp_interfaces.InterfaceNotification{
							Type: vpp_interfaces.InterfaceNotification_UPDOWN,
							State: &vpp_interfaces.InterfaceState{
								Name:         "LOOP1",
								InternalName: "loop1",
								Type:         vpp_interfaces.Interface_SOFTWARE_LOOPBACK,
								IfIndex:      1,
								AdminStatus:  vpp_interfaces.InterfaceState_UP,
								OperStatus:   vpp_interfaces.InterfaceState_UP,
							},
						}},
					},
				},
				filters: []*pb.Notification{
					{
						Notification: &pb.Notification_VppNotification{
							VppNotification: &vpp.Notification{Interface: &vpp_interfaces.InterfaceNotification{
								Type: vpp_interfaces.InterfaceNotification_COUNTERS,
							}},
						},
					},
				}},
			want: false,
		},

		{
			name: "",
			args: args{
				n: &pb.Notification{
					Notification: &pb.Notification_VppNotification{
						VppNotification: &vpp.Notification{Interface: &vpp_interfaces.InterfaceNotification{
							Type: vpp_interfaces.InterfaceNotification_UPDOWN,
							State: &vpp_interfaces.InterfaceState{
								Name:         "LOOP1",
								InternalName: "loop1",
								Type:         vpp_interfaces.Interface_SOFTWARE_LOOPBACK,
								IfIndex:      1,
								AdminStatus:  vpp_interfaces.InterfaceState_UP,
								OperStatus:   vpp_interfaces.InterfaceState_UP,
							},
						}},
					},
				},
				filters: []*pb.Notification{
					{
						Notification: &pb.Notification_VppNotification{
							VppNotification: &vpp.Notification{Interface: &vpp_interfaces.InterfaceNotification{
								State: &vpp_interfaces.InterfaceState{
									Name: "LOOP1",
								},
							}},
						},
					},
				}},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isFilter(tt.args.n, tt.args.filters); got != tt.want {
				t.Errorf("isFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}
