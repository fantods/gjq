package query

import (
	"fmt"
	"strings"
	"unicode"
)

type QueryKind int

const (
	QueryField QueryKind = iota
	QueryIndex
	QueryRange
	QueryRangeFrom
	QueryFieldWildcard
	QueryArrayWildcard
	QueryRegex
	QueryOptional
	QueryKleeneStar
	QueryDisjunction
	QuerySequence
)

type Query struct {
	Kind     QueryKind
	Field    string
	Index    int
	RangeEnd int
	Regex    string
	Children []Query
}

func NewField(name string) Query {
	return Query{Kind: QueryField, Field: name}
}

func NewIndex(idx int) Query {
	return Query{Kind: QueryIndex, Index: idx}
}

func NewRange(start, end int) Query {
	return Query{Kind: QueryRange, Index: start, RangeEnd: end}
}

func NewRangeFrom(start int) Query {
	return Query{Kind: QueryRangeFrom, Index: start}
}

func NewFieldWildcard() Query {
	return Query{Kind: QueryFieldWildcard}
}

func NewArrayWildcard() Query {
	return Query{Kind: QueryArrayWildcard}
}

func NewRegex(pattern string) Query {
	return Query{Kind: QueryRegex, Regex: pattern}
}

func NewOptional(child Query) Query {
	return Query{Kind: QueryOptional, Children: []Query{child}}
}

func NewKleeneStar(child Query) Query {
	return Query{Kind: QueryKleeneStar, Children: []Query{child}}
}

func NewDisjunction(children []Query) Query {
	return Query{Kind: QueryDisjunction, Children: children}
}

func NewSequence(children []Query) Query {
	return Query{Kind: QuerySequence, Children: children}
}

func (q Query) Depth() int {
	switch q.Kind {
	case QueryDisjunction:
		max := 0
		for _, c := range q.Children {
			if d := c.Depth(); d > max {
				max = d
			}
		}
		return 1 + max
	case QuerySequence:
		sum := 0
		for _, c := range q.Children {
			sum += c.Depth()
		}
		return sum
	case QueryOptional, QueryKleeneStar:
		return 1 + q.Children[0].Depth()
	default:
		return 1
	}
}

func (q Query) String() string {
	switch q.Kind {
	case QueryField:
		if needsQuoting(q.Field) {
			return fmt.Sprintf(`"%s"`, escapeForQuotedField(q.Field))
		}
		return q.Field
	case QueryIndex:
		return fmt.Sprintf("[%d]", q.Index)
	case QueryRange:
		return fmt.Sprintf("[%d:%d]", q.Index, q.RangeEnd)
	case QueryRangeFrom:
		return fmt.Sprintf("[%d:]", q.Index)
	case QueryFieldWildcard:
		return "*"
	case QueryArrayWildcard:
		return "[*]"
	case QueryRegex:
		return fmt.Sprintf("/%s/", q.Regex)
	case QueryOptional:
		inner := q.Children[0]
		if inner.Kind == QueryDisjunction || inner.Kind == QuerySequence {
			if len(inner.Children) > 1 {
				return fmt.Sprintf("(%s)?", inner.String())
			}
		}
		return inner.String() + "?"
	case QueryKleeneStar:
		inner := q.Children[0]
		if inner.Kind == QueryDisjunction || inner.Kind == QuerySequence {
			if len(inner.Children) > 1 {
				return fmt.Sprintf("(%s)*", inner.String())
			}
		}
		return inner.String() + "*"
	case QueryDisjunction:
		parts := make([]string, len(q.Children))
		for i, c := range q.Children {
			parts[i] = c.String()
		}
		return strings.Join(parts, " | ")
	case QuerySequence:
		if len(q.Children) == 0 {
			return ""
		}
		var buf strings.Builder
		for i, child := range q.Children {
			if i > 0 {
				inner := unwrapModifier(child)
				prevInner := unwrapModifier(q.Children[i-1])
				if prevInner.Kind != QueryField || !isBracketAccess(inner) {
					buf.WriteByte('.')
				}
			}
			if child.Kind == QueryDisjunction {
				buf.WriteByte('(')
				buf.WriteString(child.String())
				buf.WriteByte(')')
			} else {
				buf.WriteString(child.String())
			}
		}
		return buf.String()
	default:
		return ""
	}
}

func unwrapModifier(q Query) Query {
	if q.Kind == QueryOptional || q.Kind == QueryKleeneStar {
		return q.Children[0]
	}
	return q
}

func isBracketAccess(q Query) bool {
	switch q.Kind {
	case QueryIndex, QueryRange, QueryRangeFrom, QueryFieldWildcard, QueryArrayWildcard:
		return true
	}
	return false
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

func escapeForQuotedField(name string) string {
	var buf strings.Builder
	buf.Grow(len(name))
	for _, r := range name {
		switch r {
		case '"':
			buf.WriteString(`\"`)
		case '\\':
			buf.WriteString(`\\`)
		default:
			buf.WriteRune(r)
		}
	}
	return buf.String()
}
