package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shlyk/hark/skill"
)

func TestSkillInstallsUserSkillsForBothAgents(t *testing.T) {
	f := withFakes(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := run(t, f, "skill"); err != nil {
		t.Fatalf("skill error = %v", err)
	}
	for _, p := range []string{
		filepath.Join(home, ".claude", "skills", "hark", "SKILL.md"),
		filepath.Join(home, ".codex", "skills", "hark", "SKILL.md"),
	} {
		got, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("expected skill at %s: %v", p, err)
		}
		if string(got) != string(skill.Content) {
			t.Errorf("%s content differs from embedded skill", p)
		}
	}
}

func TestSkillProjectInstallsIntoCwd(t *testing.T) {
	f := withFakes(t)
	dir := t.TempDir()
	t.Chdir(dir)
	t.Setenv("HOME", t.TempDir())
	if err := run(t, f, "skill", "--project"); err != nil {
		t.Fatalf("skill --project error = %v", err)
	}
	for _, p := range []string{
		filepath.Join(dir, ".claude", "skills", "hark", "SKILL.md"),
		filepath.Join(dir, ".codex", "skills", "hark", "SKILL.md"),
	} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("expected skill at %s: %v", p, err)
		}
	}
}

func TestSkillAgentFilter(t *testing.T) {
	f := withFakes(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := run(t, f, "skill", "--agent", "claude"); err != nil {
		t.Fatalf("skill --agent claude error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".claude", "skills", "hark", "SKILL.md")); err != nil {
		t.Errorf("claude skill missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".codex")); !os.IsNotExist(err) {
		t.Errorf("codex dir should not exist, stat err = %v", err)
	}
}

func TestSkillRejectsUnknownAgent(t *testing.T) {
	f := withFakes(t)
	t.Setenv("HOME", t.TempDir())
	if err := run(t, f, "skill", "--agent", "cursor"); err == nil || !strings.Contains(err.Error(), "cursor") {
		t.Errorf("expected unknown-agent error naming cursor, got %v", err)
	}
}

func TestEmbeddedSkillHasFrontmatter(t *testing.T) {
	s := string(skill.Content)
	if !strings.HasPrefix(s, "---\nname: hark\n") || !strings.Contains(s, "description:") {
		t.Errorf("embedded skill missing required frontmatter:\n%.120s", s)
	}
}
