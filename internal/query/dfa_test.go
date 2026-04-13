package query

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/fantods/gjq/internal/jsonutil"
)

func mustParseAnyJSON(t *testing.T, input string) interface{} {
	t.Helper()
	var result interface{}
	dec := json.NewDecoder(strings.NewReader(input))

	if err := dec.Decode(&result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	return result
}

var simpleJSON = `{
  "foo": {
    "bar": "val"
  },
  "baz": [1, 2, 3, 4, 5],
  "other": 42
}`

var nestedJSON = `{
  "nested": {
    "a": {
      "b": {
        "c": "target"
      }
    }
  }
}`

var duplicateKeyNestedJSON = `{
  "c": {
    "c": {
      "c": "target"
    }
  }
}`

func mustParseJSON(t *testing.T, input string) map[string]interface{} {
	result := mustParseAnyJSON(t, input)
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected JSON object, got %T", result)
	}
	return m
}

func mustParseQuery(t *testing.T, input string) Query {
	q, err := ParseQuery(input)
	if err != nil {
		t.Fatalf("failed to parse query %q: %v", input, err)
	}
	return q
}

func TestDFAEmptyQuery(t *testing.T) {
	json := mustParseJSON(t, simpleJSON)
	q := NewSequence(nil)
	dfa := NewDFA(q, false)
	results := dfa.Find(json)
	if len(results) != 1 {
		t.Fatalf("expected 1 match (identity), got %d", len(results))
	}
}

func TestDFASingleField(t *testing.T) {
	json := mustParseJSON(t, simpleJSON)
	q := mustParseQuery(t, "foo")
	dfa := NewDFA(q, false)
	results := dfa.Find(json)
	if len(results) != 1 {
		t.Fatalf("expected 1 match, got %d", len(results))
	}
	if results[0].Value.(map[string]interface{})["bar"] != "val" {
		t.Fatalf("expected value 'val', got %v", results[0].Value)
	}
}

func TestDFAFieldSequence(t *testing.T) {
	json := mustParseJSON(t, simpleJSON)
	q := mustParseQuery(t, "foo.bar")
	dfa := NewDFA(q, false)
	results := dfa.Find(json)
	if len(results) != 1 {
		t.Fatalf("expected 1 match, got %d", len(results))
	}
	paths := results[0].Path
	if len(paths) != 2 || paths[0].Field != "foo" || paths[1].Field != "bar" {
		t.Fatalf("expected path [foo, bar], got %v", paths)
	}
}

func TestDFADisjunction(t *testing.T) {
	json := mustParseJSON(t, simpleJSON)
	q := mustParseQuery(t, "foo | baz")
	dfa := NewDFA(q, false)
	results := dfa.Find(json)
	if len(results) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(results))
	}
}

func TestDFAIndexAccess(t *testing.T) {
	json := mustParseJSON(t, simpleJSON)
	q := mustParseQuery(t, "baz[1]")
	dfa := NewDFA(q, false)
	results := dfa.Find(json)
	if len(results) != 1 {
		t.Fatalf("expected 1 match, got %d", len(results))
	}
	arr := json["baz"].([]interface{})
	if results[0].Value != arr[1] {
		t.Fatalf("expected value 2, got %v", results[0].Value)
	}
}

func TestDFAArrayWildcard(t *testing.T) {
	json := mustParseJSON(t, simpleJSON)
	q := mustParseQuery(t, "baz[*]")
	dfa := NewDFA(q, false)
	results := dfa.Find(json)
	if len(results) != 5 {
		t.Fatalf("expected 5 matches, got %d", len(results))
	}
}

func TestDFABoundedRange(t *testing.T) {
	json := mustParseJSON(t, simpleJSON)
	q := mustParseQuery(t, "baz[1:4]")
	dfa := NewDFA(q, false)
	results := dfa.Find(json)
	if len(results) != 3 {
		t.Fatalf("expected 3 matches, got %d", len(results))
	}
}

