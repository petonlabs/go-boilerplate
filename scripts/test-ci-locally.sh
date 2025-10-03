#!/usr/bin/env bash
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_step() {
    echo -e "${GREEN}==>${NC} $1"
}

print_error() {
    echo -e "${RED}ERROR:${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}WARNING:${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

# Change to backend directory
cd "$(dirname "$0")/../apps/backend" || exit 1

print_step "Starting CI simulation test..."

# Step 1: Clean environment (simulate fresh clone)
print_step "Step 1: Cleaning environment (simulating fresh clone)..."
rm -rf vendor 2>/dev/null || true
go clean -cache -modcache -testcache 2>/dev/null || print_warning "Could not clean all caches (may need sudo)"

# Step 2: Set environment variables (simulate CI)
print_step "Step 2: Setting CI environment variables..."
export GOPROXY="https://proxy.golang.org,direct"
export GOSUMDB="sum.golang.org"
export GO111MODULE="on"
export CGO_ENABLED="0"
# Set dummy values for required env vars to prevent build failures
export DATABASE_URL="${DATABASE_URL:-postgresql://dummy:dummy@localhost:5432/dummy}"
export REDIS_URL="${REDIS_URL:-redis://localhost:6379}"
export CLERK_SECRET_KEY="${CLERK_SECRET_KEY:-sk_test_dummy}"
export NEW_RELIC_LICENSE_KEY="${NEW_RELIC_LICENSE_KEY:-dummy_license_key}"
export RESEND_API_KEY="${RESEND_API_KEY:-re_dummy_key}"
print_success "CI environment configured (with dummy values for missing env vars)"

# Step 3: Verify go.mod
print_step "Step 3: Verifying go.mod..."
if ! grep -q "^go 1.24" go.mod; then
    print_error "go.mod does not specify Go 1.24+"
    exit 1
fi
echo "✓ go.mod requires: $(grep '^go ' go.mod)"

# Step 4: Download modules
print_step "Step 4: Downloading modules..."
if ! go mod download; then
    print_error "Failed to download modules"
    exit 1
fi
echo "✓ Modules downloaded successfully"

# Step 5: Verify modules
print_step "Step 5: Verifying module checksums..."
if ! go mod verify; then
    print_error "Module verification failed"
    exit 1
fi
echo "✓ All modules verified"

# Step 6: Check formatting
print_step "Step 6: Checking code formatting..."
unformatted=$(gofmt -l . 2>/dev/null | grep -v vendor || true)
if [ -n "$unformatted" ]; then
    print_error "The following files are not formatted:"
    echo "$unformatted"
    exit 1
fi
echo "✓ All files properly formatted"

# Step 7: Run go vet
print_step "Step 7: Running go vet..."
if ! go vet ./...; then
    print_error "go vet found issues"
    exit 1
fi
echo "✓ go vet passed"

# Step 8: Build all packages
print_step "Step 8: Building all packages..."
if ! go build -v ./...; then
    print_error "Build failed"
    exit 1
fi
echo "✓ Build successful"

# Step 9: Run tests
print_step "Step 9: Running tests..."
if ! go test -v ./...; then
    print_error "Tests failed"
    exit 1
fi
echo "✓ Tests passed"

# Step 10: Build binary
print_step "Step 10: Building binary..."
mkdir -p bin
if ! go build -o bin/server ./cmd/go-boilerplate; then
    print_error "Binary build failed"
    exit 1
fi
echo "✓ Binary built: bin/server"

# Step 11: Check for undefined references (typecheck)
print_step "Step 11: Type checking for undefined references..."
if ! go build -o /dev/null ./... 2>&1 | tee /tmp/build-output.txt; then
    if grep -q "undefined:" /tmp/build-output.txt; then
        print_error "Found undefined references:"
        grep "undefined:" /tmp/build-output.txt
        exit 1
    fi
fi
echo "✓ No undefined references found"

# Step 12: Verify critical imports
print_step "Step 12: Verifying critical imports in go.mod..."
critical_imports=(
    "github.com/labstack/echo/v4"
    "github.com/clerk/clerk-sdk-go/v2"
    "github.com/go-playground/validator/v10"
    "github.com/newrelic/go-agent/v3/integrations/nrecho-v4"
    "github.com/jackc/pgx/v5"
    "github.com/redis/go-redis/v9"
    "github.com/knadh/koanf/v2"
    "github.com/resend/resend-go/v2"
)

for import in "${critical_imports[@]}"; do
    if ! grep -q "$import" go.mod; then
        print_error "Missing critical import: $import"
        exit 1
    fi
    echo "  ✓ $import"
done

# Success!
echo ""
echo -e "${GREEN}╔═══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║                                                               ║${NC}"
echo -e "${GREEN}║              ✅  CI SIMULATION TEST PASSED  ✅                ║${NC}"
echo -e "${GREEN}║                                                               ║${NC}"
echo -e "${GREEN}╚═══════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo "All steps completed successfully!"
echo "Your code should pass GitHub Actions CI."
echo ""
echo "Summary:"
echo "  • Modules: Downloaded & Verified"
echo "  • Formatting: OK"
echo "  • Vet: Passed"
echo "  • Build: Success"
echo "  • Tests: Passed"
echo "  • Binary: Created"
echo "  • Type Check: No undefined references"
echo "  • Critical Imports: All present"
echo ""
