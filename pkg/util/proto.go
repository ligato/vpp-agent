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

package util

import (
	"reflect"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func ExtractProtos(from ...interface{}) (protos []proto.Message) {
	for _, v := range from {
		if reflect.ValueOf(v).IsNil() {
			continue
		}
		val := reflect.ValueOf(v).Elem()
		typ := val.Type()
		if typ.Kind() != reflect.Struct {
			return
		}
		for i := 0; i < typ.NumField(); i++ {
			field := val.Field(i)
			if field.Kind() == reflect.Slice {
				for idx := 0; idx < field.Len(); idx++ {
					elem := field.Index(idx)
					if msg, ok := elem.Interface().(proto.Message); ok {
						protos = append(protos, msg)
					}
				}
			} else if field.Kind() == reflect.Ptr && !field.IsNil() {
				if msg, ok := field.Interface().(proto.Message); ok && !field.IsNil() {
					protos = append(protos, msg)
				}
			}
		}
	}
	return
}

func PlaceProtos(protos map[string]proto.Message, dsts ...interface{}) {
	for _, prot := range protos {
		protTyp := reflect.TypeOf(prot)
		for _, dst := range dsts {
			dstVal := reflect.ValueOf(dst).Elem()
			dstTyp := dstVal.Type()
			if dstTyp.Kind() != reflect.Struct {
				return
			}
			for i := 0; i < dstTyp.NumField(); i++ {
				field := dstVal.Field(i)
				if field.Kind() == reflect.Slice {
					if protTyp.AssignableTo(field.Type().Elem()) {
						field.Set(reflect.Append(field, reflect.ValueOf(prot)))
					}
				} else {
					if field.Type() == protTyp {
						field.Set(reflect.ValueOf(prot))
					}
				}
			}
		}
	}
	return
}

// PlaceProtosIntoProtos fills dsts proto messages (direct or transitive) fields with protos values.
// The matching is done by message descriptor's full name. The <clearIgnoreLayerCount> variable controls
// how many top model structure hierarchy layers can have empty values for messages (see
// util.placeProtosInProto(...) for details)
func PlaceProtosIntoProtos(protos []proto.Message, clearIgnoreLayerCount int, dsts ...proto.Message) {
	// create help structure for insertion proto messages
	// (map values are protoreflect.Message(s) that contains proto message and its type. These messages will be
	// later wrapped into protoreflect.Value(s) and filled into destination proto message using proto reflection.
	// We could have used protoreflect.Value(s) as map values in this help structure, but protoreflect.Value type
	// is cheap (really thin wrapper) and it is unknown whether using the same value on multiple fields could
	// cause problems, so we will generate it for each field.
	messageMap := make(map[string][]protoreflect.Message)
	for _, protoMsg := range protos {
		protoName := string(protoMsg.ProtoReflect().Descriptor().FullName())
		messageMap[protoName] = append(messageMap[protoName], protoMsg.ProtoReflect())
	}

	// insert proto message to all destination containers (also proto messages)
	for _, dst := range dsts {
		placeProtosInProto(dst, messageMap, clearIgnoreLayerCount)
	}
}

// placeProtosInProto fills dst proto message (direct or transitive) fields with protos values from messageMap
// (convenient map[proto descriptor full name]= protoreflect message containing proto message and proto type).
// The matching is done by message descriptor's full name. The function is recursive and one run is handling
// only one level of proto message structure tree (only handling Message references and ignoring
// scalar/enum/... values). The <clearIgnoreLayerCount> controls how many top layer can have empty values
// for their message fields (as the algorithm backtracks the descriptor model tree, it unfortunately initialize
// empty value for visited fields). The layer below <clearIgnoreLayerCount> top layer will be cleared
// from the fake empty value. Currently unsupported are maps as fields.
func placeProtosInProto(dst proto.Message, messageMap map[string][]protoreflect.Message, clearIgnoreLayerCount int) bool {
	changed := false
	fields := dst.ProtoReflect().Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		fieldMessageDesc := field.Message()
		if fieldMessageDesc != nil { // only interested in MessageKind or GroupKind fields
			if messageForField, typeMatch := messageMap[string(fieldMessageDesc.FullName())]; typeMatch {
				// fill value(s)
				if field.IsList() {
					list := dst.ProtoReflect().Mutable(field).List()
					for _, message := range messageForField {
						list.Append(protoreflect.ValueOf(message))
						changed = true
					}
				} else if field.IsMap() { // unsupported
				} else {
					dst.ProtoReflect().Set(field, protoreflect.ValueOf(messageForField[0]))
					changed = true
				}
			} else {
				// no type match -> check deeper structure layers
				// Note: dst.ProtoReflect().Mutable(field) creates empty value that creates problems later
				// (i.e. by outputing to json/yaml) => need to check whether actual value has been assigned
				// and if not then clear the field. Additionally there is clearIgnoreLayerCount variable
				// disabling this clearing functionality for upper <clearIgnoreLayerCount> layers
				if field.IsList() {
					list := dst.ProtoReflect().Mutable(field).List()
					changeOnLowerLayer := false
					for j := 0; j < list.Len(); j++ {
						changeOnLowerLayer = changeOnLowerLayer ||
							placeProtosInProto(list.Get(j).Message().Interface(), messageMap, clearIgnoreLayerCount-1)
					}
					if !changeOnLowerLayer && clearIgnoreLayerCount <= 0 {
						dst.ProtoReflect().Clear(field)
					}
					changed = changed || changeOnLowerLayer
				} else if field.IsMap() { // unsupported
				} else {
					changeOnLowerLayer := placeProtosInProto(dst.ProtoReflect().Mutable(field).
						Message().Interface(), messageMap, clearIgnoreLayerCount-1)
					if !changeOnLowerLayer && clearIgnoreLayerCount <= 0 {
						dst.ProtoReflect().Clear(field)
					}
					changed = changed || changeOnLowerLayer
				}
			}
		}
	}
	return changed
}
