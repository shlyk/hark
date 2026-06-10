package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingFileIsZeroConfig(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() with no file error = %v", err)
	}
	if cfg != (Config{}) {
		t.Errorf("Load() = %+v, want zero config", cfg)
	}
}

func TestLoadReadsFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	path := filepath.Join(dir, "hark", "config.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	data := `{"title":"agent","smart":true,"sound":"Glass","ntfy":{"topic":"my-topic"},"escalate":{"enabled":true,"idle_seconds":120}}`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Title != "agent" || !cfg.Smart || cfg.Sound != "Glass" ||
		cfg.Ntfy.Topic != "my-topic" || !cfg.Escalate.Enabled || cfg.Escalate.IdleSeconds != 120 {
		t.Errorf("Load() = %+v", cfg)
	}
}

func TestLoadRejectsInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	path := filepath.Join(dir, "hark", "config.json")
	os.MkdirAll(filepath.Dir(path), 0o755)
	os.WriteFile(path, []byte("{broken"), 0o600)
	if _, err := Load(); err == nil {
		t.Error("Load() with invalid JSON should error, not silently ignore the config")
	}
}

func TestNtfyServerDefault(t *testing.T) {
	if got := (Ntfy{}).ServerOrDefault(); got != "https://ntfy.sh" {
		t.Errorf("ServerOrDefault() = %q", got)
	}
	if got := (Ntfy{Server: "https://my.host"}).ServerOrDefault(); got != "https://my.host" {
		t.Errorf("ServerOrDefault() = %q", got)
	}
}

func TestEscalateIdleDefault(t *testing.T) {
	if got := (Escalate{}).IdleOrDefault(); got != 300 {
		t.Errorf("IdleOrDefault() = %d, want 300", got)
	}
	if got := (Escalate{IdleSeconds: 60}).IdleOrDefault(); got != 60 {
		t.Errorf("IdleOrDefault() = %d, want 60", got)
	}
}
