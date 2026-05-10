#!/bin/bash

# Kindle Highlights Parser - Custom Installer
# This script handles dependency checks and custom binary naming.

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Kindle Highlights Parser Setup ===${NC}\n"

# 1. Check Prerequisites
echo -e "Checking prerequisites..."

if ! command -v go >/dev/null 2>&1; then
    echo -e "${RED}Error: Go is not installed. Please install Go 1.24+ before continuing.${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Go is installed: $(go version)${NC}"

if ! command -v rg >/dev/null 2>&1; then
    echo -e "${YELLOW}Warning: ripgrep (rg) is not installed. System search feature will be disabled.${NC}"
else
    echo -e "${GREEN}✓ ripgrep is installed.${NC}"
fi

# 2. Check PATH
GOPATH_BIN=$(go env GOPATH)/bin
if [[ ":$PATH:" != *":$GOPATH_BIN:"* ]]; then
    echo -e "${YELLOW}Warning: $GOPATH_BIN is not in your PATH.${NC}"
    echo -e "To run the app from anywhere, add it to your shell profile (e.g., .zshrc or .bashrc):"
    echo -e "${BLUE}export PATH=\"\$PATH:$GOPATH_BIN\"${NC}\n"
else
    echo -e "${GREEN}✓ Go bin directory is in your PATH.${NC}"
fi

# 3. Custom Binary Name
DEFAULT_NAME="kindle-parser"
echo -e "\nWhat name would you like to use for the command? (default: $DEFAULT_NAME)"
read -p "Enter name (or press Enter for default): " CUSTOM_NAME

BINARY_NAME="${CUSTOM_NAME:-$DEFAULT_NAME}"
INSTALL_PATH="$GOPATH_BIN/$BINARY_NAME"

# 4. Build and Install
echo -e "\nBuilding and installing as '${BLUE}$BINARY_NAME${NC}'..."
go build -o "$INSTALL_PATH" .

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Success! Installed to $INSTALL_PATH${NC}"
    echo -e "\nYou can now run the app using: ${BLUE}$BINARY_NAME${NC}"
else
    echo -e "${RED}Error: Build failed.${NC}"
    exit 1
fi
