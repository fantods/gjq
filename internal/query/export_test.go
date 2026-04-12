package query

// Expose internal types and fields for testing purposes only.

// QueryNFA is exported for testing purposes only.
type QueryNFA = nfa

// NFATransition is exported for testing purposes only.
type NFATransition = nfaTransition

// RangeEntry is exported for testing purposes only.
type RangeEntry = rangeEntry

// NewQueryNFA builds an NFA from a query (exported for testing).
func NewQueryNFA(q Query) *nfa {
	return newNFA(q)
}

// DFA accessors for testing.

func (d *DFA) NumStatesForTest() int          { return d.numStates }
func (d *DFA) IsAcceptingForTest() []bool      { return d.isAccepting }
func (d *DFA) KeyToIDForTest() map[string]int  { return d.keyToID }
func (d *DFA) RangesForTest() []rangeEntry     { return d.ranges }
