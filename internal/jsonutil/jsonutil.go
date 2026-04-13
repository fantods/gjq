// Package jsonutil provides JSON parsing into native Go types.
package jsonutil

import (
	"bytes"
	"encoding/json"
)

// Parse decodes a JSON string into native Go types.
// Numbers are represented as float64 by the standard decoder.
func Parse(input string) (interface{}, error) {
	var result interface{}
	dec := json.NewDecoder(bytes.NewReader([]byte(input)))
	if err := dec.Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// ParseBytes decodes JSON bytes into native Go types.
// Numbers are represented as float64 by the standard decoder.
func ParseBytes(data []byte) (interface{}, error) {
	var result interface{}
	dec := json.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}


