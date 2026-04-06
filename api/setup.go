package api

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/charmbracelet/lipgloss"

	"smallcode/config"
	"smallcode/security"
	"smallcode/types"
)

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
	} else if data, err := os.ReadFile(".smallcode/session.json"); err == nil {
		var state map[string]interface{}
		if json.Unmarshal(data, &state) == nil {
			msgs, _ := state["messages"].([]interface{})
			if len(msgs) > 0 {
				m.Model.PromptingForResume = true
				m.Model.Output = append(m.Model.Output, lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Bold(true).Render("   Previous session found. Resume? [y/n]"))
			}
		}
	}

	return m
}

func initProject() []string {
	var msgs []string

	// 1. .env (template)
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		content := "OPENROUTER_API_KEY=\nMODEL=minimax/minimax-m2.5\nMAX_TOKENS=16384\nSECURITY_LEVEL=balanced\n"
		if err := os.WriteFile(".env", []byte(content), 0644); err == nil {
			msgs = append(msgs, "✔ Created .env template")
		}
	} else {
		msgs = append(msgs, "ℹ .env already exists")
	}

	// 2. .smallcode directory and basics
	os.MkdirAll(".smallcode/skills", 0755)

	files := map[string]string{
		".gitignore":                   ".env\ndist/\n",
		".smallcode/ignore":            ".git\nnode_modules\ndist\n.smallcode\n",
		".smallcode/memory.json":       "{\n  \"version\": 1,\n  \"entries\": []\n}",
		".smallcode/todos.json":        "{\n  \"version\": 1,\n  \"todos\": []\n}",
		".smallcode/skills/example.md": "# Example Skill\n\nYou are an example skill that demonstrates the skills system.\n\n## Purpose\nSkills allow you to inject specialized instructions into your conversation on-demand.\n",
		".smallcode/skills/context.md": "# Context Audit Skill\n\nYou are an expert context manager. Your task is to audit the current project context (Memory and Todos) to ensure they are lean, relevant, and accurate.\n\n## Objectives\n1. **Audit Memory:** Use the `remember` tool with `action=forget` to remove stale or redundant facts. Use `action=update` to consolidate related facts into single, concise entries.\n2. **Audit Todos:** Use the `todo` tool with `action=list` first to see all tasks. Then:\n    - `action=close` tasks that are finished but still open.\n    - `action=remove` duplicate tasks.\n    - `action=update` tasks to add better priority or clear up blocking dependencies.\n3. **Consolidate:** If a finished task resulted in a lasting project fact, ensure that fact is in `memory` before closing the `todo`.\n\n## Goal\nKeep the system prompt under control so that it doesn't waste tokens. Be aggressive but careful not to lose critical project information.\n",
		".smallcode/skills/skills-builder.md": "# Skills Builder Skill\n\nYou are a specialized agent designed to help users create new \"skills\" for the `smallcode` project. A skill is a markdown-based set of instructions that the agent uses when triggered by an `@skillname` mention.\n\n## Process\nWhen a user invokes this skill (e.g., `@skills-builder help me make a new skill`), follow this interactive process:\n\n1.  **Initialize:** Ask the user what they want the new skill to do.\n2.  **Gather Information:** Ask targeted questions to build out the skill's content:\n    *   **Name:** What should the skill be called (e.g., `obsidian`, `git-expert`, `deploy`)?\n    *   **Purpose:** What is the high-level goal of this skill?\n    *   **Workflows/Instructions:** What specific steps or rules should the agent follow when this skill is active?\n    *   **Tools/Commands:** Are there specific CLI tools (e.g., `git`, `npm`, `obsidian`) or agent tools (e.g., `read`, `write`, `bash`) it should prioritize or use in a certain way?\n    *   **Conventions:** Are there any formatting, naming, or architectural conventions it should enforce?\n    *   **Examples:** Can the user provide example usages or commands?\n3.  **Draft:** Once you have sufficient information, present a draft of the markdown file to the user.\n4.  **Refine:** Ask if they want any changes.\n5.  **Finalize:** Use the `write` tool to save the markdown file to `.smallcode/skills/<name>.md`.\n\n## Content Structure\nNew skills should generally follow this template:\n\n```markdown\n# <Skill Name> Skill\n\n<Brief description of what this skill does.>\n\n## Purpose\n<What is the primary objective of this skill?>\n\n## Usage\n<How to trigger it and example commands/questions.>\n\n## Instructions\n<Detailed, step-by-step instructions or rules for the agent.>\n\n## Workflows (Optional)\n<Specific sequences of actions to take for common tasks.>\n\n## Conventions (Optional)\n<Coding styles, file structures, or naming rules to follow.>\n```\n\n## Goal\nMake the process as frictionless as possible while ensuring the resulting skill is high-quality and provides clear instructions to the agent.\n",
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
