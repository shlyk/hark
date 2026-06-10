// Package remote pushes notifications to an ntfy server (https://ntfy.sh),
// reaching the user's phone when they are away from the Mac.
package remote

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Client targets one ntfy topic.
type Client struct {
	Server string
	Topic  string
	HTTP   *http.Client
}

// Send publishes message to the topic. The title travels in the X-Title
// header per the ntfy protocol.
func (c Client) Send(title, message string) error {
	if c.Topic == "" {
		return fmt.Errorf("ntfy topic not configured (set ntfy.topic in the hark config)")
	}
	httpc := c.HTTP
	if httpc == nil {
		httpc = &http.Client{Timeout: 10 * time.Second}
	}
	url := strings.TrimRight(c.Server, "/") + "/" + c.Topic
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(message))
	if err != nil {
		return err
	}
	if title != "" {
		req.Header.Set("X-Title", title)
	}
	resp, err := httpc.Do(req)
	if err != nil {
		return fmt.Errorf("remote push failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("remote push failed: %s returned %s", url, resp.Status)
	}
	return nil
}
