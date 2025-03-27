# Using TokenTracker as a Private Package

TokenTracker is distributed as a private Go package and requires proper authentication setup to access it. This guide provides detailed instructions for different environments and scenarios.

## Table of Contents
- [Authentication Requirements](#authentication-requirements)
- [Local Development Setup](#local-development-setup)
- [CI/CD Integration](#cicd-integration)
- [Docker Integration](#docker-integration)
- [Versioning Guidelines](#versioning-guidelines)
- [Example Files](#example-files)
- [Troubleshooting](#troubleshooting)

## Authentication Requirements

To access TokenTracker, you'll need:
1. A GitHub account with access to the TrustSight-io organization
2. A Personal Access Token (PAT) with `read:packages` scope
3. Proper Git and Go configuration for private module access

## Local Development Setup

Setting up your local environment involves configuring Git authentication and Go module settings:

1. **Create or update your `.netrc` file**:
   
   This file stores your GitHub credentials for Git operations:
   ```
   machine github.com
   login your-github-username
   password your-personal-access-token
   ```
   
   Ensure the file has proper permissions:
   ```bash
   chmod 600 ~/.netrc
   ```

2. **Configure Go for private module access**:
   
   Tell Go to use authentication for TrustSight-io repositories:
   ```bash
   go env -w GOPRIVATE=github.com/TrustSight-io/*
   git config --global url."https://github.com/".insteadOf "https://github.com/"
   git config --global url."https://${GH_USERNAME}:${GH_TOKEN}@github.com/".insteadOf "https://github.com/"
   ```

3. **Import the package in your Go code**:
   ```go
   import (
       "github.com/TrustSight-io/tokentracker"
       "github.com/TrustSight-io/tokentracker/providers"
   )
   ```

4. **For convenient setup**, use our provided script:
   ```bash
   ./examples/setup-env.sh <github-username> <personal-access-token>
   ```

## Versioning Guidelines

Always use explicit versions in your dependencies:

```go
// In go.mod
require (
    github.com/TrustSight-io/tokentracker v1.2.3
)
```

Semantic versioning is followed:
- MAJOR version for incompatible API changes
- MINOR version for new functionality in a backward compatible manner
- PATCH version for backward compatible bug fixes

## CI/CD Integration

### GitHub Actions

For GitHub Actions workflows, add these steps:

```yaml
- name: Set up Go module authentication
  run: git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"
  env:
    GITHUB_TOKEN: ${{ secrets.GO_MODULE_TOKEN }}

- name: Configure Go private modules
  run: go env -w GOPRIVATE=github.com/TrustSight-io/*
```

### GitLab CI

For GitLab CI:

```yaml
before_script:
  - git config --global url."https://${CI_DEPLOY_USER}:${CI_DEPLOY_PASSWORD}@github.com/".insteadOf "https://github.com/"
  - go env -w GOPRIVATE=github.com/TrustSight-io/*
```

### Jenkins

For Jenkins pipelines:

```groovy
stage('Setup') {
  steps {
    sh 'git config --global url."https://${GH_TOKEN}@github.com/".insteadOf "https://github.com/"'
    sh 'go env -w GOPRIVATE=github.com/TrustSight-io/*'
  }
}
```

## Docker Integration

For Dockerfile builds:

```dockerfile
# Use build args to pass GitHub token
ARG GITHUB_TOKEN

# Install Git (required for private modules)
RUN apk add --no-cache git

# Configure Git to use token for GitHub
RUN git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"

# Set GOPRIVATE
ENV GOPRIVATE=github.com/TrustSight-io/*

# Download dependencies first (better layer caching)
COPY go.mod go.sum ./
RUN go mod download

# Build your application
COPY . .
RUN go build -o app
```

Build with:

```bash
docker build --build-arg GITHUB_TOKEN=your-github-token -t your-image .
```

## Example Files

We provide several example files to help you integrate TokenTracker:

1. **Dockerfile Example**: [examples/Dockerfile.example](../examples/Dockerfile.example)
   - Complete example for containerized applications using TokenTracker

2. **GitHub Actions Workflow**: [examples/github-workflow.example.yml](../examples/github-workflow.example.yml)
   - CI/CD workflow example for GitHub Actions

3. **Local Development Setup Script**: [examples/setup-env.sh](../examples/setup-env.sh)
   - Automated setup script for local development
   - Usage: `./examples/setup-env.sh <github-username> <personal-access-token>`

## Troubleshooting

### Common Issues and Solutions

1. **"go: module github.com/TrustSight-io/tokentracker: git ls-remote: exit status 128"**
   - **Cause**: Authentication failure when accessing the repository
   - **Solution**: 
     - Verify your PAT has the correct permissions
     - Check your `.netrc` file contains correct credentials
     - Ensure Git configuration is set up properly with `git config --global url."https://${TOKEN}@github.com/".insteadOf "https://github.com/"`

2. **"go: github.com/TrustSight-io/tokentracker@v1.2.3: no matching versions for query v1.2.3"**
   - **Cause**: Incorrect version or tag in go.mod
   - **Solution**:
     - Check available tags in the repository
     - Use `go list -m -versions github.com/TrustSight-io/tokentracker` with proper authentication
     - Contact the TokenTracker team for version information

3. **"cannot find module providing package github.com/TrustSight-io/tokentracker"**
   - **Cause**: GOPRIVATE not set up correctly
   - **Solution**:
     - Run `go env -w GOPRIVATE=github.com/TrustSight-io/*`
     - Verify with `go env GOPRIVATE`

4. **Docker build fails with authentication errors**
   - **Cause**: GitHub token not passed correctly or expired
   - **Solution**:
     - Verify token is passed correctly: `docker build --build-arg GITHUB_TOKEN=xxx`
     - Generate a new token if expired
     - Consider using Docker secrets for more secure token handling

5. **CI/CD pipeline authentication failures**
   - **Cause**: Incorrect secrets configuration
   - **Solution**:
     - Verify secrets are correctly set in your CI/CD platform
     - Ensure the token has not expired
     - Check token permissions include `read:packages`

### Authentication Debugging Tips

If you're still having issues:

1. **Test GitHub Authentication Directly**:
   ```bash
   curl -H "Authorization: token YOUR_GITHUB_TOKEN" https://api.github.com/user
   ```
   You should receive your user information if the token is valid.

2. **Try Clone with Token**:
   ```bash
   git clone https://${GITHUB_TOKEN}@github.com/TrustSight-io/tokentracker.git
   ```

3. **Check Go Environment**:
   ```bash
   go env | grep GOPRIVATE
   git config --global --get-regexp url
   ```

4. **Review Go Module Discovery Logs**:
   ```bash
   GODEBUG=netdns=go+2 go get -v github.com/TrustSight-io/tokentracker@v1.2.3
   ```

For additional support, contact the TokenTracker team or file an issue in the repository.
