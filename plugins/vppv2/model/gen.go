//  Copyright (c) 2018 Cisco and/or its affiliates.
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

//go:generate protoc --proto_path=acl --gogo_out=acl acl/acl.proto
//go:generate protoc --proto_path=interfaces --gogo_out=interfaces interfaces/dhcp.proto
//go:generate protoc --proto_path=interfaces --gogo_out=interfaces interfaces/interface.proto
//go:generate protoc --proto_path=interfaces --gogo_out=interfaces interfaces/state.proto
//go:generate protoc --proto_path=l2 --gogo_out=l2 l2/bd.proto
//go:generate protoc --proto_path=l2 --gogo_out=l2 l2/fib.proto
//go:generate protoc --proto_path=l2 --gogo_out=l2 l2/xconnect.proto
//go:generate protoc --proto_path=l3 --gogo_out=l3 l3/arp.proto
//go:generate protoc --proto_path=l3 --gogo_out=l3 l3/static_route.proto
//go:generate protoc --proto_path=nat --gogo_out=nat nat/nat.proto

package model
