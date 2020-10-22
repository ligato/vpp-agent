package client

import (
	"fmt"
	"strings"

	"go.ligato.io/cn-infra/v2/logging/logrus"

	"github.com/go-errors/errors"
	"go.ligato.io/vpp-agent/v3/pkg/models"
	"go.ligato.io/vpp-agent/v3/proto/ligato/generic"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

// field proto name/json name constants (can't be changes to not break json/yaml compatibility with configurator.Config)
const (
	// configName is Name of field in fake config root Message that hold the real config root
	configName = "config"
	// configGroupSuffix is field proto name suffix that all fields referencing config groups has
	configGroupSuffix = "_config"
)

// names is supporting structure for remembering proto field name and json name
type names struct {
	protoName, jsonName string
}

// TOOD: generate backwardCompatibleNames dynamically by searching given known model in configurator.Config
//  and extracting proto field name and json name?

// backwardCompatibleNames is mappging from dynamic Config fields (derived from currently known models) to
// hardcoded names (proto field name/json name) in hardcoded configurator.Config. This mapping should allow
// dynamically-created Config to read/write configuration from/to json/yaml files in the same way as it is
// for hardcoded configurator.Config.
var backwardCompatibleNames = map[string]names{
	"netalloc_config.IPAllocation":      names{protoName: "ip_addresses", jsonName: "ipAddresses"},
	"linux_config.Interface":            names{protoName: "interfaces", jsonName: "interfaces"},
	"linux_config.ARPEntry":             names{protoName: "arp_entries", jsonName: "arpEntries"},
	"linux_config.Route":                names{protoName: "routes", jsonName: "routes"},
	"linux_config.RuleChain":            names{protoName: "RuleChain", jsonName: "RuleChain"},
	"vpp_config.ABF":                    names{protoName: "abfs", jsonName: "abfs"},
	"vpp_config.ACL":                    names{protoName: "acls", jsonName: "acls"},
	"vpp_config.SecurityPolicyDatabase": names{protoName: "ipsec_spds", jsonName: "ipsecSpds"},
	"vpp_config.SecurityPolicy":         names{protoName: "ipsec_sps", jsonName: "ipsecSps"},
	"vpp_config.SecurityAssociation":    names{protoName: "ipsec_sas", jsonName: "ipsecSas"},
	"vpp_config.TunnelProtection":       names{protoName: "ipsec_tunnel_protections", jsonName: "ipsecTunnelProtections"},
	"vpp_config.Interface":              names{protoName: "interfaces", jsonName: "interfaces"},
	"vpp_config.Span":                   names{protoName: "spans", jsonName: "spans"},
	"vpp_config.IPFIX":                  names{protoName: "ipfix_global", jsonName: "ipfixGlobal"},
	"vpp_config.FlowProbeParams":        names{protoName: "ipfix_flowprobe_params", jsonName: "ipfixFlowprobeParams"},
	"vpp_config.FlowProbeFeature":       names{protoName: "ipfix_flowprobes", jsonName: "ipfixFlowprobes"},
	"vpp_config.BridgeDomain":           names{protoName: "bridge_domains", jsonName: "bridgeDomains"},
	"vpp_config.FIBEntry":               names{protoName: "fibs", jsonName: "fibs"},
	"vpp_config.XConnectPair":           names{protoName: "xconnect_pairs", jsonName: "xconnectPairs"},
	"vpp_config.ARPEntry":               names{protoName: "arps", jsonName: "arps"},
	"vpp_config.Route":                  names{protoName: "routes", jsonName: "routes"},
	"vpp_config.ProxyARP":               names{protoName: "proxy_arp", jsonName: "proxyArp"},
	"vpp_config.IPScanNeighbor":         names{protoName: "ipscan_neighbor", jsonName: "ipscanNeighbor"},
	"vpp_config.VrfTable":               names{protoName: "vrfs", jsonName: "vrfs"},
	"vpp_config.DHCPProxy":              names{protoName: "dhcp_proxies", jsonName: "dhcpProxies"},
	"vpp_config.L3XConnect":             names{protoName: "l3xconnects", jsonName: "l3xconnects"},
	"vpp_config.TeibEntry":              names{protoName: "teib_entries", jsonName: "teibEntries"},
	"vpp_config.Nat44Global":            names{protoName: "nat44_global", jsonName: "nat44Global"},
	"vpp_config.DNat44":                 names{protoName: "dnat44s", jsonName: "dnat44s"},
	"vpp_config.Nat44Interface":         names{protoName: "nat44_interfaces", jsonName: "nat44Interfaces"},
	"vpp_config.Nat44AddressPool":       names{protoName: "nat44_pools", jsonName: "nat44Pools"},
	"vpp_config.IPRedirect":             names{protoName: "punt_ipredirects", jsonName: "puntIpredirects"},
	"vpp_config.ToHost":                 names{protoName: "punt_tohosts", jsonName: "puntTohosts"},
	"vpp_config.Exception":              names{protoName: "punt_exceptions", jsonName: "puntExceptions"},
	"vpp_config.LocalSID":               names{protoName: "srv6_localsids", jsonName: "srv6Localsids"},
	"vpp_config.Policy":                 names{protoName: "srv6_policies", jsonName: "srv6Policies"},
	"vpp_config.Steering":               names{protoName: "srv6_steerings", jsonName: "srv6Steerings"},
	"vpp_config.SRv6Global":             names{protoName: "srv6_global", jsonName: "srv6Global"},
}

// NewDynamicConfig creates dynamically proto Message that contains all given configuration models(knowModels).
// This proto message(when all VPP-Agent models are given as input) is json/yaml compatible with
// configurator.Config. The configurator.Config config have all models hardcoded (generated from config
// proto model, but that model is hardcoded). Dynamic config can contain also custom 3rd party models
// and therefore can be used to import/export config data also for 3rd party models that are registered, but not
// part of VPP-Agent repository and therefore not know to hardcoded configurator.Config.
func NewDynamicConfig(knownModels []*ModelInfo, fileDescProtos []*descriptorpb.FileDescriptorProto) (*dynamicpb.Message, error) {
	dependencyRegistry, err := createDependencyRegistry(fileDescProtos)
	if err != nil {
		return nil, errors.Errorf("can't create dependency file descriptor registry due to: %v", err)
	}

	// get file descriptor proto for give known models
	fileDP, rootMsgName, err := createDynamicConfigDescriptorProto(knownModels, dependencyRegistry)
	if err != nil {
		return nil, errors.Errorf("can't create descriptor proto for dynamic config due to: %v", err)
	}

	// convert file descriptor proto to file descriptor
	fd, err := protodesc.NewFile(fileDP, dependencyRegistry)
	if err != nil {
		panic(err) // TODO
	}

	// get descriptor for config root message
	rootMsg := fd.Messages().ByName(rootMsgName)

	// create dynamic config message
	return dynamicpb.NewMessage(rootMsg), nil
}

func createDependencyRegistry(fileDescProtos []*descriptorpb.FileDescriptorProto) (protodesc.Resolver, error) {
	reg := &protoregistry.Files{}
	fds, err := toFileDescriptors(fileDescProtos)
	if err != nil {
		return nil, errors.Errorf("can't convert file descriptor protos to file descriptors "+
			"(for dependency registry creation) due to: %v", err)
	}
	for _, fd := range fds {
		if err := reg.RegisterFile(fd); err != nil {
			return nil, errors.Errorf("can't add proto file descriptor(%v) "+
				"to cache due to: %v", fd.Name(), err)
		}

		logrus.DefaultLogger().Debugf("Proto file %v was successfully "+
			"added to dependency registry.", fd.Path())
	}
	return reg, nil
}

// createDynamicConfigDescriptorProto creates descriptor proto for configuration. The construction of the descriptor
// proto is the way how the configuration from known models are added to the configuration proto message.
// The constructed file descriptor proto is used to get file descriptor that in turn can be used to instantiate
// proto message with all the configs from knownModels. This method conveniently provides also all referenced
// external models of provided knownModels and the configuration root message (proto file has many messages, but
// we need to know which one is the root for our configuration).
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
		Name: proto.String("Config"),
	}

	// create fake root to mimic the same usage as with hardcoded configurator.Config proto message
	// (idea is to not break anything for user that is using yaml configs from/for old
	// hardcoded configurator.Config proto message)
	fakeConfigRootDP := &descriptorpb.DescriptorProto{
		Name: proto.String("Dynamic_config"),
		Field: []*descriptorpb.FieldDescriptorProto{
			&descriptorpb.FieldDescriptorProto{
				Name:     proto.String(configName),
				Number:   proto.Int32(1), // field numbering
				JsonName: proto.String("config"),
				Type:     protoType(descriptorpb.FieldDescriptorProto_TYPE_MESSAGE),
				TypeName: proto.String(*configDP.Name),
				Label:    protoLabel(descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL),
			},
		},
	}
	rootMsgName = protoreflect.Name(*(fakeConfigRootDP.Name))

	// add new messages to proto file
	fileDP.MessageType = []*descriptorpb.DescriptorProto{fakeConfigRootDP, configDP}

	// fill dynamic message with given known models
	configGroups := make(map[string]*descriptorpb.DescriptorProto)
	importedDependency := make(map[string]struct{}) // just for deduplication checking
	for _, modelDetail := range knownModels {
		// get/create group config for this know model (all configs are grouped into groups based on their module)
		configGroupName := fmt.Sprintf("%v%v", modulePrefix(models.ToSpec(modelDetail.Spec).ModelName()), configGroupSuffix)
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
		// TODO remove
		//msgDesc := protoV1.MessageReflect(knownModel.NewInstance()).Descriptor()
		simpleProtoName := simpleProtoName(modelDetail.ProtoName)
		protoName := string(simpleProtoName)
		jsonName := string(simpleProtoName)
		compatibilityKey := fmt.Sprintf("%v.%v", configGroupName, string(simpleProtoName))
		if newNames, found := backwardCompatibleNames[compatibilityKey]; found {
			// using field names from hardcoded configurator.Config to achieve json/yaml backward compatibility
			protoName = newNames.protoName
			jsonName = newNames.jsonName
		}
		configGroup.Field = append(configGroup.Field, &descriptorpb.FieldDescriptorProto{
			Name:     proto.String(protoName),
			Number:   proto.Int32(int32(len(configGroup.Field) + 1)),
			JsonName: proto.String(jsonName),
			Label:    label,
			Type:     protoType(descriptorpb.FieldDescriptorProto_TYPE_MESSAGE),
			TypeName: proto.String(fmt.Sprintf(".%v", modelDetail.ProtoName)),
		})

		//add proto file dependency for this known model (+ check that it is in dependency file descriptor registry)
		protoFile, err := ModelOptionFor("protoFile", modelDetail.Options)
		if err != nil {
			error = errors.Errorf("can't retrieve protoFile from model options "+
				"from model %v due to: %v", modelDetail.ProtoName, err)
			return
		}
		if _, found := importedDependency[protoFile]; !found {
			importedDependency[protoFile] = struct{}{}

			//add proto file dependency for this known model
			fileDP.Dependency = append(fileDP.Dependency, protoFile)

			// checking dependency registry that should already contain the linked dependency
			if _, err := dependencyRegistry.FindFileByPath(protoFile); err != nil {
				if err == protoregistry.NotFound {
					error = errors.Errorf("proto file %v need to be referenced in dynamic config, but it "+
						"is not in dependency registry that was created from file descriptor proto input "+
						"(missing in input? check debug output from creating dependency registry) ", protoFile)
					return
				}
				error = errors.Errorf("can't verify that proto file %v is in "+
					"dependency registry, it is due to: %v", protoFile, err)
				return
			}
		}
	}
	return
}

