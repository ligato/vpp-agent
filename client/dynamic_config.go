package client

import (
	"fmt"
	"strings"

	"github.com/goccy/go-yaml"
	"go.ligato.io/cn-infra/v2/logging/logrus"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"

	"go.ligato.io/vpp-agent/v3/pkg/models"
	"go.ligato.io/vpp-agent/v3/pkg/util"
	"go.ligato.io/vpp-agent/v3/proto/ligato/generic"
)

// field proto name/json name constants (can't be changes to not break json/yaml compatibility with configurator.Config)
const (
	// configGroupSuffix is field proto name suffix that all fields referencing config groups has
	configGroupSuffix = "Config"
	// repeatedFieldsSuffix is suffix added to repeated fields inside config group message
	repeatedFieldsSuffix = "_list"
)

// names is supporting structure for remembering proto field name and json name
type names struct {
	protoName, jsonName string
}

// TODO: generate backwardCompatibleNames dynamically by searching given known model in configurator.Config
//  and extracting proto field name and json name?

// backwardCompatibleNames is mappging from dynamic Config fields (derived from currently known models) to
// hardcoded names (proto field name/json name) in hardcoded configurator.Config. This mapping should allow
// dynamically-created Config to read/write configuration from/to json/yaml files in the same way as it is
// for hardcoded configurator.Config.
var backwardCompatibleNames = map[string]names{
	"netallocConfig.IPAllocation":      names{protoName: "ip_addresses", jsonName: "ipAddresses"},
	"linuxConfig.Interface":            names{protoName: "interfaces", jsonName: "interfaces"},
	"linuxConfig.ARPEntry":             names{protoName: "arp_entries", jsonName: "arpEntries"},
	"linuxConfig.Route":                names{protoName: "routes", jsonName: "routes"},
	"linuxConfig.RuleChain":            names{protoName: "RuleChain", jsonName: "RuleChain"},
	"vppConfig.ABF":                    names{protoName: "abfs", jsonName: "abfs"},
	"vppConfig.ACL":                    names{protoName: "acls", jsonName: "acls"},
	"vppConfig.SecurityPolicyDatabase": names{protoName: "ipsec_spds", jsonName: "ipsecSpds"},
	"vppConfig.SecurityPolicy":         names{protoName: "ipsec_sps", jsonName: "ipsecSps"},
	"vppConfig.SecurityAssociation":    names{protoName: "ipsec_sas", jsonName: "ipsecSas"},
	"vppConfig.TunnelProtection":       names{protoName: "ipsec_tunnel_protections", jsonName: "ipsecTunnelProtections"},
	"vppConfig.Interface":              names{protoName: "interfaces", jsonName: "interfaces"},
	"vppConfig.Span":                   names{protoName: "spans", jsonName: "spans"},
	"vppConfig.IPFIX":                  names{protoName: "ipfix_global", jsonName: "ipfixGlobal"},
	"vppConfig.FlowProbeParams":        names{protoName: "ipfix_flowprobe_params", jsonName: "ipfixFlowprobeParams"},
	"vppConfig.FlowProbeFeature":       names{protoName: "ipfix_flowprobes", jsonName: "ipfixFlowprobes"},
	"vppConfig.BridgeDomain":           names{protoName: "bridge_domains", jsonName: "bridgeDomains"},
	"vppConfig.FIBEntry":               names{protoName: "fibs", jsonName: "fibs"},
	"vppConfig.XConnectPair":           names{protoName: "xconnect_pairs", jsonName: "xconnectPairs"},
	"vppConfig.ARPEntry":               names{protoName: "arps", jsonName: "arps"},
	"vppConfig.Route":                  names{protoName: "routes", jsonName: "routes"},
	"vppConfig.ProxyARP":               names{protoName: "proxy_arp", jsonName: "proxyArp"},
	"vppConfig.IPScanNeighbor":         names{protoName: "ipscan_neighbor", jsonName: "ipscanNeighbor"},
	"vppConfig.VrfTable":               names{protoName: "vrfs", jsonName: "vrfs"},
	"vppConfig.DHCPProxy":              names{protoName: "dhcp_proxies", jsonName: "dhcpProxies"},
	"vppConfig.L3XConnect":             names{protoName: "l3xconnects", jsonName: "l3xconnects"},
	"vppConfig.TeibEntry":              names{protoName: "teib_entries", jsonName: "teibEntries"},
	"vppConfig.Nat44Global":            names{protoName: "nat44_global", jsonName: "nat44Global"},
	"vppConfig.DNat44":                 names{protoName: "dnat44s", jsonName: "dnat44s"},
	"vppConfig.Nat44Interface":         names{protoName: "nat44_interfaces", jsonName: "nat44Interfaces"},
	"vppConfig.Nat44AddressPool":       names{protoName: "nat44_pools", jsonName: "nat44Pools"},
	"vppConfig.IPRedirect":             names{protoName: "punt_ipredirects", jsonName: "puntIpredirects"},
	"vppConfig.ToHost":                 names{protoName: "punt_tohosts", jsonName: "puntTohosts"},
	"vppConfig.Exception":              names{protoName: "punt_exceptions", jsonName: "puntExceptions"},
	"vppConfig.LocalSID":               names{protoName: "srv6_localsids", jsonName: "srv6Localsids"},
	"vppConfig.Policy":                 names{protoName: "srv6_policies", jsonName: "srv6Policies"},
	"vppConfig.Steering":               names{protoName: "srv6_steerings", jsonName: "srv6Steerings"},
	"vppConfig.SRv6Global":             names{protoName: "srv6_global", jsonName: "srv6Global"},
}

