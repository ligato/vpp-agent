package converter

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	srv6 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/srv6"

	"github.com/alecthomas/jsonschema"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/iancoleman/orderedmap"
	"github.com/xeipuuv/gojsonschema"
)

var (
	globalPkg = newProtoPackage(nil, "")

	wellKnownTypes = map[string]bool{
		"DoubleValue": true,
		"FloatValue":  true,
		"Int64Value":  true,
		"UInt64Value": true,
		"Int32Value":  true,
		"UInt32Value": true,
		"BoolValue":   true,
		"StringValue": true,
		"BytesValue":  true,
		"Value":       true,
	}
)

// min/max constants that are safe to assign to int on 32-bit systems
// The "github.com/alecthomas/jsonschema".Type has manimum and maximum defined as int, but that is insufficient
// for some types. Therefore the ranges for these types must be artificially cut to be usable with int.
var (
	intSafeMaxUint32 int = math.MaxInt32 // int32 can't hold values up to math.MaxUint32
	intSafeMinInt64  int = math.MinInt32
	intSafeMaxInt64  int = math.MaxInt32
	intSafeMaxUint64 int = math.MaxInt32
)

func init() {
	if strconv.IntSize == 64 { // override of min/max constants for 64-bit systems
		intSafeMaxUint32 = math.MaxUint32
		intSafeMinInt64 = math.MinInt64
		intSafeMaxInt64 = math.MaxInt64
		intSafeMaxUint64 = math.MaxInt64 // int64 can't hold values up to math.MaxUint64
	}
}

func (c *Converter) registerEnum(pkgName *string, enum *descriptor.EnumDescriptorProto) {
	pkg := globalPkg
	if pkgName != nil {
		for _, node := range strings.Split(*pkgName, ".") {
			if pkg == globalPkg && node == "" {
				// Skips leading "."
				continue
			}
			child, ok := pkg.children[node]
			if !ok {
				child = newProtoPackage(pkg, node)
				pkg.children[node] = child
			}
			pkg = child
		}
	}
	pkg.enums[enum.GetName()] = enum
}

func (c *Converter) registerType(pkgName *string, msg *descriptor.DescriptorProto) {
	pkg := globalPkg
	if pkgName != nil {
		for _, node := range strings.Split(*pkgName, ".") {
			if pkg == globalPkg && node == "" {
				// Skips leading "."
				continue
			}
			child, ok := pkg.children[node]
			if !ok {
				child = newProtoPackage(pkg, node)
				pkg.children[node] = child
			}
			pkg = child
		}
	}
	pkg.types[msg.GetName()] = msg
}

