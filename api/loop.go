package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"smallcode/config"
	"smallcode/security"
	"smallcode/skills"
	"smallcode/tools"
	"smallcode/types"
)

// SendMessage adds user message and calls API

func (m *Model) SendMessage(input string) tea.Cmd {
	skillRE := regexp.MustCompile(`@(\S+)`)
	matches := skillRE.FindStringSubmatch(input)

	var skillContent string
	if matches != nil {
		skillName := matches[1]
		m.Model.ActiveSkill = skillName
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
	m.Model.SequentialToolCount = 0
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
				m.logReasoning("context_prune", map[string]int{"pruned_count": pruneCount})
				m.Model.Output = append(m.Model.Output, fmt.Sprintf("%sAuto-pruned oldest conversation history to save tokens.%s", dimStyle, resetStyle))
			}
		}

		var toolsSlice []map[string]interface{}
		for _, t := range tools.Registry {
			include := false
			if m.Model.ActiveSkill == "" {
				include = true // Default: send all if no skill active (or maybe just core?)
			} else {
				for _, s := range t.Skills {
					if s == "core" || s == m.Model.ActiveSkill {
						include = true
						break
					}
				}
			}

			if include {
				toolsSlice = append(toolsSlice, map[string]interface{}{
					"name":         t.Name,
					"description":  t.Description,
					"input_schema": t.Schema,
				})
			}
		}

		// Reset ActiveSkill after it's been used to assemble tools for this turn
		// actually, maybe we should keep it for the whole conversation turn (multi-tool)
		// but SendMessage resets it next time.

		cwd, _ := os.Getwd()
		payload := map[string]interface{}{
			"model":      config.MODEL,
			"max_tokens": config.MAX_TOKENS,
			"system":     BuildSystemPrompt(cwd),
			"messages":   messages,
			"tools":      toolsSlice,
		}

		body, _ := json.Marshal(payload)
		if m.Model.Trace {
			m.logTraffic("SENT", body)
		}
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

		if m.Model.Trace {
			resBody, _ := json.Marshal(result)
			m.logTraffic("RECEIVED", resBody)
		}

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

		if m.Model.SequentialToolCount > 5 {
			return types.ToolConfirmMsg{Call: call, Reason: "loop prevention: too many sequential tool calls. Please approve to continue."}
		}

		if m.Model.Yolo {
			security.Log("YOLO", call.Name, call.Args, "bypassing security")
			return m.executeTool(call, true)
		}

		cwd, _ := os.Getwd()
		policy := security.Check(call.Name, call.Args, cwd)

		switch policy.Decision {
		case security.Allow:
			security.Log("ALLOW", call.Name, call.Args, "")
			m.logReasoning("tool_allow", map[string]interface{}{"tool": call.Name, "args": call.Args})
			return m.executeTool(call, true)

		case security.Block:
			security.Log("BLOCK", call.Name, call.Args, policy.Reason)
			m.logReasoning("tool_block", map[string]interface{}{"tool": call.Name, "args": call.Args, "reason": policy.Reason})
			return types.ToolBlockedMsg{Call: call, Reason: policy.Reason}
		case security.Confirm:
			security.Log("CONFIRM", call.Name, call.Args, "")
			m.logReasoning("tool_confirm", map[string]interface{}{"tool": call.Name, "args": call.Args, "reason": policy.Reason})
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

	for _, t := range tools.Registry {
		if t.Name == call.Name {
			if t.Name == "write" || t.Name == "edit" || t.Name == "bash" {
				security.Log("EXEC", t.Name, call.Args, "")
			}
			result := t.Handler(call.Args)
			return types.ToolExecResult{ID: call.ID, Result: result}
		}
	}

	return types.ToolExecResult{ID: call.ID, Result: "error: unknown tool"}
}
