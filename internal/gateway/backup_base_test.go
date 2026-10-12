package gateway

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yetone/magpie/internal/provider"
)

func TestForwardFailsOverToBackupBase(t *testing.T) {
	unreachable := "http://127.0.0.1:1/v1" // nothing listens there: the ask fails on the wire

	var got []byte
	var hit bool
	live := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit = true
		body, _ := io.ReadAll(r.Body)
		got = body
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"choices":[{"message":{"content":"ok"}}]}`))
	}))
	defer live.Close()

	p := provider.Provider{
		ID:      "t",
		Chat:    unreachable,
		Backups: map[string][]string{"chat": {live.URL + "/v1"}},
	}

	s := &Server{client: http.DefaultClient}
	res, err := s.forward(context.Background(), p, provider.Chat, "/chat/completions", []byte(`{"model":"m"}`), http.Header{})
	if err != nil {
		t.Fatalf("forward: %v", err)
	}
	if res.StatusCode != 200 {
		t.Fatalf("status %d", res.StatusCode)
	}
	if !hit || string(got) == "" {
		t.Fatal("backup server got no request")
	}
}
