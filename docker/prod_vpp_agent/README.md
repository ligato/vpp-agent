## Production Docker Image

This image is lightweight version of dev_vpp_agent image. It contains:

- `vpp-agent`, `agentctl` and `vpp-agent-ctl` binaries,
- installed VPP ready to be used.

#### Building Locally
At first you need to have built `dev_vpp_agent` image.
To build production image on your local machine, type:
```
./build.sh
```
This will build `prod_vpp_agent` image with agent and vpp files taken from dev image.

#### Verifying a Created Image
You can verify the newly built image as follows:

```
docker images
``` 

You should see something this:

```
REPOSITORY                                            TAG                 IMAGE ID            CREATED             SIZE
prod_vpp_agent                                        latest              e33a5551b504        7 days ago          404 MB
...
```
Get the details of the newly built image:

```
docker image inspect prod_vpp_agent
docker image history prod_vpp_agent
```

### Shrinking the Image
Prod_vpp_agent image can be shrunk by typing command:

```
./shrink.sh
```

This will build a new image with the name `prod_vpp_agent_shrink` which has removed installation-related files 
(about 150MB).
It is using docker export and import command, but due to 
[the Docker issue](https://github.com/moby/moby/issues/26173) it will fail on docker older than 1.13.

```
REPOSITORY                                            TAG                 IMAGE ID            CREATED             SIZE
prod_vpp_agent_shrink                                 latest              446b271cce26        2 days ago          257 MB
prod_vpp_agent                                        latest              e33a5551b504        2 days ago          404 MB

```
---

### Starting the Image
By default, the VPP & the Agent processes will be started automatically in the container. This is useful
e.g. for deployments with Kubernetes, as described in [this README](../k8s/README.md). However,
this option is not really demanded for development purposes. This can be overridden by specifying
another container entrypoint, e.g. bash, as we do in the following steps in this README.

To start the image, type:
```
sudo docker run -it --name vpp_agent --privileged --rm prod_vpp_agent bash
```
To open another terminal:
```
sudo docker exec -it vpp_agent bash
```

### Running VPP and the Agent
You can use the image the same way as the dev image, see [this README](../dev_vpp_agent/README.md).