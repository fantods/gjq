# Agent Instructions

## Project Overview

`gjq` is a CLI tool for querying JSON documents using **regular path expressions**. Queries are regular expressions applied to JSON paths (field names and array indices), compiled into a DFA for efficient traversal. Reference implementation: `.ref/jsongrep/`.

## Technology Stack

- **Go** 1.21+, module `github.com/fantods/gjq`
- **CLI**: [spf13/cobra](https://github.com/spf13/cobra)
- **No other external dependencies** — ANSI codes written directly, JSON parsing via stdlib

## Commands

```bash
go build ./...          # build
go test ./...           # run tests
go fmt ./...            # format
go run main.go          # run
```

Run `go test ./...` before committing.

## Code Style

- Handle errors explicitly — never ignore with `_`.
- No comments unless requested.
- Follow existing patterns in neighboring files.

## Directory Structure

```
./
├── cmd/
│   ├── root.go            # CLI entry point, flags, query pipeline
│   └── version.go
├── internal/
│   ├── output/            # (planned) Colorized JSON output
│   └── query/
│       ├── ast.go         # Query AST node types and constructors
│       ├── common.go      # PathType, JSONPointer, TransitionLabel
│       ├── dfa.go         # DFA construction + traversal (Find)
│       ├── nfa.go         # Glushkov NFA construction
│       └── parser.go      # Recursive descent query parser
├── tests/
│   ├── cli_test.go        # CLI integration tests
│   └── data/              # Test JSON fixtures
├── main.go
└── .ref/jsongrep/         # Rust reference implementation
```

---

## Implementation Plan

### Current Status

**Completed:**
- Query AST — all node types with constructors and `Depth()`
- Recursive descent parser — full grammar (fields, indices, ranges, wildcards, regex, groups, disjunction, sequence, optional, Kleene star, quoted fields)
- Glushkov NFA construction — linearization, first/last/follows sets, epsilon-free NFA
- DFA via subset construction — symbol extraction, disjoint range finalization, determinization, `Find()` traversal
- CLI stub — all flags defined, input reading (file/stdin), `--with-path`/`--no-path` resolution
- `ParseJSON()` with `json.Number` → `int`/`float64` conversion
- Case-insensitive DFA compilation and traversal
- Unit tests for parser, NFA, DFA; CLI integration tests

### Phase 5: Test Coverage

- CLI: `--count`/`--no-display`, `--depth`, `-F` multiple matches, `--compact`, `-i`, new `tests/data/nested.json`
- Output: `Depth()`, `WriteResult()` with/without paths/compact/color, broken pipe
- String round-trip: `ParseQuery(s).String() == s` for all syntax, `-F` constructed query
