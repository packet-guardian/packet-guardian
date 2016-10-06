export GO15VENDOREXPERIMENT=1

# variable definitions
NAME := packet-guardian
DESC := A captive portal for today's networks
VERSION := $(shell git describe --tags --always --dirty)
GOVERSION := $(shell go version)
BUILDTIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILDDATE := $(shell date -u +"%B %d, %Y")
BUILDER := $(shell echo "`git config user.name` <`git config user.email`>")
PKG_RELEASE ?= 1
PROJECT_URL := "https://github.com/usi-lfkeitel/$(NAME)"
BUILDTAGS ?= dball
LDFLAGS := -X 'main.version=$(VERSION)' \
			-X 'main.buildTime=$(BUILDTIME)' \
			-X 'main.builder=$(BUILDER)' \
			-X 'main.goversion=$(GOVERSION)'

.PHONY: all doc fmt alltests test coverage benchmark lint vet dhcp management dist clean docker

all: test management dhcp

# development tasks
doc:
	@godoc -http=:6060 -index

fmt:
	@go fmt $$(go list ./src/...)

alltests: test lint vet

test:
	@go test -race $$(go list ./src/...)

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

dhcp:
	GOBIN=$(PWD)/bin go install -v -ldflags "$(LDFLAGS)" -tags '$(BUILDTAGS)' ./cmd/dhcp

management:
	GOBIN=$(PWD)/bin go install -v -ldflags "$(LDFLAGS)" -tags '$(BUILDTAGS)' ./cmd/pg

dist: vet local-install
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

	(cd "dist"; tar -cz packet-guardian) > "dist/pg-dist-$(VERSION).tar.gz"

	@rm -rf dist/packet-guardian

clean:
	rm $(TARGETS)
	rm -rf ./logs/*
	rm -rf ./sessions/*

docker:
	@rm -rf docker/tmp
	@mkdir docker/tmp
	cp dist/pg-dist* docker/tmp/dist.tar.gz

	cp docker/pg-base/Dockerfile docker/tmp/Dockerfile
	cd docker/tmp; \
	sudo docker build -t pg-base --rm .

	cp docker/pg-web/Dockerfile docker/tmp/Dockerfile
	cd docker/tmp; \
	sudo docker build -t pg-web --rm .

	cp docker/pg-dhcp/Dockerfile docker/tmp/Dockerfile
	cd docker/tmp; \
	sudo docker build -t pg-dhcp --rm .

	@rm -rf docker/tmp
