// Package config loads optional user defaults from
// ~/.config/hark/config.json ($XDG_CONFIG_HOME honored).
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds user defaults. Precedence at use sites is:
// explicit flag > config > built-in default.
type Config struct {
	Title    string   `json:"title,omitempty"`
	Smart    bool     `json:"smart,omitempty"`
	Sound    string   `json:"sound,omitempty"`
	Ntfy     Ntfy     `json:"ntfy,omitempty"`
	Escalate Escalate `json:"escalate,omitempty"`
}

// Ntfy configures the remote push channel. Remote push is enabled by
// setting a topic.
type Ntfy struct {
	Server string `json:"server,omitempty"`
	Topic  string `json:"topic,omitempty"`
}

func (n Ntfy) ServerOrDefault() string {
	if n.Server == "" {
		return "https://ntfy.sh"
	}
	return n.Server
}

// Escalate configures pushing to ntfy automatically when the user is away.
type Escalate struct {
	Enabled     bool `json:"enabled,omitempty"`
	IdleSeconds int  `json:"idle_seconds,omitempty"`
}

func (e Escalate) IdleOrDefault() int {
	if e.IdleSeconds <= 0 {
		return 300
	}
	return e.IdleSeconds
}

// Path returns the config file location; the file may not exist.
func Path() (string, error) {
	root := os.Getenv("XDG_CONFIG_HOME")
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		root = filepath.Join(home, ".config")
	}
	return filepath.Join(root, "hark", "config.json"), nil
}

// Load reads the config file. A missing file yields a zero Config; a
// malformed file is an error so typos are not silently ignored.
func Load() (Config, error) {
	path, err := Path()
	if err != nil {
		return Config{}, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Config{}, nil
	}
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing %s: %w", path, err)
	}
	return cfg, nil
}
