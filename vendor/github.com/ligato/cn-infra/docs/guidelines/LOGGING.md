# Logging

## Overview
CN-Infra logging API is defined in the [logging package](../../logging/log_api.go).
The API defines interface of a logger, a log registry and a log factory.
Currently the only provided implementation of a logger is based on
[Logrus](https://github.com/sirupsen/logrus) and can be found
[here](../../logging/logrus/logger.go).

## Log Registry
The Logrus-based logger also ships with an implementation of both
the log registry and the log factory under one structure denoted
as [logRegistry](../../logging/logrus/registry.go).
The registry can be installed into a CN-Infra based applications
through the [local flavor](../../flavors/local).
On its own it allows to create a new logger with a given label and
maintain a local view of all active loggers.
The following example is a combination of code snippets presenting
the use of the registry in a user-defined plugin:
```go
/*** A very simple plugin which uses LogRegistry ***/
type Plugin struct {
	Deps
}

type Deps struct {
	LogRegistry  logging.Registry // inject
}

func (plugin *Plugin) Init() error {
    // Create a new logger
    plugin.LogRegistry.NewLogger("my-logger")
    // Set level for logging
    plugin.LogRegistry.SetLevel("my-logger", "debug")
    // Print all active loggers
    fmt.Printf("All registered loggers: %+v", plugin.LogRegistry.ListLoggers())
    return nil
}


/*** A flavor which composes the above plugin with the local flavor ***/
type Flavor struct {
	*local.FlavorLocal
	MyPlugin Plugin

	injected bool
}

func (f *Flavor) Inject() bool {
	if f.injected {
		return false
	}
	f.injected = true

	if f.FlavorLocal == nil {
		f.FlavorLocal = &local.FlavorLocal{}
	}
	f.FlavorLocal.Inject()

	f.MyPlugin.LogRegistry = f.FlavorLocal.LogRegistry()
	return true
}
```

The registry gets much more interesting in a combination with
the [RPC flavor](../../flavors/rpc).
The RPC flavor includes [log manager](../../logging/logmanager) which
provides a remote access to the log registry through REST API.
It is possible to view all active loggers via HTTP GET method and to set
the log level individually for each logger via HTTP POST method.

## Logging dependency
Plugins that need to use the logging capabilities should be defined
as dependent on [PluginLogger](../../logging/log_api.go).
Such dependency definition is already prepared and can be applied through
embedding from the structure
[PluginLogDeps](../../flavors/local/plugin_deps.go).
Plugins that interact with DB/Messaging may use
[PluginInfraDeps](../../flavors/local/plugin_deps.go) instead, which,
apart from logging, also includes some other basic dependencies likely
to be needed.

Example:
```go
import (
	"github.com/ligato/cn-infra/flavors/localdeps"
)

type Plugin struct {
	Deps
}

type Deps struct {
	localdeps.PluginInfraDeps
}
```

All plugin infra-dependencies can be fully satisfied by the local flavor.
The log registry is used to create a new logger referenced by the plugin
name and injected into the plugin as a struct member labelled **Log**
(inherited from PluginLogger).

For the injection of plugin infra-dependencies from the local flavor
there is already a helper method
[InfraDeps()](../../flavors/local/local_flavor.go) as presented in the
following example:
```go
type Flavor struct {
	*local.FlavorLocal
	GoVPP     govppmux.GOVPPPlugin // GoVPP uses PluginInfraDeps to ask for logging capabilities among other things

	injected bool
}

// Inject sets object references
func (f *Flavor) Inject() bool {
	if f.injected {
		return false
	}
	f.injected = true

	if f.FlavorLocal == nil {
		f.FlavorLocal = &local.FlavorLocal{}
	}
	f.FlavorLocal.Inject()

	// The only parameter of InfraDeps() is the name of the plugin which will be used as the label for the allocated logger.
	f.GoVPP.Deps.PluginInfraDeps = *f.FlavorLocal.InfraDeps("govpp")

	//...
	return true
}
```

Similarly, for PluginLogDeps the helper method
[LogDeps()](../../flavors/local/local_flavor.go) can be used to inject
the logger.

## Plugin Logger
The use of golang embedding from PluginLogger through
PluginLogDeps/PluginInfraDeps, plugin's own Deps and all the way
to the top definition of the plugin, allows to access the plugin logger
in a rather concise manner:
```go
 func (plugin *Plugin) Init() error {
 	plugin.Log.Debug("Initializing interface plugin")
 	// ...
 }
```

While plugins can still use the [DefaultLogger](../../logging/logrus)
(referenced as "defaultLogger"), it is recommended to use a separate
logger for each plugin through the dependency injection as explained in
[Logging dependency](#Logging dependency).
This gives the advantage of being able to set different log level for
each plugin. This is especially useful for debugging purposes as some
selected plugins may be switched to the Debug log level while others can
simultaneously remain in a less verbose mode.

For a more complicated plugin it can be preferred to split log messages
into multiple topics using **child loggers**. This is possible because
the injected plugin logger not only implements the logger API, but also
the interface of the log factory.
Method [NewLogger()](../../logging/log_api.go) allows to create a new
child logger with a name that gets prefixed with the plugin label:
```go
// Create a new child logger
childLogger := plugin.Log.NewLogger("childLogger")
// Usage of child loggers
childLogger.Infof("Log using named logger with name: %v", childLogger.GetName())
```

A full code example with plugin's own logger injected through
the use of PluginLogDeps and with one child logger can be found in
[examples/logs-plugin](../../examples/logs-plugin/main.go).