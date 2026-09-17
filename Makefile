.PHONY: start build test mock vet

start:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

test:
	go test ./...

mock:
	go generate ./...

vet:
	go vet ./...
