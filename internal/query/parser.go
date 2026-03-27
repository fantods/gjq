package query

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type QueryParseError struct {
	Message string
	Pos     int
}

func (e QueryParseError) Error() string {
	if e.Pos >= 0 {
		return fmt.Sprintf("parse error at position %d: %s", e.Pos, e.Message)
	}
	return fmt.Sprintf("parse error: %s", e.Message)
}

func ParseQuery(input string) (Query, error) {
	p := &parser{input: input, pos: 0}
	p.skipWhitespace()
	if p.pos >= len(p.input) {
		return NewSequence(nil), nil
	}
	q, err := p.parseDisjunction()
	if err != nil {
		return q, err
	}
	p.skipWhitespace()
	if p.pos < len(p.input) {
		return q, p.errorf("unexpected trailing character '%c'", p.peek())
	}
	return q, nil
}

type parser struct {
	input string
	pos   int
}

func (p *parser) peek() rune {
	if p.pos >= len(p.input) {
		return -1
	}
	r, _ := utf8.DecodeRuneInString(p.input[p.pos:])
	return r
}

func (p *parser) advance() rune {
	if p.pos >= len(p.input) {
		return -1
	}
	r, size := utf8.DecodeRuneInString(p.input[p.pos:])
	p.pos += size
	return r
}

func (p *parser) errorf(format string, args ...interface{}) QueryParseError {
	return QueryParseError{Message: fmt.Sprintf(format, args...), Pos: p.pos}
}

func (p *parser) skipWhitespace() {
	for p.pos < len(p.input) {
		r, size := utf8.DecodeRuneInString(p.input[p.pos:])
		if !unicode.IsSpace(r) {
			break
		}
		p.pos += size
	}
}

func (p *parser) parseDisjunction() (Query, error) {
	first, err := p.parseSequence()
	if err != nil {
		return first, err
	}

	var branches []Query
	branches = append(branches, first)

	for {
		p.skipWhitespace()
		if p.peek() != '|' {
			break
		}
		p.advance()
		p.skipWhitespace()
		next, err := p.parseSequence()
		if err != nil {
			return first, err
		}
		branches = append(branches, next)
	}

	if len(branches) == 1 {
		return branches[0], nil
	}
	return NewDisjunction(branches), nil
}

func (p *parser) parseSequence() (Query, error) {
	first, err := p.parseStep()
	if err != nil {
		return first, err
	}

	var steps []Query
	steps = append(steps, first)

	for {
		p.skipWhitespace()
		if p.peek() != '.' {
			break
		}
		p.advance()
		p.skipWhitespace()
		next, err := p.parseStep()
		if err != nil {
			return first, err
		}
		steps = append(steps, next)
	}

	if len(steps) == 1 {
		return steps[0], nil
	}
	return NewSequence(steps), nil
}

func (p *parser) parseStep() (Query, error) {
	var queries []Query

	r := p.peek()

	if r == '(' {
		q, err := p.parseGroup()
		if err != nil {
			return q, err
		}
		queries = append(queries, q)
	} else if r == '[' {
		q, err := p.parseBracket()
		if err != nil {
			return q, err
		}
		queries = append(queries, q)
	} else if r == '*' {
		p.advance()
		queries = append(queries, NewFieldWildcard())
	} else if r == '/' {
		q, err := p.parseRegex()
		if err != nil {
			return q, err
		}
		queries = append(queries, q)
	} else if r == '"' {
		q, err := p.parseQuotedField()
		if err != nil {
			return q, err
		}
		queries = append(queries, q)
	} else if isUnquotedFieldStart(r) {
		q, err := p.parseUnquotedField()
		if err != nil {
			return q, err
		}
		queries = append(queries, q)
	} else {
		return NewSequence(nil), p.errorf("unexpected character '%c'", r)
	}

	for p.peek() == '[' {
		q, err := p.parseBracket()
		if err != nil {
			return q, err
		}
		queries = append(queries, q)
	}

	if p.peek() == '*' || p.peek() == '?' {
		mod := p.advance()
		if len(queries) == 0 {
			return NewSequence(nil), p.errorf("modifier '%c' with no preceding query", mod)
		}
		last := queries[len(queries)-1]
		queries = queries[:len(queries)-1]
		switch mod {
		case '*':
			queries = append(queries, NewKleeneStar(last))
		case '?':
			queries = append(queries, NewOptional(last))
		}
	}

	if len(queries) == 1 {
		return queries[0], nil
	}
	return NewSequence(queries), nil
}

func (p *parser) parseGroup() (Query, error) {
	if p.peek() != '(' {
		return NewSequence(nil), p.errorf("expected '('")
	}
	p.advance()
	p.skipWhitespace()

	q, err := p.parseDisjunction()
	if err != nil {
		return q, err
	}

	p.skipWhitespace()
	if p.peek() != ')' {
		return q, p.errorf("expected ')'")
	}
	p.advance()

	return q, nil
}

