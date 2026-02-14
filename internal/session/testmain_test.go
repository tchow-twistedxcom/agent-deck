package session

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// skipIfNoTmuxServer skips the test if tmux binary is missing or server isn't running.
// This prevents test failures in CI environments where tmux is installed but no server runs.
func skipIfNoTmuxServer(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("tmux not available")
	}
	// Check if tmux server is running by trying to list sessions
	if err := exec.Command("tmux", "list-sessions").Run(); err != nil {
		t.Skip("tmux server not running")
	}
}

// createTestSessionFile creates a .jsonl file with conversation data for testing.
// Uses the current CLAUDE_CONFIG_DIR (or HOME-based default) to determine where to write.
func createTestSessionFile(t *testing.T, projectPath, sessionID string) {
	t.Helper()
	configDir := GetClaudeConfigDir()
	if configDir == "" {
		configDir = filepath.Join(os.Getenv("HOME"), ".claude")
	}
	encodedPath := ConvertToClaudeDirName(projectPath)
	if encodedPath == "" {
		encodedPath = "-"
	}
	projectDir := filepath.Join(configDir, "projects", encodedPath)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}
	data := []byte(`{"sessionId":"` + sessionID + `","type":"human"}`)
	if err := os.WriteFile(filepath.Join(projectDir, sessionID+".jsonl"), data, 0644); err != nil {
		t.Fatal(err)
	}
}

func TestMain(m *testing.M) {
	// Force test profile to prevent production data corruption
	// See CLAUDE.md: "2025-12-11 Incident: Tests with AGENTDECK_PROFILE=work overwrote ALL 36 production sessions"
	os.Setenv("AGENTDECK_PROFILE", "_test")

	// Run tests
	code := m.Run()

	// Cleanup: Kill any orphaned test sessions after tests complete
	// This prevents RAM waste from lingering test sessions
	// See CLAUDE.md: "2026-01-20 Incident: 20+ Test-Skip-Regen sessions orphaned, wasting ~3GB RAM"
	cleanupTestSessions()

	os.Exit(code)
}

// cleanupTestSessions kills any tmux sessions created during testing.
// IMPORTANT: Only match specific known test artifacts, NOT broad patterns.
// Broad patterns like HasPrefix("agentdeck_test") or Contains("test_") kill
// real user sessions with "test" in their title. Each test already has
// defer Kill() which handles cleanup reliably (runs on panic, Fatal, etc).
func cleanupTestSessions() {
	out, err := exec.Command("tmux", "list-sessions", "-F", "#{session_name}").Output()
	if err != nil {
		return
	}

	sessions := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, sess := range sessions {
		if strings.Contains(sess, "Test-Skip-Regen") {
			_ = exec.Command("tmux", "kill-session", "-t", sess).Run()
		}
	}
}
