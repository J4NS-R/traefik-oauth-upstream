package traefik_oauth_upstream_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	traefik_oauth_upstream "github.com/Koellewe/traefik-oauth-upstream"
)

func TestBlankRedirect(t *testing.T) {
	cfg := traefik_oauth_upstream.CreateConfig()
	cfg.ClientID = "cid"
	cfg.ClientSecret = "csec"
	cfg.AuthURL = "https://auth.example.org/auth"
	cfg.TokenURL = "https://auth.example.org/auth"
	cfg.PersistDir = "/tmp/oauth_persist"
	cfg.Scopes = []string{"profile"}

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	handler, err := traefik_oauth_upstream.New(ctx, next, cfg, "test-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)

	res := recorder.Result()
	if res.StatusCode != 420 {
		t.Errorf("Bad status code. Expecting 420, but got %d", res.StatusCode)
	}
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(bodyBytes), "https://localhost/_oauth") {
		t.Error("Expecting redirect URL to be in 420 payload")
	}
}

// Todo test oauth callback

// Todo test normal case: add auth header

// Todo test refresh case
