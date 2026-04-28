package browser

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// 只是编排查找顺序
func FindChrome(explicitPath string) (string, error) {

	if explicitPath != "" {
		return validateExecutable(explicitPath)
	}

	if envPath := os.Getenv("CHROME_PATH"); envPath != "" {
		return validateExecutable(envPath)
	}

	for _, path := range defaultExecutableCandidates() {
		if path == "" {
			continue
		}
		if executable, err := validateExecutable(path); err == nil {
			return executable, nil
		}
	}

	for _, binaryName := range binaryNames() {
		if resolvedPath, err := exec.LookPath(binaryName); err == nil {
			return resolvedPath, nil
		}
	}

	return "", errors.New("unable to find Chrome/Chromium executable; use -chrome-path or CHROME_PATH")
}

func validateExecutable(path string) (string, error) {
	// 检查这个路径在文件系统里是否存在，并拿到它的文件信息。
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("browser executable %q: %w", path, err)
	}

	if info.IsDir() {
		return "", fmt.Errorf("browser executable %q is a directory", path)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve browser executable %q: %w", path, err)
	}

	return absPath, nil
}

// 负责平台差异
func defaultExecutableCandidates() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
		}
	case "windows":
		return windowsCandidates()
	default:
		return []string{
			"/usr/bin/google-chrome",
			"/usr/bin/google-chrome-stable",
			"/usr/bin/chromium",
			"/usr/bin/chromium-browser",
			"/snap/bin/chromium",
		}
	}
}

// 负责 PATH 搜索名
func binaryNames() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{"chrome.exe"}
	default:
		return []string{"google-chrome", "google-chrome-stable", "chromium", "chromium-browser"}
	}
}

func windowsCandidates() []string {
	roots := []string{
		os.Getenv("ProgramFiles"),
		os.Getenv("ProgramFiles(x86)"),
		os.Getenv("LocalAppData"),
	}

	suffixes := []string{
		filepath.Join("Google", "Chrome", "Application", "chrome.exe"),
		filepath.Join("Chromium", "Application", "chrome.exe"),
	}

	candidates := make([]string, 0, len(roots)*len(suffixes))
	for _, root := range roots {
		if root == "" {
			continue
		}
		for _, suffix := range suffixes {
			candidates = append(candidates, filepath.Join(root, suffix))
		}
	}

	return candidates
}
