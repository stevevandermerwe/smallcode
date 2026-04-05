package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	OPENROUTER_KEY string
	API_URL        string
	MODEL          string
	MAX_TOKENS     int
	YOLO           bool
)

func LoadDotenv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
			if os.Getenv(key) == "" {
				os.Setenv(key, value)
			}
		}
	}
}

func Init() {
	LoadDotenv(".env")
	home, _ := os.UserHomeDir()
	LoadDotenv(filepath.Join(home, ".env"))

	OPENROUTER_KEY = os.Getenv("OPENROUTER_API_KEY")
	API_URL = "https://api.anthropic.com/v1/messages"
	if OPENROUTER_KEY != "" {
		API_URL = "https://openrouter.ai/api/v1/messages"
	}

	MODEL = os.Getenv("MODEL")
	if MODEL == "" {
		if OPENROUTER_KEY != "" {
			MODEL = "anthropic/claude-3.5-sonnet"
		} else {
			MODEL = "claude-3-5-sonnet-20241022"
		}
	}

	maxTokensStr := os.Getenv("MAX_TOKENS")
	MAX_TOKENS, _ = strconv.Atoi(maxTokensStr)
	if MAX_TOKENS == 0 {
		MAX_TOKENS = 8192
	}
}

func Provider() string {
	if OPENROUTER_KEY != "" {
		return "OpenRouter"
	}
	return "Anthropic"
}