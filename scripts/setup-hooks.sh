#!/usr/bin/env bash
# Setup Git hooks for the repository

set -euo pipefail

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo ""
echo -e "${BLUE}🔧 Setting up Git hooks...${NC}"
echo ""

# Get the repository root
REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

# Create .git/hooks directory if it doesn't exist
mkdir -p .git/hooks

# Copy pre-commit hook
if [ -f "scripts/git-hooks/pre-commit" ]; then
    cp scripts/git-hooks/pre-commit .git/hooks/pre-commit
    chmod +x .git/hooks/pre-commit
    echo -e "${GREEN}✓${NC} Installed pre-commit hook"
else
    echo -e "${YELLOW}⚠️${NC}  scripts/git-hooks/pre-commit not found"
fi

echo ""
echo -e "${GREEN}╔═══════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║                                                       ║${NC}"
echo -e "${GREEN}║          ✅  Git Hooks Installed Successfully  ✅      ║${NC}"
echo -e "${GREEN}║                                                       ║${NC}"
echo -e "${GREEN}╚═══════════════════════════════════════════════════════╝${NC}"
echo ""

echo -e "${BLUE}Installed hooks:${NC}"
echo -e "  • pre-commit: Runs golangci-lint before each commit"
echo ""

echo -e "${YELLOW}Note:${NC}"
echo -e "  • Hooks will run automatically on ${BLUE}git commit${NC}"
echo -e "  • To skip hooks temporarily: ${BLUE}git commit --no-verify${NC}"
echo -e "  • To run lint manually: ${BLUE}./scripts/lint.sh${NC}"
echo ""
