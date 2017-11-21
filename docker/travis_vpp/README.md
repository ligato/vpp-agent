# Building VPP for Travis _(Ubuntu 14.04)_

1. Checkout wanted commit in git submodule `vpp`.

2. Run `./build.sh` script.
  - this will read the commit from `vpp` submodule
  - build the `ligato/vppdeb` image
  - push the image with proper tag to Dockerhub
