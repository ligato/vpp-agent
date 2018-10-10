// Copyright (c) 2018 Cisco and/or its affiliates.
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

package utils

// KeySet is a set of keys.
type KeySet map[string]struct{}

// NewKeySet returns a new instance of empty KeySet.
func NewKeySet(keys ...string) KeySet {
	ks := make(KeySet)
	for _, key := range keys {
		ks.Add(key)
	}
	return ks
}

// String return human-readable string representation of the ket-set.
func (ks KeySet) String() string {
	str := "{"
	idx := 0
	for key := range ks {
		str += key
		if idx < len(ks)-1 {
			str += ", "
		}
		idx++
	}
	str += "}"
	return str
}

// DeepCopy returns a deep-copy of the key set.
func (ks KeySet) DeepCopy() KeySet {
	copy := make(KeySet)
	for key := range ks {
		copy[key] = struct{}{}
	}
	return copy
}

// Has returns true if the given key is in the set.
func (ks KeySet) Has(key string) bool {
	_, has := ks[key]
	return has
}

// Add adds key into the set.
func (ks KeySet) Add(key string) KeySet {
	ks[key] = struct{}{}
	return ks
}

// Del removes key from the set.
func (ks KeySet) Del(key string) KeySet {
	delete(ks, key)
	return ks
}

// Subtract removes keys from <ks> that are in both key sets.
func (ks KeySet) Subtract(ks2 KeySet) KeySet {
	for key := range ks2 {
		delete(ks, key)
	}
	return ks
}

// Intersect returns a new key set that contains keys present in both sets.
func (ks KeySet) Intersect(ks2 KeySet) KeySet {
	intersection := NewKeySet()
	for key := range ks {
		if ks2.Has(key) {
			intersection.Add(key)
		}
	}
	return intersection
}