// Convert a proto "field" (essentially a type-switch with some recursion):
func (c *Converter) convertField(curPkg *ProtoPackage, desc *descriptor.FieldDescriptorProto, msg *descriptor.DescriptorProto, duplicatedMessages map[*descriptor.DescriptorProto]string) (*jsonschema.Type, error) {
	// Prepare a new jsonschema.Type for our eventual return value:
	jsonSchemaType := &jsonschema.Type{}

	// Generate a description from src comments (if available)
	if src := c.sourceInfo.GetField(desc); src != nil {
		jsonSchemaType.Description = formatDescription(src)
	}

	// get field annotations
	var fieldAnnotations *srv6.LigatoOptions
	val, err := proto.GetExtension(desc.Options, srv6.E_LigatoOptions)
	if err != nil {
		c.logger.Debugf("Field %s.%s doesn't have ligato option extension", msg.GetName(), desc.GetName())
	} else {
		var ok bool
		if fieldAnnotations, ok = val.(*srv6.LigatoOptions); !ok {
			c.logger.Debugf("Field %s.%s have ligato option extension, but its value has "+
				"unexpected type (%T)", msg.GetName(), desc.GetName(), val)
		}
	}

	// Switch the types, and pick a JSONSchema equivalent:
	switch desc.GetType() {
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE,
		descriptor.FieldDescriptorProto_TYPE_FLOAT:
		if c.AllowNullValues {
			jsonSchemaType.OneOf = []*jsonschema.Type{
				{Type: gojsonschema.TYPE_NULL},
				{Type: gojsonschema.TYPE_NUMBER},
			}
		} else {
			jsonSchemaType.Type = gojsonschema.TYPE_NUMBER
		}

	case descriptor.FieldDescriptorProto_TYPE_INT32,
		descriptor.FieldDescriptorProto_TYPE_SFIXED32,
		descriptor.FieldDescriptorProto_TYPE_SINT32:
		if c.AllowNullValues {
			jsonSchemaType.OneOf = []*jsonschema.Type{
				{
					Type: gojsonschema.TYPE_NULL,
				}, {
					Type:    gojsonschema.TYPE_INTEGER,
					Minimum: math.MinInt32,
					Maximum: math.MaxInt32,
				},
			}
		} else {
			jsonSchemaType.Type = gojsonschema.TYPE_INTEGER
			jsonSchemaType.Minimum = math.MinInt32
			jsonSchemaType.Maximum = math.MaxInt32
		}

	case descriptor.FieldDescriptorProto_TYPE_UINT32,
		descriptor.FieldDescriptorProto_TYPE_FIXED32:
		if c.AllowNullValues {
			jsonSchemaType.OneOf = []*jsonschema.Type{
				{
					Type: gojsonschema.TYPE_NULL,
				}, {
					Type:             gojsonschema.TYPE_INTEGER,
					Minimum:          -1,
					ExclusiveMinimum: true,
					Maximum:          intSafeMaxUint32,
				},
			}
		} else {
			jsonSchemaType.Type = gojsonschema.TYPE_INTEGER
			jsonSchemaType.Minimum = -1
			jsonSchemaType.ExclusiveMinimum = true
			jsonSchemaType.Maximum = intSafeMaxUint32
		}

	case descriptor.FieldDescriptorProto_TYPE_INT64,
		descriptor.FieldDescriptorProto_TYPE_SFIXED64,
		descriptor.FieldDescriptorProto_TYPE_SINT64:
		if c.AllowNullValues {
			jsonSchemaType.OneOf = []*jsonschema.Type{
				{Type: gojsonschema.TYPE_STRING},
				{Type: gojsonschema.TYPE_NULL},
			}
		} else {
			jsonSchemaType.Type = gojsonschema.TYPE_STRING
		}
		if c.DisallowBigIntsAsStrings {
			jsonSchemaType.Type = gojsonschema.TYPE_INTEGER
			jsonSchemaType.Minimum = intSafeMinInt64
			jsonSchemaType.Maximum = intSafeMaxInt64
		}

	case descriptor.FieldDescriptorProto_TYPE_UINT64,
		descriptor.FieldDescriptorProto_TYPE_FIXED64:
		if c.AllowNullValues {
			jsonSchemaType.OneOf = []*jsonschema.Type{
				{Type: gojsonschema.TYPE_STRING},
				{Type: gojsonschema.TYPE_NULL},
			}
		} else {
			jsonSchemaType.Type = gojsonschema.TYPE_STRING
		}
		if c.DisallowBigIntsAsStrings {
			jsonSchemaType.Type = gojsonschema.TYPE_INTEGER
			jsonSchemaType.Minimum = -1
			jsonSchemaType.ExclusiveMinimum = true
			jsonSchemaType.Maximum = intSafeMaxUint64
		}

	case descriptor.FieldDescriptorProto_TYPE_STRING:
		var pType *jsonschema.Type
		if c.AllowNullValues {
			pType = &jsonschema.Type{Type: gojsonschema.TYPE_STRING}
			jsonSchemaType.OneOf = []*jsonschema.Type{
				{Type: gojsonschema.TYPE_NULL},
				pType,
			}
		} else {
			jsonSchemaType.Type = gojsonschema.TYPE_STRING
			pType = jsonSchemaType
		}
		if fieldAnnotations != nil {
			switch fieldAnnotations.Type {
			case srv6.LigatoOptions_IPV6:
				pType.Format = "ipv6"
			case srv6.LigatoOptions_IPV4:
				pType.Format = "ipv4"
			}
		}

	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		if c.AllowNullValues {
			jsonSchemaType.OneOf = []*jsonschema.Type{
				{Type: gojsonschema.TYPE_NULL},
				{Type: gojsonschema.TYPE_STRING},
			}
		} else {
			jsonSchemaType.Type = gojsonschema.TYPE_STRING
		}

	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		// Note: not setting type specification(oneof string and integer), because explicitly saying which
		// values are valid (and any other is invalid) is enough specification what can be used
		// (this also overcome bug in example creator https://json-schema-faker.js.org/ that doesn't select
		// correct type for enum value but rather chooses random type from oneof and cast value to that type)
		//
		//jsonSchemaType.OneOf = append(jsonSchemaType.OneOf, &jsonschema.Type{Type: gojsonschema.TYPE_STRING})
		//jsonSchemaType.OneOf = append(jsonSchemaType.OneOf, &jsonschema.Type{Type: gojsonschema.TYPE_INTEGER})
		if c.AllowNullValues {
			jsonSchemaType.OneOf = append(jsonSchemaType.OneOf, &jsonschema.Type{Type: gojsonschema.TYPE_NULL})
		}

		// Go through all the enums we have, see if we can match any to this field.
		fullEnumIdentifier := strings.TrimPrefix(desc.GetTypeName(), ".")
		matchedEnum, _, ok := c.lookupEnum(curPkg, fullEnumIdentifier)
		if !ok {
			return nil, fmt.Errorf("unable to resolve enum type: %s", desc.GetType().String())
		}

		// We have found an enum, append its values.
		for _, value := range matchedEnum.Value {
			jsonSchemaType.Enum = append(jsonSchemaType.Enum, value.Name)
			jsonSchemaType.Enum = append(jsonSchemaType.Enum, value.Number)
		}

	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		if c.AllowNullValues {
			jsonSchemaType.OneOf = []*jsonschema.Type{
				{Type: gojsonschema.TYPE_NULL},
				{Type: gojsonschema.TYPE_BOOLEAN},
			}
		} else {
			jsonSchemaType.Type = gojsonschema.TYPE_BOOLEAN
		}

	case descriptor.FieldDescriptorProto_TYPE_GROUP, descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		switch desc.GetTypeName() {
		case ".google.protobuf.Timestamp":
			jsonSchemaType.Type = gojsonschema.TYPE_STRING
			jsonSchemaType.Format = "date-time"
		default:
			jsonSchemaType.Type = gojsonschema.TYPE_OBJECT
			// disallowAdditionalProperties will fail validation when this message/group field have value that
			// have extra fields that are not covered by message/group schema
			if c.DisallowAdditionalProperties {
				jsonSchemaType.AdditionalProperties = []byte("false")
			} else {
				jsonSchemaType.AdditionalProperties = []byte("true")
			}
		}

	default:
		return nil, fmt.Errorf("unrecognized field type: %s", desc.GetType().String())
	}

	// Recurse array of primitive types:
	if desc.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED && jsonSchemaType.Type != gojsonschema.TYPE_OBJECT {
		jsonSchemaType.Items = &jsonschema.Type{}

		if len(jsonSchemaType.Enum) > 0 {
			jsonSchemaType.Items.Enum = jsonSchemaType.Enum
			jsonSchemaType.Enum = nil
			jsonSchemaType.Items.OneOf = nil
		} else {
			jsonSchemaType.Items.Type = jsonSchemaType.Type
			jsonSchemaType.Items.Minimum = jsonSchemaType.Minimum
			jsonSchemaType.Items.Maximum = jsonSchemaType.Maximum
			jsonSchemaType.Items.ExclusiveMinimum = jsonSchemaType.ExclusiveMinimum
			jsonSchemaType.Items.OneOf = jsonSchemaType.OneOf
		}

		if c.AllowNullValues {
			jsonSchemaType.OneOf = []*jsonschema.Type{
				{Type: gojsonschema.TYPE_NULL},
				{Type: gojsonschema.TYPE_ARRAY},
			}
		} else {
			jsonSchemaType.Type = gojsonschema.TYPE_ARRAY
			jsonSchemaType.OneOf = []*jsonschema.Type{}
		}
		return jsonSchemaType, nil
	}

	// Recurse nested objects / arrays of objects (if necessary):
	if jsonSchemaType.Type == gojsonschema.TYPE_OBJECT {

		recordType, pkgName, ok := c.lookupType(curPkg, desc.GetTypeName())
		if !ok {
			return nil, fmt.Errorf("no such message type named %s", desc.GetTypeName())
		}

		// Recurse the recordType:
		recursedJSONSchemaType, err := c.recursiveConvertMessageType(curPkg, recordType, pkgName, duplicatedMessages, false)
		if err != nil {
			return nil, err
		}

		// Maps, arrays, and objects are structured in different ways:
		switch {

		// Maps:
		case recordType.Options.GetMapEntry():
			c.logger.
				WithField("field_name", recordType.GetName()).
				WithField("msg_name", *msg.Name).
				Tracef("Is a map")

			// Make sure we have a "value":
			value, valuePresent := recursedJSONSchemaType.Properties.Get("value")
			if !valuePresent {
				return nil, fmt.Errorf("Unable to find 'value' property of MAP type")
			}

			// Marshal the "value" properties to JSON (because that's how we can pass on AdditionalProperties):
			additionalPropertiesJSON, err := json.Marshal(value)
			if err != nil {
				return nil, err
			}
			jsonSchemaType.AdditionalProperties = additionalPropertiesJSON

		// Arrays:
		case desc.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED:
			jsonSchemaType.Items = recursedJSONSchemaType
			jsonSchemaType.Type = gojsonschema.TYPE_ARRAY

			// Build up the list of required fields:
			if c.AllFieldsRequired && recursedJSONSchemaType.Properties != nil {
				for _, property := range recursedJSONSchemaType.Properties.Keys() {
					jsonSchemaType.Items.Required = append(jsonSchemaType.Items.Required, property)
				}
			}

		// Not maps, not arrays:
		default:

			// If we've got optional types then just take those:
			if recursedJSONSchemaType.OneOf != nil {
				return recursedJSONSchemaType, nil
			}

			// If we're not an object then set the type from whatever we recursed:
			if recursedJSONSchemaType.Type != gojsonschema.TYPE_OBJECT {
				jsonSchemaType.Type = recursedJSONSchemaType.Type
			}

			// Assume the attrbutes of the recursed value:
			jsonSchemaType.Properties = recursedJSONSchemaType.Properties
			jsonSchemaType.Ref = recursedJSONSchemaType.Ref
			jsonSchemaType.Required = recursedJSONSchemaType.Required

			// Build up the list of required fields:
			if c.AllFieldsRequired && recursedJSONSchemaType.Properties != nil {
				for _, property := range recursedJSONSchemaType.Properties.Keys() {
					jsonSchemaType.Required = append(jsonSchemaType.Required, property)
				}
			}
		}

		// Optionally allow NULL values:
		if c.AllowNullValues {
			jsonSchemaType.OneOf = []*jsonschema.Type{
				{Type: gojsonschema.TYPE_NULL},
				{Type: jsonSchemaType.Type},
			}
			jsonSchemaType.Type = ""
		}
	}

	jsonSchemaType.Required = dedupe(jsonSchemaType.Required)

	return jsonSchemaType, nil
}

