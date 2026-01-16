---
id: 0
group: "dev-environment"
dependencies: []
status: "completed"
created: "2026-01-15"
skills:
  - bash
  - devops
---
# Development Environment Setup

## Objective
Create a SessionStart script that automatically installs and configures required development tools for Debian/Ubuntu-based environments. This establishes the foundation for all subsequent development work.

## Skills Required
- bash: Shell scripting for installation automation
- devops: System-level tool management and configuration

## Acceptance Criteria
- [ ] `.claude/settings.json` exists and configures the SessionStart hook
- [ ] SessionStart script at `.claude/scripts/session_start.sh` executes without errors on fresh Debian/Ubuntu environment
- [ ] Go 1.25 toolchain is properly installed and accessible via `go version`
- [ ] pre-commit is installed and hooks are configured
- [ ] golangci-lint pre-commit hook runs on staged Go files
- [ ] gofmt pre-commit hook runs on staged Go files
- [ ] Conventional commit message hook validates commit messages
- [ ] Renovate config validator pre-commit hook validates `renovate.json`
- [ ] Script is idempotent (running twice produces same result)
- [ ] Script documents what it installs and why

Use your internal Todo tool to track these and keep on track.

## Technical Requirements
- Target environment: Debian/Ubuntu (apt-based)
- Go 1.25 installed via official Go binaries
- pre-commit installed via pip or pipx
- golangci-lint installed for linting
- conventional-pre-commit for commit message enforcement
- Version variables must be Renovate-compatible for automated updates

## Input Dependencies
None - this is the first task.

## Output Artifacts
- `.claude/settings.json` - Claude Code settings configuring the SessionStart hook
- `.claude/scripts/session_start.sh` - Idempotent installation script
- `.pre-commit-config.yaml` - Pre-commit hooks configuration

## Implementation Notes

<details>
<summary>Detailed Implementation Guide</summary>

### Claude Code Settings

Create `.claude/settings.json` to configure Claude Code to run the session start script:

```json
{
  "hooks": {
    "SessionStart": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": ".claude/scripts/session_start.sh"
          }
        ]
      }
    ]
  }
}
```

This configuration tells Claude Code to execute the session start script at the beginning of each session, ensuring the development environment is properly configured.

### SessionStart Script Structure

Create `.claude/scripts/session_start.sh` with the following structure:

```bash
#!/bin/bash
set -euo pipefail

# Version variables (Renovate-managed)
GO_VERSION="1.25"
GOLANGCI_LINT_VERSION="v1.64.5"

# Functions for each tool installation
install_go() {
    # Check if already installed at correct version
    # Download from https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz
    # Extract to /usr/local/go
    # Add to PATH
}

install_precommit() {
    # Install via pipx (preferred) or pip
    # pipx install pre-commit
}

install_golangci_lint() {
    # Install using official installer
    # curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin ${GOLANGCI_LINT_VERSION}
}

setup_precommit_hooks() {
    # Run pre-commit install
    # Run pre-commit install --hook-type commit-msg (for conventional commits)
}

# Main execution - call each function
main() {
    install_go
    install_precommit
    install_golangci_lint
    setup_precommit_hooks
}

main "$@"
```

### Pre-commit Configuration

Create `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: https://github.com/golangci/golangci-lint
    rev: v1.64.5
    hooks:
      - id: golangci-lint
  - repo: local
    hooks:
      - id: gofmt
        name: gofmt
        entry: gofmt -w
        language: system
        types: [go]
  - repo: https://github.com/compilerla/conventional-pre-commit
    rev: v3.4.0
    hooks:
      - id: conventional-pre-commit
        stages: [commit-msg]
  - repo: https://github.com/renovatebot/pre-commit-hooks
    rev: 39.84.0
    hooks:
      - id: renovate-config-validator
        args: ['--strict']
```

### Idempotency Requirements

Each installation function must:
1. Check if the tool is already installed at the correct version
2. Skip installation if version matches
3. Upgrade if version is outdated
4. Provide clear output about what actions were taken

### Path Configuration

Ensure the script adds necessary paths:
- `/usr/local/go/bin` for Go
- `$(go env GOPATH)/bin` for Go-installed tools
- Use `~/.profile` or `/etc/profile.d/` for persistent PATH updates

</details>
