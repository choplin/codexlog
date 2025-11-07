package view

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"agentlog/internal/codex"
	"agentlog/internal/format"

	"github.com/mattn/go-isatty"
	"golang.org/x/term"
)

// Options defines the configurable parameters for rendering a view.
type Options struct {
	Path             string
	Format           string
	Wrap             int
	MaxEvents        int
	EntryTypeArg     string
	ResponseTypeArg  string
	EventMsgTypeArg  string
	PayloadRoleArg   string
	AllFilter        bool
	ForceColor       bool
	ForceNoColor     bool
	RawFile          bool
	Out              io.Writer
	OutFile          *os.File
}

// Run renders a session log according to the provided options.
func Run(opts Options) error {
	if opts.Out == nil {
		opts.Out = os.Stdout
	}

	if opts.RawFile {
		return copyFile(opts.Out, opts.Path)
	}

	filters, err := buildViewFilters(opts.AllFilter, opts.EntryTypeArg, opts.ResponseTypeArg, opts.EventMsgTypeArg, opts.PayloadRoleArg)
	if err != nil {
		return err
	}

	formatMode := strings.ToLower(opts.Format)
	if formatMode == "" {
		formatMode = "text"
	}

	if _, err := codex.ReadSessionMeta(opts.Path); err != nil {
		return err
	}

	processEvents := func(fn func(codex.CodexEvent) error) error {
		return codex.IterateEvents(opts.Path, func(event codex.CodexEvent) error {
			// Skip session_meta unless --all is specified
			if !opts.AllFilter && event.Kind == codex.EntryTypeSessionMeta {
				return nil
			}
			if !eventMatchesFilters(event, filters) {
				return nil
			}
			return fn(event)
		})
	}

	switch formatMode {
	case "text":
		useColor := resolveColorChoice(opts)
		if opts.MaxEvents == 0 {
			count := 0
			return processEvents(func(event codex.CodexEvent) error {
				if count > 0 {
					fmt.Fprintln(opts.Out)
				}
				printEvent(opts.Out, event, count+1, opts.Wrap, useColor)
				count++
				return nil
			})
		}
		ring := newEventRing(opts.MaxEvents)
		if err := processEvents(func(event codex.CodexEvent) error {
			ring.push(event)
			return nil
		}); err != nil {
			return err
		}
		for idx, event := range ring.slice() {
			if idx > 0 {
				fmt.Fprintln(opts.Out)
			}
			printEvent(opts.Out, event, idx+1, opts.Wrap, useColor)
		}
		return nil

	case "raw":
		if opts.MaxEvents == 0 {
			return processEvents(func(event codex.CodexEvent) error {
				_, err := fmt.Fprintln(opts.Out, event.Raw)
				return err
			})
		}
		ring := newEventRing(opts.MaxEvents)
		if err := processEvents(func(event codex.CodexEvent) error {
			ring.push(event)
			return nil
		}); err != nil {
			return err
		}
		for _, event := range ring.slice() {
			fmt.Fprintln(opts.Out, event.Raw)
		}
		return nil

	case "chat":
		colorEnabled := resolveColorChoice(opts)
		width := determineWidth(opts.OutFile, opts.Wrap)

		var events []codex.CodexEvent
		if opts.MaxEvents > 0 {
			ring := newEventRing(opts.MaxEvents)
			if err := processEvents(func(event codex.CodexEvent) error {
				ring.push(event)
				return nil
			}); err != nil {
				return err
			}
			events = ring.slice()
		} else {
			collected := make([]codex.CodexEvent, 0)
			if err := processEvents(func(event codex.CodexEvent) error {
				collected = append(collected, event)
				return nil
			}); err != nil {
				return err
			}
			events = collected
		}

		if len(events) == 0 {
			return nil
		}

		lines := renderChatTranscript(events, width, colorEnabled)
		if len(lines) == 0 {
			return nil
		}
		if opts.OutFile != nil && isatty.IsTerminal(opts.OutFile.Fd()) {
			return pipeThroughPager(lines, colorEnabled)
		}
		return writeLines(opts.Out, lines)

	default:
		return fmt.Errorf("unsupported format: %s", opts.Format)
	}
}

