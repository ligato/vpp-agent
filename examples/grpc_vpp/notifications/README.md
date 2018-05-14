# GRPC notification server example

The example shows how to use vpp-agent to receive VPP notifications from the vpp-agent. 
Vpp-agent streams VPP notifications to all servers provided via grpc.conf file. 

How to run example:
1. Run vpp-agent with GRPC endpoint set.

```
vpp-agent --grpc-config=/opt/vpp-agent/dev/grpc.conf
```

2. Run GRPC server (example):
```
go run main.go
```

Two flags can be set:
* `-address=<address>` - for grpc server address (otherwise localhost will be used)
* `request-period=<time_in_sec>` - time between grpc requests

The example prints all received VPP notifications.