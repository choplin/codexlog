package view

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"agentlog/internal/codex"
	"agentlog/internal/model"
)

func TestBuildViewFiltersDefaults(t *testing.T) {
	filters, err := buildViewFilters(false, "", "", "", "")
	if err != nil {
		t.Fatalf("buildViewFilters returned error: %v", err)
	}
	if filters.entryTypes == nil || len(filters.entryTypes) != 1 {
		t.Fatalf("expected default entry type filter with single value, got %#v", filters.entryTypes)
	}
	if _, ok := filters.entryTypes[codex.EntryTypeResponseItem]; !ok {
		t.Fatalf("default entry type should include response_item")
	}
	if filters.responseItemTypes == nil || len(filters.responseItemTypes) != 1 {
		t.Fatalf("expected default response type filter with single value, got %#v", filters.responseItemTypes)
	}
	if _, ok := filters.responseItemTypes[codex.ResponseItemTypeMessage]; !ok {
		t.Fatalf("default response type should include message")
	}
	if filters.payloadRoles == nil || len(filters.payloadRoles) != 2 {
		t.Fatalf("expected default payload roles for user/assistant, got %#v", filters.payloadRoles)
	}
}

func TestEventMatchesFilters(t *testing.T) {
	filters := viewFilters{
		entryTypes: map[codex.EntryType]struct{}{
			codex.EntryTypeResponseItem: {},
		},
		payloadRoles: map[codex.PayloadRole]struct{}{
			codex.PayloadRoleAssistant: {},
		},
	}

	event := codex.CodexEvent{Kind: codex.EntryTypeResponseItem, Role: codex.PayloadRoleAssistant}
	if !eventMatchesFilters(event, filters) {
		t.Fatalf("expected event to match filters")
	}

	event.Role = codex.PayloadRoleUser
	if eventMatchesFilters(event, filters) {
		t.Fatalf("unexpected match for user role")
	}

	event.Kind = codex.EntryTypeSessionMeta
	if eventMatchesFilters(event, filters) {
		t.Fatalf("session_meta should be filtered out")
	}
}

func TestParsePayloadRoleArgUnknown(t *testing.T) {
	if _, _, err := parsePayloadRoleArg("user,unknown"); err == nil {
		t.Fatalf("expected error for unknown payload role")
	}
}

func TestRenderChatLinesAlignment(t *testing.T) {
	events := []codex.CodexEvent{
		{
			Role:      codex.PayloadRoleUser,
			Timestamp: time.Date(2025, 10, 27, 12, 0, 0, 0, time.UTC),
			Content: []model.ContentBlock{{Type: "text", Text: "hello there"}},
		},
		{
			Role:      codex.PayloadRoleAssistant,
			Timestamp: time.Date(2025, 10, 27, 12, 0, 5, 0, time.UTC),
			Content: []model.ContentBlock{{Type: "text", Text: "hi, how can I help you today?"}},
		},
		{
			Role:      codex.PayloadRoleTool,
			Timestamp: time.Date(2025, 10, 27, 12, 0, 10, 0, time.UTC),
			Content: []model.ContentBlock{{Type: "json", Text: `{"result":"ok"}`}},
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

func TestRunFormatRaw(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "sessions", "sample-simple.jsonl")
	var buf bytes.Buffer
	opts := Options{
		Path:            path,
		Format:          "raw",
		Out:             &buf,
		EntryTypeArg:    "",
		ResponseTypeArg: "",
		EventMsgTypeArg: "",
		PayloadRoleArg:  "",
	}
	if err := Run(opts); err != nil {
		t.Fatalf("Run returned error: %v", err)
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
	if buf.String() != want {
		t.Fatalf("raw output mismatch\nwant:\n%q\n\ngot:\n%q", want, buf.String())
	}
}

func TestFilterCombinations(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "sessions", "sample-full.jsonl")

	tests := []struct {
		name            string
		entryTypeArg    string
		responseTypeArg string
		eventMsgTypeArg string
		payloadRoleArg  string
		expectedCount   int
		description     string
	}{
		{
			name:            "filter event_msg by token_count",
			entryTypeArg:    "event_msg",
			eventMsgTypeArg: "token_count",
			expectedCount:   3,
			description:     "should show only token_count events",
		},
		{
			name:            "filter response_item by reasoning",
			entryTypeArg:    "response_item",
			responseTypeArg: "reasoning",
			expectedCount:   1,
			description:     "should show only reasoning entries",
		},
		{
			name:            "filter response_item by function_call",
			entryTypeArg:    "response_item",
			responseTypeArg: "function_call,function_call_output",
			payloadRoleArg:  "all",
			expectedCount:   2,
			description:     "should show function_call and function_call_output",
		},
		{
			name:            "filter multiple entry types",
			entryTypeArg:    "response_item,event_msg",
			responseTypeArg: "all",
			payloadRoleArg:  "all",
			expectedCount:   14,
			description:     "should show both response_item and event_msg entries",
		},
		{
			name:            "event_msg ignores response-type filter",
			entryTypeArg:    "event_msg",
			responseTypeArg: "message",
			expectedCount:   6,
			description:     "should show all event_msg entries (response-type filter is ignored)",
		},
		{
			name:          "filter by user role only",
			entryTypeArg:  "response_item",
			payloadRoleArg: "user",
			expectedCount: 2,
			description:   "should show only user messages",
		},
		{
			name:         "filter turn_context",
			entryTypeArg: "turn_context",
			expectedCount: 1,
			description:  "should show turn_context entry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			opts := Options{
				Path:            path,
				Format:          "raw",
				Out:             &buf,
				EntryTypeArg:    tt.entryTypeArg,
				ResponseTypeArg: tt.responseTypeArg,
				EventMsgTypeArg: tt.eventMsgTypeArg,
				PayloadRoleArg:  tt.payloadRoleArg,
			}
			if err := Run(opts); err != nil {
				t.Fatalf("Run returned error: %v", err)
			}

			output := buf.String()
			lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
			var actualCount int
			for _, line := range lines {
				if strings.TrimSpace(line) != "" {
					actualCount++
				}
			}

			if actualCount != tt.expectedCount {
				t.Fatalf("%s: expected %d entries, got %d\nOutput:\n%s",
					tt.description, tt.expectedCount, actualCount, output)
			}
		})
	}
}
