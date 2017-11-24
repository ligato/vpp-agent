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

// Package model is the parent for packages defining various GRPC
// services generated from protobuf data models.
package model

//go:generate protoc -I. -I./../../../../../../../ vpp_changes_svc.proto --go_out=plugins=grpc:.
//go:generate protoc -I. -I./../../../../../../../ vpp_resync_svc.proto --go_out=plugins=grpc:.