// toFileDescriptors convert file descriptor protos to file descriptors. This conversion handles correctly
// possible transitive dependencies, but all dependencies (direct or transitive) must be included in input
// file descriptor protos.
func toFileDescriptors(fileDescProtos []*descriptorpb.FileDescriptorProto) ([]protoreflect.FileDescriptor, error) {
	// NOTE this could be done more efficiently by creating dependency tree and
	// traversing it and all, but it seems more complicated to implement
	// => going over unresolved FileDescriptorProto over and over while resolving that FileDescriptorProto that
	// could be resolved, in the end(first round with nothing new to resolve) there is either everything resolved
	// (everything went ok) or there exist something that is not resolved and can't be resolved with given
	// input to this function (this result is considered error expecting not adding to function input
	// additional useless file descriptor protos)
	unresolvedFDProtos := make(map[string]*descriptorpb.FileDescriptorProto)
	for _, fdp := range fileDescProtos {
		unresolvedFDProtos[*fdp.Name] = fdp
	}
	resolved := make(map[string]protoreflect.FileDescriptor)

	newResolvedInLastRound := true
	for len(unresolvedFDProtos) > 0 && newResolvedInLastRound {
		newResolvedInLastRound = false
		for fdpName, fdp := range unresolvedFDProtos {
			allDepsFound := true
			reg := &protoregistry.Files{}
			for _, dependencyName := range fdp.Dependency {
				resolvedDep, found := resolved[dependencyName]
				if !found {
					allDepsFound = false
					break
				}
				if err := reg.RegisterFile(resolvedDep); err != nil {
					return nil, errors.Errorf("can't put resolved dependency %v "+
						"into descriptor registry due to: %v", resolvedDep.Name(), err)
				}
			}
			if allDepsFound {
				fd, err := protodesc.NewFile(fdp, reg)
				if err != nil {
					panic(err) // TODO
				}
				resolved[fdpName] = fd
				delete(unresolvedFDProtos, fdpName)
				newResolvedInLastRound = true
			}
		}
	}
	if len(unresolvedFDProtos) > 0 {
		keys := make([]string, 0, len(unresolvedFDProtos)) // TODO move away
		for key, _ := range unresolvedFDProtos {
			keys = append(keys, key)
		}
		return nil, errors.Errorf("can't resolve these FileDescriptorProto's %v", strings.Join(keys, ","))
	}

	result := make([]protoreflect.FileDescriptor, 0, len(resolved))
	for _, fd := range resolved {
		result = append(result, fd)
	}
	return result, nil
}

