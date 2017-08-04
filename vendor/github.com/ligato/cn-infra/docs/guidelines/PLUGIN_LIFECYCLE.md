# Plugin Lifecytle

Each plugin implements Init() and Close() method (see the [plugin_api.go](../../core/pluginapi.go)) 
and optionally AfterInit(). Those methods are called sequncialy by (see the [agent_core.go](../../core/agent_core.go)).

There are following rules for implementing the methods:
## Init()
* Initialize maps & channels here to avoid nil pointers later.
* Process configs here (see the [Config Guidelins](CONFIG.md))
* Propagate errors. If an error occurs, agent stops since it is not properly initialized and calls Close() methods.
* Start watching the GO channels here (but not subscribed yet) in a go routine.
* Initialize GO lang Context & Cancel Function to stop go routines gracefully.

## AfterInit()
* Connect clients & start servers here (see the [System Integration Guidelines](SYSTEM_INTEGRATION.md))
* Propagate errors. Agent will stop because it is not properly initialized and calls Close() methods.
* Subscribe for watching data (see the go channel in the example below).

## Close()
* Cancel the go routines by calling GO lang (Context) Cancel function.
* Disconnect clients & stop servers here, release resources. For that try to use package [safeclose](../../utils/safeclose)
* Propagate errors. Agent will log those errors.

## Example
```go
package example
import (
    "errors"
    "context"
    "io"
    "github.com/ligato/cn-infra/datasync"
    "github.com/ligato/cn-infra/logging"
    "github.com/ligato/cn-infra/utils/safeclose"
)

type PluginXY struct {
    Watcher     datasync.Watcher //Injected
    Logger      logging.Logger
    ParentCtx   context.Context
    
    resource    io.Closer
    dataChange  chan datasync.ChangeEvent
    dataResync  chan datasync.ResyncEvent
    data        map[string]interface{}
    cancel      context.CancelFunc
}

func (plugin * PluginXY) Init() (err error) {
    //initialize the resource
    if plugin.resource, err = connectResouce(); err != nil {
        return err//propagate resource
    }
    
    
    // initialize maps (to avoid segmentation fault)
    plugin.data = make(map[string]interface{})
    
    // initialize channels & start go routines
    plugin.dataChange = make(chan datasync.ChangeEvent, 100)
    plugin.dataResync = make(chan datasync.ResyncEvent, 100)
    
    // initiate context & cancel function (to stop go routine)
    var ctx context.Context
    if plugin.ParentCtx == nil {
        ctx, plugin.cancel = context.WithCancel(context.Background())    
    } else {
        ctx, plugin.cancel = context.WithCancel(plugin.ParentCtx)
    }   
    
    go func() {
        for {
            select {
            case dataChangeEvent := <-plugin.dataChange:
                plugin.Logger.Debug(dataChangeEvent)
            case dataResyncEvent := <-plugin.dataResync:
                plugin.Logger.Debug(dataResyncEvent)
            case <-ctx.Done():
                // stop watching for notifications
                return
            }
        }
    }()
    
    return nil
}

func connectResouce() (resource io.Closer, err error) {
    // do something relevant here...
    return nil, errors.New("Not implemented")
}

func (plugin * PluginXY) AfterInit() error {
    // subscribe plugin.channel for watching data (to really receive the data)
    plugin.Watcher.WatchData("watchingXY", plugin.dataChange, plugin.dataResync, "keysXY")

    return nil
}

func (plugin * PluginXY) Close() error {
    // cancel watching the channels
    plugin.cancel()
    
    // close all resources / channels
    _, err := safeclose.CloseAll(plugin.dataChange, plugin.dataResync, plugin.resource)
    return err 
}
```


