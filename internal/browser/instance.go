package browser

import "os/exec"

// Instance represents a running Chrome/Chromium process with its associated
// metadata needed for CDP communication and cleanup.
type Instance struct {
	ExecPath     string    // Path to the Chrome executable
	UserDataDir  string    // Temporary profile directory (cleaned up on Close)
	DebugPort    int       // CDP debugging port
	WebSocketURL string    // WebSocket URL for CDP connections
	Cmd          *exec.Cmd // Underlying OS process
}
