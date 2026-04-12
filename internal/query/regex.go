package query

import "strings"

// RegexExpr matches object keys against a regular expression.
type RegexExpr struct {
	Pattern string
}

func NewRegex(pattern string) Query {
	return RegexExpr{Pattern: pattern}
}

func (RegexExpr) queryNode() {}

func (RegexExpr) Depth() int { return 1 }

func (r RegexExpr) String() string {
	var b strings.Builder
	r.writeString(&b)
	return b.String()
}

func (r RegexExpr) writeString(b *strings.Builder) {
	b.WriteByte('/')
	b.WriteString(r.Pattern)
	b.WriteByte('/')
}
