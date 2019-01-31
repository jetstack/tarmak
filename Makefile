# Copyright Jetstack Ltd. See LICENSE for details.
PACKAGE_NAME ?= github.com/jetstack/tarmak
CONTAINER_DIR := /go/src/$(PACKAGE_NAME)
GO_VERSION := 1.11.5

BINDIR ?= $(CURDIR)/bin
PATH   := $(BINDIR):$(PATH)

CI_COMMIT_TAG ?= dev
CI_COMMIT_SHA ?= unknown

# A list of all types.go files in pkg/apis
TYPES_FILES = $(shell find pkg/apis -name types.go)

# List of dependent gofiles
TARMAK_GO_FILES := pkg/tarmak/binaries/binaries_bindata.go pkg/tarmak/assets/assets_bindata.go $(shell go list -f '{{ join .Deps "\n" }}' ./cmd/tarmak | xargs go list -f '{{ $$global := .}}{{ range .GoFiles }}{{ printf "%s/%s\n" $$global.Dir . }}{{ end}}')
WING_GO_FILES := $(shell go list -f '{{ join .Deps "\n" }}' ./cmd/wing | xargs go list -f '{{ $$global := .}}{{ range .GoFiles }}{{ printf "%s/%s\n" $$global.Dir . }}{{ end}}')
TAGGING_CONTROL_GO_FILES := $(shell go list -f '{{ join .Deps "\n" }}' ./cmd/tagging_control | xargs go list -f '{{ $$global := .}}{{ range .GoFiles }}{{ printf "%s/%s\n" $$global.Dir . }}{{ end}}')

HACK_DIR     ?= hack

GOPATH ?= /tmp/go

# Source URLs / hashes based on OS

UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
	TAR_EXT := tar.xz
	SHASUM := sha256sum -c
	DEP_URL := https://github.com/golang/dep/releases/download/v0.5.0/dep-linux-amd64
	DEP_HASH := 287b08291e14f1fae8ba44374b26a2b12eb941af3497ed0ca649253e21ba2f83
	NODE_NAME := node-v8.12.0-linux-x64
	NODE_URL := https://nodejs.org/dist/v8.12.0/${NODE_NAME}.$(TAR_EXT)
	NODE_HASH := 29a20479cd1e3a03396a4e74a1784ccdd1cf2f96928b56f6ffa4c8dae40c88f2
endif
ifeq ($(UNAME_S),Darwin)
	TAR_EXT := tar.gz
	SHASUM := shasum -a 256 -c
	DEP_URL := https://github.com/golang/dep/releases/download/v0.5.0/dep-darwin-amd64
	DEP_HASH := 1a7bdb0d6c31ecba8b3fd213a1170adf707657123e89dff234871af9e0498be2
	NODE_NAME := node-v8.12.0-darwin-x64
	NODE_URL := https://nodejs.org/dist/v8.12.0/${NODE_NAME}.$(TAR_EXT)
	NODE_HASH := ca131b84dfcf2b6f653a6521d31f7a108ad7d83f4d7e781945b2eca8172064aa
endif

# from https://suva.sh/posts/well-documented-makefiles/
.PHONY: help
help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

help1:
	# test      - runs go_test target
	# build     - runs generate, and then go_build targets
	# generate  - generates mocks and assets files
	# verify    - verifies generated files & scripts

.PHONY: all test verify

all: verify test build  ## runs verify, test and build targets

depend: $(BINDIR)/go-bindata $(BINDIR)/mockgen $(BINDIR)/defaulter-gen $(BINDIR)/defaulter-gen $(BINDIR)/deepcopy-gen $(BINDIR)/conversion-gen $(BINDIR)/client-gen $(BINDIR)/lister-gen $(BINDIR)/informer-gen $(BINDIR)/dep $(BINDIR)/upx $(BINDIR)/openapi-gen $(BINDIR)/gen-apidocs $(BINDIR)/node $(BINDIR)/ghr ## download all dependencies necessary for build

