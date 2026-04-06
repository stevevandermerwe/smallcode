package tools

import (
	"os"
	"strings"
	"testing"

	"smallcode/config"
)

func TestSecurity(t *testing.T) {
	// 1. Test Output Truncation in Read
	err := os.WriteFile("large_file.txt", []byte(strings.Repeat("a", MaxOutputSize+100)), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("large_file.txt")

	res := Read(map[string]interface{}{"path": "large_file.txt"})
	if !strings.Contains(res, "truncated") {
		t.Errorf("Expected output to be truncated, got: %s", res)
	}

	// 2. Test Bash Sandbox (Working Directory)
	config.BASH_TIMEOUT = 5
	res = Bash(map[string]interface{}{"cmd": "pwd"})
	cwd, _ := os.Getwd()
	if !strings.Contains(res, cwd) {
		t.Errorf("Expected bash to run in %s, got: %s", cwd, res)
	}

	// 3. Test Bash Output Truncation
	// Generate 40KB of output, which is > 32KB limit
	res = Bash(map[string]interface{}{"cmd": "head -c 40000 /dev/zero | tr '\\0' 'a'"})
	if !strings.Contains(res, "output truncated") {
		t.Errorf("Expected bash output to be truncated, got size: %d, output: %s", len(res), res)
	}

	// 4. Test YOLO Mode
	config.YOLO = true
	// In YOLO mode, limit is 1MB, so 40KB should NOT be truncated
	res = Bash(map[string]interface{}{"cmd": "head -c 40000 /dev/zero | tr '\\0' 'a'"})
	if strings.Contains(res, "output truncated") {
		t.Errorf("Expected YOLO mode to NOT truncate output, but it was truncated")
	}
	if len(res) < 40000 {
		t.Errorf("Expected large output in YOLO mode, got size: %d", len(res))
	}
	config.YOLO = false

	// 5. Test Exclusions
	os.MkdirAll("test_exclude/.git", 0755)
	os.WriteFile("test_exclude/.git/secret.txt", []byte("secret"), 0644)
	os.WriteFile("test_exclude/normal.txt", []byte("normal"), 0644)
	defer os.RemoveAll("test_exclude")

	res = Glob(map[string]interface{}{"pat": "*", "path": "test_exclude"})
	if strings.Contains(res, ".git") {
		t.Errorf("Expected Glob to exclude .git, but it was found: %s", res)
	}
	if !strings.Contains(res, "normal.txt") {
		t.Errorf("Expected Glob to find normal.txt, but it was not found: %s", res)
	}

	res = Grep(map[string]interface{}{"pat": "secret", "path": "test_exclude"})
	if res != "none" {
		t.Errorf("Expected Grep to skip .git and return none, got: %s", res)
	}
}
