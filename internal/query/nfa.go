package query

import (
	"fmt"
	"math"
	"strings"
)

type nfaTransition struct {
	LabelIdx int
	Dest     int
}

type nfa struct {
	NumStates   int
	StartState  int
	Transitions [][]nfaTransition
	PosToLabel  []TransitionLabel
	IsAccepting []bool
	IsFirst     []bool
	IsLast      []bool
	Follows     [][]int
	IsEmpty     bool
}

// posNode carries the position range [start, end) for a subtree of the query.
// It mirrors the Query tree structure but with concrete position indices,
// replacing the *int cursor pattern and countPositions helper.
type posNode struct {
	start, end int       // half-open interval of position indices
	query      Query     // original query (for type switching)
	children   []posNode // child nodes for container types
}

func newNFA(q Query) *nfa {
	labels, root := linearize(q)

	if len(labels) == 0 {
		return &nfa{
			NumStates:   1,
			StartState:  0,
			Transitions: [][]nfaTransition{{}},
			PosToLabel:  nil,
			IsAccepting: []bool{true},
			IsFirst:     nil,
			IsLast:      nil,
			Follows:     nil,
			IsEmpty:     true,
		}
	}

	alphaSize := len(labels)
	numStates := 1 + alphaSize

	isEmpty := root.computeIsEmpty()

	isFirst := make([]bool, alphaSize)
	computeFirst(isFirst, root)

	isLast := make([]bool, alphaSize)
	computeLast(isLast, root)

	follows := make([][]int, alphaSize)
	computeFollows(follows, root)

	transitions := make([][]nfaTransition, numStates)
	isAccepting := make([]bool, numStates)

	if isEmpty {
		isAccepting[0] = true
	}

	for i, first := range isFirst {
		if first {
			transitions[0] = append(transitions[0], nfaTransition{LabelIdx: i, Dest: i + 1})
		}
	}

	for i, last := range isLast {
		if last {
			isAccepting[i+1] = true
		}
	}

	for from, followers := range follows {
		for _, f := range followers {
			transitions[from+1] = append(transitions[from+1], nfaTransition{LabelIdx: f, Dest: f + 1})
		}
	}

	return &nfa{
		NumStates:   numStates,
		StartState:  0,
		Transitions: transitions,
		PosToLabel:  labels,
		IsAccepting: isAccepting,
		IsFirst:     isFirst,
		IsLast:      isLast,
		Follows:     follows,
		IsEmpty:     isEmpty,
	}
}

func (n *nfa) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "NFA States: %d\n", n.NumStates)
	fmt.Fprintf(&b, "Start State: %d\n", n.StartState)

	accepting := make([]int, 0)
	for i, a := range n.IsAccepting {
		if a {
			accepting = append(accepting, i)
		}
	}
	fmt.Fprintf(&b, "Accepting States: %v\n", accepting)
	fmt.Fprintf(&b, "IsEmpty: %v\n", n.IsEmpty)

	fmt.Fprintf(&b, "First set: ")
	for i, f := range n.IsFirst {
		if f {
			fmt.Fprintf(&b, "[%d %s] ", i, n.PosToLabel[i])
		}
	}
	fmt.Fprintln(&b)

	fmt.Fprintf(&b, "Last set: ")
	for i, l := range n.IsLast {
		if l {
			fmt.Fprintf(&b, "[%d %s] ", i, n.PosToLabel[i])
		}
	}
	fmt.Fprintln(&b)

	fmt.Fprintf(&b, "Follows:\n")
	for i, followers := range n.Follows {
		if len(followers) == 0 {
			fmt.Fprintf(&b, "  [%d %s] -> (none)\n", i, n.PosToLabel[i])
		} else {
			fmt.Fprintf(&b, "  [%d %s] -> ", i, n.PosToLabel[i])
			for _, f := range followers {
				fmt.Fprintf(&b, "[%d %s] ", f, n.PosToLabel[f])
			}
			fmt.Fprintln(&b)
		}
	}

	fmt.Fprintf(&b, "Transitions:\n")
	for st, trans := range n.Transitions {
		fmt.Fprintf(&b, "  state %d:\n", st)
		for _, tr := range trans {
			fmt.Fprintf(&b, "    on [%d %s] -> state %d\n", tr.LabelIdx, n.PosToLabel[tr.LabelIdx], tr.Dest)
		}
	}

	return b.String()
}

// linearize walks the query tree, assigns sequential position indices to leaf
// nodes, and returns both the ordered labels and a posNode tree carrying
// position ranges for every subtree.
func linearize(q Query) ([]TransitionLabel, posNode) {
	var labels []TransitionLabel
	node := doLinearize(q, &labels)
	return labels, node
}

