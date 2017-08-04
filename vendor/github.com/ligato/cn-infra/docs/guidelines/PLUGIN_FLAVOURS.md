# Plugin Flavours 

Plugin Flavours:
1. Reusable combination of multiple plugins is called ReusedFlavour. See the following code snipped. The structure ReusedFlavour 
is basicly combination of Logrus, HTTP, LogManager, ServiceLabel, StatusCheck, ETCD & Kafka plugins. All of these
plugins are instantiated implicit. The intentionally not contain pointers:
   1. to minimize number of lines in flavour
   2. those plugins are not optional (if some of the would be it would be a pointer)
   3. garbage collector ignores those field objects (since they are not pointer - small optimization) 
2. Method Inject() contains hand written code (that is normally checked by compiler rather than automatically by using reflection).
3. Method Plugin() returns sorted list (slice) of plugins for agent startup.
4. Reuse  CompositeFlavour demonstrates how to reuse ReusedFlavour in CompositeFlavour.

```go
package flavourexample

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

type CompositeFlavour struct {
	Basic    ReusedFlavour
	PluginXY PluginXY
	injected bool
}

func (flavour *CompositeFlavour) Inject() error {
	if flavour.injected {
		return nil
	}
	flavour.injected = true
	if err := flavour.Basic.Inject(); err != nil {
	    return err
	}

    flavour.PluginXY.HTTP = &flavour.Basic.HTTP
	// inject all other dependencies...
	
	return nil
}

func (flavour *CompositeFlavour) Plugins() []*core.NamedPlugin {
	flavour.Inject()
	return core.ListPluginsInFlavor(flavour)
}

type ReusedFlavour struct {
	Logrus       logrus.Plugin
	HTTP         httpmux.Plugin
	LogManager   logmanager.Plugin
	ServiceLabel servicelabel.Plugin
	StatusCheck  statuscheck.Plugin
	Etcd         etcdv3.Plugin
	Kafka        kafka.Plugin

	injected bool
}

func (flavour *ReusedFlavour) Inject() error {
	if flavour.injected {
		return nil
	}
	flavour.injected = true

	flavour.HTTP.LogFactory = &flavour.Logrus
	flavour.LogManager.ManagedLoggers = &flavour.Logrus
	flavour.LogManager.HTTP = &flavour.HTTP
	flavour.Etcd.LogFactory = &flavour.Logrus
	flavour.Etcd.ServiceLabel = &flavour.ServiceLabel
	flavour.Etcd.StatusCheck = &flavour.StatusCheck
	flavour.Kafka.LogFactory = &flavour.Logrus
	flavour.Kafka.ServiceLabel = &flavour.ServiceLabel
	flavour.Kafka.StatusCheck = &flavour.StatusCheck
	return nil
}

func (flavour *ReusedFlavour) Plugins() []*core.NamedPlugin {
	flavour.Inject()
	return core.ListPluginsInFlavor(flavour)
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