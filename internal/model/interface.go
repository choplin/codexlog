// Package model provides common interfaces and types for agent log implementations.
package model

import "time"

// SessionSummaryProvider provides common session summary information.
// Different agent implementations (Codex, Claude) can provide agent-specific
// fields while sharing this common interface.
type SessionSummaryProvider interface {
	GetID() string
	GetPath() string
	GetCWD() string
	GetStartedAt() time.Time
	GetSummary() string
	GetMessageCount() int
	GetDurationSeconds() int
}

// SessionMetaProvider provides common session metadata.
// Different agent implementations can extend this with agent-specific metadata.
type SessionMetaProvider interface {
	GetID() string
	GetPath() string
	GetCWD() string
	GetStartedAt() time.Time
}

// EventProvider provides common event information.
// Different agent implementations can have different internal structures
// while exposing these common fields for display and filtering.
type EventProvider interface {
	GetTimestamp() time.Time
	GetRole() string // Normalized role: "user", "assistant", "tool", "system"
	GetContent() []ContentBlock
	GetRaw() string // Raw JSON for debugging/export
}
