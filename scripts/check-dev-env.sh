#!/usr/bin/env bash
set -euo pipefail

# Check and optionally populate developer env vars so the app can start with
# safe defaults locally. Works with repo .env and apps/backend/.env.
# Usage: scripts/check-dev-env.sh [--fix] [--env-file PATH] [--repo-env PATH] [--dry-run] [--verbose]

ENV_FILE="apps/backend/.env"
REPO_ENV_FILE=".env"
FIX=0
DRY_RUN=0
VERBOSE=0

print_help() {
  cat <<'EOF'
Usage: check-dev-env.sh [--fix] [--env-file PATH] [--repo-env PATH] [--dry-run] [--verbose]

Scans environment sources for the variables the backend expects. If --fix is
provided missing values will be appended to the backend env file with safe
development defaults.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --fix) FIX=1; shift ;;
    --dry-run) DRY_RUN=1; shift ;;
    --env-file) ENV_FILE="$2"; shift 2 ;;
    --repo-env) REPO_ENV_FILE="$2"; shift 2 ;;
    --verbose) VERBOSE=1; shift ;;
    --help) print_help; exit 0 ;;
    *) echo "Unknown arg: $1" >&2; print_help; exit 2 ;;
  esac
done

required_vars=(
  PRIMARY__ENV
  SERVER__PORT
  SERVER__READ_TIMEOUT
  SERVER__WRITE_TIMEOUT
  SERVER__IDLE_TIMEOUT
  SERVER__CORS_ALLOWED_ORIGINS
  DATABASE__HOST
  DATABASE__PORT
  DATABASE__USER
  DATABASE__PASSWORD
  DATABASE__NAME
  DATABASE__SSL_MODE
  DATABASE__MAX_OPEN_CONNS
  DATABASE__MAX_IDLE_CONNS
  DATABASE__CONN_MAX_LIFETIME
  DATABASE__CONN_MAX_IDLE_TIME
  REDIS__ADDRESS
  AUTH__SECRET_KEY
  INTEGRATION__RESEND_API_KEY
  OBSERVABILITY__NEW_RELIC__LICENSE_KEY
  OBSERVABILITY__SERVICE_NAME
  OBSERVABILITY__LOGGING__LEVEL
  OBSERVABILITY__LOGGING__FORMAT
  OBSERVABILITY__HEALTH_CHECKS__INTERVAL
  OBSERVABILITY__HEALTH_CHECKS__TIMEOUT
  S3__ACCESS_KEY_ID
  S3__SECRET_ACCESS_KEY
  S3__BUCKET
  S3__ENDPOINT
  DSPY__AZURE_ENDPOINT
  DSPY__AZURE_API_KEY
)

default_for() {
  case "$1" in
    PRIMARY__ENV) echo development ;;
    SERVER__PORT) echo 8080 ;;
    SERVER__READ_TIMEOUT) echo 30 ;;
    SERVER__WRITE_TIMEOUT) echo 30 ;;
    SERVER__IDLE_TIMEOUT) echo 120 ;;
    SERVER__CORS_ALLOWED_ORIGINS) echo "http://localhost:3000,http://localhost:5173" ;;
    DATABASE__HOST) echo postgres ;;
    DATABASE__PORT) echo 5432 ;;
    DATABASE__USER) echo app ;;
    DATABASE__PASSWORD) echo 'g0B01l3rP@ss!2025' ;;
    DATABASE__NAME) echo app ;;
    DATABASE__SSL_MODE) echo disable ;;
    DATABASE__MAX_OPEN_CONNS) echo 25 ;;
    DATABASE__MAX_IDLE_CONNS) echo 5 ;;
    DATABASE__CONN_MAX_LIFETIME) echo 300 ;;
    DATABASE__CONN_MAX_IDLE_TIME) echo 60 ;;
    REDIS__ADDRESS) echo redis:6379 ;;
    AUTH__SECRET_KEY) echo dev-secret-key-please-change ;;
    INTEGRATION__RESEND_API_KEY) echo dev-resend-key ;;
    OBSERVABILITY__NEW_RELIC__LICENSE_KEY) echo dev-newrelic-key ;;
    OBSERVABILITY__SERVICE_NAME) echo go-boilerplate ;;
    OBSERVABILITY__LOGGING__LEVEL) echo debug ;;
    OBSERVABILITY__LOGGING__FORMAT) echo json ;;
    OBSERVABILITY__HEALTH_CHECKS__INTERVAL) echo 30s ;;
    OBSERVABILITY__HEALTH_CHECKS__TIMEOUT) echo 5s ;;
    S3__ACCESS_KEY_ID) echo "" ;;
    S3__SECRET_ACCESS_KEY) echo "" ;;
    S3__BUCKET) echo "" ;;
    S3__ENDPOINT) echo "" ;;
    DSPY__AZURE_ENDPOINT) echo "" ;;
    DSPY__AZURE_API_KEY) echo "" ;;
    *) echo "" ;;
  esac
}

