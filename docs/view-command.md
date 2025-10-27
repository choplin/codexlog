---
created: 2025-10-25
updated: 2025-10-25
---

# `codexlog view` Output Plan

## Current Issues

- Messages blur together; long sessions are hard to scan and orient within.
- Tool/event entries dump JSON inline without indentation, creating horizontal scroll.
- No option to limit the number of events displayed when exploring large histories.

## Goals

1. Add clear headers so each message exposes index, role, and timestamp at a glance.
2. Improve body readability with indentation and wrapping while preserving `--wrap` semantics.
3. Provide optional limits and filtering to help contributors focus on relevant sections.

## Proposed Changes

- Prefix each rendered event with an indexed header like `[#03][assistant][2025-10-27T12:34:56Z]`.
- Indent body text by two spaces and honour the `--wrap` width for prose content.
- Pretty-print JSON payloads (tool outputs, event messages) beneath the header with consistent indentation.
- Introduce a `--max` flag to show only the most recent _N_ events (default: show all).
- Keep the existing `--role` filter but apply it before numbering so indices reflect the displayed subset.

## Acceptance Checklist

- [ ] Headers include sequential index, role, and timestamp for every message.
- [ ] Wrapped bodies are indented and readable within the configured width.
- [ ] JSON payloads render with indentation (no inline minified blobs).
- [ ] `--max` flag limits output without breaking metadata or numbering.
- [ ] Unit tests cover header formatting, JSON pretty-printing, and `--max` behaviour.
