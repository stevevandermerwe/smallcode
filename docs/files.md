# SmallCode Internal Files (`.smallcode/`)

This document describes the purpose and structure of the files maintained by SmallCode in the `.smallcode/` directory.

## Core State & Metadata

| File | Purpose |
|------|---------|
| `memory.json` | Stores long-term project facts and technical context. Managed via the `remember` tool. |
| `todos.json` | Stores the project's task list, including priorities and dependencies. Managed via the `todo` tool. |
| `session.json` | Stores the **current active session state** (full message history and token counts). Used for crash recovery and resuming sessions. |
| `permissions.json` | (Optional) Stores tool-specific "Always Allow" permissions granted by the user. |
| `ignore` | (Optional) Custom line-delimited ignore patterns for search tools (`glob`, `grep`) and file listing. |

## Observability & Logging

| File | Purpose |
|------|---------|
| `reasoning.jsonl` | Structured JSON log of the agent's internal reasoning loop, including tool selection, security checks, and context pruning. |
| `audit.log` | Human-readable log of all tool executions and security decisions (ALLOW/BLOCK/CONFIRM). |
| `sessions.jsonl` | Historical ledger storing a summary record (start time, duration, message count) of every completed session. |
| `trace.log` | (Optional) Raw dump of all API request/response payloads. Enabled via `/trace`. |

## Specialized Instructions

### `skills/`
A directory containing Markdown files that define "Skills".
- **Trigger:** Invoked in the TUI using the `@skillname` syntax.
- **Function:** Prepend specific instructions or system prompts to the current conversation turn to specialize the agent's behavior.

#### Default Skills:
- `skills-builder.md`: An interactive assistant to help you create and save new skills.
- `context.md`: Expert context manager for auditing memory and tasks.
- `obsidian.md`: Integration for reading and writing documentation in an Obsidian vault.
- `example.md`: A simple demonstration of the skills system.
