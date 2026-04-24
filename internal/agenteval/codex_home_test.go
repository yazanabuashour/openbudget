package agenteval

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSetupCodexHomeCopiesOnlyAuth(t *testing.T) {
	t.Parallel()

	sourceHome := filepath.Join(t.TempDir(), "source-codex")
	if err := os.MkdirAll(filepath.Join(sourceHome, "sessions"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceHome, "auth.json"), []byte(`{"token":"secret"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceHome, "config.toml"), []byte("model = \"custom\""), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceHome, "sessions", "session.jsonl"), []byte("session"), 0o644); err != nil {
		t.Fatal(err)
	}

	runRoot := t.TempDir()
	if err := SetupCodexHomeFromSource(runRoot, sourceHome); err != nil {
		t.Fatalf("setup eval Codex home: %v", err)
	}

	codexHome := CodexHome(runRoot)
	authPath := filepath.Join(codexHome, "auth.json")
	authBytes, err := os.ReadFile(authPath)
	if err != nil {
		t.Fatalf("read copied auth: %v", err)
	}
	if string(authBytes) != `{"token":"secret"}` {
		t.Fatalf("auth content = %q, want copied source auth", authBytes)
	}

	info, err := os.Lstat(authPath)
	if err != nil {
		t.Fatalf("lstat auth: %v", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Fatal("auth copy must not be a symlink")
	}
	for _, unwanted := range []string{"config.toml", filepath.Join("sessions", "session.jsonl")} {
		if _, err := os.Stat(filepath.Join(codexHome, unwanted)); !os.IsNotExist(err) {
			t.Fatalf("unexpected copied %s: stat error = %v", unwanted, err)
		}
	}

	homeInfo, err := os.Stat(codexHome)
	if err != nil {
		t.Fatalf("stat eval codex home: %v", err)
	}
	if homeInfo.Mode().Perm()&0o077 != 0 {
		t.Fatalf("eval codex home permissions = %v, want no group/other access", homeInfo.Mode().Perm())
	}
}

func TestSetupCodexHomeRequiresAuth(t *testing.T) {
	t.Parallel()

	err := SetupCodexHomeFromSource(t.TempDir(), t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "run codex login") {
		t.Fatalf("setup eval Codex home error = %v, want login guidance", err)
	}
}

func TestEnvWithCodexHomeOverridesExistingValue(t *testing.T) {
	t.Parallel()

	runRoot := t.TempDir()
	env := EnvWithCodexHome([]string{"CODEX_HOME=/personal", "PATH=/bin"}, runRoot)
	joined := strings.Join(env, "\n")
	if strings.Contains(joined, "CODEX_HOME=/personal") {
		t.Fatalf("env kept personal CODEX_HOME: %v", env)
	}
	if !strings.Contains(joined, "CODEX_HOME="+CodexHome(runRoot)) {
		t.Fatalf("env missing isolated CODEX_HOME: %v", env)
	}
	if !strings.Contains(joined, "PATH=/bin") {
		t.Fatalf("env missing existing PATH: %v", env)
	}
}

func TestAddIgnoreUserConfig(t *testing.T) {
	t.Parallel()

	args := AddIgnoreUserConfig([]string{"exec", "--json"})
	if got := strings.Join(args, " "); got != "exec --json --ignore-user-config" {
		t.Fatalf("args = %q", got)
	}

	args = AddIgnoreUserConfig(args)
	if got := strings.Count(strings.Join(args, " "), "--ignore-user-config"); got != 1 {
		t.Fatalf("--ignore-user-config count = %d, want 1", got)
	}
}

func TestCountNewSessionFilesUsesEvalCodexHome(t *testing.T) {
	t.Parallel()

	runRoot := t.TempDir()
	sessionsDir := filepath.Join(CodexHome(runRoot), "sessions")
	if err := os.MkdirAll(sessionsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	marker := time.Now()
	oldPath := filepath.Join(sessionsDir, "old.jsonl")
	newPath := filepath.Join(sessionsDir, "new.jsonl")
	otherPath := filepath.Join(sessionsDir, "other.jsonl")
	for path, content := range map[string]string{
		oldPath:   runRoot,
		newPath:   runRoot,
		otherPath: "different run root",
	} {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Chtimes(oldPath, marker.Add(-time.Hour), marker.Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(newPath, marker.Add(time.Hour), marker.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(otherPath, marker.Add(time.Hour), marker.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}

	if got := CountNewSessionFiles(marker, runRoot); got != 1 {
		t.Fatalf("count new session files = %d, want 1", got)
	}
}