func doLinearize(q Query, labels *[]TransitionLabel) posNode {
	switch v := q.(type) {
	case FieldExpr:
		start := len(*labels)
		*labels = append(*labels, TransitionLabel{Kind: LabelField, Field: v.Name})
		return posNode{start: start, end: start + 1, query: q}

	case WildcardExpr:
		start := len(*labels)
		*labels = append(*labels, TransitionLabel{Kind: LabelFieldWildcard})
		return posNode{start: start, end: start + 1, query: q}

	case IndexExpr:
		start := len(*labels)
		*labels = append(*labels, TransitionLabel{Kind: LabelRange, RangeLo: v.Index, RangeHi: v.Index + 1})
		return posNode{start: start, end: start + 1, query: q}

	case RangeExpr:
		start := len(*labels)
		*labels = append(*labels, TransitionLabel{Kind: LabelRange, RangeLo: v.Start, RangeHi: v.End})
		return posNode{start: start, end: start + 1, query: q}

	case RangeFromExpr:
		start := len(*labels)
		*labels = append(*labels, TransitionLabel{Kind: LabelRange, RangeLo: v.Start, RangeHi: math.MaxInt})
		return posNode{start: start, end: start + 1, query: q}

	case ArrayWildExpr:
		start := len(*labels)
		*labels = append(*labels, TransitionLabel{Kind: LabelRange, RangeLo: 0, RangeHi: math.MaxInt})
		return posNode{start: start, end: start + 1, query: q}

	case RegexExpr:
		start := len(*labels)
		*labels = append(*labels, TransitionLabel{Kind: LabelFieldWildcard})
		return posNode{start: start, end: start + 1, query: q}

	case DisjExpr:
		start := len(*labels)
		children := make([]posNode, len(v.Branches))
		for i, c := range v.Branches {
			children[i] = doLinearize(c, labels)
		}
		return posNode{start: start, end: len(*labels), query: q, children: children}

	case SeqExpr:
		start := len(*labels)
		children := make([]posNode, len(v.Steps))
		for i, c := range v.Steps {
			children[i] = doLinearize(c, labels)
		}
		return posNode{start: start, end: len(*labels), query: q, children: children}

	case OptionalExpr:
		start := len(*labels)
		children := []posNode{doLinearize(v.Child, labels)}
		return posNode{start: start, end: len(*labels), query: q, children: children}

	case StarExpr:
		start := len(*labels)
		children := []posNode{doLinearize(v.Child, labels)}
		return posNode{start: start, end: len(*labels), query: q, children: children}

	default:
		return posNode{}
	}
}

// computeIsEmpty returns true if the subtree can match the empty string.
func (n posNode) computeIsEmpty() bool {
	switch n.query.(type) {
	case FieldExpr, IndexExpr, RangeExpr, RangeFromExpr, ArrayWildExpr, WildcardExpr, RegexExpr:
		return false
	case SeqExpr:
		for _, c := range n.children {
			if !c.computeIsEmpty() {
				return false
			}
		}
		return true
	case DisjExpr:
		for _, c := range n.children {
			if c.computeIsEmpty() {
				return true
			}
		}
		return false
	case OptionalExpr, StarExpr:
		return true
	default:
		return false
	}
}

// isAtom returns true for leaf query types.
func isAtom(q Query) bool {
	switch q.(type) {
	case FieldExpr, IndexExpr, RangeExpr, RangeFromExpr, ArrayWildExpr, WildcardExpr, RegexExpr:
		return true
	default:
		return false
	}
}

// computeFirst sets the first-position set: positions that can be matched first.
func computeFirst(first []bool, n posNode) {
	if isAtom(n.query) {
		first[n.start] = true
		return
	}
	switch n.query.(type) {
	case DisjExpr:
		for _, c := range n.children {
			computeFirst(first, c)
		}
	case SeqExpr:
		for _, c := range n.children {
			computeFirst(first, c)
			if !c.computeIsEmpty() {
				break
			}
		}
	case OptionalExpr, StarExpr:
		computeFirst(first, n.children[0])
	}
}

// computeLast sets the last-position set: positions that can be matched last.
func computeLast(last []bool, n posNode) {
	if isAtom(n.query) {
		last[n.start] = true
		return
	}
	switch n.query.(type) {
	case DisjExpr:
		for _, c := range n.children {
			computeLast(last, c)
		}
	case SeqExpr:
		for i := len(n.children) - 1; i >= 0; i-- {
			computeLast(last, n.children[i])
			if !n.children[i].computeIsEmpty() {
				break
			}
		}
	case OptionalExpr, StarExpr:
		computeLast(last, n.children[0])
	}
}

// computeFollows builds the follow-position set: for each position i, which
// positions can immediately follow i in a successful match.
func computeFollows(follows [][]int, n posNode) {
	if isAtom(n.query) {
		return
	}

	switch n.query.(type) {
	case DisjExpr:
		for _, c := range n.children {
			computeFollows(follows, c)
		}

	case SeqExpr:
		// Recurse into children first
		for _, c := range n.children {
			computeFollows(follows, c)
		}
		// For each adjacent pair of steps, link last(i) → first(j)
		for i := 0; i < len(n.children); i++ {
			left := n.children[i]
			leftLast := make([]bool, len(follows))
			computeLast(leftLast, left)

			for j := i + 1; j < len(n.children); j++ {
				right := n.children[j]
				rightFirst := make([]bool, len(follows))
				computeFirst(rightFirst, right)

				for li := left.start; li < left.end; li++ {
					if leftLast[li] {
						for ri := right.start; ri < right.end; ri++ {
							if rightFirst[ri] {
								follows[li] = append(follows[li], ri)
							}
						}
					}
				}

				if !right.computeIsEmpty() {
					break
				}
			}
		}

	case StarExpr:
		child := n.children[0]
		computeFollows(follows, child)

		lastSet := make([]bool, len(follows))
		computeLast(lastSet, child)

		firstSet := make([]bool, len(follows))
		computeFirst(firstSet, child)

		for i := child.start; i < child.end; i++ {
			if lastSet[i] {
				for j := child.start; j < child.end; j++ {
					if firstSet[j] {
						follows[i] = append(follows[i], j)
					}
				}
			}
		}

	case OptionalExpr:
		computeFollows(follows, n.children[0])
	}
}
