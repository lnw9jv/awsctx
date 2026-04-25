VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

build:
	go build -buildvcs=false -ldflags "-X main.version=$(VERSION)" -o awsctx .

install: build
	mv awsctx /usr/local/bin/awsctx

test:
	go test ./...

test-integration:
	go test -tags integration ./...
