package invoker

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"pgregory.net/rapid"
)

// ---- helpers ----------------------------------------------------------------

// largePayload returns a byte slice of exactly n bytes of valid JSON.
func largePayload(n int) []byte {
	// Build a JSON string value long enough to hit n bytes total.
	// `{"data":"<padding>"}` — adjust padding length accordingly.
	prefix := `{"data":"`
	suffix := `"}`
	padding := n - len(prefix) - len(suffix)
	if padding < 0 {
		padding = 0
	}
	return []byte(prefix + strings.Repeat("x", padding) + suffix)
}

// fileCount returns how many files matching the glob exist in dir.
func fileCount(t *testing.T, dir, glob string) int {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(dir, glob))
	if err != nil {
		t.Fatalf("glob error: %v", err)
	}
	return len(matches)
}

// ---- unit tests: writeOutputToFile ------------------------------------------

func TestWriteOutputToFile_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	data := []byte(`{"hello":"world"}`)

	filename, err := writeOutputToFile(data, "json", dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("could not read created file: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("file content mismatch: got %q, want %q", got, data)
	}
}

func TestWriteOutputToFile_NamePattern(t *testing.T) {
	dir := t.TempDir()
	filename, err := writeOutputToFile([]byte("{}"), "json", dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	base := filepath.Base(filename)
	pattern := regexp.MustCompile(`^gateway-output-\d{8}-\d{6}\.json$`)
	if !pattern.MatchString(base) {
		t.Errorf("filename %q does not match expected pattern gateway-output-YYYYMMDD-HHMMSS.json", base)
	}
}

func TestWriteOutputToFile_ExtensionPreserved(t *testing.T) {
	for _, ext := range []string{"json", "md", "txt"} {
		dir := t.TempDir()
		filename, err := writeOutputToFile([]byte("data"), ext, dir)
		if err != nil {
			t.Fatalf("ext=%s: unexpected error: %v", ext, err)
		}
		if !strings.HasSuffix(filename, "."+ext) {
			t.Errorf("ext=%s: filename %q does not end with .%s", ext, filename, ext)
		}
	}
}

func TestWriteOutputToFile_EmptyDirMeansCurrentDir(t *testing.T) {
	// Change to a temp dir so we don't pollute the repo.
	dir := t.TempDir()
	orig, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(orig) }()

	filename, err := writeOutputToFile([]byte("{}"), "json", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, statErr := os.Stat(filename); statErr != nil {
		t.Errorf("file %q not found in current dir: %v", filename, statErr)
	}
}

func TestWriteOutputToFile_BadDir(t *testing.T) {
	_, err := writeOutputToFile([]byte("{}"), "json", "/nonexistent/dir/that/does/not/exist")
	if err == nil {
		t.Error("expected an error for a non-existent directory, got nil")
	}
}

// ---- unit tests: routeOutput ------------------------------------------------

func TestRouteOutput_SmallPayload_Terminal(t *testing.T) {
	dir := t.TempDir()
	var stdout, stderr bytes.Buffer
	payload := []byte(`{"small":true}`)

	if err := routeOutput(payload, "json", &stdout, &stderr, true, dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout.String(), `"small":true`) {
		t.Errorf("expected payload in stdout, got: %q", stdout.String())
	}
	if fileCount(t, dir, "gateway-output-*.json") != 0 {
		t.Error("expected no output file for small payload, but one was created")
	}
}

func TestRouteOutput_SmallPayload_Pipe(t *testing.T) {
	dir := t.TempDir()
	var stdout, stderr bytes.Buffer
	payload := []byte(`{"small":true}`)

	if err := routeOutput(payload, "json", &stdout, &stderr, false, dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stdout.Len() == 0 {
		t.Error("expected payload in stdout for pipe mode, got nothing")
	}
	if fileCount(t, dir, "gateway-output-*.json") != 0 {
		t.Error("expected no output file in pipe mode")
	}
}

func TestRouteOutput_LargePayload_Terminal(t *testing.T) {
	dir := t.TempDir()
	var stdout, stderr bytes.Buffer
	payload := largePayload(maxOutputChars + 1)

	if err := routeOutput(payload, "json", &stdout, &stderr, true, dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stdout.Len() != 0 {
		t.Errorf("expected empty stdout for large terminal output, got %d bytes", stdout.Len())
	}
	if fileCount(t, dir, "gateway-output-*.json") != 1 {
		t.Error("expected exactly one output file for large terminal payload")
	}
	if !strings.Contains(stderr.String(), "Output too large") {
		t.Errorf("expected 'Output too large' notice on stderr, got: %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "gateway-output-") {
		t.Errorf("expected filename in stderr notice, got: %q", stderr.String())
	}
}

func TestRouteOutput_LargePayload_Pipe(t *testing.T) {
	dir := t.TempDir()
	var stdout, stderr bytes.Buffer
	payload := largePayload(maxOutputChars + 1)

	if err := routeOutput(payload, "json", &stdout, &stderr, false, dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stdout.Len() == 0 {
		t.Error("expected payload passed through to stdout in pipe mode")
	}
	if fileCount(t, dir, "gateway-output-*.json") != 0 {
		t.Error("expected no output file in pipe mode even for large payload")
	}
}

func TestRouteOutput_LargePayload_WriteFailure_FallsBackToStdout(t *testing.T) {
	badDir := "/nonexistent/dir/that/does/not/exist"
	var stdout, stderr bytes.Buffer
	payload := largePayload(maxOutputChars + 1)

	if err := routeOutput(payload, "json", &stdout, &stderr, true, badDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stdout.Len() == 0 {
		t.Error("expected fallback to stdout when file write fails")
	}
	if !strings.Contains(stderr.String(), "Warning: could not write output file") {
		t.Errorf("expected warning on stderr, got: %q", stderr.String())
	}
}

func TestRouteOutput_ExactlyAtThreshold_NotWrittenToFile(t *testing.T) {
	// len(out) == maxOutputChars should NOT trigger file write (condition is strictly >)
	dir := t.TempDir()
	var stdout, stderr bytes.Buffer
	payload := make([]byte, maxOutputChars)
	for i := range payload {
		payload[i] = 'x'
	}

	if err := routeOutput(payload, "json", &stdout, &stderr, true, dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fileCount(t, dir, "gateway-output-*.json") != 0 {
		t.Error("payload exactly at threshold should not be written to file")
	}
	if stdout.Len() == 0 {
		t.Error("expected payload at threshold to be printed to stdout")
	}
}

func TestRouteOutput_StderrNoticeContainsSizeInfo(t *testing.T) {
	dir := t.TempDir()
	var stdout, stderr bytes.Buffer
	payload := largePayload(maxOutputChars + 1024) // a bit over threshold

	_ = routeOutput(payload, "json", &stdout, &stderr, true, dir)

	notice := stderr.String()
	if !strings.Contains(notice, "KB") {
		t.Errorf("stderr notice should contain KB size, got: %q", notice)
	}
	if !strings.Contains(notice, "tokens") {
		t.Errorf("stderr notice should mention token count, got: %q", notice)
	}
}

// ---- property tests ---------------------------------------------------------

// PropRouteOutput_BelowThreshold: for any payload smaller than maxOutputChars,
// output always goes to stdout regardless of terminal flag.
func TestPropRouteOutput_BelowThreshold(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		size := rapid.IntRange(0, maxOutputChars).Draw(rt, "size")
		terminal := rapid.Bool().Draw(rt, "terminal")

		dir := t.TempDir()
		payload := rapid.SliceOfN(rapid.Byte(), size, size).Draw(rt, "payload")

		var stdout, stderr bytes.Buffer
		if err := routeOutput(payload, "json", &stdout, &stderr, terminal, dir); err != nil {
			rt.Fatalf("unexpected error: %v", err)
		}

		if fileCount(t, dir, "gateway-output-*.json") != 0 {
			rt.Fatalf("no file should be created for payloads below threshold (size=%d, terminal=%v)", size, terminal)
		}
		if stdout.Len() == 0 && len(payload) > 0 {
			rt.Fatalf("stdout should contain output below threshold (size=%d, terminal=%v)", size, terminal)
		}
	})
}

// PropRouteOutput_AboveThreshold_Terminal: for any payload larger than
// maxOutputChars with terminal=true, a file is always created and stdout is empty.
func TestPropRouteOutput_AboveThreshold_Terminal(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		extra := rapid.IntRange(1, 1000).Draw(rt, "extra")
		size := maxOutputChars + extra

		dir := t.TempDir()
		payload := rapid.SliceOfN(rapid.Byte(), size, size).Draw(rt, "payload")

		var stdout, stderr bytes.Buffer
		if err := routeOutput(payload, "json", &stdout, &stderr, true, dir); err != nil {
			rt.Fatalf("unexpected error: %v", err)
		}

		if fileCount(t, dir, "gateway-output-*.json") != 1 {
			rt.Fatalf("exactly one file should be created for large terminal output (size=%d)", size)
		}
		if stdout.Len() != 0 {
			rt.Fatalf("stdout should be empty for large terminal output (size=%d)", size)
		}
	})
}

// PropRouteOutput_AboveThreshold_Pipe: for any payload size with terminal=false,
// output always goes to stdout and no file is ever created.
func TestPropRouteOutput_AboveThreshold_Pipe(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		size := rapid.IntRange(maxOutputChars+1, maxOutputChars+10000).Draw(rt, "size")

		dir := t.TempDir()
		payload := rapid.SliceOfN(rapid.Byte(), size, size).Draw(rt, "payload")

		var stdout, stderr bytes.Buffer
		if err := routeOutput(payload, "json", &stdout, &stderr, false, dir); err != nil {
			rt.Fatalf("unexpected error: %v", err)
		}

		if fileCount(t, dir, "gateway-output-*.json") != 0 {
			rt.Fatalf("no file should be created in pipe mode (terminal=false), size=%d", size)
		}
		if stdout.Len() == 0 {
			rt.Fatalf("stdout should receive output in pipe mode (terminal=false), size=%d", size)
		}
	})
}

// PropWriteOutputToFile_ContentRoundtrip: any bytes written to a file can be
// read back identically.
func TestPropWriteOutputToFile_ContentRoundtrip(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		data := rapid.SliceOf(rapid.Byte()).Draw(rt, "data")
		dir := t.TempDir()

		filename, err := writeOutputToFile(data, "json", dir)
		if err != nil {
			rt.Fatalf("unexpected write error: %v", err)
		}

		got, err := os.ReadFile(filename)
		if err != nil {
			rt.Fatalf("could not read file back: %v", err)
		}

		if !bytes.Equal(got, data) {
			rt.Fatalf("round-trip mismatch: wrote %d bytes, read %d bytes", len(data), len(got))
		}
	})
}

// PropWriteOutputToFile_ExtensionAlwaysPreserved: the returned path always ends
// with the requested extension.
func TestPropWriteOutputToFile_ExtensionAlwaysPreserved(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		ext := rapid.SampledFrom([]string{"json", "md", "txt", "log"}).Draw(rt, "ext")
		dir := t.TempDir()

		filename, err := writeOutputToFile([]byte("x"), ext, dir)
		if err != nil {
			rt.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasSuffix(filename, fmt.Sprintf(".%s", ext)) {
			rt.Fatalf("filename %q does not end with .%s", filename, ext)
		}
	})
}
