// Copyright (c) 2018 Cisco and/or its affiliates.
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
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/bennyscetbun/jsongo"
	"github.com/sirupsen/logrus"
)

// top level objects
const (
	objTypes     = "types"
	objMessages  = "messages"
	objUnions    = "unions"
	objEnums     = "enums"
	objServices  = "services"
	objAliases   = "aliases"
	vlAPIVersion = "vl_api_version"
	objOptions   = "options"
)

// various object fields
const (
	crcField   = "crc"
	msgIdField = "_vl_msg_id"

	clientIndexField = "client_index"
	contextField     = "context"

	aliasLengthField = "length"
	aliasTypeField   = "type"

	replyField  = "reply"
	streamField = "stream"
	eventsField = "events"
)

// service name parts
const (
	serviceEventPrefix   = "want_"
	serviceDumpSuffix    = "_dump"
	serviceDetailsSuffix = "_details"
	serviceReplySuffix   = "_reply"
	serviceNoReply       = "null"
)

// field meta info
const (
	fieldMetaLimit = "limit"
)

// module options
const (
	versionOption = "version"
)

// parsePackage parses provided JSON data into objects prepared for code generation
func parsePackage(ctx *context, jsonRoot *jsongo.JSONNode) (*Package, error) {
	pkg := Package{
		Name:   ctx.packageName,
		RefMap: make(map[string]string),
	}

	// parse CRC for API version
	if crc := jsonRoot.At(vlAPIVersion); crc.GetType() == jsongo.TypeValue {
		pkg.CRC = crc.Get().(string)
	}

	// parse version string
	if opt := jsonRoot.Map(objOptions); opt.GetType() == jsongo.TypeMap {
		if ver := opt.Map(versionOption); ver.GetType() == jsongo.TypeValue {
			pkg.Version = ver.Get().(string)
		}
	}

	logf("parsing package %s (version: %s, CRC: %s)", pkg.Name, pkg.Version, pkg.CRC)
	logf(" consists of:")
	for _, key := range jsonRoot.GetKeys() {
		logf("  - %d %s", jsonRoot.At(key).Len(), key)
	}

	// parse enums
	enums := jsonRoot.Map(objEnums)
	pkg.Enums = make([]Enum, enums.Len())
	for i := 0; i < enums.Len(); i++ {
		enumNode := enums.At(i)

		enum, err := parseEnum(ctx, enumNode)
		if err != nil {
			return nil, err
		}
		pkg.Enums[i] = *enum
		pkg.RefMap[toApiType(enum.Name)] = enum.Name
	}
	// sort enums
	sort.SliceStable(pkg.Enums, func(i, j int) bool {
		return pkg.Enums[i].Name < pkg.Enums[j].Name
	})

	// parse aliases
	aliases := jsonRoot.Map(objAliases)
	if aliases.GetType() == jsongo.TypeMap {
		pkg.Aliases = make([]Alias, aliases.Len())
		for i, key := range aliases.GetKeys() {
			aliasNode := aliases.At(key)

			alias, err := parseAlias(ctx, key.(string), aliasNode)
			if err != nil {
				return nil, err
			}
			pkg.Aliases[i] = *alias
			pkg.RefMap[toApiType(alias.Name)] = alias.Name
		}
	}
	// sort aliases to ensure consistent order
	sort.Slice(pkg.Aliases, func(i, j int) bool {
		return pkg.Aliases[i].Name < pkg.Aliases[j].Name
	})

	// parse types
	types := jsonRoot.Map(objTypes)
	pkg.Types = make([]Type, types.Len())
	for i := 0; i < types.Len(); i++ {
		typNode := types.At(i)

		typ, err := parseType(ctx, typNode)
		if err != nil {
			return nil, err
		}
		pkg.Types[i] = *typ
		pkg.RefMap[toApiType(typ.Name)] = typ.Name
	}
	// sort types
	sort.SliceStable(pkg.Types, func(i, j int) bool {
		return pkg.Types[i].Name < pkg.Types[j].Name
	})

	// parse unions
	unions := jsonRoot.Map(objUnions)
	pkg.Unions = make([]Union, unions.Len())
	for i := 0; i < unions.Len(); i++ {
		unionNode := unions.At(i)

		union, err := parseUnion(ctx, unionNode)
		if err != nil {
			return nil, err
		}
		pkg.Unions[i] = *union
		pkg.RefMap[toApiType(union.Name)] = union.Name
	}
	// sort unions
	sort.SliceStable(pkg.Unions, func(i, j int) bool {
		return pkg.Unions[i].Name < pkg.Unions[j].Name
	})

	// parse messages
	messages := jsonRoot.Map(objMessages)
	pkg.Messages = make([]Message, messages.Len())
	for i := 0; i < messages.Len(); i++ {
		msgNode := messages.At(i)

		msg, err := parseMessage(ctx, msgNode)
		if err != nil {
			return nil, err
		}
		pkg.Messages[i] = *msg
	}
	// sort messages
	sort.SliceStable(pkg.Messages, func(i, j int) bool {
		return pkg.Messages[i].Name < pkg.Messages[j].Name
	})

	// parse services
	services := jsonRoot.Map(objServices)
	if services.GetType() == jsongo.TypeMap {
		pkg.Services = make([]Service, services.Len())
		for i, key := range services.GetKeys() {
			svcNode := services.At(key)

			svc, err := parseService(ctx, key.(string), svcNode)
			if err != nil {
				return nil, err
			}
			pkg.Services[i] = *svc
		}
	}
	// sort services
	sort.Slice(pkg.Services, func(i, j int) bool {
		// dumps first
		if pkg.Services[i].Stream != pkg.Services[j].Stream {
			return pkg.Services[i].Stream
		}
		return pkg.Services[i].RequestType < pkg.Services[j].RequestType
	})

	printPackage(&pkg)

	return &pkg, nil
}

