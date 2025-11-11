// Package model provides common interfaces and types for agent log implementations.
package model

import (
	"fmt"
)

// AgentType represents the type of AI agent.
type AgentType string

const (
	// AgentCodex represents the Codex CLI agent.
	AgentCodex AgentType = "codex"
	// AgentClaude represents the Claude Code agent.
	AgentClaude AgentType = "claude"
)

// ParserFactory is a function type that creates a Parser.
// We use this to avoid circular dependencies between model and agent packages.
type ParserFactory func() Parser

var (
	codexFactory  ParserFactory
	claudeFactory ParserFactory
)

// RegisterCodexParser registers the Codex parser factory.
func RegisterCodexParser(factory ParserFactory) {
	codexFactory = factory
}

// RegisterClaudeParser registers the Claude parser factory.
func RegisterClaudeParser(factory ParserFactory) {
	claudeFactory = factory
}

// NewParser creates a parser for the specified agent type.
func NewParser(agentType AgentType) (Parser, error) {
	switch agentType {
	case AgentCodex:
		if codexFactory == nil {
			return nil, fmt.Errorf("codex parser not registered")
		}
		return codexFactory(), nil
	case AgentClaude:
		if claudeFactory == nil {
			return nil, fmt.Errorf("claude parser not registered")
		}
		return claudeFactory(), nil
	default:
		return nil, fmt.Errorf("unknown agent type: %s", agentType)
	}
}
