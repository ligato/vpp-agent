package main

import (
	"fmt"
	"time"

	"github.com/ligato/cn-infra/agent"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/db/keyval/etcd"
	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/vpp-agent/examples/tutorials/04_kv-store/model"
)

//go:generate protoc --proto_path=model --gogo_out=model ./model/model.proto

func main() {
	// Create an instance of our plugin using its constructor.
	p := NewMyPlugin()

	// Create new agent with our plugin instance.
	a := agent.NewAgent(agent.AllPlugins(p))

	// Run starts the agent with plugins, wait until shutdown
	// and then stops the agent and its plugins.
	if err := a.Run(); err != nil {
		logging.Error(err)
	}
}

const keyPrefix = "/myplugin/"

// MyPlugin represents our plugin.
type MyPlugin struct {
	infra.PluginDeps
	KVStore     keyval.KvProtoPlugin
	watchCloser chan string
}

// NewMyPlugin is a constructor for our MyPlugin plugin.
func NewMyPlugin() *MyPlugin {
	p := &MyPlugin{
		watchCloser: make(chan string),
	}
	p.SetName("myplugin")
	p.Setup()
	// Initialize key-value store
	p.KVStore = &etcd.DefaultPlugin
	return p
}

// Init is executed on agent initialization.
func (p *MyPlugin) Init() error {
	if p.KVStore.Disabled() {
		return fmt.Errorf("KV store is disabled")
	}

	watcher := p.KVStore.NewWatcher(keyPrefix)

	// Start watching for changes
	err := watcher.Watch(p.onChange, p.watchCloser, "greetings/")
	if err != nil {
		return err
	}

	return nil
}

// Init is executed after agent initialization.
func (p *MyPlugin) AfterInit() error {
	go p.updater()
	return nil
}

func (p *MyPlugin) onChange(resp datasync.ProtoWatchResp) {
	value := new(model.Greetings)
	// Deserialize data
	if err := resp.GetValue(value); err != nil {
		p.Log.Errorf("GetValue for change failed: %v", err)
		return
	}
	p.Log.Infof("%v change: KEY: %q VALUE: %+v", resp.GetChangeType(), resp.GetKey(), value)
}

func (p *MyPlugin) updater() {
	broker := p.KVStore.NewBroker(keyPrefix)

	// Wait few seconds
	time.Sleep(time.Second * 2)

	// Prepare data
	value := &model.Greetings{
		Greeting: "Hello",
	}

	// Update data in KV store
	p.Log.Infof("updating value")
	if err := broker.Put("greetings/hello", value); err != nil {
		p.Log.Errorf("Put failed: %v", err)
	}
}
