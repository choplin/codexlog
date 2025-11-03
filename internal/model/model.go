package model

import "time"

// EntryType represents the top-level "type" field values observed in Codex JSONL logs.
type EntryType string

const (
	EntryTypeSessionMeta  EntryType = "session_meta"
	EntryTypeResponseItem EntryType = "response_item"
)

// PayloadType captures the "payload.type" values observed in Codex response items.
type PayloadType string

const (
	PayloadTypeMessage     PayloadType = "message"
	PayloadTypeEventMsg    PayloadType = "event_msg"
	PayloadTypeTurnContext PayloadType = "turn_context"
)

// PayloadRole captures the "payload.role" values observed in Codex response items.
type PayloadRole string

const (
	PayloadRoleUser      PayloadRole = "user"
	PayloadRoleAssistant PayloadRole = "assistant"
	PayloadRoleTool      PayloadRole = "tool"
	PayloadRoleSystem    PayloadRole = "system"
)

// SessionSummary holds lightweight information about a Codex session.
type SessionSummary struct {
	ID              string
	Path            string
	CWD             string
	Originator      string
	CLIVersion      string
	StartedAt       time.Time
	Summary         string
	MessageCount    int
	DurationSeconds int
}

// SessionMeta represents metadata stored in the session_meta payload.
type SessionMeta struct {
	ID         string
	Path       string
	CWD        string
	Originator string
	CLIVersion string
	StartedAt  time.Time
}

// Event represents a single entry in the session JSONL stream.
type Event struct {
	Timestamp   time.Time
	Kind        EntryType
	Role        PayloadRole
	MessageType PayloadType
	Content     []ContentBlock
	Raw         string
}

// ContentBlock models a portion of a response payload.
type ContentBlock struct {
	Type string
	Text string
}
