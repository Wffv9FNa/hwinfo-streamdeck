# Testing & Benchmarks
## Table of Contents
- [Unit Tests](#unit-tests)
  - [Running All Tests](#unit-tests)
- [Benchmarks](#benchmarks)
  - [Example Bench Suite](#benchmarks)
- [CI Pipeline](#ci-pipeline)
  - [GitHub Actions](#ci-pipeline)
- [Mocking Strategies](#mocking-strategies-todo)

## Unit Tests

Located alongside code under `*_test.go` files. Run all tests with:

```bash
go test ./...
```

## Benchmarks

Example performance benchmarks in `examples/bench/`.

```bash
go test -bench=. ./examples/bench
```

## CI Pipeline

_Not yet configured._ Suggestions:

* GitHub Actions matrix for `go test`, `go vet`, linters.
* Automatic artifact build of `.streamDeckPlugin` on tag.

---

> **TODO**: Add mock implementations for Stream Deck SDK and shared memory.
