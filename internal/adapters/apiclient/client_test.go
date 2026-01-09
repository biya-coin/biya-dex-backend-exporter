package apiclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_GetJSON_UnwrapsEnvelopeAndSetsBearer(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test_key" {
			t.Fatalf("Authorization header = %q", got)
		}
		if got := r.URL.Path; got != "/api/v1/account/balances" {
			t.Fatalf("path = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code":    0,
			"message": "success",
			"data": map[string]any{
				"ok": true,
			},
		})
	}))
	defer srv.Close()

	c := New(srv.URL, "test_key", 2*time.Second)

	var out struct {
		OK bool `json:"ok"`
	}
	if err := c.GetJSON(context.Background(), "/api/v1/account/balances", nil, &out); err != nil {
		t.Fatalf("GetJSON err: %v", err)
	}
	if !out.OK {
		t.Fatalf("expected ok=true")
	}
}

func TestClient_GetJSON_AcceptsCode200AsSuccess(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code":    200,
			"message": "success",
			"data": map[string]any{
				"ok": true,
			},
		})
	}))
	defer srv.Close()

	c := New(srv.URL, "", 2*time.Second)
	var out struct {
		OK bool `json:"ok"`
	}
	if err := c.GetJSON(context.Background(), "/x", nil, &out); err != nil {
		t.Fatalf("GetJSON err: %v", err)
	}
	if !out.OK {
		t.Fatalf("expected ok=true")
	}
}

func TestClient_GetJSON_AcceptsRawJSONWithoutEnvelope(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":   true,
			"name": "raw",
		})
	}))
	defer srv.Close()

	c := New(srv.URL, "", 2*time.Second)
	var out struct {
		OK   bool   `json:"ok"`
		Name string `json:"name"`
	}
	if err := c.GetJSON(context.Background(), "/x", nil, &out); err != nil {
		t.Fatalf("GetJSON err: %v", err)
	}
	if !out.OK || out.Name != "raw" {
		t.Fatalf("unexpected out: %+v", out)
	}
}

func TestClient_GetJSON_EmptyAPIKey(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// apiKey 为空时不应设置 Authorization header
		if got := r.Header.Get("Authorization"); got != "" {
			t.Fatalf("Authorization header = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code":    0,
			"message": "success",
			"data": map[string]any{
				"ok": true,
			},
		})
	}))
	defer srv.Close()

	c := New(srv.URL, "", 2*time.Second)
	var out struct {
		OK bool `json:"ok"`
	}
	if err := c.GetJSON(context.Background(), "/x", nil, &out); err != nil {
		t.Fatalf("GetJSON err: %v", err)
	}
	if !out.OK {
		t.Fatalf("expected ok=true")
	}
}
