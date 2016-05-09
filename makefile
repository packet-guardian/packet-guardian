.PHONY: build clean doc fmt install lint run test vet vendor_cleanup vendor_update vendor_updateall vendor_save vendor_saveall

VERSION?=unversioned

default: build

build: vet test
	go build -v -o ./bin/pg ./src/cmd/pg

clean:
	rm -rf ./bin/*
	rm -rf ./logs/*
	rm -rf ./sessions/*

dist: build
	rm -rf ./dist
	mkdir -p dist/packet-guardian
	cp -R config dist/packet-guardian/
	cp -R public dist/packet-guardian/
	cp -R templates dist/packet-guardian/
	cp LICENSE dist/packet-guardian/
	cp README.md dist/packet-guardian/
	mkdir dist/packet-guardian/bin
	cp bin/pg dist/packet-guardian/bin/pg

	(cd "dist"; tar -cz packet-guardian) > "dist/pg-dist-$(VERSION).tar.gz"

	rm -rf dist/packet-guardian

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
	-./bin/pg -d -c=$(CONFIG)

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
