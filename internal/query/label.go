package query

import (
	"strconv"
	"strings"
)

// LabelKind classifies the type of a transition label in the DFA alphabet.
type LabelKind int

const (
	LabelField         LabelKind = iota // Matches a specific object key
	LabelFieldWildcard                  // Matches any object key
	LabelRange                          // Matches array indices in [RangeLo, RangeHi)
	LabelOther                          // Matches keys/indices not in the query
)

// TransitionLabel represents one symbol in the DFA alphabet.
type TransitionLabel struct {
	Kind    LabelKind
	Field   string // Valid when Kind == LabelField
	RangeLo int    // Valid when Kind == LabelRange
	RangeHi int    // Valid when Kind == LabelRange
}

func (l TransitionLabel) String() string {
	var b strings.Builder
	switch l.Kind {
	case LabelField:
		b.WriteString("Field(")
		b.WriteString(l.Field)
		b.WriteByte(')')
	case LabelFieldWildcard:
		b.WriteString("FieldWildcard")
	case LabelRange:
		b.WriteString("Range(")
		b.WriteString(strconv.Itoa(l.RangeLo))
		b.WriteString(", ")
		b.WriteString(strconv.Itoa(l.RangeHi))
		b.WriteByte(')')
	case LabelOther:
		b.WriteString("Other")
	default:
		b.WriteString("Unknown")
	}
	return b.String()
}
