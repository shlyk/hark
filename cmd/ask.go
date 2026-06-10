package cmd

import (
	"fmt"
	"strings"

	"github.com/shlyk/hark/internal/notify"

	"github.com/spf13/cobra"
)

func newAskCmd(execer notify.Execer) *cobra.Command {
	var title, options string
	var input bool
	var timeout int
	cmd := &cobra.Command{
		Use:   "ask <question>",
		Short: "Ask the user a question via a dialog and print the answer",
		Long: `ask blocks until the user answers and prints the answer to stdout, so
agents can wait for a decision without the user returning to the terminal.

  hark ask "Deploy now?"                          -> Yes / No
  hark ask "Which strategy?" --options A,B,C      -> chosen option
  hark ask "Ticket number?" --input               -> typed text

Cancel or timeout exits non-zero. --timeout does not apply to --options
(AppleScript list dialogs cannot time out).`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			q := notify.Question{
				Prompt:  strings.Join(args, " "),
				Title:   title,
				Input:   input,
				Timeout: timeout,
			}
			if options != "" {
				q.Options = strings.Split(options, ",")
			}
			answer, err := notify.Ask(execer, q)
			if err != nil {
				return err
			}
			record(cmd, "ask", title, q.Prompt+" -> "+answer)
			fmt.Fprintln(cmd.OutOrStdout(), answer)
			return nil
		},
	}
	cmd.Flags().StringVarP(&title, "title", "t", "hark", "dialog title")
	cmd.Flags().StringVar(&options, "options", "", "comma-separated choices shown as a list")
	cmd.Flags().BoolVar(&input, "input", false, "ask for free text instead of buttons")
	cmd.Flags().IntVar(&timeout, "timeout", 0, "give up after N seconds (not with --options)")
	return cmd
}
