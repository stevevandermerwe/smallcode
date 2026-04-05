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
	res = Bash(map[string]interface{}{"cmd": "pwd"})
	cwd, _ := os.Getwd()
	if !strings.Contains(res, cwd) {
		t.Errorf("Expected bash to run in %s, got: %s", cwd, res)
	}

	// 3. Test Bash Output Truncation
	res = Bash(map[string]interface{}{"cmd": "printf 'a%.0s' {1..40000}"})
	if !strings.Contains(res, "output truncated") {
		t.Errorf("Expected bash output to be truncated, got size: %d", len(res))
	}

	// 4. Test YOLO Mode
	config.YOLO = true
	res = Bash(map[string]interface{}{"cmd": "printf 'a%.0s' {1..40000}"})
	if strings.Contains(res, "output truncated") {
		t.Errorf("Expected YOLO mode to NOT truncate output, but it was truncated")
	}
	if len(res) < 40000 {
		t.Errorf("Expected large output in YOLO mode, got size: %d", len(res))
	}
	config.YOLO = false
}
