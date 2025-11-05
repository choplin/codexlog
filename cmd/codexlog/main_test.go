package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClipSummary(t *testing.T) {
	if got := clipSummary("abcdef", 3); got != "ab…" {
		t.Fatalf("clipSummary unexpected result: %q", got)
	}
	if got := clipSummary("short", 10); got != "short" {
		t.Fatalf("clipSummary should not alter short text: %q", got)
	}
}

func TestCollapseWhitespace(t *testing.T) {
	text := "  line one\n\nline\t two  "
	if got := collapseWhitespace(text); got != "line one line two" {
		t.Fatalf("collapseWhitespace failed: %q", got)
	}
}

func TestViewCommandFormatRaw(t *testing.T) {
	cmd := newViewCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(io.Discard)
	path := filepath.Join("..", "..", "testdata", "sessions", "sample-simple.jsonl")
	cmd.SetArgs([]string{path, "--format", "raw"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("view command failed: %v", err)
	}
	wantBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read sample file: %v", err)
	}
	var wantLines []string
	for _, line := range strings.Split(strings.TrimRight(string(wantBytes), "\n"), "\n") {
		if strings.Contains(line, "\"type\":\"session_meta\"") {
			continue
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		wantLines = append(wantLines, line)
	}
	want := strings.Join(wantLines, "\n") + "\n"
	if got := buf.String(); got != want {
		t.Fatalf("raw output mismatch\nwant:\n%q\n\ngot:\n%q", want, got)
	}
}
