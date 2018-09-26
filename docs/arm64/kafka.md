
## Running Kafka on Local Host

You can start Kafka in a separate container:
```
sudo docker run -p 2181:2181 -p 9092:9092 --name kafka --rm \
 --env ADVERTISED_HOST=172.17.0.1 --env ADVERTISED_PORT=9092 spotify/kafka
```
**Note for ARM64:**

There is no official spotify/kafka image for ARM64 platform.
You can build an image following steps at the [repository](https://github.com/spotify/docker-kafka#build-from-source).
However you need to modify the kafka/Dockerfile before building like this:
```
#FROM java:openjdk-8-jre
#arm version needs this....
FROM openjdk:8-jre
...
...
#ENV KAFKA_VERSION 0.10.1.0
#arm version needs this....
ENV KAFKA_VERSION 0.10.2.1
...
...
```
