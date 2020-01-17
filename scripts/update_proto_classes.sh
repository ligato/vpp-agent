#!/usr/bin/env bash
#todo add option to choose proto files to pythonize and create plugins automatically
cd ..
rm -rf ansible/action_plugins/pout
mkdir pout
protoc --proto_path=vendor --proto_path=api --python_out=pout api/models/vpp/interfaces/interface.proto api/models/vpp/ipsec/ipsec.proto api/models/vpp/l3/route.proto api/models/vpp/l2/bridge-domain.proto api/models/vpp/nat/nat.proto api/models/vpp/l2/xconnect.proto api/models/vpp/l3/l3.proto api/models/linux/interfaces/interface.proto  api/models/linux/namespace/namespace.proto
mv pout ansible/action_plugins/.
cd ansible/action_plugins/pout
find . -type d -exec touch {}/__init__.py \;
