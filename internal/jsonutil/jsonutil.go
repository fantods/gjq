// Package jsonutil provides JSON parsing into native Go types.
package jsonutil

import (
	"encoding/json"
)

func Parse(input string) (interface{}, error) {
	return ParseBytes([]byte(input))
}

func ParseBytes(data []byte) (interface{}, error) {
	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}
