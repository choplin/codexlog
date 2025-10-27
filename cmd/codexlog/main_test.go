package main

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"codexlog/internal/model"
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

func TestLimitEvents(t *testing.T) {
	events := []model.Event{{}, {}, {}, {}}
	limited := limitEvents(events, 2)
	if len(limited) != 2 {
		t.Fatalf("expected 2 events, got %d", len(limited))
	}
}

func TestPrintEvent(t *testing.T) {
	event := model.Event{
		Kind:      "response_item",
		Role:      "assistant",
		Timestamp: time.Date(2025, 10, 25, 12, 0, 0, 0, time.UTC),
		Content: []model.ContentBlock{
			{Type: "text", Text: "hello world"},
		},
	}

	var buf bytes.Buffer
	printEvent(&buf, event, 3, 40)
	output := buf.String()

	if !strings.Contains(output, "[#003][assistant][2025-10-25T12:00:00Z]") {
		t.Fatalf("header not found in output: %s", output)
	}
	if !strings.Contains(output, "  hello world") {
		t.Fatalf("body not indented correctly: %s", output)
	}
}
