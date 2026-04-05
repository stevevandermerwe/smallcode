package types

import "time"

// API Types

type Message struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type ContentBlock struct {
	Type   string                 `json:"type"`
	Text   string                 `json:"text,omitempty"`
	ID     string                 `json:"id,omitempty"`
	Name   string                 `json:"name,omitempty"`
	Input  map[string]interface{} `json:"input,omitempty"`
	Result string                 `json:"content,omitempty"`
	ToolID string                 `json:"tool_use_id,omitempty"`
}

type ToolResult struct {
	Name   string
	Args   map[string]interface{}
	Result string
	Error  string
}

// API Messages

type APIResponse struct {
	Content      []ContentBlock
	ToolResults  []ContentBlock
	InputTokens  int
	OutputTokens int
	Err          error
}

type ToolOutput struct {
	Name   string
	Result string
	Err    string
}

// Model - exported fields for external access

type Model struct {
	Messages    []Message
	ToolResults []ToolResult
	Input       string
	Output      []string
	Waiting     bool
	Provider    string
	ModelName   string
	Err         error
	Cursor      int
	Width       int
	Height      int
	ScrollOffset int
	StartTime   time.Time

	// Security & tool execution
	PendingConfirm   *ToolConfirmMsg
	ToolQueue        []ToolCall
	CollectedResults []ContentBlock
	AssistantBlocks  []ContentBlock

	// Debug & telemetry
	Debug             bool
	Yolo              bool
	Summarizing       bool
	PromptingForApiKey bool
	TotalInputTokens  int
	TotalOutputTokens int
}

// Memory Types

type MemEntry struct {
	ID    string   `json:"id"`
	Key   string   `json:"key"`
	Value string   `json:"value"`
	Tags  []string `json:"tags,omitempty"`
	Added string   `json:"added"`
}

type MemFile struct {
	Version int       `json:"version"`
	Updated string    `json:"updated"`
	Entries []MemEntry `json:"entries"`
}

// Todo Types

type Source struct {
	Type string `json:"type"` // "file" or "text"
	Ref  string `json:"ref"`
}

type Todo struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Status    string   `json:"status"` // "pending", "active", "done"
	Priority  int      `json:"priority,omitempty"`
	BlockedBy []string `json:"blocked_by,omitempty"`
	Sources   []Source `json:"sources,omitempty"`
	Created   string   `json:"created"`
	Closed    string   `json:"closed,omitempty"`
}

type TodoFile struct {
	Version int    `json:"version"`
	Updated string `json:"updated"`
	Todos   []Todo `json:"todos"`
}

type SessionRecord struct {
	Session     string `json:"session"`
	DurationMin int    `json:"duration_min"`
	MsgCount    int    `json:"msg_count"`
}

// Security & Tool Execution Types

type ToolCall struct {
	ID   string
	Name string
	Args map[string]interface{}
}

type ToolConfirmMsg struct {
	Call   ToolCall
	Reason string
}

type ToolExecResult struct {
	ID     string
	Result string
}

type ToolBlockedMsg struct {
	Call   ToolCall
	Reason string
}