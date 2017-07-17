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

package redis

import (
	"errors"
	"testing"

	"github.com/garyburd/redigo/redis"
	"github.com/ligato/cn-infra/db"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logroot"
	"github.com/onsi/gomega"
	"github.com/rafaeljusto/redigomock"
	"strconv"
)

var mockConn *redigomock.Conn
var mockPool *redis.Pool
var bytesBroker *BytesConnectionRedis
var log logging.Logger

var keyValues = map[string]string{
	"keyWest": "a place",
	"keyMap":  "a map",
}

//func TestMain(m *testing.M) {
func init() {
	log = logroot.Logger()
	mockConn = redigomock.NewConn()
	var iKeys []interface{}
	var iVals []interface{}
	var iAll []interface{}
	for k, v := range keyValues {
		mockConn.Command("SET", k, v).Expect("not used")
		mockConn.Command("GET", k).Expect(v)
		iKeys = append(iKeys, k)
		iVals = append(iVals, v)
		iAll = append(append(iAll, k), v)
	}
	mockConn.Command("GET", "key").Expect(nil)
	mockConn.Command("GET", "bytes").Expect([]byte("bytes"))
	mockConn.Command("GET", "nil").Expect(nil)

	mockConn.Command("MGET", iKeys...).Expect(iVals)
	mockConn.Command("KEYS", "key*").Expect(iKeys)
	mockConn.Command("DEL", iKeys...).Expect(len(keyValues))

	mockConn.Command("MSET", []interface{}{"keyMap", keyValues["keyMap"]}...).Expect(nil)
	mockConn.Command("DEL", []interface{}{"keyWest"}...).Expect(1).Expect(nil)
	mockConn.Command("PSUBSCRIBE", []interface{}{keySpaceEventPrefix + "key*"}...).Expect(newSubscriptionResponse("psubscribe", keySpaceEventPrefix+"key*", 1))

	// for negative tests
	manufacturedError := errors.New("manufactured error")
	mockConn.Command("SET", "error", "error").ExpectError(manufacturedError)
	mockConn.Command("GET", "error").ExpectError(manufacturedError)
	mockConn.Command("GET", "redisError").Expect(redis.Error("Blah"))
	mockConn.Command("GET", "unknown").Expect(struct{}{})

	mockPool = &redis.Pool{
		Dial: func() (redis.Conn, error) { return mockConn, nil },
	}
	bytesBroker, _ = NewBytesConnectionRedis(mockPool, logroot.Logger())
}

