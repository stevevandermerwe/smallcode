# 🤖 SmallCode

An AI-powered code assistant that lives in your terminal. SmallCode helps you explore, understand, and modify your codebase with any AI model—all from a beautiful TUI interface.

Think of it as having an expert pair programmer right in your terminal that understands your entire project and follows your security rules.

## ✨ What You Can Do

- Ask your AI assistant questions about your code
- Have the AI read, understand, and modify files
- Run terminal commands with safety checks
- Keep track of project tasks and learnings across sessions
- Create custom "skills" for specialized workflows
- All with fine-grained control over what the AI can do

## 🎯 Key Features

- 🔐 **Secure by Default** — Advanced security with tiered permissions (`strict`, `balanced`, `relaxed`). You approve what the AI can do.
- 🧠 **Agentic Loop** — The AI can read files, understand your code, and take actions automatically (with your approval).
- 💾 **Session Persistence** — Your conversations, tasks, and project knowledge are saved automatically.
- 🔄 **Smart Context** — Automatically manages token usage, summarizes conversations, and keeps things organized.
- 📝 **Project Memory** — Track facts and tasks that persist across sessions with `/remember` and `/todo`.
- 🎓 **Skills System** — Create custom instruction sets for specialized workflows (e.g., `@code-review`, `@docs-writer`).
- 🗺️ **Code Mapping** — Automatically generates a "skeleton" of your codebase for quick understanding (supports Go, Python, Java, JS/TS).
- 🛡️ **Safety First** — Prevents accidental destructive operations and asks for confirmation on risky actions.
- 📊 **Rich Terminal UI** — Beautiful, modern interface with real-time token tracking and skill selection.
- 🚀 **YOLO Mode** — Optional unrestricted mode for power users who know what they're doing.

## 🚀 Quick Start

### Prerequisites

You'll need a few tools on your machine:

- **Go 1.26+** — Available at https://golang.org/dl/
- **Make** — Usually comes with your system (test with `make --version`)
- **Git** — Available at https://git-scm.com/
- **OpenRouter API Key** — Free tier available at https://openrouter.ai/ (you'll need this to use AI models)

### Installation & Setup

1. **Build the project**
   ```bash
   make build
   ```

2. **Run the TUI**
   ```bash
   ./dist/smallcode
   ```

3. **Initialize your project**
   Inside the TUI, type:
   ```
   /init
   ```
   This will create a `.env` template and set everything up. You'll need to add your API key to `.env`.

4. **Run tests** (optional, but recommended)
   ```bash
   make test          # Run all unit tests
   make test-harness  # Run end-to-end security tests
   ```

## ⚙️ Configuration

Create a `.env` file in your project root (or in your home directory `~/.env`):

```bash
# Your OpenRouter API key (get one free at https://openrouter.ai/)
OPENROUTER_API_KEY=sk-or-v1-...

# Which AI model to use (default: minimax/minimax-m2.5)
# Options: gpt-4, claude-3.5-sonnet, etc.
MODEL=minimax/minimax-m2.5

# Maximum tokens per response (higher = longer responses but costs more)
MAX_TOKENS=16384

# Security level: how careful the AI should be
# strict   = Ask before every action
# balanced = Auto-approve safe actions, ask about risky ones (recommended)
# relaxed  = Auto-approve almost everything
SECURITY_LEVEL=balanced
```

**💡 Tip:** Run `/init` inside the TUI to automatically create a `.env` template with these settings.

## ⌨️ Keyboard Controls

- **Enter** — Send a message or select a skill
- **Tab** — Auto-complete skill names
- **Up/Down** — Navigate dropdown menus
- **Ctrl+C** or **Escape** — Quit (or close a dropdown)

## 🎮 Commands

### Getting Started
- `/help` or `/h` — Show the help menu
- `/init` — Set up your project (creates `.env`, `.smallcode/`, initializes git)

### Working with Files
- `/add <path>` — Add a file's content to the conversation. Tip: Type `/add ` without a path to browse files interactively.
- `/map` or `/m` — Generate a visual skeleton of your codebase to help the AI understand the structure

### Memory & Tasks
- `/memory` or `/mem` — Show project facts and tasks you've saved
- Remember facts using the `remember` tool in conversations
- Track tasks using the `todo` tool

### Conversation Management
- `/summarize` or `/s` — Compress the conversation history (saves tokens for long conversations)
- `/clear` or `/c` — Clear the conversation history (start fresh)

### Debug & Power User
- `/debug` or `/d` — Toggle debug mode to see token counts and tool details
- `/trace` or `/t` — Log all LLM requests/responses to `.smallcode/trace.log` (great for troubleshooting)
- `/yolo` or `/y` — Toggle unrestricted mode (skips all security checks—use with caution! 🚀)

### Exiting
- `/exit`, `/q`, or `exit` — Quit SmallCode

## 🛠️ Available Tools

The AI can use these tools to help you. Which tools are available depends on your security level and active skill:

### File Operations
- **`read`** — Read a file's contents with line numbers (great for understanding code)
- **`write`** — Create or completely replace a file
- **`edit`** — Make surgical changes to a file by finding and replacing text

### Search & Explore
- **`glob`** — Find files by pattern (e.g., `*.py` or `src/**/*.go`). Automatically skips `.git`, `node_modules`, etc.
- **`grep`** — Search file contents for text/regex patterns. Also skips ignored directories.

### Project Understanding
- **`map`** — Generate a visual tree of your codebase with key symbols highlighted. Supports Go, Python, Java, JavaScript, and TypeScript.

