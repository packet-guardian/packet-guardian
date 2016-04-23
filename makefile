.PHONY: build clean clean_vendor doc fmt install lint run test vet

# Prepend our vendor directory to the system GOPATH
# so that import path resolution will prioritize
# our third party snapshots.
GOPATH := ${PWD}/vendor:${GOPATH}
export GOPATH

default: build

build: vet
	go build -v -o ./bin/pg ./src/cmd/pg

clean:
	rm -rf ./bin/*
	rm -rf ./logs/*
	rm -rf ./sessions/*

# Godep has a bug where it copies the dot-files from a dependency
# Until I have time to look at it, this job will clean them up.
clean_vendor:
	rm -rf `find ./vendor -type f -name ".*"`

doc:
	godoc -http=:6060 -index

fmt:
	go fmt ./src/...

install: vet
	go install -v ./src/cmd/pg

# https://github.com/golang/lint
# go get github.com/golang/lint/golint
lint:
	golint ./src

run: build
	-./bin/pg -dev -config=$(CONFIG)

test:
	go test ./src/...

vet:
	go vet ./src/...
