---
id: 1
group: "project-foundation"
dependencies: []
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

Use your internal Todo tool to track these and keep on track.

## Technical Requirements
- Go 1.25+ required
- Module path: `github.com/deviantintegral/replica`
- License: AGPL-3.0

## Input Dependencies
None - this is the foundational task.

## Output Artifacts
- `go.mod` file
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
  └── LICENSE
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

</details>
