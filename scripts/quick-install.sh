#!/bin/bash
set -e

echo "GitHub Auto Deployer - Quick Install Script"
echo "==========================================="
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go 1.21 or later."
    exit 1
fi

# Check if git is installed
if ! command -v git &> /dev/null; then
    echo "Error: git is not installed. Please install git."
    exit 1
fi

# Build the application
echo "Building the application..."
make build

if [ ! -f "build/deployer" ]; then
    echo "Error: Build failed. Please check the error messages above."
    exit 1
fi

echo ""
echo "Build successful!"
echo ""
echo "Next steps:"
echo "1. Set up your GitHub App (see docs/GITHUB_APP_SETUP.md)"
echo "2. Run: ./build/deployer init"
echo "3. Run: ./build/deployer add-folder"
echo "4. Run: ./build/deployer install"
echo "5. Run: systemctl --user start github-deployer"
echo "6. (Optional) Enable lingering: loginctl enable-linger \$USER"
echo ""
echo "For more information, see README.md"
