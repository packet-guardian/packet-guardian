.PHONY: all doc fmt alltests test coverage benchmark lint vet dhcp management dist clean docker

all: test

# development tasks
doc:
	@godoc -http=:6060 -index

fmt:
	@go fmt

alltests: test lint vet

test:
ifdef verbose
	@go test -race -v
else
	@go test -race
endif

coverage:
	@go test -cover

benchmark:
	@echo "Running tests..."
	@go test -bench=.

# https://github.com/golang/lint
# go get github.com/golang/lint/golint
lint:
	@golint ./src/...

vet:
	@go vet
