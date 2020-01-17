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

package models

import (
	"github.com/golang/protobuf/descriptor"
	"github.com/golang/protobuf/proto"

	"go.ligato.io/vpp-agent/v3/proto/ligato/generic"
)

func (r *Registry) checkProtoOptions(x interface{}) *KnownModel {
	p, ok := x.(descriptor.Message)
	if !ok {
		return nil
	}
	_, md := descriptor.ForMessage(p)
	s, err := proto.GetExtension(md.Options, generic.E_Model)
	if err != nil {
		return nil
	}
	if spec, ok := s.(*generic.ModelSpec); ok {
		km, err := r.Register(x, ToSpec(spec))
		if err != nil {
			panic(err)
		}
		return km
	}
	return nil
}
