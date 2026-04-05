# smallcode

Minimal Claude Code alternative powered by Bubbletea TUI.

## Features

- Full agentic loop with tool use
- Tools: `read`, `write`, `edit`, `glob`, `grep`, `bash`
- Conversation history
- Rich terminal UI with lipgloss styling
- Reads config from `~/.env`

## Configuration

Configuration is loaded from `~/.env`:

```bash
OPENROUTER_API_KEY=sk-or-v1-...
MODEL=openrouter/moonshotai/kimi-k2
MAX_TOKENS=16384
```

### OpenRouter

Uses [OpenRouter](https://openrouter.ai) to access any model. Set your API key and model in `~/.env`.

### Anthropic Direct

If no `OPENROUTER_API_KEY` is set, uses Anthropic API directly. Set `ANTHROPIC_API_KEY` in `~/.env`.

## Build & Run

```bash
make build
./dist/smallcode
```

Or use `make run`.

## Controls

- `Enter` - Send message
- `Ctrl+C` / `Escape` - Quit

## Commands

- `/h` - Show help
- `/c` - Clear conversation
- `/q` or `exit` - Quit

## Tools

| Tool | Description |
|------|-------------|
| `read` | Read file with line numbers, offset/limit |
| `write` | Write content to file |
| `edit` | Replace string in file (must be unique) |
| `glob` | Find files by pattern, sorted by mtime |
| `grep` | Search files for regex |
| `bash` | Run shell command |

## Skills

Skills are specialized instruction sets that can be invoked on-demand during conversation.

### Usage

Prefix any message with `@skillname` to activate that skill:

```
@example what is this?
```

### Creating Skills

Add markdown files to `.smallcode/skills/`:

```bash
.smallcode/skills/
├── example.md
└── your-skill.md
```

Each skill file contains markdown that gets prepended to your message when invoked.

### Example Skill

`.smallcode/skills/example.md`:
```markdown
# Example Skill

You are an example skill that demonstrates the skills system.

## Purpose
Skills allow you to inject specialized instructions into your conversation on-demand.
```

## License

MIT