verify: generate go_verify verify_boilerplate verify_codegen verify_vendor verify_gen_docs ## verifies generated files & scripts

test: go_test ## runs all defined tests, no puppet tests

generate: go_generate ## generates mocks and assets files

build: generate go_build ## runs generate, and then go_build targets

go_verify: go_fmt go_vet

go_test:
	go test $$(go list ./pkg/... ./cmd/... ./puppet)

go_fmt:
	@set -e; \
	GO_FMT=$$(git ls-files *.go | grep -v 'vendor/' | xargs gofmt -d); \
	if [ -n "$${GO_FMT}" ] ; then \
		echo "Please run go fmt"; \
		echo "$$GO_FMT"; \
		exit 1; \
	fi

clean:
	rm -rf \
		$(BINDIR) \
		_output/ \
		_release/ \
		pkg/tarmak/binaries/binaries_bindata.go \
		pkg/tarmak/assets/assets_bindata.go \
		pkg/wing/mocks/http_client.go \
		pkg/wing/mocks/command.go \
		pkg/wing/mocks/client.go

go_vet:
	go vet $$(go list ./pkg/... ./cmd/...| grep -v pkg/wing/client/clientset/internalversion/fake | grep -v pkg/wing/client/clientset/versioned/fake)

go_build: _output/tarmak _output/tagging_control_linux_amd64 _output/wing_linux_amd64

$(BINDIR)/mockgen:
	mkdir -p $(BINDIR)
	go build -o $(BINDIR)/mockgen ./vendor/github.com/golang/mock/mockgen

$(BINDIR)/ghr:
	mkdir -p $(BINDIR)
	go build -o $@ ./vendor/github.com/tcnksm/ghr

$(BINDIR)/go-bindata:
	mkdir -p $(BINDIR)
	go build -o $(BINDIR)/go-bindata ./vendor/github.com/kevinburke/go-bindata/go-bindata

$(BINDIR)/defaulter-gen:
	mkdir -p $(BINDIR)
	go build -o $@ ./vendor/k8s.io/code-generator/cmd/defaulter-gen

$(BINDIR)/deepcopy-gen:
	mkdir -p $(BINDIR)
	go build -o $@ ./vendor/k8s.io/code-generator/cmd/deepcopy-gen

$(BINDIR)/conversion-gen:
	mkdir -p $(BINDIR)
	go build -o $@ ./vendor/k8s.io/code-generator/cmd/conversion-gen

$(BINDIR)/client-gen:
	mkdir -p $(BINDIR)
	go build -o $@ ./vendor/k8s.io/code-generator/cmd/client-gen

$(BINDIR)/lister-gen:
	mkdir -p $(BINDIR)
	go build -o $@ ./vendor/k8s.io/code-generator/cmd/lister-gen

$(BINDIR)/informer-gen:
	mkdir -p $(BINDIR)
	go build -o $@ ./vendor/k8s.io/code-generator/cmd/informer-gen

$(BINDIR)/dep:
	curl -sL -o $@ $(DEP_URL)
	echo "$(DEP_HASH)  $@" | $(SHASUM)
	chmod +x $@

$(BINDIR)/node:
	curl -sL -o $(BINDIR)/$(NODE_NAME).$(TAR_EXT) $(NODE_URL)
	echo "$(NODE_HASH)  $(BINDIR)/$(NODE_NAME).$(TAR_EXT)" | $(SHASUM)
	cd $(BINDIR) && tar xf $(NODE_NAME).$(TAR_EXT)
	rm $(BINDIR)/$(NODE_NAME).$(TAR_EXT)
	ln -s $(BINDIR)/$(NODE_NAME)/bin/node $(BINDIR)/node
	ln -s $(BINDIR)/$(NODE_NAME)/bin/npm $(BINDIR)/npm

$(BINDIR)/npm: $(BINDIR)/node

