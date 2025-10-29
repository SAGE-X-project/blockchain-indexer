# CI/CD Documentation

This document describes the Continuous Integration and Continuous Deployment (CI/CD) pipelines for the Blockchain Indexer project.

## Table of Contents

- [Overview](#overview)
- [CI Pipeline](#ci-pipeline)
- [Release Pipeline](#release-pipeline)
- [Dependency Management](#dependency-management)
- [Badges and Status](#badges-and-status)

---

## Overview

The project uses **GitHub Actions** for automated CI/CD workflows:

- **CI Workflow** (`.github/workflows/ci.yml`): Runs on every push and pull request
- **Release Workflow** (`.github/workflows/release.yml`): Runs on version tags
- **Dependabot** (`.github/dependabot.yml`): Automated dependency updates

---

## CI Pipeline

### Trigger Events

The CI pipeline runs on:
- Push to `main`, `develop`, or `feat/*` branches
- Pull requests to `main` or `develop`

### Jobs

#### 1. Lint
- Runs `golangci-lint` with 5-minute timeout
- Checks code formatting with `gofmt`
- Ensures code quality and style consistency

#### 2. Unit Tests
- Tests against Go 1.21 and 1.22
- Runs all unit tests
- Parallel execution for faster feedback

#### 3. Test Coverage
- Generates coverage report
- Uploads to Codecov
- Enforces 70% minimum coverage threshold
- Comments coverage report on pull requests
- **Fails if coverage is below 70%**

#### 4. Integration Tests
- Runs integration tests
- Starts actual server for testing
- Tests all API endpoints (REST, GraphQL, gRPC)
- Only runs on non-draft pull requests

#### 5. Build
- Cross-compilation for multiple platforms:
  - Linux (amd64, arm64)
  - macOS/Darwin (amd64, arm64)
  - Windows (amd64)
- Uploads build artifacts (7-day retention)
- Verifies builds succeed before merging

#### 6. Security Scan
- Runs Gosec security scanner
- Uploads results to GitHub Security tab
- Identifies potential security vulnerabilities

#### 7. Dependency Review
- Reviews dependency changes in PRs
- Checks for known vulnerabilities
- Alerts on risky dependencies

#### 8. Race Detection
- Runs tests with Go race detector
- Identifies potential race conditions
- Critical for concurrent code

#### 9. Benchmarks
- Runs on `main` branch only
- Tracks performance metrics
- Ensures no performance regressions

### Status Checks

All jobs except benchmarks are **required** for merging pull requests.

---

## Release Pipeline

### Trigger

The release pipeline runs when a version tag is pushed:

```bash
git tag v1.0.0
git push origin v1.0.0
```

### Tag Format

Tags must follow semantic versioning: `v*.*.*`

Examples:
- `v1.0.0` - Major release
- `v1.2.0` - Minor release
- `v1.2.3` - Patch release
- `v2.0.0-beta.1` - Pre-release

### Release Jobs

#### 1. Create Release
- Generates changelog from git commits
- Creates GitHub release
- Links binaries and Docker images

#### 2. Build and Upload Binaries
- Builds for all supported platforms
- Creates archives (`.tar.gz` for Unix, `.zip` for Windows)
- Uploads to GitHub release
- Version info embedded in binaries

**Binary naming:**
```
blockchain-indexer-v1.0.0-linux-amd64.tar.gz
blockchain-indexer-v1.0.0-darwin-amd64.tar.gz
blockchain-indexer-v1.0.0-darwin-arm64.tar.gz
blockchain-indexer-v1.0.0-windows-amd64.zip
```

#### 3. Docker Release
- Builds multi-platform Docker images (amd64, arm64)
- Pushes to GitHub Container Registry (ghcr.io)
- Tags with semantic versions and `latest`

**Docker tags:**
```
ghcr.io/sage-x-project/blockchain-indexer:v1.0.0
ghcr.io/sage-x-project/blockchain-indexer:v1.0
ghcr.io/sage-x-project/blockchain-indexer:v1
ghcr.io/sage-x-project/blockchain-indexer:latest
```

#### 4. Notification
- Confirms successful release
- Lists all artifacts

### Releasing a New Version

```bash
# 1. Update version in code if needed
# 2. Commit all changes
git add .
git commit -m "chore: prepare v1.0.0 release"
git push

# 3. Create and push tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# 4. GitHub Actions will automatically:
#    - Create GitHub release
#    - Build and upload binaries
#    - Build and push Docker images
```

---

## Dependency Management

### Dependabot Configuration

Dependabot automatically:
- Checks for dependency updates weekly (Monday 9:00 UTC)
- Opens pull requests for updates
- Limits to 10 Go module PRs, 5 GitHub Actions PRs

### Update Types

1. **Go Modules**
   - Updates Go dependencies in `go.mod`
   - Excludes major version updates (breaking changes)
   - Labeled with `dependencies` and `go`

2. **GitHub Actions**
   - Updates workflow action versions
   - Ensures latest security patches
   - Labeled with `dependencies` and `github-actions`

3. **Docker**
   - Updates base images in Dockerfile
   - Keeps Alpine Linux updated
   - Labeled with `dependencies` and `docker`

### Reviewing Dependency PRs

```bash
# Check Dependabot PR
gh pr view <pr-number>

# Check for breaking changes
gh pr diff <pr-number>

# Run tests locally
gh pr checkout <pr-number>
make test

# Merge if all checks pass
gh pr merge <pr-number> --squash
```

---

## Badges and Status

### Adding Status Badges to README

```markdown
[![CI](https://github.com/sage-x-project/blockchain-indexer/actions/workflows/ci.yml/badge.svg)](https://github.com/sage-x-project/blockchain-indexer/actions/workflows/ci.yml)
[![Release](https://github.com/sage-x-project/blockchain-indexer/actions/workflows/release.yml/badge.svg)](https://github.com/sage-x-project/blockchain-indexer/actions/workflows/release.yml)
[![codecov](https://codecov.io/gh/sage-x-project/blockchain-indexer/branch/main/graph/badge.svg)](https://codecov.io/gh/sage-x-project/blockchain-indexer)
[![Go Report Card](https://goreportcard.com/badge/github.com/sage-x-project/blockchain-indexer)](https://goreportcard.com/report/github.com/sage-x-project/blockchain-indexer)
```

### Monitoring Build Status

- **GitHub Actions Tab**: View all workflow runs
- **Pull Requests**: See checks at the bottom of each PR
- **Security Tab**: Review security scan results
- **Dependabot Tab**: Manage dependency updates

---

## Troubleshooting

### CI Failures

#### Tests Failing
```bash
# Run tests locally
make test

# Run with verbose output
go test -v ./...

# Run specific test
go test -v -run TestName ./path/to/package
```

#### Coverage Below Threshold
```bash
# Check coverage locally
make test-coverage

# View coverage report
go tool cover -html=coverage.out

# Add tests for uncovered code
```

#### Linting Errors
```bash
# Run linter locally
golangci-lint run

# Auto-fix some issues
golangci-lint run --fix

# Check specific file
golangci-lint run path/to/file.go
```

#### Build Failures
```bash
# Test build locally
make build

# Cross-compile for specific platform
GOOS=linux GOARCH=amd64 go build ./cmd/indexer
```

### Release Failures

#### Tag Already Exists
```bash
# Delete local tag
git tag -d v1.0.0

# Delete remote tag
git push origin :refs/tags/v1.0.0

# Re-create tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

#### Docker Push Failed
- Check GitHub token permissions
- Verify GHCR is enabled for repository
- Check Docker build logs in Actions

---

## Best Practices

### Pull Requests
1. **Always** create feature branches from `develop`
2. **Write tests** for new features
3. **Update documentation** for changes
4. **Keep PRs small** and focused
5. **Respond to** review comments

### Commits
1. Use **conventional commits**: `feat:`, `fix:`, `chore:`, `docs:`
2. Write **clear commit messages**
3. **Reference issues** in commits: `fixes #123`

### Testing
1. **Write unit tests** for all logic
2. **Add integration tests** for APIs
3. **Run tests locally** before pushing
4. **Maintain coverage** above 70%

### Releases
1. **Follow semantic versioning**
2. **Update CHANGELOG** before release
3. **Test release candidate** before tagging
4. **Announce releases** in discussions

---

## Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Dependabot Documentation](https://docs.github.com/en/code-security/dependabot)
- [Semantic Versioning](https://semver.org/)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Go Testing](https://golang.org/pkg/testing/)
