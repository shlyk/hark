package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shlyk/hark/skill"

	"github.com/spf13/cobra"
)

// skillDirs maps supported agents to their skill directory relative to the
// install root (home directory, or the project directory with --project).
var skillDirs = map[string]string{
	"claude": filepath.Join(".claude", "skills", "hark"),
	"codex":  filepath.Join(".codex", "skills", "hark"),
}

func newSkillCmd() *cobra.Command {
	var project bool
	var agents []string
	cmd := &cobra.Command{
		Use:   "skill",
		Short: "Install the hark agent skill for Claude Code and Codex",
		Long: `skill writes the bundled SKILL.md into agent skill directories so AI
agents know when and how to notify you with hark.

By default it installs for all supported agents into your home directory
(~/.claude/skills/hark, ~/.codex/skills/hark). With --project it installs
into the current directory instead, so the skill can be committed with a
repository. Existing files are overwritten, so re-run "hark skill" after upgrading
hark.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.UserHomeDir()
			if project {
				root, err = os.Getwd()
			}
			if err != nil {
				return err
			}
			for _, agent := range agents {
				dir, ok := skillDirs[agent]
				if !ok {
					return fmt.Errorf("unknown agent %q (supported: claude, codex)", agent)
				}
				path := filepath.Join(root, dir, "SKILL.md")
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					return err
				}
				if err := os.WriteFile(path, skill.Content, 0o644); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "installed %s\n", path)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&project, "project", false, "install into the current project instead of the home directory")
	cmd.Flags().StringSliceVar(&agents, "agent", []string{"claude", "codex"}, "agents to install for (claude, codex)")
	return cmd
}
