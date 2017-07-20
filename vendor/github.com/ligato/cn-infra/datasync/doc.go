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

// Package datasync defines the interfaces for the abstraction of a data
// transport between app plugins and backend data sources/sinks (such
// as data stores, message buses, or gRPC-connected clients).
//
// The data transport APIs are centered around watching & publishing
// events. For a VPP-based VNF-Agent these events can be:
//  1. Northbound configuration related events (e.g. creation/update/deletion
//     of an interface, a bridge domain or a route)
//  2. Operational state related events (e.g. link a up/down event or a
//     counter update)
//
// These events are processed asynchronously and a feedback is given
// to the user of this API (e.g. successful configuration or an error).
//
// This APIs defines two types of events that a plugin must be able to
// process:
//  1. Data resynchronization (resync) event is defined to trigger
//     resynchronization of the whole configuration. This event is used
//     after agent start/restart, or when agent's connectivity to an
//     external data source/sink is lost and restored.
//  2. Data change event is defined to trigger incremental processing of
//     configuration changes. Data changes events are sent after the data
//     resync has completed. Each data change event contains both the
//     previous and the new/current values for the data.
//
// Data change event optimize the performance of the VPP based VNF-Agent
// in particular, because only a minimal subset of the VPP binary APIs
// needs to be called for each event. For example, only if only an IP
// address changes on a network interface, only the new IP address is
// programmed into the VPP rather than the whole network interface.
//
// Benefits of this design for VPP plugins include:
//  1. Plugins can be implemented independently from backend data
//     sources/sinks, such as ETCD, Kafka, REST, or gRPC
//  2. Plugins can work asynchronously from each other
//
// See the implementation of transports in etcdmux & httpmux for more
// details.
//
// While a separate interface definition increases complexity somewhat,
// it constitutes a standard way to use different data transport
// implementations and to provide for unit testability.
//
// The workflow is as follows:
// 1. Infrastructure plugins first register their respective transports.
// 2. VPP plugins get an instance of a transport and are able to WatchData
//    or to PublishData (see the GO interfaces)
// 3. The transport propagates to VPP plugins:
//    - configuration changes
//    - or the snapshot of the whole configuration when the data
//      resynchronization is required (see the resync plugin)
// 4. VPP plugins listen to the data change channel and the data resync
//    channel.
// 5. Best practice is that VPP plugins publish selected operational
//    state using the PublishData GO interface (for example link up/down,
//    statistics, plugin status, etc.)
package datasync
