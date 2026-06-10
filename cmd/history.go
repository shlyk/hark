package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/shlyk/hark/internal/history"

	"github.com/spf13/cobra"
)

func newHistoryCmd() *cobra.Command {
	var limit int
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show recently sent notifications",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := history.DefaultStore()
			if err != nil {
				return err
			}
			entries, err := store.Tail(limit)
			if err != nil {
				return err
			}
			if asJSON {
				if entries == nil {
					entries = []history.Entry{}
				}
				return json.NewEncoder(cmd.OutOrStdout()).Encode(entries)
			}
			for _, e := range entries {
				title := e.Title
				if title != "" {
					title = " " + title + ":"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s  [%s]%s %s\n",
					e.Time.Local().Format("2006-01-02 15:04:05"), e.Kind, title, e.Message)
			}
			return nil
		},
	}
	cmd.Flags().IntVarP(&limit, "limit", "n", 20, "max entries to show")
	cmd.Flags().BoolVar(&asJSON, "json", false, "output a JSON array")
	return cmd
}
