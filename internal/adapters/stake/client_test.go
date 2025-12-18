package stake

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_GetValidators_UsesBearerAndNormalizesBaseURL(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer k" {
			t.Fatalf("Authorization header = %q", got)
		}
		if got := r.URL.Path; got != "/stake/validators" {
			t.Fatalf("path = %q", got)
		}
		if got := r.URL.Query().Get("page"); got != "2" {
			t.Fatalf("page = %q", got)
		}
		if got := r.URL.Query().Get("pageSize"); got != "50" {
			t.Fatalf("pageSize = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code":    0,
			"message": "success",
			"data": map[string]any{
				"validators": []any{},
				"pagination": map[string]any{
					"page":       2,
					"pageSize":   50,
					"total":      "0",
					"totalPages": 1,
					"hasPrev":    true,
					"hasNext":    false,
				},
			},
		})
	}))
	defer srv.Close()

	// baseURL 带 /stake，client 应自动归一化避免 /stake/stake
	c := NewClient(srv.URL+"/stake", "k", 2*time.Second)
	resp, err := c.GetValidators(context.Background(), 2, 50)
	if err != nil {
		t.Fatalf("GetValidators err: %v", err)
	}
	if resp == nil {
		t.Fatalf("resp is nil")
	}
}
