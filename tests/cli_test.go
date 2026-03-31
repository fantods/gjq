package tests

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

func runGjq(args ...string) *exec.Cmd {
	return exec.Command("gjq", args...)
}

func runGjqSuccess(t *testing.T, args ...string) string {
	t.Helper()
	cmd := runGjq(args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("expected success but got error: %v\noutput: %s", err, output)
	}
	return string(output)
}

func runGjqFailure(t *testing.T, args ...string) {
	t.Helper()
	cmd := runGjq(args...)
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected failure but succeeded\noutput: %s", output)
	}
}

func TestNonexistentFieldSimpleQuery(t *testing.T) {
	output := runGjqSuccess(t, "does.not.exist", "data/simple.json")
	if strings.TrimSpace(output) != "" {
		t.Errorf("expected no output for nonexistent field, got: %q", output)
	}
}

func TestNonexistentFile(t *testing.T) {
	runGjqFailure(t, "", "")
}

func TestInvalidQuery(t *testing.T) {
	cmd := runGjq("unclosed\"", "data/simple.json")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected failure but succeeded\noutput: %s", output)
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 1 {
			t.Errorf("expected exit code 1, got %d", exitErr.ExitCode())
		}
	}
}

func TestSimpleQuery(t *testing.T) {
	output := runGjqSuccess(t, "age", "data/simple.json", "--with-path")

	lines := strings.Split(output, "\n")
	if len(lines) < 1 {
		t.Fatal("expected at least one line of output")
	}

	pathLine := strings.TrimRight(lines[0], "\r")
	if pathLine != "age:" {
		t.Errorf("expected path header %q, got %q", "age:", pathLine)
	}

	remaining := strings.Join(lines[1:], "\n")
	var outputVal interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(remaining)), &outputVal); err != nil {
		t.Fatalf("failed to parse output JSON: %v\noutput was: %q", err, remaining)
	}

	var expectedVal interface{}
	if err := json.Unmarshal([]byte("32"), &expectedVal); err != nil {
		t.Fatalf("failed to parse expected JSON: %v", err)
	}

	if outputVal != expectedVal {
		t.Errorf("expected %v, got %v", expectedVal, outputVal)
	}
}

func TestQuotedFieldQueryMatches(t *testing.T) {
	output := runGjqSuccess(t, `paths."/activities"`, "data/openapi_paths.json", "--count", "--no-display")
	if !strings.Contains(output, "1") {
		t.Errorf("expected 1 match for quoted field query, got: %q", output)
	}
}

func TestFixedStringFindsKeyAtAnyDepth(t *testing.T) {
	output := runGjqSuccess(t, "-F", "/activities", "data/openapi_paths.json")
	if strings.TrimSpace(output) == "" {
		t.Error("expected output for -F '/activities', got empty")
	}
}

func TestFixedStringCount(t *testing.T) {
	output := runGjqSuccess(t, "-F", "/activities", "data/openapi_paths.json", "--count", "--no-display")
	if !strings.Contains(output, "1") {
		t.Errorf("expected exactly 1 match for -F '/activities', got: %q", output)
	}
}

func TestFixedStringNoMatch(t *testing.T) {
	output := runGjqSuccess(t, "-F", "/nonexistent", "data/openapi_paths.json")
	if strings.TrimSpace(output) != "" {
		t.Errorf("expected no output for nonexistent fixed string, got: %q", output)
	}
}

func TestNoPathFlagSuppressesHeaders(t *testing.T) {
	output := runGjqSuccess(t, "age", "data/simple.json", "--no-path")
	if strings.Contains(output, "age:") {
		t.Errorf("expected no path header with --no-path, got: %q", output)
	}
	if !strings.Contains(output, "32") {
		t.Errorf("expected value to be present, got: %q", output)
	}
}

func TestWithPathFlagShowsHeaders(t *testing.T) {
	output := runGjqSuccess(t, "age", "data/simple.json", "--with-path")
	if !strings.Contains(output, "age:") {
		t.Errorf("expected path header with --with-path, got: %q", output)
	}
}

func TestPathFlagsAreMutuallyExclusive(t *testing.T) {
	runGjqFailure(t, "age", "data/simple.json", "--with-path", "--no-path")
}

func TestCountFlag(t *testing.T) {
	output := runGjqSuccess(t, "age", "data/simple.json", "--count", "--no-display")
	if !strings.Contains(output, "Found matches: 1") {
		t.Errorf("expected 'Found matches: 1', got: %q", output)
	}
	if strings.Contains(output, "32") {
		t.Errorf("expected no value output with --count --no-display, got: %q", output)
	}
}

func TestNoDisplayFlag(t *testing.T) {
	output := runGjqSuccess(t, "age", "data/simple.json", "--no-display")
	if strings.TrimSpace(output) != "" {
		t.Errorf("expected no output with --no-display, got: %q", output)
	}
}

func TestCountWithNoDisplay(t *testing.T) {
	output := runGjqSuccess(t, "age", "data/simple.json", "--count", "--no-display")
	if !strings.Contains(output, "Found matches: 1") {
		t.Errorf("expected count line, got: %q", output)
	}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 1 {
		t.Errorf("expected exactly 1 line (count only), got %d lines: %q", len(lines), output)
	}
}

func TestDepthFlag(t *testing.T) {
	output := runGjqSuccess(t, "--depth", "--no-display", "*", "data/nested.json")
	if !strings.Contains(output, "Depth: 5") {
		t.Errorf("expected 'Depth: 5', got: %q", output)
	}
}

func TestDepthFlagSimple(t *testing.T) {
	output := runGjqSuccess(t, "--depth", "--no-display", "*", "data/simple.json")
	if !strings.Contains(output, "Depth: 3") {
		t.Errorf("expected 'Depth: 3', got: %q", output)
	}
}

func TestCompactFlag(t *testing.T) {
	output := runGjqSuccess(t, "name", "data/simple.json", "--compact", "--no-path")
	if strings.Contains(output, "\n  ") {
		t.Errorf("compact output should not contain indented newlines, got: %q", output)
	}
	var parsed interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &parsed); err != nil {
		t.Fatalf("output should be valid JSON: %v\noutput: %q", err, output)
	}
}

func TestIgnoreCaseFlag(t *testing.T) {
	output := runGjqSuccess(t, "-i", "users[*].address", "data/nested.json", "--with-path")
	if !strings.Contains(output, "Address:") {
		t.Errorf("expected case-insensitive match on 'Address', got: %q", output)
	}
}

func TestFixedStringMultipleMatches(t *testing.T) {
	output := runGjqSuccess(t, "-F", "name", "data/nested.json", "--count", "--no-display")
	if !strings.Contains(output, "Found matches: 2") {
		t.Errorf("expected 2 matches for -F 'name', got: %q", output)
	}
}

func TestLargeFileWildcardQuery(t *testing.T) {
	output := runGjqSuccess(t, "results[*].nat", "data/randomusers.json", "--count", "--no-display")
	if !strings.Contains(output, "Found matches: 1000") {
		t.Errorf("expected 1000 nationalities, got: %q", output)
	}
}

func TestLargeFileDeepQuery(t *testing.T) {
	output := runGjqSuccess(t, "results[0].name.last", "data/randomusers.json", "--with-path")
	if strings.TrimSpace(output) == "" {
		t.Error("expected output for deep query into randomuser data, got empty")
	}
	if !strings.Contains(output, "Banerjee") {
		t.Errorf("expected 'Banerjee' in output, got: %q", output)
	}
}