// DynamicConfigExport exports from dynamic config the proto.Messages corresponding to known models that
// were given as input when dynamic config was created using NewDynamicConfig. This is a convenient
// method how to extract data for generic client usage (proto.Message instances) from value-filled
// dynamic config (i.e. after json/yaml loading into dynamic config).
func DynamicConfigExport(dynamicConfig *dynamicpb.Message) ([]proto.Message, error) {
	if dynamicConfig == nil {
		return nil, errors.Errorf("dynamic config can't be nil")
	}

	// moving from fake config root to real config root
	configField := dynamicConfig.Descriptor().Fields().ByName(configName)
	if configField == nil {
		return nil, errors.Errorf("can't find field %v. Was provided dynamic config created by "+
			"NewDynamicConfig(...) method or equivalently?", configName)
	}
	configMessage := dynamicConfig.Get(configField).Message()

	// handling export from inner config layers by using helper methods
	return exportFromConfigMessage(configMessage), nil
}

// exportFromConfigMessage exports proto messages from config message layer of dynamic config
func exportFromConfigMessage(configMessage protoreflect.Message) []proto.Message {
	result := make([]proto.Message, 0)
	fields := configMessage.Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		fieldName := fields.Get(i).Name()
		if strings.HasSuffix(string(fieldName), configGroupSuffix) {
			configGroupMessage := configMessage.Get(fields.Get(i)).Message()

			// handling export from inner config layers by using helper methods
			result = append(result, exportFromConfigGroupMessage(configGroupMessage)...)
		}
	}
	return result
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

func ModelOptionFor(key string, options []*generic.ModelDetail_Option) (string, error) {
	for _, option := range options {
		if option.Key == key {
			if len(option.Values) == 0 {
				return "", errors.Errorf("there is no value for key %v in model options", key)
			}
			return option.Values[0], nil
		}
	}
	return "", errors.Errorf("there is no model option with key %v (model options=%+v))", key, options)
}

func existsModelOptionFor(key string, options []*generic.ModelDetail_Option) bool {
	_, err := ModelOptionFor(key, options)
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
