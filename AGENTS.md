# Agent Instructions

## Project Overview

`gjq` is a CLI tool (Go port of [jsongrep](https://github.com/micahkepe/jsongrep)) for querying JSON documents using **regular path expressions**. Queries are regular expressions applied to JSON paths (field names and array indices), compiled into a DFA for efficient traversal. Reference implementation: `.ref/jsongrep/`.

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

**Not yet done:**
- `cmd/root.go` stub returns debug print instead of executing query pipeline
- Colorized JSON output formatting (`internal/output/`)
- `Query.String()` round-trip display method
- Binary search optimization for range lookups in DFA

### Phase 3: Wire Up CLI (`cmd/root.go`)

Replace stub `runRoot()` with:

1. **Resolve query** — `-F` → `Sequence([KleeneStar(Disjunction([FieldWildcard, ArrayWildcard])), Field(queryStr)])`. Else `ParseQuery(queryStr)`.
2. **Read input** — `readInput()`.
3. **Parse JSON** — `query.ParseJSON(string(data))`.
4. **Compile DFA** — `query.NewQueryDFA(&q, flagIgnoreCase)`.
5. **Execute** — `dfa.Find(json)`.
6. **Output** — `bufio.NewWriterSize(os.Stdout, 4096)`, then:
   - `--count` → bold blue "Found matches: N"
   - `--depth` → bold blue "Depth: N" (call `output.Depth(json)`)
   - not `--no-display` → iterate results, `output.WriteResult(...)`
   - Flush; silently ignore `syscall.EPIPE`.
7. **Errors** → exit code 1.

### Phase 4: Performance

- **4a.** Binary search in `IndexSymbolID()` via `sort.Search` over sorted `d.Ranges` — O(log n) per array element.
- **4b.** Pre-allocate `path` capacity; only clone on accept in `traverse()`.
- **4c.** Pass `bytes.Reader` directly to `json.Decoder` instead of `string(data)` copy.
- **4d.** (Future) `syscall.Mmap` for large files on darwin/linux.

### Phase 5: Test Coverage

- CLI: `--count`/`--no-display`, `--depth`, `-F` multiple matches, `--compact`, `-i`, new `tests/data/nested.json`
- Output: `Depth()`, `WriteResult()` with/without paths/compact/color, broken pipe
- String round-trip: `ParseQuery(s).String() == s` for all syntax, `-F` constructed query

### Implementation Order

1. Phase 1 + 2 (parallel, no dependencies)
2. Phase 3 (depends on 1 + 2)
3. Phase 4 (depends on 3)
4. Phase 5 (interleaved)
