package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"

	"github.com/ligato/vpp-agent/api/types"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/api"
)

func (c *Client) SchedulerDump(ctx context.Context, opts types.SchedulerDumpOptions) ([]api.KVWithMetadata, error) {
	type ProtoWithName struct {
		ProtoMsgName string
		ProtoMsgData string
	}
	type KVWithMetadata struct {
		api.KVWithMetadata
		Value ProtoWithName
	}
	var kvdump []KVWithMetadata

	query := url.Values{}
	query.Set("key-prefix", opts.KeyPrefix)
	query.Set("view", opts.View)

	resp, err := c.get(ctx, "/scheduler/dump", query, nil)
	if err != nil {
		return nil, err
	}
	if err := json.NewDecoder(resp.body).Decode(&kvdump); err != nil {
		return nil, fmt.Errorf("decoding reply failed: %v", err)
	}

	var dump []api.KVWithMetadata
	for _, kvd := range kvdump {
		d := kvd.KVWithMetadata
		if kvd.Value.ProtoMsgName == "" {
			return nil, fmt.Errorf("empty proto message name for key %s", d.Key)
		}
		valueType := proto.MessageType(kvd.Value.ProtoMsgName)
		if valueType == nil {
			return nil, fmt.Errorf("unknown proto message defined for key %s", d.Key)
		}
		d.Value = reflect.New(valueType.Elem()).Interface().(proto.Message)
		if err = jsonpb.UnmarshalString(kvd.Value.ProtoMsgData, d.Value); err != nil {
			return nil, fmt.Errorf("decoding reply failed: %v", err)
		}
		dump = append(dump, d)
	}
	return dump, nil
}

func (c *Client) SchedulerStatus(ctx context.Context, opts types.SchedulerStatusOptions) ([]*api.BaseValueStatus, error) {
	query := url.Values{}
	query.Set("key-prefix", opts.KeyPrefix)

	resp, err := c.get(ctx, "/scheduler/status", query, nil)
	if err != nil {
		return nil, err
	}

	var status []*api.BaseValueStatus
	if err := json.NewDecoder(resp.body).Decode(&status); err != nil {
		return nil, fmt.Errorf("decoding reply failed: %v", err)
	}

	return status, nil
}
