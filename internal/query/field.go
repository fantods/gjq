package query

import (
	"strings"
	"unicode"
)

// FieldExpr matches a specific object key by name.
type FieldExpr struct {
	Name string
}

func NewField(name string) Query {
	return FieldExpr{Name: name}
}

func (FieldExpr) queryNode() {}

func (f FieldExpr) Depth() int { return 1 }

func (f FieldExpr) String() string {
	var b strings.Builder
	f.writeString(&b)
	return b.String()
}

func (f FieldExpr) writeString(b *strings.Builder) {
	if needsQuoting(f.Name) {
		b.WriteByte('"')
		escapeForQuotedFieldWrite(b, f.Name)
		b.WriteByte('"')
	} else {
		b.WriteString(f.Name)
	}
}

func needsQuoting(name string) bool {
	if name == "" {
		return true
	}
	for _, r := range name {
		if isReserved(r) || unicode.IsSpace(r) || r == '"' || r == '\\' {
			return true
		}
	}
	return false
}

func escapeForQuotedFieldWrite(b *strings.Builder, name string) {
	for _, r := range name {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		default:
			b.WriteRune(r)
		}
	}
}

func escapeForQuotedField(name string) string {
	var buf strings.Builder
	buf.Grow(len(name))
	escapeForQuotedFieldWrite(&buf, name)
	return buf.String()
}

// isBracketAccess checks if a string looks like an array access expression.
func isBracketAccess(s string) bool {
	if len(s) == 0 || s[0] != '[' {
		return false
	}
	return true
}
