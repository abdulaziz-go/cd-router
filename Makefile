PROJECT_NAME=dispatcher
DISPATCHER_CLI_DIR=dispatcher_cli
DISPATCHER_SERVER_DIR=dispatcher_server
BUILD_DIR=build

.PHONY: all build clean

all: build

build:
	@echo "Building $(PROJECT_NAME) CLI..."
	@mkdir -p $(BUILD_DIR)
	@cd $(DISPATCHER_CLI_DIR) && go build -o ../$(BUILD_DIR)/$(PROJECT_NAME)
	@echo "Build complete: $(BUILD_DIR)/$(PROJECT_NAME)"

clean:
	@echo "Cleaning up build files..."
	@rm -rf $(BUILD_DIR)
	@echo "Cleanup complete."