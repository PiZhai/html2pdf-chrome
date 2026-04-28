package browser

import "os/exec"

type Instance struct {
	ExecPath     string
	UserDataDir  string
	DebugPort    int
	WebSocketURL string
	Cmd          *exec.Cmd
}
