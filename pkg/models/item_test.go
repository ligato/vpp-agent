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

	"go.ligato.io/vpp-agent/v3/pkg/models"
	testmodel "go.ligato.io/vpp-agent/v3/pkg/models/testdata/proto"
)

func TestEncode(t *testing.T) {
	tc := setupTest(t)
	defer tc.teardownTest()

	instance := &testmodel.Basic{
		Name:           "basic1",
		ValueInt:       -20,
		ValueUint:      3,
		ValueInt64:     99000000123,
		RepeatedString: []string{"alpha", "beta", "gama"},
	}
	t.Logf("instance: %#v", instance)

	item, err := models.MarshalItem(instance)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	t.Logf("marshalled:\n%+v", proto.MarshalTextString(item))

	tc.Expect(item.GetData().GetAny().GetTypeUrl()).
		To(Equal("models.ligato.io/model.Basic"))

	out, err := models.UnmarshalItem(item)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	t.Logf("unmarshalled:\n%+v", proto.MarshalTextString(out))
}

func TestDecode(t *testing.T) {
	tc := setupTest(t)
	defer tc.teardownTest()

	in := &testmodel.Basic{
		Name:           "basic1",
		ValueInt:       -20,
		ValueUint:      3,
		ValueInt64:     99000000123,
		RepeatedString: []string{"alpha", "beta", "gama"},
	}
	t.Logf("in: %#v", in)

	item, err := models.MarshalItem(in)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	t.Logf("marshalled:\n%+v", proto.MarshalTextString(item))

	tc.Expect(item.GetId().GetModel()).To(Equal("module.basic"))
	tc.Expect(item.GetId().GetName()).To(Equal("basic1"))

	tc.Expect(item.GetData().GetAny().GetTypeUrl()).
		To(Equal("models.ligato.io/model.Basic"))

	out, err := models.UnmarshalItem(item)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	t.Logf("unmarshalled:\n%+v", proto.MarshalTextString(out))

	tc.Expect(proto.Equal(in, out)).To(BeTrue())
}
