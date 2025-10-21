.PHONY: build build-linux install install-service uninstall clean test

BINARY_NAME=github-auto-deployer
INSTALL_PATH=/usr/local/bin
SERVICE_PATH=/etc/systemd/system

# Build for current platform
build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) cmd/github-auto-deployer/main.go
	@echo "Build complete: ./$(BINARY_NAME)"

# Build for Linux (useful for cross-compilation from macOS)
build-linux:
	@echo "Building $(BINARY_NAME) for Linux..."
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux-amd64 cmd/github-auto-deployer/main.go
	@echo "Build complete: ./$(BINARY_NAME)-linux-amd64"

# Install binary to system
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	sudo cp $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	sudo chmod +x $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Installation complete!"
	@echo "Run '$(BINARY_NAME) init' to configure"

# Install systemd service
install-service:
	@echo "Installing systemd service..."
	@read -p "Enter the username to run the service as: " username; \
	sudo cp systemd/github-auto-deployer.service $(SERVICE_PATH)/github-auto-deployer@$$username.service
	@echo "Service installed!"
	@echo ""
	@echo "To enable and start the service, run:"
	@echo "  sudo systemctl enable github-auto-deployer@USERNAME"
	@echo "  sudo systemctl start github-auto-deployer@USERNAME"
	@echo ""
	@echo "To check status:"
	@echo "  sudo systemctl status github-auto-deployer@USERNAME"

# Uninstall binary and service
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	sudo rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Removing systemd service files..."
	sudo rm -f $(SERVICE_PATH)/github-auto-deployer@*.service
	sudo systemctl daemon-reload
	@echo "Uninstallation complete!"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-linux-amd64
	@echo "Clean complete!"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies updated!"

# Show help
help:
	@echo "GitHub Auto-Deployer Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  build            - Build binary for current platform"
	@echo "  build-linux      - Build binary for Linux (amd64)"
	@echo "  install          - Install binary to $(INSTALL_PATH)"
	@echo "  install-service  - Install systemd service"
	@echo "  uninstall        - Remove binary and service"
	@echo "  clean            - Remove build artifacts"
	@echo "  test             - Run tests"
	@echo "  deps             - Download and update dependencies"
	@echo "  help             - Show this help message"
