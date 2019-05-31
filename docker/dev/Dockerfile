ARG VPP_IMG
FROM ${VPP_IMG} AS vppimg

RUN dpkg-query -f '${Version}' -W vpp > /vpp/version

FROM golang:1.11 AS verify-binapi

RUN apt-get update && apt-get install -y --no-install-recommends \
		patch \
 	&& rm -rf /var/lib/apt/lists/*

WORKDIR /go/src/github.com/ligato/vpp-agent

ARG VPP_BINAPI
ARG SKIP_CHECK

COPY plugins/vpp/binapi plugins/vpp/binapi
COPY vendor vendor
COPY scripts/genbinapi.sh scripts/genbinapi.sh
COPY Makefile vpp.env ./

COPY --from=vppimg /usr/share/vpp /usr/share/vpp
COPY --from=vppimg /vpp/version /vpp-version

RUN cp -r plugins/vpp/binapi /tmp/orig_binapi && \
 	make generate-binapi && \
 	diff -r plugins/vpp/binapi /tmp/orig_binapi || \
 	{ \
 		vpp_version="$(cat /vpp-version)"; \
 		echo >&2 "---------------------------------------------------------------"; \
 		echo >&2 " Generated binapi does not match with used version!"; \
 		echo >&2 " VPP version: ${vpp_version}"; \
 		echo >&2 "---------------------------------------------------------------"; \
 		[ -n "$SKIP_CHECK" ] && exit 0; \
 	}

FROM vppimg AS devimg

RUN apt-get update && apt-get install -y --no-install-recommends \
		build-essential \
		ca-certificates \
		curl \
		git \
		graphviz \
		iproute2 \
		iputils-ping \
		make \
		nano \
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
 	wget -nv -O go.tgz "https://golang.org/dl/go${GOLANG_VERSION}.${goRelArch}.tar.gz"; \
 	tar -C /usr/local -xzf go.tgz; \
 	rm go.tgz;

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

# Install debugger
RUN \
	go get -u github.com/go-delve/delve/cmd/dlv && dlv version; \
	go get -u github.com/golang/dep/cmd/dep && dep version;

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

COPY . ./

RUN VERSION=$VERSION COMMIT=$COMMIT DATE=$DATE make install

ENV SUPERVISOR_CONFIG=/opt/vpp-agent/dev/supervisor.conf

# run supervisor as the default executable
CMD rm -f /dev/shm/db /dev/shm/global_vm /dev/shm/vpe-api && \
    exec vpp-agent-init
