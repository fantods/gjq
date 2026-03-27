# gjq - Go JSON Query

`gjq` is a go-based CLI tool for querying JSON using *regular path queries*

## What is gjq?

Think of a JSON document as a labeled graph — keys and indices form the edges, and the values are the nodes. gjq gives you a compact pattern language for describing which edges to follow, borrowing the familiar building blocks of regular expressions (alternation, wildcards, repetition) and applying them to tree traversal instead of string matching.

Rather than chaining filters step-by-step (the approach jq takes), you write a single declarative pattern that describes your destination. Internally, the pattern is compiled into a deterministic finite automaton that walks the document in one pass.

## How gjq differs from jq

jq's pipeline model is expressive, but simple "find this field" queries often require boilerplate like `.. | .field? // empty`. gjq flips the model: you specify *what* to match, and the engine handles the traversal.

### Deep field lookup

```bash
# gjq — the -F flag treats the argument as a plain field name and searches the entire tree
$ curl -s https://api.nobelprize.org/v1/prize.json | gjq -F firstname | head -6
prizes.[0].laureates.[0].firstname:
"Susumu"
prizes.[0].laureates.[1].firstname:
"Richard"
prizes.[0].laureates.[2].firstname:
"Omar M."
```

```bash
# jq — recursive descent with manual null suppression
$ curl -s https://api.nobelprize.org/v1/prize.json | jq '.. | .firstname? // empty' | head -3
"Susumu"
"Richard"
"Omar M."
```

One thing to notice: gjq prints the full path to each match (e.g. `prizes.[0].laureates.[0].firstname:`), so you always know *where* a value came from. jq strips that context. (Paths are shown when output goes to a terminal; piped output omits them by default. Toggle with `--with-path` / `--no-path`.)

### Matching multiple keys

```bash
# gjq — alternation inside parentheses
$ curl -s https://api.nobelprize.org/v1/prize.json | gjq 'prizes[0].(year|category)'
prizes.[0].year:
"2025"
prizes.[0].category:
"chemistry"
```

```bash
# jq — enumerate each key separately
$ curl -s https://api.nobelprize.org/v1/prize.json | jq '.prizes[0] | .year, .category'
"2025"
"chemistry"
```

### Tallying results

```bash
# gjq
$ curl -s https://api.nobelprize.org/v1/prize.json | gjq -F firstname --count -n
Found matches: 1026
```

```bash
# jq
$ curl -s https://api.nobelprize.org/v1/prize.json | jq '[.. | .firstname? // empty] | length'
1026
```

### Formatting JSON (analogous to jq '.')

```bash
$ echo '{"name":"Ada","age":36}' | gjq ''
{
  "name": "Ada",
  "age": 36
}
```

## Installation

**TODO**

## CLI Usage

```
A JSONPath-inspired query language for JSON documents

Usage: gjq [OPTIONS] [QUERY] [FILE] [COMMAND]

Commands:
  generate  Generate additional documentation and/or completions

Arguments:
  [QUERY]  Query string (e.g., "**.name")
  [FILE]   Optional path to JSON file. If omitted, reads from STDIN

Options:
  -i, --ignore-case   Case insensitive search
      --compact       Do not pretty-print the JSON output
      --count         Display count of number of matches
      --depth         Display depth of the input document
  -n, --no-display    Do not display matched JSON values
  -F, --fixed-string  Treat the query as a literal field name and search at any depth
      --with-path     Always print the path header, even when output is piped
      --no-path       Never print the path header, even in a terminal
  -h, --help          Print help (see more with '--help')
  -V, --version       Print version
```

## Additional examples

**Pluck a field from anywhere in the structure:**

```bash
curl -s https://api.nobelprize.org/v1/prize.json | gjq -F motivation | head -4
```

**Count matches silently:**

```bash
curl -s https://api.nobelprize.org/v1/prize.json | gjq -F firstname --count -n
# Found matches: 1026
```

**Combining with standard Unix tools:**

gjq adapts its output depending on whether it's writing to a terminal or a pipe — much like ripgrep's `--heading` behavior. In a terminal you get annotated paths; through a pipe you get raw values, making it straightforward to chain into `sort`, `uniq`, `wc`, and friends.

```bash
# Values only when piped — ready for downstream processing
$ curl -s https://api.nobelprize.org/v1/prize.json | gjq -F firstname | sort | head -3
"A. Michael"
"Aage N."
"Aaron"

# Force path annotations on even when piped
$ curl -s https://api.nobelprize.org/v1/prize.json | gjq -F firstname --with-path | head -4
prizes.[0].laureates.[0].firstname:
"Susumu"
prizes.[0].laureates.[1].firstname:
"Richard"
```

## Query language reference

gjq queries are regular expressions applied to JSON paths rather than text. If you've used regex before, the operators will feel natural — they just operate on key and index segments instead of characters.

| Operator | Syntax | Meaning |
|---|---|---|
| Concatenation | `foo.bar.baz` | Follow the exact path `foo` → `bar` → `baz` |
| Alternation | `foo \| bar` | Accept either `foo` or `bar` |
| Kleene star | `**` | Zero or more field steps |
| Repetition | `foo*` | Repeat the preceding step zero or more times |
| Wildcard | `*` or `[*]` | Match any single object key or array position |
| Optional | `foo?.bar` | The `foo` step may or may not be present |
| Field literal | `foo` or `"foo bar"` | Match a specific key (quote names containing spaces) |
| Array indexing | `[0]` or `[1:3]` | Select a single index or an inclusive slice |

Operators compose freely inside parentheses. For instance, `foo.(bar|baz).qux` expands to two valid paths: `foo.bar.qux` and `foo.baz.qux`.

To descend through an arbitrary mix of objects and arrays, use `(* | [*])*` — so `(* | [*])*.foo` would locate every `foo` field no matter how deeply it's nested.

Under the hood, the query engine parses expressions into an [NFA](https://en.wikipedia.org/wiki/Nondeterministic_finite_automaton), then converts that into a [DFA](https://en.wikipedia.org/wiki/Deterministic_finite_automaton) before walking the document. See the [grammar](./src/query/grammar) directory and the [`query`](./src/query) module for implementation details.

