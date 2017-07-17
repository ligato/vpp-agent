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

// Package redis implements client API to Redis key-value data store.  The API confirms to the speciication defined in the package cn-infra/db/keyval.
//
// The entity BytesConnectionRedis provides access to CRUD as well as event subscription API's.
//
//   +-----+    --> (BytesBroker)  +------------------------+   -->  CRUD      +-------+
//   | app |                       |  BytesConnectionRedis  |                  | Redis |
//   +-----+    <-- (BytesWatcher) +------------------------+   <--  events    +-------+
//
// The code snippets below provide examples on using BytesConnectionRedis.  For simplicity, error handling is omitted:
//
// Connection
//   import  "github.com/ligato/cn-infra/db/keyval/redis"
//
//   config := redis.ServerConfig{
//       Endpoint: "localhost:6379",
//       Pool: redis.ConnPool{
//               MaxIdle:     10,
//               MaxActive:   10,
//               IdleTimeout: 60,
//               Wait:        true,
//      },
//   }
//   pool, err := redis.CreateNodeClientConnPool(config)
//   db, err := redis.NewBytesConnectionRedis(pool)
//
// You can also define server configuration in a yaml file, and load it into memory using ParseConfigFromYamlFile(yamlFile, &config) from the package github.com/ligato/cn-infra/utils/config.
// See github.com/ligato/cn-infra/db/keyval/redis/examples/node-client.yaml for an example of server configuration.
//
// CRUD
//   // put
//   err = db.Put("some-key", []byte("some-value"))
//   err = db.Put("some-temp-key", []byte("valid for 20 seconds"), keyval.WithTTL(20*time.Second))
//
//   // get
//   value, found, revision, err := db.GetValue("some-key")
//   if found {
//       ...
//   }
//
//   // list
//   keyPrefix := "some"
//   kv, err := db.ListValues(keyPrefix)
//   for {
//       kv, done := kv.GetNext()
//       if done {
//           break
//       }
//       key := kv.GetKey()
//       value := kv.GetValue()
//   }
//
//   // delete
//   found, err := db.Delete("some-key")
//
//   // transaction
//   var txn keyval.BytesTxn = db.NewTxn()
//   txn.Put("key101", []byte("val 101")).Put("key102", []byte("val 102"))
//   txn.Put("key103", []byte("val 103")).Put("key104", []byte("val 104"))
//   err := txn.Commit()
//
// Subscribe to key space events:
//   watchChan := make(chan keyval.BytesWatchResp, 10)
//   err = db.Watch(watchChan, "some-key")
//   for {
//       select {
//       case r := <-watchChan:
//           switch r.GetChangeType() {
//           case db.Put:
//               log.Infof("Watcher received %v: %s=%s", r.GetChangeType(), r.GetKey(), string(r.GetValue()))
//           case db.Delete:
//               ...
//           }
//       ...
//       }
//   }
//
package redis
