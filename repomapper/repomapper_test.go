package repomapper

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestRepoMapper(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "repomapper-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create some mock files
	files := map[string]string{
		"main.go": `package main

import "fmt"

func main() {
	s := &Store{}
	s.Save([]byte("hello"))
}
`,
		"store.go": `package main

import "fmt"

type Store struct{}

func (s *Store) Save(data []byte) error {
	fmt.Println("Saving data")
	return nil
}
`,
		"utils.py": `
def helper():
    print("Helping")

class Utils:
    def __init__(self):
        pass
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

	fmt.Println("Generated Map:")
	fmt.Println(output)

	// Basic validation
	if output == "" {
		t.Errorf("Expected output, got empty string")
	}
}
