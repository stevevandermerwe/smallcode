package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"smallcode/config"
	"smallcode/helpers"
	"smallcode/memory"
	"smallcode/security"
	"smallcode/skills"
	"smallcode/todos"
	"smallcode/tools"
	"smallcode/types"
)

// Model wraps types.Model for Bubble Tea

type Model struct {
	*types.Model
}

// Init implements tea.Model

func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle confirmation prompt
		if m.Model.PendingConfirm != nil {
			switch msg.Type {
			case tea.KeyRunes:
				if len(msg.Runes) > 0 {
					ch := string(msg.Runes[0])
					call := m.Model.PendingConfirm.Call
					if ch == "y" || ch == "Y" {
						return m, func() tea.Msg { return m.executeTool(call, true) }
					} else if ch == "n" || ch == "N" {
						return m, func() tea.Msg { return m.executeTool(call, false) }
					}
				}
			}
			return m, nil
		}

		// Normal key handling
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEscape:
			memory.WriteSessionSummary(m.Model.Messages, m.Model.StartTime)
			return m, tea.Quit
		case tea.KeyEnter:
			if m.Model.Waiting || m.Model.Input == "" {
				return m, nil
			}
			input := m.Model.Input
			m.Model.Input = ""

			if input == "/h" {
				m.Model.Output = append(m.Model.Output, helpText)
				return m, nil
			}

			if input == "/debug" {
				m.Model.Debug = !m.Model.Debug
				state := "off"
				if m.Model.Debug {
					state = "on"
				}
				m.Model.Output = append(m.Model.Output, fmt.Sprintf("%sdebug mode %s%s", dimStyle, state, resetStyle))
				return m, nil
			}

			m.Model.Output = append(m.Model.Output, fmt.Sprintf("%s❯%s %s", lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")).Render("❯"), resetStyle, input))
			return m, m.SendMessage(input)
		case tea.KeyBackspace:
			if len(m.Model.Input) > 0 {
				m.Model.Input = m.Model.Input[:len(m.Model.Input)-1]
			}
		case tea.KeySpace:
			if !m.Model.Waiting {
				m.Model.Input += " "
			}
		case tea.KeyRunes:
			if !m.Model.Waiting {
				m.Model.Input += string(msg.Runes)
			}
		}
	case tea.WindowSizeMsg:
		m.Model.Width = msg.Width
		m.Model.Height = msg.Height
	case types.APIResponse:
		if msg.Err != nil {
			m.Model.Output = append(m.Model.Output, fmt.Sprintf("%sError: %v%s", errorStyle, msg.Err, resetStyle))
			m.Model.Waiting = false
			return m, nil
		}

		// Accumulate token counts
		m.Model.TotalInputTokens += msg.InputTokens
		m.Model.TotalOutputTokens += msg.OutputTokens

		if m.Model.Debug && (msg.InputTokens > 0 || msg.OutputTokens > 0) {
			m.Model.Output = append(m.Model.Output, fmt.Sprintf("%s[tokens] in=%d out=%d total_in=%d total_out=%d%s",
				dimStyle, msg.InputTokens, msg.OutputTokens, m.Model.TotalInputTokens, m.Model.TotalOutputTokens, resetStyle))
		}

		m.Model.AssistantBlocks = msg.Content
		m.Model.CollectedResults = []types.ContentBlock{}

		// Always append assistant message to history
		m.Model.Messages = append(m.Model.Messages, types.Message{Role: "assistant", Content: msg.Content})

		// Extract tool calls and display output
		var toolCalls []types.ToolCall
		for _, block := range msg.Content {
			if block.Type == "text" {
				m.Model.Output = append(m.Model.Output, fmt.Sprintf("%s%s%s", assistantStyle, block.Text, resetStyle))
			}
			if block.Type == "tool_use" {
				m.Model.Output = append(m.Model.Output, fmt.Sprintf("%s%s(%s)%s", toolStyle, block.Name, helpers.Truncate(fmt.Sprintf("%v", block.Input), 50), resetStyle))
				toolCalls = append(toolCalls, types.ToolCall{ID: block.ID, Name: block.Name, Args: block.Input})
				if m.Model.Debug {
					m.Model.Output = append(m.Model.Output, fmt.Sprintf("%s[debug] tool call: %s args=%v%s", dimStyle, block.Name, block.Input, resetStyle))
				}
			}
		}

		m.Model.ToolQueue = toolCalls

		// Start processing tools
		if len(toolCalls) > 0 {
			return m, m.processNextTool()
		}

		m.Model.Waiting = false

	case types.ToolConfirmMsg:
		m.Model.PendingConfirm = &msg
		// View will render the confirmation prompt

	case types.ToolExecResult:
		m.Model.CollectedResults = append(m.Model.CollectedResults, types.ContentBlock{
			Type:   "tool_result",
			ToolID: msg.ID,
			Result: msg.Result,
		})
		m.Model.PendingConfirm = nil

		preview := helpers.Truncate(strings.Split(msg.Result, "\n")[0], 60)
		m.Model.Output = append(m.Model.Output, fmt.Sprintf("%s⎿  %s%s", dimStyle, preview, resetStyle))
		if m.Model.Debug {
			m.Model.Output = append(m.Model.Output, fmt.Sprintf("%s[debug] tool result id=%s: %s%s", dimStyle, msg.ID, helpers.Truncate(msg.Result, 200), resetStyle))
		}

		if len(m.Model.ToolQueue) > 0 {
			return m, m.processNextTool()
		}

		// All tools processed — assistant message already in history; send tool results back
		m.Model.Messages = append(m.Model.Messages, types.Message{Role: "user", Content: m.Model.CollectedResults})
		return m, m.CallAPI(nil)

	case types.ToolBlockedMsg:
		m.Model.CollectedResults = append(m.Model.CollectedResults, types.ContentBlock{
			Type:   "tool_result",
			ToolID: msg.Call.ID,
			Result: fmt.Sprintf("error: %s", msg.Reason),
		})
		m.Model.Output = append(m.Model.Output, fmt.Sprintf("%s⎿  blocked: %s%s", errorStyle, msg.Reason, resetStyle))

		if len(m.Model.ToolQueue) > 0 {
			return m, m.processNextTool()
		}

		m.Model.Messages = append(m.Model.Messages, types.Message{Role: "user", Content: m.Model.CollectedResults})
		return m, m.CallAPI(nil)

	case types.ToolOutput:
		m.Model.ToolResults = append(m.Model.ToolResults, types.ToolResult{Name: msg.Name, Result: msg.Result, Error: msg.Err})
		result := msg.Result
		if msg.Err != "" {
			result = msg.Err
		}
		preview := helpers.Truncate(strings.Split(result, "\n")[0], 60)
		m.Model.Output = append(m.Model.Output, fmt.Sprintf("%s⎿  %s%s", dimStyle, preview, resetStyle))
	}
	return m, nil
}

