# AI Agent Development Guide for slog

This file provides guidance to AI agents when working with code in this
repository. For developers and general project information, please refer to
[README.md](README.md) first.

## Critical Rules for AI Agents

### Never Push Directly to Main Branch

All code changes must go through pull requests. When working with this
repository:

1. Always create a feature branch for changes
2. Never use `git push origin main`
3. Always create a pull request for review
4. See [Git Workflow and Pull Requests](#git-workflow-and-pull-requests) section

## Repository Overview

`darvaza.org/slog` is a backend-agnostic structured logging interface for Go
that provides a standardised way to implement logging across different logging
backends. It offers a clean API with method chaining, multiple log levels, and
support for structured fields.

## Prerequisites

Before starting development, ensure you have:

- Go 1.23 or later installed (check with `go version`).
- `make` command available (usually pre-installed on Unix systems).
- `$GOPATH` configured correctly (typically `~/go`).
- Git configured for proper line endings.

## Common Development Commands

```bash
# Full build cycle (get deps, generate, tidy, build)
make all

# Run tests for all modules including handlers
make test

# Format code, tidy dependencies, and check grammar (run before committing)
make tidy

# Check grammar only (without formatting)
make check-grammar

# Clean build artifacts
make clean

# Update dependencies
make up

# Run go:generate directives
make generate

# Work with specific handlers
make tidy-cblog    # Tidy cblog handler
make test-filter   # Test filter handler
make build-zap     # Build zap handler
```

## Code Architecture

### Key Design Principles

- **Backend-agnostic**: Core interface doesn't depend on any specific logging
  implementation.
- **Method chaining**: All logger methods return a Logger for fluent API
  design.
- **Minimal dependencies**: Only depends on `darvaza.org/core` for base
  utilities.
- **Handler pattern**: Separate packages for different logging backends.

### Core Components

- **Logger interface** (slog.go): Defines the main Logger interface with level
  methods, field attachment, and print methods.
- **Log levels**: Six levels from Debug to Panic with clear semantics.
- **Context integration** (context.go): Type-safe context storage and retrieval
  of loggers.
- **Standard library adapter** (std.go, writer.go): Integration with Go's
  standard `log` package.

### Handler Architecture

Each handler is a separate Go module in the `handlers/` directory:

- **cblog**: Channel-based logger for receiving log entries through channels.
- **discard**: No-op logger for testing and optional logging scenarios.
- **filter**: Middleware logger for filtering and transforming log entries.
- **logrus**: Adapter for the popular logrus logging library.
- **zap**: Adapter for Uber's zap high-performance logger.
- **zerolog**: Adapter for the zerolog JSON logger.

### Code Quality Standards

The project enforces the same strict linting rules as `darvaza.org/core`:

- Max function length: 40 lines.
- Max function results: 3.
- Max arguments: 5.
- Cognitive complexity: 7.
- Cyclomatic complexity: 10.

Always run `make tidy` before committing to ensure proper formatting.

### Testing Patterns

- Test files should cover both the interface contract and implementation
  details.
- Handler tests should verify proper delegation to the underlying logger.
- Use table-driven tests for comprehensive coverage.
- Test disabled logger behavior to ensure no side effects.

## Important Notes

- The main module and each handler are separate Go modules with their own
  `go.mod` files.
- Handlers use `replace` directives during development to reference the local
  slog module - these are permanent and should not be removed.
- Fatal and Panic log levels are expected to exit/panic regardless of enabled
  state.
- Field keys must be non-empty strings; values can be any type.
- The build system automatically discovers and builds all handler modules.

## CI and Testing

### Version Selection System

The build system uses `internal/build/get_version.sh` to dynamically select
tool versions based on the Go version being used. This allows different
versions of tools like golangci-lint for different Go versions.

#### How get_version.sh Works

```bash
# Usage: get_version.sh <base_go_version> <version1> [version2] ...
# Example: $(TOOLSDIR)/get_version.sh 1.23 v1.63.4 v1.64
```

The script:

1. Detects the current Go version from `go version`
2. Compares it with the base Go version (first argument)
3. If current Go >= base version, it selects versions from the list:
   - For Go == base version: uses the first version (v1.63.4)
   - For Go > base version: increments through the list
   - Returns the last version if Go version exceeds the list

This allows the Makefile to use appropriate tool versions:

- Go 1.22: would use v1.63.4 (if base is 1.23)
- Go 1.23: uses v1.63.4 (first version after base)
- Go 1.24+: uses v1.64 (second version)

### Testing Tool Compatibility

When updating Go versions or tool versions, test compatibility using a
separate branch:

```bash
# Create test branch
git checkout -b test/ci-go-version

# Update tool version in Makefile
# Edit line: GOLANGCI_LINT_VERSION ?= \
#   $(shell $(TOOLSDIR)/get_version.sh 1.23 vX.Y.Z)

# Commit and push
git add Makefile
git commit -m "test: try golangci-lint vX.Y.Z"
git push -u origin test/ci-go-version

# Monitor CI run
gh run watch --exit-status

# Check specific job details if needed
gh run view --job=<job-id>
```

### Common Version Issues

1. **Tool built with older Go**: If golangci-lint was built with Go 1.22,
   it cannot analyze code requiring Go 1.23+.
2. **Version selection**: Ensure versions in get_version.sh calls are
   ordered correctly (older to newer).
3. **CI failures**: Check the actual Go version used by the runner, not just
   the matrix version.

For version numbering strategy when creating releases, see
[RELEASE.md Version Numbering](RELEASE.md#version-numbering).

## Working with Handlers

When developing or modifying handlers:

1. Each handler must implement the full `slog.Logger` interface.
2. All handlers should embed `internal.Loglet` for consistent field chain
   management and immutable logger behavior.
3. Handlers should properly delegate to their underlying logging library.
4. Level mapping between slog and the backend should be clearly documented.
5. Handlers should handle nil or invalid inputs gracefully.
6. Performance characteristics should match the underlying library.

### Handler Development Mode

Handlers use `replace` directives to reference the local slog module during
development:

```go
// In handlers/*/go.mod - always present
replace darvaza.org/slog => ../..
```

**IMPORTANT**: These replace directives are essential for development:

- They allow handlers to use the local slog module instead of the published
  version
- They must **remain in the repository** - do not remove them
- Go automatically ignores them when the module is imported externally
- They enable testing changes to the slog interface without publishing

### Updating Handler Dependencies

When updating slog version in handlers:

```bash
# Update all handlers to a new slog version
for handler in cblog discard filter logrus zap zerolog; do
  go -C handlers/$handler get darvaza.org/slog@v0.7.0
done

# The replace directives remain intact - this is correct behavior
```

To update all dependencies in handlers:

```bash
# Update all dependencies (use with caution)
for handler in cblog discard filter logrus zap zerolog; do
  go -C handlers/$handler get -u
done
```

### Common Development Tasks

1. **Testing a change across all handlers**:

   ```bash
   # Make your changes to slog interface
   # Then test all handlers
   make test
   ```

2. **Adding a new method to the Logger interface**:
   - Update the interface in slog.go
   - Implement the method in all handlers
   - The replace directives ensure handlers use your local changes

3. **Updating handler-specific dependencies**:

   ```bash
   # Update a specific dependency in a handler
   go -C handlers/zap get go.uber.org/zap@latest
   ```

### Common Mistakes to Avoid

1. **Do not remove replace directives** - they are needed for development
2. **Do not run `go get -u` without considering impact** - it updates all
   dependencies which may include breaking changes
3. **Always verify replace directives exist** after dependency updates
4. **Remember to test all handlers** after interface changes

## Linting and Code Quality

### Documentation Standards

When editing Markdown files, ensure compliance with:

- **Line Length**: Maximum 80 characters per line. Break long lines at
  appropriate points (after commas, before operators, at sentence boundaries).
- **LanguageTool**: Check for missing articles ("a", "an", "the"), punctuation,
  and proper hyphenation of compound modifiers.
- **Markdownlint**: Follow standard Markdown formatting rules, including:
  - Consistent heading style
  - Proper list formatting
  - Trailing newline at end of file
  - No multiple consecutive blank lines

### Common Documentation Issues to Check

1. **Line Length**: Keep lines under 80 characters.
   - ❌ Long URLs on same line with text
   - ✅ Break after colon or use reference-style links
   - For lists, indent continuation lines with 2 spaces

2. **Missing Articles**: Ensure proper use of "a", "an", and "the".
   - ❌ "provides simple standardised interface".
   - ✅ "provides a simple standardised interface".

3. **Missing Punctuation**: End all list items consistently.
   - ❌ "Handler tests should verify proper delegation".
   - ✅ "Handler tests should verify proper delegation."

4. **Compound Modifiers**: Hyphenate when used as modifiers.
   - ❌ "backend agnostic interface".
   - ✅ "backend-agnostic interface".

### Writing Documentation Guidelines

When creating or editing documentation files:

1. **File Structure**:
   - Always include a link to related documentation (e.g., AGENT.md should
     link to README.md).
   - Add prerequisites or setup instructions before diving into commands.
   - Include paths to configuration files when mentioning tools.

2. **Formatting Consistency**:
   - End all bullet points with periods for consistency.
   - Capitalize proper nouns correctly (Go, GitHub, Makefile).
   - Use consistent punctuation in examples and lists.

3. **Clarity and Context**:
   - Provide context for AI agents and developers alike.
   - Include "why" explanations, not just "what" descriptions.
   - Add examples for complex concepts or common pitfalls.

4. **Code Examples**:
   - Always include necessary import statements in code snippets.
   - Use package aliases when imports might conflict (e.g., `slogzap`).
   - Ensure examples are self-contained and would compile.
   - Include variable declarations for referenced but undefined variables.

5. **Maintenance**:
   - Update documentation when adding new handlers or changing interfaces.
   - Keep the pre-commit checklist current with project practices.
   - Review documentation changes for the issues listed above.

### Pre-commit Checklist

1. Run `make tidy` to format code and check grammar across all modules.
2. Verify all tests pass with `make test`.
3. Ensure no linting violations remain.
4. Update handler documentation if modifying handler behavior.
5. Verify handler examples still compile and run correctly.

### Grammar and Style Checking

The project now includes integrated grammar checking via LanguageTool:

```bash
# Run both formatting and grammar checks
make tidy

# Run only grammar checks (Markdown and Go files)
make check-grammar
```

LanguageTool is automatically installed via npm (using pnpx) when available.
It checks both Markdown documentation and Go source files (comments and strings).
The following rules are disabled for technical documentation compatibility:

- COMMA_PARENTHESIS_WHITESPACE (conflicts with Markdown links)
- ARROWS (used in code examples)
- EN_QUOTES (technical docs use straight quotes)
- MORFOLOGIK_RULE_EN_GB (flags technical terms)
- UPPERCASE_SENTENCE_START (conflicts with inline code)

Configuration files are located in `internal/build/`:

- `markdownlint.json` - Markdown formatting rules
- `languagetool.cfg` - Grammar checking rules for British English

## Git Workflow and Pull Requests

**IMPORTANT**: Never push directly to the main branch. All changes must go
through pull requests.

### Required Workflow

1. **Always create a feature branch**:

   ```bash
   git checkout -b feature/description
   # or
   git checkout -b chore/description
   # or
   git checkout -b docs/description
   ```

2. **Make commits on the feature branch**:

   ```bash
   # Add files individually
   git add file1.go file2.go

   # Create commit message file using Write tool to avoid shell escaping
   # Write .commit-msg-slug with content:
   # type: short description
   #
   # Longer description explaining the changes.
   # Each line under 72 characters.

   # Commit with specific files and sign-off
   git commit -s file1.go file2.go -F .commit-msg-slug

   # Clean up
   rm .commit-msg-slug
   ```

   **Commit Guidelines**:
   - Always use `-s` flag for sign-off
   - Enumerate files explicitly in commit command (avoid `-a`)
   - No AI attribution or generated-by messages
   - No heredocs in commit messages
   - Keep commit messages concise and factual

3. **Push the feature branch** (never main):

   ```bash
   git push -u origin feature/description
   ```

4. **Create a pull request**:

   ```bash
   # Create PR body file using Write tool to avoid shell escaping
   # Write .pr-body-slug with content:
   # ## Summary
   #
   # - Brief bullet points of changes
   # - No AI attribution
   #
   # ## Test plan
   #
   # - [ ] Tests pass
   # - [ ] Linting passes

   # Use concise titles following conventional commit format
   gh pr create --base main --head feature/branch \
     --title "type(scope): brief description" \
     --body-file .pr-body-slug

   # Clean up
   rm .pr-body-slug
   ```

   **PR Title Format**:
   - For handlers: `feat(logr):`, `fix(zap):`, etc.
   - For core: `feat:`, `fix:`, `docs:`, etc.
   - Keep titles under 72 characters

### Branch Naming Conventions

- `feature/` - New features or enhancements
- `fix/` - Bug fixes
- `docs/` - Documentation only changes
- `chore/` - Maintenance tasks (deps, build, etc.)
- `test/` - Test-only changes

### Git Workflow Mistakes to Avoid

1. **NEVER run `git push origin main`** - This bypasses PR review
2. **Always verify current branch** before pushing: `git branch --show-current`
3. **Create PR branch before making changes** - Don't work on main
4. **Never force push to main** - This can break the repository

### Bash Command Restrictions

When working with AI agents, follow these restrictions:

1. **No directory changes**: Never use `cd` commands - use absolute paths instead
2. **No bulk operations**: Avoid `-a` flag in git commands - enumerate files explicitly
3. **No shell escaping issues**: Use Write tool for creating files with complex content
4. **No heredocs**: Avoid `<<EOF` syntax in commands
5. **No AI attribution**: Never include "Generated by", "AI", or similar attributions

## Release Process

For information about releasing slog and its handlers, including version
numbering, release procedures, and coordinating handler releases, see
[RELEASE.md](RELEASE.md).
