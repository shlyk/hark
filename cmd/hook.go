package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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
				if settings == nil { // file contained literal null
					settings = map[string]any{}
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
			// Write atomically: settings.json must never be left truncated.
			tmp := path + ".tmp"
			if err := os.WriteFile(tmp, append(out, '\n'), 0o600); err != nil {
				return err
			}
			if err := os.Rename(tmp, path); err != nil {
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

// isRelayHook reports whether one inner hook runs exactly the relay command.
func isRelayHook(h any) bool {
	m, ok := h.(map[string]any)
	return ok && m["command"] == relayCommand
}

func innerHooks(entry any) []any {
	m, ok := entry.(map[string]any)
	if !ok {
		return nil
	}
	hs, _ := m["hooks"].([]any)
	return hs
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
		present := false
		for _, entry := range entries {
			for _, h := range innerHooks(entry) {
				if isRelayHook(h) {
					present = true
				}
			}
		}
		if !present {
			entries = append(entries, map[string]any{
				"hooks": []any{map[string]any{"type": "command", "command": relayCommand}},
			})
		}
		hooks[event] = entries
	}
}

// removeRelayHooks deletes only the relay hook itself; entry groups that
// also run other hooks are kept with the relay filtered out.
func removeRelayHooks(settings map[string]any) {
	hooks, _ := settings["hooks"].(map[string]any)
	if hooks == nil {
		return
	}
	for _, event := range hookEvents {
		entries, _ := hooks[event].([]any)
		var kept []any
		for _, entry := range entries {
			m, ok := entry.(map[string]any)
			if !ok {
				kept = append(kept, entry)
				continue
			}
			hs, _ := m["hooks"].([]any)
			var keptInner []any
			for _, h := range hs {
				if !isRelayHook(h) {
					keptInner = append(keptInner, h)
				}
			}
			if len(keptInner) == 0 && len(hs) > 0 {
				continue // group existed only for the relay hook
			}
			if len(keptInner) != len(hs) {
				m["hooks"] = keptInner
			}
			kept = append(kept, m)
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