# upx binary packer, only supported on Linux
$(BINDIR)/upx:
ifeq ($(UNAME_S),Linux)
	curl -sL -o $@.tar.xz https://github.com/upx/upx/releases/download/v3.95/upx-3.95-amd64_linux.tar.xz
	echo "b5d6856b89dd696138ad8c7245a8f0dae4b76f41b5a31c7c43a21bc72c479c4e  $@.tar.xz" | $(SHASUM)
	which xz || ( apt-get update && apt-get -y install xz-utils)
	cd $(BINDIR) && tar xvf $(shell basename $@).tar.xz upx-3.95-amd64_linux/upx --strip-components=1
	rm $@.tar.xz
else
	echo -e "#/bin/sh\nexit 0" > $@
	chmod +x $@
endif

$(BINDIR)/openapi-gen:
	mkdir -p $(BINDIR)
	go build -o $@ ./vendor/k8s.io/kube-openapi/cmd/openapi-gen

$(BINDIR)/gen-apidocs:
	mkdir -p $(BINDIR)
	go build -o $@ ./vendor/github.com/kubernetes-incubator/reference-docs/gen-apidocs


go_generate: pkg/wing/mocks/http_client.go pkg/wing/mocks/command.go pkg/wing/mocks/client.go pkg/tarmak/binaries/binaries_bindata.go pkg/tarmak/assets/assets_bindata.go pkg/tarmak/mocks/tarmak.go pkg/tarmak/mocks/amazon.go

go_codegen: depend $(TYPES_FILES)
	$(HACK_DIR)/update-codegen.sh

go_reference_docs_gen: depend
	$(HACK_DIR)/update-reference-docs.sh

go_cmd_docs_gen: depend
	$(HACK_DIR)/update-cmd-docs.sh

verify_boilerplate:
	$(HACK_DIR)/verify-boilerplate.sh

verify_codegen: depend
	$(HACK_DIR)/verify-codegen.sh

verify_vendor: $(BINDIR)/dep
	dep ensure -no-vendor -dry-run -v

verify_gen_docs: verify_reference_docs verify_cmd_docs

verify_reference_docs: $(BINDIR)/node
	$(HACK_DIR)/verify-reference-docs.sh

verify_cmd_docs:
	$(HACK_DIR)/verify-cmd-docs.sh

SUBTREES = etcd calico aws_ebs kubernetes kubernetes_addons prometheus tarmak vault_client
subtrees:
	for module in $(SUBTREES); do \
		echo $$module; \
		git subtree pull --prefix puppet/modules/$$module git://github.com/jetstack/puppet-module-$$module.git master; \
	done

prepare_release:
ifndef VERSION
	$(error VERSION is undefined)
endif
	# replace wing version in terraform
	sed -i 's/Environment=WING_VERSION=[[:digit:]].*$$/Environment=WING_VERSION=$(VERSION)/g' terraform/amazon/templates/bastion_user_data.yaml.template terraform/amazon/templates/puppet_agent_user_data.yaml.template
	# replace major version in docs
	sed -i 's#^version = u.*$$#version = u"$(shell echo "$(VERSION)" | grep -oe "^[0-9]\{1,\}\\.[0-9]\{1,\}")"#g' docs/conf.py
	sed -i 's#^release = u.*$$#release = u"$(shell echo "$(VERSION)" | grep -oe "^[0-9]\{1,\}\\.[0-9]\{1,\}")"#g' docs/conf.py
	# replace version in README
	sed -i 's#wget https://github.com/jetstack/tarmak/releases/download/.*$$#wget https://github.com/jetstack/tarmak/releases/download/$(VERSION)/tarmak_$(VERSION)_linux_amd64#g' README.md
	sed -i 's/mv tarmak_.*$$/mv tarmak_$(VERSION)_linux_amd64 tarmak/g' README.md
	git add -p docs/conf.py terraform/amazon/templates/bastion_user_data.yaml.template terraform/amazon/templates/puppet_agent_user_data.yaml.template README.md
	git commit -m "Release $(VERSION)"
	git tag $(VERSION)

_output/wing_linux_amd64: $(WING_GO_FILES)
	mkdir -p _output/
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags netgo -o $@ -ldflags '-w $(shell hack/version-ldflags.sh)' ./cmd/wing

