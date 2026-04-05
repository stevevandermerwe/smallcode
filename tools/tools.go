package tools

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"smallcode/config"
	"smallcode/types"
)

const MaxOutputSize = 32 * 1024 // 32KB

func Read(args map[string]interface{}) string {
	path, _ := args["path"].(string)
	
	// Open file with limit
	f, err := os.Open(path)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}

	size := stat.Size()
	limit := int64(MaxOutputSize)
	if config.YOLO {
		limit = 10 * 1024 * 1024 // 10MB limit for YOLO
	}
	if size > limit {
		size = limit
	}

	data := make([]byte, size)
	_, err = io.ReadFull(f, data)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return fmt.Sprintf("error: %v", err)
	}

	lines := strings.Split(string(data), "\n")
	offset := 0
	if val, ok := args["offset"].(float64); ok {
		offset = int(val)
	}
	lcount := len(lines)
	if val, ok := args["limit"].(float64); ok {
		lcount = int(val)
	}

	end := offset + lcount
	if end > len(lines) {
		end = len(lines)
	}

	var sb strings.Builder
	for i := offset; i < end; i++ {
		sb.WriteString(fmt.Sprintf("%4d| %s\n", i+1, lines[i]))
	}

	res := sb.String()
	if stat.Size() > limit {
		res += fmt.Sprintf("\n(truncated; file size %d bytes exceeds %d byte limit)", stat.Size(), limit)
	}
	return res
}

func Write(args map[string]interface{}) string {
	path, _ := args["path"].(string)
	content, _ := args["content"].(string)
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return "ok"
}

func Edit(args map[string]interface{}) string {
	path, _ := args["path"].(string)
	old, _ := args["old"].(string)
	new, _ := args["new"].(string)
	all, _ := args["all"].(bool)

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	text := string(data)

	count := strings.Count(text, old)
	if count == 0 {
		return "error: old_string not found"
	}
	if !all && count > 1 {
		return fmt.Sprintf("error: old_string appears %d times, must be unique (use all=true)", count)
	}

	var replacement string
	if all {
		replacement = strings.ReplaceAll(text, old, new)
	} else {
		replacement = strings.Replace(text, old, new, 1)
	}

	err = os.WriteFile(path, []byte(replacement), 0644)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return "ok"
}

func Glob(args map[string]interface{}) string {
	pat, _ := args["pat"].(string)
	root := "."
	if val, ok := args["path"].(string); ok {
		root = val
	}

	limit := MaxOutputSize
	if config.YOLO {
		limit = 1 * 1024 * 1024 // 1MB for YOLO
	}

	var matches []string
	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		matched, _ := filepath.Match(pat, d.Name())
		if matched || pat == "**/*" || strings.Contains(pat, "*") {
			if matched {
				matches = append(matches, path)
			} else if strings.Contains(pat, "**") {
				matches = append(matches, path)
			}
		}
		return nil
	})

	sort.Slice(matches, func(i, j int) bool {
		fi, _ := os.Stat(matches[i])
		fj, _ := os.Stat(matches[j])
		if fi == nil || fj == nil {
			return false
		}
		return fi.ModTime().After(fj.ModTime())
	})

	if len(matches) == 0 {
		return "none"
	}
	res := strings.Join(matches, "\n")
	if len(res) > limit {
		res = res[:limit] + "\n(truncated)"
	}
	return res
}

func Grep(args map[string]interface{}) string {
	pat, _ := args["pat"].(string)
	root := "."
	if val, ok := args["path"].(string); ok {
		root = val
	}

	limit := MaxOutputSize
	if config.YOLO {
		limit = 1 * 1024 * 1024 // 1MB for YOLO
	}

	re, err := regexp.Compile(pat)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}

	var hits []string
	totalSize := 0
	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNum := 1
		for scanner.Scan() {
			line := scanner.Text()
			if re.MatchString(line) {
				hit := fmt.Sprintf("%s:%d:%s", path, lineNum, line)
				hits = append(hits, hit)
				totalSize += len(hit)
			}
			lineNum++
			if totalSize >= limit || len(hits) >= 1000 {
				return io.EOF
			}
		}
		return nil
	})

	if len(hits) == 0 {
		return "none"
	}
	res := strings.Join(hits, "\n")
	if totalSize >= limit {
		res += "\n(truncated)"
	}
	return res
}

func Bash(args map[string]interface{}) string {
	cmdStr, _ := args["cmd"].(string)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cwd, _ := os.Getwd()
	cmd := exec.CommandContext(ctx, "sh", "-c", cmdStr)
	cmd.Dir = cwd

	limit := MaxOutputSize
	if config.YOLO {
		cmd.Env = os.Environ()
		limit = 1 * 1024 * 1024 // 1MB for YOLO
	} else {
		cmd.Env = []string{
			"PATH=/usr/local/bin:/usr/bin:/bin",
			"HOME=" + cwd,
			"PWD=" + cwd,
		}
	}

	var outBuf bytes.Buffer
	stdout, _ := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		return fmt.Sprintf("error: %v", err)
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if outBuf.Len()+len(line) < limit {
			outBuf.WriteString(line + "\n")
		} else {
			outBuf.WriteString("\n(output truncated)")
			break
		}
	}

	err := cmd.Wait()
	if ctx.Err() == context.DeadlineExceeded {
		outBuf.WriteString("\n(timed out after 30s)")
	} else if err != nil {
		outBuf.WriteString(fmt.Sprintf("\nexit status: %v", err))
	}

	res := strings.TrimSpace(outBuf.String())
	if res == "" {
		return "(empty)"
	}
	return res
}

func Remember(args map[string]interface{}) string {
	action, _ := args["action"].(string)
	key, _ := args["key"].(string)
	value, _ := args["value"].(string)
	tagsStr, _ := args["tags"].(string)

	if len(value) > 200 {
		value = value[:200]
	}

	os.MkdirAll(".smallcode", 0755)
	path := ".smallcode/memory.json"

	var mem types.MemFile
	mem.Version = 1
	if data, err := os.ReadFile(path); err == nil {
		json.Unmarshal(data, &mem)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	mem.Updated = now

	switch action {
	case "add":
		if len(mem.Entries) >= 30 {
			return "warning: memory full (30 entries max). Use action=update to replace an existing entry or action=forget to remove a stale one."
		}
		id := fmt.Sprintf("m%d", len(mem.Entries)+1)
		var tags []string
		for _, t := range strings.Split(tagsStr, ",") {
			if t = strings.TrimSpace(t); t != "" {
				tags = append(tags, t)
			}
		}
		mem.Entries = append(mem.Entries, types.MemEntry{ID: id, Key: key, Value: value, Tags: tags, Added: now})
	case "update":
		for i, e := range mem.Entries {
			if e.Key == key {
				mem.Entries[i].Value = value
				break
			}
		}
	case "forget":
		filtered := mem.Entries[:0]
		for _, e := range mem.Entries {
			if e.Key != key {
				filtered = append(filtered, e)
			}
		}
		mem.Entries = filtered
	default:
		return "error: unknown action (use add, update, or forget)"
	}

	data, _ := json.MarshalIndent(mem, "", "  ")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return "ok"
}
