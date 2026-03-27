package query

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
