.PHONY: start build test mock vet docker-build docker-up docker-down docker-logs

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

docker-build:
	docker compose build

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f
