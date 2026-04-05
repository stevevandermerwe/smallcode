# SmallCode

Minimal AI powered generator powered by Bubbletea TUI.

## Features

- **Full Agentic Loop:** Seamless tool-use integration.
- **Secure by Default:** Advanced security layer with Allow/Block/Confirm logic for destructive operations and shell commands.
- **YOLO Mode:** Optional unrestricted mode for power users.
- **Smart Context Management:** Includes auto-pruning, conversation summarizing (`/s`), and token warning indicators.
- **Persistent Memory:** Track project facts and tasks across sessions using `remember` and `todo` tools.
- **Skills System:** Specialized instruction sets (e.g., `@context`, `@example`) invoked on-demand.
- **Repository Mapping:** Universal codebase "skeleton" mapper using Tree-sitter for multi-language symbol extraction and ranking.
- **Automatic Exclusions:** Search tools automatically ignore `.git`, `node_modules`, and build artifacts.
- **Rich TUI:** Modern terminal interface with real-time token tracking and lipgloss styling.

## Installation & Setup

### 1. Build from Source
```bash
make build
```

### 2. Initialize Project
Run the following command inside your project directory to set up the necessary structure:
```bash
./dist/smallcode
# Inside the TUI, type:
/init
```
This will create a `.env` template, initialize the `.smallcode/` directory, and run `git init` if needed.

## Configuration

Configuration is loaded from `.env` in the current directory or `~/.env`:

```bash
OPENROUTER_API_KEY=sk-or-v1-...
MODEL=minimax/minimax-m2.5
MAX_TOKENS=16384
```

## Controls

- `Enter` - Send message
- `Ctrl+C` / `Escape` - Quit

## Commands

| Command | Description |
|---------|-------------|
| `/h`, `/help` | Show help menu |
| `/init` | Initialize project (.env, .smallcode, git init) |
| `/add <path>`| Explicitly add a file's content to the conversation context |
| `/s`, `/summarize` | Summarize conversation and clear history (keeping context lean) |
| `/c`, `/clear` | Clear conversation history |
| `/debug`, `/d` | Toggle debug mode (token counts, raw tool args) |
| `/trace`, `/t` | Toggle raw LLM traffic logging to `.smallcode/trace.log` |
| `/yolo`, `/y` | Toggle YOLO mode (bypasses all security protections) |
| `/map`, `/m` | Generate and add a repository skeleton map to context |
| `/q`, `/exit`, `exit` | Quit |

## Tools

| Tool | Description |
|------|-------------|
| `read` | Read file with line numbers, offset/limit |
| `write` | Write content to file |
| `edit` | Replace string in file (must be unique) |
| `glob` | Find files by pattern (automatically excludes .git, etc.) |
| `grep` | Search files for regex (automatically excludes .git, etc.) |
| `bash` | Run shell commands (sandboxed by default) |
| `map` | Generate a hierarchical codebase map (Go, Python, Java, JS/TS) |
| `remember` | Persist a fact to project memory (`.smallcode/memory.json`) |
| `todo` | Manage project tasks and dependencies (`.smallcode/todos.json`) |

## Skills

Skills are specialized instruction sets stored as markdown files in `.smallcode/skills/`.

### Usage
Prefix any message with `@skillname` to activate that skill:
```
@context audit my memory
```

### Included Skills
- `@context`: Audit and consolidate your project's memory and tasks.
- `@example`: A demonstration of the skills system.

## Security & Isolation

- **Sandboxing:** Bash commands run in a restricted environment with limited PATH and isolated HOME/PWD.
- **Directory Isolation:** File tools are restricted to the project root and cannot access sensitive system files (e.g., `.env`, `.ssh/`).
- **Confirmation:** Destructive actions like `write`, `edit`, and `bash` require manual user approval ('y' to approve, 'n' to deny).
- **Resource Limits:** Tool outputs are automatically truncated to prevent memory exhaustion and TUI clutter.

## Library & Modules

### RepoMapper
The `repomapper` module provides a universal way to generate a hierarchical map of a codebase for LLM context. It supports Go, Python, Java, and TypeScript/JavaScript using Tree-sitter.

#### Usage:
```go
import "smallcode/repomapper"

rm := repomapper.NewRepoMapper()
output, err := rm.GenerateMap("./")
fmt.Println(output)
```

## License

MIT
