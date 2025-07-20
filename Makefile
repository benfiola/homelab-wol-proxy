PWD := $(shell pwd)
DST ?= $(PWD)/build
BIN ?= $(PWD)/.bin

OS := $(shell go env GOOS)
ARCH := $(shell go env GOARCH)

GRARCH := $(ARCH)
ifeq ($(ARCH),amd64)
	GRARCH := x86_64
endif
GORELEASER_VERSION := 2.11.0
GORELEASER_URL := https://github.com/goreleaser/goreleaser/releases/download/v$(GORELEASER_VERSION)/goreleaser_$(OS)_$(GRARCH).tar.gz
SEMANTIC_RELEASE_VERSION := 2.31.0
SEMANTIC_RELEASE_URL := https://github.com/go-semantic-release/semantic-release/releases/download/v$(SEMANTIC_RELEASE_VERSION)/semantic-release_v$(SEMANTIC_RELEASE_VERSION)_$(OS)_$(ARCH)

.PHONY: default
default: build

.PHONY: build
build: | $(DST)
	# build the program
	goreleaser build --auto-snapshot --clean --single-target

.PHONY: install-semantic-release
install-semantic-release: $(BIN)/semantic-release
$(BIN)/semantic-release: | $(BIN)
	# download file
	curl -o $(BIN)/semantic-release -fsSL $(SEMANTIC_RELEASE_URL)
	# make executable
	chmod +x $(BIN)/semantic-release

.PHONY: install-goreleaser
install-goreleaser: $(BIN)/goreleaser 
$(BIN)/goreleaser: | $(BIN)
	# clean temp paths
	rm -rf $(BIN)/.archive.tar.gz $(BIN)/.extract && mkdir -p $(BIN)/.extract
	# download archive
	curl -o $(BIN)/.archive.tar.gz -fsSL $(GORELEASER_URL)
	# extract archive
	tar xvzf $(BIN)/.archive.tar.gz -C $(BIN)/.extract
	# copy file
	mv $(BIN)/.extract/goreleaser $(BIN)/goreleaser
	# clean temp paths
	rm -rf $(BIN)/.archive.tar.gz $(BIN)/.extract

$(BIN):
	# create bin directory
	mkdir -p $(BIN)

$(DST):
	# create dst directory
	mkdir -p $(DST)