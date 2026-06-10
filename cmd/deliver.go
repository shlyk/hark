package cmd

import (
	"github.com/shlyk/hark/internal/notify"

	"github.com/spf13/cobra"
)

// delivery is the post-flag notification pipeline shared by send and relay:
// banner, history record, optional speech.
type delivery struct {
	kind         string // history kind: "send" or "relay"
	notification notify.Notification
	speak        bool   // force speech
	smart        bool   // speech only with headphones
	once         string // dedupe key already checked by the caller
}

func deliver(cmd *cobra.Command, execer notify.Execer, d delivery) error {
	if err := notify.Send(execer, d.notification); err != nil {
		return err
	}
	// The banner is delivered at this point — record it even if the
	// optional speech below fails.
	record(cmd, d.kind, d.notification.Title, d.notification.Message, d.once)
	if d.speak || (d.smart && notify.HeadphonesConnected(execer)) {
		if err := notify.Say(execer, notify.Speech{Text: d.notification.Message}); err != nil {
			return err
		}
	}
	return nil
}
