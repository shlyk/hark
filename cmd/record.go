package cmd

import (
	"fmt"
	"time"

	"github.com/shlyk/hark/internal/history"

	"github.com/spf13/cobra"
)

// record appends to history; failures warn but never fail the command.
func record(cmd *cobra.Command, kind, title, msg string) {
	store, err := history.DefaultStore()
	if err == nil {
		err = store.Append(history.Entry{Time: time.Now(), Kind: kind, Title: title, Message: msg})
	}
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "warning: could not record history: %v\n", err)
	}
}
