package container

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const agentDebugLogPath = "/home/jimedrand/Git/reddock/.cursor/debug-b0e6fd.log"

// agentDebugLog appends one NDJSON line for debug-mode investigation (session b0e6fd).
func agentDebugLog(location, message, hypothesisID string, data map[string]any) {
	payload := map[string]any{
		"sessionId":    "b0e6fd",
		"location":     location,
		"message":      message,
		"hypothesisId": hypothesisID,
		"data":         data,
		"timestamp":    time.Now().UnixMilli(),
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return
	}
	if mkErr := os.MkdirAll(filepath.Dir(agentDebugLogPath), 0755); mkErr != nil {
		fmt.Fprintf(os.Stderr, "reddock debug log (session b0e6fd): cannot mkdir %q: %v\n", filepath.Dir(agentDebugLogPath), mkErr)
		return
	}
	f, err := os.OpenFile(agentDebugLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "reddock debug log (session b0e6fd): cannot open %q: %v\n", agentDebugLogPath, err)
		return
	}
	_, _ = f.Write(append(b, '\n'))
	_ = f.Close()
}
