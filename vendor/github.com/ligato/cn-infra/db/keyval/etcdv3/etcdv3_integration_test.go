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
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/embed"
	"github.com/coreos/etcd/etcdserver/api/v3client"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/logging/logroot"
	"github.com/onsi/gomega"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"
)

const (
	etcdStartTimeout = 30
	prefix           = "/my/prefix/"
	key              = "key"
	watchKey         = "vals/"
)

var (
	broker          *BytesConnectionEtcd
	prefixedBroker  keyval.BytesBroker
	prefixedWatcher keyval.BytesWatcher
	embd            embededEtcd
)

type embededEtcd struct {
	tmpDir string
	etcd   *embed.Etcd
	client *clientv3.Client
}

func TestDataBroker(t *testing.T) {

	//setup
	embd.start(t)
	defer embd.stop()
	gomega.RegisterTestingT(t)

	t.Run("putGetValue", testPutGetValuePrefixed)
	embd.cleanDs()
	t.Run("simpleWatcher", testPrefixedWatcher)
	embd.cleanDs()
	t.Run("listValues", testPrefixedListValues)
	embd.cleanDs()
	t.Run("txn", testPrefixedTxn)
	embd.cleanDs()
	t.Run("testDelWithPrefix", testDelWithPrefix)
}

func teardownBrokers() {
	broker.Close()
	broker = nil
	prefixedBroker = nil
	prefixedWatcher = nil
}

func testPutGetValuePrefixed(t *testing.T) {
	setupBrokers(t)
	defer teardownBrokers()

	data := []byte{1, 2, 3}

	// insert key-value pair using databroker
	err := broker.Put(prefix+key, data)
	gomega.Expect(err).To(gomega.BeNil())

	returnedData, found, _, err := prefixedBroker.GetValue(key)

	gomega.Expect(returnedData).NotTo(gomega.BeNil())
	gomega.Expect(found).To(gomega.BeTrue())
	gomega.Expect(err).To(gomega.BeNil())

	// not existing value
	returnedData, found, _, err = prefixedBroker.GetValue("unknown")
	gomega.Expect(returnedData).To(gomega.BeNil())
	gomega.Expect(found).To(gomega.BeFalse())
	gomega.Expect(err).To(gomega.BeNil())

}

func testPrefixedWatcher(t *testing.T) {
	setupBrokers(t)
	defer teardownBrokers()

	watchCh := make(chan keyval.BytesWatchResp)
	err := prefixedWatcher.Watch(watchCh, watchKey)
	gomega.Expect(err).To(gomega.BeNil())

	wg := sync.WaitGroup{}
	wg.Add(1)
	go expectWatchEvent(t, &wg, watchCh, watchKey+"val1")

	// insert kv that doesn't match the watcher subscription
	broker.Put(prefix+"/something/else/val1", []byte{0, 0, 7})

	// insert kv for watcher
	broker.Put(prefix+watchKey+"val1", []byte{0, 0, 7})

	wg.Wait()
}

func testPrefixedTxn(t *testing.T) {
	setupBrokers(t)
	defer teardownBrokers()

	tx := prefixedBroker.NewTxn()
	gomega.Expect(tx).NotTo(gomega.BeNil())

	tx.Put("b/val1", []byte{0, 1})
	tx.Put("b/val2", []byte{0, 1})
	tx.Put("b/val3", []byte{0, 1})
	tx.Commit()

	kvi, err := broker.ListValues(prefix + "b")
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(kvi).NotTo(gomega.BeNil())

	expectedKeys := []string{prefix + "b/val1", prefix + "b/val2", prefix + "b/val3"}
	for i := 0; i < 3; i++ {
		kv, all := kvi.GetNext()
		gomega.Expect(kv).NotTo(gomega.BeNil())
		gomega.Expect(all).To(gomega.BeFalse())
		gomega.Expect(kv.GetKey()).To(gomega.BeEquivalentTo(expectedKeys[i]))
	}
}

