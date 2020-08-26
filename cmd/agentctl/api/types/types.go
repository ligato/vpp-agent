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

// Version contains response of Agent REST API:
// GET "/info/version"
type Version struct {
	App       string
	Version   string
	GitCommit string
	GitBranch string

	BuildUser string
	BuildHost string
	BuildTime int64

	GoVersion string
	OS        string
	Arch      string

	APIVersion string
	OSType     string
}

type Logger struct {
	Logger string
	Level  string `json:"level,omitempty"`
}

// Model provides info about registered model.
type Model struct {
	Name         string
	Class        string
	Module       string
	Type         string
	Version      string
	KeyPrefix    string
	NameTemplate string `json:",omitempty"`
	ProtoName    string
	ProtoFile    string `json:",omitempty"`
	GoType       string `json:",omitempty"`
	PkgPath      string `json:",omitempty"`
}
