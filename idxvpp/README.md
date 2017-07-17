# IDX Map

The `idxmap` is a Core Agent Plugin that allows Agent plugins that interact 
with VPP to map between VPP interface handles (the `sw_if_index` entities) 
and the string-based object identifiers used by northbound clients of the 
Agent. It is implemented as a collection of databases, also known as 
registries, used for a storage of `index->name` pairs. In addition to basic
 mapping, it is possible to store auxiliary metadata for each pair. A base implementation,
providing a quick in-memory registration and lookup, is shipped together with
a decorator enhancing the registry with persistent storage capabilities.

The mapping are used to implement the re-configuration, and state 
re-synchronization after failures. Furthermore, a registry may be shared 
between plugins. For example, `ifplugin` exposes the `sw_if_index->name` 
mapping so that other plugins may reference interfaces from objects that 
depend on them (such as bridge domains, IP routes, etc.)

While `idxmap` was primarily designed for VPP plugins, it is also available
for other plugins that need to maintain `integer->string` pairs and the ETCD
data store would be an overkill or simply not suitable for other reasons.

**API**

*Configuration*

All registries use the same configuration defined in a YAML configuration 
file whose location can be specified by the `idxmap-config` agent command 
line option, or through the `IDXMAP_CONFIG` environment variable.

Currently, the configuration includes only a single section
(`persistent-storage`) with parameters specifying the behaviour of the 
persistent storage:

  * `location` (`/var/vnf-agent/idxmap` be default): location, i.e. the 
    directory path, for the persistent storage of index-to-name maps
  * `sync-interval` (`10s` by default): how often (in nanoseconds) to flush 
     the underlying registry into the persistent storage
  * `max-sync-start-delay` (`3s` by default): to evenly distribute I/O load,
    the start of the periodic synchronization for a given index-to-name map 
    gets delayed by a random time duration. This constant defines the maximum
    allowed delay in nanoseconds.

Example idxmap configuration file:
```
persistent-storage:
  location: /tmp
  max-sync-start-delay: 2000000000
  sync-interval: 5000000000
```
  
*Registry*

Every plugin is allowed to allocate a new registry using either the function 
`NewNameToIdxRW(owner, title)`, giving in-memory-only storage capabilities, 
or by using the function `NewNameToIdxRWPersistent(owner, title)`, returning
a persistently backed registry. In order to query registry by fields in metadata
constructor `NewNameToIdxMemWithIndexes(owner, title, indexFunction)` can be leveraged.

For the persistent storage to function
properly, each registry must have a unique name assigned to it. The name
must be unique across all plugins. Even if the persistent storage is not 
used, the preferred way is to use a unique name to facilitate logging and 
debugging. 

The registry interface supports read and write operations. While the registry
owner is allowed to do both reads and writes, only the read interface is 
typically exposed to other plugins. See for example the `sw_if_index->name` 
mapping maintained by the `ifplugin`. Its read-only interface supports index-
by-name and name-by-index lookups with the `LookupIdx` and `LookupName` 
functions. Additionally, it is possible to watch for changes in the registry
by using the `WatchNameToIdx` functions. The write access is used by the
registry owner to register a new mapping using the `RegisterName` functions,
and to remove existing pair with the `UnregisterName` function.

**Example**

Here is a simplified code snippet from `ifplugin` showing how to use the 
`sw_if_index->name` registry:

```
// Plugin allocates new registries by its name and automatically becomes
// their owner.
const PluginID pluginapi.PluginName = "ifplugin"

// InterfaceMeta defines the attributes of metadata as used by the 
// interface plugin.
type InterfaceMeta struct {
	InterfaceType intf.InterfaceType
}

// Init initializes the interface plugin
func (plugin *InterfaceConfigurator) Init() {
    // Allocate registry for sw_if_index to name mappings.
    plugin.swIfIndexes, err = idxmap.NewNameToIdxRWPersistent(PluginID, "sw_if_indexes")
    if err != nil {
        // handle error
    }
    
    // Continue with the initialization...
}

// ConfigureInterface configures a new VPP or Linux interface.
func (plugin *InterfaceConfigurator) ConfigureInterface(iface *intf.Interfaces_Interface) {
    // Create the interface ...
    // ifIdx := ...
    
    
    // Once a new interface is created in VPP/Linux, add new mapping into the registry
    // if it doesn't exist yet
    _, _, found := plugin.SwIfIndexes.LookupName(ifIdx)
    if !found {
        plugin.SwIfIndexes.RegisterName(iface.Name, ifIdx, &InterfaceMeta{iface.Type})
    }
}

// DeleteInterface removes an existing VPP or Linux interface.
func (plugin *InterfaceConfigurator) DeleteInterface(iface *intf.Interfaces_Interface) {
    // Delete the interface ...
    
    // When the interface gets deleted from VPP/Linux, the mapping must be removed as well.
    plugin.SwIfIndexes.UnregisterName(iface.Name)
}
```
