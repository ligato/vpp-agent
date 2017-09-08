## Prepare Phase
Install and run Kubernetes, e.g. using [kubeadm](https://kubernetes.io/docs/setup/independent/create-cluster-kubeadm/).

Deploy ETCD and Kafka server PODs:
```
$ kubectl apply -f etcd.yaml
$ kubectl apply -f kafka.yaml
```

Verify the ETCD and Kafka PODs are ready:
```
$ kubectl get pods
NAME            READY     STATUS    RESTARTS   AGE
etcdv3-server   1/1       Running   0          12s
kafka-server    1/1       Running   0          5s
```

## Deploy Network Service
Import ETCD configuration for the given scenario:
```
$ sudo ./etcdimport.sh scenario1/etcd.txt
```

Deploy vSwitch + VNF PODs:
```
$ kubectl apply -R -f scenario1
```

Verify the PODs are ready:
```
$ kubectl get pods
NAME             READY     STATUS    RESTARTS   AGE
etcdv3-server    1/1       Running   0          33m
kafka-server     1/1       Running   0          6h
vnf-vpp          1/1       Running   0          26s
vswitch-vpp      1/1       Running   0          26s
```

## Verify Service is Up

Telnet to the vSwitch VPP:
```
$ telnet localhost 5002
```

Telnet to the VNF VPP:
```
$ kubectl describe pod vnf-vpp | grep IP
IP:		192.168.65.193
$ telnet 192.168.65.193 5002
```
(use `vnf1-vpp` / `vnf2-vpp` instead of `vnf-vpp` for the scenario 2)


## Cleanup
Undeploy the scenario:
```
$ kubectl delete -R -f scenario1/
```

Wipe ETCD (restart it):
```
$ kubectl delete -f etcd.yaml
$ kubectl apply -f etcd.yaml
```
(make sure the POD gets undeployed after `delete` before `apply`, by executing `kubectl get pods`)