func TestDFAUnboundedRangeFrom(t *testing.T) {
	json := mustParseJSON(t, simpleJSON)
	q := mustParseQuery(t, "baz[2:]")
	dfa := NewDFA(q, false)
	results := dfa.Find(json)
	if len(results) != 3 {
		t.Fatalf("expected 3 matches, got %d", len(results))
	}
}

func TestDFAUnboundedRangeTo(t *testing.T) {
	json := mustParseJSON(t, simpleJSON)
	q := mustParseQuery(t, "baz[:2]")
	dfa := NewDFA(q, false)
	results := dfa.Find(json)
	if len(results) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(results))
	}
}

func TestDFARangeAll(t *testing.T) {
	json := mustParseJSON(t, simpleJSON)
	q := mustParseQuery(t, "baz[:]")
	dfa := NewDFA(q, false)
	results := dfa.Find(json)
	if len(results) != 5 {
		t.Fatalf("expected 5 matches, got %d", len(results))
	}
}

func TestDFAEmptyRange(t *testing.T) {
	json := mustParseJSON(t, simpleJSON)
	q := mustParseQuery(t, "baz[1:1]")
	dfa := NewDFA(q, false)
	results := dfa.Find(json)
	if len(results) != 0 {
		t.Fatalf("expected 0 matches, got %d", len(results))
	}
}

func TestDFAOptionalField(t *testing.T) {
	json := mustParseJSON(t, simpleJSON)
	q := mustParseQuery(t, "other?")
	dfa := NewDFA(q, false)
	results := dfa.Find(json)
	if len(results) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(results))
	}
}

func TestDFAKleeneStarSameKey(t *testing.T) {
	json := mustParseJSON(t, duplicateKeyNestedJSON)
	q := mustParseQuery(t, "c*")
	dfa := NewDFA(q, false)
	results := dfa.Find(json)
	if len(results) != 4 {
		t.Fatalf("expected 4 matches, got %d", len(results))
	}
}

func TestDFAFieldWildcardNotRecursive(t *testing.T) {
	json := mustParseJSON(t, nestedJSON)
	q := mustParseQuery(t, "*.c")
	dfa := NewDFA(q, false)
	results := dfa.Find(json)
	if len(results) != 0 {
		t.Fatalf("expected 0 matches (*.c is not recursive), got %d", len(results))
	}
}

func TestDFAFieldWildcardNested(t *testing.T) {
	json := mustParseJSON(t, nestedJSON)
	q := mustParseQuery(t, "nested.*.*.c")
	dfa := NewDFA(q, false)
	results := dfa.Find(json)
	if len(results) != 1 {
		t.Fatalf("expected 1 match, got %d", len(results))
	}
}

func TestDFAFieldWildcardDeep(t *testing.T) {
	json := mustParseJSON(t, nestedJSON)
	q := mustParseQuery(t, "*.*.*.c")
	dfa := NewDFA(q, false)
	results := dfa.Find(json)
	if len(results) != 1 {
		t.Fatalf("expected 1 match, got %d", len(results))
	}
}

func TestDFAFieldWildcardNonuniqueKeys(t *testing.T) {
	json := mustParseJSON(t, duplicateKeyNestedJSON)
	q := mustParseQuery(t, "c.*.c")
	dfa := NewDFA(q, false)
	results := dfa.Find(json)
	if len(results) != 1 {
		t.Fatalf("expected 1 match, got %d", len(results))
	}
}

func TestDFAMultipleOptional(t *testing.T) {
	json := mustParseJSON(t, duplicateKeyNestedJSON)
	q := mustParseQuery(t, "c*.c?.c?")
	dfa := NewDFA(q, false)
	results := dfa.Find(json)
	if len(results) != 4 {
		t.Fatalf("expected 4 matches, got %d", len(results))
	}
}

