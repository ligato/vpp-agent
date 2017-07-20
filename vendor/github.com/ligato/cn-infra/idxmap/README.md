# IDX Map

The idxmap package provides enhanced mapping structure. In addition to basic built-in map structure,
it allows to create secondary indexes that can also be leveraged for lookup. One can also subscribe 
for changes and receive notification once an item is added or removed.

Function `RegisterName` adds value(item) into the mapping. In the function call the primary index(name) for
the item is specified. The values of the primary index are unique, if the name already exists, then the item
 is overwritten. To retrieve an item identified by the primary index, use `Lookup` function.
An item can be removed from the mapping by `UnregisterName` function. The names that are currently registered
can be retrieved by `ListNames`.
 
The constructor allows to define `createIndexes` function that extracts secondary indexes from stored
items. The function returns a map indexed by names of secondary indexes and values are extracted values
for the particular item. The values of secondary indexes, are not necessarily unique. To retrieve items 
based on secondary indexes use `LookupByMetadata`. In contrast to lookup by primary index, the function
may return multiple names.

```
 Primary Index                Item                                Secondary indexes
===================================================================================
   
   Eth1              +---------------------+                 { "IP" : ["192.168.2.1", "10.0.0.8"],
                     |  Status: Enabled    |                   "Type" : ["ethernet"]
                     |  IP: 192.168.2.1    |                 }
                     |      10.0.0.8       |
                     |  Type: ethernet     |
                     |  Desc: something    |
                     +---------------------+
```

`Watch` allows to define a callback that is called when a change in the mapping occurs. There is 
a helper function `ToChan` available, which allows to deliver notifications through a channel.
