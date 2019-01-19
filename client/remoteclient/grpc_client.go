package remoteclient

import (
	"context"
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/ligato/vpp-agent/api"
	"github.com/ligato/vpp-agent/client"
	"github.com/ligato/vpp-agent/pkg/models"
	"github.com/ligato/vpp-agent/plugins/dispatcher"
)

type grpcClient struct {
	remote api.GenericConfiguratorClient
}

// NewClientGRPC returns new instance that uses given service client for requests.
func NewClientGRPC(client api.GenericConfiguratorClient) client.ConfigClient {
	return &grpcClient{client}
}

func (c *grpcClient) ActiveModels() (map[string][]api.ModelInfo, error) {
	ctx := context.Background()

	resp, err := c.remote.Capabilities(ctx, &api.CapabilitiesRequest{})
	if err != nil {
		return nil, err
	}

	modules := make(map[string][]api.ModelInfo)
	for _, info := range resp.KnownModels {
		modules[info.Model.Module] = append(modules[info.Model.Module], *info)
	}

	return modules, nil
}

func (c *grpcClient) GetConfig(dsts ...interface{}) error {
	ctx := context.Background()

	resp, err := c.remote.GetConfig(ctx, &api.GetConfigRequest{})
	if err != nil {
		return err
	}

	fmt.Printf("GetConfig: %+v\n", resp)

	protos := map[string]proto.Message{}
	for _, item := range resp.Items {
		val, err := models.UnmarshalItem(item.Item)
		if err != nil {
			return err
		}
		var key string
		if data := item.Item.GetData(); data != nil {
			key, err = models.GetKey(val)
		} else {
			// protos[item.Item.Key] = val
			key, err = models.ItemKey(item.Item)
		}
		if err != nil {
			return err
		}
		protos[key] = val
	}

	dispatcher.PlaceProtos(protos, dsts...)

	return nil
}

func (c *grpcClient) SetConfig(resync bool) client.SetConfigRequest {
	return &setConfigRequest{
		client: c.remote,
		req: &api.SetConfigRequest{
			//Options: &api.SetConfigRequest_Options{Resync: resync},
			OverwriteAll: resync,
		},
	}
}

type setConfigRequest struct {
	client api.GenericConfiguratorClient
	req    *api.SetConfigRequest
	err    error
}

func (r *setConfigRequest) Update(items ...proto.Message) {
	if r.err != nil {
		return
	}
	for _, protoModel := range items {
		item, err := models.MarshalItem(protoModel)
		if err != nil {
			r.err = err
			return
		}
		r.req.Updates = append(r.req.Updates, &api.UpdateItem{
			Item: item,
		})
	}
}

func (r *setConfigRequest) Delete(items ...proto.Message) {
	if r.err != nil {
		return
	}
	for _, protoModel := range items {
		item, err := models.MarshalItem(protoModel)
		if err != nil {
			if err != nil {
				r.err = err
				return
			}
		}
		r.req.Updates = append(r.req.Updates, &api.UpdateItem{
			/*Item: &api.Item{
				Key: item.Key,
			},*/
			Item: item,
		})
	}
}

func (r *setConfigRequest) Send(ctx context.Context) (err error) {
	if r.err != nil {
		return r.err
	}
	_, err = r.client.SetConfig(ctx, r.req)
	return err
}
