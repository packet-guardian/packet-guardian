NAME := packet-guardian
DESC := A captive portal for today's networks
VERSION := $(shell git describe --tags --always --dirty)
GITCOMMIT := $(shell git rev-parse HEAD)
GOVERSION := $(shell go version)
BUILDTIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILDER := $(shell echo "`git config user.name` <`git config user.email`>")
DIST_FILENAME ?= pg-dist-$(VERSION).tar.gz
CGO_ENABLED ?= 1
PWD := $(shell pwd)
GOBIN := $(PWD)/bin
CODECLIMATE_CODE := $(PWD)

ifeq ($(shell uname -o), Cygwin)
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

.PHONY: all doc fmt alltests test coverage benchmark lint vet management dist clean docker docker-compile docker-build codeclimate bindata

all: bindata test management

bindata:
	go-bindata -o src/bindata/bindata.go -pkg bindata templates/... public/...

management:
	go build -o bin/pg -v -ldflags "$(LDFLAGS)" -tags '$(BUILDTAGS)' ./cmd/pg

# development tasks
doc:
	@godoc -http=:6060 -index

fmt:
	@gofmt -s -l -d ./src/*

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

codeclimate:
	@docker run -i --rm \
		--env CODECLIMATE_CODE="$(CODECLIMATE_CODE)" \
		-v $(CODECLIMATE_CODE):/code \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v /tmp/cc:/tmp/cc \
		codeclimate/codeclimate analyze $(CODECLIMATE_ARGS)

dist: vet all
	@rm -rf ./dist
	@mkdir -p dist/packet-guardian
	@cp -R config dist/packet-guardian/

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

docker:
	docker build \
		--build-arg version='$(VERSION)' \
		--build-arg builddate='$(BUILDTIME)' \
		--build-arg vcsref='$(GITCOMMIT)' \
		-t packet-guardian .
