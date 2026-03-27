package query

import (
	"encoding/json"
	"strings"
	"testing"
)

func mustParseAnyJSON(t *testing.T, input string) interface{} {
	t.Helper()
	var result interface{}
	dec := json.NewDecoder(strings.NewReader(input))
	dec.UseNumber()
	if err := dec.Decode(&result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	return convertNumbers(result)
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
	dfa := NewQueryDFA(&q, false)
	results := dfa.Find(json)
	if len(results) != 1 {
		t.Fatalf("expected 1 match (identity), got %d", len(results))
	}
}

func TestDFASingleField(t *testing.T) {
	json := mustParseJSON(t, simpleJSON)
	q := mustParseQuery(t, "foo")
	dfa := NewQueryDFA(&q, false)
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
	dfa := NewQueryDFA(&q, false)
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
	dfa := NewQueryDFA(&q, false)
	results := dfa.Find(json)
	if len(results) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(results))
	}
}

func TestDFAIndexAccess(t *testing.T) {
	json := mustParseJSON(t, simpleJSON)
	q := mustParseQuery(t, "baz[1]")
	dfa := NewQueryDFA(&q, false)
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
	dfa := NewQueryDFA(&q, false)
	results := dfa.Find(json)
	if len(results) != 5 {
		t.Fatalf("expected 5 matches, got %d", len(results))
	}
}

func TestDFABoundedRange(t *testing.T) {
	json := mustParseJSON(t, simpleJSON)
	q := mustParseQuery(t, "baz[1:4]")
	dfa := NewQueryDFA(&q, false)
	results := dfa.Find(json)
	if len(results) != 3 {
		t.Fatalf("expected 3 matches, got %d", len(results))
	}
}

func TestDFAUnboundedRangeFrom(t *testing.T) {
	json := mustParseJSON(t, simpleJSON)
	q := mustParseQuery(t, "baz[2:]")
	dfa := NewQueryDFA(&q, false)
	results := dfa.Find(json)
	if len(results) != 3 {
		t.Fatalf("expected 3 matches, got %d", len(results))
	}
}

func TestDFAUnboundedRangeTo(t *testing.T) {
	json := mustParseJSON(t, simpleJSON)
	q := mustParseQuery(t, "baz[:2]")
	dfa := NewQueryDFA(&q, false)
	results := dfa.Find(json)
	if len(results) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(results))
	}
}

func TestDFARangeAll(t *testing.T) {
	json := mustParseJSON(t, simpleJSON)
	q := mustParseQuery(t, "baz[:]")
	dfa := NewQueryDFA(&q, false)
	results := dfa.Find(json)
	if len(results) != 5 {
		t.Fatalf("expected 5 matches, got %d", len(results))
	}
}

func TestDFAEmptyRange(t *testing.T) {
	json := mustParseJSON(t, simpleJSON)
	q := mustParseQuery(t, "baz[1:1]")
	dfa := NewQueryDFA(&q, false)
	results := dfa.Find(json)
	if len(results) != 0 {
		t.Fatalf("expected 0 matches, got %d", len(results))
	}
}

func TestDFAOptionalField(t *testing.T) {
	json := mustParseJSON(t, simpleJSON)
	q := mustParseQuery(t, "other?")
	dfa := NewQueryDFA(&q, false)
	results := dfa.Find(json)
	if len(results) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(results))
	}
}

func TestDFAKleeneStarSameKey(t *testing.T) {
	json := mustParseJSON(t, duplicateKeyNestedJSON)
	q := mustParseQuery(t, "c*")
	dfa := NewQueryDFA(&q, false)
	results := dfa.Find(json)
	if len(results) != 4 {
		t.Fatalf("expected 4 matches, got %d", len(results))
	}
}

func TestDFAFieldWildcardNotRecursive(t *testing.T) {
	json := mustParseJSON(t, nestedJSON)
	q := mustParseQuery(t, "*.c")
	dfa := NewQueryDFA(&q, false)
	results := dfa.Find(json)
	if len(results) != 0 {
		t.Fatalf("expected 0 matches (*.c is not recursive), got %d", len(results))
	}
}

func TestDFAFieldWildcardNested(t *testing.T) {
	json := mustParseJSON(t, nestedJSON)
	q := mustParseQuery(t, "nested.*.*.c")
	dfa := NewQueryDFA(&q, false)
	results := dfa.Find(json)
	if len(results) != 1 {
		t.Fatalf("expected 1 match, got %d", len(results))
	}
}

