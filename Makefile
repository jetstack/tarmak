# Copyright Jetstack Ltd. See LICENSE for details.
PACKAGE_NAME ?= github.com/jetstack/tarmak

BINDIR ?= $(PWD)/bin
PATH   := $(BINDIR):$(PATH)

CI_COMMIT_TAG ?= unknown
CI_COMMIT_SHA ?= unknown

# A list of all types.go files in pkg/apis
TYPES_FILES = $(shell find pkg/apis -name types.go)

HACK_DIR     ?= hack

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
	go test $$(go list ./pkg/... ./cmd/...)

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
	go vet $$(go list ./pkg/... ./cmd/...| grep -v pkg/wing/client)

go_build:
	CGO_ENABLED=0 GOOS=linux  GOARCH=amd64 go build -a -tags netgo -ldflags '-w -X main.version=$(CI_COMMIT_TAG) -X main.commit=$(CI_COMMIT_SHA) -X main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)' -o tarmak_linux_amd64  ./cmd/tarmak
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -tags netgo -ldflags '-w -X main.version=$(CI_COMMIT_TAG) -X main.commit=$(CI_COMMIT_SHA) -X main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)' -o tarmak_darwin_amd64 ./cmd/tarmak
	CGO_ENABLED=0 GOOS=linux  GOARCH=amd64 go build -a -tags netgo -ldflags '-w -X main.version=$(CI_COMMIT_TAG) -X main.commit=$(CI_COMMIT_SHA) -X main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)' -o wing_linux_amd64    ./cmd/wing

$(BINDIR)/mockgen:
	mkdir -p $(BINDIR)
	go build -o $(BINDIR)/mockgen ./vendor/github.com/golang/mock/mockgen

$(BINDIR)/go-bindata:
	mkdir -p $(BINDIR)
	go build -o $(BINDIR)/go-bindata ./vendor/github.com/jteeuwen/go-bindata/go-bindata

depend: $(BINDIR)/go-bindata $(BINDIR)/mockgen

go_generate: depend
	go generate $$(go list ./pkg/... ./cmd/...)
	# generate all pkg/client contents
	$(HACK_DIR)/update-client-gen.sh

verify_boilerplate:
	$(HACK_DIR)/verify-boilerplate.sh

verify_client_gen:
	$(HACK_DIR)/verify-client-gen.sh

verify_vendor:
	dep ensure -no-vendor -dry-run -v
