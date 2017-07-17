# Key-value datastore

The package defines API for access to key-value data store. `Broker` interface allows to read and manipulate key-value pairs.
`Watcher` provides functions for monitoring of changes in a datastore. Both interfaces are available with arguments
 of type `[]bytes` and `proto.Message`.

The package also provides a skeleton for a key-value plugin. The particular data store is selected
 in the constructor `NewSkeleton` using argument of type `CoreBrokerWatcher`. The skeleton handles the plugins life-cycle
 and provides unified access to datastore implementing `KvPlugin` interface.
