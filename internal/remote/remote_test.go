package remote

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendPostsToTopic(t *testing.T) {
	var gotPath, gotTitle, gotBody string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotTitle = r.Header.Get("X-Title")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
	}))
	defer ts.Close()

	c := Client{Server: ts.URL, Topic: "my-topic"}
	if err := c.Send("CI", "build done"); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if gotPath != "/my-topic" || gotTitle != "CI" || gotBody != "build done" {
		t.Errorf("got path=%q title=%q body=%q", gotPath, gotTitle, gotBody)
	}
}

func TestSendFailsOnServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusForbidden)
	}))
	defer ts.Close()
	c := Client{Server: ts.URL, Topic: "t"}
	if err := c.Send("", "msg"); err == nil {
		t.Error("Send() should fail on HTTP 403")
	}
}

func TestSendFailsOnConnectionError(t *testing.T) {
	c := Client{Server: "http://127.0.0.1:1", Topic: "t"}
	if err := c.Send("", "msg"); err == nil {
		t.Error("Send() should fail when server is unreachable")
	}
}

func TestSendRequiresTopic(t *testing.T) {
	c := Client{Server: "https://ntfy.sh"}
	if err := c.Send("", "msg"); err == nil {
		t.Error("Send() without topic should fail")
	}
}