func TestDFAFieldWildcardDeep(t *testing.T) {
	json := mustParseJSON(t, nestedJSON)
	q := mustParseQuery(t, "*.*.*.c")
	dfa := NewQueryDFA(&q, false)
	results := dfa.Find(json)
	if len(results) != 1 {
		t.Fatalf("expected 1 match, got %d", len(results))
	}
}

func TestDFAFieldWildcardNonuniqueKeys(t *testing.T) {
	json := mustParseJSON(t, duplicateKeyNestedJSON)
	q := mustParseQuery(t, "c.*.c")
	dfa := NewQueryDFA(&q, false)
	results := dfa.Find(json)
	if len(results) != 1 {
		t.Fatalf("expected 1 match, got %d", len(results))
	}
}

func TestDFAMultipleOptional(t *testing.T) {
	json := mustParseJSON(t, duplicateKeyNestedJSON)
	q := mustParseQuery(t, "c*.c?.c?")
	dfa := NewQueryDFA(&q, false)
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
	dfa := NewQueryDFA(&q, false)
	results := dfa.Find(json)
	if len(results) != 3 {
		t.Fatalf("expected 3 matches, got %d", len(results))
	}
}

func TestDFADisjunctionGroup(t *testing.T) {
	input := `{"x": {"y": 5, "z": { "t": 2}}}`
	json := mustParseJSON(t, input)
	q := mustParseQuery(t, "x.(y | z.t)")
	dfa := NewQueryDFA(&q, false)
	results := dfa.Find(json)
	if len(results) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(results))
	}
}

func TestDFARecursiveArrayIndexing(t *testing.T) {
	input := `[[1], [2, 3]]`
	js := mustParseAnyJSON(t, input)
	q := mustParseQuery(t, "[*]*")
	dfa := NewQueryDFA(&q, false)
	results := dfa.Find(js)
	if len(results) != 6 {
		t.Fatalf("expected 6 matches, got %d", len(results))
	}
}

func TestDFARecursiveArrayIndexingAnyLevel(t *testing.T) {
	input := `[[1], [2, 3]]`
	json := mustParseAnyJSON(t, input)
	q := mustParseQuery(t, "**.[*]*.[*]")
	dfa := NewQueryDFA(&q, false)
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
	dfa := NewQueryDFA(&q, false)
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
	dfa := NewQueryDFA(&q, false)
	results := dfa.Find(json)
	if len(results) != 5 {
		t.Fatalf("expected 5 matches, got %d", len(results))
	}
}

func TestDFAOverlappingRanges(t *testing.T) {
	json := mustParseJSON(t, simpleJSON)
	q := mustParseQuery(t, "baz[0:3] | baz[1:]")
	dfa := NewQueryDFA(&q, false)
	results := dfa.Find(json)
	if len(results) != 5 {
		t.Fatalf("expected 5 matches (union of [0:3] and [1:]), got %d", len(results))
	}
}

func TestDFAConstruction(t *testing.T) {
	q := mustParseQuery(t, "foo.bar")
	dfa := NewQueryDFA(&q, false)

	if dfa.NumStates != 3 {
		t.Fatalf("expected 3 states, got %d", dfa.NumStates)
	}
	if !dfa.IsAccepting[2] {
		t.Fatal("state 2 should be accepting")
	}
	if dfa.IsAccepting[0] || dfa.IsAccepting[1] {
		t.Fatal("states 0 and 1 should not be accepting")
	}
	if _, ok := dfa.KeyToID["foo"]; !ok {
		t.Fatal("expected 'foo' in key_to_key_id")
	}
	if _, ok := dfa.KeyToID["bar"]; !ok {
		t.Fatal("expected 'bar' in key_to_key_id")
	}
}

func TestDFACaseInsensitive(t *testing.T) {
	input := `{ "FOO": 1, "bar": 2 }`
	json := mustParseJSON(t, input)
	q := mustParseQuery(t, "foo")
	dfa := NewQueryDFA(&q, true)
	results := dfa.Find(json)
	if len(results) != 1 {
		t.Fatalf("expected 1 match, got %d", len(results))
	}
}