// Converts a proto "MESSAGE" into a JSON-Schema:
func (c *Converter) convertMessageType(curPkg *ProtoPackage, msg *descriptor.DescriptorProto) (*jsonschema.Schema, error) {

	// first, recursively find messages that appear more than once - in particular, that will break cycles
	duplicatedMessages, err := c.findDuplicatedNestedMessages(curPkg, msg)
	if err != nil {
		return nil, err
	}

	// main schema for the message
	rootType, err := c.recursiveConvertMessageType(curPkg, msg, "", duplicatedMessages, false)
	if err != nil {
		return nil, err
	}

	// and then generate the sub-schema for each duplicated message
	definitions := jsonschema.Definitions{}
	for refMsg, name := range duplicatedMessages {
		refType, err := c.recursiveConvertMessageType(curPkg, refMsg, "", duplicatedMessages, true)
		if err != nil {
			return nil, err
		}

		// need to give that schema an ID
		if refType.Extras == nil {
			refType.Extras = make(map[string]interface{})
		}
		refType.Extras["id"] = name
		definitions[name] = refType
	}

	newJSONSchema := &jsonschema.Schema{
		Type:        rootType,
		Definitions: definitions,
	}

	// Look for required fields (either by proto2 required flag, or the AllFieldsRequired option):
	for _, fieldDesc := range msg.GetField() {
		if c.AllFieldsRequired || fieldDesc.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REQUIRED {
			newJSONSchema.Required = append(newJSONSchema.Required, fieldDesc.GetName())
		}
	}

	newJSONSchema.Required = dedupe(newJSONSchema.Required)

	return newJSONSchema, nil
}

