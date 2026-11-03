# Contributing

Use Go 1.23 or newer. Run `make test`, `gofmt -w cmd internal`, and
`scripts/integration.sh` before submitting a change. Integration tests require
clang-format, clang-tidy, Cppcheck, CMake, Make and a C++ compiler.

Keep check execution separate from scoring and rendering. Add regression tests
for parsing or status changes, especially errors that could otherwise become
false green reports. Schema or scoring changes need an explicit version bump and
migration notes. Never execute compilation database command strings as shell code.