file_has_var() {
  local var="$1"
  local file="$2"
  [[ -f "$file" ]] || return 1
  # match VAR= or VAR="..." or export VAR=...
  grep -E "^\s*(export\s+)?${var}=" "$file" >/dev/null 2>&1
}

env_has_var() {
  local var="$1"
  if printenv "$var" >/dev/null 2>&1; then
    return 0
  fi
  return 1
}

present=()
missing=()

for v in "${required_vars[@]}"; do
  if env_has_var "$v"; then
    present+=("$v (env)")
    [[ $VERBOSE -eq 1 ]] && echo "FOUND $v in environment"
    continue
  fi
  if file_has_var "$v" "$ENV_FILE"; then
    present+=("$v ($ENV_FILE)")
    [[ $VERBOSE -eq 1 ]] && echo "FOUND $v in $ENV_FILE"
    continue
  fi
  if file_has_var "$v" "$REPO_ENV_FILE"; then
    present+=("$v ($REPO_ENV_FILE)")
    [[ $VERBOSE -eq 1 ]] && echo "FOUND $v in $REPO_ENV_FILE"
    continue
  fi
  missing+=("$v")
done

echo "Present: ${#present[@]} variables"
if [[ ${#present[@]} -gt 0 && $VERBOSE -eq 1 ]]; then
  for p in "${present[@]}"; do echo "  - $p"; done
fi

if [[ ${#missing[@]} -eq 0 ]]; then
  echo "All required environment variables found."
  exit 0
fi

echo "Missing ${#missing[@]} required variables:" >&2
for m in "${missing[@]}"; do echo "  - $m" >&2; done

if [[ $FIX -eq 0 ]]; then
  echo "Run with --fix to append safe development defaults into $ENV_FILE" >&2
  exit 2
fi

if [[ $DRY_RUN -eq 1 ]]; then
  echo "Dry-run: would append the following to $ENV_FILE:" >&2
  for m in "${missing[@]}"; do
    printf "%s=%s\n" "$m" "$(default_for "$m")" >&2
  done
  exit 0
fi

mkdir -p "$(dirname "$ENV_FILE")"
if [[ ! -f "$ENV_FILE" ]]; then
  echo "# Created by scripts/check-dev-env.sh on $(date)" > "$ENV_FILE"
  echo "# This file contains safe local development defaults. Do not commit secrets to source control." >> "$ENV_FILE"
  echo >> "$ENV_FILE"
fi

echo "Appending ${#missing[@]} missing vars to $ENV_FILE" >&2
for m in "${missing[@]}"; do
  val="$(default_for "$m")"
  # If value contains spaces or commas, quote it
  if [[ "$val" =~ [[:space:],] ]]; then
    printf "%s=\"%s\"\n" "$m" "$val" >> "$ENV_FILE"
  else
    printf "%s=%s\n" "$m" "$val" >> "$ENV_FILE"
  fi
done

echo "Appended missing vars to $ENV_FILE" >&2
exit 0
