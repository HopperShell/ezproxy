BINARY=ezproxy
VERSION?=0.1.0

.PHONY: build test clean release

build:
	go build -o $(BINARY) ./cmd/ezproxy

test:
	go test ./... -v

clean:
	rm -f $(BINARY) $(BINARY)-*

release:
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY)-darwin-amd64 ./cmd/ezproxy
	GOOS=darwin GOARCH=arm64 go build -o $(BINARY)-darwin-arm64 ./cmd/ezproxy
	GOOS=linux GOARCH=amd64 go build -o $(BINARY)-linux-amd64 ./cmd/ezproxy
	GOOS=linux GOARCH=arm64 go build -o $(BINARY)-linux-arm64 ./cmd/ezproxy
