ARG VPP_IMG=ligato/vpp-base:1904
FROM ${VPP_IMG} AS vppimg

FROM golang:1.11 AS verify-binapi

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
		patch \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /go/src/github.com/ligato/vpp-agent

COPY plugins/vpp/binapi plugins/vpp/binapi
COPY vendor vendor
COPY Makefile vpp.env ./

COPY --from=vppimg /usr/share/vpp /usr/share/vpp
COPY --from=vppimg /vpp/version /vpp/version

RUN cp -r plugins/vpp/binapi /tmp/orig_binapi \
 && make generate-binapi \
 && diff -r plugins/vpp/binapi /tmp/orig_binapi \
 || ( echo "Generated binapi code does not match with VPP version: `cat /vpp/version`!"; exit 1 )

FROM ${VPP_IMG} AS devimg

RUN echo "VPP version: `cat /vpp/version`"

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
		python \
		sudo \
		supervisor \
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
 /opt/vpp-agent/dev/

COPY ./docker/dev/vpp.conf /etc/vpp/vpp.conf
COPY ./docker/dev/supervisord.conf /etc/supervisord/supervisord.conf
COPY ./docker/dev/supervisord_kill.py /usr/bin/

# Build agent
WORKDIR $GOPATH/src/github.com/ligato/vpp-agent

# this prevents buildkit from skipping verify-binapi stage
COPY --from=verify-binapi $GOPATH/src/github.com/ligato/vpp-agent/plugins/vpp/binapi plugins/vpp/binapi

COPY . ./

ARG VERSION=dev
ARG COMMIT
ARG DATE

RUN VERSION=$VERSION COMMIT=$COMMIT DATE=$DATE make install

# run supervisor as the default executable
CMD rm -f /dev/shm/db /dev/shm/global_vm /dev/shm/vpe-api && \
    exec /usr/bin/supervisord -c /etc/supervisord/supervisord.conf


FROM