# C++ Report Card

[![C++ Report Card example](.github/cppreportcard-example/badge.svg)](.github/cppreportcard-example/report.md)
[![CI](https://github.com/soulteary/cppreportcard-action/actions/workflows/ci.yml/badge.svg)](https://github.com/soulteary/cppreportcard-action/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/soulteary/cppreportcard-action)](https://github.com/soulteary/cppreportcard-action/releases)

A local-first C/C++ quality report: **Go CLI + GitHub Action**, powered by
clang-format, Cppcheck and clang-tidy. Generates JSON, Markdown and a self-contained
SVG badge. No hosted analysis service or badge service is required.

The C++ badge above is the actual full-analysis result for the bundled
[CMake example](testdata/cmake), not a C++ rating of this Go implementation.
Click it to see the report. Regenerate it with `scripts/update-badge-example.sh`.

## GitHub Action

A copyable workflow is available in [examples/cppreportcard.yml](examples/cppreportcard.yml).

Start with a Linux CMake project. Install your project dependencies, configure it
and build any targets that generate headers **before** analysis:

```yaml
name: C++ quality
on: [push, pull_request, workflow_dispatch]
permissions:
  contents: read
jobs:
  quality:
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v6
      - run: |
          cmake -S . -B build -DCMAKE_EXPORT_COMPILE_COMMANDS=ON
          cmake --build build
      - id: report
        uses: soulteary/cppreportcard-action@v1
        with:
          mode: full
          compile_commands: build/compile_commands.json
      - uses: actions/upload-artifact@v4
        if: always()
        with:
          name: cppreportcard
          path: .cppreportcard/
          include-hidden-files: true
```

`v1` is the moving stable major tag; pin a full commit SHA for reproducible CI.
This Action builds its small Go CLI from the checked-out Action source. It installs
LLVM 18 analyzers and Cppcheck from Ubuntu packages by default. Cppcheck package
updates can change findings; provision pinned tools yourself and set
`install_tools: 'false'` when exact reproducibility is required.

The Action writes a Step Summary and report files before enforcing quality gates.
It does not commit files by default. Link a committed badge with:

```markdown
[![C++ Report Card](.cppreportcard/badge.svg)](.cppreportcard/report.md)
```

To opt into report commits, set `commit: 'true'` and `permissions: contents: write`.
Badges use the same compact, rounded style and grade colors as
[Go Report Card](https://github.com/soulteary/goreportcard-action). Basic-mode
badges visibly include `(basic)`; incomplete analysis is grey and shows
`incomplete` instead of a grade.

Commits only run for branch `push` events, never for pull requests. Use a dedicated
workflow or path filters to avoid unnecessary report-only runs.

### Inputs

- `directory`: project root relative to the workspace; default `.`.
- `mode`: `basic` (default) or `full`.
- `compile_commands`: compilation database relative to project root; default `build/compile_commands.json`.
- `config`: optional JSON config path relative to project root.
- `output_dir`: report directory relative to project root; default `.cppreportcard`.
- `baseline`: previous complete `report.json` relative to project root.
- `fail_on`: `none`, `errors`, `new-errors` (default), or `findings`.
- `threshold`: minimum absolute score in 0–100; default `0` disables this gate.
- `install_tools`: default `true`; automatic installation supports Ubuntu only.
- `llvm_version`: Ubuntu LLVM package major; default `18`.
- `commit`: default `false`.

Action inputs for mode and compilation database override config values. Use the
CLI directly if you want these values to come only from your config file.

Outputs: `status`, `grade`, `score`, `analysis_complete`, `new_errors`, `report`,
`json`, `badge`. File outputs are absolute paths. Incomplete analysis has an empty
score output. Invalid invocation can fail before any outputs exist.

## Local CLI

Requires Go 1.23+ to build. The analyzers are separate dependencies; no Go runtime
is needed when using a release binary.

```sh
go install github.com/soulteary/cppreportcard-action/cmd/cppreportcard-cli@latest
cppreportcard-cli --mode basic
cppreportcard-cli --mode full --compile-commands build/compile_commands.json
cppreportcard-cli --config .cppreportcard.json --json
```

Run `cppreportcard-cli --help` for all flags. `--json` changes stdout, **not exit
codes**. Every completed analysis writes `report.json`, `report.md`, `badge.svg`
to `--output-dir` (default `.cppreportcard`, relative to the project root).

Exit codes:

- `0`: complete analysis and configured gates pass.
- `1`: complete analysis but a quality gate fails.
- `2`: incomplete analysis, invalid configuration, or an operational error.

`--fail-on none` disables finding-based gates, not operational failures or an
explicit nonzero `--threshold`.

## Analysis modes

**basic:** clang-format + Cppcheck without a compilation database. The badge says
`C++ basic`; results do not imply full build-aware analysis. Missing configuration
reported by Cppcheck can still make analysis incomplete.

**full:** clang-format + Cppcheck + clang-tidy using a compilation database.
Compilation directories and files must exist on the analysis machine. Generate
and consume the database in the same environment; provide all generated headers,
include paths, macros and SDKs. Every discovered project source file must appear
in the database or be explicitly excluded. Multiple configurations of a source
file are preserved and analyzed.

No automatic configure/build command is executed. CMake, Meson, Make with Bear,
and other build systems can supply the same JSON database. CMake's export option
supports Makefile and Ninja generators. A header-only library needs a small
consumer/test translation unit in full mode.

Missing tools, timeouts, malformed output, invalid compile databases, and compiler
errors produce **incomplete**, never A+. Checks expose `passed`, `findings`,
`skipped`, or `error`, plus analyzed/total unit counts. Units differ by checker:
format counts files, full static analysis counts compilation entries.

## Configuration

Copy [.cppreportcard.example.json](.cppreportcard.example.json) to your project.
JSON is strict: unknown fields and multiple JSON objects are rejected. CLI mode
and database flags override config. The defaults are:

```json
{
  "mode": "basic",
  "compile_commands": "build/compile_commands.json",
  "exclude": [".git", "build", "dist", "vendor", "third_party", "external", "generated"],
  "tidy_checks": "-*,clang-analyzer-*,bugprone-*",
  "format_style": "LLVM",
  "timeout_seconds": 120,
  "jobs": 2,
  "tools": {
    "clang-format": "clang-format",
    "clang-tidy": "clang-tidy",
    "cppcheck": "cppcheck"
  }
}
```

`exclude` replaces the default list. Entries are literal directory/file names or
project-relative path prefixes, **not globs**. Symlinks are not traversed.
clang-format honors the nearest `.clang-format`; `format_style` is its fallback.
clang-tidy uses repository configuration, with the configured `tidy_checks`
override. Neither formatter nor analyzers are invoked in fix mode.

Timeout applies per tool invocation; `jobs` bounds parallel clang-tidy work and
Cppcheck workers. Tool paths can name versioned executables or absolute paths.
The initial Action supports Ubuntu; the CLI is cross-compiled for Linux, macOS
and Windows, but analyzer compatibility on other platforms must be validated by
the consumer. Compilation databases using platform-specific compiler flags may
need a dedicated Clang-compatible build configuration.

## Scoring and baselines

The JSON records schema version, scoring version, mode, configuration fingerprint,
tool versions, diagnostics and analysis status. Scoring v1 uses the **lowest
clean-file percentage across completed checks**; all discovered C/C++ files form
the denominator. Clang Static Analyzer rules (`clang-analyzer-*`) are normalized to error severity;
other clang-tidy warnings remain warnings. Any error-severity finding caps the score at 59. Grades: A+ ≥95,
A ≥90, B ≥80, C ≥70, D ≥60, F <60. Incomplete runs have no numeric score.

This score is a project summary, not a safety certification or a reliable ranking
between repositories. File layout and rule selection affect it. Use issue-based
gates for serious regressions; do not rely on the score alone.

```sh
# Save the report from your trusted base branch as baseline.json, then:
cppreportcard-cli --mode full --baseline baseline.json --fail-on new-errors
```

A baseline must be complete and match the schema, scoring version, mode and
configuration fingerprint. It never changes the absolute grade. Diagnostics are
deduplicated per tool, rule, file, location and message. Baseline fingerprints omit
line numbers so inserted comments do not recreate existing findings. Consequently,
identical messages from the same rule in the same file share a baseline identity;
new duplicate occurrences may be suppressed. Renames or message changes count as
new findings. Compare baselines with the same analyzer versions; version changes
can alter diagnostics. Cross-tool findings preserve their source and are not
merged into a single claim.

## Development

```sh
make test            # unit tests, race detector and vet
make build
make integration     # requires analyzers, CMake, Make and a C++ compiler
```

CI checks unit tests, formatting, shell syntax, real-tool clean/defective/missing-
header cases, and local Action execution. The integration test proves baseline
suppression does not hide analysis failures.

Tag `vX.Y.Z` to test and build six CLI archives (Linux/macOS/Windows,
amd64/arm64) and SHA-256 checksums. The release workflow runs the full CI suite, then attaches these assets to a
**draft release**. Publish the draft through GitHub to opt into Marketplace
listing. Update `v1` deliberately after publishing a stable release. Analyzer binaries are
not bundled; provision them separately.

## Scope

No hosted service, automatic dependency installation for the analyzed project,
coverage collection, SARIF upload or automatic fixes. Tests and coverage can be
added as report importers without changing the core diagnostic model.

Apache-2.0; see [LICENSE](LICENSE).

## Publishing to GitHub Marketplace

The root `action.yml` is the single Marketplace entry point. The listing uses its
name, description, icon and color; this README documents the Action's inputs,
outputs, permissions and Ubuntu requirements.

See [the maintainer publishing guide](.github/MARKETPLACE.md) for the draft release
flow, Marketplace categories, and stable major tag promotion. Creating a GitHub
Release through `gh` alone does not opt an Action into Marketplace.

## Unit-test coverage

Run `scripts/test-coverage.sh` to execute the Go unit tests with the race detector
and coverage instrumentation across all production Go packages, including the
CLI. CI requires at least **95% total statement coverage**. The gate uses exact
statement counts from `coverage.out`, not a rounded percentage, and does not
exclude production files. Analyzer protocol tests use deterministic local
processes; real-analyzer integration tests remain part of CI.
