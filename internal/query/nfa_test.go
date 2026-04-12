package query

import (
	"math"
	"testing"
)

func countTrue(bits []bool) int {
	n := 0
	for _, b := range bits {
		if b {
			n++
		}
	}
	return n
}

func TestNFAEmptyQuery(t *testing.T) {
	q := NewSequence(nil)
	nfa := NewQueryNFA(q)
	if nfa.NumStates != 1 {
		t.Fatalf("expected 1 state, got %d", nfa.NumStates)
	}
	if !nfa.IsEmpty {
		t.Fatal("expected IsEmpty=true")
	}
	if !nfa.IsAccepting[0] {
		t.Fatal("start state should be accepting")
	}
}

func TestNFASingleField(t *testing.T) {
	q := NewField("foo")
	nfa := NewQueryNFA(q)

	if nfa.NumStates != 2 {
		t.Fatalf("expected 2 states, got %d", nfa.NumStates)
	}
	if nfa.IsEmpty {
		t.Fatal("expected IsEmpty=false")
	}
	if countTrue(nfa.IsFirst) != 1 || !nfa.IsFirst[0] {
		t.Fatalf("expected first set = {0}, got %v", nfa.IsFirst)
	}
	if countTrue(nfa.IsLast) != 1 || !nfa.IsLast[0] {
		t.Fatalf("expected last set = {0}, got %v", nfa.IsLast)
	}
	if nfa.IsAccepting[0] {
		t.Fatal("start state should not be accepting")
	}
	if !nfa.IsAccepting[1] {
		t.Fatal("state 1 should be accepting")
	}
	if len(nfa.Transitions[0]) != 1 {
		t.Fatalf("expected 1 transition from start, got %d", len(nfa.Transitions[0]))
	}
	if nfa.Transitions[0][0].LabelIdx != 0 || nfa.Transitions[0][0].Dest != 1 {
		t.Fatalf("expected start -> state 1 on label 0, got %+v", nfa.Transitions[0][0])
	}
	if nfa.PosToLabel[0].Kind != LabelField || nfa.PosToLabel[0].Field != "foo" {
		t.Fatalf("expected label Field(foo), got %v", nfa.PosToLabel[0])
	}
}

func TestNFAOptionalField(t *testing.T) {
	q := NewOptional(NewField("foo"))
	nfa := NewQueryNFA(q)

	if nfa.NumStates != 2 {
		t.Fatalf("expected 2 states, got %d", nfa.NumStates)
	}
	if !nfa.IsEmpty {
		t.Fatal("expected IsEmpty=true")
	}
	if countTrue(nfa.IsAccepting) != 2 {
		t.Fatalf("expected 2 accepting states, got %d", countTrue(nfa.IsAccepting))
	}
	if !nfa.IsAccepting[0] || !nfa.IsAccepting[1] {
		t.Fatalf("expected both states accepting, got %v", nfa.IsAccepting)
	}
	if countTrue(nfa.IsFirst) != 1 || !nfa.IsFirst[0] {
		t.Fatalf("expected first set = {0}, got %v", nfa.IsFirst)
	}
	if countTrue(nfa.IsLast) != 1 || !nfa.IsLast[0] {
		t.Fatalf("expected last set = {0}, got %v", nfa.IsLast)
	}
}

func TestNFAKleeneStar(t *testing.T) {
	q := NewKleeneStar(NewField("a"))
	nfa := NewQueryNFA(q)

	if nfa.NumStates != 2 {
		t.Fatalf("expected 2 states, got %d", nfa.NumStates)
	}
	if !nfa.IsEmpty {
		t.Fatal("expected IsEmpty=true")
	}
	if !nfa.IsAccepting[0] {
		t.Fatal("start state should be accepting")
	}
	if !nfa.IsAccepting[1] {
		t.Fatal("state 1 should be accepting")
	}

	if len(nfa.Transitions[0]) != 1 {
		t.Fatalf("expected 1 transition from start, got %d", len(nfa.Transitions[0]))
	}
	if len(nfa.Transitions[1]) != 1 {
		t.Fatalf("expected 1 loopback transition from state 1, got %d", len(nfa.Transitions[1]))
	}
	if nfa.Transitions[1][0].Dest != 1 {
		t.Fatalf("expected loopback to state 1, got dest %d", nfa.Transitions[1][0].Dest)
	}
}

