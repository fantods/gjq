package query

import (
	"fmt"
)

type LabelKind int

const (
	LabelField LabelKind = iota
	LabelFieldWildcard
	LabelRange
	LabelRangeFrom
	LabelOther
)

type TransitionLabel struct {
	Kind    LabelKind
	Field   string
	RangeLo int
	RangeHi int
}

func (l TransitionLabel) String() string {
	switch l.Kind {
	case LabelField:
		return fmt.Sprintf("Field(%s)", l.Field)
	case LabelFieldWildcard:
		return "FieldWildcard"
	case LabelRange:
		return fmt.Sprintf("Range(%d, %d)", l.RangeLo, l.RangeHi)
	case LabelRangeFrom:
		return fmt.Sprintf("RangeFrom(%d)", l.RangeLo)
	case LabelOther:
		return "Other"
	default:
		return "Unknown"
	}
}

type PathTypeKind int

const (
	PathIndex PathTypeKind = iota
	PathField
)

type PathType struct {
	Kind  PathTypeKind
	Field string
	Index int
}

func (p PathType) String() string {
	switch p.Kind {
	case PathIndex:
		return fmt.Sprintf("[%d]", p.Index)
	case PathField:
		return p.Field
	default:
		return ""
	}
}

type JSONPointer struct {
	Path  []PathType
	Value interface{}
}
