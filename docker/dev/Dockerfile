ARG VPP_IMG
FROM ${VPP_IMG} AS vppimg

RUN dpkg-query -f '${Version}' -W vpp > /vpp/version

FROM golang:1.17 AS verify-binapi

RUN apt-get update && apt-get install -y --no-install-recommends \
		patch \
 	&& rm -rf /var/lib/apt/lists/*

WORKDIR /src/ligato/vpp-agent

ARG VPP_IMG
ARG VPP_VERSION
ARG SKIP_CHECK

COPY go.mod go.sum ./
COPY Makefile vpp.env ./
RUN make get-binapi-generators

COPY scripts/genbinapi.sh ./scripts/genbinapi.sh
COPY plugins/vpp/binapi ./plugins/vpp/binapi

COPY --from=vppimg /usr/share/vpp /usr/share/vpp
COPY --from=vppimg /vpp/version /vpp-version

RUN set -x; \
	gofmt -w plugins/vpp/binapi; \
	cp -r plugins/vpp/binapi /tmp/orig_binapi && \
	make generate-binapi && \
 	diff --color=always -I="\/\/.*" -r plugins/vpp/binapi /tmp/orig_binapi || \
 	{ \
 		set +ex; \
 		vpp_version="$(cat /vpp-version)"; \
 		echo >&2 "==============================================================="; \
 		echo >&2 "!!! VPP BINARY API CHECK FAILED !!!"; \
 		echo >&2 "==============================================================="; \
 		echo >&2 " - VPP installed: ${vpp_version}"; \
 		echo >&2 " - VPP_VERSION:   ${VPP_VERSION}"; \
 		echo >&2 " - VPP_IMG:       ${VPP_IMG}"; \
 		echo >&2 "---------------------------------------------------------------"; \
 		echo >&2 "Generated binapi does not seem to be up-to-date with used VPP/GoVPP!"; \
 		echo >&2 ""; \
 		echo >&2 "$(diff -q --color=always -r plugins/vpp/binapi /tmp/orig_binapi)"; \
 		echo >&2 ""; \
 		echo >&2 " This might happen when VPP API change gets merged to a branch of used VPP."; \
 		echo >&2 " Ensure that VPP base image is compatible with the selected VPP version!"; \
 		echo >&2 "---------------------------------------------------------------"; \
 		echo >&2 " To resolve this now, you could:"; \
 		echo >&2 ""; \
 		echo >&2 "  1. Ignore this check by setting: SKIP_CHECK=y"; \
 		echo >&2 "  2. Override used VPP base image by setting: VPP_IMG=ligato/vpp-base:<TAG>"; \
 		echo >&2 "---------------------------------------------------------------"; \
 		[ -n "$SKIP_CHECK" ] && { \
 			echo >&2 "SKIP_CHECK set, ignoring check failure!"; \
 			exit 0; \
 		}; \
 	}

FROM vppimg AS devimg

RUN apt-get update && apt-get install -y --no-install-recommends \
		build-essential \
		ca-certificates \
		curl \
		git \
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
ENV GOLANG_VERSION 1.17.6
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
RUN mkdir -p "$GOPATH/bin" && chmod -R 777 "$GOPATH"

# Install debugger
RUN go install github.com/go-delve/delve/cmd/dlv@v1.9.0 && dlv version

# Copy configs
COPY \
    ./docker/dev/etcd.conf \
    ./docker/dev/grpc.conf \
    ./docker/dev/logs.conf \
    ./docker/dev/vpp-ifplugin.conf \
    ./docker/dev/supervisor.conf \
 /opt/vpp-agent/dev/

COPY ./docker/dev/vpp.conf /etc/vpp/vpp.conf
COPY ./docker/dev/init_hook.sh /usr/bin/

# handle differences in vpp.conf which are between supported VPP versions
ARG VPP_VERSION
COPY ./docker/dev/legacy-nat.conf /tmp/legacy-nat.conf
RUN if [ "$VPP_VERSION" -le 2009 ]; then \
		cat /tmp/legacy-nat.conf >> /etc/vpp/vpp.conf; \
	fi; \
	rm /tmp/legacy-nat.conf

# Install agent
WORKDIR /src/ligato/vpp-agent

COPY go.mod go.sum ./
RUN go mod download
COPY . ./

ARG VERSION
ARG COMMIT
ARG BRANCH
ARG BUILD_DATE
RUN make install purge

ENV SUPERVISOR_CONFIG=/opt/vpp-agent/dev/supervisor.conf

CMD rm -f /dev/shm/db /dev/shm/global_vm /dev/shm/vpe-api && \
	mkdir -p /run/vpp && \
	exec vpp-agent-init
