---
id: 16
group: "testing"
dependencies: [15]
status: "pending"
created: "2026-01-14"
skills:
  - github-actions
  - go
---
# Coverage and Mutation Testing

## Objective
Configure code coverage reporting with 80% threshold and gremlins mutation testing with 60% threshold, both enforced in CI.

## Skills Required
- **github-actions**: CI configuration, coverage tools
- **go**: Go coverage, gremlins mutation testing tool

## Acceptance Criteria
- [ ] Code coverage reported with 80% minimum threshold
- [ ] CI fails if coverage drops below 80%
- [ ] gremlins mutation testing integrated
- [ ] Mutation score ≥60% enforced in CI
- [ ] Coverage reports uploaded to codecov (optional)

Use your internal Todo tool to track these and keep on track.

## Technical Requirements
- Go's built-in coverage tool
- gremlins for mutation testing
- Coverage threshold: 80%
- Mutation score threshold: 60%

## Input Dependencies
- Task 15: CI database matrix (coverage runs after tests)

## Output Artifacts
- Updated `.github/workflows/ci.yml` with coverage and mutation jobs
- `.gremlins.yaml` configuration
- Updated `Makefile` with coverage and mutation targets

## Implementation Notes

<details>
<summary>Detailed Implementation Instructions</summary>

### Meaningful Test Strategy Guidelines

**IMPORTANT**: When writing tests to meet coverage requirements, focus on meaningful tests that verify actual functionality.

**Definition of "Meaningful Tests":**
Tests that verify custom business logic, critical paths, and edge cases specific to the application. Focus on testing YOUR code, not the framework or library functionality.

**When TO Write Tests:**
- Custom business logic and algorithms
- Critical user workflows and data transformations
- Edge cases and error conditions for core functionality
- Integration points between different system components
- Complex validation logic or calculations

**When NOT to Write Tests:**
- Third-party library functionality (already tested upstream)
- Framework features (React hooks, Express middleware, etc.)
- Simple CRUD operations without custom logic
- Getter/setter methods or basic property access
- Configuration files or static data
- Obvious functionality that would break immediately if incorrect

1. **Install gremlins**:
   ```bash
   go install github.com/go-gremlins/gremlins/cmd/gremlins@latest
   ```

2. **Create `.gremlins.yaml`**:
   ```yaml
   # Gremlins mutation testing configuration
   timeout: 10s
   workers: 4

   # Threshold for mutation score
   threshold: 0.6

   # Directories to mutate
   include:
     - "internal/..."

   # Directories to exclude
   exclude:
     - "internal/testutil/..."
     - "**/*_test.go"
   ```

3. **Update Makefile**:
   ```makefile
   ## coverage: Run tests with coverage report
   coverage:
   	go test -race -coverprofile=coverage.out -covermode=atomic ./...
   	go tool cover -func=coverage.out
   	@echo "Checking coverage threshold..."
   	@coverage=$$(go tool cover -func=coverage.out | grep total | awk '{print substr($$3, 1, length($$3)-1)}'); \
   	if [ $$(echo "$$coverage < 80" | bc -l) -eq 1 ]; then \
   		echo "Coverage $$coverage% is below 80% threshold"; \
   		exit 1; \
   	fi

   ## mutation: Run mutation testing with gremlins
   mutation:
   	gremlins unleash --config .gremlins.yaml
   ```

4. **Update CI workflow** (`.github/workflows/ci.yml`):
   Add coverage and mutation jobs:
   ```yaml
   coverage:
     name: Coverage
     runs-on: ubuntu-latest
     needs: [test]
     steps:
       - uses: actions/checkout@v4

       - uses: actions/setup-go@v5
         with:
           go-version: ${{ env.GO_VERSION }}

       - name: Run tests with coverage
         run: go test -race -coverprofile=coverage.out -covermode=atomic ./...

       - name: Check coverage threshold
         run: |
           COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print substr($3, 1, length($3)-1)}')
           echo "Coverage: ${COVERAGE}%"
           if (( $(echo "$COVERAGE < 80" | bc -l) )); then
             echo "Coverage ${COVERAGE}% is below 80% threshold"
             exit 1
           fi

       - name: Upload coverage to Codecov
         uses: codecov/codecov-action@v4
         with:
           files: coverage.out
           fail_ci_if_error: false

   mutation:
     name: Mutation Testing
     runs-on: ubuntu-latest
     needs: [coverage]
     steps:
       - uses: actions/checkout@v4

       - uses: actions/setup-go@v5
         with:
           go-version: ${{ env.GO_VERSION }}

       - name: Install gremlins
         run: go install github.com/go-gremlins/gremlins/cmd/gremlins@latest

       - name: Run mutation testing
         run: gremlins unleash --config .gremlins.yaml
   ```

5. **Verify locally**:
   ```bash
   make coverage
   make mutation
   ```

</details>
