package api

import (
	"os"
	"testing"

	"smallcode/security"
	"smallcode/types"
)

func TestSecurityHarness(t *testing.T) {
	cwd, _ := os.Getwd()

	tests := []struct {
		name     string
		tool     string
		args     map[string]interface{}
		expected security.Decision
	}{
		{
			name:     "Block rm -rf /",
			tool:     "bash",
			args:     map[string]interface{}{"cmd": "rm -rf /"},
			expected: security.Block,
		},
		{
			name:     "Block sensitive file read",
			tool:     "read",
			args:     map[string]interface{}{"path": ".env"},
			expected: security.Block,
		},
		{
			name:     "Block outside path",
			tool:     "read",
			args:     map[string]interface{}{"path": "/etc/passwd"},
			expected: security.Block,
		},
		{
			name:     "Allow safe read",
			tool:     "read",
			args:     map[string]interface{}{"path": "main.go"},
			expected: security.Allow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := security.Check(tt.tool, tt.args, cwd)
			if res.Decision != tt.expected {
				t.Errorf("expected %v, got %v (reason: %s)", tt.expected, res.Decision, res.Reason)
			}
		})
	}
}

func TestSequentialLoopPrevention(t *testing.T) {
	m := NewModel()
	m.Model.SequentialToolCount = 6
	call := types.ToolCall{Name: "read", Args: map[string]interface{}{"path": "main.go"}}
	m.Model.ToolQueue = []types.ToolCall{call}

	msg := m.processNextTool()()
	confirm, ok := msg.(types.ToolConfirmMsg)
	if !ok {
		t.Fatalf("expected ToolConfirmMsg for loop prevention, got %T", msg)
	}
	if confirm.Reason != "loop prevention: too many sequential tool calls. Please approve to continue." {
		t.Errorf("unexpected confirm reason: %s", confirm.Reason)
	}
}
