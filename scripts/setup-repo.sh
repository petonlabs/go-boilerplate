#!/usr/bin/env bash
# Robust cross-platform setup script for the repo
# - ensures required tools are installed
# - performs common setup steps (go mod download, verify)
# - defensive: retries, timeouts, informative output
# Usage: ./scripts/setup-repo.sh [--non-interactive]

set -u

RETRY_MAX=5
RETRY_SLEEP=3
NON_INTERACTIVE=0
GO_MIN_VERSION="1.25.0"
GOLANGCI_VERSION="v2.5.0"
INSTALL_GO=0
GO_INSTALL_VERSION="1.25.1"

# Helpers
info(){ printf "[INFO] %s\n" "$*"; }
warn(){ printf "[WARN] %s\n" "$*"; }
err(){ printf "[ERROR] %s\n" "$*" >&2; }

run_retry(){
  local n=0
  local cmd="${*}"
  until [ $n -ge $RETRY_MAX ]
  do
    eval "$cmd" && return 0
    n=$((n+1))
    warn "Command failed (attempt $n/$RETRY_MAX): $cmd"
    sleep $RETRY_SLEEP
  done
  return 1
}

command_exists(){ command -v "$1" >/dev/null 2>&1; }

detect_os(){
  OS="unknown"
  UNAME=$(uname -s)
  case "$UNAME" in
    Linux*) OS=linux ;;
    Darwin*) OS=darwin ;;
    MINGW*|MSYS*|CYGWIN*) OS=windows ;;
  esac
  printf "%s" "$OS"
}

# Parse args
while [ $# -gt 0 ]; do
  case "$1" in
    --non-interactive|-y)
      NON_INTERACTIVE=1
      shift
      ;;
    --install-go)
      INSTALL_GO=1
      shift
      ;;
    --go-version)
      GO_INSTALL_VERSION="$2"
      shift 2
      ;;
    --help|-h)
      cat <<'EOF'
Usage: setup-repo.sh [--non-interactive]

This script installs developer tooling and prepares the repository for development.
It will attempt to install or verify:
 - go (>= 1.25.0) (not installed by script)
 - golangci-lint (v2.5.0)
 - govulncheck
 - gosec
 - delve (dlv)

It also runs: go mod download && go mod verify
EOF
      exit 0
      ;;
    *)
      err "Unknown option: $1"
      exit 2
      ;;
  esac
done

