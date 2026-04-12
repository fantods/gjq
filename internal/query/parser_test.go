package query

import "testing"

func mustAssertField(t *testing.T, q Query, name string) {
	t.Helper()
	f, ok := q.(FieldExpr)
	if !ok {
		t.Fatalf("expected FieldExpr, got %T", q)
	}
	if f.Name != name {
		t.Fatalf("expected field %q, got %q", name, f.Name)
	}
}

func mustAssertIndex(t *testing.T, q Query, idx int) {
	t.Helper()
	i, ok := q.(IndexExpr)
	if !ok {
		t.Fatalf("expected IndexExpr, got %T", q)
	}
	if i.Index != idx {
		t.Fatalf("expected index %d, got %d", idx, i.Index)
	}
}

func mustAssertRange(t *testing.T, q Query, start, end int) {
	t.Helper()
	r, ok := q.(RangeExpr)
	if !ok {
		t.Fatalf("expected RangeExpr, got %T", q)
	}
	if r.Start != start || r.End != end {
		t.Fatalf("expected range [%d,%d), got [%d,%d)", start, end, r.Start, r.End)
	}
}

func mustAssertRangeFrom(t *testing.T, q Query, start int) {
	t.Helper()
	r, ok := q.(RangeFromExpr)
	if !ok {
		t.Fatalf("expected RangeFromExpr, got %T", q)
	}
	if r.Start != start {
		t.Fatalf("expected rangeFrom %d, got %d", start, r.Start)
	}
}

func mustAssertSeq(t *testing.T, q Query) SeqExpr {
	t.Helper()
	s, ok := q.(SeqExpr)
	if !ok {
		t.Fatalf("expected SeqExpr, got %T", q)
	}
	return s
}

func mustAssertDisj(t *testing.T, q Query) DisjExpr {
	t.Helper()
	d, ok := q.(DisjExpr)
	if !ok {
		t.Fatalf("expected DisjExpr, got %T", q)
	}
	return d
}

func mustAssertOpt(t *testing.T, q Query) OptionalExpr {
	t.Helper()
	o, ok := q.(OptionalExpr)
	if !ok {
		t.Fatalf("expected OptionalExpr, got %T", q)
	}
	return o
}

func mustAssertStar(t *testing.T, q Query) StarExpr {
	t.Helper()
	s, ok := q.(StarExpr)
	if !ok {
		t.Fatalf("expected StarExpr, got %T", q)
	}
	return s
}

