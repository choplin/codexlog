package model

import "time"

// EntryType represents the top-level "type" field values observed in Codex JSONL logs.
type EntryType string

const (
	EntryTypeSessionMeta  EntryType = "session_meta"
	EntryTypeResponseItem EntryType = "response_item"
	EntryTypeEventMsg     EntryType = "event_msg"
	EntryTypeTurnContext  EntryType = "turn_context"
)

// ResponseItemType captures the "payload.type" values in response_item entries.
type ResponseItemType string

const (
	ResponseItemTypeMessage              ResponseItemType = "message"
	ResponseItemTypeReasoning            ResponseItemType = "reasoning"
	ResponseItemTypeFunctionCall         ResponseItemType = "function_call"
	ResponseItemTypeFunctionCallOutput   ResponseItemType = "function_call_output"
	ResponseItemTypeCustomToolCall       ResponseItemType = "custom_tool_call"
	ResponseItemTypeCustomToolCallOutput ResponseItemType = "custom_tool_call_output"
)

// EventMsgType captures the "payload.type" values in event_msg entries.
type EventMsgType string

const (
	EventMsgTypeTokenCount     EventMsgType = "token_count"
	EventMsgTypeAgentReasoning EventMsgType = "agent_reasoning"
	EventMsgTypeUserMessage    EventMsgType = "user_message"
	EventMsgTypeAgentMessage   EventMsgType = "agent_message"
	EventMsgTypeTurnAborted    EventMsgType = "turn_aborted"
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
	PayloadType string // response_item: ResponseItemType, event_msg: EventMsgType
	Content     []ContentBlock
	Raw         string
}

// ContentBlock models a portion of a response payload.
type ContentBlock struct {
	Type string
	Text string
}
