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

// Package resync integrates the data synchronization in the lifecycle of the Agent:
//  1. Data resynchronization (resync) event after agent start/restart, or when agent's connectivity to
//     an external data source/sink is lost and restored;
//  2. Init(), Close() of the datasync transport
package resync
