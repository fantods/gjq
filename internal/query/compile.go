package query

// Option configures DFA compilation.
type Option func(*compileConfig)

type compileConfig struct {
	caseInsensitive bool
}

// WithCaseInsensitive enables case-insensitive field matching.
func WithCaseInsensitive() Option {
	return func(c *compileConfig) {
		c.caseInsensitive = true
	}
}

// Compile compiles a Query into a DFA using the provided options.
// This is the primary entry point for query compilation.
//
// Example:
//
//	dfa := query.Compile(q, query.WithCaseInsensitive())
//	results := dfa.Find(root)
func Compile(q Query, opts ...Option) *DFA {
	var cfg compileConfig
	for _, o := range opts {
		o(&cfg)
	}
	return NewDFA(q, cfg.caseInsensitive)
}
