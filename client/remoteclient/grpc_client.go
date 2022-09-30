package remoteclient

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"

	"go.ligato.io/vpp-agent/v3/client"
	"go.ligato.io/vpp-agent/v3/pkg/models"
	"go.ligato.io/vpp-agent/v3/pkg/util"
	"go.ligato.io/vpp-agent/v3/proto/ligato/generic"
)

type grpcClient struct {
	manager       generic.ManagerServiceClient
	meta          generic.MetaServiceClient
	modelRegistry models.Registry
}

type NewClientOption = func(client.GenericClient) error

// NewClientGRPC returns new instance that uses given service client for requests.
func NewClientGRPC(conn grpc.ClientConnInterface, options ...NewClientOption) (client.ConfigClient, error) {
	manager := generic.NewManagerServiceClient(conn)
	meta := generic.NewMetaServiceClient(conn)
	c := &grpcClient{
		manager:       manager,
		meta:          meta,
		modelRegistry: models.DefaultRegistry,
	}
	for _, option := range options {
		if err := option(c); err != nil {
			return nil, fmt.Errorf("cannot apply option to newly created GRPC client due to: %w", err)
		}
	}
	return c, nil
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
			return nil, fmt.Errorf("cannot get protoFile from model options of "+
				"known model %v due to: %w", modelDetail.ProtoName, err)
		}
		protoFilePaths[protoFilePath] = struct{}{}
	}

	// query meta service for extracted proto files to get their file descriptor protos
	fileDescProtos := make(map[string]*descriptorpb.FileDescriptorProto) // deduplicaton + data container
	for protoFilePath := range protoFilePaths {
		ctx := context.Background()
		resp, err := c.meta.ProtoFileDescriptor(ctx, &generic.ProtoFileDescriptorRequest{
			FullProtoFileName: protoFilePath,
		})
		if err != nil {
			return nil, fmt.Errorf("cannot retrieve ProtoFileDescriptor "+
				"for proto file %v due to: %w", protoFilePath, err)
		}
		if resp.FileDescriptor == nil {
			return nil, fmt.Errorf("returned file descriptor proto "+
				"for proto file %v from meta service can't be nil", protoFilePath)
		}
		if resp.FileImportDescriptors == nil {
			return nil, fmt.Errorf("returned import file descriptors proto "+
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
		return nil, fmt.Errorf("cannot convert file descriptor protos to file descriptors "+
			"(for dependency registry creation) due to: %w", err)
	}

	// extract all messages from file descriptors
	messageDescriptors := make(map[string]protoreflect.MessageDescriptor)
	for _, fd := range fileDescriptors {
		for i := 0; i < fd.Messages().Len(); i++ {
			messageDescriptors[string(fd.Messages().Get(i).FullName())] = fd.Messages().Get(i)
		}
	}

	// pack all gathered information into correct output format
	var result []*models.ModelInfo
	for _, info := range knownModels {
		result = append(result, &models.ModelInfo{
			ModelDetail:       info,
			MessageDescriptor: messageDescriptors[info.ProtoName],
		})
	}

	return result, nil
}

func (c *grpcClient) ChangeRequest() client.ChangeRequest {
	return &setConfigRequest{
		client:        c.manager,
		modelRegistry: c.modelRegistry,
		req:           &generic.SetConfigRequest{},
	}
}

func (c *grpcClient) ResyncConfig(items ...proto.Message) error {
	req := &generic.SetConfigRequest{
		OverwriteAll: true,
	}

	for _, protoModel := range items {
		item, err := models.MarshalItemUsingModelRegistry(protoModel, c.modelRegistry)
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

func (c *grpcClient) GetFilteredConfig(filter client.Filter, dsts ...interface{}) error {
	ctx := context.Background()

	resp, err := c.manager.GetConfig(ctx, &generic.GetConfigRequest{Ids: filter.Ids, Labels: filter.Labels})
	if err != nil {
		return err
	}

	protos := map[string]proto.Message{}
	for _, item := range resp.Items {
		var key string
		val, err := models.UnmarshalItemUsingModelRegistry(item.Item, c.modelRegistry)
		if err != nil {
			return err
		}
		if data := item.Item.GetData(); data != nil {
			key, err = models.GetKeyUsingModelRegistry(val, c.modelRegistry)
		} else {
			key, err = models.GetKeyForItemUsingModelRegistry(item.Item, c.modelRegistry)
		}
		if err != nil {
			return err
		}
		protos[key] = val
	}

	protoDsts := extractProtoMessages(dsts)
	if len(dsts) == len(protoDsts) { // all dsts are proto messages
		// TODO the clearIgnoreLayerCount function argument should be a option of generic.Client
		//  (the value 1 generates from dynamic config the same json/yaml output as the hardcoded
		//  configurator.Config and therefore serves for backward compatibility)
		util.PlaceProtosIntoProtos(protoMapToList(protos), 1, protoDsts...)
	} else {
		util.PlaceProtos(protos, dsts...)
	}

	return nil
}

func (c *grpcClient) GetConfig(dsts ...interface{}) error {
	return c.GetFilteredConfig(client.Filter{}, dsts)
}

func (c *grpcClient) GetItems(ctx context.Context) ([]*client.ConfigItem, error) {
	resp, err := c.manager.GetConfig(ctx, &generic.GetConfigRequest{})
	if err != nil {
		return nil, err
	}
	return resp.GetItems(), err
}

func (c *grpcClient) UpdateItems(ctx context.Context, items []client.UpdateItem, resync bool) ([]*client.UpdateResult, error) {
	req := &generic.SetConfigRequest{
		OverwriteAll: resync,
	}
	for _, ui := range items {
		var item *generic.Item
		item, err := models.MarshalItemUsingModelRegistry(ui.Message, c.modelRegistry)
		if err != nil {
			return nil, err
		}
		req.Updates = append(req.Updates, &generic.UpdateItem{
			Item:   item,
			Labels: ui.Labels,
		})
	}
	res, err := c.manager.SetConfig(ctx, req)
	if err != nil {
		return nil, err
	}
	var updateResults []*client.UpdateResult
	for _, r := range res.Results {
		updateResults = append(updateResults, &client.UpdateResult{
			Key:    r.Key,
			Status: r.Status,
		})
	}
	return updateResults, nil
}

func (c *grpcClient) DeleteItems(ctx context.Context, items []client.UpdateItem) ([]*client.UpdateResult, error) {
	req := &generic.SetConfigRequest{}
	for _, ui := range items {
		var item *generic.Item
		item, err := models.MarshalItemUsingModelRegistry(ui.Message, c.modelRegistry)
		if err != nil {
			return nil, err
		}
		item.Data = nil // delete
		req.Updates = append(req.Updates, &generic.UpdateItem{
			Item: item,
		})
	}
	res, err := c.manager.SetConfig(ctx, req)
	if err != nil {
		return nil, err
	}
	var updateResults []*client.UpdateResult
	for _, r := range res.Results {
		updateResults = append(updateResults, &client.UpdateResult{
			Key:    r.Key,
			Status: r.Status,
		})
	}
	return updateResults, nil
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
	client        generic.ManagerServiceClient
	modelRegistry models.Registry
	req           *generic.SetConfigRequest
	err           error
}

func (r *setConfigRequest) Update(items ...proto.Message) client.ChangeRequest {
	if r.err != nil {
		return r
	}
	for _, protoModel := range items {
		var item *generic.Item
		item, r.err = models.MarshalItemUsingModelRegistry(protoModel, r.modelRegistry)
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
		item, err := models.MarshalItemUsingModelRegistry(protoModel, r.modelRegistry)
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

// UseRemoteRegistry modifies remote client to use remote model registry instead of local model registry. The
// remote model registry is filled with remote known models for given class (modelClass).
func UseRemoteRegistry(modelClass string) NewClientOption {
	return func(c client.GenericClient) error {
		if grpcClient, ok := c.(*grpcClient); ok {
			// get all remote models
			knownModels, err := grpcClient.KnownModels(modelClass)
			if err != nil {
				return fmt.Errorf("cannot retrieve remote models (in UseRemoteRegistry) due to: %w", err)
			}

			// fill them into new remote registry and use that registry instead of default local model registry
			grpcClient.modelRegistry = models.NewRemoteRegistry()
			for _, knowModel := range knownModels {
				if _, err := grpcClient.modelRegistry.Register(knowModel, models.ToSpec(knowModel.Spec)); err != nil {
					return fmt.Errorf("cannot register remote known model "+
						"for remote generic client usage due to: %w", err)
				}
			}
		}
		return nil
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
					return nil, fmt.Errorf("cannot put resolved dependency %v "+
						"into descriptor registry due to: %v", resolvedDep.Name(), err)
				}
			}
			if allDepsFound {
				fd, err := protodesc.NewFile(fdp, reg)
				if err != nil {
					return nil, fmt.Errorf("cannot create file descriptor "+
						"(from file descriptor proto named %v) due to: %v", *fdp.Name, err)
				}
				resolved[fdpName] = fd
				delete(unresolvedFDProtos, fdpName)
				newResolvedInLastRound = true
			}
		}
	}
	if len(unresolvedFDProtos) > 0 {
		return nil, fmt.Errorf("cannot resolve some FileDescriptorProtos due to missing of "+
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
	for key := range fdps {
		keys = append(keys, key)
	}
	return strings.Join(keys, ",")
}

func extractProtoMessages(dsts []interface{}) []proto.Message {
	protoDsts := make([]proto.Message, 0)
	for _, dst := range dsts {
		msg, ok := dst.(proto.Message)
		if ok {
			protoDsts = append(protoDsts, msg)
		} else {
			break
		}
	}
	return protoDsts
}

func protoMapToList(protoMap map[string]proto.Message) []proto.Message {
	result := make([]proto.Message, 0, len(protoMap))
	for _, msg := range protoMap {
		result = append(result, msg)
	}
	return result
}