func TestDFAKleeneStarRecursiveType(t *testing.T) {
	input := `{
	  "type": {
	    "type": "value1",
	    "b": {
	      "type": "value2"
	    }
	  }
	}`
	json := mustParseJSON(t, input)
	q := mustParseQuery(t, "**.type")
	dfa := NewDFA(q, false)
	results := dfa.Find(json)
	if len(results) != 3 {
		t.Fatalf("expected 3 matches, got %d", len(results))
	}
}

func TestDFADisjunctionGroup(t *testing.T) {
	input := `{"x": {"y": 5, "z": { "t": 2}}}`
	json := mustParseJSON(t, input)
	q := mustParseQuery(t, "x.(y | z.t)")
	dfa := NewDFA(q, false)
	results := dfa.Find(json)
	if len(results) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(results))
	}
}

func TestDFARecursiveArrayIndexing(t *testing.T) {
	input := `[[1], [2, 3]]`
	js := mustParseAnyJSON(t, input)
	q := mustParseQuery(t, "[*]*")
	dfa := NewDFA(q, false)
	results := dfa.Find(js)
	if len(results) != 6 {
		t.Fatalf("expected 6 matches, got %d", len(results))
	}
}

func TestDFARecursiveArrayIndexingAnyLevel(t *testing.T) {
	input := `[[1], [2, 3]]`
	json := mustParseAnyJSON(t, input)
	q := mustParseQuery(t, "**.[*]*.[*]")
	dfa := NewDFA(q, false)
	results := dfa.Find(json)
	if len(results) != 5 {
		t.Fatalf("expected 5 matches, got %d", len(results))
	}
}

func TestDFARecursiveGeojsonAnyFieldsThenArrays(t *testing.T) {
	input := `{
	   "type":"FeatureCollection",
	   "features":[
	      {
	         "geometry":{
	            "coordinates":[
	               [
	                  [
	                     1,
	                     2
	                  ]
	               ]
	            ]
	         }
	      }
	   ]
	}`
	json := mustParseJSON(t, input)
	q := mustParseQuery(t, "**.[*]*.[*]")
	dfa := NewDFA(q, false)
	results := dfa.Find(json)
	if len(results) != 1 {
		t.Fatalf("expected 1 match, got %d", len(results))
	}
}

func TestDFARecursiveGeojsonGroupAnyLevel(t *testing.T) {
	input := `{
	   "type":"FeatureCollection",
	   "features":[
	      {
	         "geometry":{
	            "coordinates":[
	               [
	                  [
	                     1,
	                     2
	                  ]
	               ]
	            ]
	         }
	      }
	   ]
	}`
	json := mustParseJSON(t, input)
	q := mustParseQuery(t, "(* | [*])*.[*]")
	dfa := NewDFA(q, false)
	results := dfa.Find(json)
	if len(results) != 5 {
		t.Fatalf("expected 5 matches, got %d", len(results))
	}
}

func TestDFAOverlappingRanges(t *testing.T) {
	json := mustParseJSON(t, simpleJSON)
	q := mustParseQuery(t, "baz[0:3] | baz[1:]")
	dfa := NewDFA(q, false)
	results := dfa.Find(json)
	if len(results) != 5 {
		t.Fatalf("expected 5 matches (union of [0:3] and [1:]), got %d", len(results))
	}
}

func TestDFAConstruction(t *testing.T) {
	q := mustParseQuery(t, "foo.bar")
	dfa := NewDFA(q, false)

	if dfa.NumStatesForTest() != 3 {
		t.Fatalf("expected 3 states, got %d", dfa.NumStatesForTest())
	}
	if !dfa.IsAcceptingForTest()[2] {
		t.Fatal("state 2 should be accepting")
	}
	if dfa.IsAcceptingForTest()[0] || dfa.IsAcceptingForTest()[1] {
		t.Fatal("states 0 and 1 should not be accepting")
	}
	if _, ok := dfa.KeyToIDForTest()["foo"]; !ok {
		t.Fatal("expected 'foo' in key_to_key_id")
	}
	if _, ok := dfa.KeyToIDForTest()["bar"]; !ok {
		t.Fatal("expected 'bar' in key_to_key_id")
	}
}

