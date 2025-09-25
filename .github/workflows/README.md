# CI/CD Workflows Status

## Temporarily Disabled Workflows

The following workflows have been temporarily disabled due to GitHub Actions billing issues:

- `main.yml.disabled` - Main CI/CD Pipeline
- `security-enhanced.yml.disabled` - Enhanced Security Scanning  
- `test-coverage.yml.disabled` - Test Coverage & Quality

## Issue Details

**Error**: "The job was not started because recent account payments have failed or your spending limit needs to be increased"

**Impact**: All PR checks fail, preventing merges even for working code

## Active Workflows

- `security-audit.yml` - Security audit (may also fail due to billing)
- `release.yml` - Release workflow
- `staging-deployment.yml` - Staging deployment

## To Re-enable

1. Resolve GitHub Actions billing issues in repository settings
2. Rename `.disabled` files back to `.yml`:
   ```bash
   mv main.yml.disabled main.yml
   mv security-enhanced.yml.disabled security-enhanced.yml
   mv test-coverage.yml.disabled test-coverage.yml
   ```

## Alternative: Local Testing

While CI/CD is disabled, run tests locally:

```bash
# Run tests
go test ./...

# Run security checks
go mod audit

# Build and test Docker image
docker build -f Dockerfile.local -t cloudbridge-relay:test .

# Run linting
golangci-lint run
```

## Temporary PR Strategy

With CI/CD disabled:
1. **Dependabot PRs can be merged** without waiting for checks
2. **Manual code review** becomes more important
3. **Local testing** before merge is critical
4. **Rollback plan** should be ready

**Status**: CI/CD disabled on 2025-09-23 due to billing issues