type viewFilters struct {
	entryTypes        map[codex.EntryType]struct{}
	responseItemTypes map[codex.ResponseItemType]struct{}
	eventMsgTypes     map[codex.EventMsgType]struct{}
	payloadRoles      map[codex.PayloadRole]struct{}
}

func buildViewFilters(allFilter bool, entryArg, responseTypeArg, eventMsgTypeArg, payloadRoleArg string) (viewFilters, error) {
	var filters viewFilters

	// If --all is specified, disable all filters
	if allFilter {
		return viewFilters{
			entryTypes:        nil,
			responseItemTypes: nil,
			eventMsgTypes:     nil,
			payloadRoles:      nil,
		}, nil
	}

	entryFilter, entryProvided, err := parseEntryTypeArg(entryArg)
	if err != nil {
		return filters, err
	}
	responseTypeFilter, responseTypeProvided, err := parseResponseTypeArg(responseTypeArg)
	if err != nil {
		return filters, err
	}
	eventMsgTypeFilter, eventMsgTypeProvided, err := parseEventMsgTypeArg(eventMsgTypeArg)
	if err != nil {
		return filters, err
	}
	payloadRoleFilter, roleProvided, err := parsePayloadRoleArg(payloadRoleArg)
	if err != nil {
		return filters, err
	}

	if entryProvided {
		filters.entryTypes = entryFilter
	} else {
		filters.entryTypes = map[codex.EntryType]struct{}{
			codex.EntryTypeResponseItem: {},
		}
	}

	if responseTypeProvided {
		filters.responseItemTypes = responseTypeFilter
	} else {
		filters.responseItemTypes = map[codex.ResponseItemType]struct{}{
			codex.ResponseItemTypeMessage: {},
		}
	}

	if eventMsgTypeProvided {
		filters.eventMsgTypes = eventMsgTypeFilter
	} else {
		// Default: no event_msg types (since EntryTypeEventMsg is excluded by default)
		filters.eventMsgTypes = nil
	}

	if roleProvided {
		filters.payloadRoles = payloadRoleFilter
	} else {
		filters.payloadRoles = map[codex.PayloadRole]struct{}{
			codex.PayloadRoleUser:      {},
			codex.PayloadRoleAssistant: {},
		}
	}

	return filters, nil
}

func parseEntryTypeArg(arg string) (map[codex.EntryType]struct{}, bool, error) {
	values := parseCSV(arg)
	if len(values) == 0 {
		return nil, false, nil
	}
	if len(values) == 1 && values[0] == "all" {
		return nil, true, nil
	}

	lookup := map[string]codex.EntryType{
		"session_meta":  codex.EntryTypeSessionMeta,
		"response_item": codex.EntryTypeResponseItem,
		"event_msg":     codex.EntryTypeEventMsg,
		"turn_context":  codex.EntryTypeTurnContext,
	}

	set := make(map[codex.EntryType]struct{}, len(values))
	for _, token := range values {
		entryType, ok := lookup[token]
		if !ok {
			return nil, true, fmt.Errorf("unknown entry type %q", token)
		}
		set[entryType] = struct{}{}
	}
	return set, true, nil
}

func parseResponseTypeArg(arg string) (map[codex.ResponseItemType]struct{}, bool, error) {
	values := parseCSV(arg)
	if len(values) == 0 {
		return nil, false, nil
	}
	if len(values) == 1 && values[0] == "all" {
		return nil, true, nil
	}

	lookup := map[string]codex.ResponseItemType{
		"message":                codex.ResponseItemTypeMessage,
		"reasoning":              codex.ResponseItemTypeReasoning,
		"function_call":          codex.ResponseItemTypeFunctionCall,
		"function_call_output":   codex.ResponseItemTypeFunctionCallOutput,
		"custom_tool_call":       codex.ResponseItemTypeCustomToolCall,
		"custom_tool_call_output": codex.ResponseItemTypeCustomToolCallOutput,
	}

	set := make(map[codex.ResponseItemType]struct{}, len(values))
	for _, token := range values {
		responseType, ok := lookup[token]
		if !ok {
			return nil, true, fmt.Errorf("unknown response type %q", token)
		}
		set[responseType] = struct{}{}
	}
	return set, true, nil
}

