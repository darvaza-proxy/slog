# Release Process for darvaza.org/slog

This document describes the release process for the darvaza.org/slog
repository, including the main module and all handler modules, ensuring
consistent and coordinated releases.

## Quick Reference

### Release Order

1. **Main module**: slog (core interface)
2. **Handler modules**: cblog, discard, filter, logrus, zap, zerolog

### Essential Commands

For development commands and build instructions, see [AGENT.md Common
Development Commands](AGENT.md#common-development-commands).

Release-specific commands:

```bash
# Check current versions
git tag --list | sort -V

# Create annotated tag for main module
git tag -a v0.7.0 -m "Release message"

# Create annotated tag for handler
git tag -a handlers/zap/v0.6.0 -m "Release message"

# Push tags
git push origin v0.7.0 handlers/zap/v0.6.0
```

## Module Structure

The repository contains:

- **Main module** (`darvaza.org/slog`): Core logging interface
- **Handler modules** (in `handlers/` directory):
  - `darvaza.org/slog/handlers/cblog`: Channel-based logger
  - `darvaza.org/slog/handlers/discard`: No-op logger
  - `darvaza.org/slog/handlers/filter`: Filtering middleware
  - `darvaza.org/slog/handlers/logrus`: Logrus adapter
  - `darvaza.org/slog/handlers/zap`: `zap` adapter
  - `darvaza.org/slog/handlers/zerolog`: Zerolog adapter

## Release Process

### 1. Pre-release Checklist

Before starting the release process:

- [ ] Ensure all tests pass: `make test`
- [ ] Run linting: `make lint`
- [ ] Update dependencies: `make up && make tidy`
- [ ] Review [AGENT.md testing patterns](AGENT.md#testing-patterns) for
  comprehensive testing
- [ ] Follow [documentation standards](AGENT.md#documentation-standards) when
  writing release notes
- [ ] Review and update CHANGELOG.md if present
- [ ] Ensure documentation is up-to-date.
- [ ] Check current versions: `git tag --list | sort -V`
- [ ] Verify no uncommitted changes: `git status`

### 2. Main Module Release

1. Check the latest tag for the main module:

   ```bash
   git tag --list | grep -E "^v[0-9]" | sort -V | tail -1
   ```

2. Create an annotated tag with comprehensive release notes:

   ```bash
   git tag -a v0.7.0 -m "darvaza.org/slog v0.7.0

   Brief description of the release

   Changes since vX.Y.Z:
   - List of interface changes
   - New features
   - Bug fixes
   - Breaking changes (if any)

   Dependencies:
   - darvaza.org/core v0.17.1
   - Go 1.23 or later"
   ```

3. Push the tag:

   ```bash
   git push origin v0.7.0
   ```

4. Wait for pkg.go.dev to index the new version (usually 5-10 minutes).

### 3. Update Handler Dependencies

1. Update each handler's go.mod to use the new slog version:

   ```bash
   # Update handlers to use the new slog version
   for handler in cblog discard filter logrus zap zerolog; do
     go -C handlers/$handler get darvaza.org/slog@v0.7.0
     go -C handlers/$handler mod tidy
   done
   ```

   Alternative using make:

   ```bash
   # If supported by Makefile
   make up tidy
   ```

2. Verify replace directives are present (they should always be there):

   ```bash
   # Check that replace directives exist
   grep -r "replace darvaza.org/slog" handlers/
   # Should show: replace darvaza.org/slog => ../../ for each handler
   ```

3. Run tests to ensure compatibility:

   ```bash
   make test
   ```

4. Commit the dependency updates:

   ```bash
   git add -A
   git commit -m "build: update slog dependency to v0.7.0 in all handlers"
   ```

### 4. Handler Module Releases

1. Check current handler versions:

   ```bash
   git tag --list | grep "^handlers/" | sort -V
   ```

2. Create annotated tags for each handler that has changes:

   ```bash
   # Example for zap handler
   git tag -a handlers/zap/v0.6.0 -m "darvaza.org/slog/handlers/zap v0.6.0

   `zap` handler for slog interface

   Changes since vX.Y.Z:
   - Update to slog v0.7.0
   - Other handler-specific changes

   Dependencies:
   - darvaza.org/slog v0.7.0
   - go.uber.org/zap v1.27.0
   - Go 1.23 or later"
   ```

3. Push all handler tags:

   ```bash
   # Push individual tags
   git push origin handlers/cblog/v0.7.0
   git push origin handlers/discard/v0.6.0
   git push origin handlers/filter/v0.6.0
   git push origin handlers/logrus/v0.7.0
   git push origin handlers/zap/v0.6.0
   git push origin handlers/zerolog/v0.6.0

   # Or push all at once
   git push origin --tags
   ```

### 5. Post-release Documentation

Document the release in relevant PR comments:

```bash
# For releases from merged PRs
gh pr comment PR_NUMBER --body "## Released: darvaza.org/slog vX.Y.Z

The changes from this PR have been released:

\`\`\`bash
go get darvaza.org/slog@vX.Y.Z
\`\`\`

This release includes all changes from PR #NUMBER:
- List key changes from the PR
- Note any breaking changes
- Mention dependency updates

View on pkg.go.dev: https://pkg.go.dev/darvaza.org/slog@vX.Y.Z"
```

For full releases with handlers:

```bash
gh pr comment PR_NUMBER --body "## slog v0.7.0 Released

Main module and all handlers have been updated:

\`\`\`bash
# Core interface
go get darvaza.org/slog@v0.7.0

# Handlers
go get darvaza.org/slog/handlers/cblog@v0.7.0
go get darvaza.org/slog/handlers/discard@v0.6.0
go get darvaza.org/slog/handlers/filter@v0.6.0
go get darvaza.org/slog/handlers/logrus@v0.7.0
go get darvaza.org/slog/handlers/zap@v0.6.0
go get darvaza.org/slog/handlers/zerolog@v0.6.0
\`\`\`

All modules now require Go 1.23 or later."
```

## Version Numbering

### Main Module

The main slog module follows semantic versioning:

- **Patch version** (v0.3.x): Bug fixes, documentation updates
- **Minor version** (v0.x.0): New features, backwards-compatible changes
- **Major version** (vx.0.0): Breaking changes to the Logger interface

### Handler Modules

Each handler maintains its own version but typically follows the main module:

- Handlers are versioned independently
- Major version changes in slog require handler updates
- Handler-specific changes may warrant independent version bumps

### Common Release Scenarios

1. **Interface changes in main module**:
   - Bump minor/major version of main module
   - Update and release all handlers with dependency update
   - Document migration path for breaking changes

2. **Handler-specific bug fixes**:
   - Release only the affected handler(s)
   - No need to release other handlers or main module

3. **Updating Go version requirement**:
   - This is a breaking change requiring minor version bump
   - Update main module and all handlers
   - Document clearly in release notes

4. **Adding new handler**:
   - No version change needed for existing modules
   - New handler starts at v0.1.0

## Handler Development Mode

For detailed information about handler development mode, replace directives,
and development workflows, see [AGENT.md Handler Development
Mode](AGENT.md#handler-development-mode).

**Key Points**:

- Handlers use `replace` directives to reference the local slog module
- These directives are permanent and essential for development
- They are automatically ignored when modules are imported externally

## Common Release Workflows

### Releasing After a Merged PR

When releasing changes from a recently merged PR (common workflow):

1. **Review the merged PR**:

   ```bash
   # List recent merged PRs
   gh pr list --state merged --limit 5

   # View specific PR changes
   gh pr view PR_NUMBER
   ```

2. **Determine version bump**:
   - Breaking changes (e.g., Go version requirement): Minor version
   - New features: Minor version (patch if very minor)
   - Bug fixes only: Patch version

3. **Create and push release tag**:

   ```bash
   # Create tag with summary of PR changes
   git tag -a vX.Y.Z -m "darvaza.org/slog vX.Y.Z

   Brief description matching PR purpose

   Changes since vA.B.C:
   - Include all significant changes from the PR
   - Note breaking changes clearly
   - List dependency updates

   Dependencies:
   - darvaza.org/core vX.Y.Z
   - Go X.Y or later"

   # Push the tag
   git push origin vX.Y.Z
   ```

4. **Update the PR with release info** (as shown in Post-release Documentation).

### Main Module Only vs Full Release

- **Main module only**: When changes don't affect handler APIs or when
  handler-specific changes will be released separately
- **Full release**: When interface changes require all handlers to be updated

## Troubleshooting

### Common Issues

1. **Handler tests fail after slog update**: Ensure all handlers are updated
   to use the new slog version and `replace` directives are removed.

2. **Missing handler tags**: Handler tags must follow the format
   `handlers/name/vX.Y.Z`.

3. **Version mismatch**: Ensure handler go.mod files reference the correct
   slog version after release.

### Rollback Procedure

If issues are discovered after release:

1. Do not delete tags (they may already be cached)
2. Release a new patch version with the fix
3. Update all affected handlers if interface changes

## Automation Considerations

For future automation:

1. Script to remove `replace` directives from all handler go.mod files
2. Batch update script for handler dependencies
3. GitHub Actions workflow for coordinated releases
4. Automated compatibility testing matrix

## Latest Releases

As of July 2025:

- **slog**: v0.7.0 (Go 1.23 required)
- **handlers/cblog**: v0.7.0
- **handlers/discard**: v0.6.0
- **handlers/filter**: v0.6.0
- **handlers/logrus**: v0.7.0
- **handlers/zap**: v0.6.0
- **handlers/zerolog**: v0.6.0

All modules require Go 1.23 or later and use darvaza.org/core v0.17.1.

## See Also

- [README.md](README.md): General repository information
- [AGENT.md](AGENT.md): Development guidelines for AI agents
- Individual handler README files in `handlers/*/README.md`
