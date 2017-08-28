# Simple example

To start examples you have to have Cassandra running on localhost, if you don't have installed it locally
you can use the following docker image.
```
sudo docker run -p 9042:9042 --name cassandra01 -d cassandra:latest
```

In the example connection to etcd is configured using `--cfg` argument.
If the file is not specified  application tries to connect
 to Cassandra on localhost on default port 9042.
 
The example contains one program:
```
go run main.go <ClientConfigFilePath>
```
