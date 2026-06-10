//go:build !windows
// +build !windows

package tmux

import (
	"os/exec"
	"strings"
	"testing"
)

// Regression tests for the socket-isolation gate on the tmux-level C-q
// detach binding (6bd1ec93). The root-table bind is server-global: installed
// on a shared default server it eats Ctrl+Q from EVERY session, including
// the user's own wrapper session when agent-deck runs nested inside tmux,
// consuming the byte before the PTY attach loop's IndexDetachKey handler
// can see it. Session.Start therefore installs the bind ONLY when
// SocketName is set, i.e. when the session lives on a private -L server
// where every session is agent-deck-managed. These tests fail if a future
// merge drops the `if s.SocketName != ""` guard around the bind-key call.

// rootKeysOutput returns the root key table of the given server, or "" when
// no server is running on that socket (no server means no bindings).
func rootKeysOutput(t *testing.T, socketName string) string {
	t.Helper()
	out, err := Exec(socketName, "list-keys", "-T", "root").CombinedOutput()
	if err != nil {
		return ""
	}
	return string(out)
}

func TestStart_CQDetachBind_SharedDefaultServer_OmitsBind(t *testing.T) {
	requireTmux(t)

	s := NewSession("cq-gate-shared-"+randomServerSuffix(t), "/tmp")
	t.Cleanup(func() {
		_ = exec.Command("tmux", "kill-session", "-t", s.Name).Run()
	})

	if err := s.Start(""); err != nil {
		t.Fatalf("Start on default server failed: %v", err)
	}

	if keys := rootKeysOutput(t, ""); strings.Contains(keys, "C-q") {
		t.Fatalf("Start with empty SocketName must NOT install the root C-q bind on the shared default server (it would intercept Ctrl+Q for all sessions, breaking detach for nested agent-deck); list-keys -T root:\n%s", keys)
	}
}

func TestStart_CQDetachBind_IsolatedSocket_InstallsBind(t *testing.T) {
	requireTmux(t)

	const socket = "adcqgate"
	s := NewSession("cq-gate-iso-"+randomServerSuffix(t), "/tmp")
	s.SocketName = socket
	t.Cleanup(func() {
		_ = exec.Command("tmux", "-L", socket, "kill-server").Run()
	})

	if err := s.Start(""); err != nil {
		t.Fatalf("Start on isolated -L server failed: %v", err)
	}

	keys := rootKeysOutput(t, socket)
	if !strings.Contains(keys, "C-q") || !strings.Contains(keys, "detach-client") {
		t.Fatalf("Start with SocketName set must install the root C-q detach bind on the isolated server (iTerm2 flow-control fallback); list-keys -T root:\n%s", keys)
	}

	// The bind must stay scoped to the private server: the shared default
	// server's root table must remain free of C-q.
	if shared := rootKeysOutput(t, ""); strings.Contains(shared, "C-q") {
		t.Fatalf("isolated-socket bind leaked to the shared default server; list-keys -T root:\n%s", shared)
	}
}