// parseEnum parses VPP binary API enum object from JSON node
func parseEnum(ctx *context, enumNode *jsongo.JSONNode) (*Enum, error) {
	if enumNode.Len() == 0 || enumNode.At(0).GetType() != jsongo.TypeValue {
		return nil, errors.New("invalid JSON for enum specified")
	}

	enumName, ok := enumNode.At(0).Get().(string)
	if !ok {
		return nil, fmt.Errorf("enum name is %T, not a string", enumNode.At(0).Get())
	}
	enumType, ok := enumNode.At(enumNode.Len() - 1).At("enumtype").Get().(string)
	if !ok {
		return nil, fmt.Errorf("enum type invalid or missing")
	}

	enum := Enum{
		Name: enumName,
		Type: enumType,
	}

	// loop through enum entries, skip first (name) and last (enumtype)
	for j := 1; j < enumNode.Len()-1; j++ {
		if enumNode.At(j).GetType() == jsongo.TypeArray {
			entry := enumNode.At(j)

			if entry.Len() < 2 || entry.At(0).GetType() != jsongo.TypeValue || entry.At(1).GetType() != jsongo.TypeValue {
				return nil, errors.New("invalid JSON for enum entry specified")
			}

			entryName, ok := entry.At(0).Get().(string)
			if !ok {
				return nil, fmt.Errorf("enum entry name is %T, not a string", entry.At(0).Get())
			}
			entryVal := entry.At(1).Get()

			enum.Entries = append(enum.Entries, EnumEntry{
				Name:  entryName,
				Value: entryVal,
			})
		}
	}

	return &enum, nil
}

// parseUnion parses VPP binary API union object from JSON node
func parseUnion(ctx *context, unionNode *jsongo.JSONNode) (*Union, error) {
	if unionNode.Len() == 0 || unionNode.At(0).GetType() != jsongo.TypeValue {
		return nil, errors.New("invalid JSON for union specified")
	}

	unionName, ok := unionNode.At(0).Get().(string)
	if !ok {
		return nil, fmt.Errorf("union name is %T, not a string", unionNode.At(0).Get())
	}
	var unionCRC string
	if unionNode.At(unionNode.Len()-1).GetType() == jsongo.TypeMap {
		unionCRC = unionNode.At(unionNode.Len() - 1).At(crcField).Get().(string)
	}

	union := Union{
		Name: unionName,
		CRC:  unionCRC,
	}

	// loop through union fields, skip first (name)
	for j := 1; j < unionNode.Len(); j++ {
		if unionNode.At(j).GetType() == jsongo.TypeArray {
			fieldNode := unionNode.At(j)

			field, err := parseField(ctx, fieldNode)
			if err != nil {
				return nil, err
			}

			union.Fields = append(union.Fields, *field)
		}
	}

	return &union, nil
}

