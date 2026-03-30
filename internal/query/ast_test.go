package query

import (
	"testing"
)

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

func TestNeedsQuoting(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"simple", "foo", false},
		{"alphanumeric", "foo123", false},
		{"underscore", "foo_bar", false},
		{"empty", "", true},
		{"dot", "a.b", true},
		{"pipe", "a|b", true},
		{"star", "a*b", true},
		{"question", "a?b", true},
		{"bracket_open", "a[b", true},
		{"bracket_close", "a]b", true},
		{"paren_open", "a(b", true},
		{"paren_close", "a)b", true},
		{"slash", "a/b", true},
		{"double_quote", `a"b`, true},
		{"backslash", `a\b`, true},
		{"space", "a b", true},
		{"tab", "a\tb", true},
		{"newline", "a\nb", true},
		{"reserved chars all", ".|*?[]()/", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := needsQuoting(tt.in); got != tt.want {
				t.Errorf("needsQuoting(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestEscapeForQuotedField(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "foo", "foo"},
		{"double_quote", `a"b`, `a\"b`},
		{"backslash", `a\b`, `a\\b`},
		{"both", `a\"b`, `a\\\"b`},
		{"empty", "", ""},
		{"other_reserved", ".|*?[]()/", ".|*?[]()/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := escapeForQuotedField(tt.in); got != tt.want {
				t.Errorf("escapeForQuotedField(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestQueryString_individualNodes(t *testing.T) {
	tests := []struct {
		name string
		q    Query
		want string
	}{
		{"field", NewField("foo"), "foo"},
		{"field needs quoting", NewField("a.b"), `"a.b"`},
		{"field with slash", NewField("/activities"), `"/activities"`},
		{"field empty", NewField(""), `""`},
		{"index", NewIndex(3), "[3]"},
		{"index zero", NewIndex(0), "[0]"},
		{"range", NewRange(2, 5), "[2:5]"},
		{"range from", NewRangeFrom(3), "[3:]"},
		{"range from zero", NewRangeFrom(0), "[0:]"},
		{"field wildcard", NewFieldWildcard(), "*"},
		{"array wildcard", NewArrayWildcard(), "[*]"},
		{"regex", NewRegex("foo.bar"), "/foo.bar/"},
		{"regex empty", NewRegex(""), "//"},
		{"optional field", NewOptional(NewField("foo")), "foo?"},
		{"kleene field", NewKleeneStar(NewField("a")), "a*"},
		{"disjunction two fields", NewDisjunction([]Query{NewField("foo"), NewField("bar")}), "foo | bar"},
		{"sequence two fields", NewSequence([]Query{NewField("foo"), NewField("bar")}), "foo.bar"},
		{"optional disjunction multi", NewOptional(NewDisjunction([]Query{NewField("a"), NewField("b")})), "(a | b)?"},
		{"optional disjunction single", NewOptional(NewDisjunction([]Query{NewField("a")})), "a?"},
		{"kleene sequence multi", NewKleeneStar(NewSequence([]Query{NewField("a"), NewField("b")})), "(a.b)*"},
		{"kleene sequence single", NewKleeneStar(NewSequence([]Query{NewField("a")})), "a*"},
		{"optional index", NewOptional(NewIndex(0)), "[0]?"},
		{"kleene disjunction", NewKleeneStar(NewDisjunction([]Query{NewFieldWildcard(), NewArrayWildcard()})), "(* | [*])*"},
		{"empty sequence", NewSequence(nil), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.q.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestQueryString_sequenceDotSuppression(t *testing.T) {
	tests := []struct {
		name string
		q    Query
		want string
	}{
		{"field then index", NewSequence([]Query{NewField("foo"), NewIndex(3)}), "foo[3]"},
		{"field then range", NewSequence([]Query{NewField("foo"), NewRange(1, 3)}), "foo[1:3]"},
		{"field then range from", NewSequence([]Query{NewField("foo"), NewRangeFrom(2)}), "foo[2:]"},
		{"field then field wildcard", NewSequence([]Query{NewField("foo"), NewFieldWildcard()}), "foo*"},
		{"field then array wildcard", NewSequence([]Query{NewField("foo"), NewArrayWildcard()}), "foo[*]"},
		{"index then field", NewSequence([]Query{NewIndex(0), NewField("bar")}), "[0].bar"},
		{"field index field", NewSequence([]Query{NewField("foo"), NewIndex(0), NewField("bar")}), "foo[0].bar"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.q.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestQueryString_sequenceWithDisjunction(t *testing.T) {
	q := NewSequence([]Query{
		NewDisjunction([]Query{NewField("foo"), NewField("bar")}),
		NewField("baz"),
	})
	want := "(foo | bar).baz"
	if got := q.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestQueryString_roundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"field", "foo", "foo"},
		{"field and index", "foo123[42]", "foo123[42]"},
		{"regex", "/foo.bar/", "/foo.bar/"},
		{"disjunction", "foo | bar", "foo | bar"},
		{"kleene", "a*", "a*"},
		{"optional", "b?", "b?"},
		{"complex", "foo.bar[0]?.baz*", "foo.bar[0]?.baz*"},
		{"multiple optional", "c*.c?.c?", "c*.c?.c?"},
		{"disjunction group in sequence", "(foo | bar).baz", "(foo | bar).baz"},
		{"any path group", "(* | [*])*", "(* | [*])*"},
		{"any path in query", "a.(* | [*])*.b?", "a.(* | [*])*.b?"},
		{"nested groups", "((foo.bar)* | bar)", "(foo.bar)* | bar"},
		{"group sequence optional", "(foo.bar.baz)?", "(foo.bar.baz)?"},
		{"empty", "", ""},
		{"reserved in quotes", `".|*?[]()/"`, `".|*?[]()/"`},
		{"group reserved", `"." | "|" | "*" | "?" | "[" | "]" | "(" | ")" | "/"`, `"." | "|" | "*" | "?" | "[" | "]" | "(" | ")" | "/"`},
		{"key with spaces", `"key space".foo`, `"key space".foo`},
		{"dot in field", `"a.b"`, `"a.b"`},
		{"backslash in field", `"a\\b"`, `"a\\b"`},
		{"inner quote in field", `"a\"b"`, `"a\"b"`},
		{"quoted field in sequence", `paths."/activities"`, `paths."/activities"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := ParseQuery(tt.input)
			if err != nil {
				t.Fatalf("ParseQuery(%q) error: %v", tt.input, err)
			}
			got := q.String()
			if got != tt.want {
				t.Errorf("ParseQuery(%q).String() = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestQueryString_nestedGroupsTrivial(t *testing.T) {
	q, err := ParseQuery("((foo))")
	if err != nil {
		t.Fatalf("ParseQuery error: %v", err)
	}
	want := "foo"
	if got := q.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestQueryString_unicodeEscape(t *testing.T) {
	q, err := ParseQuery(`"\u0041"`)
	if err != nil {
		t.Fatalf("ParseQuery error: %v", err)
	}
	want := "A"
	if got := q.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
