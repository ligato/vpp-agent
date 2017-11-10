# Building VPP for Travis _(Ubuntu 14.04)_

1. Update `VPP_COMMIT` in file `Dockerfile`

2. Build the image `docker build -t ligato/vppdeb .`

3. Tag the image:  `docker tag ligato/vppdeb:latest ligato/vppdeb:VPP_TAG`

   **VPP_TAG** is first 7 characters of **VPP_COMMIT**

4. Push the image to the docker hub: `docker push ligato/vppdeb:VPP_TAG`
