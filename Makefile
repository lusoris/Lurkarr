.PHONY: build test lint generate frontend coder-push

build:
	go build ./...

test:
	go test ./...

lint:
	golangci-lint run ./...

generate:
	go generate ./...

frontend:
	cd frontend && npm ci && npm run build

coder-push:
	./deploy/coder/push-template.sh
