# install dependencies according to glide.yaml & glide.lock (in case vendor dir was deleted)
define install_dependencies
	$(if $(shell command -v glide install 2> /dev/null),$(info glide dependency manager is ready),$(error glide dependency manager missing, info about installation can be found here https://github.com/Masterminds/glide))
	@echo "# installing dependencies, please wait ..."
	@glide install --strip-vendor
endef

# clean update dependencies according to glide.yaml (re-downloads all of them)
define update_dependencies
	$(if $(shell command -v glide install 2> /dev/null),$(info glide dependency manager is ready),$(error glide dependency manager missing, info about installation can be found here https://github.com/Masterminds/glide))
	@echo "# updating dependencies, please wait ..."
	@-cd vendor && rm -rf *
	@echo "# vendor dir cleared"
	@-rm -rf glide.lock
	@glide cc
	@echo "# glide cache cleared"
	@glide install --strip-vendor
	$(call unify_sirupsen)
endef

# install code generators
define install_generators
	$(if $(shell command -v protoc --gogo_out=. 2> /dev/null),$(info # gogo/protobuf is installed),$(error gogo/protobuf missing, please install it with go get github.com/gogo/protobuf))
    @echo "# installing binapi-generator"
	@go get git.fd.io/govpp.git/...
endef

# Make all imports of sirupsen/logrus upper-case.
define unify_sirupsen
	@cd vendor/git.fd.io/govpp.git && find . -type f -name '*.go' -exec sed -i 's/Sirupsen\/logrus/sirupsen\/logrus/' "{}" +;
	@cd vendor/github.com/docker && find . -type f -name '*.go' -exec sed -i 's/Sirupsen\/logrus/sirupsen\/logrus/' "{}" +;
	@cd vendor/github.com/opencontainers && find . -type f -name '*.go' -exec sed -i 's/Sirupsen\/logrus/sirupsen\/logrus/' "{}" +;
	@cd vendor/github.com/prometheus && find . -type f -name '*.go' -exec sed -i 's/Sirupsen\/logrus/sirupsen\/logrus/' "{}" +;
	@cd vendor/github.com && rm -rf Sirupsen
endef
