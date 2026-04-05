package memory

import (
	"encoding/json"
	"os"
	"time"

	"smallcode/types"
)

func WriteSessionSummary(msgs []types.Message, startTime time.Time) {
	if len(msgs) == 0 {
		return
	}
	os.MkdirAll(".smallcode", 0755)
	rec := types.SessionRecord{
		Session:     startTime.UTC().Format(time.RFC3339),
		DurationMin: int(time.Since(startTime).Minutes()),
		MsgCount:    len(msgs),
	}
	data, _ := json.Marshal(rec)
	f, err := os.OpenFile(".smallcode/sessions.jsonl", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	f.Write(append(data, '\n'))
}
