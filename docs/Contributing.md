# Contributing Guide
## Table of Contents
- [Development Environment](#development-environment)
- [Workflow](#workflow)
- [Code Style Guide](#code-style-guide)
- [License & CLA](#license--cla)

Thank you for considering a contribution!

## Development Environment

1. Install Go 1.22.
2. `go install golang.org/x/tools/cmd/goimports@latest` for formatting.
3. Clone repo and run `go test ./...` to ensure baseline passes.

## Workflow

1. Fork & create a feature branch.
2. Follow conventional commits (`feat:`, `fix:`, `docs:`).
3. Run `make test` before pushing.
4. Open a PR describing changes and linking to any issues.

## Code Style Guide

* `goimports` + `gofmt` enforced.
* Prefer small, composable functions.

## License & CLA

By submitting a patch, you agree to license your work under the MIT License.