func TestDFACaseInsensitive(t *testing.T) {
	input := `{ "FOO": 1, "bar": 2 }`
	json := mustParseJSON(t, input)
	q := mustParseQuery(t, "foo")
	dfa := NewDFA(q, true)
	results := dfa.Find(json)
	if len(results) != 1 {
		t.Fatalf("expected 1 match, got %d", len(results))
	}
}

func TestDFACaseInsensitiveSequence(t *testing.T) {
	input := `{ "Foo": { "BAR": "found" } }`
	json := mustParseJSON(t, input)
	q := mustParseQuery(t, "foo.bar")
	dfa := NewDFA(q, true)
	results := dfa.Find(json)
	if len(results) != 1 {
		t.Fatalf("expected 1 match, got %d", len(results))
	}
	if s, ok := results[0].Value.(string); !ok || s != "found" {
		t.Fatalf("expected 'found', got %v", results[0].Value)
	}
}

func TestDFACaseInsensitiveDisjunctionDedup(t *testing.T) {
	input := `{ "foo": 1, "bar": 2 }`
	json := mustParseJSON(t, input)
	q := mustParseQuery(t, "Foo | foo")
	dfa := NewDFA(q, true)
	results := dfa.Find(json)
	if len(results) != 1 {
		t.Fatalf("expected 1 match (deduped), got %d", len(results))
	}
}

func TestDFACaseSensitiveDefault(t *testing.T) {
	input := `{ "FOO": 1, "foo": 2 }`
	json := mustParseJSON(t, input)
	q := mustParseQuery(t, "foo")
	dfa := NewDFA(q, false)
	results := dfa.Find(json)
	if len(results) != 1 {
		t.Fatalf("expected 1 match, got %d", len(results))
	}
	if n, ok := results[0].Value.(float64); !ok || n != 2 {
		t.Fatalf("expected 2, got %v", results[0].Value)
	}
}

func TestDFAQuotedFieldWithSlash(t *testing.T) {
	input := `{ "/activities": { "get": "list" } }`
	json := mustParseJSON(t, input)
	q := mustParseQuery(t, `"/activities"`)
	dfa := NewDFA(q, false)
	results := dfa.Find(json)
	if len(results) != 1 {
		t.Fatalf("expected 1 match, got %d", len(results))
	}
}

func TestDFAQuotedFieldSequence(t *testing.T) {
	input := `{
	  "paths": {
	    "/activities": { "get": "list" },
	    "/users": { "get": "list_users" }
	  }
	}`
	json := mustParseJSON(t, input)
	q := mustParseQuery(t, `paths."/activities"`)
	dfa := NewDFA(q, false)
	results := dfa.Find(json)
	if len(results) != 1 {
		t.Fatalf("expected 1 match, got %d", len(results))
	}
}

func TestDFAQuotedFieldWithDot(t *testing.T) {
	input := `{ "a.b": 42, "a": { "b": 99 } }`
	json := mustParseJSON(t, input)
	q := mustParseQuery(t, `"a.b"`)
	dfa := NewDFA(q, false)
	results := dfa.Find(json)
	if len(results) != 1 {
		t.Fatalf("expected 1 match, got %d", len(results))
	}
	if n, ok := results[0].Value.(float64); !ok || n != 42 {
		t.Fatalf("expected 42 (from literal key 'a.b'), got %v", results[0].Value)
	}
}

func TestDFAQuotedFieldDisjunction(t *testing.T) {
	input := `{
	  "paths": {
	    "/activities": { "get": "list" },
	    "/users": { "get": "list_users" }
	  }
	}`
	json := mustParseJSON(t, input)
	q := mustParseQuery(t, `paths.("/activities" | "/users")`)
	dfa := NewDFA(q, false)
	results := dfa.Find(json)
	if len(results) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(results))
	}
}

