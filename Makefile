VERSION ?= 1.0.0
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: build build-all clean test

build:
	go build $(LDFLAGS) -o bin/jmvn .

build-all:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/jmvn-windows-amd64.exe .
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/jmvn-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/jmvn-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/jmvn-linux-amd64 .

test:
	go test ./...

clean:
	rm -rf bin
