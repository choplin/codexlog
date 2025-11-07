package model

// ContentBlock models a portion of a response payload.
// This is agent-agnostic and used by all implementations.
type ContentBlock struct {
	Type string
	Text string
}
