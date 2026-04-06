package api

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"smallcode/config"
	"smallcode/helpers"
	"smallcode/memory"
	"smallcode/repomapper"
	"smallcode/security"
	"smallcode/skills"
	"smallcode/types"
)

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
			if m.Model.ShowSkills {
				m.Model.ShowSkills = false
				return m, nil
			}
			if m.Model.ShowFiles {
				m.Model.ShowFiles = false
				return m, nil
			}
			memory.WriteSessionSummary(m.Model.Messages, m.Model.StartTime)
			return m, tea.Quit
		case tea.KeyUp:
			if m.Model.ShowSkills {
				if m.Model.SelectedSkillIdx > 0 {
					m.Model.SelectedSkillIdx--
				} else {
					m.Model.SelectedSkillIdx = len(m.Model.FilteredSkills) - 1
				}
				return m, nil
			}
			if m.Model.ShowFiles {
				if m.Model.SelectedFileIdx > 0 {
					m.Model.SelectedFileIdx--
				} else {
					m.Model.SelectedFileIdx = len(m.Model.FilteredFiles) - 1
				}
				return m, nil
			}
		case tea.KeyDown:
			if m.Model.ShowSkills {
				if m.Model.SelectedSkillIdx < len(m.Model.FilteredSkills)-1 {
					m.Model.SelectedSkillIdx++
				} else {
					m.Model.SelectedSkillIdx = 0
				}
				return m, nil
			}
			if m.Model.ShowFiles {
				if m.Model.SelectedFileIdx < len(m.Model.FilteredFiles)-1 {
					m.Model.SelectedFileIdx++
				} else {
					m.Model.SelectedFileIdx = 0
				}
				return m, nil
			}
		case tea.KeyTab, tea.KeyEnter:
			if m.Model.ShowSkills && len(m.Model.FilteredSkills) > 0 {
				selected := m.Model.FilteredSkills[m.Model.SelectedSkillIdx]
				// Replace the @part with the full skill name
				lastWordIdx := -1
				for i := len(m.Model.Input) - 1; i >= 0; i-- {
					if m.Model.Input[i] == '@' {
						lastWordIdx = i
						break
					}
				}
				if lastWordIdx != -1 {
					m.Model.Input = m.Model.Input[:lastWordIdx] + "@" + selected + " "
				}
				m.Model.ShowSkills = false
				return m, nil
			}
			if m.Model.ShowFiles && len(m.Model.FilteredFiles) > 0 {
				selected := m.Model.FilteredFiles[m.Model.SelectedFileIdx]
				m.Model.Input = "/add " + selected
				m.Model.ShowFiles = false
				return m, nil
			}
			if msg.Type == tea.KeyTab {
				return m, nil
			}
			if m.Model.Waiting || m.Model.Input == "" {
				return m, nil
			}
			input := m.Model.Input
			m.Model.Input = ""

			if m.Model.PromptingForResume {
				ch := strings.ToLower(strings.TrimSpace(input))
				if ch == "y" || ch == "yes" {
					if data, err := os.ReadFile(".smallcode/session.json"); err == nil {
						var state map[string]interface{}
						if json.Unmarshal(data, &state) == nil {
							// Restore messages
							if msgs, ok := state["messages"].([]interface{}); ok {
								jsonData, _ := json.Marshal(msgs)
								var restored []types.Message
								json.Unmarshal(jsonData, &restored)
								m.Model.Messages = restored
							}
							// Restore tokens
							if val, ok := state["input_tokens"].(float64); ok {
								m.Model.TotalInputTokens = int(val)
							}
							if val, ok := state["out_tokens"].(float64); ok {
								m.Model.TotalOutputTokens = int(val)
							}
							// Restore start time
							if val, ok := state["start_time"].(string); ok {
								if t, err := time.Parse(time.RFC3339, val); err == nil {
									m.Model.StartTime = t
								}
							}
							m.Model.Output = append(m.Model.Output, dimStyle.Render("   ✔ Session restored."))
						}
					}
				} else {
					os.Remove(".smallcode/session.json")
					m.Model.Output = append(m.Model.Output, dimStyle.Render("   ℹ Starting fresh session."))
				}
				m.Model.PromptingForResume = false
				return m, nil
			}

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

			if input == "/h" || input == "/help" {
				m.Model.Output = append(m.Model.Output, helpText)
				return m, nil
			}

			if input == "/debug" || input == "/d" {
				m.Model.Debug = !m.Model.Debug
				state := "off"
				if m.Model.Debug {
					state = "on"
				}
				m.Model.Output = append(m.Model.Output, dimStyle.Render(fmt.Sprintf("   debug mode %s", state)))
				return m, nil
			}

			if input == "/trace" || input == "/t" {
				m.Model.Trace = !m.Model.Trace
				state := "off"
				if m.Model.Trace {
					state = "on"
				}
				m.Model.Output = append(m.Model.Output, dimStyle.Render(fmt.Sprintf("   trace mode %s (logging to .smallcode/trace.log)", state)))
				return m, nil
			}

			if input == "/yolo" || input == "/y" {
				m.Model.Yolo = !m.Model.Yolo
				config.YOLO = m.Model.Yolo
				state := "off"
				if m.Model.Yolo {
					state = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true).Render("ON")
				}
				m.Model.Output = append(m.Model.Output, dimStyle.Render(fmt.Sprintf("   YOLO mode %s", state)))
				return m, nil
			}

			if input == "/clear" || input == "/c" {
				m.Model.Messages = []types.Message{}
				m.Model.Output = append(m.Model.Output, dimStyle.Render("   Conversation cleared."))
				return m, nil
			}

			if input == "/map" || input == "/m" {
				m.Model.Output = append(m.Model.Output, fmt.Sprintf("%s %s", userPrefix, input))
				m.Model.Output = append(m.Model.Output, dimStyle.Render("   Generating repository map..."))
				
				rm := repomapper.NewRepoMapper()
				repoMap, err := rm.GenerateMap(".")
				if err != nil {
					m.Model.Output = append(m.Model.Output, fmt.Sprintf("   %s Error generating map: %v", errorStyle.Render("✘"), err))
					return m, nil
				}
				
				content := fmt.Sprintf("Repository Skeleton Map:\n\n```\n%s\n```", repoMap)
				m.Model.Messages = append(m.Model.Messages, types.Message{Role: "user", Content: content})
				m.Model.Output = append(m.Model.Output, dimStyle.Render("   Added repository map to context."))
				return m, nil
			}

			if input == "/memory" || input == "/mem" {
				m.Model.Output = append(m.Model.Output, fmt.Sprintf("%s %s", userPrefix, input))
				m.Model.Output = append(m.Model.Output, dimStyle.Render("   Loading project memory and tasks..."))
				
				ctx := LoadMemoryContext()
				if ctx == "" {
					m.Model.Output = append(m.Model.Output, dimStyle.Render("   Memory and tasks are empty."))
					return m, nil
				}
				
				m.Model.Messages = append(m.Model.Messages, types.Message{Role: "user", Content: ctx})
				m.Model.Output = append(m.Model.Output, dimStyle.Render("   Added memory and tasks to context."))
				return m, nil
			}

			if input == "/summarize" || input == "/s" {
				m.Model.Output = append(m.Model.Output, fmt.Sprintf("%s %s", userPrefix, input))
				m.Model.Output = append(m.Model.Output, dimStyle.Render("   Summarizing conversation..."))
				prompt := "Please summarize our conversation so far. Highlight key decisions and current state. If there are important facts or pending tasks, use the `remember` or `todo` tools to persist them before concluding."
				m.Model.Messages = append(m.Model.Messages, types.Message{Role: "user", Content: prompt})
				m.Model.Summarizing = true
				m.Model.Waiting = true
				return m, m.CallAPI(nil)
			}

			if input == "/quit" || input == "/exit" || input == "/q" || input == "/e" {
				memory.WriteSessionSummary(m.Model.Messages, m.Model.StartTime)
				return m, tea.Quit
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
			m.logReasoning("user_input", map[string]string{"input": input})
			m.saveSession()
			return m, m.SendMessage(input)
		case tea.KeyBackspace:
			if len(m.Model.Input) > 0 {
				m.Model.Input = m.Model.Input[:len(m.Model.Input)-1]
			}
			// Re-filter or hide
			if m.Model.ShowSkills {
				lastWordIdx := -1
				for i := len(m.Model.Input) - 1; i >= 0; i-- {
					if m.Model.Input[i] == '@' {
						lastWordIdx = i
						break
					}
				}
				if lastWordIdx == -1 {
					m.Model.ShowSkills = false
				} else {
					filter := m.Model.Input[lastWordIdx+1:]
					m.Model.FilteredSkills = []string{}
					for _, s := range skills.List() {
						if strings.HasPrefix(s, filter) {
							m.Model.FilteredSkills = append(m.Model.FilteredSkills, s)
						}
					}
					if len(m.Model.FilteredSkills) == 0 {
						m.Model.ShowSkills = false
					}
					if m.Model.SelectedSkillIdx >= len(m.Model.FilteredSkills) {
						m.Model.SelectedSkillIdx = 0
					}
				}
			}
			if m.Model.ShowFiles {
				if !strings.HasPrefix(m.Model.Input, "/add ") {
					m.Model.ShowFiles = false
				} else {
					filter := m.Model.Input[5:]
					allFiles := m.listAllFiles()
					m.Model.FilteredFiles = []string{}
					for _, f := range allFiles {
						if strings.Contains(f, filter) {
							m.Model.FilteredFiles = append(m.Model.FilteredFiles, f)
						}
					}
					if len(m.Model.FilteredFiles) == 0 {
						m.Model.ShowFiles = false
					}
					if m.Model.SelectedFileIdx >= len(m.Model.FilteredFiles) {
						m.Model.SelectedFileIdx = 0
					}
				}
			}
		case tea.KeySpace:
			if !m.Model.Waiting {
				m.Model.Input += " "
				m.Model.ShowSkills = false
				if strings.HasPrefix(m.Model.Input, "/add ") && !m.Model.ShowFiles {
					m.Model.ShowFiles = true
					m.Model.FilteredFiles = m.listAllFiles()
					m.Model.SelectedFileIdx = 0
				}
			}
		case tea.KeyRunes:
			if !m.Model.Waiting {
				m.Model.Input += string(msg.Runes)
				input := m.Model.Input

				// Handle Skills
				lastWordIdx := -1
				for i := len(input) - 1; i >= 0; i-- {
					if input[i] == ' ' {
						break
					}
					if input[i] == '@' {
						lastWordIdx = i
						break
					}
				}

				if lastWordIdx != -1 {
					m.Model.ShowSkills = true
					filter := input[lastWordIdx+1:]
					m.Model.FilteredSkills = []string{}
					for _, s := range skills.List() {
						if strings.HasPrefix(s, filter) {
							m.Model.FilteredSkills = append(m.Model.FilteredSkills, s)
						}
					}
					if len(m.Model.FilteredSkills) == 0 {
						m.Model.ShowSkills = false
					}
					if m.Model.SelectedSkillIdx >= len(m.Model.FilteredSkills) {
						m.Model.SelectedSkillIdx = 0
					}
				} else {
					m.Model.ShowSkills = false
				}

				// Handle Files
				if strings.HasPrefix(input, "/add ") {
					m.Model.ShowFiles = true
					filter := input[5:]
					allFiles := m.listAllFiles()
					m.Model.FilteredFiles = []string{}
					for _, f := range allFiles {
						if strings.Contains(f, filter) {
							m.Model.FilteredFiles = append(m.Model.FilteredFiles, f)
						}
					}
					if len(m.Model.FilteredFiles) == 0 {
						m.Model.ShowFiles = false
					}
					if m.Model.SelectedFileIdx >= len(m.Model.FilteredFiles) {
						m.Model.SelectedFileIdx = 0
					}
				} else {
					m.Model.ShowFiles = false
				}
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
		m.logReasoning("api_response", map[string]interface{}{"blocks": msg.Content})
		m.saveSession()

		// Start processing tools
		if len(toolCalls) > 0 {
			m.Model.SequentialToolCount++
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
		m.saveSession()
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
		m.saveSession()
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
		if m.Model.ShowSkills && len(m.Model.FilteredSkills) > 0 {
			for i, skill := range m.Model.FilteredSkills {
				prefix := "  "
				style := dimStyle
				if i == m.Model.SelectedSkillIdx {
					prefix = "> "
					style = skillStyle
				}
				lines = append(lines, fmt.Sprintf("    %s%s", prefix, style.Render(skill)))
			}
		}

		if m.Model.ShowFiles && len(m.Model.FilteredFiles) > 0 {
			// Show up to 10 files
			start := 0
			end := len(m.Model.FilteredFiles)
			if end > 10 {
				if m.Model.SelectedFileIdx > 5 {
					start = m.Model.SelectedFileIdx - 5
					end = start + 10
					if end > len(m.Model.FilteredFiles) {
						end = len(m.Model.FilteredFiles)
						start = end - 10
					}
				} else {
					end = 10
				}
			}

			for i := start; i < end; i++ {
				file := m.Model.FilteredFiles[i]
				prefix := "  "
				style := dimStyle
				if i == m.Model.SelectedFileIdx {
					prefix = "> "
					style = toolStyle
				}
				lines = append(lines, fmt.Sprintf("    %s%s", prefix, style.Render(file)))
			}
			if len(m.Model.FilteredFiles) > 10 {
				lines = append(lines, fmt.Sprintf("    %s (%d more...)", dimStyle.Render("  ..."), len(m.Model.FilteredFiles)-10))
			}
		}

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
