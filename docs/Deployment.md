# Deployment

VPP Agent can run on any server where VPP is installed. It can run on bare
metal, in a VM, or in a container.
 
Benefits of putting a VPP-based VNF (bascially, just VPP and the VPP Agent) 
into a container are:
 * simplified upgrades and bringup/shutdown, and better scaling  
 * Container-based VNFs are in essence data plane lightweight and reusable 
   data plane microservices which can be used to build lrger systems and 
   applications
 * supports container healing 
 
## K8s integration
The following diagram shows VPP deployement in:
- Data Plane vSwitch
- Control Plane vSwitch (TBD [Contiv](http://contiv.github.io/) integration)
- VPP VNF Container
- Non-VPP Container

![K8s integration](imgs/k8s_deployment.png "VPP Agent - K8s integration")

K8s:
- starts/stops the containers on multiple hosts
- checks containers health (using probes - HTTP calls)

## NB (Nort-bound) configuration vs. deployment
VPP + Agent can be deployed in different environments. Several deployment 
alternatives are briefly described in the following sub-chapters. Regardless
of the deployment, the VPP Agent can be configured using the same Client v1
interface. There are three different implementations of the interface:
 - local client
 - remote client using Data Broker
 - remote client using GRPC

### Key Value Data Store for NB
The Control Plane using remote client writes configuration to the Data Store 
(tested with ETCD, Redis). VPP Agent watches particular key prefixes in Data 
Store using dbsync package.

![deployment with data store](imgs/deployment_with_data_store.png)
TBD links to the code

### GRPC 
The Control Plane using remote client sends configuration to the Data Store 
(tested with ETCD, Redis). VPP Agent watches particular key prefixes in Data 
Store using grpcsync package.

![grpc northbound](imgs/deployment_nb_grpc.png)
TBD links to the code

### Embedded deployment
VPP Agent can be embedded in a different project. For integration with 
[Contiv](http://contiv.github.io/) we use the embedded deployment. In this 
case the VPP Agent gets the configuration from Local client v1 through 
in-memory calls (Go API).

![embeded deployment](imgs/deployment_embeded.png)
TBD links to the code