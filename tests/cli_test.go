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
