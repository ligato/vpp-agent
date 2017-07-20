package main

import (
	"os"
	"time"

	"github.com/ligato/cn-infra/db"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/db/keyval/redis"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logroot"
	"github.com/ligato/cn-infra/utils/config"
)

var log = logroot.Logger()

var broker keyval.BytesBroker
var watcher keyval.BytesWatcher

func main() {
	log.SetLevel(logging.DebugLevel)

	redisConn := createConnection(os.Args[1])
	broker = redisConn.NewBroker("")
	watcher = redisConn.NewWatcher("")

	runSimpleExmple(redisConn)
}

func createConnection(yamlFile string) *redis.BytesConnectionRedis {
	var err error
	//generateServerConfig(yamlFile)
	var nodeClient redis.NodeClientConfig
	err = config.ParseConfigFromYamlFile(yamlFile, &nodeClient)
	if err != nil {
		log.Panicf("ParseConfigFromYamlFile() failed: %s", err)
	}
	pool, err := redis.CreateNodeClientConnPool(nodeClient)
	if err != nil {
		log.Panicf("CreateNodeClientConnPool() failed: %s", err)
	}
	var redisConn *redis.BytesConnectionRedis
	redisConn, err = redis.NewBytesConnectionRedis(pool, log)
	if err != nil {
		pool.Close()
		log.Panicf("NewBytesConnectionRedis() failed: %s", err)
	}
	return redisConn
}

func runSimpleExmple(redisConn *redis.BytesConnectionRedis) {
	var err error

	var key1, key2, key3 = "key1", "key2", "key3"
	keyPrefix := key1[:3]

	respChan := make(chan keyval.BytesWatchResp, 10)
	err = watcher.Watch(respChan, keyPrefix)
	if err != nil {
		log.Errorf("Watch(%s): %s", keyPrefix, err)
	}
	go func() {
		for {
			select {
			case r, ok := <-respChan:
				if ok {
					switch r.GetChangeType() {
					case db.Put:
						log.Infof("Watcher received %v: %s=%s", r.GetChangeType(), r.GetKey(), string(r.GetValue()))
					case db.Delete:
						log.Infof("Watcher received %v: %s", r.GetChangeType(), r.GetKey())
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
	time.Sleep(2 * time.Second)
	put(key1, "val 1")
	put(key2, "val 2")
	put(key3, "val 3", keyval.WithTTL(time.Second))

	time.Sleep(2 * time.Second)
	get(key1)
	get(key2)
	get(key3)      // key3 should've expired
	get(keyPrefix) // keyPrefix shouldn't find anything
	listKeys(keyPrefix)
	listVal(keyPrefix)

	del(keyPrefix)

	get(key1)
	get(key2)

	txn()

	listVal(keyPrefix)

	log.Info("Sleep for 5 seconds")
	time.Sleep(5 * time.Second)

	// Done watching.  Close the channel.
	log.Infof("Closing broker/watcher")
	//close(respChan)
	redisConn.Close()

	del(keyPrefix)

	log.Info("Sleep for 30 seconds")
	time.Sleep(30 * time.Second)
}

func put(key, value string, opts ...keyval.PutOption) {
	err := broker.Put(key, []byte(value), opts...)
	if err != nil {
		log.Panicf("Put(%s): %s", key, err)
	}
}

func get(key string) {
	val, found, revision, err := broker.GetValue(key)
	if err != nil {
		log.Errorf("GetValue(%s): %s", key, err)
	} else {
		if found {
			log.Infof("GetValue(%s) = %t ; val = %s ; revision = %d", key, found, val, revision)
		} else {
			log.Infof("GetValue(%s) = %t", key, found)
		}
	}
}

func listKeys(keyPrefix string) {
	k, err := broker.ListKeys(keyPrefix)
	if err != nil {
		log.Errorf("ListKeys(%s): %s", keyPrefix, err)
	} else {
		for {
			key, rev, done := k.GetNext()
			if done {
				break
			}
			log.Infof("ListKeys(%s):  %s (rev %d)", keyPrefix, key, rev)
		}
	}
}

func listVal(keyPrefix string) {
	kv, err := broker.ListValues(keyPrefix)
	if err != nil {
		log.Errorf("ListValues(%s): %s", keyPrefix, err)
	} else {
		for {
			kv, done := kv.GetNext()
			if done {
				break
			}
			log.Infof("ListValues(%s):  %s = %s (rev %d)", keyPrefix, kv.GetKey(), kv.GetValue(), kv.GetRevision())
		}
	}
}

func del(keyPrefix string) {
	found, err := broker.Delete(keyPrefix)
	if err != nil {
		log.Errorf("Delete(%s): %s", keyPrefix, err)
	}
	log.Infof("Delete(%s): found = %t", keyPrefix, found)
}

func txn() {
	var key101, key102, key103, key104 = "key101", "key102", "key103", "key104"
	txn := broker.NewTxn()
	txn.Put(key101, []byte("val 101")).Put(key102, []byte("val 102"))
	txn.Put(key103, []byte("val 103")).Put(key104, []byte("val 104"))
	txn.Delete(key101)
	err := txn.Commit()
	if err != nil {
		log.Errorf("txn: %s", err)
	}
}

func generateServerConfig(path string) {
	config := redis.NodeClientConfig{
		Endpoint: "localhost:6379",
		Pool: redis.ConnPoolConfig{
			MaxIdle:     10,
			MaxActive:   10,
			IdleTimeout: 60,
			Wait:        true,
		},
	}
	redis.GenerateConfig(config, path)
}
