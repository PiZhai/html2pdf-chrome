package browser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// LaunchOptions configures Chrome process startup behavior.
type LaunchOptions struct {
	// DebugLog enables Chrome stderr output to os.Stderr for debugging.
	DebugLog bool

	// NoSandbox disables Chrome's sandbox via --no-sandbox flag.
	// Required when running as root in Docker containers.
	NoSandbox bool
}

// Launch starts a new headless Chrome process with an isolated user-data-dir
// and a random debugging port. It waits up to 5 seconds for the WebSocket
// debugging endpoint to become available.
func Launch(execPath string, options LaunchOptions) (*Instance, error) {
	// Each launch uses a fresh profile directory to avoid state pollution
	// (cookies, localStorage, cache) between tasks.
	userDataDir, err := os.MkdirTemp("", "html2pdf-chrome-profile-*")
	if err != nil {
		return nil, fmt.Errorf("create user data dir: %w", err)
	}

	debugPort, err := pickFreePort()
	if err != nil {
		_ = os.RemoveAll(userDataDir)
		return nil, err
	}

	args := []string{
		"--headless=new",
		"--disable-gpu",
		"--no-first-run",
		"--no-default-browser-check",
		"--disable-background-networking",
		"--disable-background-timer-throttling",
		"--disable-default-apps",
		"--disable-sync",
		"--metrics-recording-only",
		"--mute-audio",
		"--hide-scrollbars",
		"--remote-debugging-port=" + strconv.Itoa(debugPort),
		"--user-data-dir=" + userDataDir,
	}

	if options.NoSandbox {
		args = append(args, "--no-sandbox")
	}

	cmd := exec.Command(execPath, args...)
	var stderr bytes.Buffer
	if options.DebugLog {
		cmd.Stdout = io.Discard
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	} else {
		cmd.Stdout = io.Discard
		cmd.Stderr = &stderr
	}

	if err := cmd.Start(); err != nil {
		_ = os.RemoveAll(userDataDir)

		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return nil, fmt.Errorf("start chrome: %w: %s", err, msg)
		}

		return nil, fmt.Errorf("start chrome: %w", err)
	}

	wsURL, err := waitForWebSocketURL(debugPort)
	if err != nil {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
		_ = os.RemoveAll(userDataDir)

		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return nil, fmt.Errorf("%w: %s", err, msg)
		}

		return nil, err
	}

	instance := &Instance{
		ExecPath:     execPath,
		UserDataDir:  userDataDir,
		DebugPort:    debugPort,
		Cmd:          cmd,
		WebSocketURL: wsURL,
	}

	return instance, nil
}

// Close terminates the Chrome process and removes its temporary user-data-dir.
func (i *Instance) Close() error {
	if i == nil {
		return nil
	}

	if i.Cmd != nil && i.Cmd.Process != nil {
		_ = i.Cmd.Process.Kill()
		_, _ = i.Cmd.Process.Wait()
	}

	if i.UserDataDir != "" {
		if err := os.RemoveAll(i.UserDataDir); err != nil {
			return fmt.Errorf("remove user data dir: %w", err)
		}
	}

	return nil
}

// pickFreePort finds an available TCP port on localhost. There is a small race
// window between releasing the port and Chrome binding to it, which is
// acceptable for this use case.
func pickFreePort() (int, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("pick free port: %w", err)
	}
	defer ln.Close()

	addr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		return 0, fmt.Errorf("pick free port: unexpected addr type %T", ln.Addr())
	}

	return addr.Port, nil
}

type versionInfo struct {
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

// waitForWebSocketURL polls Chrome's /json/version endpoint until the
// WebSocket debugger URL is available, with a 5-second deadline.
func waitForWebSocketURL(debugPort int) (string, error) {
	endpoint := "http://127.0.0.1:" + strconv.Itoa(debugPort) + "/json/version"
	deadline := time.Now().Add(5 * time.Second)

	for time.Now().Before(deadline) {
		resp, err := http.Get(endpoint)
		if err == nil {
			var info versionInfo
			decodeErr := json.NewDecoder(resp.Body).Decode(&info)
			_ = resp.Body.Close()

			if decodeErr == nil && info.WebSocketDebuggerURL != "" {
				return info.WebSocketDebuggerURL, nil
			}
		}

		time.Sleep(100 * time.Millisecond)
	}

	return "", fmt.Errorf("wait for Chrome debugging endpoint on port %d: timeout", debugPort)
}
