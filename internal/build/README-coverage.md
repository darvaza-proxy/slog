# Coverage System Documentation

This directory contains a comprehensive dual coverage system for Go monorepos
that provides both self-coverage and integration coverage perspectives. The
system is designed to be portable across different Go monorepo structures.

## Overview

The coverage system generates two types of coverage data:

- **Self-coverage**: Per-package coverage testing each package in isolation.
- **Integration coverage**: Cross-package coverage using `-coverpkg=./...`.

## Prerequisites

- Go 1.23 or later.
- A monorepo with modules defined in an index file (`.tmp/index`).
- Make build system with generated rules.
- Standard directory structure with modules as subdirectories.

## Scripts

### `make_coverage.sh`

Core coverage generation script that produces dual coverage reports for a
single module.

**Usage**: `make_coverage.sh <module_name> <module_dir> <coverage_dir>`

**Generates**:

- `coverage_${module}.prof` - Integration coverage profile.
- `coverage_${module}_self.prof` - Merged self-coverage profile.
- `coverage_${module}/` - Hierarchical per-package profiles.
- `coverage_${module}.func` - Function-level coverage reports (auto-generated).
- `coverage_${module}_self.func` - Self-coverage function reports
  (auto-generated).
- `coverage_${module}.html` - HTML coverage report (auto-generated from
  self-coverage).

**Key features**:

- Uses full import paths for individual profiles to prevent naming collisions.
- Merges per-package profiles into unified self-coverage profile.
- Consistent use of `-C` and `-o` flags for all Go commands.
- Dual coverage display showing both perspectives.
- Portable across different monorepo structures.

### `make_codecov.sh`

Generates intelligent upload script for Codecov integration.

**Usage**: `make_codecov.sh [index_file] [coverage_dir]`

**Generates**: `${coverage_dir}/codecov.sh` - Upload script for CI.

**Upload strategy**:

1. **Prefers self-coverage** when file has meaningful content (>1 line).
2. **Falls back to integration coverage** when self-coverage unavailable.
3. **Skips modules** with no meaningful coverage data.

**Features**:

- SHA256 integrity verification of Codecov CLI.
- Automatic cleanup of Codecov binary via trap handlers.
- GitHub Actions grouping for clean CI output.
- Module-specific flags for monorepo support.
- Efficient upload function for streamlined CI integration.

### `merge_coverage.sh`

Merges multiple Go coverage profiles into a single file.

**Usage**: `merge_coverage.sh <input_files...> > output.prof`

**Process**:

1. Takes header from first non-empty file.
2. Appends data lines from all input files.
3. Skips empty files gracefully.

### `gen_mk.sh` (Coverage Integration)

Generates make rules for coverage system integration.

**Generated variables**:

- `COVERAGE_INTEGRATION_FILES` - List of integration coverage files.
- `COVERAGE_SELF_FILES` - List of self-coverage files.

**Generated targets**:

- `coverage` - Generates coverage for all modules.
- `coverage-<module>` - Generate coverage for specific module.
- `merged-coverage` - Creates merged repository-wide profiles from module data.
- `codecov` - Generates merged coverage and creates upload script.
- `clean-coverage` - Remove all coverage data.

**Merging architecture**: Module coverage generation is separate from merging.
The `merged-coverage` target creates repository-wide profiles with proper
dependencies, while `codecov` combines merging with upload script generation
for CI integration.

**Integration requirements**: The Makefile generator expects:

<!-- cspell:ignore TOOLSDIR -->

- Index file format: `name:dir:module_path:dependencies`.
- Variables `COVERAGE_DIR` and `TOOLSDIR` defined in main Makefile.
- Module directories containing `go.mod` files.

## Usage Examples

### Local Development

```bash
# Generate coverage for specific module
make coverage-<module>

# Generate coverage for all modules
make coverage

# Create merged repository-wide profiles
make merged-coverage

# View dual coverage output (example)
# coverage: module/package1           94.4% (21.4%)
# coverage: module/package2           76.4% (41.3%)
# coverage: module/package3           90.6% (32.9%)

# Open HTML coverage report (self-coverage based)
open .tmp/coverage/coverage_<module>.html
```

### CI Integration

```bash
# Generate merged coverage and create upload script
make codecov

# Upload best coverage data to Codecov
./.tmp/coverage/codecov.sh
```

### Manual Profile Analysis

```bash
# Generate merged profiles
make merged-coverage

# View merged integration coverage
go tool cover -func=.tmp/coverage/coverage.out

# View merged self-coverage
go tool cover -func=.tmp/coverage/coverage_self.out

# View specific module function-level coverage
go tool cover -func=.tmp/coverage/coverage_<module>_self.prof
```

## Coverage Display Format

The system uses an intelligent dual coverage display:

- **Self-coverage with integration context**: `100.0% (22.6%)`
  - First number: Package's own test coverage.
  - Parenthetical: Contribution to module-wide integration testing.

- **Integration coverage only**: `39.0%`
  - Used when self-coverage is not meaningful (e.g., organisational packages).

## File Structure

```text
.tmp/coverage/
├── coverage.out                      # Merged integration coverage (all)
├── coverage_self.out                 # Merged self-coverage (all)
├── coverage_<module>.prof            # Integration coverage (per module)
├── coverage_<module>_self.prof       # Self-coverage (per module)
├── coverage_<module>.func            # Integration function reports
├── coverage_<module>_self.func       # Self-coverage function reports
├── coverage_<module>.html            # HTML report
├── coverage_<module>/                # Individual package profiles
│   └── <module-import-path>/
│       ├── <package1>.prof
│       ├── <package2>.prof
│       └── <package3>.prof
└── codecov.sh                        # Generated upload script
```

## Environment Variables

- `GO` - Go command (default: `go`).
- `GOTEST_FLAGS` - Additional flags for `go test`.
- `CODECOV_TOKEN` - Required for Codecov uploads (CI only).

## Integration with GitHub Actions

The system integrates seamlessly with CI workflows:

```yaml
- name: Run tests with coverage
  run: make codecov

- name: Upload coverage to Codecov
  run: ./.tmp/coverage/codecov.sh
  env:
    CODECOV_TOKEN: ${{ secrets.CODECOV_TOKEN }}
```

The upload script automatically selects the best coverage perspective for each
module and uploads with appropriate module flags for monorepo support.