func parseEventMsgTypeArg(arg string) (map[codex.EventMsgType]struct{}, bool, error) {
	values := parseCSV(arg)
	if len(values) == 0 {
		return nil, false, nil
	}
	if len(values) == 1 && values[0] == "all" {
		return nil, true, nil
	}

	lookup := map[string]codex.EventMsgType{
		"token_count":     codex.EventMsgTypeTokenCount,
		"agent_reasoning": codex.EventMsgTypeAgentReasoning,
		"user_message":    codex.EventMsgTypeUserMessage,
		"agent_message":   codex.EventMsgTypeAgentMessage,
		"turn_aborted":    codex.EventMsgTypeTurnAborted,
	}

	set := make(map[codex.EventMsgType]struct{}, len(values))
	for _, token := range values {
		eventMsgType, ok := lookup[token]
		if !ok {
			return nil, true, fmt.Errorf("unknown event_msg type %q", token)
		}
		set[eventMsgType] = struct{}{}
	}
	return set, true, nil
}

func parsePayloadRoleArg(arg string) (map[codex.PayloadRole]struct{}, bool, error) {
	values := parseCSV(arg)
	if len(values) == 0 {
		return nil, false, nil
	}
	if len(values) == 1 && values[0] == "all" {
		return nil, true, nil
	}

	lookup := map[string]codex.PayloadRole{
		"user":      codex.PayloadRoleUser,
		"assistant": codex.PayloadRoleAssistant,
		"tool":      codex.PayloadRoleTool,
		"system":    codex.PayloadRoleSystem,
	}

	set := make(map[codex.PayloadRole]struct{}, len(values))
	for _, token := range values {
		role, ok := lookup[token]
		if !ok {
			return nil, true, fmt.Errorf("unknown payload role %q", token)
		}
		set[role] = struct{}{}
	}
	return set, true, nil
}

func parseCSV(arg string) []string {
	if strings.TrimSpace(arg) == "" {
		return nil
	}
	parts := strings.Split(arg, ",")
	output := make([]string, 0, len(parts))
	for _, part := range parts {
		token := strings.TrimSpace(strings.ToLower(part))
		if token != "" {
			output = append(output, token)
		}
	}
	return output
}

func eventMatchesFilters(event codex.CodexEvent, filters viewFilters) bool {
	if filters.entryTypes != nil {
		if _, ok := filters.entryTypes[event.Kind]; !ok {
			return false
		}
	}

	switch event.Kind {
	case codex.EntryTypeResponseItem:
		if filters.responseItemTypes != nil {
			if _, ok := filters.responseItemTypes[codex.ResponseItemType(event.PayloadType)]; !ok {
				return false
			}
		}
		if filters.payloadRoles != nil {
			if _, ok := filters.payloadRoles[event.Role]; !ok {
				return false
			}
		}
	case codex.EntryTypeEventMsg:
		if filters.eventMsgTypes != nil {
			if _, ok := filters.eventMsgTypes[codex.EventMsgType(event.PayloadType)]; !ok {
				return false
			}
		}
	}

	return true
}

type eventRing struct {
	data   []codex.CodexEvent
	start  int
	length int
}

func newEventRing(capacity int) *eventRing {
	if capacity <= 0 {
		return &eventRing{}
	}
	return &eventRing{data: make([]codex.CodexEvent, capacity)}
}

func (r *eventRing) push(event codex.CodexEvent) {
	if len(r.data) == 0 {
		return
	}
	idx := (r.start + r.length) % len(r.data)
	r.data[idx] = event
	if r.length < len(r.data) {
		r.length++
		return
	}
	r.start = (r.start + 1) % len(r.data)
}

func (r *eventRing) slice() []codex.CodexEvent {
	if r.length == 0 {
		return nil
	}
	result := make([]codex.CodexEvent, r.length)
	for i := 0; i < r.length; i++ {
		result[i] = r.data[(r.start+i)%len(r.data)]
	}
	return result
}

func determineWidth(out *os.File, wrap int) int {
	if wrap > 0 {
		return wrap
	}
	if out != nil {
		if w, _, err := term.GetSize(int(out.Fd())); err == nil && w > 0 {
			return w
		}
	}
	if colsStr := os.Getenv("COLUMNS"); colsStr != "" {
		if v, err := strconv.Atoi(colsStr); err == nil && v > 0 {
			return v
		}
	}
	return 80
}

