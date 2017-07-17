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

package mem

import (
	"github.com/ligato/cn-infra/idxmap"
	"github.com/ligato/cn-infra/logging/logroot"
	"github.com/onsi/gomega"
	"testing"
	"time"
)

func TestNewNamedMappingMem(t *testing.T) {
	gomega.RegisterTestingT(t)
	title := "Title"
	mapping := NewNamedMapping(logroot.Logger(), "owner", title, nil)
	returnedTitle := mapping.GetRegistryTitle()
	gomega.Expect(returnedTitle).To(gomega.BeEquivalentTo(title))

	names := mapping.ListNames()
	gomega.Expect(names).To(gomega.BeNil())
}

func TestCrudOps(t *testing.T) {
	gomega.RegisterTestingT(t)
	mapping := NewNamedMapping(logroot.Logger(), "owner", "title", nil)

	mapping.RegisterName("Name1", "value1")
	meta, found := mapping.Lookup("Name1")
	gomega.Expect(found).To(gomega.BeTrue())
	gomega.Expect(meta).To(gomega.BeEquivalentTo("value1"))

	mapping.RegisterName("Name2", "value2")
	meta, found = mapping.Lookup("Name2")
	gomega.Expect(found).To(gomega.BeTrue())
	gomega.Expect(meta).To(gomega.BeEquivalentTo("value2"))

	mapping.RegisterName("Name3", "value3")
	meta, found = mapping.Lookup("Name3")
	gomega.Expect(found).To(gomega.BeTrue())
	gomega.Expect(meta).To(gomega.BeEquivalentTo("value3"))

	names := mapping.ListNames()
	gomega.Expect(names).To(gomega.ContainElement("Name1"))
	gomega.Expect(names).To(gomega.ContainElement("Name2"))
	gomega.Expect(names).To(gomega.ContainElement("Name3"))

	meta, found = mapping.UnregisterName("Name2")
	gomega.Expect(found).To(gomega.BeTrue())
	gomega.Expect(meta).To(gomega.BeEquivalentTo("value2"))

	meta, found = mapping.Lookup("Name2")
	gomega.Expect(found).To(gomega.BeFalse())
	gomega.Expect(meta).To(gomega.BeNil())

	meta, found = mapping.UnregisterName("Unknown")
	gomega.Expect(found).To(gomega.BeFalse())
	gomega.Expect(meta).To(gomega.BeNil())
}

func TestSecondaryIndexes(t *testing.T) {
	gomega.RegisterTestingT(t)
	const secondaryIx = "secondary"
	mapping := NewNamedMapping(logroot.Logger(), "owner", "title", func(meta interface{}) map[string][]string {
		res := map[string][]string{}
		if str, ok := meta.(string); ok {
			res[secondaryIx] = []string{str}
		}
		return res
	})

	mapping.RegisterName("Name1", "value")
	meta, found := mapping.Lookup("Name1")
	gomega.Expect(found).To(gomega.BeTrue())
	gomega.Expect(meta).To(gomega.BeEquivalentTo("value"))

	mapping.RegisterName("Name2", "value")
	meta, found = mapping.Lookup("Name2")
	gomega.Expect(found).To(gomega.BeTrue())
	gomega.Expect(meta).To(gomega.BeEquivalentTo("value"))

	mapping.RegisterName("Name3", "different")
	meta, found = mapping.Lookup("Name3")
	gomega.Expect(found).To(gomega.BeTrue())
	gomega.Expect(meta).To(gomega.BeEquivalentTo("different"))

	names := mapping.LookupByMetadata(secondaryIx, "value")
	gomega.Expect(names).To(gomega.ContainElement("Name1"))
	gomega.Expect(names).To(gomega.ContainElement("Name2"))

	names = mapping.LookupByMetadata(secondaryIx, "unknown")
	gomega.Expect(names).To(gomega.BeNil())
	names = mapping.LookupByMetadata("Unknown index", "value")
	gomega.Expect(names).To(gomega.BeNil())

	mapping.RegisterName("Name2", "different")
	names = mapping.LookupByMetadata(secondaryIx, "different")
	gomega.Expect(names).To(gomega.ContainElement("Name2"))
	gomega.Expect(names).To(gomega.ContainElement("Name3"))

}

func TestNotifications(t *testing.T) {
	gomega.RegisterTestingT(t)
	mapping := NewNamedMapping(logroot.Logger(), "owner", "title", nil)

	ch := make(chan idxmap.NamedMappingDto, 10)
	err := mapping.Watch("subscriber", ToChan(ch))
	gomega.Expect(err).To(gomega.BeNil())

	mapping.RegisterName("Name1", "value")
	meta, found := mapping.Lookup("Name1")
	gomega.Expect(found).To(gomega.BeTrue())
	gomega.Expect(meta).To(gomega.BeEquivalentTo("value"))

	select {
	case notif := <-ch:
		gomega.Expect(notif.RegistryTitle).To(gomega.BeEquivalentTo("title"))
		gomega.Expect(notif.Del).To(gomega.BeFalse())
		gomega.Expect(notif.Name).To(gomega.BeEquivalentTo("Name1"))
		gomega.Expect(notif.Metadata).To(gomega.BeEquivalentTo("value"))
	case <-time.After(time.Second):
		t.FailNow()
	}

	mapping.RegisterName("Name1", "modified")
	meta, found = mapping.Lookup("Name1")
	gomega.Expect(found).To(gomega.BeTrue())
	gomega.Expect(meta).To(gomega.BeEquivalentTo("modified"))

	select {
	case notif := <-ch:
		gomega.Expect(notif.RegistryTitle).To(gomega.BeEquivalentTo("title"))
		gomega.Expect(notif.Del).To(gomega.BeFalse())
		gomega.Expect(notif.Name).To(gomega.BeEquivalentTo("Name1"))
		gomega.Expect(notif.Metadata).To(gomega.BeEquivalentTo("modified"))
	case <-time.After(time.Second):
		t.FailNow()
	}

	mapping.UnregisterName("Name1")
	meta, found = mapping.Lookup("Name1")
	gomega.Expect(found).To(gomega.BeFalse())
	gomega.Expect(meta).To(gomega.BeNil())

	select {
	case notif := <-ch:
		gomega.Expect(notif.RegistryTitle).To(gomega.BeEquivalentTo("title"))
		gomega.Expect(notif.Del).To(gomega.BeTrue())
		gomega.Expect(notif.Name).To(gomega.BeEquivalentTo("Name1"))
		gomega.Expect(notif.Metadata).To(gomega.BeEquivalentTo("modified"))
	case <-time.After(time.Second):
		t.FailNow()
	}

	close(ch)
}
