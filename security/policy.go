package security

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Decision int

const (
	Allow Decision = iota
	Confirm
	Block
)

type PolicyResult struct {
	Decision Decision
	Reason   string
}

var alwaysAllowed = make(map[string]bool)

func LoadPermissions() {
	path := ".smallcode/permissions.json"
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var perms struct {
		AlwaysAllow []string `json:"always_allow"`
	}
	if err := json.Unmarshal(data, &perms); err == nil {
		for _, tool := range perms.AlwaysAllow {
			alwaysAllowed[tool] = true
		}
	}
}

func SavePermissions() {
	os.MkdirAll(".smallcode", 0755)
	path := ".smallcode/permissions.json"
	var perms struct {
		AlwaysAllow []string `json:"always_allow"`
	}
	for tool := range alwaysAllowed {
		perms.AlwaysAllow = append(perms.AlwaysAllow, tool)
	}
	data, _ := json.MarshalIndent(perms, "", "  ")
	os.WriteFile(path, data, 0644)
}

func AllowAlways(toolName string) {
	alwaysAllowed[toolName] = true
	SavePermissions()
}

var bashDenyList = []string{
	"rm -rf /",
	"sudo ",
	"mkfs",
	"dd if=",
	"> /dev/",
	"chmod -R 777",
	":(){ :|:&",
	"curl ",
	"wget ",
	"chown ",
	"mv /",
	"rm /",
	"find / ",
	"systemctl",
	"shutdown",
	"reboot",
	"passwd",
	"shadow",
	"| sh",
	"| bash",
	"> /etc/",
	"> /var/",
	"> /usr/",
	"> /bin/",
}

func Check(toolName string, args map[string]interface{}, cwd string) PolicyResult {
	var res PolicyResult

	switch toolName {
	case "bash":
		res = checkBash(args)
	case "read":
		res = checkRead(args, cwd)
	case "write":
		res = checkWrite(args, cwd)
	case "edit":
		res = checkEdit(args, cwd)
	case "glob":
		res = checkGlob(args, cwd)
	case "grep":
		res = checkGrep(args, cwd)
	default:
		res = PolicyResult{Decision: Allow, Reason: ""}
	}

	if res.Decision == Confirm && alwaysAllowed[toolName] {
		return PolicyResult{Decision: Allow, Reason: "always allow"}
	}

	return res
}

func checkBash(args map[string]interface{}) PolicyResult {
	cmd, _ := args["cmd"].(string)

	// Check deny-list
	for _, pattern := range bashDenyList {
		if strings.Contains(cmd, pattern) {
			return PolicyResult{
				Decision: Block,
				Reason:   fmt.Sprintf("bash command blocked: contains %q", pattern),
			}
		}
	}

	return PolicyResult{
		Decision: Confirm,
		Reason:   fmt.Sprintf("Run bash command: %s", cmd),
	}
}

func checkRead(args map[string]interface{}, cwd string) PolicyResult {
	path, _ := args["path"].(string)
	if !isPathInProject(path, cwd) {
		return PolicyResult{
			Decision: Block,
			Reason:   "path outside project directory",
		}
	}
	if isSensitivePath(path) {
		return PolicyResult{
			Decision: Block,
			Reason:   "sensitive file",
		}
	}
	return PolicyResult{Decision: Allow, Reason: ""}
}

func checkWrite(args map[string]interface{}, cwd string) PolicyResult {
	path, _ := args["path"].(string)
	if !isPathInProject(path, cwd) {
		return PolicyResult{
			Decision: Block,
			Reason:   "path outside project directory",
		}
	}
	if isSensitivePath(path) {
		return PolicyResult{
			Decision: Block,
			Reason:   "sensitive file",
		}
	}
	return PolicyResult{
		Decision: Confirm,
		Reason:   path,
	}
}

func checkEdit(args map[string]interface{}, cwd string) PolicyResult {
	path, _ := args["path"].(string)
	if !isPathInProject(path, cwd) {
		return PolicyResult{
			Decision: Block,
			Reason:   "path outside project directory",
		}
	}
	if isSensitivePath(path) {
		return PolicyResult{
			Decision: Block,
			Reason:   "sensitive file",
		}
	}
	return PolicyResult{
		Decision: Confirm,
		Reason:   path,
	}
}

func checkGlob(args map[string]interface{}, cwd string) PolicyResult {
	root := "."
	if val, ok := args["path"].(string); ok {
		root = val
	}
	if !isPathInProject(root, cwd) {
		return PolicyResult{
			Decision: Block,
			Reason:   "path outside project directory",
		}
	}
	return PolicyResult{Decision: Allow, Reason: ""}
}

func checkGrep(args map[string]interface{}, cwd string) PolicyResult {
	root := "."
	if val, ok := args["path"].(string); ok {
		root = val
	}
	if !isPathInProject(root, cwd) {
		return PolicyResult{
			Decision: Block,
			Reason:   "path outside project directory",
		}
	}
	return PolicyResult{Decision: Allow, Reason: ""}
}

func isPathInProject(path string, cwd string) bool {
	if path == "" {
		return true
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	cwdAbs, err := filepath.Abs(cwd)
	if err != nil {
		return false
	}

	// Normalize paths
	abs = filepath.Clean(abs)
	cwdAbs = filepath.Clean(cwdAbs)

	return strings.HasPrefix(abs, cwdAbs)
}

func isSensitivePath(path string) bool {
	base := filepath.Base(path)

	sensitivePatterns := []string{
		".env",
		"id_rsa",
		"id_ed25519",
		"credentials",
	}

	// Check exact matches
	for _, pattern := range sensitivePatterns {
		if base == pattern || strings.HasPrefix(base, pattern+".") {
			return true
		}
	}

	// Check contains
	if strings.Contains(path, ".ssh") ||
		strings.Contains(path, ".pem") ||
		strings.Contains(path, ".key") {
		return true
	}

	return false
}
