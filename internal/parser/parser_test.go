package parser

import (
	"path/filepath"
	"testing"
	"time"

	"agentlog/internal/model"
)

func fixturePath(parts ...string) string {
	elems := append([]string{"..", "..", "testdata", "sessions"}, parts...)
	return filepath.Join(elems...)
}

func TestReadSessionMeta(t *testing.T) {
	path := fixturePath("sample-simple.jsonl")

	meta, err := ReadSessionMeta(path)
	if err != nil {
		t.Fatalf("ReadSessionMeta returned error: %v", err)
	}

	if meta.ID != "test-simple-session" {
		t.Fatalf("unexpected session id: %s", meta.ID)
	}
	if got := meta.StartedAt.Format(time.RFC3339); got != "2025-11-05T09:00:00Z" {
		t.Fatalf("unexpected start time: %s", got)
	}
	if meta.CWD != "/Users/test/simple" {
		t.Fatalf("unexpected cwd: %s", meta.CWD)
	}
}

func TestFirstUserSummary(t *testing.T) {
	path := fixturePath("sample-simple.jsonl")

	summary, count, last, err := FirstUserSummary(path)
	if err != nil {
		t.Fatalf("FirstUserSummary returned error: %v", err)
	}

	if summary != "Hello, can you help me?" {
		t.Fatalf("unexpected summary: %q", summary)
	}
	if count != 4 {
		t.Fatalf("unexpected message count: %d", count)
	}
	if got := last.Format(time.RFC3339); got != "2025-11-05T09:00:04Z" {
		t.Fatalf("unexpected last timestamp: %s", got)
	}
}

func TestIterateEvents_Filtered(t *testing.T) {
	path := fixturePath("sample-simple.jsonl")

	var events []model.PayloadRole
	err := IterateEvents(path, func(evt model.Event) error {
		if evt.Kind == model.EntryTypeResponseItem {
			events = append(events, evt.Role)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("IterateEvents returned error: %v", err)
	}

	if len(events) != 4 {
		t.Fatalf("expected 4 response events, got %d", len(events))
	}
}
