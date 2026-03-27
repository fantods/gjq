package query

import (
	"fmt"
	"math"
	"strings"
)

type NFATransition struct {
	LabelIdx int
	Dest     int
}

type QueryNFA struct {
	NumStates     int
	StartState    int
	Transitions   [][]NFATransition
	PosToLabel    []TransitionLabel
	IsAccepting   []bool
	IsFirst       []bool
	IsLast        []bool
	Follows       [][]int
	ContainsEmpty bool
}

func NewQueryNFA(q *Query) *QueryNFA {
	var posToLabel []TransitionLabel
	linearize(q, &posToLabel)

	if len(posToLabel) == 0 {
		return &QueryNFA{
			NumStates:     1,
			StartState:    0,
			Transitions:   [][]NFATransition{{}},
			PosToLabel:    nil,
			IsAccepting:   []bool{true},
			IsFirst:       nil,
			IsLast:        nil,
			Follows:       nil,
			ContainsEmpty: true,
		}
	}

	alphaSize := len(posToLabel)
	numStates := 1 + alphaSize

	containsEmpty := computeContainsEmpty(q)

	isFirst := make([]bool, alphaSize)
	pos := 0
	computeFirst(isFirst, q, &pos)

	isLast := make([]bool, alphaSize)
	pos = 0
	computeLast(isLast, q, &pos)

	follows := make([][]int, alphaSize)
	pos = 0
	computeFollows(follows, q, &pos)

	transitions := make([][]NFATransition, numStates)
	isAccepting := make([]bool, numStates)

	if containsEmpty {
		isAccepting[0] = true
	}

	for i, first := range isFirst {
		if first {
			transitions[0] = append(transitions[0], NFATransition{LabelIdx: i, Dest: i + 1})
		}
	}

	for i, last := range isLast {
		if last {
			isAccepting[i+1] = true
		}
	}

	for from, followers := range follows {
		for _, f := range followers {
			transitions[from+1] = append(transitions[from+1], NFATransition{LabelIdx: f, Dest: f + 1})
		}
	}

	return &QueryNFA{
		NumStates:     numStates,
		StartState:    0,
		Transitions:   transitions,
		PosToLabel:    posToLabel,
		IsAccepting:   isAccepting,
		IsFirst:       isFirst,
		IsLast:        isLast,
		Follows:       follows,
		ContainsEmpty: containsEmpty,
	}
}

