package claude

import (
	"path/filepath"
	"testing"
	"time"
)

func fixturePath(parts ...string) string {
	elems := append([]string{"..", "..", "testdata", "claude-sessions"}, parts...)
	return filepath.Join(elems...)
}

func TestReadSessionMeta(t *testing.T) {
	path := fixturePath("sample-simple.jsonl")

	meta, err := ReadSessionMeta(path)
	if err != nil {
		t.Fatalf("ReadSessionMeta returned error: %v", err)
	}

	if meta.ID != "test-claude-session" {
		t.Fatalf("unexpected session id: %s", meta.ID)
	}
	if got := meta.StartedAt.Format(time.RFC3339); got != "2025-01-05T10:00:00Z" {
		t.Fatalf("unexpected start time: %s", got)
	}
	if meta.CWD != "/Users/test/project" {
		t.Fatalf("unexpected cwd: %s", meta.CWD)
	}
	if meta.Version != "1.0.35" {
		t.Fatalf("unexpected version: %s", meta.Version)
	}
}

func TestFirstUserSummary(t *testing.T) {
	path := fixturePath("sample-simple.jsonl")

	summary, count, last, err := FirstUserSummary(path)
	if err != nil {
		t.Fatalf("FirstUserSummary returned error: %v", err)
	}

	if summary != "What is Python?" {
		t.Fatalf("unexpected summary: %q", summary)
	}
	if count != 4 {
		t.Fatalf("unexpected message count: %d", count)
	}
	if got := last.Format(time.RFC3339); got != "2025-01-05T10:00:07Z" {
		t.Fatalf("unexpected last timestamp: %s", got)
	}
}

func TestFirstUserSummary_WithSummaryEntry(t *testing.T) {
	path := fixturePath("sample-with-tools.jsonl")

	summary, count, last, err := FirstUserSummary(path)
	if err != nil {
		t.Fatalf("FirstUserSummary returned error: %v", err)
	}

	// Should use the first user message, not the summary entry
	if summary != "Read the README file" {
		t.Fatalf("unexpected summary: %q", summary)
	}
	if count != 4 {
		t.Fatalf("unexpected message count: %d", count)
	}
	if got := last.Format(time.RFC3339); got != "2025-01-05T11:00:04Z" {
		t.Fatalf("unexpected last timestamp: %s", got)
	}
}

func TestIterateEvents_Simple(t *testing.T) {
	path := fixturePath("sample-simple.jsonl")

	var events []ClaudeEvent
	err := IterateEvents(path, func(evt ClaudeEvent) error {
		events = append(events, evt)
		return nil
	})
	if err != nil {
		t.Fatalf("IterateEvents returned error: %v", err)
	}

	if len(events) != 4 {
		t.Fatalf("expected 4 events, got %d", len(events))
	}

	// Check first user message
	if events[0].Kind != EntryTypeUser {
		t.Fatalf("expected first event to be user, got %s", events[0].Kind)
	}
	if events[0].Role != "user" {
		t.Fatalf("expected role to be user, got %s", events[0].Role)
	}
	if len(events[0].Content) == 0 || events[0].Content[0].Text != "What is Python?" {
		t.Fatalf("unexpected content in first message")
	}

	// Check first assistant message
	if events[1].Kind != EntryTypeAssistant {
		t.Fatalf("expected second event to be assistant, got %s", events[1].Kind)
	}
	if events[1].MessageID != "msg_01abc" {
		t.Fatalf("unexpected message id: %s", events[1].MessageID)
	}
	if events[1].Model != "claude-sonnet-4-20250514" {
		t.Fatalf("unexpected model: %s", events[1].Model)
	}
	if events[1].Usage == nil {
		t.Fatal("expected usage to be populated")
	}
	if events[1].Usage.InputTokens != 10 {
		t.Fatalf("unexpected input tokens: %d", events[1].Usage.InputTokens)
	}
	if events[1].Usage.OutputTokens != 15 {
		t.Fatalf("unexpected output tokens: %d", events[1].Usage.OutputTokens)
	}
}

func TestIterateEvents_WithTools(t *testing.T) {
	path := fixturePath("sample-with-tools.jsonl")

	var events []ClaudeEvent
	err := IterateEvents(path, func(evt ClaudeEvent) error {
		events = append(events, evt)
		return nil
	})
	if err != nil {
		t.Fatalf("IterateEvents returned error: %v", err)
	}

	if len(events) != 5 {
		t.Fatalf("expected 5 events, got %d", len(events))
	}

	// Check tool use message
	toolUseEvent := events[1]
	if toolUseEvent.Kind != EntryTypeAssistant {
		t.Fatalf("expected assistant event, got %s", toolUseEvent.Kind)
	}
	if len(toolUseEvent.Content) == 0 {
		t.Fatal("expected content to be populated")
	}
	if toolUseEvent.Content[0].Type != "tool_use" {
		t.Fatalf("expected tool_use content, got %s", toolUseEvent.Content[0].Type)
	}

	// Check tool result message
	toolResultEvent := events[2]
	if toolResultEvent.Kind != EntryTypeUser {
		t.Fatalf("expected user event for tool result, got %s", toolResultEvent.Kind)
	}
	if len(toolResultEvent.Content) == 0 {
		t.Fatal("expected content to be populated")
	}
	if toolResultEvent.Content[0].Type != "tool_result" {
		t.Fatalf("expected tool_result content, got %s", toolResultEvent.Content[0].Type)
	}

	// Check summary entry
	summaryEvent := events[4]
	if summaryEvent.Kind != EntryTypeSummary {
		t.Fatalf("expected summary event, got %s", summaryEvent.Kind)
	}
	if summaryEvent.SummaryText != "Reading and discussing README file" {
		t.Fatalf("unexpected summary text: %s", summaryEvent.SummaryText)
	}
	if summaryEvent.LeafUUID != "asst-msg-2" {
		t.Fatalf("unexpected leaf uuid: %s", summaryEvent.LeafUUID)
	}
}
