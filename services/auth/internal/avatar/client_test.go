package avatar

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("seed") != "ada" {
			t.Fatalf("seed = %q", r.URL.Query().Get("seed"))
		}
		_, _ = w.Write([]byte(`{"results":[{"picture":{"large":"https://example.com/ada.png"}}]}`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	avatar, err := client.Get(context.Background(), "ada")
	if err != nil || avatar.URL != "https://example.com/ada.png" {
		t.Fatalf("avatar = %#v, err = %v", avatar, err)
	}
}

func TestClientIsDisabledWithoutURL(t *testing.T) {
	client, err := NewClient("")
	if err != nil || client != nil {
		t.Fatalf("client = %#v, err = %v", client, err)
	}
}
