.PHONY: build test test-integration test-all lint generate frontend coder-push security cover

build:
	go build ./...

test:
	go test ./...

test-integration:
	go test -tags integration -count=1 -timeout 300s ./internal/database/

test-all: test test-integration

lint:
	golangci-lint run ./...

generate:
	go generate ./...

frontend:
	cd frontend && npm ci && npm run build

coder-push:
	./deploy/coder/push-template.sh

security:
	govulncheck ./...
	gosec ./...

cover:
	go test -coverprofile=coverage/coverage.out ./...
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@echo "Coverage report: coverage/coverage.html"
