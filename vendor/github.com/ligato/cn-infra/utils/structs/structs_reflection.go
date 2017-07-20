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

package structs

import (
	"reflect"
	"strings"
)

// FindField compares the pointers (pointerToAField with all fields in pointerToAStruct)
func FindField(pointerToAField interface{}, pointerToAStruct interface{}) (field *reflect.StructField, found bool) {
	fieldVal := reflect.ValueOf(pointerToAField)

	if fieldVal.Kind() != reflect.Ptr {
		panic("pointerToAField must be a pointer")
	}

	strct := reflect.Indirect(reflect.ValueOf(pointerToAStruct))
	numField := strct.NumField()
	for i := 0; i < numField; i++ {
		sf := strct.Field(i)

		if sf.CanAddr() {
			if fieldVal.Pointer() == sf.Addr().Pointer() {
				field := strct.Type().Field(i)
				return &field, true
			}
		}
	}

	return nil, false
}

// ListExportedFields returns all fields of a structure that starts wit uppercase letter
func ListExportedFields(val interface{}) []*reflect.StructField {
	valType := reflect.Indirect(reflect.ValueOf(val)).Type()
	len := valType.NumField()
	ret := []*reflect.StructField{}
	for i := 0; i < len; i++ {
		field := valType.Field(i)
		if field.Name[0] == strings.ToUpper(string(field.Name[0]))[0] {
			// if exported
			ret = append(ret, &field)
		}
	}

	return ret
}

// ListExportedFieldsWithVals returns all fields of a structure that starts wit uppercase letter with values
func ListExportedFieldsWithVals(val interface{}) (fields []*reflect.StructField, values []interface{}) {
	valRefl := reflect.Indirect(reflect.ValueOf(val))
	valType := valRefl.Type()
	len := valType.NumField()
	fields = []*reflect.StructField{}
	values = []interface{}{}
	for i := 0; i < len; i++ {
		field := valType.Field(i)
		if field.Name[0] == strings.ToUpper(string(field.Name[0]))[0] {
			// if exported
			fields = append(fields, &field)
			values = append(values, valRefl.Field(i).Interface())
		}
	}

	return fields, values
}
