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

// Package statuscheck collects health status from other plugins through
// a plugin status report API and exposes the collected overall health
// status to the outside world via ETCD and a REST API.
//
// The API provides only two functions, one for registering the plugin for status change reporting and one
// for reporting status changes.
//
// To register a plugin for providing status reports, use Register() function:
//	statuscheck.Register(PluginID, probe)
//
// If probe is not nil, statuscheck will periodically probe the plugin state through the provided function.
// Otherwise, it is expected that the plugin itself will report state updates through ReportStateChange():
//	statuscheck.ReportStateChange(PluginID, statuscheck.OK, nil)
//
// The default status of a plugin after registering is Init.
package statuscheck
