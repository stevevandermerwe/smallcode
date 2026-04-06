package api

import (
	"github.com/charmbracelet/lipgloss"
)

// Styles

var (
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("235")).
			Padding(0, 1).
			Bold(true).
			MarginBottom(1)

	resetStyle      = lipgloss.NewStyle()
	dimStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	errorStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	assistantStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	userPrefix      = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true).Render("❯")
	assistantPrefix = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true).Render("◇")
	toolStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	skillStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Bold(true)
)

const helpText = `smallcode - Commands & Tips

Commands:
  /h, /help         Show this help
  /init             Initialize project (.env, .smallcode, git init)
  /debug, /d       Toggle debug mode (shows API details, token counts, tool args)
  /trace, /t       Toggle trace mode (logs ALL raw traffic to .smallcode/trace.log)
  /yolo, /y         Toggle YOLO mode (bypasses ALL security protections)
  /add <path>       Add a file's content to the context
                    Tip: Type '/add ' to see a recursive file selection dropdown.
  /map, /m          Add repository skeleton map to context
  /memory, /mem     Add project memory and tasks to context
  /summarize, /s    Summarize conversation and prompt to save facts
  /clear, /c        Clear conversation
  /quit, /exit, /q  Quit

Skills:
  @skillname        Activate a skill (e.g. @example)
                    Tip: Type '@' to see a dropdown of available skills.
  @skills-builder   Interactive tool to help you create new skills.

Tools:
  read        Read file (path, offset?, limit?)
  write       Write file (path, content)
  edit        Edit file (path, old, new, all?)
  glob        Find files (pat, path?)
  grep        Search regex (pat, path?)
  bash        Run shell (cmd)
  map         Codebase map (path?)
  remember    Memory (action, key, value?, tags?)
  todo        Todos (action, id?, title?, priority?, blocked_by?, sources?, status?)`