func (p *parser) parseBracket() (Query, error) {
	if p.peek() != '[' {
		return NewSequence(nil), p.errorf("expected '['")
	}
	p.advance()
	p.skipWhitespace()

	if p.peek() == '*' {
		p.advance()
		p.skipWhitespace()
		if p.peek() != ']' {
			return NewSequence(nil), p.errorf("expected ']' after [*")
		}
		p.advance()
		return NewArrayWildcard(), nil
	}

	var start *int
	if p.peek() != ':' {
		n, err := p.parseNumber()
		if err != nil {
			return NewSequence(nil), err
		}
		start = &n
	}

	p.skipWhitespace()
	if p.peek() != ':' {
		if start != nil {
			p.skipWhitespace()
			if p.peek() == ']' {
				p.advance()
				return NewIndex(*start), nil
			}
			return NewSequence(nil), p.errorf("expected ']' or ':'")
		}
		return NewSequence(nil), p.errorf("expected ':'")
	}

	p.advance()
	p.skipWhitespace()

	var end *int
	if p.peek() != ']' {
		n, err := p.parseNumber()
		if err != nil {
			return NewSequence(nil), err
		}
		end = &n
	}

	p.skipWhitespace()
	if p.peek() != ']' {
		return NewSequence(nil), p.errorf("expected ']'")
	}
	p.advance()

	switch {
	case start == nil && end == nil:
		return NewArrayWildcard(), nil
	case start == nil && end != nil:
		return NewRange(0, *end), nil
	case start != nil && end == nil:
		return NewRangeFrom(*start), nil
	default:
		return NewRange(*start, *end), nil
	}
}

func (p *parser) parseNumber() (int, error) {
	start := p.pos
	for p.pos < len(p.input) && p.input[p.pos] >= '0' && p.input[p.pos] <= '9' {
		p.pos++
	}
	if p.pos == start {
		return 0, p.errorf("expected number")
	}
	return strconv.Atoi(p.input[start:p.pos])
}

func (p *parser) parseRegex() (Query, error) {
	if p.peek() != '/' {
		return NewSequence(nil), p.errorf("expected '/'")
	}
	p.advance()

	var buf strings.Builder
	for {
		r := p.peek()
		if r == -1 {
			return NewSequence(nil), p.errorf("unclosed regex")
		}
		if r == '/' {
			break
		}
		p.advance()
		if r == '\\' && p.peek() == '/' {
			buf.WriteRune('/')
			p.advance()
			continue
		}
		buf.WriteRune(r)
	}

	if p.peek() != '/' {
		return NewSequence(nil), p.errorf("unclosed regex")
	}
	p.advance()

	return NewRegex(buf.String()), nil
}

func (p *parser) parseQuotedField() (Query, error) {
	if p.peek() != '"' {
		return NewSequence(nil), p.errorf("expected '\"'")
	}
	p.advance()

	name, err := p.parseStringInner()
	if err != nil {
		return NewSequence(nil), err
	}

	if p.peek() != '"' {
		return NewSequence(nil), p.errorf("unclosed quoted field")
	}
	p.advance()

	return NewField(name), nil
}

func (p *parser) parseStringInner() (string, error) {
	var buf strings.Builder
	for {
		r := p.peek()
		if r == -1 {
			return "", p.errorf("unclosed string")
		}
		if r == '"' {
			break
		}
		p.advance()
		if r == '\\' {
			next := p.peek()
			switch next {
			case '"':
				buf.WriteRune('"')
				p.advance()
			case '\\':
				buf.WriteRune('\\')
				p.advance()
			case '/':
				buf.WriteRune('/')
				p.advance()
			case 'b':
				buf.WriteRune('\b')
				p.advance()
			case 'f':
				buf.WriteRune('\f')
				p.advance()
			case 'n':
				buf.WriteRune('\n')
				p.advance()
			case 'r':
				buf.WriteRune('\r')
				p.advance()
			case 't':
				buf.WriteRune('\t')
				p.advance()
			case 'u':
				p.advance()
				var hex string
				for i := 0; i < 4; i++ {
					c := p.peek()
					if !isHexDigit(c) {
						return "", p.errorf("expected 4 hex digits after \\u")
					}
					hex += string(c)
					p.advance()
				}
				cp, err := strconv.ParseInt(hex, 16, 32)
				if err != nil {
					return "", p.errorf("invalid unicode escape \\u%s", hex)
				}
				buf.WriteRune(rune(cp))
			case -1:
				buf.WriteRune('\\')
			default:
				buf.WriteRune('\\')
				buf.WriteRune(next)
				p.advance()
			}
		} else {
			buf.WriteRune(r)
		}
	}
	return buf.String(), nil
}

func (p *parser) parseUnquotedField() (Query, error) {
	start := p.pos
	for p.pos < len(p.input) {
		r, size := utf8.DecodeRuneInString(p.input[p.pos:])
		if isReserved(r) || unicode.IsSpace(r) || r == '"' {
			break
		}
		p.pos += size
	}
	if p.pos == start {
		return NewSequence(nil), p.errorf("expected field name")
	}
	return NewField(p.input[start:p.pos]), nil
}

func isReserved(r rune) bool {
	switch r {
	case '.', '|', '*', '?', '[', ']', '(', ')', '/':
		return true
	}
	return false
}

func isHexDigit(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}

func isUnquotedFieldStart(r rune) bool {
	return r != -1 && !isReserved(r) && !unicode.IsSpace(r) && r != '"'
}