// NewDynamicConfig creates dynamically proto Message that contains all given configuration models(knowModels).
// This proto message(when all VPP-Agent models are given as input) is json/yaml compatible with
// configurator.Config. The configurator.Config config have all models hardcoded (generated from config
// proto model, but that model is hardcoded). Dynamic config can contain also custom 3rd party models
// and therefore can be used to import/export config data also for 3rd party models that are registered, but not
// part of VPP-Agent repository and therefore not know to hardcoded configurator.Config.
func NewDynamicConfig(knownModels []*models.ModelInfo) (*dynamicpb.Message, error) {
	// create dependency registry
	dependencyRegistry, err := createFileDescRegistry(knownModels)
	if err != nil {
		return nil, fmt.Errorf("cannot create dependency file descriptor registry due to: %w", err)
	}

	// get file descriptor proto for give known models
	fileDP, rootMsgName, err := createDynamicConfigDescriptorProto(knownModels, dependencyRegistry)
	if err != nil {
		return nil, fmt.Errorf("cannot create descriptor proto for dynamic config due to: %v", err)
	}

	// convert file descriptor proto to file descriptor
	fd, err := protodesc.NewFile(fileDP, dependencyRegistry)
	if err != nil {
		return nil, fmt.Errorf("cannot convert file descriptor proto to file descriptor due to: %v", err)
	}

	// get descriptor for config root message
	rootMsg := fd.Messages().ByName(rootMsgName)

	// create dynamic config message
	return dynamicpb.NewMessage(rootMsg), nil
}

// createFileDescRegistry extracts file descriptors from given known models and returns them in convenient
// registry (in form of protodesc.Resolver).
func createFileDescRegistry(knownModels []*models.ModelInfo) (protodesc.Resolver, error) {
	reg := &protoregistry.Files{}
	for _, knownModel := range knownModels {
		fileDesc := knownModel.MessageDescriptor.ParentFile()
		if _, err := reg.FindDescriptorByName(fileDesc.FullName()); err == protoregistry.NotFound {
			if e := reg.RegisterFile(fileDesc); e != nil {
				logrus.DefaultLogger().Debugf("Failed to add Proto file %v "+
					"to dependency registry: %v.", fileDesc.Path(), e)
			} else {
				logrus.DefaultLogger().Debugf("Proto file %v was successfully "+
					"added to dependency registry.", fileDesc.Path())
			}
		}
	}
	return reg, nil
}

