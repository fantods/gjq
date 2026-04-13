package query

import (
	"math"
	"math/bits"
	"sort"
	"strings"
)

type rangeEntry struct {
	Start    int
	End      int
	SymbolID int
}

// DFA is a deterministic finite automaton compiled from a query.
// It matches paths through JSON documents.
type DFA struct {
	numStates       int
	startState      int
	isAccepting     []bool
	transitions     [][]int
	alphabet        []TransitionLabel
	keyToID         map[string]int
	ranges          []rangeEntry
	caseInsensitive bool
}

// NewDFA compiles a Query into a DFA. The caseInsensitive flag
// controls whether field matching ignores case.
func NewDFA(q Query, caseInsensitive bool) *DFA {
	if seq, ok := q.(SeqExpr); ok && len(seq.Steps) == 0 {
		return &DFA{
			numStates:       1,
			startState:      0,
			isAccepting:     []bool{true},
			transitions:     nil,
			alphabet:        nil,
			keyToID:         map[string]int{},
			ranges:          nil,
			caseInsensitive: caseInsensitive,
		}
	}

	b := &dfaBuilder{
		alphabet:        []TransitionLabel{{Kind: LabelOther}},
		keyToID:         map[string]int{},
		caseInsensitive: caseInsensitive,
	}

	b.extractSymbols(q)
	b.finalizeRanges()

	n := newNFA(q)
	return b.determinize(n)
}

// FieldSymbolID returns the alphabet symbol ID for a field name.
// Returns 0 (the "other" symbol) if the field is not in the query.
func (d *DFA) FieldSymbolID(field string) int {
	if d.caseInsensitive && hasUpper(field) {
		if id, ok := d.keyToID[strings.ToLower(field)]; ok {
			return id
		}
		return 0
	}
	if id, ok := d.keyToID[field]; ok {
		return id
	}
	return 0
}

func hasUpper(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 'A' && s[i] <= 'Z' {
			return true
		}
	}
	return false
}

// IndexSymbolID returns the alphabet symbol ID for an array index.
// Returns (0, false) if the index doesn't match any range in the query.
func (d *DFA) IndexSymbolID(index int) (int, bool) {
	i := sort.Search(len(d.ranges), func(i int) bool {
		return d.ranges[i].Start > index
	})
	if i == 0 {
		return 0, false
	}
	re := d.ranges[i-1]
	if index < re.End {
		return re.SymbolID, true
	}
	return 0, false
}

// Transition returns the next state for a given (state, symbol) pair.
// Returns (0, false) if no transition exists.
func (d *DFA) Transition(state, symbolID int) (int, bool) {
	if state >= d.numStates || symbolID >= len(d.alphabet) {
		return 0, false
	}
	next := d.transitions[state][symbolID]
	if next == -1 {
		return 0, false
	}
	return next, true
}

// IsAcceptingState returns whether the given state is an accepting state.
func (d *DFA) IsAcceptingState(state int) bool {
	return state < d.numStates && d.isAccepting[state]
}

// NumStates returns the number of states in the DFA.
func (d *DFA) NumStates() int {
	return d.numStates
}

// StartState returns the start state of the DFA.
func (d *DFA) StartState() int {
	return d.startState
}

// Find traverses a JSON document and returns all values whose paths
// match the query.
func (d *DFA) Find(root interface{}) []JSONPointer {
	var results []JSONPointer
	path := make([]PathType, 0, 16)
	d.traverse(d.startState, path, root, &results)
	return results
}

func (d *DFA) traverse(state int, path []PathType, value interface{}, results *[]JSONPointer) {
	if d.IsAcceptingState(state) {
		p := make([]PathType, len(path))
		copy(p, path)
		*results = append(*results, JSONPointer{Path: p, Value: value})
	}

	switch v := value.(type) {
	case map[string]interface{}:
		for key, val := range v {
			symID := d.FieldSymbolID(key)
			next, ok := d.Transition(state, symID)
			if ok {
				path = append(path, PathType{Kind: PathField, Field: key})
				d.traverse(next, path, val, results)
				path = path[:len(path)-1]
			}
		}
	case []interface{}:
		for idx, val := range v {
			symID, ok := d.IndexSymbolID(idx)
			if !ok {
				continue
			}
			next, ok := d.Transition(state, symID)
			if ok {
				path = append(path, PathType{Kind: PathIndex, Index: idx})
				d.traverse(next, path, val, results)
				path = path[:len(path)-1]
			}
		}
	}
}

type dfaBuilder struct {
	alphabet        []TransitionLabel
	keyToID         map[string]int
	collectedRanges [][2]int
	ranges          []rangeEntry
	caseInsensitive bool
}

func (b *dfaBuilder) extractSymbols(q Query) {
	switch v := q.(type) {
	case FieldExpr:
		normalized := v.Name
		if b.caseInsensitive {
			normalized = strings.ToLower(v.Name)
		}
		if _, exists := b.keyToID[normalized]; !exists {
			id := len(b.alphabet)
			b.alphabet = append(b.alphabet, TransitionLabel{Kind: LabelField, Field: normalized})
			b.keyToID[normalized] = id
		}
	case WildcardExpr:
		// Wildcard matches any field — no specific symbol needed
	case IndexExpr:
		b.collectedRanges = append(b.collectedRanges, [2]int{v.Index, v.Index + 1})
	case RangeExpr:
		b.collectedRanges = append(b.collectedRanges, [2]int{v.Start, v.End})
	case RangeFromExpr:
		b.collectedRanges = append(b.collectedRanges, [2]int{v.Start, math.MaxInt})
	case ArrayWildExpr:
		b.collectedRanges = append(b.collectedRanges, [2]int{0, math.MaxInt})
	case RegexExpr:
		// Regex handled as field wildcard at NFA level
	case DisjExpr:
		for _, c := range v.Branches {
			b.extractSymbols(c)
		}
	case SeqExpr:
		for _, c := range v.Steps {
			b.extractSymbols(c)
		}
	case OptionalExpr:
		b.extractSymbols(v.Child)
	case StarExpr:
		b.extractSymbols(v.Child)
	}
}

