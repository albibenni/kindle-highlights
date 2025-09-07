.PHONY: build test run clean dev compose migrate-up migrate-down lint generate

#TODO: check how to get params here
build:
	go build -o kindle-parser && ./kindle-parser

test:
	go test ./...

run:
	go run .

clean:
	rm -rf server

compose:
	docker compose down && docker compose up -d

install:
	go mod download

lint:
	golangci-lint run

docker-build:
	docker build -t myapp .

deploy:
	./scripts/deploy.sh
