# System Integration

System integration is about exposing services or consuming services, microservices or any other servers (including
database, message bus, RPC calls).

Please follow:
# Timeouts
Timeouts are very important when implementing system integration.

```
TODO link to code of db or messaging that allows to configure global timeout
```

```
TODO link to code of db or messaging that allows to configure method level timeout using varargs mehotd(args, WithTimout())
```

# Reconnection
It needs to be possible to reconnect after successful recovery of consumed service, if there was previously an established
connection.

```
TODO link to code/doc of db or messaging 
```

# AfterInit() failed to connect
If it is not able to connect during timeout, the plugin needs to propagate errors. The Agent will not start. 
TODO assuming that there is a default deployment strategy for container-based cloud (as is with K8s) that 
will try to heal the container and basically recreates it.

```
TODO link to code of db or messaging plugin that propagates error 
```
