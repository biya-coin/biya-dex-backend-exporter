package explorer

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_GetAccountTransactions_UsesNestedPaginationParams(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer k" {
			t.Fatalf("Authorization header = %q", got)
		}
		if got := r.URL.Path; got != "/api/v1/account/transactions" {
			t.Fatalf("path = %q", got)
		}
		q := r.URL.Query()
		if got := q.Get("address"); got != "biya1xxx" {
			t.Fatalf("address = %q", got)
		}
		if got := q.Get("pagination.page"); got != "1" {
			t.Fatalf("pagination.page = %q", got)
		}
		if got := q.Get("pagination.pageSize"); got != "10" {
			t.Fatalf("pagination.pageSize = %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code":    0,
			"message": "success",
			"data": map[string]any{
				"transactions": []any{},
				"pagination": map[string]any{
					"page":       1,
					"pageSize":   10,
					"total":      "0",
					"totalPages": 1,
					"hasPrev":    false,
					"hasNext":    false,
				},
			},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "k", 2*time.Second)
	_, err := c.GetAccountTransactions(context.Background(), "biya1xxx", NestedPagination{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("GetAccountTransactions err: %v", err)
	}
}
