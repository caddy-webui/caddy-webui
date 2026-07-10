APP_NAME := caddy-webui
BUILD_DIR := bin
GO := go
LDFLAGS := -ldflags="-s -w"

.PHONY: build clean run

build:
	CGO_ENABLED=1 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) .

clean:
	rm -rf $(BUILD_DIR)

run:
	CGO_ENABLED=1 $(GO) run .

cross-build:
	CGO_ENABLED=1 CC=x86_64-linux-gnu-gcc GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 .

deps:
	$(GO) mod tidy

fmt:
	gofmt -w .
