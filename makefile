.PHONY: build clean doc fmt install lint test vet vendor_cleanup vendor_update vendor_updateall vendor_save vendor_saveall

VERSION?=unversioned

default: build

build: test
	go build -v -o ./bin/pg ./src/cmd/pg
	go build -v -o ./bin/dhcp ./src/cmd/dhcp

clean:
	rm -rf ./bin/*
	rm -rf ./logs/*
	rm -rf ./sessions/*

dist: build
	@rm -rf ./dist
	@mkdir -p dist/packet-guardian
	@cp -R config dist/packet-guardian/
	@cp -R public dist/packet-guardian/
	@cp -R templates dist/packet-guardian/

	@cp LICENSE dist/packet-guardian/
	@cp README.md dist/packet-guardian/
	@cp Dockerfile dist/packet-guardian/
	@cp install.sh dist/packet-guardian/
	@cp uninstall.sh dist/packet-guardian/

	@mkdir dist/packet-guardian/bin
	@cp bin/pg dist/packet-guardian/bin/pg
	@cp bin/dhcp dist/packet-guardian/bin/dhcp

	@mkdir dist/packet-guardian/sessions

	(cd "dist"; tar -cz packet-guardian) > "dist/pg-dist-$(VERSION).tar.gz"

	@rm -rf dist/packet-guardian

doc:
	godoc -http=:6060 -index

fmt:
	go fmt ./src/...

install: test
	GOBIN=$(PWD)/bin go install -v ./src/cmd/pg
	GOBIN=$(PWD)/bin go install -v ./src/cmd/dhcp

# https://github.com/golang/lint
# go get github.com/golang/lint/golint
lint:
	golint ./src/...

test: vet
ifdef verbose
		go test -v ./src/...
else ifdef vverbose
		PG_TEST_LOG=true go test -v ./src/...
else
		go test ./src/...
endif

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
