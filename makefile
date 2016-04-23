.PHONY: build clean doc fmt install lint run test vet vendor_cleanup vendor_update vendor_updateall vendor_save vendor_saveall

default: build

build: vet
	go build -v -o ./bin/pg ./src/cmd/pg

clean:
	rm -rf ./bin/*
	rm -rf ./logs/*
	rm -rf ./sessions/*

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

# Godep has a bug where it copies the dot-files from a dependency
# Until I have time to look at it, this job will clean them up.
vendor_cleanup:
	rm -rf `find ./vendor -type f -name ".*"`

# Godep: go get github.com/tools/godep
vendor_updateall: vendor_update vendor_cleanup
vendor_saveall: vendor_save vendor_cleanup

vendor_update:
	godep update ./src/...

vendor_save:
	godep save ./src/...
