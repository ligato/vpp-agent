# CN-Infra

[![Build Status](https://travis-ci.org/ligato/cn-infra.svg?branch=master)](https://travis-ci.org/ligato/cn-infra)
[![Coverage Status](https://coveralls.io/repos/github/ligato/cn-infra/badge.svg?branch=master)](https://coveralls.io/github/ligato/cn-infra?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/ligato/cn-infra)](https://goreportcard.com/report/github.com/ligato/cn-infra)
[![GoDoc](https://godoc.org/github.com/ligato/cn-infra?status.svg)](https://godoc.org/github.com/ligato/cn-infra)
[![GitHub license](https://img.shields.io/badge/license-Apache%20license%202.0-blue.svg)](https://github.com/ligato/cn-infra/blob/master/LICENSE.md)

CN-Infra (cloud-native infrastructure) is a Golang platform for building
custom management/control plane applications for cloud-native Virtual 
Network Functions (VNFs). Cloud-native VNFs are also known as "CNFs". 

Each management/control plane app built on top of the CN-Infra platform is 
basically a set of modules called "plugins" in CN-Infra lingo, where each 
plugin provides a very specific/focused functionality. Some plugins are 
provided by the CN-Infra platform itself, some are written by the app's 
implementors. In other words, the CN-Infra platform itself is implemented
as a set of plugins that together provide the platform's functionality, 
such as logging, health checks, messaging (e.g. Kafka), a common front-end
API and back-end connectivity to various KV data stores (Etcd, Cassandra, 
Redis, ...), and REST and gRPC APIs. App writers can pick and choose only
those platform plugins that are required by their app; for example, if an
app does not need a KV store, the CN-Infra platform KV data store plugins
would not be included in the app. 

An example of a VNF control/management plane built on top of the CN-Infra
platform is the [VPP Agent](https://github.com/ligato/vpp-agent).


## Architecture

![arch](high_level_arch_cninfra.png "High Level Architecture of cn-infra")

The CN-Infra platform comprises a **[Core](core)** that provides plugin
lifecycle management (initialization and graceful shutdown of plugins) 
and a set of platform plugins. The platform plugins implement the following
functions:

* **RPC** - allows to expose application's API via REST or gRPC
* **DB** - provides a common API and connectivity to the various KV data 
    stores ([etcd](db/keyval/etcdv3), [Redis](db/keyval/redis), 
    [Casssandra](db/sql/cassandra))
* **Messaging** - provides a common API and connectivity to message buses 
    ([Kafka](messaging/kafka), ...)
* **Logging** - Integrated [Logrus](logging/logrus) for logging and a 
    [logmanager plugin](logging/logmanager) for setting of log level at 
    runtime. An app writer can create multiple loggers (for example, each 
    app plugin can have its own logger) and the log level for each logger
    can be set individually via a REST API.
* **[Health](statuscheck)** - Self health check mechanism between plugins 
    plus RPCs:
  *  probes (callable remotely from K8s)
  *  status (health check status) 

## Quickstart
The following code show the initialization/start of a simple agent application
built on the CN-Infra platform. The entire code can be found 
[here](examples/simple-agent/agent.go).
```
func main() {
	flavour := Flavour{}
	agent := core.NewAgent(logroot.Logger(), 15*time.Second, flavour.Plugins()...)

	err := core.EventLoopWithInterrupt(agent, nil)
	if err != nil {
		os.Exit(1)
	}
}
```

GoDoc can be browsed [online](https://godoc.org/github.com/ligato/cn-infra).

## Available CN-Infra Plugins

The repository contains following plugins:

- [Logging](logging/plugin) - generic skeleton that allows to create logger instance
  - [Logrus](logging/logrus) - implements logging skeleton using Logrus library
- [LogMangemet](logging/logmanager) - allows to modify log level of loggers using REST API
- [ServiceLabel](servicelabel) - exposes the identification string of the particular VNF
- [Keyval](db/keyval/plugin) - generic skeleton that provides access to a key-value datastore
  - [etcd](db/keyval/etcdv3) - implements keyval skeleton provides access to etcd
  - [redis](db/keyval/redis) - implements keyval skeleton provides access to redis
- [Kafka](messaging/kafka) - provides access to Kafka brokers
- [HTTPmux](httpmux) - allows to handle HTTP requests
- [StatusCheck](statuscheck) - allows to monitor the status of plugins and exposes it via HTTP
- [Resync](datasync/resync) - manages data synchronization in plugin life-cycle


## Contributing

If you are interested in contributing, please see the [contribution guidelines](CONTRIBUTING.md).