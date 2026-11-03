# Changelog

## 1.0.3

- Capitalize report badge labels and refresh the README badge example.
- Capitalize displayed analysis modes and incomplete status without changing machine-readable report fields.

## 1.0.2

- Match the original Go Report Card badge style with rounded corners, compact text-sized segments, and grade-specific colors.
- Preserve visible basic-mode labels and grey incomplete-analysis badges.
- Display a real CMake example badge and its analysis report in the README.
- Add regression tests for badge colors, analysis state and SVG escaping.
- Verify and upload the generated README example in CI.

## 1.0.1

- Prepare Marketplace release drafts after full CI validation.
- Document Marketplace publication and stable major tag promotion.
- Clarify Action metadata and add a reusable workflow example.
- Include hidden report files when uploading artifacts.
- Update CI checkout to the Node.js 24-based v6 release.

## 1.0.0

- Add local Go CLI and Ubuntu Composite Action.
- Integrate clang-format, Cppcheck and build-aware clang-tidy.
- Distinguish incomplete analysis from quality failures.
- Generate self-contained SVG, Markdown and versioned JSON reports.
- Add configuration-aware baselines, quality gates and bounded analysis workers.
- Add real-analyzer integration tests and multi-platform CLI release artifacts.
