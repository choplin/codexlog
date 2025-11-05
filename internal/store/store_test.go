package store

import (
	"path/filepath"
	"testing"
	"time"
)

func TestListSessions(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "sessions")

	res, err := ListSessions(ListOptions{Root: root, MaxSummary: 80})
	if err != nil {
		t.Fatalf("ListSessions returned error: %v", err)
	}

	if len(res.Summaries) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(res.Summaries))
	}

	// Both sessions have same timestamp, order may vary
	ids := map[string]bool{}
	for _, s := range res.Summaries {
		ids[s.ID] = true
	}

	if !ids["test-full-session"] {
		t.Fatalf("expected test-full-session in results")
	}

	if !ids["test-simple-session"] {
		t.Fatalf("expected test-simple-session in results")
	}

	if len(res.Warnings) != 0 {
		t.Fatalf("expected no warnings, got %d", len(res.Warnings))
	}
}

func TestListSessionsFilters(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "sessions")
	after := time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC)

	res, err := ListSessions(ListOptions{Root: root, After: &after})
	if err != nil {
		t.Fatalf("ListSessions returned error: %v", err)
	}

	if len(res.Summaries) != 2 {
		t.Fatalf("expected 2 summaries after 2025-11-01, got %d", len(res.Summaries))
	}

	if res.Summaries[0].DurationSeconds == 0 {
		t.Fatalf("expected duration to be populated")
	}
}

func TestFindSessionPath(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "sessions")
	path, err := FindSessionPath(root, "test-simple-session")
	if err != nil {
		t.Fatalf("FindSessionPath returned error: %v", err)
	}

	expected := filepath.Join(root, "sample-simple.jsonl")
	if path != expected {
		t.Fatalf("unexpected path: %s", path)
	}
}

func TestListSessionsExactCWD(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "sessions")
	res, err := ListSessions(ListOptions{Root: root, CWD: "/Users/test/project", ExactCWD: true})
	if err != nil {
		t.Fatalf("ListSessions returned error: %v", err)
	}

	if len(res.Summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(res.Summaries))
	}

	summary := res.Summaries[0]
	if summary.ID != "test-full-session" {
		t.Fatalf("unexpected session id: %s", summary.ID)
	}
	if summary.DurationSeconds == 0 {
		t.Fatalf("expected duration to be populated")
	}
}