main(){
  info "Starting repository setup"
  OS=$(detect_os)
  info "Detected OS: $OS"

  # Check Go
  if ! command_exists go; then
    err "Go is not installed. Please install Go ${GO_MIN_VERSION}+ first: https://go.dev/dl/"
    exit 1
  fi

  # Verify Go version (simple semver prefix match)
  GOVERSION=$(go version | awk '{print $3}')
  info "Go version: $GOVERSION"

  if [ "$INSTALL_GO" -eq 1 ]; then
    if command_exists go; then
      info "Go already installed: $GOVERSION"
    else
      info "--install-go requested. Attempting to install Go ${GO_INSTALL_VERSION}"
      install_go
      if ! command_exists go; then
        err "Go installation failed or 'go' not in PATH. Please install Go manually."
        exit 1
      fi
      info "Installed Go: $(go version)"
    fi
  fi

  # Ensure GOPATH/bin and GOBIN are in PATH so go install'd binaries are usable
  GOPATH_BIN="$(go env GOPATH 2>/dev/null)/bin"
  GOBIN="$(go env GOBIN 2>/dev/null)"
  if [ -n "$GOBIN" ] && [ -d "$GOBIN" ]; then
    PATH="$PATH:$GOBIN"
  fi
  if [ -d "$GOPATH_BIN" ]; then
    PATH="$PATH:$GOPATH_BIN"
  fi
  export PATH

  # Install golangci-lint
  if command_exists golangci-lint; then
    info "golangci-lint already installed: $(golangci-lint --version | head -n1)"
  else
    info "Installing golangci-lint ${GOLANGCI_VERSION}"
    if command_exists curl; then
      run_retry "curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$(go env GOPATH)/bin ${GOLANGCI_VERSION}"
    elif command_exists wget; then
      run_retry "wget -qO- https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$(go env GOPATH)/bin ${GOLANGCI_VERSION}"
    else
      err "Neither curl nor wget is available. Install one to proceed."
      exit 1
    fi
  fi

  # Install govulncheck
  if command_exists govulncheck; then
    info "govulncheck already installed: $(govulncheck --version 2>/dev/null || true)"
  else
    info "Installing govulncheck"
    run_retry "go install golang.org/x/vuln/cmd/govulncheck@latest"
  fi

  # Install gosec
  if command_exists gosec; then
    info "gosec already installed"
  else
    info "Installing gosec"
    run_retry "go install github.com/securego/gosec/v2/cmd/gosec@latest"
  fi

  # Install delve (optional for macOS/Linux)
  if command_exists dlv; then
    info "delve (dlv) already installed"
  else
    if [ "$OS" = "linux" ] || [ "$OS" = "darwin" ]; then
      info "Installing Delve (dlv)"
      run_retry "go install github.com/go-delve/delve/cmd/dlv@latest"
    else
      warn "Skipping Delve installation on OS: $OS"
    fi
  fi

  # Run go mod download & verify with retries
  # Detect all module roots (go.mod files) under the repository and operate per-module
  info "Looking for go modules in the repo..."
  MODULE_DIRS=()
  while IFS= read -r modfile; do
    dir=$(dirname "$modfile")
    MODULE_DIRS+=("$dir")
  done < <(find . -name 'go.mod' -not -path './.git/*' -print)

  if [ ${#MODULE_DIRS[@]} -eq 0 ]; then
    warn "No go.mod files found in repository. If this is intended, skip module setup."
  else
    for m in "${MODULE_DIRS[@]}"; do
      info "Processing module: $m"
      pushd "$m" >/dev/null || continue
      info "Downloading modules for $m"
      if ! run_retry "go mod download"; then
        warn "go mod download failed for $m"
      fi
      info "Verifying modules for $m"
      if ! run_retry "go mod verify"; then
        warn "go mod verify failed for $m"
      fi

      # Run golangci-lint for this module
      info "Running golangci-lint for $m"
      if command_exists golangci-lint; then
        if ! run_retry "golangci-lint run ./..."; then
          warn "golangci-lint reported issues or failed in $m"
        fi
      else
        warn "golangci-lint not found; skipping lint for $m"
      fi

      # Run govulncheck for this module (only if govulncheck exists)
      if command_exists govulncheck; then
        info "Running govulncheck for $m"
        if ! govulncheck ./...; then
          warn "govulncheck found issues or failed in $m"
        fi
      else
        warn "govulncheck not found; skipping vulnerability scan for $m"
      fi

      popd >/dev/null || true
    done
  fi

  # If no modules were found, run quick checks in current dir as a fallback
  if [ ${#MODULE_DIRS[@]} -eq 0 ]; then
    info "Running quick checks in repository root"
    if command_exists golangci-lint; then
      if ! run_retry "golangci-lint run ./..."; then
        warn "golangci-lint reported issues or failed in repo root"
      fi
    fi
    if command_exists govulncheck; then
      if ! govulncheck ./...; then
        warn "govulncheck reported issues or failed in repo root"
      fi
    fi
  fi

  info "Setup complete. If anything failed above, please inspect the logs and rerun with --non-interactive for automation."
}

install_go(){
  info "Installing Go ${GO_INSTALL_VERSION} for OS=${OS}"
  ARCH=$(uname -m)
  case "$ARCH" in
    x86_64|amd64) GO_ARCH=amd64 ;;
    arm64|aarch64) GO_ARCH=arm64 ;;
    *) GO_ARCH=amd64 ;;
  esac

  # Try platform-specific package managers first
  if [ "$OS" = "darwin" ]; then
    if command_exists brew; then
      info "Using brew to install go"
      if ! run_retry "brew install go"; then
        warn "brew install failed, will try manual tarball install"
      else
        return 0
      fi
    fi
  elif [ "$OS" = "linux" ]; then
    if command_exists apt-get && [ "$NON_INTERACTIVE" -eq 1 ]; then
      info "Attempting apt-get install golang (may not be latest)"
      if run_retry "sudo apt-get update && sudo apt-get install -y golang"; then
        return 0
      fi
    fi
  elif [ "$OS" = "windows" ]; then
    if command_exists choco; then
      info "Using choco to install golang"
      if run_retry "choco install golang -y"; then
        return 0
      fi
    fi
    if command_exists scoop; then
      info "Using scoop to install golang"
      if run_retry "scoop install go"; then
        return 0
      fi
    fi
  fi

  # Fallback: download tarball from go.dev and extract to /usr/local (requires sudo)
  FNAME="go${GO_INSTALL_VERSION}.${OS}-${GO_ARCH}.tar.gz"
  URL="https://go.dev/dl/${FNAME}"
  TMPFILE="/tmp/${FNAME}"
  info "Downloading ${URL}"
  if command_exists curl; then
    if ! run_retry "curl -fsSL -o ${TMPFILE} ${URL}"; then
      err "Failed to download Go tarball from ${URL}"
      return 1
    fi
  elif command_exists wget; then
    if ! run_retry "wget -q -O ${TMPFILE} ${URL}"; then
      err "Failed to download Go tarball from ${URL}"
      return 1
    fi
  else
    err "No downloader (curl/wget) available to fetch Go tarball"
    return 1
  fi

  info "Extracting to /usr/local (requires sudo)"
  if command_exists sudo; then
    if ! run_retry "sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf ${TMPFILE}"; then
      err "Failed to extract Go to /usr/local"
      return 1
    fi
  else
    err "sudo not available; cannot install Go to /usr/local. Please extract ${TMPFILE} manually."
    return 1
  fi

  # Ensure /usr/local/go/bin on PATH for this session
  if [ -d "/usr/local/go/bin" ]; then
    PATH="$PATH:/usr/local/go/bin"
    export PATH
  fi

  # Verify installation
  if command_exists go; then
    info "go installed: $(go version)"
    return 0
  fi
  return 1
}

main "$@"
