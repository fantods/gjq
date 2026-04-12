package query

import (
	"strings"
)

// OptionalExpr makes its child query optional (matches 0 or 1 times).
type OptionalExpr struct {
	Child Query
}

func NewOptional(q Query) Query {
	return OptionalExpr{Child: q}
}

func (OptionalExpr) queryNode() {}

func (o OptionalExpr) Depth() int { return 1 + o.Child.Depth() }

func (o OptionalExpr) String() string {
	var b strings.Builder
	o.writeString(&b)
	return b.String()
}

func (o OptionalExpr) writeString(b *strings.Builder) {
	if needsParens(o.Child) {
		b.WriteByte('(')
		o.Child.(stringWriter).writeString(b)
		b.WriteByte(')')
	} else {
		o.Child.(stringWriter).writeString(b)
	}
	b.WriteByte('?')
}

// StarExpr applies the Kleene star to its child (matches 0 or more times).
type StarExpr struct {
	Child Query
}

func NewKleeneStar(q Query) Query {
	return StarExpr{Child: q}
}

func (StarExpr) queryNode() {}

func (s StarExpr) Depth() int { return 1 + s.Child.Depth() }

func (s StarExpr) String() string {
	var b strings.Builder
	s.writeString(&b)
	return b.String()
}

func (s StarExpr) writeString(b *strings.Builder) {
	if needsParens(s.Child) {
		b.WriteByte('(')
		s.Child.(stringWriter).writeString(b)
		b.WriteByte(')')
	} else {
		s.Child.(stringWriter).writeString(b)
	}
	b.WriteByte('*')
}

// DisjExpr matches if any of its branches match.
type DisjExpr struct {
	Branches []Query
}

func NewDisjunction(branches []Query) Query {
	if len(branches) == 1 {
		return branches[0]
	}
	return DisjExpr{Branches: branches}
}

func (DisjExpr) queryNode() {}

func (d DisjExpr) Depth() int {
	max := 0
	for _, b := range d.Branches {
		if depth := b.Depth(); depth > max {
			max = depth
		}
	}
	return 1 + max
}

func (d DisjExpr) String() string {
	var b strings.Builder
	d.writeString(&b)
	return b.String()
}

func (d DisjExpr) writeString(b *strings.Builder) {
	for i, branch := range d.Branches {
		if i > 0 {
			b.WriteString(" | ")
		}
		branch.(stringWriter).writeString(b)
	}
}

// SeqExpr matches a sequence of query steps in order.
type SeqExpr struct {
	Steps []Query
}

func NewSequence(steps []Query) Query {
	return SeqExpr{Steps: steps}
}

func (SeqExpr) queryNode() {}

func (s SeqExpr) Depth() int {
	if len(s.Steps) == 0 {
		return 0
	}
	sum := 0
	for _, step := range s.Steps {
		sum += step.Depth()
	}
	return sum
}

func (s SeqExpr) String() string {
	var b strings.Builder
	s.writeString(&b)
	return b.String()
}

func (s SeqExpr) writeString(b *strings.Builder) {
	for i, step := range s.Steps {
		if i > 0 && !needsNoDotBefore(step) {
			b.WriteByte('.')
		}
		switch v := step.(type) {
		case SeqExpr:
			// Flatten nested sequences
			v.writeString(b)
		case DisjExpr:
			if len(v.Branches) > 1 {
				b.WriteByte('(')
				v.writeString(b)
				b.WriteByte(')')
			} else {
				v.writeString(b)
			}
		default:
			v.(stringWriter).writeString(b)
		}
	}
}

// stringWriter is an unexported interface used internally to avoid
// allocating separate strings.Builder instances for each nested node.
type stringWriter interface {
	writeString(b *strings.Builder)
}

// needsParens returns true if a query needs parentheses when used
// as the child of OptionalExpr or StarExpr.
func needsParens(q Query) bool {
	switch q.(type) {
	case SeqExpr:
		return len(q.(SeqExpr).Steps) > 1
	case DisjExpr:
		return true
	default:
		return false
	}
}

// needsNoDotBefore returns true if the query renders starting with '[' or '*'.
func needsNoDotBefore(q Query) bool {
	switch v := q.(type) {
	case IndexExpr, RangeExpr, RangeFromExpr, ArrayWildExpr, WildcardExpr:
		return true
	case OptionalExpr:
		return needsNoDotBefore(v.Child)
	case StarExpr:
		return needsNoDotBefore(v.Child)
	default:
		return false
	}
}

// unwrapModifier strips an outer OptionalExpr or StarExpr and returns the inner query.
func unwrapModifier(q Query) Query {
	if opt, ok := q.(OptionalExpr); ok {
		return opt.Child
	}
	if star, ok := q.(StarExpr); ok {
		return star.Child
	}
	return q
}
