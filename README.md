# gjq - Go JSON Query

`gjq` is a CLI tool for querying JSON using *regular path querys*

## What is gjq?

Think of a JSON document as a labeled graph — keys and indices form the edges, and the values are the nodes. gjq gives you a compact pattern language for describing which edges to follow, borrowing the familiar building blocks of regular expressions (alternation, wildcards, repetition) and applying them to tree traversal instead of string matching.

| Pattern | Meaning |
|---|---|
| `.name` | **Deep scan** — locate "name" anywhere in the object hierarchy |
| `items[*].id` | **Array wildcard** — grab every `id` from the `items` collection |
| `(debug|trace).*` | **Branching** — collect everything nested under either key |
| `(*|[*])*.host` | **Recursive descent through mixed types** — find "host" regardless of how many objects or arrays sit above it |

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
