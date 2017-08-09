# Plugin Flavors 

Plugin Flavors:
1. A reusable combination of multiple plugins is called a 'ReusedFlavor'. 
   Consider the following code snippet. The ReusedFlavor structure is 
   basically a combination of Logrus, HTTP, LogManager, ServiceLabel, 
   StatusCheck, ETCD and Kafka plugins. All these plugins are implicitly
   instantiated. They do intentionally not contain pointers:
    1. to minimize the number of lines in the Flavor
    2. those plugins are not optional (if some of them would be, it would
       be a pointer)
    3. garbage collector ignores those field objects (since they are not 
       pointers - small optimization) 
2. The Inject() method contains hand written code (that is normally checked
   by the compiler rather than instantiated automatically by using reflection.
3. The Plugin() method returns a sorted list (slice) of plugins for agent 
   startup.
4. The CompositeFlavor example below demonstrates how to reuse ReusedFlavor
   in CompositeFlavor.

```go
package flavorexample

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/db/keyval/etcdv3"
	"github.com/ligato/cn-infra/httpmux"
	"github.com/ligato/cn-infra/logging/logmanager"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/messaging/kafka"
	"github.com/ligato/cn-infra/servicelabel"
	"github.com/ligato/cn-infra/statuscheck"
)

type CompositeFlavor struct {
	Basic    ReusedFlavor
	PluginXY PluginXY
	injected bool
}

func (Flavor *CompositeFlavor) Inject() error {
	if Flavor.injected {
		return nil
	}
	Flavor.injected = true
	if err := Flavor.Basic.Inject(); err != nil {
	    return err
	}

    Flavor.PluginXY.HTTP = &Flavor.Basic.HTTP
	// inject all other dependencies...
	
	return nil
}

func (Flavor *CompositeFlavor) Plugins() []*core.NamedPlugin {
	Flavor.Inject()
	return core.ListPluginsInFlavor(Flavor)
}

type ReusedFlavor struct {
	Logrus       logrus.Plugin
	HTTP         httpmux.Plugin
	LogManager   logmanager.Plugin
	ServiceLabel servicelabel.Plugin
	StatusCheck  statuscheck.Plugin
	Etcd         etcdv3.Plugin
	Kafka        kafka.Plugin

	injected bool
}

func (Flavor *ReusedFlavor) Inject() error {
	if Flavor.injected {
		return nil
	}
	Flavor.injected = true

	Flavor.HTTP.LogFactory = &Flavor.Logrus
	Flavor.LogManager.ManagedLoggers = &Flavor.Logrus
	Flavor.LogManager.HTTP = &Flavor.HTTP
	Flavor.Etcd.LogFactory = &Flavor.Logrus
	Flavor.Etcd.ServiceLabel = &Flavor.ServiceLabel
	Flavor.Etcd.StatusCheck = &Flavor.StatusCheck
	Flavor.Kafka.LogFactory = &Flavor.Logrus
	Flavor.Kafka.ServiceLabel = &Flavor.ServiceLabel
	Flavor.Kafka.StatusCheck = &Flavor.StatusCheck
	return nil
}

func (Flavor *ReusedFlavor) Plugins() []*core.NamedPlugin {
	Flavor.Inject()
	return core.ListPluginsInFlavor(Flavor)
}

type PluginXY struct {
    HTTP httpmux.HttpHandlers
}

func (plugin* PluginXY) Init() error {
    // do something
    return nil
}

func (plugin* PluginXY) Close() error {
    // do something
    return nil
}
```
