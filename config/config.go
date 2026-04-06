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
	BASH_TIMEOUT   int // seconds
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
	API_URL = "https://openrouter.ai/api/v1/messages"

	MODEL = os.Getenv("MODEL")
	if MODEL == "" {
		MODEL = "minimax/minimax-m2.5"
	}

	maxTokensStr := os.Getenv("MAX_TOKENS")
	MAX_TOKENS, _ = strconv.Atoi(maxTokensStr)
	if MAX_TOKENS == 0 {
		MAX_TOKENS = 16384
	}

	BASH_TIMEOUT = 30 // default 30 seconds
	timeoutStr := os.Getenv("BASH_TIMEOUT")
	if timeoutStr != "" {
		if val, err := strconv.Atoi(timeoutStr); err == nil && val > 0 {
			BASH_TIMEOUT = val
		}
	}
}

func Provider() string {
	return "OpenRouter"
}
