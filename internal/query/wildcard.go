package query

import "strings"

// WildcardExpr matches any single object key.
type WildcardExpr struct{}

func NewFieldWildcard() Query {
	return WildcardExpr{}
}

func (WildcardExpr) queryNode() {}

func (WildcardExpr) Depth() int { return 1 }

func (WildcardExpr) String() string { return "*" }

func (WildcardExpr) writeString(b *strings.Builder) { b.WriteByte('*') }

// ArrayWildExpr matches any array index.
type ArrayWildExpr struct{}

func NewArrayWildcard() Query {
	return ArrayWildExpr{}
}

func (ArrayWildExpr) queryNode() {}

func (ArrayWildExpr) Depth() int { return 1 }

func (ArrayWildExpr) String() string { return "[*]" }

func (ArrayWildExpr) writeString(b *strings.Builder) { b.WriteString("[*]") }
