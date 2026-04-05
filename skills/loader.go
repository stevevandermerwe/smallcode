package skills

import (
	"os"
	"path/filepath"
	"strings"
)

var SkillsDir = ".smallcode/skills"

func Load(name string) (string, error) {
	data, err := os.ReadFile(filepath.Join(SkillsDir, name+".md"))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func List() []string {
	entries, err := os.ReadDir(SkillsDir)
	if err != nil {
		return nil
	}
	var skillNames []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			name := strings.TrimSuffix(e.Name(), ".md")
			skillNames = append(skillNames, name)
		}
	}
	return skillNames
}