func TestParseEmpty(t *testing.T) {
	q, err := ParseQuery("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := mustAssertSeq(t, q)
	if len(s.Steps) != 0 {
		t.Fatalf("expected empty sequence, got %d steps", len(s.Steps))
	}
}

func TestParseField(t *testing.T) {
	q, err := ParseQuery("foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mustAssertField(t, q, "foo")
}

func TestParseQuotedField(t *testing.T) {
	q, err := ParseQuery(`"foo bar"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mustAssertField(t, q, "foo bar")
}

func TestParseIndex(t *testing.T) {
	q, err := ParseQuery("[3]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mustAssertIndex(t, q, 3)
}

func TestParseRange(t *testing.T) {
	q, err := ParseQuery("[2:5]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mustAssertRange(t, q, 2, 5)
}

func TestParseRangeOpenEnd(t *testing.T) {
	q, err := ParseQuery("[3:]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mustAssertRangeFrom(t, q, 3)
}

func TestParseRangeOpenStart(t *testing.T) {
	q, err := ParseQuery("[:5]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mustAssertRange(t, q, 0, 5)
}

func TestParseArrayWildcard(t *testing.T) {
	q, err := ParseQuery("[*]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := q.(ArrayWildExpr); !ok {
		t.Fatalf("expected ArrayWildExpr, got %T", q)
	}
}

func TestParseRangeAll(t *testing.T) {
	q, err := ParseQuery("[:]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := q.(ArrayWildExpr); !ok {
		t.Fatalf("expected ArrayWildExpr for [:], got %T", q)
	}
}

func TestParseFieldWildcard(t *testing.T) {
	q, err := ParseQuery("*")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := q.(WildcardExpr); !ok {
		t.Fatalf("expected WildcardExpr, got %T", q)
	}
}

func TestParseRegex(t *testing.T) {
	q, err := ParseQuery("/foo.bar/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r, ok := q.(RegexExpr)
	if !ok {
		t.Fatalf("expected RegexExpr, got %T", q)
	}
	if r.Pattern != "foo.bar" {
		t.Fatalf("expected pattern 'foo.bar', got %q", r.Pattern)
	}
}

func TestParseRegexEscapedSlash(t *testing.T) {
	q, err := ParseQuery(`/foo\/bar/`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r, ok := q.(RegexExpr)
	if !ok {
		t.Fatalf("expected RegexExpr, got %T", q)
	}
	if r.Pattern != "foo/bar" {
		t.Fatalf("expected pattern 'foo/bar', got %q", r.Pattern)
	}
}

func TestParseSequence(t *testing.T) {
	q, err := ParseQuery("foo.bar")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := mustAssertSeq(t, q)
	if len(s.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(s.Steps))
	}
	mustAssertField(t, s.Steps[0], "foo")
	mustAssertField(t, s.Steps[1], "bar")
}

func TestParseDisjunction(t *testing.T) {
	q, err := ParseQuery("foo | bar")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d := mustAssertDisj(t, q)
	if len(d.Branches) != 2 {
		t.Fatalf("expected 2 branches, got %d", len(d.Branches))
	}
	mustAssertField(t, d.Branches[0], "foo")
	mustAssertField(t, d.Branches[1], "bar")
}

func TestParseDisjunctionSingle(t *testing.T) {
	q, err := ParseQuery("foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := q.(DisjExpr); ok {
		t.Fatalf("single element should not be wrapped in DisjExpr")
	}
}

func TestParseOptional(t *testing.T) {
	q, err := ParseQuery("foo?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	o := mustAssertOpt(t, q)
	mustAssertField(t, o.Child, "foo")
}

func TestParseKleeneStar(t *testing.T) {
	q, err := ParseQuery("a*")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := mustAssertStar(t, q)
	mustAssertField(t, s.Child, "a")
}

func TestParseFieldWithIndex(t *testing.T) {
	q, err := ParseQuery("foo[3]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := mustAssertSeq(t, q)
	if len(s.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(s.Steps))
	}
	mustAssertField(t, s.Steps[0], "foo")
	mustAssertIndex(t, s.Steps[1], 3)
}

func TestParseFieldWithArrayWildcard(t *testing.T) {
	q, err := ParseQuery("foo[*]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := mustAssertSeq(t, q)
	if len(s.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(s.Steps))
	}
	if _, ok := s.Steps[1].(ArrayWildExpr); !ok {
		t.Fatalf("expected ArrayWildExpr, got %T", s.Steps[1])
	}
}

func TestParseFieldWithRange(t *testing.T) {
	q, err := ParseQuery("foo[2:5]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := mustAssertSeq(t, q)
	if len(s.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(s.Steps))
	}
	mustAssertRange(t, s.Steps[1], 2, 5)
}

func TestParseComplexQuery(t *testing.T) {
	q, err := ParseQuery("foo.bar[0]?.baz*")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := mustAssertSeq(t, q)
	if len(s.Steps) != 3 {
		t.Fatalf("expected 3 steps, got %d: %+v", len(s.Steps), s)
	}
	mustAssertField(t, s.Steps[0], "foo")

	// bar[0] is parsed as a SeqExpr{bar, [0]?}
	inner := mustAssertSeq(t, s.Steps[1])
	if len(inner.Steps) != 2 {
		t.Fatalf("expected 2 inner steps, got %d", len(inner.Steps))
	}
	mustAssertField(t, inner.Steps[0], "bar")
	opt := mustAssertOpt(t, inner.Steps[1])
	mustAssertIndex(t, opt.Child, 0)

	star := mustAssertStar(t, s.Steps[2])
	mustAssertField(t, star.Child, "baz")
}

func TestParseGroup(t *testing.T) {
	q, err := ParseQuery("(foo | bar).baz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := mustAssertSeq(t, q)
	if len(s.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(s.Steps))
	}
	mustAssertDisj(t, s.Steps[0])
	mustAssertField(t, s.Steps[1], "baz")
}

func TestParseAnyPathGroup(t *testing.T) {
	q, err := ParseQuery("(* | [*])*")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	star := mustAssertStar(t, q)
	d := mustAssertDisj(t, star.Child)
	if len(d.Branches) != 2 {
		t.Fatalf("expected 2 branches, got %d", len(d.Branches))
	}
	if _, ok := d.Branches[0].(WildcardExpr); !ok {
		t.Fatalf("expected first branch WildcardExpr, got %T", d.Branches[0])
	}
	if _, ok := d.Branches[1].(ArrayWildExpr); !ok {
		t.Fatalf("expected second branch ArrayWildExpr, got %T", d.Branches[1])
	}
}

func TestParseNestedGroups(t *testing.T) {
	q, err := ParseQuery("((foo.bar)* | bar)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d := mustAssertDisj(t, q)
	if len(d.Branches) != 2 {
		t.Fatalf("expected 2 branches, got %d", len(d.Branches))
	}
	mustAssertStar(t, d.Branches[0])
}

func TestParseGroupSequence(t *testing.T) {
	q, err := ParseQuery("(foo.bar.baz)?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	o := mustAssertOpt(t, q)
	inner := mustAssertSeq(t, o.Child)
	if len(inner.Steps) != 3 {
		t.Fatalf("expected 3 steps inside Optional, got %d", len(inner.Steps))
	}
}

func TestParseNestedGroupsTrivial(t *testing.T) {
	q, err := ParseQuery("((foo))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mustAssertField(t, q, "foo")
}

func TestParseMultipleOptional(t *testing.T) {
	q, err := ParseQuery("c*.c?.c?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := mustAssertSeq(t, q)
	if len(s.Steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(s.Steps))
	}
}

func TestParseAnyPathInQuery(t *testing.T) {
	q, err := ParseQuery("a.(* | [*])*.b?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := mustAssertSeq(t, q)
	if len(s.Steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(s.Steps))
	}
	mustAssertField(t, s.Steps[0], "a")
	mustAssertStar(t, s.Steps[1])
	mustAssertOpt(t, s.Steps[2])
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
	mustAssertField(t, q, ".|*?[]()/")
}

func TestParseQuotedFieldInSequence(t *testing.T) {
	q, err := ParseQuery(`paths."/activities"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := mustAssertSeq(t, q)
	if len(s.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(s.Steps))
	}
	mustAssertField(t, s.Steps[0], "paths")
	mustAssertField(t, s.Steps[1], "/activities")
}

func TestParseQuotedFieldWithSlash(t *testing.T) {
	q, err := ParseQuery(`"/activities"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mustAssertField(t, q, "/activities")
}

func TestParseQuotedFieldWithDot(t *testing.T) {
	q, err := ParseQuery(`"a.b"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mustAssertField(t, q, "a.b")
}

func TestParseQuotedFieldUnescapeBackslash(t *testing.T) {
	q, err := ParseQuery(`"a\\b"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mustAssertField(t, q, "a\\b")
}

func TestParseQuotedFieldUnescapeInnerQuote(t *testing.T) {
	q, err := ParseQuery(`"a\"b"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	f, ok := q.(FieldExpr)
	if !ok {
		t.Fatalf("expected FieldExpr, got %T", q)
	}
	if f.Name != `a"b` {
		t.Fatalf("expected 'a\"b', got %q", f.Name)
	}
}

func TestParseQuotedFieldUnescapeUnicode(t *testing.T) {
	q, err := ParseQuery(`"\u0041"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mustAssertField(t, q, "A")
}

func TestParseQuotedFieldUnescapeEscapeSequences(t *testing.T) {
	q, err := ParseQuery(`"\n\r\t\b\f"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	f, ok := q.(FieldExpr)
	if !ok {
		t.Fatalf("expected FieldExpr, got %T", q)
	}
	if f.Name != "\n\r\t\b\f" {
		t.Fatalf("expected escape sequences, got %q", f.Name)
	}
}

func TestParseGroupAnyReservedCharsInDoubleQuotes(t *testing.T) {
	q, err := ParseQuery(`("." | "|" | "*" | "?" | "[" | "]" | "(" | ")" | "/")*`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	star := mustAssertStar(t, q)
	d := mustAssertDisj(t, star.Child)
	if len(d.Branches) != 9 {
		t.Fatalf("expected 9 branches, got %d", len(d.Branches))
	}
}

func TestParseWhitespaceAroundPipe(t *testing.T) {
	q, err := ParseQuery("foo|bar")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d := mustAssertDisj(t, q)
	if len(d.Branches) != 2 {
		t.Fatalf("expected 2 branches, got %d", len(d.Branches))
	}
}

func TestParseThreeWayDisjunction(t *testing.T) {
	q, err := ParseQuery("a | b | c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d := mustAssertDisj(t, q)
	if len(d.Branches) != 3 {
		t.Fatalf("expected 3 branches, got %d", len(d.Branches))
	}
}

func TestParseFieldWithAlphanumeric(t *testing.T) {
	q, err := ParseQuery("foo123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mustAssertField(t, q, "foo123")
}

func TestParseFieldWithNumberAndIndex(t *testing.T) {
	q, err := ParseQuery("foo123[42]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := mustAssertSeq(t, q)
	if len(s.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(s.Steps))
	}
	mustAssertField(t, s.Steps[0], "foo123")
	mustAssertIndex(t, s.Steps[1], 42)
}

func TestParseStandaloneIndexOptional(t *testing.T) {
	q, err := ParseQuery("[3]?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	o := mustAssertOpt(t, q)
	mustAssertIndex(t, o.Child, 3)
}

func TestParseWildcardOptional(t *testing.T) {
	q, err := ParseQuery("*?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	o := mustAssertOpt(t, q)
	if _, ok := o.Child.(WildcardExpr); !ok {
		t.Fatalf("expected Optional(WildcardExpr), got %T", o.Child)
	}
}

func TestParseArrayWildcardKleeneStar(t *testing.T) {
	q, err := ParseQuery("[*]*")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := mustAssertStar(t, q)
	if _, ok := s.Child.(ArrayWildExpr); !ok {
		t.Fatalf("expected StarExpr(ArrayWildExpr), got %T", s.Child)
	}
}
