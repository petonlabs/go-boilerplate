#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

mkdir -p "$ROOT_DIR/tmp"

# Load backend .env if present (simple parser, ignores comments)
if [ -f "$ROOT_DIR/apps/backend/.env" ]; then
  set -o allexport
  # shellcheck disable=SC2046
  eval $(grep -v '^#' "$ROOT_DIR/apps/backend/.env" | sed -E 's/"/\\"/g' | xargs -0 2>/dev/null || true)
  set +o allexport
fi

# Defaults (match docker-compose dev defaults)
: "${DB_DSN:=postgresql://app:${POSTGRES_PASSWORD:-g0B01l3rP@ss!2025}@localhost:5432/app?sslmode=disable}"
: "${REDIS_PASS:=${REDIS_PASSWORD:-g0R3d1sP@ss!2025}}"

NO_FRONTEND=0
FOREGROUND=0
while [ $# -gt 0 ]; do
  case "$1" in
    --no-frontend) NO_FRONTEND=1 ;;
    --foreground) FOREGROUND=1 ;;
    --help)
      cat <<'USAGE'
Usage: dev-start.sh [--no-frontend] [--foreground]

Options:
  --no-frontend   Skip starting frontend dev server
  --foreground    Run backend in foreground (useful for debugging)
USAGE
      exit 0
      ;;
  esac
  shift
done

echo "Starting docker compose (backend, postgres, redis, backup)..."
docker compose up -d

echo -n "Waiting for Postgres to become ready"
for i in $(seq 1 60); do
  if docker compose exec -T postgres pg_isready -U "${DATABASE_USER:-app}" -d "${DATABASE_NAME:-app}" >/dev/null 2>&1; then
    echo " -> ready"
    break
  fi
  echo -n "."
  sleep 1
done

echo -n "Waiting for Redis to become ready"
for i in $(seq 1 60); do
  if docker compose exec -T redis sh -lc "redis-cli -a '${REDIS_PASS}' ping" >/dev/null 2>&1; then
    echo " -> ready (auth)"
    break
  fi
  if docker compose exec -T redis redis-cli ping >/dev/null 2>&1; then
    echo " -> ready (no auth)"
    break
  fi
  echo -n "."
  sleep 1
done

echo "Applying DB migrations (if tern installed)..."
cd "$ROOT_DIR/apps/backend"
if command -v tern >/dev/null 2>&1; then
  DB_DSN_ARG="${DB_DSN:-postgresql://app:${POSTGRES_PASSWORD:-g0B01l3rP@ss!2025}@localhost:5432/app?sslmode=disable}"
  tern migrate -m ./internal/database/migrations --conn-string "$DB_DSN_ARG" || true
else
  echo "tern not found; skipping migrations (install: https://github.com/go-pg/tern)"
fi

if [ "$FOREGROUND" -eq 1 ]; then
  echo "Starting backend in foreground (go run)":
  exec go run ./cmd/go-boilerplate
else
  echo "Starting backend in background (go run) -> $ROOT_DIR/tmp/backend.log"
  nohup go run ./cmd/go-boilerplate > "$ROOT_DIR/tmp/backend.log" 2>&1 &
  echo "Backend pid: $!"
fi

if [ "$NO_FRONTEND" -eq 0 ] && [ -d "$ROOT_DIR/apps/frontend" ]; then
  echo "Building frontend helper packages (openapi)"
  if [ -d "$ROOT_DIR/packages/openapi" ]; then
    (cd "$ROOT_DIR/packages/openapi" && (bun run build || npm run build || true))
  fi
  echo "Starting frontend dev server in background -> $ROOT_DIR/tmp/frontend.log"
  nohup sh -c "cd $ROOT_DIR/apps/frontend && bun dev" > "$ROOT_DIR/tmp/frontend.log" 2>&1 &
  echo "Frontend pid: $!"
fi

echo
echo "Done. Tail logs with: tail -f tmp/backend.log"
