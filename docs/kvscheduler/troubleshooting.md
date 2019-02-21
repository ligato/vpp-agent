# KVScheduler Troubleshooting Guide

**WORK IN-PROGRESS**

A where relevant, link how-to:
   - run agent with debug logs INITIAL_LOGLVL=debug (not kvscheduler-specific)
   - disaplay loaded plugins with DEBUG_INFRA=lookup (not kvscheduler-specific)
   - read transaction logs
   - display and read the graph visualization (after txn and current)
   - run in the verification mode (KVSCHED_VERIFY_MODE)
   - understand graph walk-through (advanced) - add picture

B common errors:
    - `value (...) has invalid type for key: ...`
    - `metadata has invalid type for key: ...`
    - `failed to retrieve values, refresh for the descriptor will be skipped`
    - `operation Create is not implemented`
    - `operation Delete is not implemented`
    
C common programming errors / bad practises:
    - changing the value content (should be manipulated as if it was mutable)
       - metadata can be edited in-place, but only in Update()
    - implementing stateful descriptors
       - especially the relations (derived, depends-on) should only be determined
         from the value itself (not even from metadata)     
    - trying to use metadata with derived values
    - deriving the same key from different values 
    - not using models for non-derived values or mixing them with custom key
      building/parsing methods
        - models are mandatory
    - using internal caches instead of metadata
    - leaving descriptor methods which are not needed defined (too much
      copy-paste from prepared skeletons)
    - unsupported Retrieve defined to always return empty set of values
    - in ValueComparator, comparing parts of the value which were derived out
    - implementing Retrieve method for descriptor with only derived values
      in its scope
    - not implementing Retrieve method for values announced to KvScheduler
      as OBTAINED via notifications
    - sleeping/blocking inside descriptor methods
    - exposing metadata with write access (only kvscheduler should edit)
    

D various common problems:
   1. value requested via NB is not configured
    1.1. txn is not triggered or the value is missing in the transaction input
      - check if model is registered - link common error of not using models
      - check if plugin is loaded
      - check if descriptor is registered
      - check if the prefix is watched (probably only using orchestrator logs)
      - check if the key of the value is right
         - bad prefix (bad suffix would make it unimplemented)
    ```
      DEBU[0005] => received RESYNC event (1 prefixes)         loc="orchestrator/orchestrator.go(150)" logger=orchestrator.dispatcher
      DEBU[0005]  -- key: config/mock/v1/interfaces/tap1       loc="orchestrator/orchestrator.go(168)" logger=orchestrator.dispatcher
      DEBU[0005]  -- key: config/mock/v1/interfaces/loopback1  loc="orchestrator/orchestrator.go(168)" logger=orchestrator.dispatcher
      DEBU[0005] - "config/mock/v1/interfaces/" (2 items)      loc="orchestrator/orchestrator.go(173)" logger=orchestrator.dispatcher
      DEBU[0005] 	 - "config/mock/v1/interfaces/tap1": (rev: 0)  loc="orchestrator/orchestrator.go(178)" logger=orchestrator.dispatcher
      DEBU[0005] 	 - "config/mock/v1/interfaces/loopback1": (rev: 0)  loc="orchestrator/orchestrator.go(178)" logger=orchestrator.dispatcher
      DEBU[0005] Resync with 2 items                           loc="orchestrator/orchestrator.go(181)" logger=orchestrator.dispatcher
      DEBU[0005] Pushing data with 2 KV pairs (source: watcher)  loc="orchestrator/dispatcher.go(67)" logger=orchestrator.dispatcher
      DEBU[0005]  - PUT: "config/mock/v1/interfaces/tap1"      loc="orchestrator/dispatcher.go(78)" logger=orchestrator.dispatcher
      DEBU[0005]  - PUT: "config/mock/v1/interfaces/loopback1"   loc="orchestrator/dispatcher.go(78)" logger=orchestrator.dispatcher
    ```
    
    ```
      DEBU[0012] => received CHANGE event (1 changes)          loc="orchestrator/orchestrator.go(121)" logger=orchestrator.dispatcher
      DEBU[0012] Pushing data with 1 KV pairs (source: watcher)  loc="orchestrator/dispatcher.go(67)" logger=orchestrator.dispatcher
      DEBU[0012]  - UPDATE: "config/mock/v1/interfaces/tap2"   loc="orchestrator/dispatcher.go(93)" logger=orchestrator.dispatcher
    ```   
      
    1.2 txn was triggered
      1.2.1 the value is pending
             - display graph and check dependencies (black arrow coming out)
             - either dependency is missing or has failed
               - could be that the plugin implementing the dependency is not loaded
             - unintended dependency added - check Dependencies method
      1.2.2 the value is in the unimplemented state
             - invalid key suffix (i.e. prefix is good, but the section build using value primary fields is bas)
             - or if KeySelector, NBKeyPrefix do not use the model
                - e.g. NBKeyPrefix of this or another descriptor selects the value,
                  but KeySelector does not
      1.2.3 the value failed to get applied
              - display the graph after txn - can be failed/reverted, failed/retrying, invalid
              - common error: "value has invalid type for key" - caused by mismatch between
                descriptor and the model
              - maybe missing dependency (i.e. ordering issue - check docs for SB)
      1.2.4 derived value is treated as PROPERTY when it should have CRUD operations assigned        
           
      
      
   x. Resync triggers some operations even if the SB is in fact in-sync with NB
       - consider running in the verification mode to check for CRUD inconsistencies
       - either inconsistent CRUD, or
           - ValueComparator is missing some equivalency or default/undefined value
           - ValueComparator is not plugged into the Descriptor

   x. Resync tries to create objects which already exist
       - forgot to plug Retrieve, or it has returned an error 
       ```
       ERRO[0005] failed to retrieve values, refresh for the descriptor will be skipped  descriptor=mock-interface loc="kvscheduler/refresh.go(104)" logger=kvscheduler
       ```
          - values not listed in the GRAPH dump (TODO: add example with the dump)
             
       
   x. Resync removes item not configured by NB
       - the object should be Retrieved with FromSB
       - UnknownOrigin can also be used (defaults to FromSB when history is empty)       
   
   x. Retrieve fails to find associated object in the metadata from another plugin
       - missing RetrieveDependency
       
   x. Value re-created when just Update should be called instead
       - either forgot to plug

   x. Metadata-related problems:
       - Retrieve fails to find associated object in the metadata from another plugin
                - missing RetrieveDependency
       - metadata are unexpectedly nil 
          - WithMetadata is null (not enough to define the factory)
          - metadata for derived values are not supported
          - forgot to return the new metadata in Update even if they have not changed
                     
       
   x. Unexpected transaction plan (e.g. wrong ordering)
       - check graph visualization:
           - if the derived values and dependencies (relations) are as expected
           - if any of the values is in an unexpected state 
       - explain graph walk (advanced) - how to display it - it could help to
         find the divergence point        