func (n *QueryNFA) String() string {
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
	fmt.Fprintf(&b, "ContainsEmpty: %v\n", n.ContainsEmpty)

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

func linearize(q *Query, labels *[]TransitionLabel) {
	switch q.Kind {
	case QueryField:
		*labels = append(*labels, TransitionLabel{Kind: LabelField, Field: q.Field})
	case QueryFieldWildcard:
		*labels = append(*labels, TransitionLabel{Kind: LabelFieldWildcard})
	case QueryIndex:
		*labels = append(*labels, TransitionLabel{Kind: LabelRange, RangeLo: q.Index, RangeHi: q.Index + 1})
	case QueryRange:
		*labels = append(*labels, TransitionLabel{Kind: LabelRange, RangeLo: q.Index, RangeHi: q.RangeEnd})
	case QueryRangeFrom:
		*labels = append(*labels, TransitionLabel{Kind: LabelRangeFrom, RangeLo: q.Index})
	case QueryArrayWildcard:
		*labels = append(*labels, TransitionLabel{Kind: LabelRange, RangeLo: 0, RangeHi: math.MaxInt})
	case QueryRegex:
		*labels = append(*labels, TransitionLabel{Kind: LabelFieldWildcard})
	case QueryDisjunction, QuerySequence:
		for _, c := range q.Children {
			linearize(&c, labels)
		}
	case QueryOptional, QueryKleeneStar:
		linearize(&q.Children[0], labels)
	}
}

func computeContainsEmpty(q *Query) bool {
	switch q.Kind {
	case QueryField, QueryIndex, QueryRange, QueryRangeFrom, QueryArrayWildcard, QueryFieldWildcard, QueryRegex:
		return false
	case QuerySequence:
		for _, c := range q.Children {
			if !computeContainsEmpty(&c) {
				return false
			}
		}
		return true
	case QueryDisjunction:
		for _, c := range q.Children {
			if computeContainsEmpty(&c) {
				return true
			}
		}
		return false
	case QueryOptional, QueryKleeneStar:
		return true
	default:
		return false
	}
}

func countPositions(q *Query) int {
	switch q.Kind {
	case QueryField, QueryIndex, QueryRange, QueryRangeFrom, QueryArrayWildcard, QueryFieldWildcard, QueryRegex:
		return 1
	case QuerySequence, QueryDisjunction:
		sum := 0
		for _, c := range q.Children {
			sum += countPositions(&c)
		}
		return sum
	case QueryOptional, QueryKleeneStar:
		return countPositions(&q.Children[0])
	default:
		return 0
	}
}

func computeFirst(first []bool, q *Query, pos *int) {
	switch q.Kind {
	case QueryField, QueryIndex, QueryRange, QueryRangeFrom, QueryArrayWildcard, QueryFieldWildcard, QueryRegex:
		if *pos < len(first) {
			first[*pos] = true
		}
		*pos++
	case QueryDisjunction:
		for _, c := range q.Children {
			startPos := *pos
			branchLen := countPositions(&c)
			computeFirst(first, &c, pos)
			*pos = startPos + branchLen
		}
	case QuerySequence:
		for _, c := range q.Children {
			computeFirst(first, &c, pos)
			if !computeContainsEmpty(&c) {
				break
			}
		}
	case QueryOptional, QueryKleeneStar:
		computeFirst(first, &q.Children[0], pos)
	}
}

func computeLast(last []bool, q *Query, pos *int) {
	switch q.Kind {
	case QueryField, QueryIndex, QueryRange, QueryRangeFrom, QueryArrayWildcard, QueryFieldWildcard, QueryRegex:
		if *pos < len(last) {
			last[*pos] = true
		}
		*pos++
	case QueryDisjunction:
		for _, c := range q.Children {
			startPos := *pos
			branchLen := countPositions(&c)
			computeLast(last, &c, pos)
			*pos = startPos + branchLen
		}
	case QuerySequence:
		subLengths := make([]int, len(q.Children))
		for i, c := range q.Children {
			subLengths[i] = countPositions(&c)
		}
		seqStart := *pos
		for i := len(q.Children) - 1; i >= 0; i-- {
			subStart := seqStart
			for j := 0; j < i; j++ {
				subStart += subLengths[j]
			}
			computeLast(last, &q.Children[i], &subStart)
			if !computeContainsEmpty(&q.Children[i]) {
				break
			}
		}
		total := 0
		for _, l := range subLengths {
			total += l
		}
		*pos = seqStart + total
	case QueryOptional, QueryKleeneStar:
		computeLast(last, &q.Children[0], pos)
	}
}

func computeFollows(follows [][]int, q *Query, pos *int) {
	switch q.Kind {
	case QueryField, QueryIndex, QueryRange, QueryRangeFrom, QueryArrayWildcard, QueryFieldWildcard, QueryRegex:
		*pos++

	case QueryDisjunction:
		for _, c := range q.Children {
			computeFollows(follows, &c, pos)
		}

	case QuerySequence:
		type subRange struct{ start, end int }
		ranges := make([]subRange, len(q.Children))
		for i, c := range q.Children {
			subStart := *pos
			subLen := countPositions(&c)
			computeFollows(follows, &c, pos)
			ranges[i] = subRange{start: subStart, end: subStart + subLen}
		}

		for i := 0; i < len(q.Children); i++ {
			leftQuery := &q.Children[i]
			leftStart := ranges[i].start
			leftEnd := ranges[i].end

			leftLast := make([]bool, len(follows))
			leftPos := leftStart
			computeLast(leftLast, leftQuery, &leftPos)

			for j := i + 1; j < len(q.Children); j++ {
				canSkip := true
				for k := i + 1; k < j; k++ {
					if !computeContainsEmpty(&q.Children[k]) {
						canSkip = false
						break
					}
				}
				if !canSkip {
					continue
				}

				rightStart := ranges[j].start
				rightEnd := ranges[j].end

				rightFirst := make([]bool, len(follows))
				rightPos := rightStart
				computeFirst(rightFirst, &q.Children[j], &rightPos)

				for li := leftStart; li < leftEnd; li++ {
					if leftLast[li] {
						for ri := rightStart; ri < rightEnd; ri++ {
							if rightFirst[ri] {
								follows[li] = append(follows[li], ri)
							}
						}
					}
				}

				if !computeContainsEmpty(&q.Children[j]) {
					break
				}
			}
		}

	case QueryKleeneStar:
		startPos := *pos
		qLen := countPositions(&q.Children[0])

		computeFollows(follows, &q.Children[0], pos)

		lastSet := make([]bool, len(follows))
		lastPos := startPos
		computeLast(lastSet, &q.Children[0], &lastPos)

		firstSet := make([]bool, len(follows))
		firstPos := startPos
		computeFirst(firstSet, &q.Children[0], &firstPos)

		for i := startPos; i < startPos+qLen; i++ {
			if lastSet[i] {
				for j := startPos; j < startPos+qLen; j++ {
					if firstSet[j] {
						follows[i] = append(follows[i], j)
					}
				}
			}
		}

	case QueryOptional:
		computeFollows(follows, &q.Children[0], pos)
	}
}
