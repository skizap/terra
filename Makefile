GOOS?=linux
GOARCH?=amd64
COMMIT=`git rev-parse --short HEAD`
NAMESPACE?=stellarproject
IMAGE_NAMESPACE?=$(NAMESPACE)
APP=terra
CLI=tctl
REPO?=$(NAMESPACE)/$(APP)
TAG?=dev
BUILD?=-dev
BUILD_ARGS?=
PACKAGES=$(shell go list ./... | grep -v -e /vendor/)
CWD=$(PWD)

all: binaries

generate:
	@echo ${PACKAGES} | xargs protobuild -quiet

binaries: app cli
	@echo " -> Built $(TAG) version ${COMMIT} (${GOOS}/${GOARCH})"

bindir:
	@mkdir -p bin

app: bindir
	@cd cmd/$(APP) && CGO_ENABLED=0 go build -installsuffix cgo -ldflags "-w -X github.com/$(REPO)/version.GitCommit=$(COMMIT) -X github.com/$(REPO)/version.Build=$(BUILD)" -o ../../bin/$(APP) .

cli: bindir
	@cd cmd/$(CLI) && CGO_ENABLED=0 go build -installsuffix cgo -ldflags "-w -X github.com/$(REPO)/version.GitCommit=$(COMMIT) -X github.com/$(REPO)/version.Build=$(BUILD)" -o ../../bin/$(CLI) .

vet:
	@echo " -> $@"
	@test -z "$$(go vet ${PACKAGES} 2>&1 | tee /dev/stderr)"

lint:
	@echo " -> $@"
	@golint -set_exit_status ${PACKAGES}

check: vet lint

test:
	@go test -short -v -cover $(TEST_ARGS) ${PACKAGES}

install:
	@install -D -m 755 cmd/$(APP)/$(APP) /usr/local/bin/

clean:
	@rm -rf bin/
	@rm -rf ./*.deb

.PHONY: generate clean check test install app binaries
