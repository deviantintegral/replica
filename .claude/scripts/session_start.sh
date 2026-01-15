#!/usr/bin/env bash
#
# Session Start Script - Development Environment Setup
#
# This script automatically installs and configures the required development
# tools for this project. It is designed to be idempotent (safe to run multiple
# times) and targets Debian/Ubuntu-based environments.
#
# What this script installs:
#   - Go: The Go programming language (for building the project)
#   - golangci-lint: A fast Go linters runner (for code quality)
#   - pre-commit: A framework for managing git pre-commit hooks
#   - pre-commit hooks: Configured hooks for this repository
#
# Usage:
#   This script is automatically executed by Claude Code's SessionStart hook.
#   It can also be run manually: ./.claude/scripts/session_start.sh
#

set -euo pipefail

# =============================================================================
# Version Configuration (for Renovate compatibility)
# =============================================================================
GO_VERSION="1.25"
GOLANGCI_LINT_VERSION="v1.64.5"

# =============================================================================
# Color Output Helpers
# =============================================================================
info() {
    echo -e "\033[0;34m[INFO]\033[0m $*"
}

success() {
    echo -e "\033[0;32m[OK]\033[0m $*"
}

warn() {
    echo -e "\033[0;33m[WARN]\033[0m $*"
}

# =============================================================================
# PATH Configuration
# =============================================================================
export PATH="/usr/local/go/bin:$PATH"

# Add GOPATH/bin to PATH if Go is installed
if command -v go &> /dev/null; then
    export PATH="$(go env GOPATH)/bin:$PATH"
fi

# =============================================================================
# Go Installation
# =============================================================================
install_go() {
    local current_version

    # Check if Go is already installed with the correct version
    if command -v go &> /dev/null; then
        current_version=$(go version | grep -oP 'go\K[0-9]+\.[0-9]+' || echo "0.0")
        if [[ "$current_version" == "$GO_VERSION" ]]; then
            success "Go $GO_VERSION is already installed"
            return 0
        fi
        info "Go $current_version found, upgrading to $GO_VERSION..."
    else
        info "Installing Go $GO_VERSION..."
    fi

    # Determine architecture
    local arch
    case $(uname -m) in
        x86_64)  arch="amd64" ;;
        aarch64) arch="arm64" ;;
        armv7l)  arch="armv6l" ;;
        *)
            warn "Unsupported architecture: $(uname -m)"
            return 1
            ;;
    esac

    # Download and install Go
    local go_tarball="go${GO_VERSION}.linux-${arch}.tar.gz"
    local go_url="https://go.dev/dl/${go_tarball}"

    info "Downloading Go from $go_url..."
    curl -fsSL "$go_url" -o "/tmp/${go_tarball}"

    # Remove existing Go installation and extract new one
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf "/tmp/${go_tarball}"
    rm "/tmp/${go_tarball}"

    # Update PATH for this session
    export PATH="/usr/local/go/bin:$PATH"
    export PATH="$(go env GOPATH)/bin:$PATH"

    success "Go $GO_VERSION installed successfully"
}

# =============================================================================
# golangci-lint Installation
# =============================================================================
install_golangci_lint() {
    local current_version

    # Check if golangci-lint is already installed with the correct version
    if command -v golangci-lint &> /dev/null; then
        current_version=$(golangci-lint --version 2>/dev/null | grep -oP 'v[0-9]+\.[0-9]+\.[0-9]+' || echo "v0.0.0")
        if [[ "$current_version" == "$GOLANGCI_LINT_VERSION" ]]; then
            success "golangci-lint $GOLANGCI_LINT_VERSION is already installed"
            return 0
        fi
        info "golangci-lint $current_version found, upgrading to $GOLANGCI_LINT_VERSION..."
    else
        info "Installing golangci-lint $GOLANGCI_LINT_VERSION..."
    fi

    # Install using the official installer script
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
        sh -s -- -b "$(go env GOPATH)/bin" "$GOLANGCI_LINT_VERSION"

    success "golangci-lint $GOLANGCI_LINT_VERSION installed successfully"
}

# =============================================================================
# pre-commit Installation
# =============================================================================
install_precommit() {
    # Check if pre-commit is already installed
    if command -v pre-commit &> /dev/null; then
        success "pre-commit is already installed ($(pre-commit --version))"
        return 0
    fi

    info "Installing pre-commit via pipx..."

    # Ensure pipx is available
    if ! command -v pipx &> /dev/null; then
        info "Installing pipx first..."
        sudo apt-get update -qq
        sudo apt-get install -y -qq pipx
        pipx ensurepath
        export PATH="$HOME/.local/bin:$PATH"
    fi

    # Install pre-commit
    pipx install pre-commit

    success "pre-commit installed successfully"
}

# =============================================================================
# Pre-commit Hooks Setup
# =============================================================================
setup_precommit_hooks() {
    # Check if we're in a git repository
    if ! git rev-parse --git-dir &> /dev/null; then
        warn "Not in a git repository, skipping pre-commit hook installation"
        return 0
    fi

    # Check if .pre-commit-config.yaml exists
    if [[ ! -f ".pre-commit-config.yaml" ]]; then
        warn ".pre-commit-config.yaml not found, skipping hook installation"
        return 0
    fi

    # Check if hooks are already installed
    if [[ -f ".git/hooks/pre-commit" ]] && grep -q "pre-commit" ".git/hooks/pre-commit" 2>/dev/null; then
        success "pre-commit hooks are already installed"
        return 0
    fi

    info "Installing pre-commit hooks..."
    pre-commit install
    pre-commit install --hook-type commit-msg

    success "pre-commit hooks installed successfully"
}

# =============================================================================
# Main
# =============================================================================
main() {
    info "Starting development environment setup..."
    echo ""

    install_go
    install_golangci_lint
    install_precommit
    setup_precommit_hooks

    echo ""
    success "Development environment setup complete!"
}

main "$@"