// View renders the UI

func (m Model) View() string {
	var lines []string

	tokenInfo := ""
	if m.Model.TotalInputTokens > 0 || m.Model.TotalOutputTokens > 0 {
		tokenInfo = fmt.Sprintf(" | %stokens ↑%d ↓%d%s", dimStyle, m.Model.TotalInputTokens, m.Model.TotalOutputTokens, resetStyle)
	}
	debugInfo := ""
	if m.Model.Debug {
		debugInfo = fmt.Sprintf(" | %s[debug]%s", lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render("[debug]"), resetStyle)
	}
	header := fmt.Sprintf("%s smallcode %s| %s (%s)%s | %s%s%s",
		lipgloss.NewStyle().Bold(true).Render("smallcode"),
		resetStyle,
		dimStyle,
		m.Model.ModelName,
		resetStyle,
		m.Model.Provider,
		tokenInfo,
		debugInfo,
	)
	lines = append(lines, header, "")

	visibleLines := m.Model.Height - 10
	if visibleLines < 0 {
		visibleLines = 20
	}
	start := 0
	if len(m.Model.Output) > visibleLines {
		start = len(m.Model.Output) - visibleLines
	}
	displayLines := m.Model.Output[start:]
	lines = append(lines, displayLines...)

	if len(displayLines) > 0 {
		lines = append(lines, "")
	}

	// Render confirmation prompt if pending
	if m.Model.PendingConfirm != nil {
		confirmLine := fmt.Sprintf("%s⚠ %s: %s%s", errorStyle, m.Model.PendingConfirm.Call.Name, m.Model.PendingConfirm.Reason, resetStyle)
		lines = append(lines, confirmLine)
		promptLine := fmt.Sprintf("  [y] approve  [n] deny")
		lines = append(lines, promptLine)
	} else {
		prompt := fmt.Sprintf("%s%s❯%s ", lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")).Render("❯"), resetStyle, dimStyle)
		if m.Model.Waiting {
			prompt = fmt.Sprintf("%s%s⏳%s ", lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")).Render("⏳"), resetStyle, dimStyle)
		}
		inputLine := prompt + m.Model.Input
		if m.Model.Input == "" {
			inputLine += dimStyle.Render("...") + resetStyle.Render("")
		}
		lines = append(lines, inputLine)
	}

	return lipgloss.NewStyle().Margin(1).Render(strings.Join(lines, "\n"))
}

// SendMessage adds user message and calls API

func (m *Model) SendMessage(input string) tea.Cmd {
	skillRE := regexp.MustCompile(`@(\S+)`)
	matches := skillRE.FindStringSubmatch(input)

	var skillContent string
	if matches != nil {
		skillName := matches[1]
		if content, err := skills.Load(skillName); err == nil {
			skillContent = fmt.Sprintf("[Skill: %s]\n\n%s\n\n", skillName, content)
		} else {
			m.Model.Output = append(m.Model.Output, fmt.Sprintf("%sSkill not found: %s%s", errorStyle, skillName, resetStyle))
			m.Model.Waiting = false
			return nil
		}
		input = skillRE.ReplaceAllString(input, "")
		input = strings.TrimSpace(input)
	}

	fullMessage := skillContent + input
	m.Model.Messages = append(m.Model.Messages, types.Message{Role: "user", Content: fullMessage})
	m.Model.Waiting = true
	return m.CallAPI(nil)
}

// CallAPI makes the API request

func (m *Model) CallAPI(toolResults []types.ContentBlock) tea.Cmd {
	return func() tea.Msg {
		messages := make([]types.Message, len(m.Model.Messages))
		copy(messages, m.Model.Messages)
		if toolResults != nil {
			messages = append(messages, types.Message{Role: "user", Content: toolResults})
		}

		toolsSlice := []map[string]interface{}{
			{"name": "read", "description": "Read file with line numbers", "input_schema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{"path": map[string]interface{}{"type": "string"}, "offset": map[string]interface{}{"type": "integer"}, "limit": map[string]interface{}{"type": "integer"}}, "required": []string{"path"}}},
			{"name": "write", "description": "Write content to file", "input_schema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{"path": map[string]interface{}{"type": "string"}, "content": map[string]interface{}{"type": "string"}}, "required": []string{"path", "content"}}},
			{"name": "edit", "description": "Replace old with new in file", "input_schema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{"path": map[string]interface{}{"type": "string"}, "old": map[string]interface{}{"type": "string"}, "new": map[string]interface{}{"type": "string"}, "all": map[string]interface{}{"type": "boolean"}}, "required": []string{"path", "old", "new"}}},
			{"name": "glob", "description": "Find files by pattern", "input_schema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{"pat": map[string]interface{}{"type": "string"}, "path": map[string]interface{}{"type": "string"}}, "required": []string{"pat"}}},
			{"name": "grep", "description": "Search files for regex", "input_schema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{"pat": map[string]interface{}{"type": "string"}, "path": map[string]interface{}{"type": "string"}}, "required": []string{"pat"}}},
			{"name": "bash", "description": "Run shell command", "input_schema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{"cmd": map[string]interface{}{"type": "string"}}, "required": []string{"cmd"}}},
			{"name": "remember", "description": "Persist a fact to project memory", "input_schema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{"action": map[string]interface{}{"type": "string", "enum": []string{"add", "update", "forget"}}, "key": map[string]interface{}{"type": "string"}, "value": map[string]interface{}{"type": "string"}, "tags": map[string]interface{}{"type": "string"}}, "required": []string{"action", "key"}}},
			{"name": "todo", "description": "Manage project todos with priorities and dependencies", "input_schema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{"action": map[string]interface{}{"type": "string", "enum": []string{"add", "update", "close", "reopen", "remove", "list"}}, "id": map[string]interface{}{"type": "string"}, "title": map[string]interface{}{"type": "string"}, "priority": map[string]interface{}{"type": "integer"}, "blocked_by": map[string]interface{}{"type": "string"}, "sources": map[string]interface{}{"type": "string"}, "status": map[string]interface{}{"type": "string"}}, "required": []string{"action"}}},
		}

		cwd, _ := os.Getwd()
		payload := map[string]interface{}{
			"model":      config.MODEL,
			"max_tokens": config.MAX_TOKENS,
			"system":     BuildSystemPrompt(cwd),
			"messages":   messages,
			"tools":      toolsSlice,
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", config.API_URL, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("anthropic-version", "2023-06-01")
		if config.OPENROUTER_KEY != "" {
			req.Header.Set("Authorization", "Bearer "+config.OPENROUTER_KEY)
		} else {
			req.Header.Set("x-api-key", os.Getenv("ANTHROPIC_API_KEY"))
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return types.APIResponse{Err: err}
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		if resp.StatusCode != 200 {
			errMsg := fmt.Sprintf("API error %d", resp.StatusCode)
			if errObj, ok := result["error"].(map[string]interface{}); ok {
				if msg, ok := errObj["message"].(string); ok {
					errMsg = fmt.Sprintf("API error %d: %s", resp.StatusCode, msg)
				}
			}
			return types.APIResponse{Err: fmt.Errorf("%s", errMsg)}
		}

		// Extract token usage
		inputTokens, outputTokens := 0, 0
		if usage, ok := result["usage"].(map[string]interface{}); ok {
			if v, ok := usage["input_tokens"].(float64); ok {
				inputTokens = int(v)
			}
			if v, ok := usage["output_tokens"].(float64); ok {
				outputTokens = int(v)
			}
		}

		contentBlocks, _ := result["content"].([]interface{})
		var assistantBlocks []types.ContentBlock

		for _, b := range contentBlocks {
			block, ok := b.(map[string]interface{})
			if !ok {
				continue
			}
			blockType, _ := block["type"].(string)

			if blockType == "text" {
				text, _ := block["text"].(string)
				assistantBlocks = append(assistantBlocks, types.ContentBlock{Type: "text", Text: text})
			}
			if blockType == "tool_use" {
				id, _ := block["id"].(string)
				name, _ := block["name"].(string)
				args, _ := block["input"].(map[string]interface{})
				if args == nil {
					args = map[string]interface{}{}
				}
				assistantBlocks = append(assistantBlocks, types.ContentBlock{Type: "tool_use", ID: id, Name: name, Input: args})
			}
		}

		return types.APIResponse{Content: assistantBlocks, InputTokens: inputTokens, OutputTokens: outputTokens}
	}
}

// processNextTool processes the next tool in the queue
func (m *Model) processNextTool() tea.Cmd {
	return func() tea.Msg {
		if len(m.Model.ToolQueue) == 0 {
			return nil
		}

		call := m.Model.ToolQueue[0]
		m.Model.ToolQueue = m.Model.ToolQueue[1:]

		cwd, _ := os.Getwd()
		policy := security.Check(call.Name, call.Args, cwd)

		switch policy.Decision {
		case security.Allow:
			security.Log("ALLOW", call.Name, call.Args, "")
			return m.executeTool(call, true)
		case security.Block:
			security.Log("BLOCK", call.Name, call.Args, policy.Reason)
			return types.ToolBlockedMsg{Call: call, Reason: policy.Reason}
		case security.Confirm:
			security.Log("CONFIRM", call.Name, call.Args, "")
			return types.ToolConfirmMsg{Call: call, Reason: policy.Reason}
		}

		return nil
	}
}

// executeTool executes a tool and returns the result message
func (m *Model) executeTool(call types.ToolCall, approved bool) types.ToolExecResult {
	if !approved {
		security.Log("DENY", call.Name, call.Args, "user denied")
		return types.ToolExecResult{ID: call.ID, Result: "error: user denied this operation"}
	}

	var result string

	switch call.Name {
	case "read":
		result = tools.Read(call.Args)
	case "write":
		result = tools.Write(call.Args)
		security.Log("EXEC", call.Name, call.Args, "")
	case "edit":
		result = tools.Edit(call.Args)
		security.Log("EXEC", call.Name, call.Args, "")
	case "glob":
		result = tools.Glob(call.Args)
	case "grep":
		result = tools.Grep(call.Args)
	case "bash":
		result = tools.Bash(call.Args)
		security.Log("EXEC", call.Name, call.Args, "")
	case "remember":
		result = tools.Remember(call.Args)
	case "todo":
		result = todos.Execute(call.Args)
	default:
		result = "error: unknown tool"
	}

	return types.ToolExecResult{ID: call.ID, Result: result}
}

// Styles

var (
	resetStyle     = lipgloss.NewStyle()
	dimStyle       = lipgloss.NewStyle().Faint(true)
	errorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
	assistantStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	toolStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
)

const helpText = `smallcode - Commands & Tips

Commands:
  /h          Show this help
  /debug      Toggle debug mode (shows API details, token counts, tool args)
  /c          Clear conversation
  /q, exit    Quit

Skills:
  @skillname  Activate a skill (e.g. @example)

Tools:
  read        Read file (path, offset?, limit?)
  write       Write file (path, content)
  edit        Edit file (path, old, new, all?)
  glob        Find files (pat, path?)
  grep        Search regex (pat, path?)
  bash        Run shell (cmd)
  remember    Memory (action, key, value?, tags?)
  todo        Todos (action, id?, title?, priority?, blocked_by?, sources?, status?)`

// BuildSystemPrompt creates the system prompt with memory context

func BuildSystemPrompt(cwd string) string {
	base := fmt.Sprintf("Concise coding assistant. cwd: %s", cwd)
	ctx := LoadMemoryContext()
	if ctx == "" {
		return base
	}
	return base + "\n\n" + ctx + "\n## Memory & Task Management\n- Use `todo` tool to manage work items (action=add|update|close|reopen|remove|list)\n- Set priority 1-3 (1=urgent). Use blocked_by to track dependencies (comma-separated IDs).\n- Add file:path or text:snippet sources when a todo relates to specific code.\n- Use `remember` tool to persist key project facts (action=add|update|forget)\n- When closing a todo that produced a lasting project fact, also use `remember` to record it.\n- Keep memory values under 200 chars; prefer update over adding duplicate keys"
}

// LoadMemoryContext loads memory and todos for the system prompt

func LoadMemoryContext() string {
	var sb strings.Builder

	if data, err := os.ReadFile(".smallcode/memory.json"); err == nil {
		var mem types.MemFile
		if json.Unmarshal(data, &mem) == nil && len(mem.Entries) > 0 {
			sb.WriteString("## Project Memory\n")
			for _, e := range mem.Entries {
				sb.WriteString(fmt.Sprintf("- %s: %s\n", e.Key, e.Value))
			}
		}
	}

	if data, err := os.ReadFile(".smallcode/todos.json"); err == nil {
		var tf types.TodoFile
		if json.Unmarshal(data, &tf) == nil && len(tf.Todos) > 0 {
			todoPrompt := todos.FormatForPrompt(tf.Todos)
			if todoPrompt != "" {
				sb.WriteString("\n" + todoPrompt)
			}
		}
	}

	return sb.String()
}

// NewModel creates a new API model

func NewModel() *Model {
	config.Init()
	return &Model{
		Model: &types.Model{
			Messages:  []types.Message{},
			Output:    []string{},
			Provider:  config.Provider(),
			ModelName: config.MODEL,
			StartTime: time.Now(),
		},
	}
}
