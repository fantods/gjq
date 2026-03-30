package output

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/fantods/gjq/internal/query"
)

func TestDepth(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  int
	}{
		{"nil", nil, 1},
		{"bool", true, 1},
		{"int", int(42), 1},
		{"float", float64(3.14), 1},
		{"string", "hello", 1},
		{"empty array", []interface{}{}, 1},
		{"flat array", []interface{}{1, "two", true}, 2},
		{"empty object", map[string]interface{}{}, 1},
		{"flat object", map[string]interface{}{"a": 1, "b": "two"}, 2},
		{"nested array", []interface{}{[]interface{}{1}}, 3},
		{"nested object", map[string]interface{}{"a": map[string]interface{}{"b": 1}}, 3},
		{"deep nesting", map[string]interface{}{
			"a": []interface{}{
				map[string]interface{}{
					"b": []interface{}{1},
				},
			},
		}, 5},
		{"mixed", map[string]interface{}{
			"x": []interface{}{
				1,
				map[string]interface{}{"y": nil},
			},
		}, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Depth(tt.input)
			if got != tt.want {
				t.Errorf("Depth() = %d, want %d", got, tt.want)
			}
		})
	}
}

func stripAnsi(s string) string {
	var out strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && s[j] != 'm' {
				j++
			}
			i = j + 1
		} else {
			out.WriteByte(s[i])
			i++
		}
	}
	return out.String()
}