func TestNFASequence(t *testing.T) {
	q := NewSequence([]Query{NewField("foo"), NewField("bar"), NewField("baz")})
	nfa := NewQueryNFA(q)

	if nfa.NumStates != 4 {
		t.Fatalf("expected 4 states, got %d", nfa.NumStates)
	}
	if countTrue(nfa.IsAccepting) != 1 {
		t.Fatalf("expected 1 accepting state, got %d", countTrue(nfa.IsAccepting))
	}
	if !nfa.IsAccepting[3] {
		t.Fatal("state 3 should be accepting (last position)")
	}
	if countTrue(nfa.IsFirst) != 1 || !nfa.IsFirst[0] {
		t.Fatalf("expected first set = {0}, got %v", nfa.IsFirst)
	}
	if countTrue(nfa.IsLast) != 1 || !nfa.IsLast[2] {
		t.Fatalf("expected last set = {2}, got %v", nfa.IsLast)
	}

	if len(nfa.Transitions[0]) != 1 || nfa.Transitions[0][0].Dest != 1 {
		t.Fatalf("expected start -> 1, got %v", nfa.Transitions[0])
	}
	if len(nfa.Transitions[1]) != 1 || nfa.Transitions[1][0].Dest != 2 {
		t.Fatalf("expected state 1 -> 2, got %v", nfa.Transitions[1])
	}
	if len(nfa.Transitions[2]) != 1 || nfa.Transitions[2][0].Dest != 3 {
		t.Fatalf("expected state 2 -> 3, got %v", nfa.Transitions[2])
	}
}

func TestNFADisjunction(t *testing.T) {
	q := NewDisjunction([]Query{NewField("foo"), NewField("bar")})
	nfa := NewQueryNFA(q)

	if nfa.NumStates != 3 {
		t.Fatalf("expected 3 states, got %d", nfa.NumStates)
	}
	if countTrue(nfa.IsAccepting) != 2 {
		t.Fatalf("expected 2 accepting states, got %d", countTrue(nfa.IsAccepting))
	}
	if !nfa.IsAccepting[1] || !nfa.IsAccepting[2] {
		t.Fatalf("expected states 1 and 2 accepting, got %v", nfa.IsAccepting)
	}
	if countTrue(nfa.IsFirst) != 2 {
		t.Fatalf("expected 2 in first set, got %d", countTrue(nfa.IsFirst))
	}
	if !nfa.IsFirst[0] || !nfa.IsFirst[1] {
		t.Fatalf("expected first set = {0,1}, got %v", nfa.IsFirst)
	}
	if !nfa.IsLast[0] || !nfa.IsLast[1] {
		t.Fatalf("expected last set = {0,1}, got %v", nfa.IsLast)
	}
	if len(nfa.Transitions[0]) != 2 {
		t.Fatalf("expected 2 transitions from start, got %d", len(nfa.Transitions[0]))
	}
}

func TestNFAFieldBranchDisjunction(t *testing.T) {
	q := NewDisjunction([]Query{
		NewSequence([]Query{NewField("foo"), NewField("a")}),
		NewSequence([]Query{NewField("foo"), NewField("b")}),
	})
	nfa := NewQueryNFA(q)

	if countTrue(nfa.IsFirst) != 2 {
		t.Fatalf("expected 2 in first set, got %d", countTrue(nfa.IsFirst))
	}
	if !nfa.IsFirst[0] || !nfa.IsFirst[2] {
		t.Fatalf("expected first set = {0,2}, got %v", nfa.IsFirst)
	}
	if countTrue(nfa.IsLast) != 2 {
		t.Fatalf("expected 2 in last set, got %d", countTrue(nfa.IsLast))
	}
	if !nfa.IsLast[1] || !nfa.IsLast[3] {
		t.Fatalf("expected last set = {1,3}, got %v", nfa.IsLast)
	}
	if countTrue(nfa.IsAccepting) != 2 {
		t.Fatalf("expected 2 accepting states, got %d", countTrue(nfa.IsAccepting))
	}
}

func TestNFADisjunctionSequences(t *testing.T) {
	q := NewDisjunction([]Query{
		NewSequence([]Query{NewField("foo"), NewField("bar"), NewField("baz")}),
		NewSequence([]Query{NewField("x"), NewField("y"), NewField("z")}),
	})
	nfa := NewQueryNFA(q)

	if countTrue(nfa.IsAccepting) != 2 {
		t.Fatalf("expected 2 accepting, got %d", countTrue(nfa.IsAccepting))
	}
	if !nfa.IsFirst[0] || !nfa.IsFirst[3] {
		t.Fatalf("expected first set = {0,3}, got %v", nfa.IsFirst)
	}
	if !nfa.IsLast[2] || !nfa.IsLast[5] {
		t.Fatalf("expected last set = {2,5}, got %v", nfa.IsLast)
	}
}

