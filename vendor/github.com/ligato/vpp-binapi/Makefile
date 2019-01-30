VPP_REPO_URL ?= https://github.com/FDio/vpp
VPP_VERSION ?= $(shell cd vpp && git describe --always --tags --dirty)

generate: gen-vppapi gen-binapi apply-patches

gen-vppapi:
	@echo "=> generating vppapi (.api.json)"
	rm -rf vppapi/*
	find vpp -name \*.api -printf "echo %p \
	 && vpp/src/tools/vppapigen/vppapigen --includedir vpp/src \
	 --input %p --output vppapi/%f.json JSON\n" | xargs -0 sh -c

gen-binapi:
	@echo "=> generating binapi (Go code)"
	rm -rf binapi/*
	binapi-generator --input-dir=vppapi --output-dir=binapi --include-apiver
#	docker run  --rm -v ${PWD}:/vpp-binapi golang:1.11 bash -c \
#		"go install -mod=readonly git.fd.io/govpp.git/cmd/binapi-generator && \
#		binapi-generator --input-dir=vppapi --output-dir=binapi -include-apiver && \
#		chown -R 1000:1000 binapi"
	@echo -e "\n=> binapi generated for VPP ${VPP_VERSION}"
	@echo ${VPP_VERSION} > VPP_VERSION

apply-patches:
	@echo "=> applying patches"
	find patches -maxdepth 1 -type f -name '*.patch' -exec patch --no-backup-if-mismatch -p1 -i {} \;

.PHONY: generate gen-vppapi gen-binapi apply-patches

