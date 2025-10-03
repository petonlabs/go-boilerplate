#!/usr/bin/env bash
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
check_script="$here/check-dev-env.sh"

tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT

echo "Test 1: no envs anywhere -> expect missing and --fix to create file"
mkdir -p "$tmpdir/repo"
pushd "$tmpdir/repo" >/dev/null
"$check_script" --env-file "apps/backend/.env" --repo-env ".env" --dry-run || true

echo "Test 2: fix mode -> writes apps/backend/.env"
"$check_script" --env-file "apps/backend/.env" --repo-env ".env" --fix
[[ -f apps/backend/.env ]] || { echo "FAIL: apps/backend/.env not created"; exit 2; }

echo "Test 3: re-run finds everything present"
"$check_script" --env-file "apps/backend/.env" --repo-env ".env"

popd >/dev/null
echo "All tests completed (temp dir: $tmpdir)"
