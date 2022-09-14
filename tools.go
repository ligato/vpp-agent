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

// +build tools

// Manage tool dependencies using Go modules.
//
//  https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
//  https://github.com/golang/go/issues/25922
//
package vppagent

import (
	_ "go.fd.io/govpp/cmd/binapi-generator"
)
