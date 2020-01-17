// Copyright (c) 2019 Cisco and/or its affiliates.
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

package model

import (
	"go.ligato.io/vpp-agent/v3/pkg/models"
)

// ModuleName is the module name used for all the models of this plugin.
const ModuleName = "skeleton"

var (
	ValueModel = models.Register(&ValueSkeleton{}, models.Spec{
		Module:  ModuleName,
		Version: "v1",
		Type:    "skeleton-value",
	})
)

// ValueSkeletonKey returns the key used in NB DB to store the configuration
// of a skeleton value with the given logical name.
func ValueSkeletonKey(name string) string {
	return models.Key(&ValueSkeleton{
		Name: name,
	})
}
