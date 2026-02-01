.PHONY: build test run clean dev compose migrate-up migrate-down lint generate deps install setup-config

build:
	go build -o kindle-parser && ./kindle-parser $(ARGS) # example make build ARGS="test Linux"

build-solo:
	go build -o kindle-parser

test:
	go test ./...

run:
	go run .

clean:
	rm -rf server

compose:
	docker compose down && docker compose up -d

deps:
	go mod download

lint:
	golangci-lint run

docker-build:
	docker build -t myapp .

deploy:
	./scripts/deploy.sh

install:
	go build -o $(HOME)/go/bin/kindle-parser .

setup-config:
	mkdir -p $(HOME)/.config/kindle-highlights
	@if [ -f mac.env ]; then \
		echo "Copying mac.env to global config..."; \
		cp mac.env $(HOME)/.config/kindle-highlights/.env; \
	elif [ -f linux.env ]; then \
		echo "Copying linux.env to global config..."; \
		cp linux.env $(HOME)/.config/kindle-highlights/.env; \
	else \
		echo "Error: No .env template found (mac.env or linux.env missing)"; \
		exit 1; \
	fi