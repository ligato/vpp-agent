## CN-Infra Core

The `core` package contains the CN-Infra Core that manages the startup
and shutdown of an CN-Infra based management/control plane app. The 
`core` package also defines the CN-Infra Core's SPI that must be 
implemented by each plugin. The SPI is used by the Core to init, start
and shut down each plugin. 

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
