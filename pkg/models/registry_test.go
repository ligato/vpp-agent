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

package models_test

import (
	"testing"

	"github.com/golang/protobuf/proto"
	. "github.com/onsi/gomega"

	. "github.com/ligato/vpp-agent/pkg/models"
	testmodel "github.com/ligato/vpp-agent/pkg/models/testdata/testmodel"
)

func ResetDefaultRegistry() {
	DefaultRegistry = NewRegistry()
}

func TestRegister(t *testing.T) {
	g := NewGomegaWithT(t)
	ResetDefaultRegistry()

	basicModel := Register(&testmodel.Basic{}, Spec{
		Module:  "module",
		Version: "v1",
		Type:    "basic",
		Class:   "config",
	})

	registered := RegisteredModels()
	g.Expect(registered).To(HaveLen(1))

	//model := registered[0].GetSpec()
	g.Expect(proto.Equal(registered[0].Spec().Proto(), basicModel.Spec().Proto())).To(BeTrue())
	/*g.Expect(model.GetModule()).To(BeEquivalentTo(basicModel.Spec().Module))
	g.Expect(model.GetVersion()).To(BeEquivalentTo(basicModel.Spec().Version))
	g.Expect(model.GetType()).To(BeEquivalentTo(basicModel.Spec().Type))
	g.Expect(model.GetClass()).To(BeEquivalentTo(basicModel.Spec().Class))*/
}

func TestRegisterDuplicate(t *testing.T) {
	g := NewGomegaWithT(t)
	ResetDefaultRegistry()

	g.Expect(Register(&testmodel.Basic{}, Spec{
		Module:  "module",
		Version: "v1",
		Type:    "basic",
		Class:   "config",
	})).ToNot(BeNil())
	g.Expect(func() {
		Register(&testmodel.Basic{}, Spec{
			Module:  "module",
			Version: "v1",
			Type:    "basic2",
			Class:   "config",
		})
	}).To(Panic())
}

func TestRegisterClassFallback(t *testing.T) {
	g := NewGomegaWithT(t)
	ResetDefaultRegistry()

	Register(&testmodel.Basic{}, Spec{
		Module:  "module",
		Version: "v1",
		Type:    "basic",
		// Class is not set
	})

	model, err := GetModelFor(&testmodel.Basic{})
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(model.Spec().Class).To(Equal("config"))
}
