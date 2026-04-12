package jsonutil

import (
	"fmt"
	"testing"
)

func TestParseString(t *testing.T) {
	result, err := Parse(`{"a": 1, "b": "hello"}`)
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if v, ok := m["a"].(int); !ok || v != 1 {
		t.Fatalf("expected int 1, got %v", m["a"])
	}
	if v, ok := m["b"].(string); !ok || v != "hello" {
		t.Fatalf("expected string 'hello', got %v", m["b"])
	}
}

func TestParseBytes(t *testing.T) {
	result, err := ParseBytes([]byte(`{"x": 42}`))
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if v, ok := m["x"].(int); !ok || v != 42 {
		t.Fatalf("expected int 42, got %v", m["x"])
	}
}

func TestParseEquivalence(t *testing.T) {
	inputs := []string{
		`{"a": 1, "b": "hello"}`,
		`[1, 2, 3, "four", null, true, false]`,
		`{"nested": {"deep": {"val": 42.5}}}`,
		`{"num": 42, "float": 3.14, "big": 999999999999}`,
		`null`,
		`42`,
		`"hello"`,
		`[]`,
		`{}`,
	}
	for _, input := range inputs {
		resultStr, err := Parse(input)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", input, err)
		}
		resultBytes, err := ParseBytes([]byte(input))
		if err != nil {
			t.Fatalf("ParseBytes(%q) error: %v", input, err)
		}
		if fmt.Sprintf("%v", resultStr) != fmt.Sprintf("%v", resultBytes) {
			t.Errorf("Parse and ParseBytes returned different results for %q:\n  string: %v\n  bytes:  %v", input, resultStr, resultBytes)
		}
	}
}

func TestNumberConversion(t *testing.T) {
	input := `{"int_val": 42, "float_val": 3.14, "big_int": 999999999999}`
	result, err := ParseBytes([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if v, ok := m["int_val"].(int); !ok || v != 42 {
		t.Fatalf("expected int 42, got %v", m["int_val"])
	}
	if v, ok := m["float_val"].(float64); !ok || v != 3.14 {
		t.Fatalf("expected float64 3.14, got %v", m["float_val"])
	}
}

func TestParseNull(t *testing.T) {
	result, err := Parse("null")
	if err != nil {
		t.Fatal(err)
	}
	if result != nil {
		t.Fatalf("expected nil, got %v", result)
	}
}

func TestParseArray(t *testing.T) {
	result, err := Parse(`[1, "two", 3.0, null, true]`)
	if err != nil {
		t.Fatal(err)
	}
	arr := result.([]interface{})
	if len(arr) != 5 {
		t.Fatalf("expected 5 elements, got %d", len(arr))
	}
}

func TestParseNestedObject(t *testing.T) {
	input := `{"outer": {"inner": {"value": 42}}}`
	result, err := Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	outer := result.(map[string]interface{})["outer"].(map[string]interface{})
	inner := outer["inner"].(map[string]interface{})
	if v, ok := inner["value"].(int); !ok || v != 42 {
		t.Fatalf("expected int 42, got %v", inner["value"])
	}
}

func TestParseInvalidJSON(t *testing.T) {
	_, err := Parse("{invalid}")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func BenchmarkParseString(b *testing.B) {
	var buf string
	buf = `{"items": [`
	for i := 0; i < 1000; i++ {
		if i > 0 {
			buf += ", "
		}
		buf += fmt.Sprintf(`{"id": %d, "name": "item_%d", "value": %.1f}`, i, i, float64(i)*1.1)
	}
	buf += `]}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Parse(buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseBytes(b *testing.B) {
	var buf []byte
	buf = append(buf, `{"items": [`...)
	for i := 0; i < 1000; i++ {
		if i > 0 {
			buf = append(buf, ", "...)
		}
		buf = append(buf, fmt.Sprintf(`{"id": %d, "name": "item_%d", "value": %.1f}`, i, i, float64(i)*1.1)...)
	}
	buf = append(buf, `]}`...)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseBytes(buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}
