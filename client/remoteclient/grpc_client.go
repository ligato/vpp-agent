package remoteclient

import (
	"context"
	"github.com/go-errors/errors"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"go.ligato.io/vpp-agent/v3/client"
	"go.ligato.io/vpp-agent/v3/pkg/models"
	"go.ligato.io/vpp-agent/v3/pkg/util"
	"go.ligato.io/vpp-agent/v3/proto/ligato/generic"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"strings"
)

// option keys
const (
	externallyKnownModels = "externallyKnownModels"
	messageTypeResolver   = "messageTypeResolver"
)

type grpcClient struct {
	manager        generic.ManagerServiceClient
	meta           generic.MetaServiceClient
	apiFuncOptions client.APIFuncOptions
}

// NewClientGRPC returns new instance that uses given service client for requests.
func NewClientGRPC(conn grpc.ClientConnInterface) client.ConfigClient {
	manager := generic.NewManagerServiceClient(conn)
	meta := generic.NewMetaServiceClient(conn)
	return &grpcClient{
		manager: manager,
		meta:    meta,
	}
}

func (c *grpcClient) KnownModels(class string) ([]*client.ModelInfo, error) {
	ctx := context.Background()

	// get known models from meta service
	resp, err := c.meta.KnownModels(ctx, &generic.KnownModelsRequest{
		Class: class,
	})
	if err != nil {
		return nil, err
	}
	knownModels := resp.KnownModels

	// extract proto files for meta service known models
	protoFilePaths := make(map[string]struct{}) // using map as set for deduplication
	for _, modelDetail := range knownModels {
		protoFilePath, err := client.ModelOptionFor("protoFile", modelDetail.Options)
		if err != nil {
			return nil, errors.Errorf("can't get protoFile from model options of "+
				"known model %v due to: %v", modelDetail.ProtoName, err)
		}
		protoFilePaths[protoFilePath] = struct{}{}
	}

	// query meta service for extracted proto files to get their file descriptor protos
	fileDescProtos := make(map[string]*descriptor.FileDescriptorProto) // deduplicaton + data container
	for protoFilePath, _ := range protoFilePaths {
		ctx := context.Background()
		resp, err := c.meta.ProtoFileDescriptor(ctx, &generic.ProtoFileDescriptorRequest{
			FullProtoFileName: protoFilePath,
		})
		if err != nil {
			return nil, errors.Errorf("can't retrieve ProtoFileDescriptor "+
				"for proto file %v due to: %v", protoFilePath, err)
		}
		if resp.FileDescriptor == nil {
			return nil, errors.Errorf("returned file descriptor proto "+
				"for proto file %v from meta service can't be nil", protoFilePath)
		}
		if resp.FileImportDescriptors == nil {
			return nil, errors.Errorf("returned import file descriptors proto "+
				"for proto file %v from meta service can't be nil", protoFilePath)
		}

		fileDescProtos[*resp.FileDescriptor.Name] = resp.FileDescriptor
		for _, fid := range resp.FileImportDescriptors.File {
			if fid != nil {
				fileDescProtos[*fid.Name] = fid
			}
		}
	}

	// convert file descriptor protos to file descriptors
	fileDescProtosSlice := make([]*descriptorpb.FileDescriptorProto, 0) // conversion set to slice
	for _, fdp := range fileDescProtos {
		fileDescProtosSlice = append(fileDescProtosSlice, fdp)
	}
	fileDescriptors, err := toFileDescriptors(fileDescProtosSlice)
	if err != nil {
		return nil, errors.Errorf("can't convert file descriptor protos to file descriptors "+
			"(for dependency registry creation) due to: %v", err)
	}

	// extract all messages from file descriptors
	messageDescriptors := make(map[string]protoreflect.MessageDescriptor)
	for _, fd := range fileDescriptors {
		for i:=0; i < fd.Messages().Len(); i++ {
			messageDescriptors[string(fd.Messages().Get(i).FullName())] = fd.Messages().Get(i)
		}
	}

	// pack all gathered information into correct output format
	var result []*models.ModelInfo
	for _, info := range knownModels {
		result = append(result, &models.ModelInfo{
			ModelDetail:       *info,
			MessageDescriptor: messageDescriptors[info.ProtoName],
		})
	}

	return result, nil
}

func (c *grpcClient) ChangeRequest(options ...client.ChangeRequestOption) client.ChangeRequest {
	changeRequest := &setConfigRequest{
		client: c.manager,
		req:    &generic.SetConfigRequest{},
	}
	for _, option := range options {
		option(changeRequest)
	}
	return changeRequest
}

func (c *grpcClient) ResyncConfig(items ...proto.Message) error {
	req := &generic.SetConfigRequest{
		OverwriteAll: true,
	}

	for _, protoModel := range items {
		item, err := models.MarshalItem(protoModel)
		if err != nil {
			return err
		}
		req.Updates = append(req.Updates, &generic.UpdateItem{
			Item: item,
		})
	}

	_, err := c.manager.SetConfig(context.Background(), req)
	return err
}

