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

package keyval

// Root denotes that no prefix is prepended to the keys.
const Root = ""

// KvPlugin provides unifying interface for different key-value datastore implementations.
type KvPlugin interface {
	// NewPrefixedBroker returns a ProtoBroker instance that prepends given keyPrefix to all keys in its calls. To avoid
	// using a prefix pass keyval.Root constant as argument.
	NewBroker(keyPrefix string) ProtoBroker
	// NewPrefixedWatcher returns a ProtoWatcher instance. Given key prefix is prepended to keys during watch subscribe phase.
	// The prefix is removed from the key retrieved by GetKey() in ProtoWatchResp. To avoid  using a prefix pass keyval.Root constant as argument.
	NewWatcher(keyPrefix string) ProtoWatcher
}
