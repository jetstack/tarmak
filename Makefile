BINDIR ?= $(PWD)/bin
PATH   := $(BINDIR):$(PATH)

CI_COMMIT_TAG ?= unknown
CI_COMMIT_SHA ?= unknown

help:
	# all       - runs verify, build targets
	# test      - runs go_test target
	# build     - runs generate, and then go_build targets
	# generate  - generates mocks and assets files
	# verify    - verifies generated files & scripts

.PHONY: all test verify

test: go_test

verify: generate go_verify

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

go_vet:
	go vet $$(go list ./pkg/... ./cmd/...)

go_build:
	CGO_ENABLED=0 GOOS=linux  GOARCH=amd64 go build -a -tags netgo -ldflags '-w -X main.version=$(CI_COMMIT_TAG) -X main.commit=$(CI_COMMIT_SHA) -X main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)' -o tarmak_linux_amd64  ./cmd/tarmak
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -tags netgo -ldflags '-w -X main.version=$(CI_COMMIT_TAG) -X main.commit=$(CI_COMMIT_SHA) -X main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)' -o tarmak_darwin_amd64 ./cmd/tarmak
	CGO_ENABLED=0 GOOS=linux  GOARCH=amd64 go build -a -tags netgo -ldflags '-w -X main.version=$(CI_COMMIT_TAG) -X main.commit=$(CI_COMMIT_SHA) -X main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)' -o wing_linux_amd64    ./cmd/wing

go_codegen:
	mockgen -imports .=github.com/jetstack/tarmak/pkg/tarmak/interfaces -package=mocks -source=pkg/tarmak/interfaces/interfaces.go > pkg/tarmak/mocks/tarmak.go
	mockgen -package=mocks -source=pkg/tarmak/provider/aws/aws.go > pkg/tarmak/mocks/aws.go

.exes_prepare:
	mkdir -p $(BINDIR)

bin/mockgen:
	mkdir -p $(BINDIR)
	go build -o $(BINDIR)/mockgen ./vendor/github.com/golang/mock/mockgen

bin/go-bindata:
	mkdir -p $(BINDIR)
	go build -o $(BINDIR)/go-bindata ./vendor/github.com/jteeuwen/go-bindata/go-bindata

depend: bin/go-bindata bin/mockgen

go_generate: depend
	go generate $$(go list ./pkg/... ./cmd/...)
