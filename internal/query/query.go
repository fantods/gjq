package query

// Query represents a parsed query expression. It is sealed — only types
// within this package can implement it.
type Query interface {
	queryNode()
	Depth() int
	String() string
}
