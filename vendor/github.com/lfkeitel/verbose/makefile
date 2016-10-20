.PHONE: doc lint test vet full generate

default: lint vet test

doc:
	godoc -http=:6060 -index

lint:
	golint ./...

test:
	go test -v ./...

vet:
	go vet ./...

generate:
	go generate

full: generate lint vet test
