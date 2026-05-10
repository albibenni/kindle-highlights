.PHONY: build test run clean dev compose migrate-up migrate-down lint generate deps install setup-config coverage install-hooks setup

setup:
	@./scripts/install.sh

install-hooks:
	@echo "Installing pre-commit hooks..."
	@printf "#!/bin/sh\n\n# Run linting\nmake lint\nif [ \$$? -ne 0 ]; then\n    echo 'Linting failed. Commit aborted.'\n    exit 1\nfi\n\n# Run tests\nmake test\nif [ \$$? -ne 0 ]; then\n    echo 'Tests failed. Commit aborted.'\n    exit 1\nfi\n\necho 'All checks passed. Proceeding with commit.'\n" > .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "Hooks installed successfully."

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@rm coverage.out

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
	@command -v rg >/dev/null 2>&1 || { echo >&2 "Warning: ripgrep (rg) is not installed. System search feature will not work."; }
	go install .
	@echo "Installed kindle-parser to $$(go env GOPATH)/bin"

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