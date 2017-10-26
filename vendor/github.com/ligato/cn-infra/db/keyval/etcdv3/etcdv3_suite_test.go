// Copyright (c) 2017 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package etcdv3

import (
	"errors"
	"testing"

	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/logging/logroot"
	"github.com/onsi/gomega"
	"golang.org/x/net/context"
)

var dataBroker *BytesConnectionEtcd
var dataBrokerErr *BytesConnectionEtcd
var pluginDataBroker *BytesBrokerWatcherEtcd

// Mock data broker err
type MockKVErr struct {
	// NO-OP
}

func (mock *MockKVErr) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	return nil, errors.New("test-error")
}

func (mock *MockKVErr) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	return nil, errors.New("test-error")
}

func (mock *MockKVErr) Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	return nil, errors.New("test-error")
}

func (mock *MockKVErr) Compact(ctx context.Context, rev int64, opts ...clientv3.CompactOption) (*clientv3.CompactResponse, error) {
	return nil, nil
}

func (mock *MockKVErr) Do(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
	return clientv3.OpResponse{}, nil
}

func (mock *MockKVErr) Txn(ctx context.Context) clientv3.Txn {
	return &MockTxn{}
}

func (mock *MockKVErr) Watch(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan {
	return nil
}

func (mock *MockKVErr) Close() error {
	return nil
}

// Mock KV
type MockKV struct {
	// NO-OP
}

func (mock *MockKV) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	return nil, nil
}

func (mock *MockKV) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	response := *new(clientv3.GetResponse)
	kvs := new(mvccpb.KeyValue)
	kvs.Key = []byte{1}
	kvs.Value = []byte{73, 0x6f, 0x6d, 65, 0x2d, 0x6a, 73, 0x6f, 0x6e} //some-json
	response.Kvs = []*mvccpb.KeyValue{kvs}
	return &response, nil
}

func (mock *MockKV) Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	response := *new(clientv3.DeleteResponse)
	response.PrevKvs = []*mvccpb.KeyValue{}
	return &response, nil
}

func (mock *MockKV) Compact(ctx context.Context, rev int64, opts ...clientv3.CompactOption) (*clientv3.CompactResponse, error) {
	return nil, nil
}

func (mock *MockKV) Do(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
	return clientv3.OpResponse{}, nil
}

func (mock *MockKV) Txn(ctx context.Context) clientv3.Txn {
	return &MockTxn{}
}