func TestDFAFieldSymbolIDLookup(t *testing.T) {
	q := mustParseQuery(t, "foo.bar")
	dfa := NewDFA(q, false)
	fooID := dfa.FieldSymbolID("foo")
	barID := dfa.FieldSymbolID("bar")
	otherID := dfa.FieldSymbolID("unknown")
	if otherID != 0 {
		t.Fatalf("expected 'unknown' to map to Other (0), got %d", otherID)
	}
	if fooID == otherID || barID == otherID {
		t.Fatal("foo and bar should not map to Other")
	}
}

func TestDFAIndexSymbolIDLookup(t *testing.T) {
	q := mustParseQuery(t, "baz[2:5]")
	dfa := NewDFA(q, false)

	for i := 2; i < 5; i++ {
		id, ok := dfa.IndexSymbolID(i)
		if !ok || id == 0 {
			t.Fatalf("expected index %d to resolve to a range symbol, got id=%d ok=%v", i, id, ok)
		}
	}
	id, ok := dfa.IndexSymbolID(0)
	if ok && id != 0 {
		t.Fatalf("index 0 should not be in range [2,5), got id=%d ok=%v", id, ok)
	}
}

func TestDFATransition(t *testing.T) {
	q := mustParseQuery(t, "foo.bar")
	dfa := NewDFA(q, false)
	fooID := dfa.FieldSymbolID("foo")
	barID := dfa.FieldSymbolID("bar")
	otherID := dfa.FieldSymbolID("baz")

	next, ok := dfa.Transition(0, fooID)
	if !ok {
		t.Fatal("expected transition on foo from state 0")
	}
	next2, ok := dfa.Transition(next, barID)
	if !ok || !dfa.IsAcceptingState(next2) {
		t.Fatal("expected transition on bar leading to accepting state")
	}
	_, ok = dfa.Transition(0, otherID)
	if ok {
		t.Fatal("expected no transition on 'baz' from state 0")
	}
}

func TestDFANoRangeOverlaps(t *testing.T) {
	q := mustParseQuery(t, "foo[1:5].baz[2]")
	dfa := NewDFA(q, false)
	prevEnd := 0
	for _, re := range dfa.RangesForTest() {
		if re.Start < prevEnd {
			t.Fatalf("overlapping range detected: [%d,%d) overlaps with previous end %d", re.Start, re.End, prevEnd)
		}
		prevEnd = re.End
	}
}

func TestDFARangeFromLookup(t *testing.T) {
	q := mustParseQuery(t, "baz[3:]")
	dfa := NewDFA(q, false)
	id, ok := dfa.IndexSymbolID(5)
	if !ok {
		t.Fatal("expected index 5 to be found in range [3, MaxInt)")
	}
	_ = id
	id2, ok := dfa.IndexSymbolID(2)
	if ok && id2 != 0 {
		t.Fatalf("index 2 should not be in range [3, MaxInt), got id=%d", id2)
	}
}

func TestDFAGetAllArrayAfterField(t *testing.T) {
	input := `{ "root": [["1", "2"], ["3"]] }`
	json := mustParseJSON(t, input)
	q := mustParseQuery(t, "**.[*]")
	dfa := NewDFA(q, false)
	results := dfa.Find(json)
	if len(results) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(results))
	}
}

func TestDFATwoFieldWildcards(t *testing.T) {
	input := `{ "root": { "foo": "bar" } }`
	json := mustParseJSON(t, input)
	q := mustParseQuery(t, "*.*")
	dfa := NewDFA(q, false)
	results := dfa.Find(json)
	if len(results) != 1 {
		t.Fatalf("expected 1 match, got %d", len(results))
	}
}

