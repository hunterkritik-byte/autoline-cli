.PHONY: test vet build lint tidy

test:
	go test ./...

vet:
	go vet ./...

build:
	go build -o bin/autoline ./cmd/autoline

lint: vet

tidy:
	go mod tidy
