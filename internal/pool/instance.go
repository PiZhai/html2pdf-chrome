package pool

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/PiZhai/html2pdf-chrome/internal/browser"
)

// PooledInstance wraps a browser.Instance with pool-specific metadata.
type PooledInstance struct {
	instance   *browser.Instance
	createdAt  time.Time
	lastUsedAt time.Time
	taskCount  int
}

// IsHealthy checks whether the underlying Chrome process is still alive and
// responsive by hitting the /json/version debugging endpoint.
func (pi *PooledInstance) IsHealthy() bool {
	if pi.instance == nil || pi.instance.Cmd == nil || pi.instance.Cmd.Process == nil {
		return false
	}

	// Check if process has already exited.
	// On Unix, Process.Signal(0) returns nil if the process is alive.
	if pi.instance.Cmd.ProcessState != nil {
		return false
	}

	endpoint := "http://127.0.0.1:" + strconv.Itoa(pi.instance.DebugPort) + "/json/version"
	client := &http.Client{Timeout: 2 * time.Second}

	resp, err := client.Get(endpoint)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	var info struct {
		WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return false
	}

	return info.WebSocketDebuggerURL != ""
}

// Close shuts down the underlying Chrome instance.
func (pi *PooledInstance) Close() error {
	if pi.instance == nil {
		return nil
	}
	return pi.instance.Close()
}

// WebSocketURL returns the debugging WebSocket URL for CDP connections.
func (pi *PooledInstance) WebSocketURL() string {
	if pi.instance == nil {
		return ""
	}
	return pi.instance.WebSocketURL
}

// String returns a human-readable description for logging.
func (pi *PooledInstance) String() string {
	if pi.instance == nil {
		return "PooledInstance{nil}"
	}
	return fmt.Sprintf("PooledInstance{port=%d, tasks=%d}", pi.instance.DebugPort, pi.taskCount)
}
