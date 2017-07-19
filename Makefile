BINDIR        ?= $(PWD)/bin

CI_COMMIT_TAG ?= unknown
CI_COMMIT_SHA ?= unknown


help:
	# all 		- runs verify, build targets
	# test 		- runs go_test target
	# build 	- runs generate, and then go_build targets
	# generate 	- generates mocks and assets files
	# verify 	- verifies generated files & scripts

.PHONY: all test verify

verify: generate go_verify

all: verify build

build: generate go_build

generate: .generate_files

go_verify: go_fmt go_vet go_test

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

.generate_exes:
	@echo "Grabbing dependencies..."
	go get -u github.com/golang/mock/mockgen
	go get -u github.com/jteeuwen/go-bindata/...
	go get -u github.com/golang/dep/cmd/dep
	@touch $@

.generate_files: .generate_exes
	dep ensure
	go generate $$(go list ./pkg/... ./cmd/...)
	@touch $@
