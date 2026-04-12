// Package jsonutil provides JSON parsing with automatic number conversion.
// It decodes JSON using json.Number and converts numeric values to int or float64
// so callers don't need to handle json.Number strings.
package jsonutil

import (
	"bytes"
	"encoding/json"
	"strings"
)

// Parse decodes a JSON string and converts json.Number values to int or float64.
func Parse(input string) (interface{}, error) {
	var result interface{}
	dec := json.NewDecoder(strings.NewReader(input))
	dec.UseNumber()
	if err := dec.Decode(&result); err != nil {
		return nil, err
	}
	return convertNumbers(result), nil
}

// ParseBytes decodes JSON bytes and converts json.Number values to int or float64.
func ParseBytes(data []byte) (interface{}, error) {
	var result interface{}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&result); err != nil {
		return nil, err
	}
	return convertNumbers(result), nil
}

// ConvertNumbers recursively converts json.Number values in a decoded JSON
// structure to int64 or float64. Maps and slices are converted recursively.
func ConvertNumbers(v interface{}) interface{} {
	return convertNumbers(v)
}

func convertNumbers(v interface{}) interface{} {
	switch val := v.(type) {
	case json.Number:
		if i, err := val.Int64(); err == nil {
			return int(i)
		}
		if f, err := val.Float64(); err == nil {
			return f
		}
		return val.String()
	case map[string]interface{}:
		m := make(map[string]interface{}, len(val))
		for k, v := range val {
			m[k] = convertNumbers(v)
		}
		return m
	case []interface{}:
		a := make([]interface{}, len(val))
		for i, v := range val {
			a[i] = convertNumbers(v)
		}
		return a
	default:
		return v
	}
}
