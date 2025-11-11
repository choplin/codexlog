// Package model provides common interfaces and types for agent log implementations.
package model

// Parser defines the common interface for parsing agent session logs.
// Each agent implementation (Codex, Claude) provides its own parser
// that conforms to this interface.
type Parser interface {
	// ReadSessionMeta reads basic session metadata from the log file.
	ReadSessionMeta(path string) (SessionMetaProvider, error)

	// FirstUserSummary extracts the first user message or summary from the log file.
	// This is used for displaying a brief description of the session.
	FirstUserSummary(path string) (string, error)

	// IterateEvents reads all events from the log file and calls the provided
	// function for each event. The function should return an error to stop iteration.
	IterateEvents(path string, fn func(EventProvider) error) error
}
