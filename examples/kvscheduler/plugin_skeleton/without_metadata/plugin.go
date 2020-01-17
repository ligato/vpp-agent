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

/*******************************************************************************
 Feel free to use this skeleton of a VPP-Agent plugin with a single descriptor
 (variant without metadata) to build your own plugin in 5 steps:
  1. for each of your value types (e.g. interface, route, ...), define a separate
     protobuf message to carry the value data:
      - use the attached bare value definition, provided in model/model.proto
        as "Value", to start from
      - rename the proto message and the file appropriately, such that it is
        clear to which value type it belongs
      - for each of your value types, write a separate proto file
      - replicate and update the go:generate comment below (*) for each of your
        proto messages
      - when done, re-generate the golang code for the proto messages with
        `go generate .`
  2. define model for each of your value types in model/keys.go, which will be
     then used by the agent and the descriptor to build and parse keys for value
     instances
  3. for each of your value types, including derived types (with the exception
     of derived properties), implement one descriptor in the "descriptor"
     sub-package:
      - first replicate and update the go:generate comment below (**) to generate
        adapter for every descriptor to be implemented
      - generate new/updated adapters with `go generate .`
        (don't forget to remove skeleton.go adapter attached as an example)
      - use the attached skeleton descriptor to start from
      - please rename the descriptor(s) and the file(s), such that it is clear
        to which value type each of them belongs
      - implement CRUD methods and other attributes of the descriptor(s)
      - please remove callbacks/attributes which are not needed/relevant, or
        for which the default behaviour is sufficient
   4. register all your descriptors with KVScheduler in the Init method of the
      plugin
   5. finally, remove comments whose purpose was solely to guide you through
      the process of implementing a new plugin - like this one
      - make sure the remaining comments have no mention of the word "skeleton"

Beware: Extensive copy-pasting is actually a bad practise, so use the skeleton
        with caution and eventually learn how to write your own plugins from the
        scratch, using the skeleton only as a reference.
*******************************************************************************/

// (*) generate golang code from your protobuf models here:
//go:generate protoc --proto_path=. --go_out=paths=source_relative:. model/model.proto

// (**) generate adapter(s) for your descriptor(s) here:
//go:generate descriptor-adapter --descriptor-name Skeleton --value-type *model.ValueSkeleton --import "go.ligato.io/vpp-agent/v3/examples/kvscheduler/plugin_skeleton/without_metadata/model" --output-dir "descriptor"

package plugin

import (
	"github.com/ligato/cn-infra/infra"

	"go.ligato.io/vpp-agent/v3/examples/kvscheduler/plugin_skeleton/without_metadata/descriptor"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
)

// SkeletonPlugin is a plugin skeleton that you can start building your own plugins
// from.
type SkeletonPlugin struct {
	Deps
}

// Deps lists dependencies of the mock interface plugin.
type Deps struct {
	infra.PluginDeps

	// the plugin depends on KVScheduler because it needs to register the descriptor(s)
	KVScheduler kvs.KVScheduler
}

// Init method usually:
//  - loads configuration from a file (if any)
//  - registers descriptors for all objects the plugin implements
//  - potentially starts go routine to watch for some asynchronous events
//    (from which usually sends notifications to KVScheduler via PushSBNotification)
//  - etc.
func (p *SkeletonPlugin) Init() error {
	var err error

	// init & register descriptor(s) here:
	skeletonDescriptor := descriptor.NewSkeletonDescriptor(p.Log)
	err = p.KVScheduler.RegisterKVDescriptor(skeletonDescriptor)
	if err != nil {
		return err
	}

	return nil
}

// Close method usually:
//  - stops all the associated go routines (if any)
//  - closes channels, registrations, etc..
// Note: it is not needed to un-register descriptors - there is no method for
//       that anyway
func (p *SkeletonPlugin) Close() error {
	return nil
}
