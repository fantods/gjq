package query

import (
	"bytes"
	"encoding/json"
	"math"
	"sort"
	"strings"
)

type RangeEntry struct {
	Start    int
	End      int
	SymbolID int
}

type QueryDFA struct {
	NumStates       int
	StartState      int
	IsAccepting     []bool
	Transitions     [][]int
	Alphabet        []TransitionLabel
	KeyToID         map[string]int
	Ranges          []RangeEntry
	CaseInsensitive bool
}

func NewQueryDFA(q *Query, caseInsensitive bool) *QueryDFA {
	if q.Kind == QuerySequence && len(q.Children) == 0 {
		return &QueryDFA{
			NumStates:       1,
			StartState:      0,
			IsAccepting:     []bool{true},
			Transitions:     nil,
			Alphabet:        nil,
			KeyToID:         map[string]int{},
			Ranges:          nil,
			CaseInsensitive: caseInsensitive,
		}
	}

	b := &dfaBuilder{
		alphabet:        []TransitionLabel{{Kind: LabelOther}},
		keyToID:         map[string]int{},
		caseInsensitive: caseInsensitive,
	}

	b.extractSymbols(q)
	b.finalizeRanges()

	nfa := NewQueryNFA(q)
	return b.determinize(nfa)
}

func (d *QueryDFA) FieldSymbolID(field string) int {
	normalized := field
	if d.CaseInsensitive {
		normalized = strings.ToLower(field)
	}
	if id, ok := d.KeyToID[normalized]; ok {
		return id
	}
	return 0
}

func (d *QueryDFA) IndexSymbolID(index int) (int, bool) {
	i := sort.Search(len(d.Ranges), func(i int) bool {
		return d.Ranges[i].Start > index
	})
	if i == 0 {
		return 0, false
	}
	re := d.Ranges[i-1]
	if index < re.End {
		return re.SymbolID, true
	}
	return 0, false
}

func (d *QueryDFA) Transition(state, symbolID int) (int, bool) {
	if state >= d.NumStates || symbolID >= len(d.Alphabet) {
		return 0, false
	}
	next := d.Transitions[state][symbolID]
	if next == -1 {
		return 0, false
	}
	return next, true
}

func (d *QueryDFA) IsAcceptingState(state int) bool {
	return state < d.NumStates && d.IsAccepting[state]
}

func (d *QueryDFA) Find(root interface{}) []JSONPointer {
	var results []JSONPointer
	path := make([]PathType, 0, 16)
	d.traverse(d.StartState, path, root, &results)
	return results
}

func (d *QueryDFA) traverse(state int, path []PathType, value interface{}, results *[]JSONPointer) {
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
	ranges          []RangeEntry
	caseInsensitive bool
}

func (b *dfaBuilder) extractSymbols(q *Query) {
	switch q.Kind {
	case QueryField:
		normalized := q.Field
		if b.caseInsensitive {
			normalized = strings.ToLower(q.Field)
		}
		if _, exists := b.keyToID[normalized]; !exists {
			id := len(b.alphabet)
			b.alphabet = append(b.alphabet, TransitionLabel{Kind: LabelField, Field: normalized})
			b.keyToID[normalized] = id
		}
	case QueryFieldWildcard:
	case QueryIndex:
		b.collectedRanges = append(b.collectedRanges, [2]int{q.Index, q.Index + 1})
	case QueryRange:
		b.collectedRanges = append(b.collectedRanges, [2]int{q.Index, q.RangeEnd})
	case QueryRangeFrom:
		b.collectedRanges = append(b.collectedRanges, [2]int{q.Index, math.MaxInt})
	case QueryArrayWildcard:
		b.collectedRanges = append(b.collectedRanges, [2]int{0, math.MaxInt})
	case QueryRegex:
	case QueryDisjunction, QuerySequence:
		for _, c := range q.Children {
			b.extractSymbols(&c)
		}
	case QueryOptional, QueryKleeneStar:
		b.extractSymbols(&q.Children[0])
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
			b.ranges = append(b.ranges, RangeEntry{Start: start, End: end, SymbolID: symID})
		}
	}
}

type bitmapKey struct {
	bits []bool
}

