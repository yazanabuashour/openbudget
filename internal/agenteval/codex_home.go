package agenteval

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func CodexHome(runRoot string) string {
	return filepath.Join(runRoot, "codex-home")
}

func EnvWithCodexHome(env []string, runRoot string) []string {
	return EnvWithOverride(env, "CODEX_HOME", CodexHome(runRoot))
}

func EnvWithOverride(env []string, key string, value string) []string {
	prefix := key + "="
	out := make([]string, 0, len(env)+1)
	for _, entry := range env {
		if !strings.HasPrefix(entry, prefix) {
			out = append(out, entry)
		}
	}
	return append(out, prefix+value)
}

func SourceCodexHome() (string, error) {
	if home := os.Getenv("CODEX_HOME"); home != "" {
		return home, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".codex"), nil
}

func SetupCodexHome(runRoot string) error {
	sourceHome, err := SourceCodexHome()
	if err != nil {
		return err
	}
	return SetupCodexHomeFromSource(runRoot, sourceHome)
}

func SetupCodexHomeFromSource(runRoot string, sourceHome string) error {
	authPath := filepath.Join(sourceHome, "auth.json")
	authBytes, err := os.ReadFile(authPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("missing Codex auth at %s; run codex login before running evals", authPath)
		}
		return err
	}

	codexHome := CodexHome(runRoot)
	if err := os.RemoveAll(codexHome); err != nil {
		return err
	}
	if err := os.MkdirAll(codexHome, 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(codexHome, "auth.json"), authBytes, 0o600); err != nil {
		return err
	}
	return nil
}

func AddIgnoreUserConfig(args []string) []string {
	for _, arg := range args {
		if arg == "--ignore-user-config" {
			return args
		}
	}
	return append(args, "--ignore-user-config")
}

func CountNewSessionFiles(marker time.Time, runRoot string) int {
	sessionsDir := filepath.Join(CodexHome(runRoot), "sessions")
	count := 0
	_ = filepath.WalkDir(sessionsDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil || !info.ModTime().After(marker) {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		if strings.Contains(string(body), runRoot) {
			count++
		}
		return nil
	})
	return count
}
