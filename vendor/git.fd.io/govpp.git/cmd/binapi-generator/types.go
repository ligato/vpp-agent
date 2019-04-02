// Copyright (c) 2019 Cisco and/or its affiliates.
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

package main

import (
	"fmt"
	"strconv"
	"strings"
)

// toApiType returns name that is used as type reference in VPP binary API
func toApiType(name string) string {
	return fmt.Sprintf("vl_api_%s_t", name)
}

// binapiTypes is a set of types used VPP binary API for translation to Go types
var binapiTypes = map[string]string{
	"u8":  "uint8",
	"i8":  "int8",
	"u16": "uint16",
	"i16": "int16",
	"u32": "uint32",
	"i32": "int32",
	"u64": "uint64",
	"i64": "int64",
	"f64": "float64",
}

func getBinapiTypeSize(binapiType string) int {
	if _, ok := binapiTypes[binapiType]; ok {
		b, err := strconv.Atoi(strings.TrimLeft(binapiType, "uif"))
		if err == nil {
			return b / 8
		}
	}
	return -1
}

// convertToGoType translates the VPP binary API type into Go type
func convertToGoType(ctx *context, binapiType string) (typ string) {
	if t, ok := binapiTypes[binapiType]; ok {
		// basic types
		typ = t
	} else if r, ok := ctx.packageData.RefMap[binapiType]; ok {
		// specific types (enums/types/unions)
		typ = camelCaseName(r)
	} else {
		switch binapiType {
		case "bool", "string":
			typ = binapiType
		default:
			// fallback type
			log.Warnf("found unknown VPP binary API type %q, using byte", binapiType)
			typ = "byte"
		}
	}
	return typ
}

func getSizeOfType(typ *Type) (size int) {
	for _, field := range typ.Fields {
		size += getSizeOfBinapiTypeLength(field.Type, field.Length)
	}
	return size
}

func getSizeOfBinapiTypeLength(typ string, length int) (size int) {
	if n := getBinapiTypeSize(typ); n > 0 {
		if length > 0 {
			return n * length
		} else {
			return n
		}
	}
	return
}

func getTypeByRef(ctx *context, ref string) *Type {
	for _, typ := range ctx.packageData.Types {
		if ref == toApiType(typ.Name) {
			return &typ
		}
	}
	return nil
}

func getAliasByRef(ctx *context, ref string) *Alias {
	for _, alias := range ctx.packageData.Aliases {
		if ref == toApiType(alias.Name) {
			return &alias
		}
	}
	return nil
}

func getUnionSize(ctx *context, union *Union) (maxSize int) {
	for _, field := range union.Fields {
		typ := getTypeByRef(ctx, field.Type)
		if typ != nil {
			if size := getSizeOfType(typ); size > maxSize {
				maxSize = size
			}
			continue
		}
		alias := getAliasByRef(ctx, field.Type)
		if alias != nil {
			if size := getSizeOfBinapiTypeLength(alias.Type, alias.Length); size > maxSize {
				maxSize = size
			}
			continue
		} else {
			logf("no type or alias found for union %s field type %q", union.Name, field.Type)
			continue
		}
	}
	return
}
