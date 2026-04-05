# GEMINI.md - smallcode Project Context

This document provides essential context and instructions for AI agents working on the `smallcode` project.

## Project Overview

`smallcode` is a minimal, terminal-based AI agentic loop inspired by Claude Code. It is built using **Go** and the **Bubbletea** TUI framework. It features a full agentic loop with tool use, persistent memory, and a specialized skills system.

### Core Technologies
- **Language:** Go (1.26+)
- **TUI Framework:** [Bubbletea](https://github.com/charmbracelet/bubbletea)
- **Styling:** [Lipgloss](https://github.com/charmbracelet/lipgloss)
- **APIs:** Anthropic (direct) or OpenRouter (for various models)

### Architecture
- `main.go`: Application entry point.
- `api/client.go`: Main agentic loop, API interaction, and tool orchestration.
- `tools/`: Implementation of core agent tools (`read`, `write`, `edit`, `glob`, `grep`, `bash`, `remember`).
- `security/`: Safety layer that intercepts tool calls to `Allow`, `Block`, or request `Confirm` (e.g., for shell commands or file writes).
- `todos/`: Task management logic persisted to `.smallcode/todos.json`.
- `memory/`: Session summary and persistence logic.
- `skills/`: Loader for specialized markdown-based instructions triggered by `@skillname`.
- `config/`: Configuration management (loading from `.env`).

## Building and Running

| Task | Command |
|------|---------|
| Build | `make build` |
| Run | `make run` |
| Clean | `make clean` |
| Tidy Dependencies | `go mod tidy` |

## Development Conventions

### Agentic Loop & Tools
- The agent uses a tool-use loop. Tool results are fed back into the conversation history.
- **Security First:** Always respect the `security` package logic. Destructive tools or shell commands require user confirmation.
- **Memory & Context:** The system prompt is dynamically built in `api/client.go` using project memory (`.smallcode/memory.json`) and active todos (`.smallcode/todos.json`).

### Coding Style
- Standard Go idioms and formatting.
- Error handling should be explicit.
- TUI updates must follow the Bubbletea `Model`/`Update`/`View` pattern.

### Persistent State
- Project-specific data is stored in the `.smallcode/` directory:
  - `memory.json`: Key-value facts about the project.
  - `todos.json`: Structured list of tasks.
  - `audit.log`: Log of all tool executions and security decisions.
  - `sessions.jsonl`: Summary of past sessions.

## Key Files & Directories

- `api/client.go`: The heart of the agent. Check here for API payload structure and tool routing.
- `tools/tools.go`: Implementation of file and shell operations.
- `security/policy.go`: Definition of what tools require confirmation or are blocked.
- `types/types.go`: Central location for shared data structures.
- `.smallcode/skills/`: Directory for markdown files that define "Skills" (e.g., `@obsidian`).

## Future Interaction Guidelines

- When adding new tools, register them in `api/client.go` and implement the logic in `tools/`.
- Ensure new features that modify the filesystem are screened by the `security` package.
- Maintain the "minimalist" philosophy of the project.