_output/tagging_control_linux_amd64: $(TAGGING_CONTROL_GO_FILES)
	mkdir -p _output/
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags netgo -o $@ -ldflags '-w $(shell hack/version-ldflags.sh)' ./cmd/tagging_control

_output/tarmak: $(TARMAK_GO_FILES)
	mkdir -p _output/
	CGO_ENABLED=0 go build -tags netgo -o $@ -ldflags '-w $(shell hack/version-ldflags.sh)' ./cmd/tarmak

_output/tarmak_linux_amd64: $(TARMAK_GO_FILES)
	mkdir -p _output/
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags netgo -o $@ -ldflags '-w $(shell hack/version-ldflags.sh)' ./cmd/tarmak

_output/tarmak_darwin_amd64: $(TARMAK_GO_FILES)
	mkdir -p _output/
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -tags netgo -o $@ -ldflags '-w $(shell hack/version-ldflags.sh)' ./cmd/tarmak


pkg/tarmak/mocks/tarmak.go: pkg/tarmak/interfaces/interfaces.go $(BINDIR)/mockgen
	mockgen -package=mocks -source=pkg/tarmak/interfaces/interfaces.go -destination $@

pkg/tarmak/mocks/amazon.go: pkg/tarmak/provider/amazon/amazon.go $(BINDIR)/mockgen
	mockgen -package=mocks -source=pkg/tarmak/provider/amazon/amazon.go -destination $@

pkg/tarmak/binaries/binaries_bindata.go: _output/wing_linux_amd64 _output/tagging_control_linux_amd64 pkg/tarmak/binaries/binaries.go $(BINDIR)/go-bindata
	go generate ./pkg/tarmak/binaries

pkg/tarmak/assets/assets_bindata.go: $(shell find packer/ puppet/ terraform/ -type f) pkg/tarmak/assets/assets.go $(BINDIR)/go-bindata
	go generate ./pkg/tarmak/assets

pkg/wing/mocks/http_client.go: vendor/k8s.io/client-go/rest/request.go $(BINDIR)/mockgen
	mockgen -package=mocks -source=vendor/k8s.io/client-go/rest/request.go -destination $@

pkg/wing/mocks/command.go: pkg/wing/command.go $(BINDIR)/mockgen
	mockgen -package=mocks -source=pkg/wing/command.go -destination $@

pkg/wing/mocks/client.go: $(shell go list -f '{{ $$global := .}}{{ range .GoFiles }}{{ printf "%s/%s\n" $$global.Dir . }}{{ end}}' k8s.io/client-go/rest) $(BINDIR)/mockgen
	mockgen -destination $@ -package=mocks k8s.io/client-go/rest Interface


## Release instructions
.PHONY: release
release: _output/tarmak_linux_amd64 _output/tarmak_darwin_amd64 _output/wing_linux_amd64 _output/tagging_control_linux_amd64 $(BINDIR)/upx $(BINDIR)/ghr ## Build release binaries and publish them to github
	hack/publish-release.sh

docker_%:
	# create a container
	$(eval CONTAINER_ID := $(shell docker create \
		-i \
		-w $(CONTAINER_DIR) \
		golang:${GO_VERSION} \
		/bin/bash -c "make $*" \
	))

	# copy stuff into container
	(git ls-files && git ls-files --others --exclude-standard) | tar cf -  -T - | docker cp - $(CONTAINER_ID):$(CONTAINER_DIR)

	# run build inside container
	docker start -a -i $(CONTAINER_ID)

	# copy artifacts over
	docker cp $(CONTAINER_ID):$(CONTAINER_DIR)/wing_linux_amd64 wing_linux_amd64
	docker cp $(CONTAINER_ID):$(CONTAINER_DIR)/tarmak_linux_amd64 tarmak_linux_amd64
	docker cp $(CONTAINER_ID):$(CONTAINER_DIR)/tarmak_darwin_amd64 tarmak_darwin_amd64

	# remove container
	docker rm $(CONTAINER_ID)
