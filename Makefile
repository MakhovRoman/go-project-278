.PHONY: run-debug run-prod
run-debug:
	go run main.go

run-prod:
	GIN_MODE=release go run main.go