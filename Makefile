.PHONY: all build test clean example

# Variables
BINARY_NAME=tokentracker
EXAMPLE_BINARY=tokentracker-example
CMD_DIR=./cmd
EXAMPLE_DIR=./example

all: build test

build:
	@echo "Building TokenTracker library..."
	@go build -v ./...

test:
	@echo "Running tests..."
	@go test -v ./...

clean:
	@echo "Cleaning up..."
	@go clean
	@rm -f $(BINARY_NAME)
	@rm -f $(EXAMPLE_BINARY)
	@rm -f $(CMD_DIR)/$(BINARY_NAME)
	@rm -f $(EXAMPLE_DIR)/$(EXAMPLE_BINARY)

example-original:
	@echo "Building and running original example application..."
	@go build -o $(EXAMPLE_DIR)/$(EXAMPLE_BINARY) $(EXAMPLE_DIR)/main.go
	@$(EXAMPLE_DIR)/$(EXAMPLE_BINARY)

example:
	@echo "Building and running new example application..."
	@go build -o $(CMD_DIR)/$(BINARY_NAME) $(CMD_DIR)/main.go
	@$(CMD_DIR)/$(BINARY_NAME)

# For docs and coverage
docs:
	@echo "Generating documentation..."
	@go doc -all ./... > ./docs/api.txt

coverage:
	@echo "Running tests with coverage..."
	@go test -cover ./...
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