func (mock *MockKV) Watch(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan {
	return nil
}

func (mock *MockKV) Close() error {
	return nil
}

// Mock Txn
type MockTxn struct {
}

func (mock *MockTxn) If(cs ...clientv3.Cmp) clientv3.Txn {
	return &MockTxn{}
}

func (mock *MockTxn) Then(ops ...clientv3.Op) clientv3.Txn {
	return &MockTxn{}
}

func (mock *MockTxn) Else(ops ...clientv3.Op) clientv3.Txn {
	return &MockTxn{}
}

func (mock *MockTxn) Commit() (*clientv3.TxnResponse, error) {
	return nil, nil
}

// Tests

func init() {
	mockKv := &MockKV{}
	mockKvErr := &MockKVErr{}
	dataBroker = &BytesConnectionEtcd{Logger: logroot.StandardLogger(), etcdClient: &clientv3.Client{KV: mockKv, Watcher: mockKv}}
	dataBrokerErr = &BytesConnectionEtcd{Logger: logroot.StandardLogger(), etcdClient: &clientv3.Client{KV: mockKvErr, Watcher: mockKvErr}}
	pluginDataBroker = &BytesBrokerWatcherEtcd{Logger: logroot.StandardLogger(), closeCh: make(chan string), kv: mockKv, watcher: mockKv}
}

func TestNewTxn(t *testing.T) {
	gomega.RegisterTestingT(t)
	newTxn := dataBroker.NewTxn()
	gomega.Expect(newTxn).NotTo(gomega.BeNil())
}

func TestTxnPut(t *testing.T) {
	gomega.RegisterTestingT(t)
	newTxn := dataBroker.NewTxn()
	result := newTxn.Put("key", []byte("data"))
	gomega.Expect(result).NotTo(gomega.BeNil())
}

func TestTxnDelete(t *testing.T) {
	gomega.RegisterTestingT(t)
	newTxn := dataBroker.NewTxn()
	gomega.Expect(newTxn).NotTo(gomega.BeNil())
	result := newTxn.Delete("key")
	gomega.Expect(result).NotTo(gomega.BeNil())
}

func TestTxnCommit(t *testing.T) {
	gomega.RegisterTestingT(t)
	newTxn := dataBroker.NewTxn()
	result := newTxn.Commit()
	gomega.Expect(result).To(gomega.BeNil())
}

func TestPut(t *testing.T) {
	// regular case
	gomega.RegisterTestingT(t)
	err := dataBroker.Put("key", []byte("data"))
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	// error case
	err = dataBrokerErr.Put("key", []byte("data"))
	gomega.Expect(err).Should(gomega.HaveOccurred())
	gomega.Expect(err.Error()).To(gomega.BeEquivalentTo("test-error"))
}

func TestGetValue(t *testing.T) {
	// regular case
	gomega.RegisterTestingT(t)
	result, found, _, err := dataBroker.GetValue("key")
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	gomega.Expect(result).NotTo(gomega.BeNil())
	// error case
	result, found, _, err = dataBrokerErr.GetValue("key")
	gomega.Expect(err).Should(gomega.HaveOccurred())
	gomega.Expect(found).To(gomega.BeFalse())
	gomega.Expect(result).To(gomega.BeNil())
	gomega.Expect(err.Error()).To(gomega.BeEquivalentTo("test-error"))
}

func TestListValues(t *testing.T) {
	// regular case
	gomega.RegisterTestingT(t)
	result, err := dataBroker.ListValues("key")
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	gomega.Expect(result).ToNot(gomega.BeNil())

	// error case
	result, err = dataBrokerErr.ListValues("key")
	gomega.Expect(err).Should(gomega.HaveOccurred())
	gomega.Expect(result).To(gomega.BeNil())
	gomega.Expect(err.Error()).To(gomega.BeEquivalentTo("test-error"))
}

func TestListValuesRange(t *testing.T) {
	// regular case
	gomega.RegisterTestingT(t)
	result, err := dataBroker.ListValuesRange("AKey", "ZKey")
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	gomega.Expect(result).ToNot(gomega.BeNil())

	// error case
	result, err = dataBrokerErr.ListValuesRange("AKey", "ZKey")
	gomega.Expect(err).Should(gomega.HaveOccurred())
	gomega.Expect(result).To(gomega.BeNil())
	gomega.Expect(err.Error()).To(gomega.BeEquivalentTo("test-error"))
}

func TestDelete(t *testing.T) {
	// regular case
	gomega.RegisterTestingT(t)
	response, err := dataBroker.Delete("vnf")
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	gomega.Expect(response).To(gomega.BeFalse())
	// error case
	response, err = dataBrokerErr.Delete("vnf")
	gomega.Expect(err).Should(gomega.HaveOccurred())
	gomega.Expect(response).To(gomega.BeFalse())
	gomega.Expect(err.Error()).To(gomega.BeEquivalentTo("test-error"))
}

func TestNewBroker(t *testing.T) {
	gomega.RegisterTestingT(t)
	pdb := dataBroker.NewBroker("/pluginname")
	gomega.Expect(pdb).NotTo(gomega.BeNil())
}

func TestNewWatcher(t *testing.T) {
	gomega.RegisterTestingT(t)
	pdb := dataBroker.NewWatcher("/pluginname")
	gomega.Expect(pdb).NotTo(gomega.BeNil())
}

func TestWatch(t *testing.T) {
	gomega.RegisterTestingT(t)
	err := pluginDataBroker.Watch(func(keyval.BytesWatchResp) {}, nil,"key")
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func TestWatchPutResp(t *testing.T) {
	var rev int64 = 1
	value := []byte("data")
	prevVal := []byte("prevData")
	key := "key"
	gomega.RegisterTestingT(t)
	createResp := NewBytesWatchPutResp(key, value, prevVal, rev)
	gomega.Expect(createResp).NotTo(gomega.BeNil())
	gomega.Expect(createResp.GetChangeType()).To(gomega.BeEquivalentTo(datasync.Put))
	gomega.Expect(createResp.GetKey()).To(gomega.BeEquivalentTo(key))
	gomega.Expect(createResp.GetValue()).To(gomega.BeEquivalentTo(value))
	gomega.Expect(createResp.GetPrevValue()).To(gomega.BeEquivalentTo(prevVal))
	gomega.Expect(createResp.GetRevision()).To(gomega.BeEquivalentTo(rev))
}

func TestWatchDeleteResp(t *testing.T) {
	var rev int64 = 1
	key := "key"
	gomega.RegisterTestingT(t)
	createResp := NewBytesWatchDelResp(key, rev)
	gomega.Expect(createResp).NotTo(gomega.BeNil())
	gomega.Expect(createResp.GetChangeType()).To(gomega.BeEquivalentTo(datasync.Delete))
	gomega.Expect(createResp.GetKey()).To(gomega.BeEquivalentTo(key))
	gomega.Expect(createResp.GetValue()).To(gomega.BeNil())
	gomega.Expect(createResp.GetRevision()).To(gomega.BeEquivalentTo(rev))
}

func TestConfig(t *testing.T) {
	gomega.RegisterTestingT(t)
	cfg := &Config{DialTimeout: time.Second, OpTimeout: time.Second}
	etcdCfg, err := ConfigToClientv3(cfg)
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(etcdCfg).NotTo(gomega.BeNil())
	gomega.Expect(etcdCfg.OpTimeout).To(gomega.BeEquivalentTo(time.Second))
	gomega.Expect(etcdCfg.DialTimeout).To(gomega.BeEquivalentTo(time.Second))
	gomega.Expect(etcdCfg.TLS).To(gomega.BeNil())
}
