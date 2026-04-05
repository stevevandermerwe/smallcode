package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
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
					} else if ch == "a" || ch == "A" {
						security.AllowAlways(call.Name)
						security.Log("ALWAYS", call.Name, call.Args, "granted permanent permission")
						return m, func() tea.Msg { return m.executeTool(call, true) }
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

			if m.Model.PromptingForApiKey {
				key := strings.TrimSpace(input)
				if key != "" {
					// Read existing .env and replace key
					data, _ := os.ReadFile(".env")
					lines := strings.Split(string(data), "\n")
					keyUpdated := false
					for i, line := range lines {
						if strings.HasPrefix(line, "OPENROUTER_API_KEY=") {
							lines[i] = "OPENROUTER_API_KEY=" + key
							keyUpdated = true
						}
					}
					if !keyUpdated {
						lines = append(lines, "OPENROUTER_API_KEY="+key)
					}
					os.WriteFile(".env", []byte(strings.Join(lines, "\n")), 0644)
					config.Init() // Re-initialize with the new key and model
					m.Model.Provider = config.Provider()
					m.Model.ModelName = config.MODEL
					m.Model.Output = append(m.Model.Output, dimStyle.Render("   ✔ OpenRouter API Key saved and configuration updated"))
				} else {
					m.Model.Output = append(m.Model.Output, dimStyle.Render("   ℹ Skipping API Key addition (none provided)"))
				}
				m.Model.PromptingForApiKey = false
				m.Model.Output = append(m.Model.Output, dimStyle.Render("   Initialization complete! Enjoy SmallCode."))
				return m, nil
			}

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
				m.Model.Output = append(m.Model.Output, dimStyle.Render(fmt.Sprintf("   debug mode %s", state)))
				return m, nil
			}

			if input == "/yolo" {
				m.Model.Yolo = !m.Model.Yolo
				config.YOLO = m.Model.Yolo
				state := "off"
				if m.Model.Yolo {
					state = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true).Render("ON")
				}
				m.Model.Output = append(m.Model.Output, dimStyle.Render(fmt.Sprintf("   YOLO mode %s", state)))
				return m, nil
			}

			if input == "/c" {
				m.Model.Messages = []types.Message{}
				m.Model.Output = append(m.Model.Output, dimStyle.Render("   Conversation cleared."))
				return m, nil
			}

			if input == "/s" {
				m.Model.Output = append(m.Model.Output, fmt.Sprintf("%s %s", userPrefix, input))
				m.Model.Output = append(m.Model.Output, dimStyle.Render("   Summarizing conversation..."))
				prompt := "Please summarize our conversation so far. Highlight key decisions and current state. If there are important facts or pending tasks, use the `remember` or `todo` tools to persist them before concluding."
				m.Model.Messages = append(m.Model.Messages, types.Message{Role: "user", Content: prompt})
				m.Model.Summarizing = true
				m.Model.Waiting = true
				return m, m.CallAPI(nil)
			}

			if strings.HasPrefix(input, "/add ") {
				path := strings.TrimSpace(input[5:])
				data, err := os.ReadFile(path)
				if err != nil {
					m.Model.Output = append(m.Model.Output, fmt.Sprintf("   %s Error reading file: %v", errorStyle.Render("✘"), err))
					return m, nil
				}
				content := fmt.Sprintf("File: %s\n\n```\n%s\n```", path, string(data))
				m.Model.Messages = append(m.Model.Messages, types.Message{Role: "user", Content: content})
				m.Model.Output = append(m.Model.Output, dimStyle.Render(fmt.Sprintf("   Added %s to context.", path)))
				return m, nil
			}

			if input == "/init" {
				m.Model.Output = append(m.Model.Output, fmt.Sprintf("%s %s", userPrefix, input))
				msgs := initProject()
				for _, msg := range msgs {
					m.Model.Output = append(m.Model.Output, dimStyle.Render("   "+msg))
				}
				m.Model.PromptingForApiKey = true
				m.Model.Output = append(m.Model.Output, lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Bold(true).Render("   Enter your OpenRouter API Key:"))
				return m, nil
			}

			m.Model.Output = append(m.Model.Output, fmt.Sprintf("%s %s", userPrefix, input))
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

		if m.Model.Summarizing {
			// Clear all previous history and start fresh with just the summary
			m.Model.Messages = []types.Message{{Role: "assistant", Content: msg.Content}}
			m.Model.Summarizing = false
		} else {
			// Always append assistant message to history
			m.Model.Messages = append(m.Model.Messages, types.Message{Role: "assistant", Content: msg.Content})
		}

		// Extract tool calls and display output
		var toolCalls []types.ToolCall
		for _, block := range msg.Content {
			if block.Type == "text" {
				m.Model.Output = append(m.Model.Output, fmt.Sprintf("%s %s", assistantPrefix, block.Text))
			}
			if block.Type == "tool_use" {
				m.Model.Output = append(m.Model.Output, fmt.Sprintf("   %s %s(%s)", toolStyle.Render("⚒"), block.Name, dimStyle.Render(helpers.Truncate(fmt.Sprintf("%v", block.Input), 50))))
				toolCalls = append(toolCalls, types.ToolCall{ID: block.ID, Name: block.Name, Args: block.Input})
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

		preview := dimStyle.Render(helpers.Truncate(strings.Split(msg.Result, "\n")[0], 60))
		m.Model.Output = append(m.Model.Output, fmt.Sprintf("     %s %s", dimStyle.Render("↳"), preview))

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
		m.Model.Output = append(m.Model.Output, fmt.Sprintf("     %s %s", errorStyle.Render("✕"), dimStyle.Render(msg.Reason)))

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

	// 1. Header
	tokenInfo := ""
	if m.Model.TotalInputTokens > 0 || m.Model.TotalOutputTokens > 0 {
		color := "240"
		if config.MAX_TOKENS > 0 && m.Model.TotalInputTokens > int(float64(config.MAX_TOKENS)*0.8) {
			color = "3" // yellow warning
		}
		tokenInfo = fmt.Sprintf(" | %stokens ↑%d ↓%d%s", lipgloss.NewStyle().Foreground(lipgloss.Color(color)), m.Model.TotalInputTokens, m.Model.TotalOutputTokens, resetStyle)
	}

	debugInfo := ""
	if m.Model.Debug {
		debugInfo = fmt.Sprintf(" | %s[debug]%s", lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render("[debug]"), resetStyle)
	}

	yoloInfo := ""
	if m.Model.Yolo {
		yoloInfo = fmt.Sprintf(" | %s[YOLO]%s", lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true).Render("[YOLO]"), resetStyle)
	}

	headerText := fmt.Sprintf("%s %s (%s)%s%s%s | %s",
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")).Render("smallcode"),
		dimStyle.Render(m.Model.ModelName),
		dimStyle.Render(m.Model.Provider),
		tokenInfo,
		debugInfo,
		yoloInfo,
		dimStyle.Render("/h for help"),
	)

	header := headerStyle.Render(headerText)
	lines = append(lines, header, "")

	// 2. Body (Output)
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

	// 3. Footer / Input Area
	if m.Model.PendingConfirm != nil {
		confirmLine := fmt.Sprintf("%s⚠ %s: %s%s", errorStyle, m.Model.PendingConfirm.Call.Name, m.Model.PendingConfirm.Reason, resetStyle)
		lines = append(lines, confirmLine)
		promptLine := fmt.Sprintf("  %s[y] approve  [n] deny  [a] always allow%s", dimStyle.Render(""), resetStyle.Render(""))
		lines = append(lines, promptLine)
	} else {
		promptSym := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")).Render("❯")
		if m.Model.Waiting {
			promptSym = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("13")).Render("⏳")
		}

		inputContent := m.Model.Input
		if inputContent == "" && !m.Model.Waiting {
			inputContent = dimStyle.Render("Type a message...")
		}

		inputLine := fmt.Sprintf("%s %s", promptSym, inputContent)
		lines = append(lines, inputLine)
	}

	return lipgloss.NewStyle().
		Padding(1, 2).
		Width(m.Model.Width).
		Height(m.Model.Height).
		Render(strings.Join(lines, "\n"))
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
			m.Model.Output = append(m.Model.Output, dimStyle.Render(fmt.Sprintf("   Applied skill: %s", skillName)))
		} else {
			m.Model.Output = append(m.Model.Output, fmt.Sprintf("   %s Skill not found: %s", errorStyle.Render("✘"), skillName))
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

		// Auto-Pruning: Slice history if it's over 90% of MAX_TOKENS
		if config.MAX_TOKENS > 0 && m.Model.TotalInputTokens > int(float64(config.MAX_TOKENS)*0.9) {
			pruneCount := len(messages) / 5 // Prune oldest 20%
			if pruneCount > 0 {
				messages = messages[pruneCount:]
				// We update the local copy used for the request, but also update the model's history to reflect this permanent pruning
				m.Model.Messages = m.Model.Messages[pruneCount:]
				m.Model.Output = append(m.Model.Output, fmt.Sprintf("%sAuto-pruned oldest conversation history to save tokens.%s", dimStyle, resetStyle))
			}
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
		req.Header.Set("Authorization", "Bearer "+config.OPENROUTER_KEY)
		req.Header.Set("HTTP-Referer", "SmallCode")
		req.Header.Set("X-Title", "SmallCode")

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

		if m.Model.Yolo {
			security.Log("YOLO", call.Name, call.Args, "bypassing security")
			return m.executeTool(call, true)
		}

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
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("235")).
			Padding(0, 1).
			Bold(true).
			MarginBottom(1)

	resetStyle     = lipgloss.NewStyle()
	dimStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	errorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	assistantStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	userPrefix     = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true).Render("❯")
	assistantPrefix = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true).Render("◇")
	toolStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
)