// createDynamicConfigDescriptorProto creates descriptor proto for configuration. The construction of the descriptor
// proto is the way how the configuration from known models are added to the configuration proto message.
// The constructed file descriptor proto is used to get file descriptor that in turn can be used to instantiate
// proto message with all the configs from knownModels. This method conveniently provides also the configuration
// root message (proto file has many messages, but we need to know which one is the root for our configuration).
func createDynamicConfigDescriptorProto(knownModels []*ModelInfo, dependencyRegistry protodesc.Resolver) (
	fileDP *descriptorpb.FileDescriptorProto, rootMsgName protoreflect.Name, error error) {

	// file descriptor proto for dynamic config proto model
	fileDP = &descriptorpb.FileDescriptorProto{
		Syntax:  proto.String("proto3"),
		Name:    proto.String("ligato/configurator/dynamicconfigurator.proto"),
		Package: proto.String("ligato.configurator"),
	}

	// create config message
	configDP := &descriptorpb.DescriptorProto{
		Name: proto.String("Dynamic_config"),
	}

	// add config message to proto file
	fileDP.MessageType = []*descriptorpb.DescriptorProto{configDP}

	// define configuration root (for users of this function)
	rootMsgName = protoreflect.Name(*(configDP.Name))

	// fill dynamic message with given known models
	configGroups := make(map[string]*descriptorpb.DescriptorProto)
	importedDependency := make(map[string]struct{}) // just for deduplication checking
	for _, modelDetail := range knownModels {
		// get/create group config for this know model (all configs are grouped into groups based on their module)
		configGroupName := DynamicConfigGroupFieldNaming(modelDetail)
		configGroup, found := configGroups[configGroupName]
		if !found { // create it!
			// create new message that represents new config group
			configGroup = &descriptorpb.DescriptorProto{
				Name: proto.String(configGroupName),
			}

			// add config group message to message definitions
			fileDP.MessageType = append(fileDP.MessageType, configGroup)

			// create reference to the new config group message from main config message
			configDP.Field = append(configDP.Field, &descriptorpb.FieldDescriptorProto{
				Name:     proto.String(configGroupName),
				Number:   proto.Int32(int32(len(configDP.Field) + 1)),
				JsonName: proto.String(configGroupName),
				Type:     protoType(descriptorpb.FieldDescriptorProto_TYPE_MESSAGE),
				TypeName: proto.String(fmt.Sprintf(".%v.%v", *fileDP.Package, *configGroup.Name)),
				Label:    protoLabel(descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL),
			})

			// cache config group for reuse by other known models
			configGroups[configGroupName] = configGroup
		}

		// fill config group message with currently handled known model
		label := protoLabel(descriptorpb.FieldDescriptorProto_LABEL_REPEATED)
		if !existsModelOptionFor("nameTemplate", modelDetail.Options) {
			label = protoLabel(descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL)
		}
		protoName, jsonName := DynamicConfigKnownModelFieldNaming(modelDetail)
		configGroup.Field = append(configGroup.Field, &descriptorpb.FieldDescriptorProto{
			Name:     proto.String(protoName),
			Number:   proto.Int32(int32(len(configGroup.Field) + 1)),
			JsonName: proto.String(jsonName),
			Label:    label,
			Type:     protoType(descriptorpb.FieldDescriptorProto_TYPE_MESSAGE),
			TypeName: proto.String(fmt.Sprintf(".%v", modelDetail.ProtoName)),
		})

		// add proto file dependency for this known model (+ check that it is in dependency file descriptor registry)
		protoFile, err := models.ModelOptionFor("protoFile", modelDetail.Options)
		if err != nil {
			error = fmt.Errorf("cannot retrieve protoFile from model options "+
				"from model %v due to: %v", modelDetail.ProtoName, err)
			return
		}
		if _, found := importedDependency[protoFile]; !found {
			importedDependency[protoFile] = struct{}{}

			// add proto file dependency for this known model
			fileDP.Dependency = append(fileDP.Dependency, protoFile)

			// checking dependency registry that should already contain the linked dependency
			if _, err := dependencyRegistry.FindFileByPath(protoFile); err != nil {
				if err == protoregistry.NotFound {
					error = fmt.Errorf("proto file %v need to be referenced in dynamic config, but it "+
						"is not in dependency registry that was created from file descriptor proto input "+
						"(missing in input? check debug output from creating dependency registry) ", protoFile)
					return
				}
				error = fmt.Errorf("cannot verify that proto file %v is in "+
					"dependency registry, it is due to: %v", protoFile, err)
				return
			}
		}
	}
	return
}

// DynamicConfigGroupFieldNaming computes for given known model the naming of configuration group proto field
// containing the instances of given model inside the dynamic config describing the whole VPP-Agent configuration.
// The json name of the field is the same as proto name of field.
func DynamicConfigGroupFieldNaming(modelDetail *models.ModelInfo) string {
	return fmt.Sprintf("%v%v", modulePrefix(models.ToSpec(modelDetail.Spec).ModelName()), configGroupSuffix)
}

// DynamicConfigKnownModelFieldNaming compute for given known model the (proto and json) naming of proto field
// containing the instances of given model inside the dynamic config describing the whole VPP-Agent configuration.
func DynamicConfigKnownModelFieldNaming(modelDetail *models.ModelInfo) (protoName, jsonName string) {
	simpleProtoName := simpleProtoName(modelDetail.ProtoName)
	configGroupName := DynamicConfigGroupFieldNaming(modelDetail)
	compatibilityKey := fmt.Sprintf("%v.%v", configGroupName, simpleProtoName)

	if newNames, found := backwardCompatibleNames[compatibilityKey]; found {
		// using field names from hardcoded configurator.Config to achieve json/yaml backward compatibility
		protoName = newNames.protoName
		jsonName = newNames.jsonName
	} else if !existsModelOptionFor("nameTemplate", modelDetail.Options) {
		protoName = simpleProtoName
		jsonName = simpleProtoName
	} else {
		protoName = simpleProtoName + repeatedFieldsSuffix
		jsonName = simpleProtoName + repeatedFieldsSuffix
	}

	return protoName, jsonName
}

