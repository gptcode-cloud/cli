APP_NAME=gptcode
APP_PATH=./cmd/gptcode

GOBIN?=$(shell go env GOBIN)
ifeq ($(GOBIN),)
    GOBIN=$(HOME)/go/bin
endif

ML_DIR=ml/complexity_detection

.PHONY: all build install dev clean test train-ml train-complexity install-ml

all: build

build:
	@echo "-> Building $(APP_NAME)..."
	@go build -o bin/$(APP_NAME) $(APP_PATH)

install: build
	@echo "-> Installing $(APP_NAME) to $(GOBIN)..."
	@mkdir -p $(GOBIN)
	@cp bin/$(APP_NAME) $(GOBIN)/
	@echo "-> Running gptcode setup..."
	@$(GOBIN)/$(APP_NAME) setup
	@echo "-> Done."

install-ml: install
	@echo "-> Setting up ML models..."
	@$(MAKE) train-complexity
	@echo "-> ML models ready."

dev:
	@echo "-> Running in dev mode..."
	@go run $(APP_PATH)

clean:
	@echo "-> Cleaning..."
	@rm -rf bin/
	@rm -rf $(ML_DIR)/venv
	@rm -f $(ML_DIR)/models/*.json

test:
	@echo "-> Running Go tests..."
	@go test ./...

# ML Training targets
train-ml:
	@gptcode train

train-complexity:
	@gptcode train complexity_detection
