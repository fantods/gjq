package query

import "strconv"

// PathTypeKind distinguishes between field and index path components.
type PathTypeKind int

const (
	PathIndex PathTypeKind = iota
	PathField
)

// PathType represents one component of a JSON pointer path.
type PathType struct {
	Kind  PathTypeKind
	Field string // Valid when Kind == PathField
	Index int    // Valid when Kind == PathIndex
}

func (p PathType) String() string {
	if p.Kind == PathIndex {
		var buf [16]byte
		b := strconv.AppendInt(buf[:0], int64(p.Index), 10)
		result := make([]byte, 0, 2+len(b))
		result = append(result, '[')
		result = append(result, b...)
		result = append(result, ']')
		return string(result)
	}
	return p.Field
}

// JSONPointer represents a matched path through a JSON document.
type JSONPointer struct {
	Path  []PathType
	Value interface{}
}