// findDuplicatedNestedMessages takes a message, and returns a map mapping pointers to messages that appear more than once
// (typically because they're part of a reference cycle) to the sub-schema name that we give them.
func (c *Converter) findDuplicatedNestedMessages(curPkg *ProtoPackage, msg *descriptor.DescriptorProto) (map[*descriptor.DescriptorProto]string, error) {
	all := make(map[*descriptor.DescriptorProto]*nameAndCounter)
	if err := c.recursiveFindDuplicatedNestedMessages(curPkg, msg, msg.GetName(), all); err != nil {
		return nil, err
	}

	result := make(map[*descriptor.DescriptorProto]string)
	for m, nameAndCounter := range all {
		if nameAndCounter.counter > 1 && !strings.HasPrefix(nameAndCounter.name, ".google.protobuf.") {
			result[m] = strings.TrimLeft(nameAndCounter.name, ".")
		}
	}

	return result, nil
}

type nameAndCounter struct {
	name    string
	counter int
}

func (c *Converter) recursiveFindDuplicatedNestedMessages(curPkg *ProtoPackage, msg *descriptor.DescriptorProto, typeName string, alreadySeen map[*descriptor.DescriptorProto]*nameAndCounter) error {
	if nameAndCounter, present := alreadySeen[msg]; present {
		nameAndCounter.counter++
		return nil
	}
	alreadySeen[msg] = &nameAndCounter{
		name:    typeName,
		counter: 1,
	}

	for _, desc := range msg.GetField() {
		descType := desc.GetType()
		if descType != descriptor.FieldDescriptorProto_TYPE_MESSAGE && descType != descriptor.FieldDescriptorProto_TYPE_GROUP {
			// no nested messages
			continue
		}

		typeName := desc.GetTypeName()
		recordType, _, ok := c.lookupType(curPkg, typeName)
		if !ok {
			return fmt.Errorf("no such message type named %s", typeName)
		}
		if err := c.recursiveFindDuplicatedNestedMessages(curPkg, recordType, typeName, alreadySeen); err != nil {
			return err
		}
	}

	return nil
}

