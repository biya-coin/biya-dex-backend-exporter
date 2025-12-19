package collectors

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/biya-coin/biya-dex-backend-exporter/internal/adapters/stake"
	"github.com/biya-coin/biya-dex-backend-exporter/internal/metrics"
)

func TestRealtimeStakeCollector_ProvideMDValidatorsActiveLen(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/stake/validators" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code":    0,
			"message": "success",
			"data": map[string]any{
				"validators": []any{
					map[string]any{
						"id":               "1",
						"moniker":          "v1",
						"operatorAddress":  "op1",
						"consensusAddress": "cons1",
						"jailed":           false,
						"status":           3,
						"tokens":           "0",
						"uptimePercentage": 99.0,
					},
					map[string]any{
						"id":               "2",
						"moniker":          "v2",
						"operatorAddress":  "op2",
						"consensusAddress": "cons2",
						"jailed":           true,
						"status":           3,
						"tokens":           "0",
						"uptimePercentage": 50.0,
					},
				},
				"pagination": map[string]any{
					"page":       1,
					"pageSize":   100,
					"total":      "2",
					"totalPages": 1,
					"hasPrev":    false,
					"hasNext":    false,
				},
			},
		})
	}))
	defer srv.Close()

	_, m := metrics.New("biya", "dev", "none")
	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, &slog.HandlerOptions{}))
	cli := stake.NewClient(srv.URL+"/stake", "k", 2*time.Second)

	c := NewRealtimeStakeCollector(logger, m, cli)
	if err := c.Run(context.Background()); err != nil {
		t.Fatalf("collector run err: %v", err)
	}

	out := m.RenderText()
	assertContains(t, out, "\nbiya_validators_active 2\n")
}