func testPrefixedListValues(t *testing.T) {
	setupBrokers(t)
	defer teardownBrokers()

	var err error
	// insert values using databroker
	err = broker.Put(prefix+"a/val1", []byte{0, 0, 7})
	gomega.Expect(err).To(gomega.BeNil())
	err = broker.Put(prefix+"a/val2", []byte{0, 0, 7})
	gomega.Expect(err).To(gomega.BeNil())
	err = broker.Put(prefix+"a/val3", []byte{0, 0, 7})
	gomega.Expect(err).To(gomega.BeNil())

	// list values using pluginDatabroker
	kvi, err := prefixedBroker.ListValues("a")
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(kvi).NotTo(gomega.BeNil())

	expectedKeys := []string{"a/val1", "a/val2", "a/val3"}
	for i := 0; i < 3; i++ {
		kv, all := kvi.GetNext()
		gomega.Expect(kv).NotTo(gomega.BeNil())
		gomega.Expect(all).To(gomega.BeFalse())
		// verify that prefix of BytesBrokerWatcherEtcd is trimmed
		gomega.Expect(kv.GetKey()).To(gomega.BeEquivalentTo(expectedKeys[i]))
	}
}

func testDelWithPrefix(t *testing.T) {
	setupBrokers(t)
	defer teardownBrokers()

	err := broker.Put("something/a/val1", []byte{0, 0, 7})
	gomega.Expect(err).To(gomega.BeNil())
	err = broker.Put("something/a/val2", []byte{0, 0, 7})
	gomega.Expect(err).To(gomega.BeNil())
	err = broker.Put("something/a/val3", []byte{0, 0, 7})
	gomega.Expect(err).To(gomega.BeNil())

	_, found, _, err := broker.GetValue("something/a/val1")
	gomega.Expect(found).To(gomega.BeTrue())
	gomega.Expect(err).To(gomega.BeNil())

	_, found, _, err = broker.GetValue("something/a/val2")
	gomega.Expect(found).To(gomega.BeTrue())
	gomega.Expect(err).To(gomega.BeNil())

	_, found, _, err = broker.GetValue("something/a/val3")
	gomega.Expect(found).To(gomega.BeTrue())
	gomega.Expect(err).To(gomega.BeNil())

	_, err = broker.Delete("something/a", keyval.WithPrefix())
	gomega.Expect(err).To(gomega.BeNil())

	_, found, _, err = broker.GetValue("something/a/val1")
	gomega.Expect(found).To(gomega.BeFalse())
	gomega.Expect(err).To(gomega.BeNil())

	_, found, _, err = broker.GetValue("something/a/val2")
	gomega.Expect(found).To(gomega.BeFalse())
	gomega.Expect(err).To(gomega.BeNil())

	_, found, _, err = broker.GetValue("something/a/val3")
	gomega.Expect(found).To(gomega.BeFalse())
	gomega.Expect(err).To(gomega.BeNil())

}

func expectWatchEvent(t *testing.T, wg *sync.WaitGroup, watchCh chan keyval.BytesWatchResp, expectedKey string) {
	select {
	case resp := <-watchCh:
		gomega.Expect(resp).NotTo(gomega.BeNil())
		gomega.Expect(resp.GetKey()).To(gomega.BeEquivalentTo(expectedKey))
	case <-time.After(1 * time.Second):
		t.Error("Watch resp not received")
		t.FailNow()
	}
	wg.Done()
}

func (embd *embededEtcd) start(t *testing.T) {
	dir, err := ioutil.TempDir("", "etcd")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	cfg := embed.NewConfig()
	cfg.Dir = dir
	embd.etcd, err = embed.StartEtcd(cfg)
	if err != nil {
		t.Error(err)
		t.FailNow()

	}

	select {
	case <-embd.etcd.Server.ReadyNotify():
		logroot.Logger().Debug("Server is ready!")
	case <-time.After(etcdStartTimeout * time.Second):
		embd.etcd.Server.Stop() // trigger a shutdown
		t.Error("Server took too long to start!")
		t.FailNow()
	}
	embd.client = v3client.New(embd.etcd.Server)
}

func (embd *embededEtcd) stop() {
	embd.etcd.Close()
	os.RemoveAll(embd.tmpDir)
}

// cleanDs deletes all key-value pair stored
func (embd *embededEtcd) cleanDs() {
	if embd.client != nil {
		embd.client.Delete(context.Background(), "", clientv3.WithPrefix())
	}
}

func setupBrokers(t *testing.T) {
	var err error
	broker, err = NewEtcdConnectionUsingClient(v3client.New(embd.etcd.Server), logroot.Logger())

	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(broker).NotTo(gomega.BeNil())
	// create BytesBrokerWatcherEtcd with prefix
	prefixedBroker = broker.NewBroker(prefix)
	prefixedWatcher = broker.NewWatcher(prefix)
	gomega.Expect(prefixedBroker).NotTo(gomega.BeNil())
	gomega.Expect(prefixedWatcher).NotTo(gomega.BeNil())

}
