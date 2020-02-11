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

package abfidx_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/abfplugin/abfidx"
)

func TestABFIndexLookupByName(t *testing.T) {
	RegisterTestingT(t)
	abfIndex := abfidx.NewABFIndex(logging.DefaultLogger, "abf-index")

	abfIndex.Put("val1", &abfidx.ABFMetadata{Index: 10})
	abfIndex.Put("val2", &abfidx.ABFMetadata{Index: 20})
	abfIndex.Put("val3", 10)

	metadata, exists := abfIndex.LookupByName("val1")
	Expect(exists).To(BeTrue())
	Expect(metadata).ToNot(BeNil())
	Expect(metadata.Index).To(Equal(uint32(10)))

	metadata, exists = abfIndex.LookupByName("val2")
	Expect(exists).To(BeTrue())
	Expect(metadata).ToNot(BeNil())
	Expect(metadata.Index).To(Equal(uint32(20)))

	metadata, exists = abfIndex.LookupByName("val3")
	Expect(exists).To(BeFalse())
	Expect(metadata).To(BeNil())

	metadata, exists = abfIndex.LookupByName("val4")
	Expect(exists).To(BeFalse())
	Expect(metadata).To(BeNil())
}

func TestABFIndexLookupByIndex(t *testing.T) {
	RegisterTestingT(t)
	abfIndex := abfidx.NewABFIndex(logging.DefaultLogger, "abf-index")

	abfIndex.Put("val1", &abfidx.ABFMetadata{Index: 10})
	abfIndex.Put("val2", &abfidx.ABFMetadata{Index: 20})

	name, metadata, exists := abfIndex.LookupByIndex(10)
	Expect(exists).To(BeTrue())
	Expect(name).To(Equal("val1"))
	Expect(metadata).ToNot(BeNil())
	Expect(metadata.Index).To(Equal(uint32(10)))

	name, metadata, exists = abfIndex.LookupByIndex(20)
	Expect(exists).To(BeTrue())
	Expect(name).To(Equal("val2"))
	Expect(metadata).ToNot(BeNil())
	Expect(metadata.Index).To(Equal(uint32(20)))

	name, metadata, exists = abfIndex.LookupByIndex(30)
	Expect(exists).To(BeFalse())
	Expect(name).To(Equal(""))
	Expect(metadata).To(BeNil())
}
