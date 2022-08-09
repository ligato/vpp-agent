package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"go.ligato.io/vpp-agent/v3/cmd/agentctl/api/types"
	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/proto/ligato/kvscheduler"
)

func (c *Client) SchedulerDump(ctx context.Context, opts types.SchedulerDumpOptions) ([]api.RecordedKVWithMetadata, error) {
	query := url.Values{}
	query.Set("key-prefix", opts.KeyPrefix)
	query.Set("view", opts.View)

	resp, err := c.get(ctx, "/scheduler/dump", query, nil)
	if err != nil {
		return nil, err
	}
	var dump []api.RecordedKVWithMetadata
	if err := json.NewDecoder(resp.body).Decode(&dump); err != nil {
		return nil, fmt.Errorf("decoding reply failed: %v", err)
	}
	return dump, nil
}

func (c *Client) SchedulerValues(ctx context.Context, opts types.SchedulerValuesOptions) ([]*kvscheduler.BaseValueStatus, error) {
	query := url.Values{}
	query.Set("key-prefix", opts.KeyPrefix)
	query.Set("key", opts.Key)

	resp, err := c.get(ctx, "/scheduler/status", query, nil)
	if err != nil {
		return nil, err
	}
	var status []*kvscheduler.BaseValueStatus
	if opts.Key != "" {
		status = []*kvscheduler.BaseValueStatus{{}}
		err = json.NewDecoder(resp.body).Decode(status[0])
	} else {
		err = json.NewDecoder(resp.body).Decode(&status)
	}
	if err != nil {
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
