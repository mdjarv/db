BINARY := db
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

.PHONY: build fmt lint test test-integration clean

build:
	go build -ldflags "-X github.com/mdjarv/db/cmd.version=$(VERSION)" -o $(BINARY) .

fmt:
	gofmt -w .

lint: fmt
	golangci-lint run ./...

test:
	go test ./...

test-integration:
	go test -tags integration ./...

clean:
	rm -f $(BINARY)
