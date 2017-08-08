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
func ListExportedFields(val interface{}, predicates ...ExportedPredicate) []*reflect.StructField {
	valType := reflect.Indirect(reflect.ValueOf(val)).Type()
	len := valType.NumField()
	ret := []*reflect.StructField{}
	for i := 0; i < len; i++ {
		structField := valType.Field(i)

		if FieldExported(&structField, predicates...) {
			ret = append(ret, &structField)
		}
	}

	return ret
}

// ListExportedFieldsWithVals returns all fields of a structure that starts wit uppercase letter with values
func ListExportedFieldsWithVals(val interface{}, predicates ...ExportedPredicate) (fields []*reflect.StructField, values []interface{}) {
	valRefl := reflect.Indirect(reflect.ValueOf(val))
	valType := valRefl.Type()
	len := valType.NumField()
	fields = []*reflect.StructField{}
	values = []interface{}{}
	for i := 0; i < len; i++ {
		structField := valType.Field(i)

		if FieldExported(&structField, predicates...) {
			// if exported
			fields = append(fields, &structField)
			values = append(values, valRefl.Field(i).Interface())
		}
	}

	return fields, values
}

// ExportedPredicate defines a callback (used in func FieldExported)
type ExportedPredicate func(field *reflect.StructField) bool

// FieldExported returns true if field name starts with uppercase
func FieldExported(field *reflect.StructField, predicates ...ExportedPredicate) (exported bool) {
	if field.Name[0] == strings.ToUpper(string(field.Name[0]))[0] {
		expPredic := true
		for _, predicate := range predicates {
			if !predicate(field) {
				expPredic = false
				break
			}
		}

		return expPredic
	}

	return false
}

// ListExportedFieldsPtrs iterates struct fields and return slice of pointers to field values
func ListExportedFieldsPtrs(val interface{}, predicates ...ExportedPredicate) []interface{} {
	rVal := reflect.Indirect(reflect.ValueOf(val))
	ptrs := []interface{}{}
	for i := 0; i < rVal.NumField(); i++ {
		field := rVal.Field(i)
		structField := rVal.Type().Field(i)
		if !FieldExported(&structField, predicates...) {
			continue
		}

		switch field.Kind() {
		case reflect.Ptr, reflect.Interface:
			if field.IsNil() {
				p := reflect.New(field.Type().Elem())
				field.Set(p)
				ptrs = append(ptrs, p.Interface())
			} else {
				ptrs = append(ptrs, field.Interface())
			}
		case reflect.Slice, reflect.Chan, reflect.Map:
			if field.IsNil() {
				p := reflect.New(field.Type())
				field.Set(p.Elem())
				ptrs = append(ptrs, field.Addr().Interface())
			} else {
				ptrs = append(ptrs, field.Interface())
			}
		default:
			if field.CanAddr() {
				ptrs = append(ptrs, field.Addr().Interface())
			} else if field.IsValid() {
				ptrs = append(ptrs, field.Interface())
			} else {
				panic("invalid field")
			}
		}
	}

	return ptrs
}