// parseType parses VPP binary API type object from JSON node
func parseType(ctx *context, typeNode *jsongo.JSONNode) (*Type, error) {
	if typeNode.Len() == 0 || typeNode.At(0).GetType() != jsongo.TypeValue {
		return nil, errors.New("invalid JSON for type specified")
	}

	typeName, ok := typeNode.At(0).Get().(string)
	if !ok {
		return nil, fmt.Errorf("type name is %T, not a string", typeNode.At(0).Get())
	}
	var typeCRC string
	if lastField := typeNode.At(typeNode.Len() - 1); lastField.GetType() == jsongo.TypeMap {
		typeCRC = lastField.At(crcField).Get().(string)
	}

	typ := Type{
		Name: typeName,
		CRC:  typeCRC,
	}

	// loop through type fields, skip first (name)
	for j := 1; j < typeNode.Len(); j++ {
		if typeNode.At(j).GetType() == jsongo.TypeArray {
			fieldNode := typeNode.At(j)

			field, err := parseField(ctx, fieldNode)
			if err != nil {
				return nil, err
			}

			typ.Fields = append(typ.Fields, *field)
		}
	}

	return &typ, nil
}

// parseAlias parses VPP binary API alias object from JSON node
func parseAlias(ctx *context, aliasName string, aliasNode *jsongo.JSONNode) (*Alias, error) {
	if aliasNode.Len() == 0 || aliasNode.At(aliasTypeField).GetType() != jsongo.TypeValue {
		return nil, errors.New("invalid JSON for alias specified")
	}

	alias := Alias{
		Name: aliasName,
	}

	if typeNode := aliasNode.At(aliasTypeField); typeNode.GetType() == jsongo.TypeValue {
		typ, ok := typeNode.Get().(string)
		if !ok {
			return nil, fmt.Errorf("alias type is %T, not a string", typeNode.Get())
		}
		if typ != "null" {
			alias.Type = typ
		}
	}

	if lengthNode := aliasNode.At(aliasLengthField); lengthNode.GetType() == jsongo.TypeValue {
		length, ok := lengthNode.Get().(float64)
		if !ok {
			return nil, fmt.Errorf("alias length is %T, not a float64", lengthNode.Get())
		}
		alias.Length = int(length)
	}

	return &alias, nil
}

// parseMessage parses VPP binary API message object from JSON node
func parseMessage(ctx *context, msgNode *jsongo.JSONNode) (*Message, error) {
	if msgNode.Len() == 0 || msgNode.At(0).GetType() != jsongo.TypeValue {
		return nil, errors.New("invalid JSON for message specified")
	}

	msgName, ok := msgNode.At(0).Get().(string)
	if !ok {
		return nil, fmt.Errorf("message name is %T, not a string", msgNode.At(0).Get())
	}
	msgCRC, ok := msgNode.At(msgNode.Len() - 1).At(crcField).Get().(string)
	if !ok {

		return nil, fmt.Errorf("message crc invalid or missing")
	}

	msg := Message{
		Name: msgName,
		CRC:  msgCRC,
	}

	// loop through message fields, skip first (name) and last (crc)
	for j := 1; j < msgNode.Len()-1; j++ {
		if msgNode.At(j).GetType() == jsongo.TypeArray {
			fieldNode := msgNode.At(j)

			field, err := parseField(ctx, fieldNode)
			if err != nil {
				return nil, err
			}

			msg.Fields = append(msg.Fields, *field)
		}
	}

	return &msg, nil
}

