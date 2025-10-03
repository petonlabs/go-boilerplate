#!/usr/bin/env bash
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Change to backend directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$SCRIPT_DIR/../apps/backend"

cd "$BACKEND_DIR" || {
    echo -e "${RED}ERROR:${NC} Could not find backend directory"
    exit 1
}

echo -e "${BLUE}Running golangci-lint...${NC}"
echo ""

# Check if golangci-lint is installed
if ! command -v golangci-lint &> /dev/null; then
    echo -e "${YELLOW}WARNING:${NC} golangci-lint not found. Installing..."
    
    # Install golangci-lint
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        if command -v brew &> /dev/null; then
            brew install golangci-lint
        else
            curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.59.0
        fi
    else
        # Linux
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.59.0
    fi
    
    if ! command -v golangci-lint &> /dev/null; then
        echo -e "${RED}ERROR:${NC} Failed to install golangci-lint"
        echo "Please install it manually: https://golangci-lint.run/usage/install/"
        exit 1
    fi
fi

# Get golangci-lint version
LINT_VERSION=$(golangci-lint --version | head -n1)
echo -e "${GREEN}✓${NC} Using $LINT_VERSION"
echo ""

# Run golangci-lint
echo -e "${BLUE}Analyzing code...${NC}"
if golangci-lint run --config .golangci.yml ./...; then
    echo ""
    echo -e "${GREEN}╔═══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║                                                               ║${NC}"
    echo -e "${GREEN}║                  ✅  LINTING PASSED  ✅                        ║${NC}"
    echo -e "${GREEN}║                                                               ║${NC}"
    echo -e "${GREEN}╚═══════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    exit 0
else
    echo ""
    echo -e "${RED}╔═══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${RED}║                                                               ║${NC}"
    echo -e "${RED}║                  ❌  LINTING FAILED  ❌                        ║${NC}"
    echo -e "${RED}║                                                               ║${NC}"
    echo -e "${RED}╚═══════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${YELLOW}Fix the issues above before committing.${NC}"
    echo -e "${YELLOW}Or use 'git commit --no-verify' to bypass (not recommended).${NC}"
    echo ""
    exit 1
fi
