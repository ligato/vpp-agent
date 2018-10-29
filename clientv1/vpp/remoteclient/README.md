# Remote client

Remote client enables remote management of VPP configuration. It is possible to remotely call 
database request using required broker, or via GRPC service.

To create a new data request object, call `NewDataRequestDB` (for database access) or `NewDataRequestGRPC`
(for GRPC). DB request helper required broker, GRPC request helper requires resync/change service client 
(as defined in RPC proto [model](/plugins/vpp/model/rpc)). Both methods return data request object with 
two methods:

* `Resync()` can be used to create data resync call
* `Change()` can be used to put od remove data 

Call is executed with `Send()`. The data request object does not need to be created anew after call execution,
it can be reused and ensures that all the future calls use the same parameters (broker, GRPC clients)