const helpText = `smallcode - Commands & Tips

Commands:
  /h          Show this help
  /init       Initialize project (.env, .smallcode, git init)
  /debug      Toggle debug mode (shows API details, token counts, tool args)
  /yolo       Toggle YOLO mode (bypasses ALL security protections)
  /add <path> Add a file's content to the context
  /s          Summarize conversation and prompt to save facts
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
	base += "\n\nSearch tools (glob, grep) automatically exclude .git, node_modules, and other build artifacts. Users may explicitly add files to context using `/add <path>`."
	if config.YOLO {
		base += "\n\n[WARNING: YOLO MODE ACTIVE] Security protections are disabled. You have full system access. Use extreme caution."
	}
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
	security.LoadPermissions()
	m := &Model{
		Model: &types.Model{
			Messages:  []types.Message{},
			Output:    []string{},
			Provider:  config.Provider(),
			ModelName: config.MODEL,
			StartTime: time.Now(),
		},
	}

	// Check if project is initialized
	if _, err := os.Stat(".smallcode"); os.IsNotExist(err) {
		m.Model.Output = append(m.Model.Output, fmt.Sprintf("%sProject not initialized. Type %s/init%s to get started.%s", errorStyle, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")).Render(""), "", resetStyle))
	}

	return m
}

func initProject() []string {
	var msgs []string

	// 1. .env (template)
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		content := "OPENROUTER_API_KEY=\nMODEL=minimax/minimax-m2.5\nMAX_TOKENS=16384\n"
		if err := os.WriteFile(".env", []byte(content), 0644); err == nil {
			msgs = append(msgs, "✔ Created .env template")
		}
	} else {
		msgs = append(msgs, "ℹ .env already exists")
	}

	// 2. .smallcode directory and basics
	os.MkdirAll(".smallcode/skills", 0755)
	
	files := map[string]string{
		".gitignore":             ".env\ndist/\n",
		".smallcode/ignore":      ".git\nnode_modules\ndist\n.smallcode\n",
		".smallcode/memory.json": "{\n  \"version\": 1,\n  \"entries\": []\n}",
		".smallcode/todos.json":  "{\n  \"version\": 1,\n  \"todos\": []\n}",
		".smallcode/skills/example.md": "# Example Skill\n\nYou are an example skill that demonstrates the skills system.\n\n## Purpose\nSkills allow you to inject specialized instructions into your conversation on-demand.\n",
		".smallcode/skills/context.md": "# Context Audit Skill\n\nYou are an expert context manager. Your task is to audit the current project context (Memory and Todos) to ensure they are lean, relevant, and accurate.\n\n## Objectives\n1. **Audit Memory:** Use the `remember` tool with `action=forget` to remove stale or redundant facts. Use `action=update` to consolidate related facts into single, concise entries.\n2. **Audit Todos:** Use the `todo` tool with `action=list` first to see all tasks. Then:\n    - `action=close` tasks that are finished but still open.\n    - `action=remove` duplicate tasks.\n    - `action=update` tasks to add better priority or clear up blocking dependencies.\n3. **Consolidate:** If a finished task resulted in a lasting project fact, ensure that fact is in `memory` before closing the `todo`.\n\n## Goal\nKeep the system prompt under control so that it doesn't waste tokens. Be aggressive but careful not to lose critical project information.\n",
	}

	for path, content := range files {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.WriteFile(path, []byte(content), 0644); err == nil {
				msgs = append(msgs, fmt.Sprintf("✔ Created %s", path))
			}
		}
	}

	// 3. git init
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		cmd := exec.Command("git", "init")
		if err := cmd.Run(); err == nil {
			msgs = append(msgs, "✔ Initialized git repository")
		} else {
			msgs = append(msgs, fmt.Sprintf("✘ Failed to initialize git: %v", err))
		}
	} else {
		msgs = append(msgs, "ℹ Git repository already exists")
	}

	return msgs
}
