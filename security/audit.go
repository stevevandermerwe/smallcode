package security

import (
	"fmt"
	"os"
	"time"
)

const auditLogPath = ".smallcode/audit.log"

func Log(action, toolName string, args map[string]interface{}, detail string) {
	os.MkdirAll(".smallcode", 0755)

	timestamp := time.Now().UTC().Format(time.RFC3339)
	var logLine string

	if detail != "" {
		logLine = fmt.Sprintf("%s %s %-6s %s=%s reason=%s\n", timestamp, action, toolName, argKeyForTool(toolName), detail, "")
	} else {
		logLine = fmt.Sprintf("%s %s %-6s\n", timestamp, action, toolName)
	}

	f, err := os.OpenFile(auditLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	f.WriteString(logLine)
}

func argKeyForTool(toolName string) string {
	switch toolName {
	case "bash":
		return "cmd"
	case "read", "write", "edit", "glob", "grep":
		return "path"
	default:
		return "args"
	}
}
