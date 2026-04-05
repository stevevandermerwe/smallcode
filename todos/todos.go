package todos

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"smallcode/types"
)

const todoFile = ".smallcode/todos.json"

func Execute(args map[string]interface{}) string {
	action, ok := args["action"].(string)
	if !ok {
		return "error: action required"
	}

	switch action {
	case "add":
		return add(args)
	case "update":
		return update(args)
	case "close":
		return closeTask(args)
	case "reopen":
		return reopen(args)
	case "remove":
		return remove(args)
	case "list":
		return list(args)
	default:
		return fmt.Sprintf("error: unknown action %s", action)
	}
}

func add(args map[string]interface{}) string {
	title, ok := args["title"].(string)
	if !ok || title == "" {
		return "error: title required"
	}

	tf, err := loadTodos()
	if err != nil {
		tf = &types.TodoFile{Version: 1, Todos: []types.Todo{}}
	}

	// Generate next ID
	nextID := generateNextID(tf.Todos)

	// Parse optional fields
	priority := 0
	if p, ok := args["priority"].(float64); ok {
		priority = int(p)
	}

	var blockedBy []string
	if bb, ok := args["blocked_by"].(string); ok && bb != "" {
		blockedBy = strings.Split(bb, ",")
		for i := range blockedBy {
			blockedBy[i] = strings.TrimSpace(blockedBy[i])
		}
	}

	var sources []types.Source
	if src, ok := args["sources"].(string); ok && src != "" {
		sources = parseSources(src)
	}

	todo := types.Todo{
		ID:        nextID,
		Title:     title,
		Status:    "pending",
		Priority:  priority,
		BlockedBy: blockedBy,
		Sources:   sources,
		Created:   time.Now().Format(time.RFC3339),
	}

	tf.Todos = append(tf.Todos, todo)
	tf.Updated = time.Now().Format(time.RFC3339)

	if err := saveTodos(tf); err != nil {
		return fmt.Sprintf("error: %v", err)
	}

	return fmt.Sprintf("added todo %s: %s", nextID, title)
}

func update(args map[string]interface{}) string {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return "error: id required"
	}

	tf, err := loadTodos()
	if err != nil {
		return "error: no todos file"
	}

	idx := -1
	for i, t := range tf.Todos {
		if t.ID == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Sprintf("error: todo %s not found", id)
	}

	// Update fields if provided
	if title, ok := args["title"].(string); ok && title != "" {
		tf.Todos[idx].Title = title
	}

	if p, ok := args["priority"].(float64); ok {
		tf.Todos[idx].Priority = int(p)
	}

	if status, ok := args["status"].(string); ok {
		if status == "pending" || status == "active" || status == "done" {
			tf.Todos[idx].Status = status
		} else {
			return fmt.Sprintf("error: invalid status %s", status)
		}
	}

	if bb, ok := args["blocked_by"].(string); ok {
		if bb == "" {
			tf.Todos[idx].BlockedBy = []string{}
		} else {
			parts := strings.Split(bb, ",")
			for i := range parts {
				parts[i] = strings.TrimSpace(parts[i])
			}
			tf.Todos[idx].BlockedBy = parts
		}
	}

	if src, ok := args["sources"].(string); ok {
		if src == "" {
			tf.Todos[idx].Sources = []types.Source{}
		} else {
			tf.Todos[idx].Sources = parseSources(src)
		}
	}

	tf.Updated = time.Now().Format(time.RFC3339)

	if err := saveTodos(tf); err != nil {
		return fmt.Sprintf("error: %v", err)
	}

	return fmt.Sprintf("updated todo %s", id)
}

func closeTask(args map[string]interface{}) string {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return "error: id required"
	}

	tf, err := loadTodos()
	if err != nil {
		return "error: no todos file"
	}

	idx := -1
	for i, t := range tf.Todos {
		if t.ID == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Sprintf("error: todo %s not found", id)
	}

	tf.Todos[idx].Status = "done"
	tf.Todos[idx].Closed = time.Now().Format(time.RFC3339)
	tf.Updated = time.Now().Format(time.RFC3339)

	if err := saveTodos(tf); err != nil {
		return fmt.Sprintf("error: %v", err)
	}

	return fmt.Sprintf("closed todo %s", id)
}

func reopen(args map[string]interface{}) string {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return "error: id required"
	}

	tf, err := loadTodos()
	if err != nil {
		return "error: no todos file"
	}

	idx := -1
	for i, t := range tf.Todos {
		if t.ID == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Sprintf("error: todo %s not found", id)
	}

	tf.Todos[idx].Status = "pending"
	tf.Todos[idx].Closed = ""
	tf.Updated = time.Now().Format(time.RFC3339)

	if err := saveTodos(tf); err != nil {
		return fmt.Sprintf("error: %v", err)
	}

	return fmt.Sprintf("reopened todo %s", id)
}

func remove(args map[string]interface{}) string {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return "error: id required"
	}

	tf, err := loadTodos()
	if err != nil {
		return "error: no todos file"
	}

	idx := -1
	for i, t := range tf.Todos {
		if t.ID == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Sprintf("error: todo %s not found", id)
	}

	// Remove from all blocked_by lists (cascade)
	for i := range tf.Todos {
		var cleaned []string
		for _, blocker := range tf.Todos[i].BlockedBy {
			if blocker != id {
				cleaned = append(cleaned, blocker)
			}
		}
		tf.Todos[i].BlockedBy = cleaned
	}

	// Remove the todo
	tf.Todos = append(tf.Todos[:idx], tf.Todos[idx+1:]...)
	tf.Updated = time.Now().Format(time.RFC3339)

	if err := saveTodos(tf); err != nil {
		return fmt.Sprintf("error: %v", err)
	}

	return fmt.Sprintf("removed todo %s", id)
}

