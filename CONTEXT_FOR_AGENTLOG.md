---
created: 2025-11-06
updated: 2025-11-06
---

# Context for agentlog Integration

## Background

This document provides context for integrating Claude Code session browsing functionality into the agentlog repository.

## Current Situation

### cclog Repository

- **Location**: `/Users/aki/workspace/cclog`
- **Original Implementation**: Shell (bash) + Python hybrid architecture
- **Purpose**: Browse Claude Code conversation history with fzf integration
- **Status**: Being refactored to Go-based architecture

### agentlog Repository

- **Location**: `/Users/aki/workspace/agentlog`
- **Implementation**: Pure Go with production-grade architecture
- **Purpose**: Browse Codex CLI session logs
- **Architecture**:
  - `cmd/agentlog/` - CLI entry point (Cobra)
  - `internal/model/` - Data structures (86 lines)
  - `internal/parser/` - JSONL parsing (400+ lines)
  - `internal/store/` - Session discovery (250+ lines)
  - `internal/format/` - Output formatting (300+ lines)
  - `internal/view/` - Transcript rendering (400+ lines)

## Key Insight

**Most components can be shared between Claude Code and Codex log browsing:**

- ✅ **Shareable**: store, format, view (session discovery, output formatting, rendering)
- ❌ **Tool-specific**: parser, model (log formats differ between Claude Code and Codex)

## Proposed Approach

Extend agentlog to support both Codex CLI and Claude Code session logs by:

1. **Add Claude Code parser**
   - Create new parser in `internal/parser/` for Claude Code JSONL format
   - Reference existing Python implementation in cclog for parsing logic:
     - Parse session metadata from first entry
     - Handle summary format vs conversation format
     - Extract first user message and message count
     - Support path decoding for encoded project directories

2. **Add Claude Code models**
   - Create Claude Code-specific data structures in `internal/model/`
   - Map Claude Code events to common event model for rendering

3. **Extend CLI to support both tools**
   - Add `--log-type` or `--tool` flag (`claude` | `codex`)
   - Auto-detect based on sessions directory path:
     - `~/.claude/projects/` → Claude Code
     - `~/.codex/sessions/` → Codex CLI
   - Unified commands: `list`, `info`, `view`, `projects`

4. **Reuse existing agentlog components**
   - store: Session discovery works for both
   - format: Output formatting is tool-agnostic
   - view: Transcript rendering can handle both formats

## Claude Code Log Format

### Session Storage

- **Location**: `~/.claude/projects/{encoded-path}/*.jsonl`
- **Encoding**: Project paths are URL-encoded (/, ., \_ → -)
- **Files**: Multiple JSONL files per project (one per session)

### JSONL Format

Claude Code sessions use JSONL with several entry types:

1. **Summary entries** (optional, at file start):

   ```json
   { "summary": "Topic description", "uuid": "session-id" }
   ```

2. **Conversation entries** (main content):

   ```json
   {
     "timestamp": "2025-07-08T12:34:56.789Z",
     "role": "user" | "assistant" | "tool",
     "content": "message text or structured content"
   }
   ```

3. **Array format** (alternative):
   ```json
   {
     "timestamp": "2025-07-08T12:34:56.789Z",
     "role": "user",
     "content": ["array", "of", "content", "blocks"]
   }
   ```

### Key Parsing Requirements

- Handle mixed files (summary + conversation)
- Support both string and array content formats
- Efficient parsing: only read first 20 lines for user message
- Extract first/last timestamps for duration calculation
- Build summary index (UUID → summary mappings)
- Decode encoded project paths back to original paths

## Reference Implementation

### Python Parser Logic (from cclog_helper.py)

Key functions to port:

- `parse_session_minimal()` - Fast minimal parsing
- `get_session_list()` - Generate formatted session list
- `get_session_info()` - Extract session metadata
- `view_session()` - Display conversation content
- `build_summary_index()` - Create UUID → summary mappings
- `decode_project_path()` - Reverse engineer paths from encoded names

### Test Fixtures

Available in `/Users/aki/workspace/cclog/tests/fixtures/`:

- Original format (timestamp in first line)
- Summary format (summary entries before conversation)
- Array content format (content as array)
- Mixed files (summary + conversation)
- Edge cases (empty files, malformed JSON)

## Migration Plan

### Phase 1: Add Claude Code Support to agentlog

1. Create `internal/model/claude/` for Claude Code data structures
2. Create `internal/parser/claude/` for Claude Code JSONL parsing
3. Extend CLI to support `--tool claude` flag
4. Add auto-detection based on sessions directory
5. Write tests using existing cclog fixtures

### Phase 2: Unified CLI

1. Make tool detection automatic (default behavior)
2. Add `agentlog projects --tool claude` for Claude projects
3. Ensure all output formats work with both tools
4. Update documentation

### Phase 3: Deprecate cclog

1. Archive cclog repository or redirect to agentlog
2. Update README with migration instructions
3. Rename `agentlog` to universal name (optional: `devlog`, `ailog`, etc.)

## Goals

- **Single binary** that works with both Claude Code and Codex CLI
- **Shared codebase** for common functionality (80% shared, 20% tool-specific)
- **Production quality** following agentlog architecture patterns
- **Comprehensive tests** covering both log formats
- **Rich documentation** explaining both integrations

## Success Criteria

✅ Can browse Claude Code sessions: `agentlog list --tool claude`
✅ Can browse Codex sessions: `agentlog list --tool codex`
✅ Auto-detects tool based on directory
✅ All output formats work for both tools
✅ Path decoding works for Claude Code projects
✅ Summary indexing works for Claude Code sessions
✅ Test coverage >80% for both parsers
✅ Documentation covers both tools

## Next Steps

1. Review agentlog architecture
2. Design Claude Code parser interface matching Codex parser
3. Implement parser with test fixtures
4. Extend CLI with tool selection
5. Test end-to-end with real Claude Code sessions
6. Document Claude Code integration

## Files to Reference

- `/Users/aki/workspace/cclog/cclog_helper.py` - Python parsing logic
- `/Users/aki/workspace/cclog/tests/` - Test fixtures and test cases
- `/Users/aki/workspace/cclog/claude-projects-jsonl-specification.md` - Format spec
- `/Users/aki/workspace/agentlog/internal/parser/parser.go` - Codex parser reference
- `/Users/aki/workspace/agentlog/docs/architecture.md` - Architecture patterns

---

**Note**: This integration will make agentlog a universal CLI tool for browsing AI assistant conversation logs, supporting multiple backends (Codex, Claude Code, potentially others in the future).
