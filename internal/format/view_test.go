package format

import (
	"strings"
	"testing"
	"time"

	"codexlog/internal/model"
)

func TestRenderEventLines_Text(t *testing.T) {
	event := model.Event{
		Kind: "response_item",
		Role: "assistant",
		Content: []model.ContentBlock{
			{Type: "text", Text: "one two three four five six"},
		},
	}

	lines := RenderEventLines(event, 10)
	if len(lines) < 2 {
		t.Fatalf("expected wrapped lines, got %v", lines)
	}
	if strings.TrimSpace(lines[0]) == "" {
		t.Fatalf("first line should contain text: %v", lines)
	}
}

func TestRenderEventLines_JSON(t *testing.T) {
	event := model.Event{
		Kind:      "event_msg",
		Timestamp: time.Date(2025, 10, 25, 12, 0, 0, 0, time.UTC),
		Content: []model.ContentBlock{
			{Type: "json", Text: `{"foo":1,"bar":{"baz":2}}`},
		},
	}

	lines := RenderEventLines(event, 80)
	if len(lines) < 2 {
		t.Fatalf("expected pretty-printed JSON lines, got %v", lines)
	}
	if !strings.HasPrefix(strings.TrimSpace(lines[0]), "{") {
		t.Fatalf("first line should start with '{': %v", lines[0])
	}
	if !strings.HasPrefix(lines[1], "  ") {
		t.Fatalf("json indentation missing: %v", lines[1])
	}
}
