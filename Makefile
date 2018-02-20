# Copyright Jetstack Ltd. See LICENSE for details.
PACKAGE_NAME ?= github.com/jetstack/tarmak

BINDIR ?= $(CURDIR)/bin
PATH   := $(BINDIR):$(PATH)

CI_COMMIT_TAG ?= unknown
CI_COMMIT_SHA ?= unknown

# A list of all types.go files in pkg/apis
TYPES_FILES = $(shell find pkg/apis -name types.go)

HACK_DIR     ?= hack

GOPATH ?= /tmp/go

# Source URLs / hashes based on OS

UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
	DEP_URL := https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64
	DEP_HASH := 31144e465e52ffbc0035248a10ddea61a09bf28b00784fd3fdd9882c8cbb2315
	GORELEASER_URL := https://github.com/goreleaser/goreleaser/releases/download/v0.54.0/goreleaser_Linux_x86_64.tar.gz
	GORELEASER_HASH := 895df4293580dd8f9b0daf0ef5456f2238a2fbfc51d9f75dde6e2c63ca4fccc2
endif
ifeq ($(UNAME_S),Darwin)
	DEP_URL := https://github.com/golang/dep/releases/download/v0.4.1/dep-darwin-amd64
	DEP_HASH := f170008e2bf8b196779c361a4eaece1b03450d23bbf32d1a0beaa9b00b6a5ab4
	GORELEASER_URL := https://github.com/goreleaser/goreleaser/releases/download/v0.54.0/goreleaser_Darwin_x86_64.tar.gz
	GORELEASER_HASH := 9d927528a599174eed4d0d6a1ce6bdc810463c4cb105b0d2319c7c63ec642c9b
endif


help:
	# all       - runs verify, build targets
	# test      - runs go_test target
	# build     - runs generate, and then go_build targets
	# generate  - generates mocks and assets files
	# verify    - verifies generated files & scripts

.PHONY: all test verify

test: go_test

verify: generate go_verify verify_boilerplate verify_client_gen verify_vendor

all: verify test build

build: generate go_build

generate: go_generate

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
	rm -rf $(BINDIR)

go_vet:
	go vet $$(go list ./pkg/... ./cmd/...| grep -v pkg/wing/client/fake | grep -v pkg/wing/clients/internalclientset/fake)

go_build:
	# Make sure you add all binaries to the .goreleaser.yml as well
	CGO_ENABLED=0 GOOS=linux  GOARCH=amd64 go build -a -tags netgo -ldflags '-w -X main.version=$(CI_COMMIT_TAG) -X main.commit=$(CI_COMMIT_SHA) -X main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)' -o tarmak_linux_amd64  ./cmd/tarmak
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -tags netgo -ldflags '-w -X main.version=$(CI_COMMIT_TAG) -X main.commit=$(CI_COMMIT_SHA) -X main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)' -o tarmak_darwin_amd64 ./cmd/tarmak
	CGO_ENABLED=0 GOOS=linux  GOARCH=amd64 go build -a -tags netgo -ldflags '-w -X main.version=$(CI_COMMIT_TAG) -X main.commit=$(CI_COMMIT_SHA) -X main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)' -o wing_linux_amd64    ./cmd/wing
	CGO_ENABLED=0 GOOS=linux  GOARCH=amd64 go build -a -tags netgo -ldflags '-w -X main.version=$(CI_COMMIT_TAG) -X main.commit=$(CI_COMMIT_SHA) -X main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)' -o terraform-provider-awstag_linux_amd64	./cmd/terraform-provider-awstag

$(BINDIR)/mockgen:
	mkdir -p $(BINDIR)
	go build -o $(BINDIR)/mockgen ./vendor/github.com/golang/mock/mockgen

$(BINDIR)/go-bindata:
	mkdir -p $(BINDIR)
	go build -o $(BINDIR)/go-bindata ./vendor/github.com/jteeuwen/go-bindata/go-bindata

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
	echo "$(DEP_HASH)  $@" | sha256sum -c
	chmod +x $@

$(BINDIR)/goreleaser:
	curl -sL -o $@.tar.gz $(GORELEASER_URL)
	echo "$(GORELEASER_HASH) $@.tar.gz" | sha256sum -c
	cd $(BINDIR) && tar xzvf $(shell basename $@).tar.gz goreleaser

depend: $(BINDIR)/go-bindata $(BINDIR)/mockgen $(BINDIR)/defaulter-gen $(BINDIR)/defaulter-gen $(BINDIR)/deepcopy-gen $(BINDIR)/conversion-gen $(BINDIR)/client-gen $(BINDIR)/lister-gen $(BINDIR)/informer-gen $(BINDIR)/dep $(BINDIR)/goreleaser

