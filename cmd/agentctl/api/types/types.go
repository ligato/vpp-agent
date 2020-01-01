//  Copyright (c) 2019 Cisco and/or its affiliates.
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

package types

// ErrorResponse represents an error.
type ErrorResponse struct {
	Message string `json:"message"`
}

// ComponentVersion describes the version information for a specific component.
type ComponentVersion struct {
	Name    string
	Version string
	Details map[string]string `json:",omitempty"`
}

// Version contains response of Engine API:
// GET "/version"
type Version struct {
	Components []ComponentVersion

	Version       string
	APIVersion    string
	MinAPIVersion string
	GitCommit     string
	GoVersion     string
	Os            string
	Arch          string
	KernelVersion string
	BuildTime     string
}

// Ping contains response of Engine API:
// GET "/_ping"
type Ping struct {
	APIVersion string
	OSType     string
}

type Logger struct {
	Logger string `json:"logger,omitempty"`
	Level  string `json:"level,omitempty"`
}