// parseField parses VPP binary API object field from JSON node
func parseField(ctx *context, field *jsongo.JSONNode) (*Field, error) {
	if field.Len() < 2 || field.At(0).GetType() != jsongo.TypeValue || field.At(1).GetType() != jsongo.TypeValue {
		return nil, errors.New("invalid JSON for field specified")
	}

	fieldType, ok := field.At(0).Get().(string)
	if !ok {
		return nil, fmt.Errorf("field type is %T, not a string", field.At(0).Get())
	}
	fieldName, ok := field.At(1).Get().(string)
	if !ok {
		return nil, fmt.Errorf("field name is %T, not a string", field.At(1).Get())
	}

	f := &Field{
		Name: fieldName,
		Type: fieldType,
	}

	if field.Len() >= 3 {
		if field.At(2).GetType() == jsongo.TypeValue {
			fieldLength, ok := field.At(2).Get().(float64)
			if !ok {
				return nil, fmt.Errorf("field length is %T, not float64", field.At(2).Get())
			}
			f.Length = int(fieldLength)
		} else if field.At(2).GetType() == jsongo.TypeMap {
			fieldMeta := field.At(2)

			for _, key := range fieldMeta.GetKeys() {
				metaNode := fieldMeta.At(key)

				switch metaName := key.(string); metaName {
				case fieldMetaLimit:
					f.Meta.Limit = int(metaNode.Get().(float64))
				default:
					logrus.Warnf("unknown meta info (%s) for field (%s)", metaName, fieldName)
				}
			}
		} else {
			return nil, errors.New("invalid JSON for field specified")
		}
	}
	if field.Len() >= 4 {
		fieldLengthFrom, ok := field.At(3).Get().(string)
		if !ok {
			return nil, fmt.Errorf("field length from is %T, not a string", field.At(3).Get())
		}
		f.SizeFrom = fieldLengthFrom
	}

	return f, nil
}

// parseService parses VPP binary API service object from JSON node
func parseService(ctx *context, svcName string, svcNode *jsongo.JSONNode) (*Service, error) {
	if svcNode.Len() == 0 || svcNode.At(replyField).GetType() != jsongo.TypeValue {
		return nil, errors.New("invalid JSON for service specified")
	}

	svc := Service{
		Name:        svcName,
		RequestType: svcName,
	}

	if replyNode := svcNode.At(replyField); replyNode.GetType() == jsongo.TypeValue {
		reply, ok := replyNode.Get().(string)
		if !ok {
			return nil, fmt.Errorf("service reply is %T, not a string", replyNode.Get())
		}
		if reply != serviceNoReply {
			svc.ReplyType = reply
		}
	}

	// stream service (dumps)
	if streamNode := svcNode.At(streamField); streamNode.GetType() == jsongo.TypeValue {
		var ok bool
		svc.Stream, ok = streamNode.Get().(bool)
		if !ok {
			return nil, fmt.Errorf("service stream is %T, not a string", streamNode.Get())
		}
	}

	// events service (event subscription)
	if eventsNode := svcNode.At(eventsField); eventsNode.GetType() == jsongo.TypeArray {
		for j := 0; j < eventsNode.Len(); j++ {
			event := eventsNode.At(j).Get().(string)
			svc.Events = append(svc.Events, event)
		}
	}

	// validate service
	if len(svc.Events) > 0 {
		// EVENT service
		if !strings.HasPrefix(svc.RequestType, serviceEventPrefix) {
			logrus.Debugf("unusual EVENTS service: %+v\n"+
				"- events service %q does not have %q prefix in request.",
				svc, svc.Name, serviceEventPrefix)
		}
	} else if svc.Stream {
		// STREAM service
		if !strings.HasSuffix(svc.RequestType, serviceDumpSuffix) ||
			!strings.HasSuffix(svc.ReplyType, serviceDetailsSuffix) {
			logrus.Debugf("unusual STREAM service: %+v\n"+
				"- stream service %q does not have %q suffix in request or reply does not have %q suffix.",
				svc, svc.Name, serviceDumpSuffix, serviceDetailsSuffix)
		}
	} else if svc.ReplyType != "" && svc.ReplyType != serviceNoReply {
		// REQUEST service
		// some messages might have `null` reply (for example: memclnt)
		if !strings.HasSuffix(svc.ReplyType, serviceReplySuffix) {
			logrus.Debugf("unusual REQUEST service: %+v\n"+
				"- service %q does not have %q suffix in reply.",
				svc, svc.Name, serviceReplySuffix)
		}
	}

	return &svc, nil
}