func TestNFAComplexDisjunction(t *testing.T) {
	q := NewDisjunction([]Query{
		NewSequence([]Query{NewField("foo"), NewField("bar")}),
		NewOptional(NewField("bar")),
		NewKleeneStar(NewField("baz")),
	})
	nfa := NewQueryNFA(q)

	if countTrue(nfa.IsAccepting) != 4 {
		t.Fatalf("expected 4 accepting states, got %d: %v", countTrue(nfa.IsAccepting), nfa.IsAccepting)
	}
	if countTrue(nfa.IsFirst) != 3 {
		t.Fatalf("expected 3 in first set, got %d", countTrue(nfa.IsFirst))
	}
	if !nfa.IsFirst[0] || !nfa.IsFirst[2] || !nfa.IsFirst[3] {
		t.Fatalf("expected first set = {0,2,3}, got %v", nfa.IsFirst)
	}
	if countTrue(nfa.IsLast) != 3 {
		t.Fatalf("expected 3 in last set, got %d", countTrue(nfa.IsLast))
	}
	if !nfa.IsLast[1] || !nfa.IsLast[2] || !nfa.IsLast[3] {
		t.Fatalf("expected last set = {1,2,3}, got %v", nfa.IsLast)
	}
}

func TestNFARangeOverlap(t *testing.T) {
	q := NewDisjunction([]Query{
		NewSequence([]Query{NewField("foo"), NewIndex(1)}),
		NewSequence([]Query{NewField("foo"), NewArrayWildcard()}),
	})
	nfa := NewQueryNFA(q)

	if countTrue(nfa.IsFirst) != 2 {
		t.Fatalf("expected 2 in first set, got %d", countTrue(nfa.IsFirst))
	}
	if !nfa.IsFirst[0] || !nfa.IsFirst[2] {
		t.Fatalf("expected first set = {0,2}, got %v", nfa.IsFirst)
	}
	if !nfa.IsLast[1] || !nfa.IsLast[3] {
		t.Fatalf("expected last set = {1,3}, got %v", nfa.IsLast)
	}

	if nfa.PosToLabel[1].Kind != LabelRange || nfa.PosToLabel[1].RangeLo != 1 || nfa.PosToLabel[1].RangeHi != 2 {
		t.Fatalf("expected label Range(1,2) at pos 1, got %v", nfa.PosToLabel[1])
	}
	if nfa.PosToLabel[3].Kind != LabelRange || nfa.PosToLabel[3].RangeLo != 0 || nfa.PosToLabel[3].RangeHi != math.MaxInt {
		t.Fatalf("expected label Range(0,MaxInt) at pos 3, got %v", nfa.PosToLabel[3])
	}
}

func TestNFAKleeneWithFollowing(t *testing.T) {
	q := NewSequence([]Query{
		NewKleeneStar(NewField("a")),
		NewField("b"),
	})
	nfa := NewQueryNFA(q)

	if countTrue(nfa.IsAccepting) != 1 {
		t.Fatalf("expected 1 accepting state, got %d", countTrue(nfa.IsAccepting))
	}
	if !nfa.IsLast[1] {
		t.Fatalf("expected last set = {1} (position 'b'), got %v", nfa.IsLast)
	}
	if countTrue(nfa.IsFirst) != 2 {
		t.Fatalf("expected 2 in first set (a or b), got %d", countTrue(nfa.IsFirst))
	}
	if !nfa.IsFirst[0] || !nfa.IsFirst[1] {
		t.Fatalf("expected first set = {0,1}, got %v", nfa.IsFirst)
	}

	if len(nfa.Transitions[0]) != 2 {
		t.Fatalf("expected 2 transitions from start, got %d", len(nfa.Transitions[0]))
	}

	hasLoopback := false
	for _, tr := range nfa.Transitions[1] {
		if tr.Dest == 1 {
			hasLoopback = true
		}
	}
	if !hasLoopback {
		t.Fatal("expected loopback transition on state 1 (position 'a')")
	}
}