func TestDFAArrayObjNoFields(t *testing.T) {
	input := `[{"root": { "foo": "bar" }}]`
	json := mustParseAnyJSON(t, input)
	q := mustParseQuery(t, "*.*")
	dfa := NewDFA(q, false)
	results := dfa.Find(json)
	if len(results) != 0 {
		t.Fatalf("expected 0 matches (array root, no field match), got %d", len(results))
	}
}

func TestDFACaseInsensitiveRecursiveWildcard(t *testing.T) {
	input := `{
	  "a": {
	    "FOO": "deep"
	  },
	  "FOO": "shallow"
	}`
	json := mustParseJSON(t, input)
	q := mustParseQuery(t, "**.foo")
	dfa := NewDFA(q, true)
	results := dfa.Find(json)
	if len(results) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(results))
	}
}

func TestDFAIndexSymbolIDBinarySearchCorrectness(t *testing.T) {
	q := mustParseQuery(t, "a[0:10] | a[5:20] | a[15:100]")
	dfa := NewDFA(q, false)

	type expectation struct {
		index       int
		wantOK      bool
		wantNonZero bool
	}
	cases := []expectation{
		{0, true, true},
		{1, true, true},
		{9, true, true},
		{10, true, true},
		{19, true, true},
		{50, true, true},
		{99, true, true},
		{100, false, false},
		{200, false, false},
	}
	for _, tc := range cases {
		id, ok := dfa.IndexSymbolID(tc.index)
		if ok != tc.wantOK {
			t.Errorf("index %d: got ok=%v, want %v", tc.index, ok, tc.wantOK)
		}
		if tc.wantNonZero && ok && id == 0 {
			t.Errorf("index %d: expected non-zero symbol id", tc.index)
		}
	}

	for i := 0; i < 100; i++ {
		id, ok := dfa.IndexSymbolID(i)
		if !ok {
			t.Errorf("index %d should be in some range", i)
		}
		if id == 0 {
			t.Errorf("index %d should have non-zero symbol id", i)
		}
	}
}

func TestDFAIndexSymbolIDBinarySearchOverlappingRanges(t *testing.T) {
	q := mustParseQuery(t, "x[0:3] | x[2:5] | x[4:8]")
	dfa := NewDFA(q, false)

	for i := 0; i < 8; i++ {
		id, ok := dfa.IndexSymbolID(i)
		if !ok {
			t.Fatalf("index %d: expected to be found", i)
		}
		if id == 0 {
			t.Fatalf("index %d: expected non-zero symbol id", i)
		}
	}
	id, ok := dfa.IndexSymbolID(8)
	if ok && id != 0 {
		t.Fatalf("index 8 should not be in any range")
	}
	id, ok = dfa.IndexSymbolID(-1)
	if ok && id != 0 {
		t.Fatalf("index -1 should not be in any range")
	}
}

func TestDFAIndexSymbolIDBinarySearchSingleRange(t *testing.T) {
	q := mustParseQuery(t, "a[3:7]")
	dfa := NewDFA(q, false)

	for i := 3; i < 7; i++ {
		id, ok := dfa.IndexSymbolID(i)
		if !ok || id == 0 {
			t.Fatalf("index %d: expected in range [3,7)", i)
		}
	}
	for _, idx := range []int{0, 1, 2, 7, 8, 100} {
		id, ok := dfa.IndexSymbolID(idx)
		if ok && id != 0 {
			t.Fatalf("index %d: should not be in range [3,7)", idx)
		}
	}
}

func TestDFAIndexSymbolIDBinarySearchArrayWildcard(t *testing.T) {
	q := mustParseQuery(t, "a[*]")
	dfa := NewDFA(q, false)

	for _, idx := range []int{0, 1, 50, 9999} {
		id, ok := dfa.IndexSymbolID(idx)
		if !ok || id == 0 {
			t.Fatalf("index %d: expected in wildcard range", idx)
		}
	}
}

