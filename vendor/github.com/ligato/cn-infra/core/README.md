## Agent

The cn-infra provides building blocks to construct a management tool also known as an agent. The agent is designed to be
composed of multiple small blocks providing a certain functionality. These blocks are called plugins. Lifecycle of the plugins
is managed by the core. This package defines API between Agent Core and Agent Plugins (illustrated also on following diagram).

```
                                       +-----------------------+
+------------------+                   |                       |
|                  |                   |      Agent Plugin     |
|                  |                   |                       |
|    Agent Core    |                   +-----------------------+
|     (setup)      |        +--------->| Plugin global var     |
|                  |        |          |   + Init() error      |
|                  |        |          |   + AfterInit() error |
|                  |        |          |   + Close() error     |
|                  |        |          +-----------------------+
+------------------+        |
|                  +--------+          +-----------------------+
|    Init Plugin   |                   |                       |
|                  +--------+          |      Agent Plugin     |
+------------------+        |          |                       |
                            |          +-----------------------+
                            +--------->| Plugin global var     |
                                       |   + Init() error      |
                                       |   + AfterInit() error |
                                       |   + Close() error     |
                                       +-----------------------+
```

**Plugins**

The repository contains following plugins:

- [Logging](../logging/plugin) - generic skeleton that allows to create logger instance
  - [Logrus](../logging/logrus) - implements logging skeleton using Logrus library
- [LogMangemet](../logging/logmanager) - allows to modify log level of loggers using REST api
- [ServiceLabel](../servicelabel) - exposes the identification string of the particular VNF
- [Keyval](../db/keyval/plugin) - generic skeleton that provides access to a key-value datastore
  - [etcd](../db/keyval/etcdv3) - implements keyval skeleton provides access to etcd
  - [redis](../db/keyval/redis) - implements keyval skeleton provides access to redis
- [Kafka](../messaging/kafka) - provides access to Kafka brokers
- [HTTPmux](../httpmux) - allows to handle HTTP requests
- [StatusCheck](../statuscheck) - allows to monitor the status of plugins and exposes it via HTTP
- [Resync](../datasync/resync) - manages data synchronization in plugin life-cycle
 