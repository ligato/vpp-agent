ARG VPP_IMG
FROM ${VPP_IMG} AS vppimg

FROM golang:1.11 AS verify-binapi

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
		patch \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /go/src/github.com/ligato/vpp-agent

ARG VPP_BINAPI

COPY plugins/vpp/binapi plugins/vpp/binapi
COPY vendor vendor
COPY scripts/genbinapi.sh scripts/genbinapi.sh
COPY Makefile vpp.env ./

COPY --from=vppimg /usr/share/vpp /usr/share/vpp

RUN cp -r plugins/vpp/binapi /tmp/orig_binapi \
 && make generate-binapi \
 && diff -r plugins/vpp/binapi /tmp/orig_binapi \
 || ( echo "Generated binapi does not match used VPP version!"; exit 1 )

FROM ${VPP_IMG} AS devimg

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
		build-essential \
		ca-certificates \
		curl \
		git \
		graphviz \
		iproute2 \
		iputils-ping \
		make \
		patch \
		sudo \
		unzip \
		wget \
 && rm -rf /var/lib/apt/lists/*

# Install Go
ENV GOLANG_VERSION 1.11.10
RUN set -eux; \
	dpkgArch="$(dpkg --print-architecture)"; \
		case "${dpkgArch##*-}" in \
			amd64) goRelArch='linux-amd64'; ;; \
			armhf) goRelArch='linux-armv6l'; ;; \
			arm64) goRelArch='linux-arm64'; ;; \
	esac; \
 	wget -qO go.tgz "https://golang.org/dl/go${GOLANG_VERSION}.${goRelArch}.tar.gz"; \
 	tar -C /usr/local -xzf go.tgz; \
 	rm go.tgz;

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

# Install debugger
#RUN \
#	go get -u github.com/go-delve/delve/cmd/dlv && dlv version; \
#	go get -u github.com/golang/dep/cmd/dep && dep version;

# Copy configs
COPY \
    ./docker/dev/etcd.conf \
    ./docker/dev/logs.conf \
    ./docker/dev/vpp-ifplugin.conf \
    ./docker/dev/supervisor.conf \
 /opt/vpp-agent/dev/

COPY ./docker/dev/vpp.conf /etc/vpp/vpp.conf
COPY ./docker/dev/init_hook.sh /usr/bin/

# Build agent
WORKDIR $GOPATH/src/github.com/ligato/vpp-agent

ARG VERSION
ARG COMMIT
ARG DATE

# this prevents buildkit from skipping verify-binapi stage
COPY --from=verify-binapi $GOPATH/src/github.com/ligato/vpp-agent/plugins/vpp/binapi plugins/vpp/binapi

COPY . ./

RUN VERSION=$VERSION COMMIT=$COMMIT DATE=$DATE make install

ENV SUPERVISOR_CONFIG=/opt/vpp-agent/dev/supervisor.conf

# run supervisor as the default executable
CMD rm -f /dev/shm/db /dev/shm/global_vm /dev/shm/vpe-api && \
    exec vpp-agent-init
