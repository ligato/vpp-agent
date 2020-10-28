package remoteclient

import (
	"context"

	"github.com/go-errors/errors"
	"github.com/golang/protobuf/proto"
	"go.ligato.io/vpp-agent/v3/client"
	"go.ligato.io/vpp-agent/v3/pkg/models"
	"go.ligato.io/vpp-agent/v3/pkg/util"
	"go.ligato.io/vpp-agent/v3/proto/ligato/generic"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoregistry"
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

	resp, err := c.meta.KnownModels(ctx, &generic.KnownModelsRequest{
		Class: class,
	})
	if err != nil {
		return nil, err
	}

	var modules []*client.ModelInfo
	for _, info := range resp.KnownModels {
		modules = append(modules, info)
	}

	return modules, nil
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
