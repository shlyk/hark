package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/shlyk/hark/internal/history"

	"github.com/spf13/cobra"
)

// followInterval is the --follow poll period; tests shorten it.
var followInterval = time.Second

func newHistoryCmd() *cobra.Command {
	var limit int
	var asJSON, follow bool
	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show recently sent notifications",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if follow && asJSON {
				return fmt.Errorf("--follow cannot be combined with --json")
			}
			store, err := history.DefaultStore()
			if err != nil {
				return err
			}
			entries, err := store.Tail(0)
			if err != nil {
				return err
			}
			shown := entries
			if limit > 0 && len(shown) > limit {
				shown = shown[len(shown)-limit:]
			}
			if asJSON {
				if shown == nil {
					shown = []history.Entry{}
				}
				return json.NewEncoder(cmd.OutOrStdout()).Encode(shown)
			}
			for _, e := range shown {
				printEntry(cmd.OutOrStdout(), e)
			}
			if !follow {
				return nil
			}
			count := len(entries)
			ticker := time.NewTicker(followInterval)
			defer ticker.Stop()
			for {
				select {
				case <-cmd.Context().Done():
					return nil
				case <-ticker.C:
					all, err := store.Tail(0)
					if err != nil {
						continue
					}
					for _, e := range all[min(count, len(all)):] {
						printEntry(cmd.OutOrStdout(), e)
					}
					count = len(all)
				}
			}
		},
	}
	cmd.Flags().IntVarP(&limit, "limit", "n", 20, "max entries to show (0 = all)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "output a JSON array")
	cmd.Flags().BoolVar(&follow, "follow", false, "keep watching and print new notifications as they arrive")
	return cmd
}

func printEntry(w io.Writer, e history.Entry) {
	title := e.Title
	if title != "" {
		title = " " + title + ":"
	}
	fmt.Fprintf(w, "%s  [%s]%s %s\n",
		e.Time.Local().Format("2006-01-02 15:04:05"), e.Kind, title, e.Message)
}