go_generate: depend
	go generate $$(go list ./pkg/... ./cmd/...)

go_generate_types: depend $(TYPES_FILES)
	# generate types
	defaulter-gen \
		--v 1 --logtostderr \
		--go-header-file "$(HACK_DIR)/boilerplate/boilerplate.go.txt" \
		--input-dirs "$(PACKAGE_NAME)/pkg/apis/tarmak/v1alpha1" \
		--input-dirs "$(PACKAGE_NAME)/pkg/apis/cluster/v1alpha1" \
		--input-dirs "$(PACKAGE_NAME)/pkg/apis/wing" \
		--input-dirs "$(PACKAGE_NAME)/pkg/apis/wing/v1alpha1" \
		--extra-peer-dirs "$(PACKAGE_NAME)/pkg/apis/cluster/v1alpha1" \
		--extra-peer-dirs "$(PACKAGE_NAME)/pkg/apis/tarmak/v1alpha1" \
		--extra-peer-dirs "$(PACKAGE_NAME)/pkg/apis/wing" \
		--extra-peer-dirs "$(PACKAGE_NAME)/pkg/apis/wing/v1alpha1" \
		--output-file-base "zz_generated.defaults"
	# generate deep copies
	deepcopy-gen \
		--v 1 --logtostderr \
		--go-header-file "$(HACK_DIR)/boilerplate/boilerplate.go.txt" \
		--input-dirs "$(PACKAGE_NAME)/pkg/apis/tarmak/v1alpha1" \
		--input-dirs "$(PACKAGE_NAME)/pkg/apis/cluster/v1alpha1" \
		--input-dirs "$(PACKAGE_NAME)/pkg/apis/wing" \
		--input-dirs "$(PACKAGE_NAME)/pkg/apis/wing/v1alpha1" \
		--output-file-base zz_generated.deepcopy
	# generate conversions
	conversion-gen \
		--v 1 --logtostderr \
		--go-header-file "$(HACK_DIR)/boilerplate/boilerplate.go.txt" \
		--input-dirs "$(PACKAGE_NAME)/pkg/apis/wing" \
		--input-dirs "$(PACKAGE_NAME)/pkg/apis/wing/v1alpha1" \
		--output-file-base zz_generated.conversion
	# generate all pkg/client contents
	$(HACK_DIR)/update-client-gen.sh

verify_boilerplate:
	$(HACK_DIR)/verify-boilerplate.sh

verify_client_gen:
	$(HACK_DIR)/verify-client-gen.sh

verify_vendor: $(BINDIR)/dep
	dep ensure -no-vendor -dry-run -v

SUBTREES = etcd calico aws_ebs kubernetes kubernetes_addons prometheus tarmak vault_client
subtrees:
	for module in $(SUBTREES); do \
		echo $$module; \
		git subtree pull --prefix puppet/modules/$$module git://github.com/jetstack/puppet-module-$$module.git master; \
	done

release:
ifndef VERSION
	$(error VERSION is undefined)
endif
	# replace wing version in terraform
	sed -i 's/Environment=WING_VERSION=.*$$/Environment=WING_VERSION=$(VERSION)/g' terraform/amazon/tools/templates/bastion_user_data.yaml terraform/amazon/kubernetes/templates/puppet_agent_user_data.yaml
	# replace major version in docs
	sed -i 's#^version = u.*$$#version = u"$(shell echo "$(VERSION)" | grep -oe "^[0-9]\{1,\}\\.[0-9]\{1,\}")"#g' docs/conf.py
	sed -i 's#^release = u.*$$#release = u"$(shell echo "$(VERSION)" | grep -oe "^[0-9]\{1,\}\\.[0-9]\{1,\}")"#g' docs/conf.py
	# replace version in README
	sed -i 's#wget https://github.com/jetstack/tarmak/releases/download/.*$$#wget https://github.com/jetstack/tarmak/releases/download/$(VERSION)/tarmak_$(VERSION)_linux_amd64#g' README.md
	sed -i 's/mv tarmak_.*$$/mv tarmak_$(VERSION)_linux_amd64 tarmak/g' README.md
	# replace version in Dockerfile
	sed -i 's#^ENV TERRAFORM_PROVIDER_AWSTAG_VERSION .*$$#ENV TERRAFORM_PROVIDER_AWSTAG_VERSION $(VERSION)#g' terraform/Dockerfile
