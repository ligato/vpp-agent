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
