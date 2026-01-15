---
id: 16
group: "testing"
dependencies: [15]
status: "pending"
created: "2026-01-15"
skills:
  - github-actions
  - go
---
# Coverage and Mutation Testing

## Objective
Configure code coverage reporting with 80% threshold and gremlins mutation testing with 60% threshold. This ensures code quality through comprehensive testing metrics.

## Skills Required
- github-actions: CI configuration for testing tools
- go: Go test coverage and mutation testing tools

## Acceptance Criteria
- [ ] Code coverage reports are generated and uploaded
- [ ] CI fails if coverage drops below 80%
- [ ] gremlins mutation testing is configured
- [ ] CI fails if mutation score drops below 60%
- [ ] Coverage badge is available for README

Use your internal Todo tool to track these and keep on track.

## Technical Requirements
- Go test coverage with -coverprofile
- gremlins for mutation testing
- Coverage threshold enforcement in CI
- Mutation score threshold enforcement in CI

## Input Dependencies
- Task 16: CI database matrix with all tests passing

## Output Artifacts
- Updated `.github/workflows/ci.yml` with coverage thresholds
- `.gremlins.yaml` - Mutation testing configuration
- Makefile targets for local coverage/mutation testing

## Implementation Notes

<details>
<summary>Detailed Implementation Guide</summary>

**Meaningful Test Strategy Guidelines**

Your critical mantra for test generation is: "write a few tests, mostly integration".

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

### gremlins Configuration

Create `.gremlins.yaml`:

```yaml
# gremlins mutation testing configuration

# Timeout for each mutant test run
timeout: 30s

# Packages to mutate (default: ./...)
packages:
  - ./internal/...

# Exclude test files and generated code
exclude:
  - ".*_test\\.go$"
  - ".*/testutil/.*"
  - ".*/migrations/.*"

# Mutators to apply
mutators:
  - CONDITIONALS_BOUNDARY
  - CONDITIONALS_NEGATION
  - INCREMENT_DECREMENT
  - INVERT_NEGATIVES
  - ARITHMETIC_BASE

# Minimum mutation score threshold (0.0 - 1.0)
threshold: 0.6

# Output format
output: text
```

### Updated CI Workflow

Add coverage and mutation testing jobs to `.github/workflows/ci.yml`:

```yaml
  coverage:
    name: Coverage Check
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'

      - name: Install SQLite dependencies
        run: sudo apt-get update && sudo apt-get install -y libsqlite3-dev

      - name: Run tests with coverage
        run: go test -race -coverprofile=coverage.out -covermode=atomic ./...

      - name: Check coverage threshold
        run: |
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          echo "Coverage: ${COVERAGE}%"
          if (( $(echo "$COVERAGE < 80" | bc -l) )); then
            echo "Coverage ${COVERAGE}% is below threshold of 80%"
            exit 1
          fi

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v4
        with:
          files: coverage.out
          fail_ci_if_error: false

  mutation:
    name: Mutation Testing
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'

      - name: Install SQLite dependencies
        run: sudo apt-get update && sudo apt-get install -y libsqlite3-dev

      - name: Install gremlins
        run: go install github.com/go-gremlins/gremlins/cmd/gremlins@latest

      - name: Run mutation testing
        run: |
          gremlins unleash --threshold 0.6
```

### Makefile Targets

Add to `Makefile`:

```makefile
## coverage: Run tests with coverage report
coverage:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out
	@COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "Total coverage: $${COVERAGE}%"; \
	if [ $$(echo "$${COVERAGE} < 80" | bc -l) -eq 1 ]; then \
		echo "Coverage $${COVERAGE}% is below threshold of 80%"; \
		exit 1; \
	fi

## coverage-html: Generate HTML coverage report
coverage-html: coverage
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## mutation: Run mutation testing
mutation:
	gremlins unleash --threshold 0.6

## mutation-html: Run mutation testing with HTML report
mutation-html:
	gremlins unleash --threshold 0.6 --output html
```

### Coverage Badge

Add to README.md (when repository is public):

```markdown
[![Coverage](https://codecov.io/gh/deviantintegral/replica/branch/main/graph/badge.svg)](https://codecov.io/gh/deviantintegral/replica)
```

### Local Testing

```bash
# Run coverage check
make coverage

# Generate HTML report
make coverage-html

# Run mutation testing (requires gremlins installed)
go install github.com/go-gremlins/gremlins/cmd/gremlins@latest
make mutation
```

### Verification

1. Run `make coverage` locally to verify threshold check
2. Run `make mutation` locally to verify mutation testing
3. Push changes and verify CI jobs pass
4. Confirm coverage and mutation scores meet thresholds

</details>
