---
id: 0
group: "development-environment"
dependencies: []
status: "pending"
created: "2026-01-14"
skills:
  - bash
  - devops
---
# Development Environment Setup

## Objective
Create a SessionStart script that automatically installs and configures required development tools for Debian/Ubuntu-based environments, ensuring all developers have consistent tooling.

## Skills Required
- **bash**: Shell scripting for idempotent installation scripts
- **devops**: System-level tool configuration and environment setup

## Acceptance Criteria
- [ ] SessionStart script at `.claude/scripts/session_start.sh` executes without errors on fresh Debian/Ubuntu environment
- [ ] Go 1.25 toolchain is properly installed and accessible via `go version`
- [ ] pre-commit is installed and hooks are configured
- [ ] golangci-lint pre-commit hook runs on staged Go files
- [ ] gofmt pre-commit hook runs on staged Go files
- [ ] Conventional commit message hook validates commit messages
- [ ] Script is idempotent (running twice produces same result)
- [ ] Script documents what it installs and why

## Technical Requirements
- Target environment: Debian/Ubuntu (apt-based)
- Go 1.25 installation via official Go binaries
- pre-commit installation via pip or pipx
- golangci-lint for linting
- conventional-pre-commit for commit message validation

## Input Dependencies
None - this is the first task in the plan.

## Output Artifacts
- `.claude/scripts/session_start.sh` - Idempotent installation script
- `.pre-commit-config.yaml` - Pre-commit hooks configuration

## Implementation Notes

<details>
<summary>Detailed Implementation Guide</summary>

### 1. Create Directory Structure
```bash
mkdir -p .claude/scripts
```

### 2. Create SessionStart Script (`.claude/scripts/session_start.sh`)

The script should:
1. Check if running on Debian/Ubuntu (exit gracefully on other platforms)
2. Install Go 1.25 if not already installed or if version is older
3. Install pre-commit via pipx (preferred) or pip
4. Install golangci-lint
5. Run `pre-commit install` to set up git hooks

**Script structure:**
```bash
#!/usr/bin/env bash
set -euo pipefail

# Document what this script does
# This script sets up the development environment for the Replica project.
# It installs: Go 1.25, pre-commit, golangci-lint, and configures git hooks.

# Idempotency checks
GO_VERSION="1.25"
GOLANGCI_LINT_VERSION="v1.62.0"  # Use latest stable

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check Go version
check_go_version() {
    if command_exists go; then
        current=$(go version | grep -oP 'go\K[0-9]+\.[0-9]+')
        if [[ "$(printf '%s\n' "$GO_VERSION" "$current" | sort -V | head -n1)" == "$GO_VERSION" ]]; then
            return 0
        fi
    fi
    return 1
}

# Install Go if needed
install_go() {
    if check_go_version; then
        echo "Go $GO_VERSION or newer already installed"
        return 0
    fi
    echo "Installing Go $GO_VERSION..."
    curl -LO "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz"
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf "go${GO_VERSION}.linux-amd64.tar.gz"
    rm "go${GO_VERSION}.linux-amd64.tar.gz"
    export PATH=$PATH:/usr/local/go/bin
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
}

# Install pre-commit
install_precommit() {
    if command_exists pre-commit; then
        echo "pre-commit already installed"
        return 0
    fi
    echo "Installing pre-commit..."
    if command_exists pipx; then
        pipx install pre-commit
    else
        pip install --user pre-commit
    fi
}

# Install golangci-lint
install_golangci_lint() {
    if command_exists golangci-lint; then
        echo "golangci-lint already installed"
        return 0
    fi
    echo "Installing golangci-lint..."
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin $GOLANGCI_LINT_VERSION
}

# Main execution
main() {
    echo "=== Replica Development Environment Setup ==="
    install_go
    install_precommit
    install_golangci_lint

    # Configure pre-commit hooks if .pre-commit-config.yaml exists
    if [[ -f .pre-commit-config.yaml ]]; then
        echo "Installing pre-commit hooks..."
        pre-commit install
        pre-commit install --hook-type commit-msg
    fi

    echo "=== Setup Complete ==="
}

main "$@"
```

### 3. Create Pre-commit Configuration (`.pre-commit-config.yaml`)

```yaml
# Pre-commit hooks for Replica project
# See https://pre-commit.com for more information
repos:
  # Go linting with golangci-lint
  - repo: https://github.com/golangci/golangci-lint
    rev: v1.62.0
    hooks:
      - id: golangci-lint
        args: [--timeout=5m]

  # Go formatting
  - repo: local
    hooks:
      - id: gofmt
        name: gofmt
        entry: gofmt -w -s
        language: system
        types: [go]
        pass_filenames: true

  # Conventional commit messages
  - repo: https://github.com/compilerla/conventional-pre-commit
    rev: v3.4.0
    hooks:
      - id: conventional-pre-commit
        stages: [commit-msg]
        args: [--strict]
```

### 4. Make Script Executable
```bash
chmod +x .claude/scripts/session_start.sh
```

### 5. Testing Idempotency
Run the script twice and verify:
- No errors on second run
- No duplicate installations
- Same end state

### Notes
- The script uses `set -euo pipefail` for strict error handling
- Each installation function checks if the tool is already installed
- PATH updates are added to `.bashrc` for persistence
- pre-commit hooks are installed for both pre-commit and commit-msg stages

</details>
