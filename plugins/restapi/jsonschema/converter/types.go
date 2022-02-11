package converter

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/alecthomas/jsonschema"
	"github.com/iancoleman/orderedmap"
	"github.com/xeipuuv/gojsonschema"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"go.ligato.io/vpp-agent/v3/proto/ligato"
)

const (
	PatternIpv6WithMask = "^(::|(([a-fA-F0-9]{1,4}):){7}(([a-fA-F0-9]{1,4}))|(:(:([a-fA-F0-9]{1,4})){1,6})|((([a-fA-F0-9]{1,4}):){1,6}:)|((([a-fA-F0-9]{1,4}):)(:([a-fA-F0-9]{1,4})){1,6})|((([a-fA-F0-9]{1,4}):){2}(:([a-fA-F0-9]{1,4})){1,5})|((([a-fA-F0-9]{1,4}):){3}(:([a-fA-F0-9]{1,4})){1,4})|((([a-fA-F0-9]{1,4}):){4}(:([a-fA-F0-9]{1,4})){1,3})|((([a-fA-F0-9]{1,4}):){5}(:([a-fA-F0-9]{1,4})){1,2}))(\\\\/(12[0-8]|1[0-1][0-9]|[1-9][0-9]|[0-9]))$"
	PatternIpv4WithMask = "^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])(/(3[0-2]|[1-2][0-9]|[0-9]))$"
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

func (c *Converter) registerEnum(pkgName *string, enum *descriptorpb.EnumDescriptorProto) {
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

func (c *Converter) registerType(pkgName *string, msg *descriptorpb.DescriptorProto) {
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

// applyAllowNullValuesOption applies schema changes to schema while handling possibility of use Null values
// (if enabled). This is a convenience method for handling the NULL values option.
func (c *Converter) applyAllowNullValuesOption(schema *jsonschema.Type, schemaChanges *jsonschema.Type) {
	if c.AllowNullValues { // insert possibility of using NULL type
		if len(schemaChanges.OneOf) == 0 {
			schema.OneOf = []*jsonschema.Type{
				{
					Type: gojsonschema.TYPE_NULL,
				}, {
					Type:             schemaChanges.Type,
					Format:           schemaChanges.Format,
					Pattern:          schemaChanges.Pattern,
					Minimum:          schemaChanges.Minimum,
					ExclusiveMinimum: schemaChanges.ExclusiveMinimum,
					Maximum:          schemaChanges.Maximum,
					ExclusiveMaximum: schemaChanges.ExclusiveMaximum,
				},
			}
		} else {
			schema.OneOf = append([]*jsonschema.Type{
				{Type: gojsonschema.TYPE_NULL},
			}, schemaChanges.OneOf...)
		}
	} else { // direct mapping (schema could be already partially built -> need to fill new values into it)
		schema.Type = schemaChanges.Type
		schema.Format = schemaChanges.Format
		schema.Pattern = schemaChanges.Pattern
		schema.Minimum = schemaChanges.Minimum
		schema.ExclusiveMinimum = schemaChanges.ExclusiveMinimum
		schema.Maximum = schemaChanges.Maximum
		schema.ExclusiveMaximum = schemaChanges.ExclusiveMaximum
		schema.OneOf = schemaChanges.OneOf
	}
}

// Convert a proto "field" (essentially a type-switch with some recursion):
func (c *Converter) convertField(curPkg *ProtoPackage, desc *descriptorpb.FieldDescriptorProto, msg *descriptorpb.DescriptorProto, duplicatedMessages map[*descriptorpb.DescriptorProto]string) (*jsonschema.Type, error) {
	// Prepare a new jsonschema.Type for our eventual return value:
	jsonSchemaType := &jsonschema.Type{}

	// Generate a description from src comments (if available)
	if src := c.sourceInfo.GetField(desc); src != nil {
		jsonSchemaType.Description = formatDescription(src)
	}

	c.logger.Tracef("(PKG: %v) CONVERT FIELD %v", curPkg.name, desc)

	// get field annotations
	var fieldAnnotations *ligato.LigatoOptions

	if proto.HasExtension(desc.Options, ligato.E_LigatoOptions) {
		val := proto.GetExtension(desc.Options, ligato.E_LigatoOptions)
		var ok bool
		if fieldAnnotations, ok = val.(*ligato.LigatoOptions); !ok {
			c.logger.Debugf("Field %s.%s have ligato option extension, but its value has "+
				"unexpected type (%T)", msg.GetName(), desc.GetName(), val)
		}
	} else {
		c.logger.Debugf("Field %s.%s doesn't have ligato option extension", msg.GetName(), desc.GetName())
	}

	// Switch the types, and pick a JSONSchema equivalent:
	switch desc.GetType() {
	case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE,
		descriptorpb.FieldDescriptorProto_TYPE_FLOAT:
		c.applyAllowNullValuesOption(jsonSchemaType, &jsonschema.Type{Type: gojsonschema.TYPE_NUMBER})

	case descriptorpb.FieldDescriptorProto_TYPE_INT32,
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED32,
		descriptorpb.FieldDescriptorProto_TYPE_SINT32:
		schema := &jsonschema.Type{
			Type:    gojsonschema.TYPE_INTEGER,
			Minimum: math.MinInt32,
			Maximum: math.MaxInt32,
		}
		c.applyIntRangeFieldAnnotation(fieldAnnotations, schema)
		c.applyAllowNullValuesOption(jsonSchemaType, schema)

	case descriptorpb.FieldDescriptorProto_TYPE_UINT32,
		descriptorpb.FieldDescriptorProto_TYPE_FIXED32:
		schema := &jsonschema.Type{
			Type:             gojsonschema.TYPE_INTEGER,
			Minimum:          -1,
			ExclusiveMinimum: true,
			Maximum:          intSafeMaxUint32,
		}
		c.applyIntRangeFieldAnnotation(fieldAnnotations, schema)
		c.applyAllowNullValuesOption(jsonSchemaType, schema)

	case descriptorpb.FieldDescriptorProto_TYPE_INT64,
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED64,
		descriptorpb.FieldDescriptorProto_TYPE_SINT64:
		if !c.DisallowBigIntsAsStrings {
			c.applyAllowNullValuesOption(jsonSchemaType, &jsonschema.Type{Type: gojsonschema.TYPE_STRING})
		} else {
			schema := &jsonschema.Type{
				Type:    gojsonschema.TYPE_INTEGER,
				Minimum: intSafeMinInt64,
				Maximum: intSafeMaxInt64,
			}
			c.applyIntRangeFieldAnnotation(fieldAnnotations, schema)
			c.applyAllowNullValuesOption(jsonSchemaType, schema)
		}

	case descriptorpb.FieldDescriptorProto_TYPE_UINT64,
		descriptorpb.FieldDescriptorProto_TYPE_FIXED64:
		if !c.DisallowBigIntsAsStrings {
			c.applyAllowNullValuesOption(jsonSchemaType, &jsonschema.Type{Type: gojsonschema.TYPE_STRING})
		} else {
			schema := &jsonschema.Type{
				Type:             gojsonschema.TYPE_INTEGER,
				Minimum:          -1,
				ExclusiveMinimum: true,
				Maximum:          intSafeMaxUint64,
			}
			c.applyIntRangeFieldAnnotation(fieldAnnotations, schema)
			c.applyAllowNullValuesOption(jsonSchemaType, schema)
		}

	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		schema := &jsonschema.Type{}
		switch fieldAnnotations.GetType() {
		case ligato.LigatoOptions_IPV6:
			schema.Type = gojsonschema.TYPE_STRING
			schema.Format = "ipv6"
		case ligato.LigatoOptions_IPV4:
			schema.Type = gojsonschema.TYPE_STRING
			schema.Format = "ipv4"
		case ligato.LigatoOptions_IP:
			schema.OneOf = []*jsonschema.Type{
				{
					Type:   gojsonschema.TYPE_STRING,
					Format: "ipv4",
				},
				{
					Type:   gojsonschema.TYPE_STRING,
					Format: "ipv6",
				},
			}
		case ligato.LigatoOptions_IPV4_WITH_MASK:
			schema.Type = gojsonschema.TYPE_STRING
			schema.Pattern = PatternIpv4WithMask
		case ligato.LigatoOptions_IPV6_WITH_MASK:
			schema.Type = gojsonschema.TYPE_STRING
			schema.Pattern = PatternIpv6WithMask
		case ligato.LigatoOptions_IP_WITH_MASK:
			schema.OneOf = []*jsonschema.Type{
				{
					Type:    gojsonschema.TYPE_STRING,
					Pattern: PatternIpv4WithMask,
				},
				{
					Type:    gojsonschema.TYPE_STRING,
					Pattern: PatternIpv6WithMask,
				},
			}
		case ligato.LigatoOptions_IPV4_OPTIONAL_MASK:
			schema.OneOf = []*jsonschema.Type{
				{
					Type:   gojsonschema.TYPE_STRING,
					Format: "ipv4",
				},
				{
					Type:    gojsonschema.TYPE_STRING,
					Pattern: PatternIpv4WithMask,
				},
			}
		case ligato.LigatoOptions_IPV6_OPTIONAL_MASK:
			schema.OneOf = []*jsonschema.Type{
				{
					Type:   gojsonschema.TYPE_STRING,
					Format: "ipv6",
				},
				{
					Type:    gojsonschema.TYPE_STRING,
					Pattern: PatternIpv6WithMask,
				},
			}
		case ligato.LigatoOptions_IP_OPTIONAL_MASK:
			schema.OneOf = []*jsonschema.Type{
				{
					Type:   gojsonschema.TYPE_STRING,
					Format: "ipv4",
				},
				{
					Type:    gojsonschema.TYPE_STRING,
					Pattern: PatternIpv4WithMask,
				},
				{
					Type:   gojsonschema.TYPE_STRING,
					Format: "ipv6",
				},
				{
					Type:    gojsonschema.TYPE_STRING,
					Pattern: PatternIpv6WithMask,
				},
			}
		default: // no annotations or annotation used are not applicable here
			schema.Type = gojsonschema.TYPE_STRING
		}
		c.applyAllowNullValuesOption(jsonSchemaType, schema)

	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		c.applyAllowNullValuesOption(jsonSchemaType, &jsonschema.Type{Type: gojsonschema.TYPE_STRING})

	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		// Note: not setting type specification(oneof string and integer), because explicitly saying which
		// values are valid (and any other is invalid) is enough specification what can be used
		// (this also overcome bug in example creator https://json-schema-faker.js.org/ that doesn't select
		// correct type for enum value but rather chooses random type from oneof and cast value to that type)
		//
		// jsonSchemaType.OneOf = append(jsonSchemaType.OneOf, &jsonschema.Type{Type: gojsonschema.TYPE_STRING})
		// jsonSchemaType.OneOf = append(jsonSchemaType.OneOf, &jsonschema.Type{Type: gojsonschema.TYPE_INTEGER})
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

	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		c.applyAllowNullValuesOption(jsonSchemaType, &jsonschema.Type{Type: gojsonschema.TYPE_BOOLEAN})

	case descriptorpb.FieldDescriptorProto_TYPE_GROUP, descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
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
	if desc.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED && jsonSchemaType.Type != gojsonschema.TYPE_OBJECT {
		jsonSchemaType.Items = &jsonschema.Type{}

		if len(jsonSchemaType.Enum) > 0 {
			jsonSchemaType.Items.Enum = jsonSchemaType.Enum
			jsonSchemaType.Enum = nil
			jsonSchemaType.Items.OneOf = nil
		} else { // move schema of primitive type to item schema
			// copy
			jsonSchemaType.Items.Type = jsonSchemaType.Type
			jsonSchemaType.Items.Format = jsonSchemaType.Format
			jsonSchemaType.Items.Minimum = jsonSchemaType.Minimum
			jsonSchemaType.Items.Maximum = jsonSchemaType.Maximum
			jsonSchemaType.Items.ExclusiveMinimum = jsonSchemaType.ExclusiveMinimum
			jsonSchemaType.Items.OneOf = jsonSchemaType.OneOf

			// cleanup
			jsonSchemaType.Type = ""
			jsonSchemaType.Format = ""
			jsonSchemaType.Minimum = 0
			jsonSchemaType.Maximum = 0
			jsonSchemaType.ExclusiveMinimum = false
			jsonSchemaType.OneOf = nil
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
		case desc.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED:
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
			if jsonSchemaType.Ref != "" {
				// clean some fields because usage of REF makes them unnecessary (and in some validator
				// implementation it cause problems/warnings)
				jsonSchemaType.AdditionalProperties = []byte{}
			}
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

// applyIntRangeFieldAnnotation applies new int range for int schema (if the annotation is present)
func (c *Converter) applyIntRangeFieldAnnotation(fieldAnnotations *ligato.LigatoOptions, schema *jsonschema.Type) {
	if fieldAnnotations.GetIntRange() != nil {
		// correct value due for "exclusive" boundary usage
		correctedMinimum := schema.Minimum
		correctedMaximum := schema.Maximum
		if schema.ExclusiveMinimum {
			correctedMinimum = schema.Minimum + 1
		}
		if schema.ExclusiveMaximum {
			correctedMaximum = schema.Maximum - 1
		}

		// compute new range
		schema.Minimum = int(math.Max(float64(fieldAnnotations.GetIntRange().Minimum), float64(correctedMinimum)))
		schema.Maximum = int(math.Min(float64(fieldAnnotations.GetIntRange().Maximum), float64(correctedMaximum)))
		schema.ExclusiveMinimum = false
		schema.ExclusiveMaximum = false

		// apply workaround for 'omitempty' problem (default value is omitted from jsonschema marshaling and
		// the boundary is missing in generated schema)
		if schema.Minimum == 0 {
			schema.Minimum = -1
			schema.ExclusiveMinimum = true
		}
		if schema.Maximum == 0 {
			schema.Maximum = 1
			schema.ExclusiveMaximum = true
		}
	}
}

// Converts a proto "MESSAGE" into a JSON-Schema:
func (c *Converter) convertMessageType(curPkg *ProtoPackage, msg *descriptorpb.DescriptorProto) (*jsonschema.Schema, error) {

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

	// Look for required fields (either by proto required flag, or the AllFieldsRequired option):
	for _, fieldDesc := range msg.GetField() {
		if c.AllFieldsRequired || fieldDesc.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REQUIRED {
			newJSONSchema.Required = append(newJSONSchema.Required, fieldDesc.GetName())
		}
	}

	newJSONSchema.Required = dedupe(newJSONSchema.Required)

	return newJSONSchema, nil
}

// findDuplicatedNestedMessages takes a message, and returns a map mapping pointers to messages that appear more than once
// (typically because they're part of a reference cycle) to the sub-schema name that we give them.
func (c *Converter) findDuplicatedNestedMessages(curPkg *ProtoPackage, msg *descriptorpb.DescriptorProto) (map[*descriptorpb.DescriptorProto]string, error) {
	all := make(map[*descriptorpb.DescriptorProto]*nameAndCounter)
	if err := c.recursiveFindDuplicatedNestedMessages(curPkg, msg, msg.GetName(), all); err != nil {
		return nil, err
	}

	result := make(map[*descriptorpb.DescriptorProto]string)
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

func (c *Converter) recursiveFindDuplicatedNestedMessages(curPkg *ProtoPackage, msg *descriptorpb.DescriptorProto, typeName string, alreadySeen map[*descriptorpb.DescriptorProto]*nameAndCounter) error {
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
		if descType != descriptorpb.FieldDescriptorProto_TYPE_MESSAGE && descType != descriptorpb.FieldDescriptorProto_TYPE_GROUP {
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

func (c *Converter) recursiveConvertMessageType(curPkg *ProtoPackage, msg *descriptorpb.DescriptorProto, pkgName string, duplicatedMessages map[*descriptorpb.DescriptorProto]string, ignoreDuplicatedMessages bool) (*jsonschema.Type, error) {
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

	c.logger.WithField("message_str", prototext.Format(msg)).Trace("Converting message")
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

			// Look for required fields (either by proto required flag, or the AllFieldsRequired option):
			if fieldDesc.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REQUIRED {
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

func formatDescription(sl *descriptorpb.SourceCodeInfo_Location) string {
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