func list(args map[string]interface{}) string {
	tf, err := loadTodos()
	if err != nil {
		return "no todos yet"
	}

	if len(tf.Todos) == 0 {
		return "no todos"
	}

	// Group by computed state
	var active, ready, blocked, done []types.Todo

	for _, t := range tf.Todos {
		state := computeState(t, tf.Todos)
		switch state {
		case "active":
			active = append(active, t)
		case "ready":
			ready = append(ready, t)
		case "blocked":
			blocked = append(blocked, t)
		case "done":
			done = append(done, t)
		}
	}

	var sb strings.Builder
	sb.WriteString("## Todos\n")

	if len(active) > 0 {
		sb.WriteString("\nActive:\n")
		for _, t := range active {
			sb.WriteString(formatTodo(t))
		}
	}

	if len(ready) > 0 {
		sb.WriteString("\nReady:\n")
		for _, t := range ready {
			sb.WriteString(formatTodo(t))
		}
	}

	if len(blocked) > 0 {
		sb.WriteString("\nBlocked:\n")
		for _, t := range blocked {
			sb.WriteString(formatTodo(t))
		}
	}

	if len(done) > 0 {
		sb.WriteString("\nDone:\n")
		for _, t := range done {
			sb.WriteString(formatTodo(t))
		}
	}

	return sb.String()
}

// Helpers

func loadTodos() (*types.TodoFile, error) {
	data, err := os.ReadFile(todoFile)
	if err != nil {
		return nil, err
	}

	var tf types.TodoFile
	if err := json.Unmarshal(data, &tf); err != nil {
		return nil, err
	}
	return &tf, nil
}

func saveTodos(tf *types.TodoFile) error {
	data, err := json.MarshalIndent(tf, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(todoFile, data, 0644)
}

func generateNextID(todos []types.Todo) string {
	maxNum := 0
	for _, t := range todos {
		if strings.HasPrefix(t.ID, "t") {
			var num int
			fmt.Sscanf(t.ID, "t%d", &num)
			if num > maxNum {
				maxNum = num
			}
		}
	}
	return fmt.Sprintf("t%d", maxNum+1)
}

func computeState(todo types.Todo, allTodos []types.Todo) string {
	if todo.Status == "done" {
		return "done"
	}
	if todo.Status == "active" {
		return "active"
	}
	// status == "pending"
	if len(todo.BlockedBy) == 0 {
		return "ready"
	}

	for _, blockerID := range todo.BlockedBy {
		for _, other := range allTodos {
			if other.ID == blockerID && other.Status != "done" {
				return "blocked"
			}
		}
	}
	return "ready"
}

func formatTodo(t types.Todo) string {
	var line strings.Builder
	line.WriteString(fmt.Sprintf("- [%s]", t.ID))

	if t.Priority > 0 {
		line.WriteString(fmt.Sprintf(" P%d", t.Priority))
	}

	line.WriteString(fmt.Sprintf(" %s", t.Title))

	if len(t.BlockedBy) > 0 {
		line.WriteString(fmt.Sprintf(" (blocked by %s)", strings.Join(t.BlockedBy, ", ")))
	}

	if len(t.Sources) > 0 {
		var srcs []string
		for _, s := range t.Sources {
			srcs = append(srcs, s.Ref)
		}
		line.WriteString(fmt.Sprintf("\n  sources: %s", strings.Join(srcs, ", ")))
	}

	line.WriteString("\n")
	return line.String()
}

func parseSources(src string) []types.Source {
	var sources []types.Source
	parts := strings.Split(src, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, ":") {
			kv := strings.SplitN(part, ":", 2)
			sources = append(sources, types.Source{
				Type: kv[0],
				Ref:  kv[1],
			})
		}
	}
	return sources
}

// FormatForPrompt returns a compact grouped view for system prompt injection
func FormatForPrompt(todos []types.Todo) string {
	if len(todos) == 0 {
		return ""
	}

	// Group by computed state
	byState := make(map[string][]types.Todo)
	for _, t := range todos {
		if t.Status != "done" {
			state := computeState(t, todos)
			byState[state] = append(byState[state], t)
		}
	}

	// Sort each group by priority (descending) then by title
	for state := range byState {
		sort.Slice(byState[state], func(i, j int) bool {
			if byState[state][i].Priority != byState[state][j].Priority {
				return byState[state][i].Priority < byState[state][j].Priority // 1 comes before 2
			}
			return byState[state][i].Title < byState[state][j].Title
		})
	}

	var sb strings.Builder
	sb.WriteString("## Todos\n")

	for _, state := range []string{"active", "ready", "blocked"} {
		if todos, ok := byState[state]; ok && len(todos) > 0 {
			sb.WriteString(fmt.Sprintf("%s:\n", strings.Title(state)))
			for _, t := range todos {
				sb.WriteString(fmt.Sprintf("- [%s]", t.ID))
				if t.Priority > 0 {
					sb.WriteString(fmt.Sprintf(" P%d", t.Priority))
				}
				sb.WriteString(fmt.Sprintf(" %s", t.Title))
				if len(t.BlockedBy) > 0 {
					sb.WriteString(fmt.Sprintf(" (blocked by %s)", strings.Join(t.BlockedBy, ", ")))
				}
				sb.WriteString("\n")
			}
		}
	}

	return sb.String()
}
