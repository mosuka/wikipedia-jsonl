GOOS ?=
GOARCH ?=
GO111MODULE ?= on
CGO_ENABLED ?= 0
CGO_CFLAGS ?=
CGO_LDFLAGS ?=
BUILD_TAGS ?=
TAG ?=
BIN_EXT ?=
DOCKER_REPOSITORY ?= mosuka

PACKAGES = $(shell $(GO) list ./... | grep -v '/vendor/')

PROTOBUFS = $(shell find . -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq | grep -v /vendor/)
TARGET_PACKAGES = $(shell find $(CURDIR) -name 'main.go' -print0 | xargs -0 -n1 dirname | sort | uniq | grep -v /vendor/)

ifeq ($(GOOS),)
  GOOS = $(shell go version | awk -F ' ' '{print $$NF}' | awk -F '/' '{print $$1}')
endif

ifeq ($(GOARCH),)
  GOARCH = $(shell go version | awk -F ' ' '{print $$NF}' | awk -F '/' '{print $$2}')
endif

ifeq ($(TAG),)
  TAG = latest
endif
LDFLAGS = -ldflags "-s -w -X \"github.com/mosuka/wikipedia-jsonl/version.Version=$(TAG)\""

ifeq ($(GOOS),windows)
  BIN_EXT = .exe
endif

BUILD_FLAGS := GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED) CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) GO111MODULE=$(GO111MODULE)

GO := $(BUILD_FLAGS) go

.DEFAULT_GOAL := build

.PHONY: show-env
show-env:
	@echo ">> show env"
	@echo "   GOOS              = $(GOOS)"
	@echo "   GOARCH            = $(GOARCH)"
	@echo "   GO111MODULE       = $(GO111MODULE)"
	@echo "   CGO_ENABLED       = $(CGO_ENABLED)"
	@echo "   CGO_CFLAGS        = $(CGO_CFLAGS)"
	@echo "   CGO_LDFLAGS       = $(CGO_LDFLAGS)"
	@echo "   BUILD_TAGS        = $(BUILD_TAGS)"
	@echo "   TAG               = $(TAG)"
	@echo "   BIN_EXT           = $(BIN_EXT)"
	@echo "   LDFLAGS           = $(LDFLAGS)"
	@echo "   PACKAGES          = $(PACKAGES)"
	@echo "   PROTOBUFS         = $(PROTOBUFS)"
	@echo "   TARGET_PACKAGES   = $(TARGET_PACKAGES)"

.PHONY: fmt
fmt: show-env
	@echo ">> formatting code"
	$(GO) fmt $(PACKAGES)

.PHONY: test
test: show-env
	@echo ">> testing all packages"
	$(GO) clean -testcache
	$(GO) test -v -tags="$(BUILD_TAGS)" $(PACKAGES)

.PHONY: clean
clean: show-env
	@echo ">> cleaning repository"
	$(GO) clean
	rm -rf ./bin

.PHONY: build
build: show-env
	@echo ">> building binaries"
	mkdir -p bin
	$(GO) build $(LDFLAGS) -o bin/wikipedia-jsonl

.PHONY: tag
tag: show-env
	@echo ">> tagging github"
ifeq ($(TAG),$(filter $(TAG),latest master ""))
	@echo "please specify TAG"
else
	git tag -a $(TAG) -m "Release $(TAG)"
	git push origin $(TAG)
endif
