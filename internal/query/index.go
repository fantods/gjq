package query

import (
	"strconv"
	"strings"
)

// IndexExpr matches a specific array index.
type IndexExpr struct {
	Index int
}

func NewIndex(idx int) Query {
	return IndexExpr{Index: idx}
}

func (IndexExpr) queryNode() {}

func (IndexExpr) Depth() int { return 1 }

func (i IndexExpr) String() string {
	var b strings.Builder
	i.writeString(&b)
	return b.String()
}

func (i IndexExpr) writeString(b *strings.Builder) {
	b.WriteByte('[')
	b.WriteString(strconv.Itoa(i.Index))
	b.WriteByte(']')
}

// RangeExpr matches array indices in the half-open interval [Start, End).
type RangeExpr struct {
	Start int
	End   int
}

func NewRange(start, end int) Query {
	return RangeExpr{Start: start, End: end}
}

func (RangeExpr) queryNode() {}

func (RangeExpr) Depth() int { return 1 }

func (r RangeExpr) String() string {
	var b strings.Builder
	r.writeString(&b)
	return b.String()
}

func (r RangeExpr) writeString(b *strings.Builder) {
	b.WriteByte('[')
	b.WriteString(strconv.Itoa(r.Start))
	b.WriteByte(':')
	b.WriteString(strconv.Itoa(r.End))
	b.WriteByte(']')
}

// RangeFromExpr matches array indices >= Start.
type RangeFromExpr struct {
	Start int
}

func NewRangeFrom(start int) Query {
	return RangeFromExpr{Start: start}
}

func (RangeFromExpr) queryNode() {}

func (RangeFromExpr) Depth() int { return 1 }

func (r RangeFromExpr) String() string {
	var b strings.Builder
	r.writeString(&b)
	return b.String()
}

func (r RangeFromExpr) writeString(b *strings.Builder) {
	b.WriteByte('[')
	b.WriteString(strconv.Itoa(r.Start))
	b.WriteString(":]")
}