func TestDFACaseInsensitiveSequence(t *testing.T) {
	input := `{ "Foo": { "BAR": "found" } }`
	json := mustParseJSON(t, input)
	q := mustParseQuery(t, "foo.bar")
	dfa := NewQueryDFA(&q, true)
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
	dfa := NewQueryDFA(&q, true)
	results := dfa.Find(json)
	if len(results) != 1 {
		t.Fatalf("expected 1 match (deduped), got %d", len(results))
	}
}

func TestDFACaseSensitiveDefault(t *testing.T) {
	input := `{ "FOO": 1, "foo": 2 }`
	json := mustParseJSON(t, input)
	q := mustParseQuery(t, "foo")
	dfa := NewQueryDFA(&q, false)
	results := dfa.Find(json)
	if len(results) != 1 {
		t.Fatalf("expected 1 match, got %d", len(results))
	}
	if n, ok := results[0].Value.(int); !ok || n != 2 {
		t.Fatalf("expected 2, got %v", results[0].Value)
	}
}

func TestDFAQuotedFieldWithSlash(t *testing.T) {
	input := `{ "/activities": { "get": "list" } }`
	json := mustParseJSON(t, input)
	q := mustParseQuery(t, `"/activities"`)
	dfa := NewQueryDFA(&q, false)
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
	dfa := NewQueryDFA(&q, false)
	results := dfa.Find(json)
	if len(results) != 1 {
		t.Fatalf("expected 1 match, got %d", len(results))
	}
}

func TestDFAQuotedFieldWithDot(t *testing.T) {
	input := `{ "a.b": 42, "a": { "b": 99 } }`
	json := mustParseJSON(t, input)
	q := mustParseQuery(t, `"a.b"`)
	dfa := NewQueryDFA(&q, false)
	results := dfa.Find(json)
	if len(results) != 1 {
		t.Fatalf("expected 1 match, got %d", len(results))
	}
	if n, ok := results[0].Value.(int); !ok || n != 42 {
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
	dfa := NewQueryDFA(&q, false)
	results := dfa.Find(json)
	if len(results) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(results))
	}
}

func TestDFAFieldSymbolIDLookup(t *testing.T) {
	q := mustParseQuery(t, "foo.bar")
	dfa := NewQueryDFA(&q, false)
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
	dfa := NewQueryDFA(&q, false)

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
	dfa := NewQueryDFA(&q, false)
	fooID := dfa.FieldSymbolID("foo")
	barID := dfa.FieldSymbolID("bar")
	otherID := dfa.FieldSymbolID("baz")

	next, ok := dfa.Transition(0, fooID)
	if !ok {
		t.Fatal("expected transition on foo from state 0")
	}
	next2, ok := dfa.Transition(next, barID)
	if !ok || !dfa.IsAccepting[next2] {
		t.Fatal("expected transition on bar leading to accepting state")
	}
	_, ok = dfa.Transition(0, otherID)
	if ok {
		t.Fatal("expected no transition on 'baz' from state 0")
	}
}

func TestDFANoRangeOverlaps(t *testing.T) {
	q := mustParseQuery(t, "foo[1:5].baz[2]")
	dfa := NewQueryDFA(&q, false)
	prevEnd := 0
	for _, re := range dfa.Ranges {
		if re.Start < prevEnd {
			t.Fatalf("overlapping range detected: [%d,%d) overlaps with previous end %d", re.Start, re.End, prevEnd)
		}
		prevEnd = re.End
	}
}

func TestDFARangeFromLookup(t *testing.T) {
	q := mustParseQuery(t, "baz[3:]")
	dfa := NewQueryDFA(&q, false)
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
	dfa := NewQueryDFA(&q, false)
	results := dfa.Find(json)
	if len(results) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(results))
	}
}

func TestDFATwoFieldWildcards(t *testing.T) {
	input := `{ "root": { "foo": "bar" } }`
	json := mustParseJSON(t, input)
	q := mustParseQuery(t, "*.*")
	dfa := NewQueryDFA(&q, false)
	results := dfa.Find(json)
	if len(results) != 1 {
		t.Fatalf("expected 1 match, got %d", len(results))
	}
}

func TestDFAArrayObjNoFields(t *testing.T) {
	input := `[{"root": { "foo": "bar" }}]`
	json := mustParseAnyJSON(t, input)
	q := mustParseQuery(t, "*.*")
	dfa := NewQueryDFA(&q, false)
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
	dfa := NewQueryDFA(&q, true)
	results := dfa.Find(json)
	if len(results) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(results))
	}
}
