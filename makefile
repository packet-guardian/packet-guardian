NAME := packet-guardian
DESC := A captive portal for today's networks
VERSION := $(shell git describe --tags --always --dirty)
GOVERSION := $(shell go version)
BUILDTIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILDDATE := $(shell date -u +"%B %d, %Y")
BUILDER := $(shell echo "`git config user.name` <`git config user.email`>")
DIST_FILENAME ?= pg-dist-$(VERSION).tar.gz
CGO_ENABLED ?= 1
PWD := $(shell pwd)
GOBIN := $(PWD)/bin

ifeq ($(shell uname -o), Cygwin)
PWD := $(shell cygpath -w -a `pwd`)
GOBIN := $(PWD)\bin
endif

DOCKER_DIR := /go/src/github.com/usi-lfkeitel/packet-guardian

PROJECT_URL := "https://github.com/usi-lfkeitel/$(NAME)"
BUILDTAGS ?= dball
LDFLAGS := -X 'main.version=$(VERSION)' \
			-X 'main.buildTime=$(BUILDTIME)' \
			-X 'main.builder=$(BUILDER)' \
			-X 'main.goversion=$(GOVERSION)'

.PHONY: all doc fmt alltests test coverage benchmark lint vet dhcp management dist clean docker docker-compile docker-build

all: test management dhcp

dhcp:
	GOBIN="$(GOBIN)" go install -v -ldflags "$(LDFLAGS)" -tags '$(BUILDTAGS)' ./cmd/dhcp

management:
	GOBIN="$(GOBIN)" go install -v -ldflags "$(LDFLAGS)" -tags '$(BUILDTAGS)' ./cmd/pg

# development tasks
doc:
	@godoc -http=:6060 -index

fmt:
	@go fmt $$(go list ./src/...)

alltests: test lint vet

test:
ifeq (CGO_ENABLED, 1)
	@go test -race $$(go list ./src/...)
else
	@go test $$(go list ./src/...)
endif

coverage:
	@go test -cover $$(go list ./src/...)

benchmark:
	@echo "Running tests..."
	@go test -bench=. $$(go list ./src/...)

# https://github.com/golang/lint
# go get github.com/golang/lint/golint
lint:
	@golint ./src/...

vet:
	@go vet $$(go list ./src/...)

dist: vet all
	@rm -rf ./dist
	@mkdir -p dist/packet-guardian
	@cp -R config dist/packet-guardian/
	@cp -R public dist/packet-guardian/
	@cp -R templates dist/packet-guardian/

	@cp LICENSE dist/packet-guardian/
	@cp README.md dist/packet-guardian/
	@cp -R scripts dist/packet-guardian/

	@mkdir dist/packet-guardian/bin
	@cp bin/pg dist/packet-guardian/bin/pg
	@cp bin/dhcp dist/packet-guardian/bin/dhcp

	@mkdir dist/packet-guardian/sessions

	(cd "dist"; tar -cz packet-guardian) > "dist/$(DIST_FILENAME)"

	@rm -rf dist/packet-guardian

clean:
	rm -rf ./bin/*
	rm -rf ./logs/*
	rm -rf ./sessions/*

docker: docker-compile docker-build

docker-compile:
	docker run --rm -v $(PWD):$(DOCKER_DIR) -w $(DOCKER_DIR) -e CGO_ENABLED=0 -e BUILDTAGS=dbmysql -e DIST_FILENAME=dist.tar.gz golang:1.8 make dist

docker-build:
	docker build -t packet-guardian .
