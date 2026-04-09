.PHONY: run-debug run-prod lint test build full-flow
run-debug:
	go run main.go

run-prod:
	GIN_MODE=release go run main.go

lint:
	golangci-lint run

test:
	go test ./...

build:
	go build

full-flow: lint test build