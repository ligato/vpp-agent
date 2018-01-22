# Statuscheck example

### Requirements

To start the example you have to have ETCD running first.
if you don't have it installed locally you can use the following docker
image.
```
sudo docker run -p 22379:2379 --name etcd --rm \
    quay.io/coreos/etcd:v3.0.16 /usr/local/bin/etcd \
    -advertise-client-urls http://0.0.0.0:2379 \
    -listen-client-urls http://0.0.0.0:2379
```

It will bring up ETCD listening on port 2379 for client communication.

### Usage

In the example, the location of the ETCD configuration file is defined
with the `-etcdv3-config` argument or through the `ETCDV3_CONFIG`
environment variable.
By default, the application will try to search for `etcd.conf`
in the current working directory.
If the configuration file cannot be loaded, the initialization
of the etcd plugin will be skipped and the example scenario will thus
not execute in its entirety.

To run the example, type:
```
go run main.go [-etcdv3-config <config-filepath>]
```

The status of connection to etcd is printed once per second. You can stop
and start the etcd again. You can observe the state change in logs.

```
INFO[0008] Status[etcdv3] = state:OK last_change:1516188524 last_update:1516188524   loc="statuscheck-plugin/main.go(84)" logger=statuscheck-example
INFO[0009] Status[etcdv3] = state:OK last_change:1516188524 last_update:1516188524   loc="statuscheck-plugin/main.go(84)" logger=statuscheck-example
ERRO[0013] etcdv3 put error: context deadline exceeded   loc="etcdv3/bytes_broker_impl.go(272)" logger=etcdv3
ERRO[0013] etcdv3 error: context deadline exceeded       loc="etcdv3/bytes_broker_impl.go(337)" logger=etcdv3
ERRO[0016] etcdv3 put error: context deadline exceeded   loc="etcdv3/bytes_broker_impl.go(272)" logger=etcdv3
INFO[0016] Status[etcdv3] = state:OK last_change:1516188524 last_update:1516188532   loc="statuscheck-plugin/main.go(84)" logger=statuscheck-example
INFO[0016] Agent plugin state update.                    lastErr="context deadline exceeded" loc="statuscheck/plugin_impl_statuscheck.go(189)" logger=status-check plugin=etcdv3 state=error
ERRO[0019] etcdv3 put error: context deadline exceeded   loc="etcdv3/bytes_broker_impl.go(272)" logger=etcdv3
INFO[0021] Status[etcdv3] = state:ERROR last_change:1516188535 last_update:1516188535 error:"context deadline exceeded"   loc="statuscheck-plugin/main.go(84)" logger=statuscheck-example
INFO[0022] Status[etcdv3] = state:ERROR last_change:1516188535 last_update:1516188535 error:"context deadline exceeded"   loc="statuscheck-plugin/main.go(84)" logger=statuscheck-example
INFO[0023] Status[etcdv3] = state:ERROR last_change:1516188535 last_update:1516188535 error:"context deadline exceeded"   loc="statuscheck-plugin/main.go(84)" logger=statuscheck-example
INFO[0024] Status[etcdv3] = state:ERROR last_change:1516188535 last_update:1516188535 error:"context deadline exceeded"   loc="statuscheck-plugin/main.go(84)" logger=statuscheck-example
INFO[0025] Status[etcdv3] = state:ERROR last_change:1516188535 last_update:1516188535 error:"context deadline exceeded"   loc="statuscheck-plugin/main.go(84)" logger=statuscheck-example
INFO[0026] Agent plugin state update.                    lastErr=<nil> loc="statuscheck/plugin_impl_statuscheck.go(189)" logger=status-check plugin=etcdv3 state=ok
INFO[0026] Status[etcdv3] = state:OK last_change:1516188546 last_update:1516188546   loc="statuscheck-plugin/main.go(84)" logger=statuscheck-example
INFO[0027] Status[etcdv3] = state:OK last_change:1516188546 last_update:1516188546   loc="statuscheck-plugin/main.go(84)" logger=statuscheck-example
```

