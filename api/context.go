package api

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"smallcode/config"
	"smallcode/todos"
	"smallcode/types"
)

// BuildSystemPrompt creates the system prompt with memory context

func BuildSystemPrompt(cwd string) string {
	base := fmt.Sprintf("Concise coding assistant. cwd: %s", cwd)
	base += "\n\nTool Registry: Your capabilities are defined in `tools/registry.go`. Each tool has a name, description, and JSON schema."
	base += "\n\nSearch tools (glob, grep) automatically exclude .git, node_modules, and other build artifacts. Users may explicitly add files to context using `/add <path>` or generate a repository map with `/map`."
	if config.YOLO {
		base += "\n\n[WARNING: YOLO MODE ACTIVE] Security protections are disabled. You have full system access. Use extreme caution."
	}
	return base + "\n\n## Memory & Task Management\n- Use `todo` tool to manage work items (action=add|update|close|reopen|remove|list)\n- Set priority 1-3 (1=urgent). Use blocked_by to track dependencies (comma-separated IDs).\n- Add file:path or text:snippet sources when a todo relates to specific code.\n- Use `remember` tool to persist key project facts (action=add|update|forget)\n- When closing a todo that produced a lasting project fact, also use `remember` to record it.\n- Keep memory values under 200 chars; prefer update over adding duplicate keys\n- Users may explicitly add project memory/tasks to context using `/memory`."
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