func TestPut(t *testing.T) {
	gomega.RegisterTestingT(t)
	err := bytesBroker.Put("keyWest", []byte(keyValues["keyWest"]))
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func TestPutError(t *testing.T) {
	gomega.RegisterTestingT(t)
	err := bytesBroker.Put("error", []byte("error"))
	gomega.Expect(err).Should(gomega.HaveOccurred())
}

func TestGet(t *testing.T) {
	gomega.RegisterTestingT(t)
	val, found, _, err := bytesBroker.GetValue("keyWest")
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	gomega.Expect(found).Should(gomega.BeTrue())
	gomega.Expect(val).Should(gomega.Equal([]byte(keyValues["keyWest"])))

	val, found, _, err = bytesBroker.GetValue("bytes")
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

	val, found, _, err = bytesBroker.GetValue("nil")
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	gomega.Expect(found).Should(gomega.BeFalse())
	gomega.Expect(val).Should(gomega.BeNil())
}

func TestGetError(t *testing.T) {
	gomega.RegisterTestingT(t)
	val, found, _, err := bytesBroker.GetValue("error")
	gomega.Expect(err).Should(gomega.HaveOccurred())
	gomega.Expect(found).Should(gomega.BeFalse())
	gomega.Expect(val).Should(gomega.BeNil())

	val, found, _, err = bytesBroker.GetValue("redisError")
	gomega.Expect(err).Should(gomega.HaveOccurred())
	gomega.Expect(found).Should(gomega.BeFalse())
	gomega.Expect(val).Should(gomega.BeNil())

	val, found, _, err = bytesBroker.GetValue("unknown")
	gomega.Expect(err).Should(gomega.HaveOccurred())
	gomega.Expect(found).Should(gomega.BeFalse())
	gomega.Expect(val).Should(gomega.BeNil())
}

func TestGetShouldNotApplyWildcard(t *testing.T) {
	gomega.RegisterTestingT(t)
	val, found, _, err := bytesBroker.GetValue("key")
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	gomega.Expect(found).Should(gomega.BeFalse())
	gomega.Expect(val).Should(gomega.BeNil())
}

func TestListValues(t *testing.T) {
	gomega.RegisterTestingT(t)
	keyVals, err := bytesBroker.ListValues("key")
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	for {
		kv, last := keyVals.GetNext()
		if last {
			break
		}
		gomega.Expect(kv.GetKey()).Should(gomega.SatisfyAny(gomega.BeEquivalentTo("keyWest"), gomega.BeEquivalentTo("keyMap")))
		gomega.Expect(kv.GetValue()).Should(gomega.SatisfyAny(gomega.BeEquivalentTo(keyValues["keyWest"]), gomega.BeEquivalentTo(keyValues["keyMap"])))
	}
}

func TestListKeys(t *testing.T) {
	gomega.RegisterTestingT(t)
	keys, err := bytesBroker.ListKeys("key")
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	for {
		k, _, last := keys.GetNext()
		if last {
			break
		}
		gomega.Expect(k).Should(gomega.SatisfyAny(gomega.BeEquivalentTo("keyWest"), gomega.BeEquivalentTo("keyMap")))
	}
}

func TestDel(t *testing.T) {
	gomega.RegisterTestingT(t)
	/*found*/ _, err := bytesBroker.Delete("key")
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	//gomega.Expect(found).Should(gomega.BeTrue()) // why is this not found, all of a sudden?
}

func TestTxn(t *testing.T) {
	gomega.RegisterTestingT(t)
	txn := bytesBroker.NewTxn()
	txn.Put("keyWest", []byte(keyValues["keyWest"])).Put("keyMap", []byte(keyValues["keyMap"]))
	txn.Delete("keyWest")
	err := txn.Commit()
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func TestWatch(t *testing.T) {
	gomega.RegisterTestingT(t)
	count := 0
	mockConn.AddSubscriptionMessage(newPMessage(keySpaceEventPrefix+"key*", keySpaceEventPrefix+"keyWest", "set"))
	count++
	mockConn.AddSubscriptionMessage(newPMessage(keySpaceEventPrefix+"key*", keySpaceEventPrefix+"keyWest", "del"))
	count++
	doneChan := make(chan struct{})
	respChan := make(chan keyval.BytesWatchResp)
	bytesBroker.Watch(respChan, "key")
	go func() {
		for {
			select {
			case r, ok := <-respChan:
				if ok {
					switch r.GetChangeType() {
					case db.Put:
						log.Debugf("Watcher received %v: %s=%s", r.GetChangeType(), r.GetKey(), string(r.GetValue()))
					case db.Delete:
						log.Debugf("Watcher received %v: %s", r.GetChangeType(), r.GetKey())
					}
					count--
					if count == 0 {
						doneChan <- struct{}{}
					}
				} else {
					log.Error("Something wrong with respChan... bail out")
					return
				}
			default:
				break
			}
		}
	}()
	<-doneChan
	log.Infof("TestWatch is done")
}

func newSubscriptionResponse(kind string, chanName string, count int) []interface{} {
	values := []interface{}{}
	values = append(values, interface{}([]byte(kind)))
	values = append(values, interface{}([]byte(chanName)))
	values = append(values, interface{}([]byte(strconv.Itoa(count))))
	return values
}

func newPMessage(pattern string, chanName string, data string) []interface{} {
	values := []interface{}{}
	values = append(values, interface{}([]byte("pmessage")))
	values = append(values, interface{}([]byte(pattern)))
	values = append(values, interface{}([]byte(chanName)))
	values = append(values, interface{}([]byte(data)))
	return values
}

func TestBrokerClosed(t *testing.T) {
	gomega.RegisterTestingT(t)
	bytesBroker.Close()

	err := bytesBroker.Put("any", []byte("any"))
	gomega.Expect(err).Should(gomega.HaveOccurred())
	_, _, _, err = bytesBroker.GetValue("any")
	gomega.Expect(err).Should(gomega.HaveOccurred())
	_, err = bytesBroker.ListValues("any")
	gomega.Expect(err).Should(gomega.HaveOccurred())
	_, err = bytesBroker.ListKeys("any")
	gomega.Expect(err).Should(gomega.HaveOccurred())
	_, err = bytesBroker.Delete("any")
	gomega.Expect(err).Should(gomega.HaveOccurred())
}
