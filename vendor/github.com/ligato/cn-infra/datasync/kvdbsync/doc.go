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

// Package kvdbsync implements a key-value data store client and server
// satisfying the datasync API. The datasync API contains the Data Broker
// & KeyValProtoWatcher APIs, which are basically just facades in front of
// different key-value or SQL stores.
//
// A key-value store is used as a transport channel between a remote client
// and the agent(server). It stores data/configuration for multiple agents
// (servers). Therefore, a client only needs to know the address of the
// key-value store but not the addresses of individual agents. The client
// can write data/configuration independently of the agent's (server's)
// lifecycle.
//
// The Data KeyValProtoWatcher is used during regular operation to efficiently
// propagate data/configuration changes from the key & value store to the
// agents (servers). Upon receiving a data change event, the watcher makes
// an incremental update to its data. When data resynchronization (RESYNC) is
// triggered, then the Data Broker is used to read all particular keys &
// values from the key-value store. Reading all particular keys & values is
// a more reliable but less efficient data synchronization method.
package kvdbsync
