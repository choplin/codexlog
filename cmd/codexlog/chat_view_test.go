package main

import (
	"strings"
	"testing"
	"time"

	"codexlog/internal/model"
)

func TestRenderChatLinesAlignment(t *testing.T) {
	events := []model.Event{
		{
			Role:      model.PayloadRoleUser,
			Timestamp: time.Date(2025, 10, 27, 12, 0, 0, 0, time.UTC),
			Content: []model.ContentBlock{
				{Type: "text", Text: "hello there"},
			},
		},
		{
			Role:      model.PayloadRoleAssistant,
			Timestamp: time.Date(2025, 10, 27, 12, 0, 5, 0, time.UTC),
			Content: []model.ContentBlock{
				{Type: "text", Text: "hi, how can I help you today?"},
			},
		},
		{
			Role:      model.PayloadRoleTool,
			Timestamp: time.Date(2025, 10, 27, 12, 0, 10, 0, time.UTC),
			Content: []model.ContentBlock{
				{Type: "json", Text: `{"result":"ok"}`},
			},
		},
	}

	lines := renderChatTranscript(events, 80, false)
	if len(lines) == 0 {
		t.Fatal("expected chat lines")
	}

	userTop := findPrefix(lines, "╭")
	if userTop < 0 {
		t.Fatalf("failed to locate user bubble: %v", lines)
	}

	next := findPrefix(lines[userTop+1:], "╭")
	if next < 0 {
		t.Fatalf("failed to locate assistant bubble: %v", lines)
	}
	assistantTop := next + userTop + 1

	if idx := strings.Index(lines[userTop], "╭"); idx <= 2 {
		t.Fatalf("user bubble should be right aligned, got index %d line %q", idx, lines[userTop])
	}

	if !strings.HasPrefix(lines[assistantTop], "  ╭") {
		t.Fatalf("assistant bubble should be left aligned: %q", lines[assistantTop])
	}
}

func findPrefix(lines []string, prefix string) int {
	for i, line := range lines {
		if strings.HasPrefix(line, prefix) || strings.Contains(line, prefix) {
			return i
		}
	}
	return -1
}
