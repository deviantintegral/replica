---
id: 1
group: "project-foundation"
dependencies: [0]
status: "pending"
created: "2026-01-14"
skills:
  - go
  - project-setup
---
# Go Module and Project Structure

## Objective
Initialize the Go module and establish the directory structure following Go conventions for the Replica project.

## Skills Required
- **go**: Go module initialization, dependency management
- **project-setup**: Creating standard Go project layouts

## Acceptance Criteria
- [ ] Go module initialized with path `github.com/deviantintegral/replica` and Go 1.25+
- [ ] Directory structure created per project conventions
- [ ] LICENSE file (AGPL-3.0) added
- [ ] Basic `.gitignore` for Go projects
- [ ] `go.mod` and `go.sum` committed
- [ ] `renovate.json` configured with auto-merge after 3 days and custom managers
- [ ] `renovate.json` validates with `npx renovate-config-validator --strict`

Use your internal Todo tool to track these and keep on track.

## Technical Requirements
- Go 1.25+ required
- Module path: `github.com/deviantintegral/replica`
- License: AGPL-3.0

## Input Dependencies
- Task 00: Development Environment Setup (Go toolchain must be installed)

## Output Artifacts
- `go.mod` file
- `renovate.json` - Renovate configuration with auto-merge and custom managers
- Directory structure:
  ```
  replica/
  ├── cmd/
  │   └── replica/
  ├── internal/
  │   ├── config/
  │   ├── database/
  │   │   ├── migrations/
  │   │   ├── sqlite/
  │   │   ├── mysql/
  │   │   └── postgres/
  │   └── testutil/
  ├── LICENSE
  └── renovate.json
  ```

## Implementation Notes

<details>
<summary>Detailed Implementation Instructions</summary>

1. **Initialize Go Module**:
   ```bash
   go mod init github.com/deviantintegral/replica
   ```
   Ensure `go.mod` specifies `go 1.25` (or later).

2. **Create Directory Structure**:
   ```bash
   mkdir -p cmd/replica
   mkdir -p internal/config
   mkdir -p internal/database/migrations
   mkdir -p internal/database/sqlite
   mkdir -p internal/database/mysql
   mkdir -p internal/database/postgres
   mkdir -p internal/testutil
   ```

3. **Add LICENSE File**:
   Create `LICENSE` with AGPL-3.0 text. Use the standard AGPL-3.0 license text from https://www.gnu.org/licenses/agpl-3.0.txt

4. **Add .gitignore**:
   Create `.gitignore` with standard Go entries:
   ```
   # Binaries
   *.exe
   *.exe~
   *.dll
   *.so
   *.dylib
   replica

   # Test binary
   *.test

   # Output of go coverage
   *.out
   coverage.html

   # Dependency directories
   vendor/

   # IDE
   .idea/
   .vscode/
   *.swp

   # OS files
   .DS_Store

   # Local config
   replica.yaml
   *.db
   ```

5. **Create placeholder main.go**:
   Create `cmd/replica/main.go` with minimal content:
   ```go
   package main

   func main() {
       // Placeholder - will be implemented with Cobra CLI
   }
   ```

6. **Create Renovate Configuration** (`renovate.json`):
   Create `renovate.json` with auto-merge after 3 days and custom managers for bash/curl dependencies:
   ```json
   {
     "$schema": "https://docs.renovatebot.com/renovate-schema.json",
     "extends": ["config:recommended"],
     "schedule": ["before 9am on monday"],
     "packageRules": [
       {
         "matchUpdateTypes": ["minor", "patch"],
         "automerge": true,
         "minimumReleaseAge": "3 days"
       }
     ],
     "customManagers": [
       {
         "customType": "regex",
         "fileMatch": ["^\\.claude/scripts/session_start\\.sh$"],
         "matchStrings": [
           "GO_VERSION=\"(?<currentValue>\\d+\\.\\d+(\\.\\d+)?)\""
         ],
         "depNameTemplate": "golang",
         "datasourceTemplate": "golang-version"
       },
       {
         "customType": "regex",
         "fileMatch": ["^\\.claude/scripts/session_start\\.sh$"],
         "matchStrings": [
           "GOLANGCI_LINT_VERSION=\"(?<currentValue>v\\d+\\.\\d+\\.\\d+)\""
         ],
         "depNameTemplate": "golangci/golangci-lint",
         "datasourceTemplate": "github-releases"
       },
       {
         "customType": "regex",
         "fileMatch": ["^Dockerfile$"],
         "matchStrings": [
           "FROM\\s+(?<depName>[^:]+):(?<currentValue>[^\\s]+)"
         ],
         "datasourceTemplate": "docker"
       },
       {
         "customType": "regex",
         "fileMatch": ["^docker-compose\\.ya?ml$"],
         "matchStrings": [
           "image:\\s*(?<depName>[^:]+):(?<currentValue>[^\\s]+)"
         ],
         "datasourceTemplate": "docker"
       }
     ]
   }
   ```

7. **Validate Renovate Configuration**:
   ```bash
   npx --yes --package renovate -- renovate-config-validator --strict
   ```

</details>