### Project Memory
- **`remember`** — Save a fact that the AI should know. Saved to `.smallcode/memory.json` and restored each session.
- **`todo`** — Create, update, and track project tasks. Saved to `.smallcode/todos.json`.

### Power User
- **`bash`** — Run shell commands (with safety checks). Great for running tests, git commands, deployments, etc.

> **Note:** The AI will only use tools appropriate for the current task and your security settings.

## 🎓 Skills (Custom Instructions)

Skills are like "personalities" or "specializations" for the AI. They give it specific instructions for different types of tasks.

### Built-in Skills
- **`@context`** — Review and organize your project memory and tasks
- **`@skills-builder`** — Interactive tool to create new custom skills
- **`@example`** — A demo skill showing how the skills system works

### Creating Your Own Skills
You can create custom skills for your workflows:

```
Type: @skills-builder
```

This opens an interactive guide to create skills like:
- `@code-review` — For reviewing code quality
- `@docs-writer` — For writing documentation
- `@refactor` — For systematic code refactoring

Skills are saved in `.smallcode/skills/` and available across all your projects.

### Using Skills
1. Type `@skillname` before your message
2. Or type `@` and select from a dropdown menu

## 🔒 Security & Safety

SmallCode is designed to be safe by default. Here's how it protects your code:

### Three Security Levels
- **`strict`** — Ask before every action (most cautious)
- **`balanced`** — Auto-approve safe actions, ask about risky ones (recommended for most users)
- **`relaxed`** — Auto-approve almost everything (trust the AI, but less safe)

### How It Works
- **Risk Assessment** — Every action the AI wants to take is classified as Low, Medium, or High risk
- **Auto-Approval** — Low-risk actions are handled automatically based on your security level
- **Confirmation** — High-risk actions (like deleting files) always require your approval
- **Sandboxing** — Shell commands run in a restricted environment with limited access
- **No System Access** — File tools can't access your `.env`, `.ssh/`, or other sensitive system files
- **Context Limits** — Conversations are automatically pruned to prevent accidental context overflows

### Emergency Override
- **YOLO Mode** (`/yolo`) — Disables all safety checks (useful when you trust the AI completely, but be careful!)

> **Best Practice:** Keep `SECURITY_LEVEL=balanced` unless you have a specific reason to change it.

## 📚 RepoMapper (For Developers)

If you want to use SmallCode's code mapping library in your own Go projects, you can import it:

```go
import "smallcode/repomapper"

rm := repomapper.NewRepoMapper()
tree, err := rm.GenerateMap("./")
if err != nil {
    log.Fatal(err)
}
fmt.Println(tree)  // Prints a hierarchical code structure
```

This generates a visual tree of your codebase and extracts key symbols (functions, classes, etc.) for AI understanding. It supports:
- 🔵 **Go** — Functions, methods, types
- 🐍 **Python** — Functions, classes
- ☕ **Java** — Methods, classes, interfaces
- 📘 **JavaScript/TypeScript** — Functions, classes, methods, interfaces


## 🚀 Deployment (Linux)

Want to run SmallCode on a Linux server or VM? Here's how:

### Building for Linux

```bash
# Build for Linux x86-64
make build-linux-amd64

# Build for Linux ARM64 (Raspberry Pi, Apple Silicon VM, etc.)
make build-linux-arm64

# Or build both at once
make build-linux
```

The binaries will be in `dist/`.

### Running on Linux

1. Copy the binary to your Linux machine
2. Make it executable: `chmod +x smallcode-linux-amd64`
3. Run it: `./smallcode-linux-amd64`

### VM Deployment (macOS to Linux)

If you're using **Tart** for virtualization on macOS:

```bash
# Install Tart
brew install tart sshpass

# Clone a Debian image
tart clone ghcr.io/cirruslabs/debian:latest my-debian-vm

# Start the VM
tart run my-debian-vm --no-graphics &

# Deploy SmallCode
make build-linux-arm64
bash deploy.sh  # Automated deployment script
```

The script will:
- Copy the binary to `~/bin/smallcode` on the VM
- Set up permissions
- Add `~/bin` to your PATH

Then just run `smallcode` on the VM!



## ❓ FAQ & Tips

**Q: How do I get an API key?**
A: Go to https://openrouter.ai/, sign up (free), and generate an API key in your account settings.

**Q: Can I use a different AI model?**
A: Yes! Set the `MODEL` variable in `.env`. OpenRouter supports Claude, GPT-4, and many others.

**Q: What if the AI wants to do something I don't trust?**
A: Don't approve it! Your security level controls how much auto-approval happens. Higher `SECURITY_LEVEL=strict` means you approve everything.

**Q: Can I use SmallCode offline?**
A: Not currently—it needs internet to reach OpenRouter's API.

**Q: How do I contribute?**
A: We'd love your help! Check out the code and feel free to submit PRs.

**Q: Is my code safe?**
A: Your code is sent to OpenRouter and the selected AI model. Review your security level and only use trusted models. For sensitive code, consider self-hosting.

**Q: Can I create custom skills?**
A: Absolutely! Type `@skills-builder` to create custom skills for your workflows.

## 💡 Pro Tips

- Use `/add` to add large files to context before asking questions
- Use `/summarize` to compress long conversations and save tokens
- Use `/map` to help the AI understand your codebase structure
- Save facts with `/remember` so the AI knows about your project
- Create skills for tasks you do repeatedly

## 📝 License

MIT
