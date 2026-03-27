package query

import "testing"

func TestNewField(t *testing.T) {
	q := NewField("foo")
	if q.Kind != QueryField {
		t.Fatalf("expected QueryField, got %d", q.Kind)
	}
	if q.Field != "foo" {
		t.Fatalf("expected field 'foo', got %q", q.Field)
	}
}

func TestNewIndex(t *testing.T) {
	q := NewIndex(3)
	if q.Kind != QueryIndex {
		t.Fatalf("expected QueryIndex, got %d", q.Kind)
	}
	if q.Index != 3 {
		t.Fatalf("expected index 3, got %d", q.Index)
	}
}

func TestNewRange(t *testing.T) {
	q := NewRange(2, 5)
	if q.Kind != QueryRange {
		t.Fatalf("expected QueryRange, got %d", q.Kind)
	}
	if q.Index != 2 || q.RangeEnd != 5 {
		t.Fatalf("expected range [2,5), got [%d,%d)", q.Index, q.RangeEnd)
	}
}

func TestNewRangeFrom(t *testing.T) {
	q := NewRangeFrom(3)
	if q.Kind != QueryRangeFrom {
		t.Fatalf("expected QueryRangeFrom, got %d", q.Kind)
	}
	if q.Index != 3 {
		t.Fatalf("expected rangeFrom 3, got %d", q.Index)
	}
}

func TestNewFieldWildcard(t *testing.T) {
	q := NewFieldWildcard()
	if q.Kind != QueryFieldWildcard {
		t.Fatalf("expected QueryFieldWildcard, got %d", q.Kind)
	}
}

func TestNewArrayWildcard(t *testing.T) {
	q := NewArrayWildcard()
	if q.Kind != QueryArrayWildcard {
		t.Fatalf("expected QueryArrayWildcard, got %d", q.Kind)
	}
}

func TestNewRegex(t *testing.T) {
	q := NewRegex("^foo")
	if q.Kind != QueryRegex {
		t.Fatalf("expected QueryRegex, got %d", q.Kind)
	}
	if q.Regex != "^foo" {
		t.Fatalf("expected regex '^foo', got %q", q.Regex)
	}
}

func TestNewOptional(t *testing.T) {
	inner := NewField("foo")
	q := NewOptional(inner)
	if q.Kind != QueryOptional {
		t.Fatalf("expected QueryOptional, got %d", q.Kind)
	}
	if len(q.Children) != 1 || q.Children[0].Kind != QueryField {
		t.Fatal("optional should wrap one child")
	}
}

func TestNewKleeneStar(t *testing.T) {
	inner := NewField("foo")
	q := NewKleeneStar(inner)
	if q.Kind != QueryKleeneStar {
		t.Fatalf("expected QueryKleeneStar, got %d", q.Kind)
	}
	if len(q.Children) != 1 || q.Children[0].Kind != QueryField {
		t.Fatal("kleene star should wrap one child")
	}
}

func TestNewDisjunction(t *testing.T) {
	q := NewDisjunction([]Query{NewField("foo"), NewField("bar")})
	if q.Kind != QueryDisjunction {
		t.Fatalf("expected QueryDisjunction, got %d", q.Kind)
	}
	if len(q.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(q.Children))
	}
}

func TestNewSequence(t *testing.T) {
	q := NewSequence([]Query{NewField("foo"), NewField("bar"), NewIndex(0)})
	if q.Kind != QuerySequence {
		t.Fatalf("expected QuerySequence, got %d", q.Kind)
	}
	if len(q.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(q.Children))
	}
}

func TestQueryDepthAtom(t *testing.T) {
	if d := NewField("foo").Depth(); d != 1 {
		t.Fatalf("expected depth 1 for atom, got %d", d)
	}
}

func TestQueryDepthSequence(t *testing.T) {
	q := NewSequence([]Query{NewField("a"), NewField("b")})
	if d := q.Depth(); d != 2 {
		t.Fatalf("expected depth 2 for sequence, got %d", d)
	}
}

func TestQueryDepthDisjunction(t *testing.T) {
	q := NewDisjunction([]Query{
		NewSequence([]Query{NewField("a"), NewField("b")}),
		NewField("c"),
	})
	if d := q.Depth(); d != 3 {
		t.Fatalf("expected depth 3 for disjunction (1 + max(2,1)), got %d", d)
	}
}

func TestQueryDepthKleeneStar(t *testing.T) {
	q := NewKleeneStar(NewField("foo"))
	if d := q.Depth(); d != 2 {
		t.Fatalf("expected depth 2 for kleene star over atom, got %d", d)
	}
}

func TestQueryDepthOptional(t *testing.T) {
	q := NewOptional(NewField("foo"))
	if d := q.Depth(); d != 2 {
		t.Fatalf("expected depth 2 for optional over atom, got %d", d)
	}
}
