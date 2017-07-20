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

// Package resynceventimpl implements the interfaces (events, registration) of parent package (see it's comments)
//
// Intent: implementation is separated from the API and is meant to be used mainly internally. There are exported
// methods that are not in the api interfaces (can be changed without any impact on the users of the API).
package resynceventimpl
