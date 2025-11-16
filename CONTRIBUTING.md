# Contributing to GoBlog

Thank you for considering contributing to GoBlog! We welcome contributions from the community and appreciate your effort to make this project better.

## Philosophy

GoBlog follows a **welcoming but selective** approach to contributions:

- **All contributions are welcome** - We encourage everyone to participate
- **Quality over quantity** - We maintain high standards to ensure project stability
- **Clear communication** - We'll provide constructive feedback on all submissions
- **Learning-friendly** - First-time contributors are encouraged to ask questions

## Quick Start

1. Fork the repository
2. Create a feature branch (`git checkout -b feat/amazing-feature`)
3. Make your changes following our guidelines below
4. Write/update tests
5. Ensure code passes all checks
6. Commit with Conventional Commits format
7. Push to your fork
8. Open a Pull Request

## Commit Message Convention

**REQUIRED:** We enforce [Conventional Commits](https://www.conventionalcommits.org/) for versioning purposes.

### Format

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Types

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `refactor:` - Code refactoring (no functional changes)
- `test:` - Adding or updating tests
- `chore:` - Maintenance tasks (dependencies, tooling, etc.)
- `perf:` - Performance improvements
- `style:` - Code style changes (formatting, whitespace, etc.)
- `ci:` - CI/CD pipeline changes

### Examples

```bash
feat(server): add CORS support for embedded use cases

fix(parser): handle malformed frontmatter gracefully

docs(readme): update installation instructions for Go 1.23+

test(feeds): add RSS feed generation test cases

chore(deps): upgrade goldmark to v1.7.0
```

### Breaking Changes

Breaking changes are **welcome but require discussion** and will likely be denied without prior communication.

To indicate a breaking change:

```
feat(api)!: change Options struct field names

BREAKING CHANGE: Renamed EnableCache to CacheEnabled for consistency
```

**Before submitting breaking changes:**
1. Open an issue to discuss the rationale
2. Wait for maintainer feedback
3. Provide migration guide in PR description

## Testing Requirements

### For New Features

**Tests are MANDATORY.** New features will not be merged without appropriate test coverage.

- Write unit tests for new functions/methods
- Write integration tests for new workflows
- Ensure tests cover edge cases and error conditions
- Aim for >80% code coverage on new code

### For Bug Fixes

**Use a test-first approach:**

1. Write or update a test that **fails** due to the bug
2. Run the test to confirm it fails
3. Fix the bug
4. Run the test to confirm it now passes
5. Include the test in your PR

This ensures:
- The bug is reproducible
- The fix actually works
- Regression prevention

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test ./internal/gen/parser/...
```

## Code Style

**REQUIRED:** All code must pass the following checks:

### 1. gofmt

Code must be formatted with `gofmt`:

```bash
# Format all files
gofmt -w .

# Check if formatting is needed
gofmt -l .
```

### 2. go vet

Code must pass `go vet`:

```bash
go vet ./...
```

### 3. golangci-lint

Code must pass `golangci-lint`:

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run
```

### Pre-Commit Checklist

Before committing, ensure:

```bash
# Format code
gofmt -w .

# Run vet
go vet ./...

# Run linter
golangci-lint run

# Run tests
go test ./...
```

## What NOT to Contribute

### CLAUDE.md is Protected

**DO NOT modify `CLAUDE.md`** under any circumstances. This file contains development context for AI assistants and is maintained exclusively by the project owner.

Any PR that modifies `CLAUDE.md` will be automatically rejected.

### Other Hard No's

- **Dependency bloat** - Don't add dependencies without strong justification
- **Scope creep** - Keep PRs focused on a single concern
- **Uncommented complex code** - If it's not obvious, add comments
- **Ignored test failures** - Never commit with failing tests

## Contribution Value Ranking

To help prioritize your efforts, here's how we rank contribution types by value:

### Tier 1 (Highest Value)

**Bug fixes that affect production use cases**
- Crashes, data corruption, security issues
- Broken core functionality

**Critical missing features**
- Features needed for 1.0.0 milestone
- Features that block common use cases

### Tier 2 (High Value)

**Performance improvements**
- Measurable speed increases
- Reduced memory usage
- Better scalability

**Well-tested new features**
- Adds functionality without breaking existing code
- Includes comprehensive tests and docs

**Documentation improvements**
- Clarifies confusing sections
- Adds missing examples
- Fixes inaccuracies

### Tier 3 (Medium Value)

**Code refactoring**
- Improves maintainability
- Reduces complexity
- Does NOT change behavior

**Developer experience improvements**
- Better error messages
- Improved logging
- Tooling enhancements

### Tier 4 (Lower Value)

**Minor bug fixes**
- Edge cases unlikely to be hit
- Non-critical issues

**Code style improvements**
- Whitespace, formatting (if gofmt compliant)
- Variable renaming (without good reason)

### Tier 5 (Lowest Value)

**Trivial changes**
- Typo fixes in comments
- One-line formatting changes
- "Prefer X over Y" without measurable benefit

**Note:** Lower tier doesn't mean "unwelcome"! All valid contributions are appreciated. This ranking helps you understand what to prioritize if you want to make the biggest impact.

## Pull Request Process

### Before Opening a PR

1. **Search existing issues/PRs** - Avoid duplicates
2. **Discuss large changes** - Open an issue first for major features
3. **Keep it focused** - One concern per PR
4. **Write tests** - As outlined above
5. **Update docs** - If you change behavior, update README/docs

### PR Description Template

```markdown
## Description
Brief summary of changes

## Motivation
Why is this change needed?

## Type of Change
- [ ] Bug fix (non-breaking change that fixes an issue)
- [ ] New feature (non-breaking change that adds functionality)
- [ ] Breaking change (fix or feature that breaks existing functionality)
- [ ] Documentation update

## Testing
How has this been tested?

## Checklist
- [ ] Code follows project style (gofmt, go vet, golangci-lint)
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] Commits follow Conventional Commits
- [ ] No breaking changes (or discussed in issue first)
- [ ] CLAUDE.md not modified
```

### Review Process

**Response Time Expectations:**
- **Initial review**: 1-2 weeks
- **Follow-up reviews**: 3-5 days
- **Final merge decision**: 1-2 weeks after approval

**What to Expect:**
- Constructive feedback on code quality
- Requests for tests if missing
- Suggestions for improvements
- Possible requests for changes

**Maintainer Decision:**
- Merge as-is
- Merge with minor changes
- Request changes (PR stays open)
- Close with explanation (not aligned with project goals)

## Issue Reporting

### Bug Reports

Use this template:

```markdown
**Description**
Clear description of the bug

**Steps to Reproduce**
1. Step one
2. Step two
3. See error

**Expected Behavior**
What should happen?

**Actual Behavior**
What actually happens?

**Environment**
- GoBlog version:
- Go version:
- OS:

**Additional Context**
Logs, screenshots, etc.
```

### Feature Requests

Use this template:

```markdown
**Use Case**
What problem does this solve?

**Proposed Solution**
How should it work?

**Alternatives Considered**
What else did you think about?

**Additional Context**
Examples, mockups, etc.
```

## Development Setup

### Prerequisites

- Go 1.23 or later
- Git
- golangci-lint (optional but recommended)

### Clone and Build

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/GoBlog.git
cd GoBlog

# Install dependencies
go mod download

# Build binaries
go build ./cmd/GoBlogGen
go build ./cmd/GoBlogServ

# Run tests
go test ./...
```

### Project Structure

```
GoBlog/
├── cmd/               # Binary entry points
├── internal/          # Private application code
├── pkg/               # Public API/SDK
├── examples/          # Example usage
├── templates/         # Default templates
└── static/            # Default static assets
```

See `CLAUDE.md` for detailed architecture documentation (read-only).

## Code of Conduct

### Our Standards

- **Be respectful** - Treat everyone with kindness
- **Be constructive** - Focus on solutions, not blame
- **Be patient** - Remember maintainers are volunteers
- **Be open-minded** - Consider other perspectives

### Unacceptable Behavior

- Harassment, discrimination, or hate speech
- Personal attacks or insults
- Trolling or inflammatory comments
- Spam or off-topic discussions

### Enforcement

Violations may result in:
1. Warning
2. Temporary ban from participation
3. Permanent ban from the project

Report violations to: contact@harryday.xyz

## License

By contributing to GoBlog, you agree that your contributions will be licensed under the MIT License.

## Questions?

- **General questions**: Open a [GitHub Discussion](https://github.com/harrydayexe/GoBlog/discussions)
- **Bug reports**: Open a [GitHub Issue](https://github.com/harrydayexe/GoBlog/issues)
- **Security issues**: Email contact@harryday.xyz (see SECURITY.md)
- **Other inquiries**: contact@harryday.xyz

## Recognition

Contributors will be:
- Listed in release notes
- Credited in CHANGELOG.md
- Mentioned in security hall of fame (for security reports)

Thank you for helping make GoBlog better!

---

*Last updated: 2025-01-16*
