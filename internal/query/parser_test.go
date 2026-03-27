package query

import "testing"

func TestParseEmpty(t *testing.T) {
	q, err := ParseQuery("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QuerySequence || len(q.Children) != 0 {
		t.Fatalf("expected empty sequence, got %+v", q)
	}
}

func TestParseField(t *testing.T) {
	q, err := ParseQuery("foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QueryField || q.Field != "foo" {
		t.Fatalf("expected Field('foo'), got %+v", q)
	}
}

func TestParseQuotedField(t *testing.T) {
	q, err := ParseQuery(`"foo bar"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QueryField || q.Field != "foo bar" {
		t.Fatalf("expected Field('foo bar'), got %+v", q)
	}
}

func TestParseIndex(t *testing.T) {
	q, err := ParseQuery("[3]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QueryIndex || q.Index != 3 {
		t.Fatalf("expected Index(3), got %+v", q)
	}
}

func TestParseRange(t *testing.T) {
	q, err := ParseQuery("[2:5]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QueryRange || q.Index != 2 || q.RangeEnd != 5 {
		t.Fatalf("expected Range(2,5), got %+v", q)
	}
}

func TestParseRangeOpenEnd(t *testing.T) {
	q, err := ParseQuery("[3:]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QueryRangeFrom || q.Index != 3 {
		t.Fatalf("expected RangeFrom(3), got %+v", q)
	}
}

func TestParseRangeOpenStart(t *testing.T) {
	q, err := ParseQuery("[:5]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QueryRange || q.Index != 0 || q.RangeEnd != 5 {
		t.Fatalf("expected Range(0,5), got %+v", q)
	}
}

func TestParseArrayWildcard(t *testing.T) {
	q, err := ParseQuery("[*]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QueryArrayWildcard {
		t.Fatalf("expected ArrayWildcard, got %+v", q)
	}
}

func TestParseRangeAll(t *testing.T) {
	q, err := ParseQuery("[:]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QueryArrayWildcard {
		t.Fatalf("expected ArrayWildcard for [:], got %+v", q)
	}
}

func TestParseFieldWildcard(t *testing.T) {
	q, err := ParseQuery("*")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QueryFieldWildcard {
		t.Fatalf("expected FieldWildcard, got %+v", q)
	}
}

func TestParseRegex(t *testing.T) {
	q, err := ParseQuery("/foo.bar/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QueryRegex || q.Regex != "foo.bar" {
		t.Fatalf("expected Regex('foo.bar'), got %+v", q)
	}
}

func TestParseRegexEscapedSlash(t *testing.T) {
	q, err := ParseQuery(`/foo\/bar/`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QueryRegex || q.Regex != "foo/bar" {
		t.Fatalf("expected Regex('foo/bar'), got %+v", q)
	}
}

func TestParseSequence(t *testing.T) {
	q, err := ParseQuery("foo.bar")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QuerySequence || len(q.Children) != 2 {
		t.Fatalf("expected Sequence of 2, got %+v", q)
	}
	if q.Children[0].Kind != QueryField || q.Children[0].Field != "foo" {
		t.Fatalf("expected first child Field('foo'), got %+v", q.Children[0])
	}
	if q.Children[1].Kind != QueryField || q.Children[1].Field != "bar" {
		t.Fatalf("expected second child Field('bar'), got %+v", q.Children[1])
	}
}

func TestParseDisjunction(t *testing.T) {
	q, err := ParseQuery("foo | bar")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QueryDisjunction || len(q.Children) != 2 {
		t.Fatalf("expected Disjunction of 2, got %+v", q)
	}
	if q.Children[0].Kind != QueryField || q.Children[0].Field != "foo" {
		t.Fatalf("expected first branch Field('foo'), got %+v", q.Children[0])
	}
	if q.Children[1].Kind != QueryField || q.Children[1].Field != "bar" {
		t.Fatalf("expected second branch Field('bar'), got %+v", q.Children[1])
	}
}

func TestParseDisjunctionSingle(t *testing.T) {
	q, err := ParseQuery("foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind == QueryDisjunction {
		t.Fatalf("single element should not be wrapped in Disjunction")
	}
}

func TestParseOptional(t *testing.T) {
	q, err := ParseQuery("foo?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QueryOptional {
		t.Fatalf("expected Optional, got %+v", q)
	}
	if len(q.Children) != 1 || q.Children[0].Kind != QueryField || q.Children[0].Field != "foo" {
		t.Fatalf("expected Optional(Field('foo')), got %+v", q)
	}
}

func TestParseKleeneStar(t *testing.T) {
	q, err := ParseQuery("a*")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QueryKleeneStar {
		t.Fatalf("expected KleeneStar, got %+v", q)
	}
	if len(q.Children) != 1 || q.Children[0].Kind != QueryField || q.Children[0].Field != "a" {
		t.Fatalf("expected KleeneStar(Field('a')), got %+v", q)
	}
}

func TestParseFieldWithIndex(t *testing.T) {
	q, err := ParseQuery("foo[3]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QuerySequence || len(q.Children) != 2 {
		t.Fatalf("expected Sequence(Field, Index), got %+v", q)
	}
	if q.Children[0].Kind != QueryField || q.Children[0].Field != "foo" {
		t.Fatalf("expected first child Field('foo'), got %+v", q.Children[0])
	}
	if q.Children[1].Kind != QueryIndex || q.Children[1].Index != 3 {
		t.Fatalf("expected second child Index(3), got %+v", q.Children[1])
	}
}

func TestParseFieldWithArrayWildcard(t *testing.T) {
	q, err := ParseQuery("foo[*]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QuerySequence || len(q.Children) != 2 {
		t.Fatalf("expected Sequence(Field, ArrayWildcard), got %+v", q)
	}
	if q.Children[1].Kind != QueryArrayWildcard {
		t.Fatalf("expected second child ArrayWildcard, got %+v", q.Children[1])
	}
}

func TestParseFieldWithRange(t *testing.T) {
	q, err := ParseQuery("foo[2:5]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QuerySequence || len(q.Children) != 2 {
		t.Fatalf("expected Sequence(Field, Range), got %+v", q)
	}
	if q.Children[1].Kind != QueryRange || q.Children[1].Index != 2 || q.Children[1].RangeEnd != 5 {
		t.Fatalf("expected Range(2,5), got %+v", q.Children[1])
	}
}

func TestParseComplexQuery(t *testing.T) {
	q, err := ParseQuery("foo.bar[0]?.baz*")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QuerySequence {
		t.Fatalf("expected Sequence, got %+v", q)
	}
	if len(q.Children) != 3 {
		t.Fatalf("expected 3 children, got %d: %+v", len(q.Children), q)
	}
	if q.Children[0].Kind != QueryField || q.Children[0].Field != "foo" {
		t.Fatalf("expected Field('foo'), got %+v", q.Children[0])
	}
	if q.Children[1].Kind != QuerySequence || len(q.Children[1].Children) != 2 {
		t.Fatalf("expected Sequence(Field, Optional), got %+v", q.Children[1])
	}
	step := q.Children[1].Children
	if step[0].Kind != QueryField || step[0].Field != "bar" {
		t.Fatalf("expected Field('bar'), got %+v", step[0])
	}
	if step[1].Kind != QueryOptional {
		t.Fatalf("expected Optional, got %+v", step[1])
	}
	optInner := step[1].Children[0]
	if optInner.Kind != QueryIndex || optInner.Index != 0 {
		t.Fatalf("expected Optional(Index(0)), got %+v", optInner)
	}
	if q.Children[2].Kind != QueryKleeneStar {
		t.Fatalf("expected KleeneStar, got %+v", q.Children[2])
	}
	ks := q.Children[2].Children[0]
	if ks.Kind != QueryField || ks.Field != "baz" {
		t.Fatalf("expected KleeneStar(Field('baz')), got %+v", ks)
	}
}

func TestParseGroup(t *testing.T) {
	q, err := ParseQuery("(foo | bar).baz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QuerySequence || len(q.Children) != 2 {
		t.Fatalf("expected Sequence of 2, got %+v", q)
	}
	if q.Children[0].Kind != QueryDisjunction {
		t.Fatalf("expected first child Disjunction, got %+v", q.Children[0])
	}
	if q.Children[1].Kind != QueryField || q.Children[1].Field != "baz" {
		t.Fatalf("expected second child Field('baz'), got %+v", q.Children[1])
	}
}

func TestParseAnyPathGroup(t *testing.T) {
	q, err := ParseQuery("(* | [*])*")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QueryKleeneStar {
		t.Fatalf("expected KleeneStar, got %+v", q)
	}
	inner := q.Children[0]
	if inner.Kind != QueryDisjunction {
		t.Fatalf("expected Disjunction inside KleeneStar, got %+v", inner)
	}
	if len(inner.Children) != 2 {
		t.Fatalf("expected 2 branches, got %d", len(inner.Children))
	}
	if inner.Children[0].Kind != QueryFieldWildcard {
		t.Fatalf("expected first branch FieldWildcard, got %+v", inner.Children[0])
	}
	if inner.Children[1].Kind != QueryArrayWildcard {
		t.Fatalf("expected second branch ArrayWildcard, got %+v", inner.Children[1])
	}
}

func TestParseNestedGroups(t *testing.T) {
	q, err := ParseQuery("((foo.bar)* | bar)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QueryDisjunction || len(q.Children) != 2 {
		t.Fatalf("expected Disjunction of 2, got %+v", q)
	}
	if q.Children[0].Kind != QueryKleeneStar {
		t.Fatalf("expected first branch KleeneStar, got %+v", q.Children[0])
	}
}

func TestParseGroupSequence(t *testing.T) {
	q, err := ParseQuery("(foo.bar.baz)?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QueryOptional {
		t.Fatalf("expected Optional, got %+v", q)
	}
	inner := q.Children[0]
	if inner.Kind != QuerySequence || len(inner.Children) != 3 {
		t.Fatalf("expected Sequence of 3 inside Optional, got %+v", inner)
	}
}

func TestParseNestedGroupsTrivial(t *testing.T) {
	q, err := ParseQuery("((foo))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QueryField || q.Field != "foo" {
		t.Fatalf("expected Field('foo'), got %+v", q)
	}
}

func TestParseMultipleOptional(t *testing.T) {
	q, err := ParseQuery("c*.c?.c?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QuerySequence || len(q.Children) != 3 {
		t.Fatalf("expected Sequence of 3, got %+v", q)
	}
}

func TestParseAnyPathInQuery(t *testing.T) {
	q, err := ParseQuery("a.(* | [*])*.b?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QuerySequence || len(q.Children) != 3 {
		t.Fatalf("expected Sequence of 3, got %+v", q)
	}
	if q.Children[0].Kind != QueryField || q.Children[0].Field != "a" {
		t.Fatalf("expected first child Field('a'), got %+v", q.Children[0])
	}
	if q.Children[1].Kind != QueryKleeneStar {
		t.Fatalf("expected second child KleeneStar, got %+v", q.Children[1])
	}
	if q.Children[2].Kind != QueryOptional {
		t.Fatalf("expected third child Optional, got %+v", q.Children[2])
	}
}

func TestParseInvalidNumber(t *testing.T) {
	_, err := ParseQuery("foo[abc]")
	if err == nil {
		t.Fatal("expected error for non-numeric index")
	}
}

func TestParseUnclosedRegex(t *testing.T) {
	_, err := ParseQuery("/unclosed")
	if err == nil {
		t.Fatal("expected error for unclosed regex")
	}
}

func TestParseUnclosedQuote(t *testing.T) {
	_, err := ParseQuery(`"`)
	if err == nil {
		t.Fatal("expected error for unclosed quote")
	}
}

func TestParseInvalidKeyWithSpaces(t *testing.T) {
	_, err := ParseQuery("spaces not allowed without double quotes")
	if err == nil {
		t.Fatal("expected error for unquoted spaces")
	}
}

func TestParseInvalidKeyWithReservedChars(t *testing.T) {
	_, err := ParseQuery("][")
	if err == nil {
		t.Fatal("expected error for reserved chars")
	}
}

func TestParseQuotedFieldWithReservedChars(t *testing.T) {
	q, err := ParseQuery(`".|*?[]()/"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QueryField || q.Field != ".|*?[]()/" {
		t.Fatalf("expected Field('.|*?[]()/'), got %+v", q)
	}
}

func TestParseQuotedFieldInSequence(t *testing.T) {
	q, err := ParseQuery(`paths."/activities"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QuerySequence || len(q.Children) != 2 {
		t.Fatalf("expected Sequence of 2, got %+v", q)
	}
	if q.Children[0].Field != "paths" {
		t.Fatalf("expected first child 'paths', got %q", q.Children[0].Field)
	}
	if q.Children[1].Field != "/activities" {
		t.Fatalf("expected second child '/activities', got %q", q.Children[1].Field)
	}
}

func TestParseQuotedFieldWithSlash(t *testing.T) {
	q, err := ParseQuery(`"/activities"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QueryField || q.Field != "/activities" {
		t.Fatalf("expected Field('/activities'), got %+v", q)
	}
}

func TestParseQuotedFieldWithDot(t *testing.T) {
	q, err := ParseQuery(`"a.b"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QueryField || q.Field != "a.b" {
		t.Fatalf("expected Field('a.b'), got %+v", q)
	}
}

func TestParseQuotedFieldUnescapeBackslash(t *testing.T) {
	q, err := ParseQuery(`"a\\b"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QueryField || q.Field != "a\\b" {
		t.Fatalf("expected Field('a\\b'), got %+v", q)
	}
}

func TestParseQuotedFieldUnescapeInnerQuote(t *testing.T) {
	q, err := ParseQuery(`"a\"b"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QueryField || q.Field != `a"b` {
		t.Fatalf("expected Field('a\"b'), got %+v", q)
	}
}

func TestParseQuotedFieldUnescapeUnicode(t *testing.T) {
	q, err := ParseQuery(`"\u0041"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QueryField || q.Field != "A" {
		t.Fatalf("expected Field('A'), got %+v", q)
	}
}

func TestParseQuotedFieldUnescapeEscapeSequences(t *testing.T) {
	q, err := ParseQuery(`"\n\r\t\b\f"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QueryField || q.Field != "\n\r\t\b\f" {
		t.Fatalf("expected Field with escape sequences, got %+v", q)
	}
}

func TestParseGroupAnyReservedCharsInDoubleQuotes(t *testing.T) {
	q, err := ParseQuery(`("." | "|" | "*" | "?" | "[" | "]" | "(" | ")" | "/")*`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QueryKleeneStar {
		t.Fatalf("expected KleeneStar, got %+v", q)
	}
	inner := q.Children[0]
	if inner.Kind != QueryDisjunction || len(inner.Children) != 9 {
		t.Fatalf("expected Disjunction of 9, got %+v", inner)
	}
}

func TestParseWhitespaceAroundPipe(t *testing.T) {
	q, err := ParseQuery("foo|bar")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QueryDisjunction || len(q.Children) != 2 {
		t.Fatalf("expected Disjunction of 2, got %+v", q)
	}
}

func TestParseThreeWayDisjunction(t *testing.T) {
	q, err := ParseQuery("a | b | c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QueryDisjunction || len(q.Children) != 3 {
		t.Fatalf("expected Disjunction of 3, got %+v", q)
	}
}

func TestParseFieldWithAlphanumeric(t *testing.T) {
	q, err := ParseQuery("foo123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QueryField || q.Field != "foo123" {
		t.Fatalf("expected Field('foo123'), got %+v", q)
	}
}

func TestParseFieldWithNumberAndIndex(t *testing.T) {
	q, err := ParseQuery("foo123[42]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QuerySequence || len(q.Children) != 2 {
		t.Fatalf("expected Sequence(Field, Index), got %+v", q)
	}
	if q.Children[0].Field != "foo123" {
		t.Fatalf("expected Field('foo123'), got %+v", q.Children[0])
	}
	if q.Children[1].Index != 42 {
		t.Fatalf("expected Index(42), got %+v", q.Children[1])
	}
}

func TestParseStandaloneIndexOptional(t *testing.T) {
	q, err := ParseQuery("[3]?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QueryOptional {
		t.Fatalf("expected Optional, got %+v", q)
	}
	inner := q.Children[0]
	if inner.Kind != QueryIndex || inner.Index != 3 {
		t.Fatalf("expected Optional(Index(3)), got %+v", inner)
	}
}

func TestParseWildcardOptional(t *testing.T) {
	q, err := ParseQuery("*?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QueryOptional {
		t.Fatalf("expected Optional, got %+v", q)
	}
	if q.Children[0].Kind != QueryFieldWildcard {
		t.Fatalf("expected Optional(FieldWildcard), got %+v", q.Children[0])
	}
}

func TestParseArrayWildcardKleeneStar(t *testing.T) {
	q, err := ParseQuery("[*]*")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Kind != QueryKleeneStar {
		t.Fatalf("expected KleeneStar, got %+v", q)
	}
	if q.Children[0].Kind != QueryArrayWildcard {
		t.Fatalf("expected KleeneStar(ArrayWildcard), got %+v", q.Children[0])
	}
}
