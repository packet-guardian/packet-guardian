NAME := packet-guardian
DESC := A captive portal for today's networks
VERSION := $(shell git describe --tags --always --dirty)
GITCOMMIT := $(shell git rev-parse HEAD)
GOVERSION := $(shell go version)
BUILDTIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILDER := $(shell echo "`git config user.name` <`git config user.email`>")
DIST_FILENAME ?= pg-dist-$(VERSION).tar.gz
export CGO_ENABLED ?= 0
PWD := $(shell pwd)
GOBIN := $(PWD)/bin
CODECLIMATE_CODE := $(PWD)

ifeq ($(shell uname -s), Cygwin)
CODECLIMATE_CODE := //c/cygwin64$(PWD)
PWD := $(shell cygpath -w -a `pwd`)
GOBIN := $(PWD)\bin
endif

DOCKER_DIR := /go/src/github.com/packet-guardian/packet-guardian

PROJECT_URL := "https://github.com/packet-guardian/$(NAME)"
BUILDTAGS ?= dball
LDFLAGS := -X 'main.version=$(VERSION)' \
			-X 'main.buildTime=$(BUILDTIME)' \
			-X 'main.builder=$(BUILDER)' \
			-X 'main.goversion=$(GOVERSION)'

.PHONY: all dev fmt alltests test benchmark lint build dist clean docker codeclimate bindata yarn yarn-dev

all: yarn bindata test build
dev: yarn-dev bindata test build

yarn:
	yarn run build:prod

yarn-dev:
	yarn run build:dev

# go get github.com/go-bindata/go-bindata/...
bindata:
	rm public/dist/js/*.map
	go-bindata -o src/bindata/bindata.go -pkg bindata templates/... public/dist/...

build:
	go build -o bin/pg -v -ldflags "$(LDFLAGS)" -tags '$(BUILDTAGS)' ./cmd/pg

# development tasks
fmt:
	@gofmt -s -l -d ./src/*

alltests: test lint

test:
	@go test ./src/...

benchmark:
	@go test -bench=. $$(go list ./src/...)

# https://github.com/golang/lint
# go get github.com/golang/lint/golint
lint:
	@golint ./src/...

codeclimate:
	@docker run -i --rm \
		--env CODECLIMATE_CODE="$(CODECLIMATE_CODE)" \
		-v $(CODECLIMATE_CODE):/code \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v /tmp/cc:/tmp/cc \
		codeclimate/codeclimate analyze $(CODECLIMATE_ARGS)

dist: all
	@rm -rf ./dist
	@mkdir -p dist/packet-guardian
	@cp -R config dist/packet-guardian/

	@cp LICENSE dist/packet-guardian/
	@cp README.md dist/packet-guardian/
	@cp -R scripts dist/packet-guardian/
	@rm -rf dist/packet-guardian/scripts/dev-docker

	@mkdir dist/packet-guardian/bin
	@cp bin/pg dist/packet-guardian/bin/pg

	@mkdir dist/packet-guardian/sessions

	(cd "dist"; tar -cz packet-guardian) > "dist/$(DIST_FILENAME)"

	@rm -rf dist/packet-guardian

clean:
	rm -rf ./bin/*
	rm -rf ./logs/*
	rm -rf ./sessions/*

docker:
	docker build \
		--build-arg version='$(VERSION)' \
		--build-arg builddate='$(BUILDTIME)' \
		--build-arg vcsref='$(GITCOMMIT)' \
		-t packet-guardian .