func pipeThroughPager(lines []string, colorEnabled bool) error {
	text := strings.Join(lines, "\n")
	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}

	pagerCmd := os.Getenv("PAGER")
	var cmd *exec.Cmd
	if pagerCmd == "" {
		args := []string{"less"}
		if colorEnabled {
			args = append(args, "-R")
		}
		cmd = exec.Command(args[0], args[1:]...) // #nosec G204
	} else {
		cmd = exec.Command("sh", "-c", pagerCmd) // #nosec G204
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("create pager pipe: %w", err)
	}
	go func() {
		defer stdin.Close()
		io.WriteString(stdin, text) //nolint:errcheck
	}()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run pager: %w", err)
	}

	return nil
}

func writeLines(out io.Writer, lines []string) error {
	for _, line := range lines {
		if _, err := fmt.Fprintln(out, line); err != nil {
			return err
		}
	}
	return nil
}

func printEvent(out io.Writer, event codex.CodexEvent, index int, wrap int, useColor bool) {
	roleLabel := string(event.Role)
	if roleLabel == "" {
		roleLabel = string(event.Kind)
	}
	if roleLabel == "" {
		roleLabel = "event"
	}
	roleLabel = strings.ToLower(roleLabel)

	ts := "-"
	if !event.Timestamp.IsZero() {
		ts = event.Timestamp.Format(time.RFC3339)
	}
	headerPlain := fmt.Sprintf("[#%03d] %s | %s", index, roleLabel, ts)

	indexText := fmt.Sprintf("#%03d", index)
	roleText := roleLabel
	tsText := ts
	separator := "|"

	if useColor {
		indexText = colorize(true, ansiBoldWhite, indexText)
		roleText = colorize(true, roleColor(roleLabel), roleText)
		tsText = colorize(true, ansiTimestamp, tsText)
		separator = colorize(true, ansiSeparator, "|")
	}

	header := fmt.Sprintf("[%s] %s %s %s", indexText, roleText, separator, tsText)
	fmt.Fprintln(out, header)
	fmt.Fprintln(out, strings.Repeat("-", len(headerPlain)))

	lines := format.RenderEventLines(event, wrap)
	if len(lines) == 0 {
		prefix := "|"
		if useColor {
			prefix = colorize(true, ansiSeparator, "|")
		}
		fmt.Fprintf(out, "%s %s\n", prefix, "(no content)")
		return
}
	linePrefix := "| "
	emptyPrefix := "|"
	if useColor {
		separatorColor := colorize(true, ansiSeparator, "|")
		linePrefix = separatorColor + " "
		emptyPrefix = separatorColor
	}
	for _, line := range lines {
		if line == "" {
			fmt.Fprintln(out, emptyPrefix)
			continue
		}
		fmt.Fprintf(out, "%s%s\n", linePrefix, line)
	}
}

const (
	ansiReset     = "\x1b[0m"
	ansiBoldWhite = "\x1b[1;97m"
	ansiTimestamp = "\x1b[38;5;245m"
	ansiSeparator = "\x1b[38;5;240m"
	ansiAssistant = "\x1b[38;5;44m"
	ansiUser      = "\x1b[38;5;220m"
	ansiTool      = "\x1b[38;5;207m"
)

func colorize(enabled bool, code string, text string) string {
	if !enabled {
		return text
	}
	return code + text + ansiReset
}

func roleColor(role string) string {
	switch codex.PayloadRole(role) {
	case codex.PayloadRoleAssistant:
		return ansiAssistant
	case codex.PayloadRoleUser:
		return ansiUser
	case codex.PayloadRoleTool, codex.PayloadRoleSystem:
		return ansiTool
	default:
		return ansiSeparator
	}
}

func resolveColorChoice(opts Options) bool {
	if opts.ForceColor {
		return true
	}
	if opts.ForceNoColor {
		return false
	}
	return shouldUseColorAuto(opts.Out)
}

func shouldUseColorAuto(out io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	file, ok := out.(*os.File)
	if !ok {
		return false
	}
	fd := file.Fd()
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}

func copyFile(dst io.Writer, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(dst, f)
	return err
}