func BenchmarkIndexSymbolID(b *testing.B) {
	q := mustParseQueryB(b, "a[0:10] | a[5:50] | a[40:200] | a[150:500] | a[400:1000]")
	dfa := NewDFA(q, false)

	indices := make([]int, 1000)
	for i := range indices {
		indices[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, idx := range indices {
			dfa.IndexSymbolID(idx)
		}
	}
}

func mustParseQueryB(b *testing.B, input string) Query {
	q, err := ParseQuery(input)
	if err != nil {
		b.Fatalf("failed to parse query %q: %v", input, err)
	}
	return q
}

func TestParseJSONFromBytesEquivalence(t *testing.T) {
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
		resultStr, err := jsonutil.Parse(input)
		if err != nil {
			t.Fatalf("ParseJSON(%q) error: %v", input, err)
		}
		resultBytes, err := jsonutil.ParseBytes([]byte(input))
		if err != nil {
			t.Fatalf("ParseJSONFromBytes(%q) error: %v", input, err)
		}
		if fmt.Sprintf("%v", resultStr) != fmt.Sprintf("%v", resultBytes) {
			t.Errorf("ParseJSON and ParseJSONFromBytes returned different results for %q:\n  string: %v\n  bytes:  %v", input, resultStr, resultBytes)
		}
	}
}

