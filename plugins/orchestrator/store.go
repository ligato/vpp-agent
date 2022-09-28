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
	"sort"

	"google.golang.org/protobuf/proto"
)

// KVStore describes an interface for key-value store used by dispatcher.
type KVStore interface {
	ListAll() KVPairs
	List(dataSrc string) KVPairs
	Update(dataSrc, key string, val proto.Message)
	Delete(dataSrc, key string)
	Reset(dataSrc string)
}

// KLStore describes an interface for key-label store used by dispatcher.
type KLStore interface {
	ListLabels(key string) Labels
	AddLabel(key, lkey, lval string)
	HasLabel(key, lkey string) bool
	DeleteLabel(key, lkey string)
	ResetLabels(key string)
}

type Store interface {
	KLStore
	KVStore
}

// memStore is KStore implementation that stores data in memory.
type memStore struct {
	db  map[string]KVPairs
	ldb map[string]Labels
}

func newMemStore() *memStore {
	return &memStore{
		db:  make(map[string]KVPairs),
		ldb: make(map[string]Labels),
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

func (s *memStore) ListLabels(key string) Labels {
	labels := make(Labels, len(s.ldb[key]))
	for lkey, lval := range s.ldb[key] {
		labels[lkey] = lval
	}
	return labels
}

func (s *memStore) AddLabel(key, lkey, lval string) {
	if _, ok := s.ldb[key]; !ok {
		s.ldb[key] = make(Labels)
	}
	s.ldb[key][lkey] = lval
}

func (s *memStore) HasLabel(key, lkey string) bool {
	_, ok := s.ldb[key][lkey]
	return ok
}

func (s *memStore) DeleteLabel(key, lkey string) {
	delete(s.ldb[key], lkey)
}

func (s *memStore) ResetLabels(key string) {
	delete(s.ldb, key)
}
