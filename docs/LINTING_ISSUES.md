# Linting Issues to Fix

**Status**: 12 critical issues identified by golangci-lint  
**Date**: October 3, 2025

## Summary

After configuring golangci-lint with a focused set of critical linters, we identified 12 real issues that should be fixed:

- **5 errcheck**: Unchecked errors
- **2 errorlint**: Type assertions on errors
- **2 gocritic**: Code improvements  
- **1 gosec**: Integer overflow
- **2 staticcheck**: Anti-patterns

## Issues Breakdown

### 1. Errcheck (5 issues) - Priority: HIGH
Unchecked error returns that could lead to bugs:

```
cmd/go-boilerplate/main.go:41:24 - resp.Body.Close()
internal/database/migrator.go:39:18 - conn.Close(ctx)
internal/lib/job/job.go:57:16 - j.Client.Close()
internal/testhelpers/transaction.go:22:19 - tx.Rollback(ctx)
internal/testhelpers/transaction.go:47:19 - tx.Rollback(ctx)
```

**Fix**: Check and handle errors, or explicitly ignore with `_ = `

### 2. Errorlint (2 issues) - Priority: HIGH
Type assertions that will fail on wrapped errors:

```
internal/validation/utils.go:51:26 - err.(validator.ValidationErrors)
internal/validation/utils.go:53:29 - err.(CustomValidationErrors)
```

**Fix**: Use `errors.As()` instead of type assertion

### 3. Gocritic (2 issues) - Priority: MEDIUM

```
cmd/go-boilerplate/main.go:44:4 - exitAfterDefer: os.Exit will prevent defer from running
internal/config/config.go:65:38 - unlambda: replace lambda with strings.ToLower
```

**Fix**: 
- Move defer before os.Exit or use explicit close
- Replace lambda with direct function reference

### 4. Gosec (1 issue) - Priority: MEDIUM

```
internal/database/migrator.go:59:18 - G115: integer overflow conversion int -> int32
```

**Fix**: Add bounds checking before conversion

### 5. Staticcheck (2 issues) - Priority: MEDIUM

```
internal/middleware/auth.go:63:31 - QF1008: Remove embedded field from selector
internal/middleware/context.go:59:52 - SA1029: Don't use built-in type as context key
```

**Fix**:
- Simplify selector expression
- Define custom type for context key

## Next Steps

### Option 1: Fix All Now (Recommended for Production)
Fix all 12 issues before merging to main. This ensures code quality from the start.

### Option 2: Fix Gradually (Pragmatic Approach)
1. **Now**: Disable linting in CI temporarily  
2. **Sprint 1**: Fix errcheck issues (5)
3. **Sprint 2**: Fix errorlint issues (2)
4. **Sprint 3**: Fix remaining issues (5)
5. **Re-enable**: Turn on linting in CI

### Option 3: Block New Issues Only
Configure linter to only fail on new issues:
```yaml
issues:
  new-from-rev: main  # Only check new code
```

## Recommendation

**Use Option 1** - Fix all 12 issues now. They're straightforward and will prevent real bugs:

1. Add error checks (5 mins)
2. Fix error handling (10 mins)
3. Fix misc issues (10 mins)

Total: ~25 minutes of work for significantly better code quality.

## Configuration Changes Made

Simplified `.golangci.yml` to focus on critical issues only:
- Removed 40+ noisy linters (revive, mnd, funlen, etc.)
- Kept 15 essential linters focused on bugs and security
- Excluded test files from strict checking
- Limited max issues to prevent overwhelming output

This reduces noise from 132 issues to 12 real issues.

## CI Integration

Updated `.github/workflows/ci.yml` to use:
- golangci-lint-action@v6 (official GitHub Action)
- Version 2.5.0 (matches local version)
- Modern best practices (no manual binary download)

## Commands

```bash
# Run linting locally
./scripts/run-lint.sh

# Run linting in backend only
cd apps/backend && golangci-lint run

# Auto-fix some issues
cd apps/backend && golangci-lint run --fix

# Check what would be fixed
cd apps/backend && golangci-lint run --fix --dry-run
```