func (c *grpcClient) GetConfig(dsts ...interface{}) error {
	ctx := context.Background()

	resp, err := c.manager.GetConfig(ctx, &generic.GetConfigRequest{})
	if err != nil {
		return err
	}

	knownModels, dontUseLocalModelRegistry := c.apiFuncOptions[externallyKnownModels]
	resolver, resolverFound := c.apiFuncOptions[messageTypeResolver]
	if dontUseLocalModelRegistry && !resolverFound {
		return errors.Errorf("when not using local model registry then message type resolver is needed")
	}

	protos := map[string]proto.Message{}
	for _, item := range resp.Items {
		var val proto.Message
		var key string
		if dontUseLocalModelRegistry {
			knownmodels := knownModels.([]*client.ModelInfo)
			msgTypeResolver := resolver.(*protoregistry.Types)
			val, err = models.UnmarshalItemWithExternallyKnownModels(item.Item, knownmodels, msgTypeResolver)
			if err != nil {
				return err
			}
			if data := item.Item.GetData(); data != nil {
				key, err = models.GetKeyWithExternallyKnownModels(val, knownmodels)
			} else {
				key, err = models.GetKeyForItemWithExternallyKnownModels(item.Item, knownmodels)
			}
		} else {
			val, err = models.UnmarshalItem(item.Item)
			if err != nil {
				return err
			}
			if data := item.Item.GetData(); data != nil {
				key, err = models.GetKey(val)
			} else {
				key, err = models.GetKeyForItem(item.Item)
			}
		}
		if err != nil {
			return err
		}
		protos[key] = val
	}

	util.PlaceProtos(protos, dsts...) // TODO make proto desc version with unlimited struct nesting

	return nil
}

func (c *grpcClient) DumpState() ([]*client.StateItem, error) {
	ctx := context.Background()

	resp, err := c.manager.DumpState(ctx, &generic.DumpStateRequest{})
	if err != nil {
		return nil, err
	}

	return resp.GetItems(), nil
}

type setConfigRequest struct {
	client                generic.ManagerServiceClient
	req                   *generic.SetConfigRequest
	externallyKnownModels []*client.ModelInfo
	err                   error
}

func (r *setConfigRequest) Update(items ...proto.Message) client.ChangeRequest {
	if r.err != nil {
		return r
	}
	for _, protoModel := range items {
		var item *generic.Item
		if r.externallyKnownModels != nil {
			item, r.err = models.MarshalItemWithExternallyKnownModels(protoModel, r.externallyKnownModels)
		} else {
			item, r.err = models.MarshalItem(protoModel)
		}
		if r.err != nil {
			return r
		}
		r.req.Updates = append(r.req.Updates, &generic.UpdateItem{
			Item: item,
		})
	}
	return r
}

func (r *setConfigRequest) Delete(items ...proto.Message) client.ChangeRequest {
	if r.err != nil {
		return r
	}
	for _, protoModel := range items {
		item, err := models.MarshalItem(protoModel)
		if err != nil {
			r.err = err
			return r
		}
		item.Data = nil // delete
		r.req.Updates = append(r.req.Updates, &generic.UpdateItem{
			Item: item,
		})
	}
	return r
}

func (r *setConfigRequest) Send(ctx context.Context) (err error) {
	if r.err != nil {
		return r.err
	}
	_, err = r.client.SetConfig(ctx, r.req)
	return err
}

// WithExternallyKnownModels uses for remote client given list of known models to use instead of local
// model registry that is created by models included in compilation. This can be used to separate models
// between compiled programs (i.e. to have generic agenctl that doesn't have custom models of customized
// vpp-agent fork).
func WithExternallyKnownModels(knownModels []*client.ModelInfo) client.ChangeRequestOption {
	return func(changeRequest client.ChangeRequest) {
		if request, ok := changeRequest.(*setConfigRequest); ok {
			request.externallyKnownModels = knownModels
		}
	}
}

func (c *grpcClient) WithOptions(callFunc func(client.GenericClient), options ...client.APIFuncOptions) {
	newClient := &grpcClient{
		manager:        c.manager,
		meta:           c.meta,
		apiFuncOptions: make(client.APIFuncOptions),
	}
	for _, optionSlice := range options {
		for k, v := range optionSlice {
			newClient.apiFuncOptions[k] = v
		}
	}
	callFunc(newClient)
}

// UseExternallyKnownModels returns properly filled APIFuncOption to use externally known models
func UseExternallyKnownModels(knownModels []*client.ModelInfo) client.APIFuncOptions {
	return client.APIFuncOptions{
		externallyKnownModels: knownModels,
	}
}

// UseMessageTypeResolver returns properly filled APIFuncOption to use message type resolver
func UseMessageTypeResolver(msgTypeResolver *protoregistry.Types) client.APIFuncOptions {
	return client.APIFuncOptions{
		messageTypeResolver: msgTypeResolver,
	}
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
					return nil, errors.Errorf("can't create file descriptor "+
						"(from file descriptor proto named %v) due to: %v", *fdp.Name, err)
				}
				resolved[fdpName] = fd
				delete(unresolvedFDProtos, fdpName)
				newResolvedInLastRound = true
			}
		}
	}
	if len(unresolvedFDProtos) > 0 {
		return nil, errors.Errorf("can't resolve some FileDescriptorProtos due to missing of "+
			"some protos of their imports (FileDescriptorProtos with unresolvable imports: %v)",
			fileDescriptorProtoMapToString(unresolvedFDProtos))
	}

	result := make([]protoreflect.FileDescriptor, 0, len(resolved))
	for _, fd := range resolved {
		result = append(result, fd)
	}
	return result, nil
}

func fileDescriptorProtoMapToString(fdps map[string]*descriptorpb.FileDescriptorProto) string {
	keys := make([]string, 0, len(fdps))
	for key, _ := range fdps {
		keys = append(keys, key)
	}
	return strings.Join(keys, ",")
}