func (b *dfaBuilder) determinize(nfa *QueryNFA) *QueryDFA {
	alphaLen := len(b.alphabet)
	nfaStates := nfa.NumStates

	stateToID := map[string]int{}
	var dfaStates [][]bool
	var transitions [][]int
	var isAccepting []bool

	startSet := make([]bool, nfaStates)
	startSet[nfa.StartState] = true
	startKey := bitmapStr(startSet)

	stateToID[startKey] = 0
	dfaStates = append(dfaStates, startSet)
	transitions = append(transitions, make([]int, alphaLen))
	for i := range transitions[0] {
		transitions[0][i] = -1
	}
	isAccepting = append(isAccepting, nfa.IsAccepting[nfa.StartState])

	queue := [][]bool{startSet}

	for len(queue) > 0 {
		currentSet := queue[0]
		queue = queue[1:]
		currentID := stateToID[bitmapStr(currentSet)]

		for symID := 0; symID < alphaLen; symID++ {
			nextSet := make([]bool, nfaStates)

			for nfaState := 0; nfaState < nfaStates; nfaState++ {
				if !currentSet[nfaState] {
					continue
				}
				for _, tr := range nfa.Transitions[nfaState] {
					nfaLabel := nfa.PosToLabel[tr.LabelIdx]
					dfaSym := b.alphabet[symID]

					if nfaLabelMatchesDFA(nfaLabel, dfaSym, b.caseInsensitive) {
						nextSet[tr.Dest] = true
					}
				}
			}

			hasAny := false
			for _, v := range nextSet {
				if v {
					hasAny = true
					break
				}
			}
			if !hasAny {
				continue
			}

			nextKey := bitmapStr(nextSet)
			if nextID, exists := stateToID[nextKey]; exists {
				transitions[currentID][symID] = nextID
			} else {
				nextID := len(dfaStates)
				stateToID[nextKey] = nextID
				dfaStates = append(dfaStates, nextSet)
				transitions = append(transitions, make([]int, alphaLen))
				for i := range transitions[nextID] {
					transitions[nextID][i] = -1
				}
				accept := false
				for i, v := range nextSet {
					if v && nfa.IsAccepting[i] {
						accept = true
						break
					}
				}
				isAccepting = append(isAccepting, accept)
				queue = append(queue, nextSet)
				transitions[currentID][symID] = nextID
			}
		}
	}

	return &QueryDFA{
		NumStates:       len(dfaStates),
		StartState:      0,
		IsAccepting:     isAccepting,
		Transitions:     transitions,
		Alphabet:        b.alphabet,
		KeyToID:         b.keyToID,
		Ranges:          b.ranges,
		CaseInsensitive: b.caseInsensitive,
	}
}

func bitmapStr(bits []bool) string {
	b := make([]byte, len(bits))
	for i, v := range bits {
		if v {
			b[i] = 1
		}
	}
	return string(b)
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

	case LabelRangeFrom:
		if dfaSym.Kind == LabelRange {
			return nfaLabel.RangeLo <= dfaSym.RangeLo
		}
	}
	return false
}

func NewDFAFromQueryString(queryStr string, caseInsensitive bool) (*QueryDFA, error) {
	q, err := ParseQuery(queryStr)
	if err != nil {
		return nil, err
	}
	return NewQueryDFA(&q, caseInsensitive), nil
}

func FindWithQuery(root interface{}, queryStr string, caseInsensitive bool) ([]JSONPointer, error) {
	dfa, err := NewDFAFromQueryString(queryStr, caseInsensitive)
	if err != nil {
		return nil, err
	}
	return dfa.Find(root), nil
}

func ParseJSON(input string) (interface{}, error) {
	var result interface{}
	dec := json.NewDecoder(strings.NewReader(input))
	dec.UseNumber()
	if err := dec.Decode(&result); err != nil {
		return nil, err
	}
	return convertNumbers(result), nil
}

func ParseJSONFromBytes(data []byte) (interface{}, error) {
	var result interface{}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&result); err != nil {
		return nil, err
	}
	return convertNumbers(result), nil
}

func convertNumbers(v interface{}) interface{} {
	switch val := v.(type) {
	case json.Number:
		if i, err := val.Int64(); err == nil {
			return int(i)
		}
		if f, err := val.Float64(); err == nil {
			return f
		}
		return val.String()
	case map[string]interface{}:
		m := make(map[string]interface{}, len(val))
		for k, v := range val {
			m[k] = convertNumbers(v)
		}
		return m
	case []interface{}:
		a := make([]interface{}, len(val))
		for i, v := range val {
			a[i] = convertNumbers(v)
		}
		return a
	default:
		return v
	}
}