func (b *dfaBuilder) finalizeRanges() {
	pointSet := map[int]bool{}
	for _, r := range b.collectedRanges {
		if r[0] < r[1] {
			pointSet[r[0]] = true
			pointSet[r[1]] = true
		}
	}

	points := make([]int, 0, len(pointSet))
	for p := range pointSet {
		points = append(points, p)
	}
	sort.Ints(points)

	b.ranges = nil
	for i := 0; i+1 < len(points); i++ {
		start := points[i]
		end := points[i+1]
		if start < end {
			symID := len(b.alphabet)
			b.alphabet = append(b.alphabet, TransitionLabel{Kind: LabelRange, RangeLo: start, RangeHi: end})
			b.ranges = append(b.ranges, rangeEntry{Start: start, End: end, SymbolID: symID})
		}
	}
}

func (b *dfaBuilder) determinize(n *nfa) *DFA {
	alphaLen := len(b.alphabet)
	nfaStates := n.NumStates
	nfaWords := (nfaStates + 63) / 64

	stateToID := map[uint64]int{}
	var dfaStates [][]uint64
	var transitions [][]int
	var isAccepting []bool

	startSet := make([]uint64, nfaWords)
	startSet[n.StartState/64] |= 1 << (n.StartState % 64)
	startHash := bitsetHash(startSet)

	stateToID[startHash] = 0
	dfaStates = append(dfaStates, startSet)
	transitions = append(transitions, make([]int, alphaLen))
	for i := range transitions[0] {
		transitions[0][i] = -1
	}
	isAccepting = append(isAccepting, n.IsAccepting[n.StartState])

	queue := [][]uint64{startSet}

	for len(queue) > 0 {
		currentSet := queue[0]
		queue = queue[1:]
		currentID := stateToID[bitsetHash(currentSet)]

		for symID := 0; symID < alphaLen; symID++ {
			nextSet := make([]uint64, nfaWords)

			for wordIdx := 0; wordIdx < nfaWords; wordIdx++ {
				words := currentSet[wordIdx]
				for words != 0 {
					bit := bits.TrailingZeros64(words)
					nfaState := wordIdx*64 + bit
					if nfaState >= nfaStates {
						break
					}
					words &= words - 1 // clear lowest set bit

					for _, tr := range n.Transitions[nfaState] {
						nfaLabel := n.PosToLabel[tr.LabelIdx]
						dfaSym := b.alphabet[symID]

						if nfaLabelMatchesDFA(nfaLabel, dfaSym, b.caseInsensitive) {
							nextSet[tr.Dest/64] |= 1 << (tr.Dest % 64)
						}
					}
				}
			}

			hasAny := false
			for _, w := range nextSet {
				if w != 0 {
					hasAny = true
					break
				}
			}
			if !hasAny {
				continue
			}

			nextHash := bitsetHash(nextSet)
			if nextID, exists := stateToID[nextHash]; exists {
				transitions[currentID][symID] = nextID
			} else {
				nextID := len(dfaStates)
				stateToID[nextHash] = nextID
				dfaStates = append(dfaStates, nextSet)
				transitions = append(transitions, make([]int, alphaLen))
				for i := range transitions[nextID] {
					transitions[nextID][i] = -1
				}
				accept := false
				for i, w := range nextSet {
					if w != 0 {
						base := i * 64
						wr := w
						for wr != 0 {
							bit := bits.TrailingZeros64(wr)
							st := base + bit
							if st < nfaStates && n.IsAccepting[st] {
								accept = true
								goto foundAccept
							}
							wr &= wr - 1
						}
					}
				}
			foundAccept:
				isAccepting = append(isAccepting, accept)
				queue = append(queue, nextSet)
				transitions[currentID][symID] = nextID
			}
		}
	}

	return &DFA{
		numStates:       len(dfaStates),
		startState:      0,
		isAccepting:     isAccepting,
		transitions:     transitions,
		alphabet:        b.alphabet,
		keyToID:         b.keyToID,
		ranges:          b.ranges,
		caseInsensitive: b.caseInsensitive,
	}
}

// bitsetHash computes a hash of a uint64 bitset for use as a map key.
func bitsetHash(bits []uint64) uint64 {
	h := uint64(0x9e3779b97f4a7c15)
	for _, w := range bits {
		h ^= w
		h *= 0xbf58476d1ce4e5b9
		h = (h >> 27) | (h << 37)
	}
	return h
}

func nfaLabelMatchesDFA(nfaLabel, dfaSym TransitionLabel, caseInsensitive bool) bool {
	switch nfaLabel.Kind {
	case LabelField:
		if dfaSym.Kind == LabelField {
			nfaField := nfaLabel.Field
			dfaField := dfaSym.Field
			if caseInsensitive {
				nfaField = strings.ToLower(nfaField)
			}
			return nfaField == dfaField
		}

	case LabelFieldWildcard:
		if dfaSym.Kind == LabelOther || dfaSym.Kind == LabelField {
			return true
		}

	case LabelRange:
		if dfaSym.Kind == LabelRange {
			return nfaLabel.RangeLo <= dfaSym.RangeLo && dfaSym.RangeHi <= nfaLabel.RangeHi
		}
	}
	return false
}