func TestNFAMultipleOptional(t *testing.T) {
	q := NewSequence([]Query{
		NewKleeneStar(NewField("a")),
		NewOptional(NewField("b")),
		NewOptional(NewField("c")),
	})
	nfa := NewQueryNFA(q)

	if countTrue(nfa.IsAccepting) != 4 {
		t.Fatalf("expected 4 accepting states, got %d: %v", countTrue(nfa.IsAccepting), nfa.IsAccepting)
	}
	if countTrue(nfa.IsFirst) != 3 {
		t.Fatalf("expected 3 in first set, got %d", countTrue(nfa.IsFirst))
	}
	if !nfa.IsFirst[0] || !nfa.IsFirst[1] || !nfa.IsFirst[2] {
		t.Fatalf("expected first set = {0,1,2}, got %v", nfa.IsFirst)
	}
	if countTrue(nfa.IsLast) != 3 {
		t.Fatalf("expected 3 in last set, got %d", countTrue(nfa.IsLast))
	}
	if !nfa.IsLast[0] || !nfa.IsLast[1] || !nfa.IsLast[2] {
		t.Fatalf("expected last set = {0,1,2}, got %v", nfa.IsLast)
	}
}

func TestNFAKleeneSequence(t *testing.T) {
	q := NewSequence([]Query{
		NewKleeneStar(NewFieldWildcard()),
		NewKleeneStar(NewArrayWildcard()),
		NewArrayWildcard(),
	})
	nfa := NewQueryNFA(q)

	if len(nfa.Follows) < 1 {
		t.Fatal("expected follows set")
	}
	found := false
	for _, f := range nfa.Follows[0] {
		if f == 2 {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected FieldWildcard (pos 0) to be followed by second ArrayWildcard (pos 2), follows[0]=%v", nfa.Follows[0])
	}
}

func TestNFAIndex(t *testing.T) {
	q := NewIndex(5)
	nfa := NewQueryNFA(q)

	if nfa.PosToLabel[0].Kind != LabelRange {
		t.Fatalf("expected Range label, got %v", nfa.PosToLabel[0])
	}
	if nfa.PosToLabel[0].RangeLo != 5 || nfa.PosToLabel[0].RangeHi != 6 {
		t.Fatalf("expected Range(5,6), got Range(%d,%d)", nfa.PosToLabel[0].RangeLo, nfa.PosToLabel[0].RangeHi)
	}
}

func TestNFAArrayWildcardLabel(t *testing.T) {
	q := NewArrayWildcard()
	nfa := NewQueryNFA(q)

	if nfa.PosToLabel[0].Kind != LabelRange {
		t.Fatalf("expected Range label for ArrayWildcard, got %v", nfa.PosToLabel[0])
	}
	if nfa.PosToLabel[0].RangeLo != 0 || nfa.PosToLabel[0].RangeHi != math.MaxInt {
		t.Fatalf("expected Range(0,MaxInt), got Range(%d,%d)", nfa.PosToLabel[0].RangeLo, nfa.PosToLabel[0].RangeHi)
	}
}

func TestNFARangeFromLabel(t *testing.T) {
	q := NewRangeFrom(7)
	nfa := NewQueryNFA(q)

	if nfa.PosToLabel[0].Kind != LabelRange {
		t.Fatalf("expected Range label, got %v", nfa.PosToLabel[0])
	}
	if nfa.PosToLabel[0].RangeLo != 7 {
		t.Fatalf("expected RangeLo=7, got %d", nfa.PosToLabel[0].RangeLo)
	}
	if nfa.PosToLabel[0].RangeHi != math.MaxInt {
		t.Fatalf("expected RangeHi=MaxInt for RangeFrom, got %d", nfa.PosToLabel[0].RangeHi)
	}
}

func TestNFAFieldWildcardLabel(t *testing.T) {
	q := NewFieldWildcard()
	nfa := NewQueryNFA(q)

	if nfa.PosToLabel[0].Kind != LabelFieldWildcard {
		t.Fatalf("expected FieldWildcard label, got %v", nfa.PosToLabel[0])
	}
}

func TestNFAParseAndBuild(t *testing.T) {
	q, err := ParseQuery("foo.bar[0]?.baz*")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	nfa := NewQueryNFA(q)

	if nfa.NumStates == 0 {
		t.Fatal("expected non-zero states")
	}
	if len(nfa.PosToLabel) == 0 {
		t.Fatal("expected non-zero labels")
	}

	accepting := 0
	for _, a := range nfa.IsAccepting {
		if a {
			accepting++
		}
	}
	if accepting == 0 {
		t.Fatal("expected at least one accepting state")
	}
}

func TestNFAAnyPathGroup(t *testing.T) {
	q, err := ParseQuery("(* | [*])*")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	nfa := NewQueryNFA(q)

	if !nfa.IsEmpty {
		t.Fatal("expected IsEmpty=true for (* | [*])*")
	}
	if !nfa.IsAccepting[0] {
		t.Fatal("start state should be accepting")
	}
	if nfa.NumStates < 3 {
		t.Fatalf("expected at least 3 states (start + FieldWildcard + ArrayWildcard), got %d", nfa.NumStates)
	}
}
