package tools

import (
	"smallcode/todos"
	"smallcode/types"
)

var Registry = []types.Tool{
	{
		Name:        "read",
		Description: "Read file with line numbers",
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path":   map[string]interface{}{"type": "string"},
				"offset": map[string]interface{}{"type": "integer"},
				"limit":  map[string]interface{}{"type": "integer"},
			},
			"required": []string{"path"},
		},
		Handler: Read,
		Skills:  []string{"core", "dev", "ops", "skills-builder"},
		Tier:    types.RiskLow,
	},
	{
		Name:        "write",
		Description: "Write content to file",
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path":    map[string]interface{}{"type": "string"},
				"content": map[string]interface{}{"type": "string"},
			},
			"required": []string{"path", "content"},
		},
		Handler: Write,
		Skills:  []string{"core", "dev", "ops", "skills-builder"},
		Tier:    types.RiskMedium,
	},
	{
		Name:        "edit",
		Description: "Replace old with new in file",
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{"type": "string"},
				"old":  map[string]interface{}{"type": "string"},
				"new":  map[string]interface{}{"type": "string"},
				"all":  map[string]interface{}{"type": "boolean"},
			},
			"required": []string{"path", "old", "new"},
		},
		Handler: Edit,
		Skills:  []string{"core", "dev", "ops"},
		Tier:    types.RiskMedium,
	},
	{
		Name:        "glob",
		Description: "Find files by pattern",
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"pat":  map[string]interface{}{"type": "string"},
				"path": map[string]interface{}{"type": "string"},
			},
			"required": []string{"pat"},
		},
		Handler: Glob,
		Skills:  []string{"core", "dev", "ops"},
		Tier:    types.RiskLow,
	},
	{
		Name:        "grep",
		Description: "Search files for regex",
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"pat":  map[string]interface{}{"type": "string"},
				"path": map[string]interface{}{"type": "string"},
			},
			"required": []string{"pat"},
		},
		Handler: Grep,
		Skills:  []string{"core", "dev", "ops"},
		Tier:    types.RiskLow,
	},
	{
		Name:        "bash",
		Description: "Run shell command",
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"cmd": map[string]interface{}{"type": "string"},
			},
			"required": []string{"cmd"},
		},
		Handler: Bash,
		Skills:  []string{"dev", "ops"},
		Tier:    types.RiskHigh,
	},
	{
		Name:        "map",
		Description: "Generate a hierarchical codebase map (Go, Python, Java, JS/TS) with symbol ranking",
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Directory to map (default: .)",
				},
			},
		},
		Handler: Map,
		Skills:  []string{"dev", "ops"},
		Tier:    types.RiskLow,
	},
	{
		Name:        "remember",
		Description: "Persist a fact to project memory",
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"action": map[string]interface{}{
					"type": "string",
					"enum": []string{"add", "update", "forget"},
				},
				"key":   map[string]interface{}{"type": "string"},
				"value": map[string]interface{}{"type": "string"},
				"tags":  map[string]interface{}{"type": "string"},
			},
			"required": []string{"action", "key"},
		},
		Handler: Remember,
		Skills:  []string{"core", "mem"},
		Tier:    types.RiskLow,
	},
	{
		Name:        "todo",
		Description: "Manage project todos with priorities and dependencies",
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"action": map[string]interface{}{
					"type": "string",
					"enum": []string{"add", "update", "close", "reopen", "remove", "list"},
				},
				"id":         map[string]interface{}{"type": "string"},
				"title":      map[string]interface{}{"type": "string"},
				"priority":   map[string]interface{}{"type": "integer"},
				"blocked_by": map[string]interface{}{"type": "string"},
				"sources":    map[string]interface{}{"type": "string"},
				"status":     map[string]interface{}{"type": "string"},
			},
			"required": []string{"action"},
		},
		Handler: todos.Execute,
		Skills:  []string{"core", "mem"},
		Tier:    types.RiskLow,
	},
}
