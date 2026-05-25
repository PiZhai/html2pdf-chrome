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

type LaunchOptions struct {
	DebugLog  bool
	NoSandbox bool
}

func Launch(execPath string, options LaunchOptions) (*Instance, error) {
	// 为每次启动创建独立 profile，避免缓存、cookie、localStorage 污染任务。
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

// pickFreePort() 有竞态窗口
// 这是正常的第一版实现。
// 意思是：你拿到端口后关闭 listener，到 Chrome 真正绑定前，理论上可能被别的进程抢走。第一版可以接受，后面如果真要增强稳定性，再考虑更稳的做法。
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
