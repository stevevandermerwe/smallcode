# SmallCode

Minimal AI powered generator powered by Bubbletea TUI.

## Features

- **Full Agentic Loop:** Seamless tool-use integration.
- **Secure by Default:** Advanced security layer with Allow/Block/Confirm logic for destructive operations and shell commands.
- **Tiered Permissions:** Configurable security levels (`strict`, `balanced`, `relaxed`) to manage tool approvals.
- **Session Restoration:** Automatically saves session state to `.smallcode/session.json`. Resumes conversation and token counts on startup.
- **Loop Prevention:** Automatically pauses and requests confirmation if the agent enters a high-frequency tool loop.
- **YOLO Mode:** Optional unrestricted mode for power users.
- **Smart Context Management:** Includes auto-pruning, conversation summarizing (`/s`), and token warning indicators.
- **Persistent Memory:** Track project facts and tasks across sessions using `remember` and `todo` tools.
- **Skills System:** Specialized instruction sets (e.g., `@context`, `@skills-builder`) invoked on-demand.
- **Repository Mapping:** Universal codebase "skeleton" mapper using Tree-sitter for multi-language symbol extraction and ranking.
- **Automatic Exclusions:** Search tools automatically ignore `.git`, `node_modules`, and build artifacts.
- **Observability:** Structured logs for raw traffic (`trace.log`), tool execution (`audit.log`), and agent reasoning (`reasoning.jsonl`).
- **Rich TUI:** Modern terminal interface with real-time token tracking and interactive skill selection.

## Prerequisites

To build and run `smallcode`, you need the following tools installed:

- **Go 1.26+**: The core language used for the project.
- **Make**: For running build and development commands.
- **Git**: For version control and project initialization.

## Installation & Setup

### 1. Build from Source
```bash
make build
```

### 2. Run Tests
```bash
make test          # Run all unit tests
make test-harness  # Run end-to-end security harness tests
```

### 3. Initialize Project
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
SECURITY_LEVEL=balanced  # strict, balanced, relaxed
```

## Controls

- `Enter` - Send message / Select skill
- `Tab` - Autocomplete skill selection
- `Up/Down` - Navigate skill selection dropdown
- `Ctrl+C` / `Escape` - Quit (or close skill dropdown)

## Commands

| Command | Description |
|---------|-------------|
| `/h`, `/help` | Show help menu |
| `/init` | Initialize project (.env, .smallcode, git init) |
| `/add <path>`| Explicitly add a file's content to the conversation context. Tip: Type `/add ` to trigger a recursive file selection dropdown. |
| `/s`, `/summarize` | Summarize conversation and clear history (keeping context lean) |
| `/c`, `/clear` | Clear conversation history |
| `/debug`, `/d` | Toggle debug mode (token counts, raw tool args) |
| `/trace`, `/t` | Toggle raw LLM traffic logging to `.smallcode/trace.log` |
| `/yolo`, `/y` | Toggle YOLO mode (bypasses all security protections) |
| `/map`, `/m` | Generate and add a repository skeleton map to context |
| `/memory`, `/mem` | Add project memory and tasks to context |
| `/q`, `/exit`, `exit` | Quit |

## Tools

SmallCode uses a **Metadata-First Tool Registry** (defined in `tools/registry.go`). Tools are dynamically filtered based on the active skill to reduce context overhead.

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
- Prefix any message with `@skillname` to activate that skill.
- **Tip:** Type `@` to trigger an interactive dropdown of available skills.

### Included Skills
- `@context`: Audit and consolidate your project's memory and tasks.
- `@skills-builder`: An interactive assistant to help you define and save new skills.
- `@example`: A demonstration of the skills system.

## Security & Isolation

- **Tiered Risk Management:** Every tool is assigned a Risk Tier (Low, Medium, High). `SECURITY_LEVEL` controls auto-approval for lower tiers.
- **Sandboxing:** Bash commands run in a restricted environment with limited PATH and isolated HOME/PWD.
- **Directory Isolation:** File tools are restricted to the project root and cannot access sensitive system files (e.g., `.env`, `.ssh/`).
- **Confirmation:** High-risk actions always require manual user approval unless YOLO mode is active.
- **Resource Limits:** Tool outputs and conversation history are automatically managed to prevent context overflow.

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