// DynamicConfigExport exports from dynamic config the proto.Messages corresponding to known models that
// were given as input when dynamic config was created using NewDynamicConfig. This is a convenient
// method how to extract data for generic client usage (proto.Message instances) from value-filled
// dynamic config (i.e. after json/yaml loading into dynamic config).
func DynamicConfigExport(dynamicConfig *dynamicpb.Message) ([]proto.Message, error) {
	if dynamicConfig == nil {
		return nil, fmt.Errorf("dynamic config cannot be nil")
	}

	// iterate over config group messages and extract proto message from them
	result := make([]proto.Message, 0)
	fields := dynamicConfig.Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		fieldName := fields.Get(i).Name()
		if strings.HasSuffix(string(fieldName), configGroupSuffix) {
			configGroupMessage := dynamicConfig.Get(fields.Get(i)).Message()

			// handling export from inner config layers by using helper method
			result = append(result, exportFromConfigGroupMessage(configGroupMessage)...)
		}
	}
	return result, nil
}

// ExportDynamicConfigStructure is a debugging helper function revealing current structure of dynamic config.
// Debugging tools can't reveal that because dynamic config is dynamic proto message with no fields named by
// proto fields as it is in generated proto messages.
func ExportDynamicConfigStructure(dynamicConfig proto.Message) (string, error) {
	// fill dynamic message with nothing (one proto message that will not map to anything), but relaying
	// on side effect that will fill the structure with empty messages
	anyProtoMessage := []proto.Message{&generic.Item{}}
	util.PlaceProtosIntoProtos(anyProtoMessage, 1000, dynamicConfig)

	// export dynamic config to json and then into yaml format
	m := protojson.MarshalOptions{
		Indent: "",
		// this will also fill non-Message fields (Message fields are filled by util.PlaceProtosIntoProtos side effect)
		EmitUnpopulated: true,
	}
	b, err := m.Marshal(dynamicConfig)
	if err != nil {
		return "", fmt.Errorf("cannot marshal dynamic config to json due to: %v", err)
	}
	var jsonObj interface{}
	err = yaml.UnmarshalWithOptions(b, &jsonObj, yaml.UseOrderedMap())
	if err != nil {
		return "", fmt.Errorf("cannot convert dynamic config's json bytes to "+
			"json struct for yaml marshalling due to: %v", err)
	}
	bb, err := yaml.Marshal(jsonObj)
	if err != nil {
		return "", fmt.Errorf("cannot marshal dynamic config from json to yaml due to: %v", err)
	}
	return string(bb), nil
}

// exportFromConfigGroupMessage exports proto messages from config group message layer of dynamic config
func exportFromConfigGroupMessage(configGroupMessage protoreflect.Message) []proto.Message {
	result := make([]proto.Message, 0)
	fields := configGroupMessage.Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		groupField := fields.Get(i)
		if groupField.IsList() { // repeated field
			repeatedValue := configGroupMessage.Get(groupField).List()
			for j := 0; j < repeatedValue.Len(); j++ {
				result = append(result, repeatedValue.Get(j).Message().Interface())
			}
		} else { // optional field (there are only optional and repeated fields)
			fieldValue := configGroupMessage.Get(groupField)
			if fieldValue.Message().IsValid() {
				// use only non-nil real value (validity check used for this is implementation
				// dependent, but there seems to be no other way)
				result = append(result, fieldValue.Message().Interface())
			}
		}
	}
	return result
}

func simpleProtoName(fullProtoName string) string {
	nameSplit := strings.Split(fullProtoName, ".")
	return nameSplit[len(nameSplit)-1]
}

func existsModelOptionFor(key string, options []*generic.ModelDetail_Option) bool {
	_, err := models.ModelOptionFor(key, options)
	return err == nil
}

func modulePrefix(modelName string) string {
	return strings.Split(modelName, ".")[0] // modelname = modulname(it has modulname prefix) + simple name of model
}

func protoType(typ descriptorpb.FieldDescriptorProto_Type) *descriptorpb.FieldDescriptorProto_Type {
	return &typ
}

func protoLabel(label descriptorpb.FieldDescriptorProto_Label) *descriptorpb.FieldDescriptorProto_Label {
	return &label
}