func TestWriteColoredJSON_primitives(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  string
	}{
		{"null", nil, "null"},
		{"true", true, "true"},
		{"false", false, "false"},
		{"int", int(42), "42"},
		{"float", float64(3.14), "3.14"},
		{"negative int", int(-7), "-7"},
		{"negative float", float64(-0.5), "-0.5"},
		{"string", "hello", "\"hello\""},
		{"string with quotes", `he said "hi"`, `"he said \"hi\""`},
		{"string with newline", "line1\nline2", `"line1\nline2"`},
		{"empty string", "", `""`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := writeColoredJSON(&buf, tt.input, 0, true, false)
			if err != nil {
				t.Fatalf("writeColoredJSON() error: %v", err)
			}
			got := stripAnsi(buf.String())
			if got != tt.want {
				t.Errorf("writeColoredJSON() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWriteColoredJSON_color(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		colorize    bool
		wantPlain   string
		wantHasAnsi bool
	}{
		{"null no color", nil, false, "null", false},
		{"null color", nil, true, "null", true},
		{"bool color", true, true, "true", true},
		{"int color", int(42), true, "42", true},
		{"float color", float64(1.5), true, "1.5", true},
		{"string color", "x", true, "\"x\"", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := writeColoredJSON(&buf, tt.input, 0, true, tt.colorize)
			if err != nil {
				t.Fatalf("writeColoredJSON() error: %v", err)
			}
			plain := stripAnsi(buf.String())
			if plain != tt.wantPlain {
				t.Errorf("plain = %q, want %q", plain, tt.wantPlain)
			}
			hasAnsi := buf.String() != plain
			if hasAnsi != tt.wantHasAnsi {
				t.Errorf("hasAnsi = %v, want %v", hasAnsi, tt.wantHasAnsi)
			}
		})
	}
}

func TestWriteColoredJSON_array(t *testing.T) {
	input := []interface{}{int(1), int(2), int(3)}

	t.Run("pretty", func(t *testing.T) {
		var buf bytes.Buffer
		err := writeColoredJSON(&buf, input, 0, true, false)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		want := "[\n  1,\n  2,\n  3\n]"
		got := stripAnsi(buf.String())
		if got != want {
			t.Errorf("got:\n%s\nwant:\n%s", got, want)
		}
	})

	t.Run("compact", func(t *testing.T) {
		var buf bytes.Buffer
		err := writeColoredJSON(&buf, input, 0, false, false)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		want := "[1,2,3]"
		got := stripAnsi(buf.String())
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("empty", func(t *testing.T) {
		var buf bytes.Buffer
		err := writeColoredJSON(&buf, []interface{}{}, 0, true, false)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		got := stripAnsi(buf.String())
		if got != "[]" {
			t.Errorf("got %q, want %q", got, "[]")
		}
	})

	t.Run("nested pretty", func(t *testing.T) {
		input := []interface{}{[]interface{}{int(1), int(2)}}
		var buf bytes.Buffer
		err := writeColoredJSON(&buf, input, 0, true, false)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		want := "[\n  [\n    1,\n    2\n  ]\n]"
		got := stripAnsi(buf.String())
		if got != want {
			t.Errorf("got:\n%s\nwant:\n%s", got, want)
		}
	})
}

func TestWriteColoredJSON_object(t *testing.T) {
	input := map[string]interface{}{"a": int(1), "b": "two"}

	t.Run("compact", func(t *testing.T) {
		var buf bytes.Buffer
		err := writeColoredJSON(&buf, input, 0, false, false)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		got := stripAnsi(buf.String())
		if !strings.HasPrefix(got, "{") || !strings.HasSuffix(got, "}") {
			t.Errorf("expected object wrapper, got %q", got)
		}
		if !strings.Contains(got, `"a":1`) {
			t.Errorf("missing \"a\":1 in %q", got)
		}
		if !strings.Contains(got, `"b":"two"`) {
			t.Errorf("missing \"b\":\"two\" in %q", got)
		}
	})

	t.Run("empty", func(t *testing.T) {
		var buf bytes.Buffer
		err := writeColoredJSON(&buf, map[string]interface{}{}, 0, true, false)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		got := stripAnsi(buf.String())
		if got != "{}" {
			t.Errorf("got %q, want %q", got, "{}")
		}
	})

	t.Run("nested pretty", func(t *testing.T) {
		input := map[string]interface{}{
			"outer": map[string]interface{}{
				"inner": int(1),
			},
		}
		var buf bytes.Buffer
		err := writeColoredJSON(&buf, input, 0, true, false)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		want := "{\n  \"outer\": {\n    \"inner\": 1\n  }\n}"
		got := stripAnsi(buf.String())
		if got != want {
			t.Errorf("got:\n%s\nwant:\n%s", got, want)
		}
	})

	t.Run("pretty key separator has space", func(t *testing.T) {
		input := map[string]interface{}{"k": int(1)}
		var buf bytes.Buffer
		_ = writeColoredJSON(&buf, input, 0, true, false)
		got := stripAnsi(buf.String())
		want := "{\n  \"k\": 1\n}"
		if got != want {
			t.Errorf("got:\n%s\nwant:\n%s", got, want)
		}
	})

	t.Run("compact key separator no space", func(t *testing.T) {
		input := map[string]interface{}{"k": int(1)}
		var buf bytes.Buffer
		_ = writeColoredJSON(&buf, input, 0, false, false)
		got := stripAnsi(buf.String())
		if !strings.Contains(got, `"k":1`) {
			t.Errorf("expected compact separator in %q", got)
		}
	})
}

func TestWriteResult_noPath(t *testing.T) {
	var buf bytes.Buffer
	err := WriteResult(&buf, int(42), nil, true, false, false)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	got := stripAnsi(buf.String())
	want := "42\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWriteResult_withPath(t *testing.T) {
	path := []query.PathType{
		{Kind: query.PathField, Field: "users"},
		{Kind: query.PathIndex, Index: 0},
		{Kind: query.PathField, Field: "name"},
	}
	var buf bytes.Buffer
	err := WriteResult(&buf, "Alice", path, true, true, false)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	got := stripAnsi(buf.String())
	want := "users.[0].name:\n\"Alice\"\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWriteResult_withPath_color(t *testing.T) {
	path := []query.PathType{
		{Kind: query.PathField, Field: "x"},
	}
	var buf bytes.Buffer
	err := WriteResult(&buf, int(1), path, true, true, true)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	raw := buf.String()
	if !strings.Contains(raw, "\x1b[") {
		t.Error("expected ANSI escape codes in output")
	}
	plain := stripAnsi(raw)
	if !strings.HasPrefix(plain, "x:\n") {
		t.Errorf("expected path header in %q", plain)
	}
}

func TestWriteResult_compact(t *testing.T) {
	input := map[string]interface{}{"a": int(1)}
	var buf bytes.Buffer
	err := WriteResult(&buf, input, nil, false, false, false)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	got := stripAnsi(buf.String())
	if strings.Contains(got, "\n  ") {
		t.Errorf("compact mode should not indent: %q", got)
	}
}

func TestWriteResult_colorSuppression(t *testing.T) {
	path := []query.PathType{{Kind: query.PathField, Field: "k"}}
	var buf bytes.Buffer
	err := WriteResult(&buf, "v", path, true, true, false)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if strings.Contains(buf.String(), "\x1b[") {
		t.Error("expected no ANSI codes when colorize=false")
	}
}

func TestWriteResult_emptyPathShowPath(t *testing.T) {
	var buf bytes.Buffer
	err := WriteResult(&buf, int(1), nil, true, true, false)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	got := stripAnsi(buf.String())
	if strings.Contains(got, ":") && strings.HasPrefix(got, "1") == false {
		t.Errorf("empty path with showPath=true should not print header, got %q", got)
	}
	want := "1\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWriteResult_brokenPipe(t *testing.T) {
	path := []query.PathType{{Kind: query.PathField, Field: "x"}}
	w := &brokenPipeWriter{}
	err := WriteResult(w, "value", path, true, true, false)
	if err != nil {
		t.Errorf("expected nil on broken pipe, got: %v", err)
	}
}

type brokenPipeWriter struct{}

func (w *brokenPipeWriter) Write(p []byte) (int, error) {
	return 0, fmt.Errorf("write |1: broken pipe")
}

func TestWriteColoredJSON_mixedNested(t *testing.T) {
	input := map[string]interface{}{
		"arr": []interface{}{
			int(1),
			map[string]interface{}{"nested": true},
			nil,
		},
		"str": "hello",
	}
	var buf bytes.Buffer
	err := writeColoredJSON(&buf, input, 0, true, false)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	got := stripAnsi(buf.String())

	if !strings.Contains(got, `"arr"`) {
		t.Error("missing arr key")
	}
	if !strings.Contains(got, `"nested": true`) {
		t.Error("missing nested:true value")
	}
	if !strings.Contains(got, "null") {
		t.Error("missing null")
	}
	if !strings.Contains(got, `"str"`) {
		t.Error("missing str key")
	}
	if !strings.Contains(got, `"hello"`) {
		t.Error("missing hello string")
	}
}
