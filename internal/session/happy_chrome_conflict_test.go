package session

import (
	"strings"
	"testing"
)

// happy (happy-coder) re-parses claude's flags through its own commander layer
// and rejects --chrome with "error: unknown option '--chrome'", so a session
// configured with both the happy wrapper and --chrome dies on start. These
// tests pin the detection helper and the command-builder safety net that keep
// the two from producing a broken spawn command.

func TestClaudeOptionsConflict(t *testing.T) {
	cases := []struct {
		name      string
		opts      *ClaudeOptions
		extraArgs []string
		want      bool
	}{
		{"nil opts", nil, nil, false},
		{"happy+chrome", &ClaudeOptions{UseHappy: true, UseChrome: true}, nil, true},
		{"happy only", &ClaudeOptions{UseHappy: true}, nil, false},
		{"chrome only (no happy)", &ClaudeOptions{UseChrome: true}, nil, false},
		{"happy + --chrome extra arg", &ClaudeOptions{UseHappy: true}, []string{"--verbose", "--chrome"}, true},
		{"chrome extra arg without happy", &ClaudeOptions{}, []string{"--chrome"}, false},
		{"happy + unrelated extra arg", &ClaudeOptions{UseHappy: true}, []string{"--model", "opus"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ClaudeOptionsConflict(tc.opts, tc.extraArgs); got != tc.want {
				t.Fatalf("ClaudeOptionsConflict(%+v, %v) = %v, want %v", tc.opts, tc.extraArgs, got, tc.want)
			}
		})
	}
}

// TestBuildClaudeExtraFlags_HappyDropsChromeFlag: with the happy wrapper,
// --chrome must NOT be emitted from the UseChrome option (it would crash happy).
func TestBuildClaudeExtraFlags_HappyDropsChromeFlag(t *testing.T) {
	channelsTestEnv(t)

	inst := NewInstanceWithTool("happy-chrome", t.TempDir(), "claude")
	flags := inst.buildClaudeExtraFlags(&ClaudeOptions{UseHappy: true, UseChrome: true})

	if strings.Contains(flags, "--chrome") {
		t.Fatalf("happy session must not emit --chrome (crashes happy); got:\n%s", flags)
	}
}

// TestBuildClaudeExtraFlags_NonHappyKeepsChrome: without happy, UseChrome still
// emits --chrome (plain claude accepts it).
func TestBuildClaudeExtraFlags_NonHappyKeepsChrome(t *testing.T) {
	channelsTestEnv(t)

	inst := NewInstanceWithTool("claude-chrome", t.TempDir(), "claude")
	flags := inst.buildClaudeExtraFlags(&ClaudeOptions{UseChrome: true})

	if !strings.Contains(flags, "--chrome") {
		t.Fatalf("non-happy session with UseChrome must emit --chrome; got:\n%s", flags)
	}
}

// TestBuildClaudeExtraFlags_HappyDropsChromeFromExtraArgs: the config.toml
// [claude] extra_args default is ["--chrome"], so a happy session can carry
// --chrome via ExtraArgs even when UseChrome is false. It must be stripped.
func TestBuildClaudeExtraFlags_HappyDropsChromeFromExtraArgs(t *testing.T) {
	channelsTestEnv(t)

	inst := NewInstanceWithTool("happy-extra-chrome", t.TempDir(), "claude")
	inst.ExtraArgs = []string{"--chrome"}
	flags := inst.buildClaudeExtraFlags(&ClaudeOptions{UseHappy: true})
	if strings.Contains(flags, "--chrome") {
		t.Fatalf("happy session must strip --chrome from extra args; got:\n%s", flags)
	}

	// Control: without happy, the same extra arg survives.
	inst2 := NewInstanceWithTool("claude-extra-chrome", t.TempDir(), "claude")
	inst2.ExtraArgs = []string{"--chrome"}
	flags2 := inst2.buildClaudeExtraFlags(&ClaudeOptions{})
	if !strings.Contains(flags2, "--chrome") {
		t.Fatalf("non-happy session must keep --chrome in extra args; got:\n%s", flags2)
	}
}
