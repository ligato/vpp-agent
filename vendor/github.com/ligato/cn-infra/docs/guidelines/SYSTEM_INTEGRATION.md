# System Integration

System integration is about exposing som services or consuming services microservices or any other servers (including
database, message bus, RPC calls).

Please follow:
# Timeouts
Timeouts are very important when doing system integration.

```
TODO link to code of db or messaging that allows to configure global timeout
```

```
TODO link to code of db or messaging that allows to configure method level timeout using varargs mehotd(args, WithTimout())
```

# Reconnection
It needs to be possible to reconnect after successful recovery of consumed service if there was previously established
connection.

```
TODO link to code/doc of db or messaging 
```

# AfterInit() failed to connect
Plugin needs to propagate errors if it is not able to connect during timeout. The Agent will not start. 
TODO Assuming that there is default deployment strategy for container base cloud (like with K8s) that 
will try to heal the container and basically recreates the container.

```
TODO link to code of db or messaging plugin that propagates error 
```