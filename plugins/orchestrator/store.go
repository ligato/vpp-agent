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

package orchestrator

import (
	"context"
	"sort"

	"github.com/gogo/protobuf/proto"
)

// KVStore describes an interface for key-value store used by dispatcher.
type KVStore interface {
	ListAll() KVPairs
	List(dataSrc string) KVPairs
	Update(dataSrc, key string, val proto.Message)
	Delete(dataSrc, key string)
	Reset(dataSrc string)
}

// memStore is KVStore implementation that stores data in memory.
type memStore struct {
	db map[string]KVPairs
}

func newMemStore() *memStore {
	return &memStore{
		db: make(map[string]KVPairs),
	}
}

// List lists all key-value pairs.
func (s *memStore) ListAll() KVPairs {
	var dataSrcs []string
	for dataSrc := range s.db {
		dataSrcs = append(dataSrcs, dataSrc)
	}
	sort.Strings(dataSrcs)
	pairs := make(KVPairs)
	for _, dataSrc := range dataSrcs {
		for k, v := range s.List(dataSrc) {
			pairs[k] = v
		}
	}
	return pairs
}

// List lists actual key-value pairs.
func (s *memStore) List(dataSrc string) KVPairs {
	pairs := make(KVPairs, len(s.db[dataSrc]))
	for k, v := range s.db[dataSrc] {
		pairs[k] = v
	}
	return pairs
}

// Update updates value stored under key with given value.
func (s *memStore) Update(dataSrc, key string, val proto.Message) {
	if _, ok := s.db[dataSrc]; !ok {
		s.db[dataSrc] = make(KVPairs)
	}
	s.db[dataSrc][key] = val
}

// Delete deletes value stored under given key.
func (s *memStore) Delete(dataSrc, key string) {
	delete(s.db[dataSrc], key)
}

// Reset clears all key-value data.
func (s *memStore) Reset(dataSrc string) {
	delete(s.db, dataSrc)
}

type dataSrcKeyT string

var dataSrcKey = dataSrcKeyT("dataSrc")

func DataSrcContext(ctx context.Context, dataSrc string) context.Context {
	return context.WithValue(ctx, dataSrcKey, dataSrc)
}

func DataSrcFromContext(ctx context.Context) (dataSrc string, ok bool) {
	dataSrc, ok = ctx.Value(dataSrcKey).(string)
	return
}
