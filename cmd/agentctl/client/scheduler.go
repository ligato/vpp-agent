package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"

	"go.ligato.io/vpp-agent/v3/cmd/agentctl/api/types"
	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/proto/ligato/kvscheduler"
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
	query := url.Values{}
	query.Set("key-prefix", opts.KeyPrefix)
	query.Set("view", opts.View)

	resp, err := c.get(ctx, "/scheduler/dump", query, nil)
	if err != nil {
		return nil, err
	}
	var kvdump []KVWithMetadata
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

		if len(kvd.Value.ProtoMsgData) > 0 && kvd.Value.ProtoMsgData[0] == '{' {
			err = jsonpb.UnmarshalString(kvd.Value.ProtoMsgData, d.Value)
		} else {
			err = proto.UnmarshalText(kvd.Value.ProtoMsgData, d.Value)
		}
		if err != nil {
			return nil, fmt.Errorf("decoding dump reply for %v failed: %v", valueType, err)
		}
		dump = append(dump, d)
	}
	return dump, nil
}

func (c *Client) SchedulerValues(ctx context.Context, opts types.SchedulerValuesOptions) ([]*kvscheduler.BaseValueStatus, error) {
	query := url.Values{}
	query.Set("key-prefix", opts.KeyPrefix)

	resp, err := c.get(ctx, "/scheduler/status", query, nil)
	if err != nil {
		return nil, err
	}
	var status []*kvscheduler.BaseValueStatus
	if err := json.NewDecoder(resp.body).Decode(&status); err != nil {
		return nil, fmt.Errorf("decoding reply failed: %v", err)
	}

	return status, nil
}

func (c *Client) SchedulerResync(ctx context.Context, opts types.SchedulerResyncOptions) (*api.RecordedTxn, error) {
	query := url.Values{}
	if opts.Retry {
		query.Set("retry", "1")
	}
	if opts.Verbose {
		query.Set("verbose", "1")
	}

	resp, err := c.post(ctx, "/scheduler/downstream-resync", query, nil, nil)
	if err != nil {
		return nil, err
	}

	var rectxn api.RecordedTxn
	if err := json.NewDecoder(resp.body).Decode(&rectxn); err != nil {
		return nil, fmt.Errorf("decoding reply failed: %v", err)
	}

	return &rectxn, nil
}

func (c *Client) SchedulerHistory(ctx context.Context, opts types.SchedulerHistoryOptions) (api.RecordedTxns, error) {
	query := url.Values{}
	if opts.SeqNum >= 0 {
		query.Set("seq-num", fmt.Sprint(opts.SeqNum))
	}

	resp, err := c.get(ctx, "/scheduler/txn-history", query, nil)
	if err != nil {
		return nil, err
	}

	if opts.SeqNum >= 0 {
		var rectxn api.RecordedTxn
		if err := json.NewDecoder(resp.body).Decode(&rectxn); err != nil {
			return nil, fmt.Errorf("decoding reply failed: %v", err)
		}
		return api.RecordedTxns{&rectxn}, nil
	}

	var rectxn api.RecordedTxns
	if err := json.NewDecoder(resp.body).Decode(&rectxn); err != nil {
		return nil, fmt.Errorf("decoding reply failed: %v", err)
	}

	return rectxn, nil
}
