package cmd

import (
	"fmt"
	"os"
	"time"

	"hark/internal/history"
)

// record appends to history; failures warn but never fail the command.
func record(kind, title, msg string) {
	store, err := history.DefaultStore()
	if err == nil {
		err = store.Append(history.Entry{Time: time.Now(), Kind: kind, Title: title, Message: msg})
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not record history: %v\n", err)
	}
}
