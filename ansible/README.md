# Ansible Action Plugin for VPP ETCD

This is a plugin that contains generated python files out of
.proto files. Ansible plugin uses them to parse and validate
json that we want to submit to etcd database. These files need
to be regenerated in case of changes in .proto files

## Prerequisites

    protoc

### Install protoc

Install prerequisits for protoc:

    apt-get install autoconf automake libtool curl make g++ unzip

1. download the protobuf-all\[VERSION\].tar.gz.
https://github.com/protocolbuffers/protobuf/releases/tag/v3.6.1
2. Extract the contents and change in the directory
3. Run following commands (This might take several minutes)

        ./configure
        make
        make check
        sudo make install
        sudo ldconfig # refresh shared library cache.

Check your installation:

    protoc --version

Expect similar output to this:

    libprotoc 3.6.1


### Use protoc to regenerate the python modules

Run [update_proto_classes.sh](../scripts/update_proto_classes.sh) script
to update automatically created python files if needed. This script will
generate python objects used with ansible plugin to pout directory that
will replace [pout](action_plugins/pout).

In case we need to use new .proto files they need to be added to the
[update_proto_classes.sh](../scripts/update_proto_classes.sh) script.
and python validation module needs to be written in a similar way as
[interface.py](action_plugins/plugins/interface.py)
