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

// Package datasync defines the datasync API, which abstracts the data
// transport between app plugins and backend data sources. Data sources
// can be data stores, clients connected to a message bus, or remote clients
// connected to a CN-Infra app. Transport can be, for example, HTTP or gRPC.
//
// These events are processed asynchronously.
// The app plugin that watches data changes gives callback for each event
// (e.g. successful configuration or an error).
//
// See the examples under the dedicated examples package.
package datasync
