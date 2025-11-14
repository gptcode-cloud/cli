APP_NAME=chu
APP_PATH=./cmd/chu

GOBIN?=$(shell go env GOBIN)
ifeq ($(GOBIN),)
    GOBIN=$(HOME)/go/bin
endif

.PHONY: all build install dev clean test

all: build

build:
	@echo "-> Building $(APP_NAME)..."
	@go build -o bin/$(APP_NAME) $(APP_PATH)

install: build
	@echo "-> Installing $(APP_NAME) to $(GOBIN)..."
	@mkdir -p $(GOBIN)
	@cp bin/$(APP_NAME) $(GOBIN)/
	@echo "-> Running chu setup..."
	@$(GOBIN)/$(APP_NAME) setup
	@echo "-> Done."

dev:
	@echo "-> Running in dev mode..."
	@go run $(APP_PATH)

clean:
	@echo "-> Cleaning..."
	@rm -rf bin/

test:
	@echo "-> Running Go tests..."
	@go test ./...
