# Tutorial: KV Store

In this tutorial we will learn how to use key-value store to update data and watch for changes.

We will be using [Etcd][1] in this tutorial, but there are several other 
implementations of key-value store available: [Consul][2], [BoltDB][3], [FileDB][4], [Redis][5].

The common interface for all key-value store implementations is `KvProtoPlugin`, defined as:

```go
type KvProtoPlugin interface {
	NewBroker(keyPrefix string) ProtoBroker
	NewWatcher(keyPrefix string) ProtoWatcher
	Disabled() bool
	OnConnect(func() error)
	String() string
}
```

To use the KV store plugin we simply define field for it in our plugin and 
set the instance in our constructor to default Etcd plugin.

```go
type MyPlugin struct {
	infra.PluginDeps
	KVStore keyval.KvProtoPlugin
}

func NewMyPlugin() *MyPlugin {
	// ...
	p.KVStore = &etcd.DefaultPlugin
	return p
}
```

First, we need to create a new broker. The broker needs to be initialized with
a key prefix and for this example, we are going to use `/myplugin/` as a key prefix.
The broker uses this key prefix for all of its operations (Get, List, Put, Delete).

```go
broker := p.KVStore.NewBroker("/myplugin/")
```

Note: The KV store might be disabled, which usually happens if its config file 
is not found. In our case the `etcd.conf` file needs to be in the same folder.

Since, the broker accepts `proto.Message` in its methods, we are going to use
our Protobuf model `Greetings` defined in [`model.proto`][6] file.

```proto
message Greetings {
    string greeting = 1;
}
```

The Go code was generated using `go:generate` directive from the example.

```go
//go:generate protoc --proto_path=model --gogo_out=model ./model/model.proto
```

To update some value in a KV store we will use broker's `Put` method.

```go
value := &model.Greetings{
	Greeting: "Hello",
}
err := broker.Put("greetings/hello", value)
if err != nil {
	// handle error
}
```

The value above will be updated for key `/myplugin/greetings/hello`.

To retrieve some value from a KV store we will use broker's `GetValue` method.

```go
value := new(model.Greetings)
found, rev, err := broker.GetValue("greetings/hello", value)
if err != nil {
	// handle error
}else if !found {
	// handle not found
}
```

To watch for changes in the KV store we need to initialize a watcher.

```go
watcher := p.KVStore.NewWatcher("/myplugin/")
```

Then we need to define our callback function that will process the changes.

```go
onChange := func(resp datasync.ProtoWatchResp) {
	key := resp.GetKey()
	value := new(model.Greetings)
	if err := resp.GetValue(value); err != nil {
		// handle error
	}
	// process change
}
```

Now we can start watching for a key prefix(es).

```go
cancelWatch := make(chan string)
err := watcher.Watch(onChange, cancelWatch, "greetings/")
if err != nil {
	// handle error
}
```

The channel `cancelWatch` can be used to cancel watching.

Complete working example can be found at [examples/tutorials/04_kv-store](../../examples/tutorials/04_kv-store).

[1]: https://github.com/ligato/cn-infra/tree/master/db/keyval/etcd
[2]: https://github.com/ligato/cn-infra/tree/master/db/keyval/consul
[3]: https://github.com/ligato/cn-infra/tree/master/db/keyval/bolt
[4]: https://github.com/ligato/cn-infra/tree/master/db/keyval/filedb
[5]: https://github.com/ligato/cn-infra/tree/master/db/keyval/redis
[6]: ../../examples/tutorials/04_kv-store/model/model.proto
