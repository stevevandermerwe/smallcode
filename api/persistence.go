package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// logTraffic writes raw API communication to .smallcode/trace.log
func (m *Model) logTraffic(direction string, data []byte) {
	os.MkdirAll(".smallcode", 0755)
	f, err := os.OpenFile(".smallcode/trace.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	var pretty bytes.Buffer
	json.Indent(&pretty, data, "", "  ")

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	separator := strings.Repeat("-", 80)
	entry := fmt.Sprintf("\n%s\n[%s] %s\n%s\n%s\n", separator, timestamp, direction, separator, pretty.String())

	f.WriteString(entry)
}

func (m *Model) saveSession() {
	os.MkdirAll(".smallcode", 0755)
	data, _ := json.Marshal(map[string]interface{}{
		"messages":     m.Model.Messages,
		"input_tokens": m.Model.TotalInputTokens,
		"out_tokens":   m.Model.TotalOutputTokens,
		"start_time":   m.Model.StartTime,
	})
	os.WriteFile(".smallcode/session.json", data, 0644)
}

func (m *Model) logReasoning(event string, data interface{}) {
	os.MkdirAll(".smallcode", 0755)
	f, err := os.OpenFile(".smallcode/reasoning.jsonl", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	entry := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"event":     event,
		"data":      data,
	}
	jsonBytes, _ := json.Marshal(entry)
	f.Write(append(jsonBytes, '\n'))
}
