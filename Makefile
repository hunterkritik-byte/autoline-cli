.PHONY: test race vet build lint tidy snapshot doctor

test:
	go test ./...

race:
	go test -race ./...

vet:
	go vet ./...

build:
	go build -trimpath -ldflags='-s -w' -o bin/autoline ./cmd/autoline

lint: vet

snapshot: test
	@echo "Template snapshots are validated by generator tests."

doctor:
	go run ./cmd/autoline doctor

tidy:
	go mod tidy
