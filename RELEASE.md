# Release Process for darvaza.org/slog

<!-- cSpell:words GOPROXY -->

This document describes the release process for the darvaza.org/slog
repository, including the main module and all handler modules, ensuring
consistent and coordinated releases.

## Quick Reference

### Release Order

1. **Main module**: slog (core interface)
2. **Handler modules**: cblog, discard, filter, logr, logrus, zap, zerolog

### Essential Commands

For development commands and build instructions, see [AGENTS.md Common
Development Commands](AGENTS.md#common-development-commands).

Release-specific commands:

```bash
# Check current versions
git tag --list | sort -V

# Draft each tag body to a file, then create signed, annotated tags
git tag -sF .tmp/tag-v0.7.0.txt v0.7.0
git tag -sF .tmp/tag-zap-v0.6.0.txt handlers/zap/v0.6.0

# Push tags
git push origin v0.7.0 handlers/zap/v0.6.0

# Prompt the proxy to index, then publish GitHub releases
GOPROXY=https://proxy.golang.org go list -m darvaza.org/slog@v0.7.0
gh release create v0.7.0 --verify-tag --latest --notes-file .tmp/rel-v0.7.0.md
```

## Module Structure

The repository contains:

- **Main module** (`darvaza.org/slog`): Core logging interface
- **Handler modules** (in `handlers/` directory):
  - `darvaza.org/slog/handlers/cblog`: Channel-based logger
  - `darvaza.org/slog/handlers/discard`: No-op logger
  - `darvaza.org/slog/handlers/filter`: Filtering middleware
  - `darvaza.org/slog/handlers/logr`: logr adapter
  - `darvaza.org/slog/handlers/logrus`: Logrus adapter
  - `darvaza.org/slog/handlers/zap`: `zap` adapter
  - `darvaza.org/slog/handlers/zerolog`: Zerolog adapter

## Release Process

### 1. Pre-release Checklist

Before starting the release process:

- [ ] Ensure all tests pass: `make test`
- [ ] Run checks: `make tidy` (formats, tidies, and runs spell/shell checks)
- [ ] Update dependencies: `make up && make tidy`
- [ ] Review [AGENTS.md testing patterns](AGENTS.md#testing-patterns) for
  comprehensive testing
- [ ] Follow [documentation standards](AGENTS.md#documentation-standards) when
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

2. Draft the tag body to a file (for example `.tmp/tag-v0.7.0.txt`). It is
   rich release-notes prose — lead with consumer impact, then sections:

   ```text
   darvaza.org/slog v0.7.0

   Brief description of the release

   Changes since vX.Y.Z:
   - List of interface changes
   - New features
   - Bug fixes
   - Breaking changes (if any)

   Dependencies:
   - darvaza.org/core v0.19.1
   - Go 1.24 or later
   ```

   Create a signed, annotated tag from the file:

   ```bash
   git tag -sF .tmp/tag-v0.7.0.txt v0.7.0
   ```

3. Push the tag:

   ```bash
   git push origin v0.7.0
   ```

4. Prompt the module proxy to index the new version (faster than waiting for
   pkg.go.dev, which follows the proxy):

   ```bash
   GOPROXY=https://proxy.golang.org go list -m darvaza.org/slog@v0.7.0
   ```

### 3. Update Handler Dependencies

1. Update each handler's go.mod to use the new slog version. Because the
   handlers resolve slog from the module proxy, the new main-module tag
   must already be published and the proxy primed (see [Main Module
   Release](#2-main-module-release)):

   ```bash
   # Update handlers to use the new slog version
   for handler in cblog discard filter logr logrus zap zerolog; do
     go -C handlers/$handler get darvaza.org/slog@v0.7.0
     go -C handlers/$handler mod tidy
   done
   ```

   Alternative using make:

   ```bash
   # If supported by Makefile
   make up tidy
   ```

2. Run tests to ensure compatibility:

   ```bash
   make test
   ```

3. Commit the dependency updates. Stage only the handler manifests by
   explicit path — never `git add -A`:

   ```bash
   git commit -s handlers/*/go.mod handlers/*/go.sum \
     -m "chore(deps): update handlers to slog v0.7.0"
   ```

### 4. Handler Module Releases

1. Check current handler versions:

   ```bash
   git tag --list | grep "^handlers/" | sort -V
   ```

2. Draft each handler's tag body to a file (for example
   `.tmp/tag-zap-v0.6.0.txt`), then create a signed, annotated tag from it:

   ```text
   darvaza.org/slog/handlers/zap v0.6.0

   `zap` handler for slog interface

   Changes since vX.Y.Z:
   - Update to slog v0.7.0
   - Other handler-specific changes

   Dependencies:
   - darvaza.org/slog v0.7.0
   - go.uber.org/zap v1.28.0
   - Go 1.24 or later
   ```

   ```bash
   git tag -sF .tmp/tag-zap-v0.6.0.txt handlers/zap/v0.6.0
   ```

3. Push all handler tags by explicit name (avoid `--tags`, which would also
   push any unrelated local tags):

   ```bash
   git push origin \
     handlers/cblog/v0.7.0 \
     handlers/discard/v0.6.0 \
     handlers/filter/v0.6.0 \
     handlers/logrus/v0.7.0 \
     handlers/zap/v0.6.0 \
     handlers/zerolog/v0.6.0
   ```

4. Prompt the proxy to index each handler:

   ```bash
   GOPROXY=https://proxy.golang.org go list -m \
     darvaza.org/slog/handlers/zap@v0.6.0 \
     darvaza.org/slog/handlers/zerolog@v0.6.0
   ```

### 5. GitHub Releases

Publish a GitHub release for every tag. The release body is a Markdown page
read in a browser, so it carries section headings, PR cross-references
(`#NNN`), and an install snippet — it is not a copy of the plain-text tag
body.

1. Draft each release body to a file (for example `.tmp/rel-zap-v0.6.0.md`),
   then create the release from its existing signed tag:

   ```bash
   # Main module — give it the repository's "Latest" badge
   gh release create v0.7.0 --verify-tag --latest \
     --title "darvaza.org/slog v0.7.0" \
     --notes-file .tmp/rel-v0.7.0.md

   # Handlers — never "Latest"; the main module owns that badge
   gh release create handlers/zap/v0.6.0 --verify-tag --latest=false \
     --title "darvaza.org/slog/handlers/zap v0.6.0" \
     --notes-file .tmp/rel-zap-v0.6.0.md
   ```

   Use `--title` matching the tag header (`<module path> <version>`) to follow
   the existing release list, and `--verify-tag` so a typo'd tag name fails
   loudly instead of creating a dangling release.

2. `--latest` semantics: pass `--latest` only on the main `slog` release — it
   is the primary module and should own the **Latest** badge — and
   `--latest=false` on every handler so a handler never steals it.

3. Ordering on the releases page: GitHub lists releases by publish time,
   newest first. Create them in dependency order (main module first), but
   create the most noteworthy handler **last** so it lands at the top of the
   list. A dependency-only bump is unremarkable; a handler that gained a
   feature is what readers came to see.

### 6. Post-release Documentation

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

For detailed information about handler development mode and multi-module
workflows, see [AGENTS.md Handler Development
Mode](AGENTS.md#handler-development-mode).

**Key Points**:

- Each handler depends on a released slog version through `require`, so
  ordinary builds and external consumers resolve the published module.
- Local cross-module development (editing slog and a handler together)
  needs a Go workspace so the handler sees the working tree rather than
  the released version. The workspace file is developer-local and
  uncommitted:

  ```bash
  # From the repository root, once per checkout
  go work init . ./handlers/*
  ```

- A handler can only be bumped to a slog version already tagged and
  available on the proxy (see [Update Handler
  Dependencies](#3-update-handler-dependencies)).

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

3. **Tag, push, and publish the release**:

   ```bash
   # Draft the tag body to .tmp/tag-vX.Y.Z.txt (summary of PR changes),
   # then create a signed tag from the file and push it
   git tag -sF .tmp/tag-vX.Y.Z.txt vX.Y.Z
   git push origin vX.Y.Z

   # Prompt the proxy, then publish the GitHub release
   GOPROXY=https://proxy.golang.org go list -m darvaza.org/slog@vX.Y.Z
   gh release create vX.Y.Z --verify-tag --latest --notes-file .tmp/rel-vX.Y.Z.md
   ```

4. **Update the PR with release info** (see GitHub Releases and Post-release
   Documentation).

### Main Module Only vs Full Release

- **Main module only**: When changes don't affect handler APIs or when
  handler-specific changes will be released separately
- **Full release**: When interface changes require all handlers to be updated

## Troubleshooting

### Common Issues

1. **Handler tests fail after slog update**: Ensure every handler's go.mod
   references the new slog version and that the version is tagged and
   available on the proxy. For local work against an untagged slog, create
   a Go workspace (see [Handler Development
   Mode](#handler-development-mode)) so the handler builds against the
   working tree.

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

1. Batch update script for handler dependencies
2. GitHub Actions workflow for coordinated releases
3. Automated compatibility testing matrix

## Latest Releases

### As of June 2026

- **slog**: v0.9.1 (Go 1.24 or later required).
- **handlers/cblog**: v0.9.1.
- **handlers/discard**: v0.7.1.
- **handlers/filter**: v0.8.1.
- **handlers/logr**: v0.2.1.
- **handlers/logrus**: v0.8.1.
- **handlers/zap**: v0.9.0.
- **handlers/zerolog**: v0.7.1.

All modules require Go 1.24 or later and use darvaza.org/core v0.19.1.

## See Also

- [README.md](README.md): General repository information
- [AGENTS.md](AGENTS.md): Development guidelines for AI agents
- Individual handler README files in `handlers/*/README.md`
