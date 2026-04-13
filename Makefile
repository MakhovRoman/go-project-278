.PHONY: run-debug run-prod lint test build full-flow client
run-debug:
	go run .

run-prod:
	GIN_MODE=release go run .

lint:
	golangci-lint run

test:
	go test ./...

build:
	go build

full-flow: lint test build

client:
	npx start-hexlet-url-shortener-frontend