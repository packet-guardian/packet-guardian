.PHONY: build clean_full doc fmt lint run test vendor_clean vendor_get vendor_update vet

# Prepend our vendor directory to the system GOPATH
# so that import path resolution will prioritize
# our third party snapshots.
GOPATH := ${PWD}/vendor:${GOPATH}
export GOPATH

default: build

build: vet
	go build -v -o ./bin/pg ./src/cmd/pg

clean_full:
	rm -rf ./bin/*
	rm -rf ./logs/*
	rm -rf ./sessions/*

doc:
	godoc -http=:6060 -index

# http://golang.org/cmd/go/#hdr-Run_gofmt_on_package_sources
fmt:
	go fmt ./src/...

# https://github.com/golang/lint
# go get github.com/golang/lint/golint
lint:
	golint ./src

run: build
	-./bin/pg -dev -config=$(CONFIG)

test:
	go test ./src/...

vendor_clean:
	rm -dRf ./vendor/*

# We have to set GOPATH to just the vendor
# directory to ensure that `go get` doesn't
# update packages in our primary GOPATH instead.
# This will happen if you already have the package
# installed in GOPATH since `go get` will use
# that existing location as the destination.
vendor_get: vendor_clean
	GOPATH=${PWD}/vendor go get -d -u -v \
	github.com/DATA-DOG/go-sqlmock \
	github.com/dragonrider23/go-logger \
	github.com/gorilla/mux \
	github.com/gorilla/sessions \
	github.com/gorilla/context \
	github.com/mattn/go-sqlite3 \
	github.com/naoina/toml \
	golang.org/x/crypto/bcrypt
	mv ./vendor/src/* ./vendor
	rm -df ./vendor/src

vendor_update: vendor_get
	rm -rf `find ./vendor -type d -name .git` \
	&& rm -rf `find ./vendor -type d -name .hg` \
	&& rm -rf `find ./vendor -type d -name .bzr` \
	&& rm -rf `find ./vendor -type d -name .svn`
	rm -rf `find ./vendor -type f -name "*_test.go"`
	rm -rf `find ./vendor -type d -name "*_test*"`
	rm -rf `find ./vendor -type d -name "*example*"`
	rm -rf `find ./vendor -type f -name ".*"`

# http://godoc.org/code.google.com/p/go.tools/cmd/vet
# go get code.google.com/p/go.tools/cmd/vet
vet:
	go vet ./src/...
