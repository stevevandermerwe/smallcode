package repomapper

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRepoMapper(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "repomapper-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	files := map[string]string{
		"main.go": `package main

import "fmt"

func main() {
	s := &Store{}
	s.Save([]byte("hello"))
}`,
		"store.go": `package main

type Store struct{}

func (s *Store) Save(data []byte) error {
	return nil
}`,
		"utils.py": `
def helper():
    print("Helping")

class Utils:
    def __init__(self):
        pass
`,
		"main.py": `
from utils import helper, Utils
helper()
u = Utils()
`,
	}

	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write mock file %s: %v", path, err)
		}
	}

	rm := NewRepoMapper()
	output, err := rm.GenerateMap(tmpDir)
	if err != nil {
		t.Fatalf("GenerateMap failed: %v", err)
	}

	// Validate output contains all files
	for name := range files {
		if !strings.Contains(output, name) {
			t.Errorf("Expected file %s in output, got:\n%s", name, output)
		}
	}

	// Validate Go-specific parsing
	if !strings.Contains(output, "func main") {
		t.Errorf("Expected Go function 'main' in output")
	}
	if !strings.Contains(output, "func (s *Store) Save") {
		t.Errorf("Expected Go method 'Save' in output")
	}

	// Validate Python-specific parsing
	if !strings.Contains(output, "def helper") {
		t.Errorf("Expected Python function 'helper' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "class Utils") {
		t.Errorf("Expected Python class 'Utils' in output")
	}
}
