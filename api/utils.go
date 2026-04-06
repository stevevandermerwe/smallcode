package api

import (
	"os"
	"path/filepath"
	"strings"
)

func (m *Model) listAllFiles() []string {
	var files []string
	excluded := []string{".git", "node_modules", "dist", ".smallcode"}
	if data, err := os.ReadFile(".smallcode/ignore"); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, l := range lines {
			if l = strings.TrimSpace(l); l != "" {
				excluded = append(excluded, l)
			}
		}
	}

	filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		parts := strings.Split(filepath.ToSlash(path), "/")
		for _, p := range parts {
			for _, e := range excluded {
				if p == e {
					if d.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
			}
		}
		if !d.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files
}
