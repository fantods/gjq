# Autoresearch Ideas

## Completed optimizations
- ✅ Remove convertNumbers() pass - eliminated full tree re-walk, all numbers are float64
- ✅ Use json.Unmarshal instead of json.NewDecoder().Decode() - avoids io.Reader overhead, fewer allocations
- ✅ Optimize output path - replaced fmt.Fprint/Fprintf with io.WriteString and strconv functions
- ✅ hasUpper fast-check in FieldSymbolID - avoids strings.ToLower allocation for lowercase field names

## Tried and rejected
- ❌ Build with -ldflags="-s -w" - no improvement, slightly worse
- ❌ debug.SetGCPercent(-1) - no measurable effect for CLI workloads
- ❌ Streaming JSON via json.Decoder.Token() - 2.3x SLOWER than Unmarshal due to interface boxing
- ❌ Streaming ParseReader(file) - json.Decoder buffers internally, no savings
- ❌ Inlined DFA hot path - regression due to binary layout change
- ❌ Pre-allocate results slice - no measurable improvement
- ❌ Split FieldSymbolID into sub-methods - regression due to binary layout
- ❌ strings.EqualFold for CI matching - O(n) iteration vs O(1) hash lookup, 13% regression
- ❌ Conditional caseInsensitive compilation - regression, binary layout sensitivity

## Key insights
- JSON parsing is 83%+ of in-process time (~6ms for 1MB). Cannot optimize with stdlib.
- Process startup is ~2-3ms. Cannot reduce (Go runtime overhead).
- DFA traversal is already very fast (~0.3-0.5ms for 1000 results).
- Go binary performance is VERY sensitive to code layout changes. Even trivial code changes can cause 5-10% variance.
- Measurements must use single-binary comparisons to avoid rebuild noise.

## Deferred optimizations (high effort, uncertain payoff)
- Custom JSON lexer that integrates with DFA traversal - would avoid building full tree
- Replace cobra with flag package - would save ~1-2ms startup but big interface change
- Binary layout optimization via function ordering - Go doesn't provide control over this