func TestParseJSONFromBytesNumberRepresentation(t *testing.T) {
	input := `{"int_val": 42, "float_val": 3.14, "big_int": 999999999999}`
	result, err := jsonutil.ParseBytes([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if v, ok := m["int_val"].(float64); !ok || v != 42 {
		t.Fatalf("expected float64 42, got %v", m["int_val"])
	}
	if v, ok := m["float_val"].(float64); !ok || v != 3.14 {
		t.Fatalf("expected float64 3.14, got %v", m["float_val"])
	}
}

func BenchmarkParseJSONString(b *testing.B) {
	var buf bytes.Buffer
	buf.WriteString(`{"items": [`)
	for i := 0; i < 1000; i++ {
		if i > 0 {
			buf.WriteString(", ")
		}
		fmt.Fprintf(&buf, `{"id": %d, "name": "item_%d", "value": %.1f}`, i, i, float64(i)*1.1)
	}
	buf.WriteString(`]}`)
	data := buf.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := jsonutil.Parse(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseJSONBytes(b *testing.B) {
	var buf bytes.Buffer
	buf.WriteString(`{"items": [`)
	for i := 0; i < 1000; i++ {
		if i > 0 {
			buf.WriteString(", ")
		}
		fmt.Fprintf(&buf, `{"id": %d, "name": "item_%d", "value": %.1f}`, i, i, float64(i)*1.1)
	}
	buf.WriteString(`]}`)
	data := buf.Bytes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := jsonutil.ParseBytes(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

var deepJSON = `{
  "l1": {
    "l2": {
      "l3": {
        "l4": {
          "l5": {
            "l6": {
              "l7": {
                "l8": {
                  "l9": {
                    "l10": "deep_value"
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}`

func TestDFAFindDeepPathCorrectness(t *testing.T) {
	root := mustParseJSON(t, deepJSON)
	q := mustParseQuery(t, "l1.l2.l3.l4.l5.l6.l7.l8.l9.l10")
	dfa := NewDFA(q, false)
	results := dfa.Find(root)
	if len(results) != 1 {
		t.Fatalf("expected 1 match, got %d", len(results))
	}
	if len(results[0].Path) != 10 {
		t.Fatalf("expected path length 10, got %d", len(results[0].Path))
	}
	if results[0].Value != "deep_value" {
		t.Fatalf("expected 'deep_value', got %v", results[0].Value)
	}
	expected := []string{"l1", "l2", "l3", "l4", "l5", "l6", "l7", "l8", "l9", "l10"}
	for i, pt := range results[0].Path {
		if pt.Field != expected[i] {
			t.Errorf("path[%d]: expected %q, got %q", i, expected[i], pt.Field)
		}
	}
}

func TestDFAFindMultipleDeepPaths(t *testing.T) {
	input := `{
	  "a": { "b": { "c": { "target": 1 } } },
	  "x": { "y": { "z": { "target": 2 } } }
	}`
	root := mustParseJSON(t, input)
	q := mustParseQuery(t, "*.*.*.target")
	dfa := NewDFA(q, false)
	results := dfa.Find(root)
	if len(results) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(results))
	}
	vals := map[float64]bool{}
	for _, r := range results {
		if v, ok := r.Value.(float64); ok {
			vals[v] = true
		}
	}
	if !vals[1] || !vals[2] {
		t.Fatalf("expected values 1 and 2, got %v", vals)
	}
	for _, r := range results {
		if len(r.Path) != 4 {
			t.Errorf("expected path length 4, got %d: %v", len(r.Path), r.Path)
		}
	}
}

func BenchmarkDFAFind(b *testing.B) {
	root := mustParseAnyJSONB(b, deepJSON)
	q := mustParseQueryB(b, "l1.l2.l3.l4.l5.l6.l7.l8.l9.l10")
	dfa := NewDFA(q, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dfa.Find(root)
	}
}

func BenchmarkDFAFindWildcard(b *testing.B) {
	input := `{"items": [`
	for i := 0; i < 100; i++ {
		if i > 0 {
			input += ", "
		}
		input += fmt.Sprintf(`{"name": "item_%d", "tags": ["a", "b", "c"]}`, i)
	}
	input += `]}`
	root := mustParseAnyJSONB(b, input)
	q := mustParseQueryB(b, "items[*].tags[*]")
	dfa := NewDFA(q, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dfa.Find(root)
	}
}

func mustParseAnyJSONB(b *testing.B, input string) interface{} {
	b.Helper()
	var result interface{}
	dec := json.NewDecoder(strings.NewReader(input))

	if err := dec.Decode(&result); err != nil {
		b.Fatalf("failed to parse JSON: %v", err)
	}
	return result
}

func loadRandomUserJSON(b *testing.B) interface{} {
	b.Helper()
	data, err := os.ReadFile("../../tests/data/randomusers.json")
	if err != nil {
		b.Fatalf("failed to load randomusers.json: %v", err)
	}
	root, err := jsonutil.ParseBytes(data)
	if err != nil {
		b.Fatalf("failed to parse randomusers.json: %v", err)
	}
	return root
}

func BenchmarkRandomUserParse(b *testing.B) {
	data, err := os.ReadFile("../../tests/data/randomusers.json")
	if err != nil {
		b.Fatalf("failed to load randomusers.json: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := jsonutil.ParseBytes(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRandomUserFindNat(b *testing.B) {
	root := loadRandomUserJSON(b)
	q := mustParseQueryB(b, "results[*].nat")
	dfa := NewDFA(q, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dfa.Find(root)
	}
}

func BenchmarkRandomUserFindLastName(b *testing.B) {
	root := loadRandomUserJSON(b)
	q := mustParseQueryB(b, "results[*].name[*].last")
	dfa := NewDFA(q, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dfa.Find(root)
	}
}

func BenchmarkRandomUserFindRecursive(b *testing.B) {
	root := loadRandomUserJSON(b)
	q := mustParseQueryB(b, "**.first")
	dfa := NewDFA(q, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dfa.Find(root)
	}
}

func BenchmarkRandomUserFindCaseInsensitive(b *testing.B) {
	root := loadRandomUserJSON(b)
	q := mustParseQueryB(b, "**.first")
	dfa := NewDFA(q, true)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dfa.Find(root)
	}
}

func BenchmarkRandomUserFindFixedString(b *testing.B) {
	root := loadRandomUserJSON(b)
	q := NewSequence([]Query{
		NewKleeneStar(NewDisjunction([]Query{
			NewFieldWildcard(),
			NewArrayWildcard(),
		})),
		NewField("email"),
	})
	dfa := NewDFA(q, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dfa.Find(root)
	}
}