func (c *Converter) recursiveConvertMessageType(curPkg *ProtoPackage, msg *descriptor.DescriptorProto, pkgName string, duplicatedMessages map[*descriptor.DescriptorProto]string, ignoreDuplicatedMessages bool) (*jsonschema.Type, error) {
	// Handle google's well-known types:
	if msg.Name != nil && wellKnownTypes[*msg.Name] && pkgName == ".google.protobuf" {
		var typeSchema *jsonschema.Type
		switch *msg.Name {
		case "DoubleValue", "FloatValue":
			typeSchema = &jsonschema.Type{Type: gojsonschema.TYPE_NUMBER}
		case "Int32Value":
			typeSchema = &jsonschema.Type{
				Type:    gojsonschema.TYPE_INTEGER,
				Minimum: math.MinInt32,
				Maximum: math.MaxInt32,
			}
		case "UInt32Value":
			typeSchema = &jsonschema.Type{
				Type:             gojsonschema.TYPE_INTEGER,
				Minimum:          -1,
				ExclusiveMinimum: true,
				Maximum:          intSafeMaxUint32,
			}
		case "Int64Value":
			typeSchema = &jsonschema.Type{
				Type:    gojsonschema.TYPE_INTEGER,
				Minimum: intSafeMinInt64,
				Maximum: intSafeMaxInt64,
			}
		case "UInt64Value":
			typeSchema = &jsonschema.Type{
				Type:             gojsonschema.TYPE_INTEGER,
				Minimum:          -1,
				ExclusiveMinimum: true,
				Maximum:          intSafeMaxUint64,
			}
		case "BoolValue":
			typeSchema = &jsonschema.Type{Type: gojsonschema.TYPE_BOOLEAN}
		case "BytesValue", "StringValue":
			typeSchema = &jsonschema.Type{Type: gojsonschema.TYPE_STRING}
		case "Value":
			typeSchema = &jsonschema.Type{Type: gojsonschema.TYPE_OBJECT}
		}

		// If we're allowing nulls then prepare a OneOf:
		if c.AllowNullValues {
			return &jsonschema.Type{
				OneOf: []*jsonschema.Type{
					{Type: gojsonschema.TYPE_NULL},
					typeSchema,
				},
			}, nil
		}

		// Otherwise just return this simple type:
		return typeSchema, nil
	}

	if refName, ok := duplicatedMessages[msg]; ok && !ignoreDuplicatedMessages {
		return &jsonschema.Type{
			Version: jsonschema.Version,
			Ref:     refName,
		}, nil
	}

	// Prepare a new jsonschema:
	jsonSchemaType := &jsonschema.Type{
		Properties: orderedmap.New(),
		Version:    jsonschema.Version,
	}

	// Generate a description from src comments (if available)
	if src := c.sourceInfo.GetMessage(msg); src != nil {
		jsonSchemaType.Description = formatDescription(src)
	}

	// Optionally allow NULL values:
	if c.AllowNullValues {
		jsonSchemaType.OneOf = []*jsonschema.Type{
			{Type: gojsonschema.TYPE_NULL},
			{Type: gojsonschema.TYPE_OBJECT},
		}
	} else {
		jsonSchemaType.Type = gojsonschema.TYPE_OBJECT
	}

	// disallowAdditionalProperties will prevent validation where extra fields are found (outside of the schema):
	if c.DisallowAdditionalProperties {
		jsonSchemaType.AdditionalProperties = []byte("false")
	} else {
		jsonSchemaType.AdditionalProperties = []byte("true")
	}

	// create support jsonchema.Type structures for proto oneof fields
	protoOneOfJsonOneOfType := make(map[int32]*jsonschema.Type)
	if len(msg.OneofDecl) == 1 { // single proto oneof in proto message
		jsonSchemaType.PatternProperties = make(map[string]*jsonschema.Type)
		protoOneOfJsonOneOfType[0] = jsonSchemaType
	} else if len(msg.OneofDecl) > 1 { // multiple proto oneof in proto message
		jsonSchemaType.PatternProperties = make(map[string]*jsonschema.Type)
		for i := range msg.OneofDecl {
			jsonOneOfType := &jsonschema.Type{}
			jsonSchemaType.AllOf = append(jsonSchemaType.AllOf, jsonOneOfType)
			protoOneOfJsonOneOfType[int32(i)] = jsonOneOfType
		}
	}

	c.logger.WithField("message_str", proto.MarshalTextString(msg)).Trace("Converting message")
	for _, fieldDesc := range msg.GetField() {
		// get field schema
		recursedJSONSchemaType, err := c.convertField(curPkg, fieldDesc, msg, duplicatedMessages)
		if err != nil {
			c.logger.WithError(err).WithField("field_name", fieldDesc.GetName()).WithField("message_name", msg.GetName()).Error("Failed to convert field")
			return nil, err
		}
		c.logger.WithField("field_name", fieldDesc.GetName()).WithField("type", recursedJSONSchemaType.Type).Trace("Converted field")

		// Figure out which field names we want to use:
		var fieldNames []string
		switch {
		case c.UseJSONFieldnamesOnly:
			fieldNames = append(fieldNames, fieldDesc.GetJsonName())
		case c.UseProtoAndJSONFieldnames:
			fieldNames = append(fieldNames, fieldDesc.GetName())
			fieldNames = append(fieldNames, fieldDesc.GetJsonName())
		default:
			fieldNames = append(fieldNames, fieldDesc.GetName())
		}

		if fieldDesc.OneofIndex != nil { // field is part of proto oneof structure
			for _, fieldName := range fieldNames {
				// allow usage of all proto oneof possible fields without sacrifice of enabling additional properties
				// (additionalProperties to true would allow also other random names fields and that would cause
				// external example generator to create for-vpp-agent-unknown fields that will cause problems
				// in proto parsing)
				jsonSchemaType.PatternProperties[fmt.Sprintf("^%s$", fieldName)] = &jsonschema.Type{}

				// adding additional restriction that allow to use only one of the proto oneof fields
				properties := orderedmap.New()
				properties.Set(fieldName, recursedJSONSchemaType) // apply field schema
				singleOneofUsageCase := &jsonschema.Type{
					Type:       "object",
					Required:   []string{fieldName},
					Properties: properties,
				}
				jsonOneOfType := protoOneOfJsonOneOfType[*fieldDesc.OneofIndex]
				jsonOneOfType.OneOf = append(jsonOneOfType.OneOf, singleOneofUsageCase)
			}
		} else { // normal field
			// apply field schemas
			for _, fieldName := range fieldNames {
				jsonSchemaType.Properties.Set(fieldName, recursedJSONSchemaType)
			}

			// Look for required fields (either by proto2 required flag, or the AllFieldsRequired option):
			if fieldDesc.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REQUIRED {
				jsonSchemaType.Required = append(jsonSchemaType.Required, fieldDesc.GetName())
			}
		}
	}

	// Remove empty properties to keep the final output as clean as possible:
	if len(jsonSchemaType.Properties.Keys()) == 0 {
		jsonSchemaType.Properties = nil
	}

	return jsonSchemaType, nil
}

func formatDescription(sl *descriptor.SourceCodeInfo_Location) string {
	var lines []string
	for _, str := range sl.GetLeadingDetachedComments() {
		if s := strings.TrimSpace(str); s != "" {
			lines = append(lines, s)
		}
	}
	if s := strings.TrimSpace(sl.GetLeadingComments()); s != "" {
		lines = append(lines, s)
	}
	if s := strings.TrimSpace(sl.GetTrailingComments()); s != "" {
		lines = append(lines, s)
	}
	return strings.Join(lines, "\n\n")
}

func dedupe(inputStrings []string) []string {
	appended := make(map[string]bool)
	outputStrings := []string{}

	for _, inputString := range inputStrings {
		if !appended[inputString] {
			outputStrings = append(outputStrings, inputString)
			appended[inputString] = true
		}
	}
	return outputStrings
}
