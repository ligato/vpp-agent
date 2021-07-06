module go.ligato.io/vpp-agent/v3

go 1.13

require (
	git.fd.io/govpp.git v0.3.6-0.20200907135408-e517439567ad
	github.com/Microsoft/go-winio v0.4.15-0.20200113171025-3fe6c5262873 // indirect
	github.com/Microsoft/hcsshim v0.8.9 // indirect
	github.com/Shopify/sarama v1.20.1 // indirect
	github.com/alecthomas/jsonschema v0.0.0-20200217214135-7152f22193c9
	github.com/alicebob/miniredis v2.5.0+incompatible // indirect
	github.com/common-nighthawk/go-figure v0.0.0-20200609044655-c4b36f998cf2
	github.com/containerd/containerd v1.3.4 // indirect
	github.com/containerd/continuity v0.0.0-20200413184840-d3ef23f19fbb // indirect
	github.com/coreos/go-iptables v0.4.5
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/docker/cli v0.0.0-20190822175708-578ab52ece34
	github.com/docker/docker v17.12.0-ce-rc1.0.20200505174321-1655290016ac+incompatible
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/fsouza/go-dockerclient v1.6.3
	github.com/ghodss/yaml v1.0.0
	github.com/go-errors/errors v1.0.1
	github.com/goccy/go-graphviz v0.0.6
	github.com/goccy/go-yaml v1.8.0
	github.com/golang/protobuf v1.5.2
	github.com/google/go-cmp v0.5.6
	github.com/google/uuid v1.1.2 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/iancoleman/orderedmap v0.0.0-20190318233801-ac98e3ecb4b0
	github.com/jhump/protoreflect v1.7.0
	github.com/lunixbochs/struc v0.0.0-20200521075829-a4cb8d33dbbe
	github.com/mitchellh/go-ps v0.0.0-20170309133038-4fdf99ab2936
	github.com/mitchellh/mapstructure v1.1.2
	github.com/moby/sys/mount v0.1.0 // indirect
	github.com/moby/term v0.0.0-20200429084858-129dac9f73f6 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/namsral/flag v1.7.4-pre
	github.com/olekukonko/tablewriter v0.0.4
	github.com/onsi/gomega v1.10.3
	github.com/opencontainers/runc v1.0.0-rc5 // indirect
	github.com/pkg/errors v0.9.1
	github.com/pkg/profile v1.5.0
	github.com/prometheus/client_golang v1.4.0
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v0.0.7
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.4.0
	github.com/unrolled/render v0.0.0-20180914162206-b9786414de4d
	github.com/vishvananda/netlink v0.0.0-20180910184128-56b1bd27a9a3
	github.com/vishvananda/netns v0.0.0-20180720170159-13995c7128cc
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonschema v1.1.0
	github.com/yuin/gopher-lua v0.0.0-20190514113301-1cd887cd7036 // indirect
	go.etcd.io/etcd v0.5.0-alpha.5.0.20210419091813-4276c3302675
	go.ligato.io/cn-infra/v2 v2.5.0-alpha.0.20210706123127-fc38c3cdecad
	go.uber.org/multierr v1.2.0 // indirect
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c
	google.golang.org/grpc v1.38.0
	google.golang.org/protobuf v1.27.1
	gotest.tools/v3 v3.0.2 // indirect
)

replace (
	github.com/coreos/etcd => go.etcd.io/etcd v0.5.0-alpha.5.0.20210419091813-4276c3302675
	go.etcd.io/etcd => github.com/coreos/etcd v0.5.0-alpha.5.0.20210419091813-4276c3302675

	google.golang.org/grpc => google.golang.org/grpc v1.29.1
)
