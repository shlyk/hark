package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const relayCommand = "hark relay claude"

// hookEvents are the Claude Code hook events hark attaches to.
var hookEvents = []string{"Notification", "Stop"}

func newHookCmd() *cobra.Command {
	var remove bool
	cmd := &cobra.Command{
		Use:   "hook claude",
		Short: "Wire hark into Claude Code hooks (notify on attention/finish)",
		Long: `hook adds "` + relayCommand + `" to the Notification and Stop hooks in
~/.claude/settings.json, so every Claude Code session pings you when it
needs attention or finishes. The previous settings file is backed up to
settings.json.bak. Idempotent; --remove undoes it.`,
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgs: []string{"claude"},
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			path := filepath.Join(home, ".claude", "settings.json")
			settings := map[string]any{}
			data, err := os.ReadFile(path)
			switch {
			case err == nil:
				if err := json.Unmarshal(data, &settings); err != nil {
					return fmt.Errorf("parsing %s: %w", path, err)
				}
				if err := os.WriteFile(path+".bak", data, 0o600); err != nil {
					return fmt.Errorf("backing up settings: %w", err)
				}
			case os.IsNotExist(err):
				if remove {
					return nil // nothing to remove
				}
			default:
				return err
			}

			if remove {
				removeRelayHooks(settings)
			} else {
				addRelayHooks(settings)
			}

			out, err := json.MarshalIndent(settings, "", "  ")
			if err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(path, append(out, '\n'), 0o600); err != nil {
				return err
			}
			verb := "installed in"
			if remove {
				verb = "removed from"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "hark hooks %s %s\n", verb, path)
			return nil
		},
	}
	cmd.Flags().BoolVar(&remove, "remove", false, "remove the hark hooks instead of adding them")
	return cmd
}

// addRelayHooks idempotently merges the relay command into each hook event.
func addRelayHooks(settings map[string]any) {
	hooks, _ := settings["hooks"].(map[string]any)
	if hooks == nil {
		hooks = map[string]any{}
		settings["hooks"] = hooks
	}
	for _, event := range hookEvents {
		entries, _ := hooks[event].([]any)
		if !containsRelay(entries) {
			entries = append(entries, map[string]any{
				"hooks": []any{map[string]any{"type": "command", "command": relayCommand}},
			})
		}
		hooks[event] = entries
	}
}

func removeRelayHooks(settings map[string]any) {
	hooks, _ := settings["hooks"].(map[string]any)
	if hooks == nil {
		return
	}
	for _, event := range hookEvents {
		entries, _ := hooks[event].([]any)
		var kept []any
		for _, e := range entries {
			if !containsRelay([]any{e}) {
				kept = append(kept, e)
			}
		}
		if len(kept) == 0 {
			delete(hooks, event)
		} else {
			hooks[event] = kept
		}
	}
	if len(hooks) == 0 {
		delete(settings, "hooks")
	}
}

// containsRelay reports whether any hook entry runs the relay command.
func containsRelay(entries []any) bool {
	raw, err := json.Marshal(entries)
	return err == nil && strings.Contains(string(raw), relayCommand)
}
