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

	"github.com/biya-coin/biya-dex-backend-exporter/internal/adapters/tendermint"
	"github.com/biya-coin/biya-dex-backend-exporter/internal/config"
	"github.com/biya-coin/biya-dex-backend-exporter/internal/metrics"
)

func TestMinuteChainCollector_MempoolCapacityAndSize(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/num_unconfirmed_txs":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"total": "42",
					"n_txs": "41",
				},
			})
			return
		case "/status":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"node_info": map[string]any{"network": "biya"},
					"sync_info": map[string]any{
						"latest_block_height": "1",
						"latest_block_time":   "2025-01-01T00:00:00Z",
						"catching_up":         false,
					},
				},
			})
			return
		case "/block":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"block": map[string]any{
						"data": map[string]any{"txs": []any{}},
					},
				},
			})
			return
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}))
	defer srv.Close()

	_, m := metrics.New("biya", "dev", "none")
	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, &slog.HandlerOptions{}))
	tm := tendermint.NewClient(srv.URL, 2*time.Second)

	c := NewMinuteChainCollector(logger, m, tm, config.MockConfig{Enabled: false}, 5000)
	if err := c.Run(context.Background()); err != nil {
		t.Fatalf("collector run err: %v", err)
	}

	out := m.RenderText()
	assertContains(t, out, "\nbiya_mempool_capacity 5000\n")
	assertContains(t, out, "\nbiya_mempool_size 42\n")
	assertContains(t, out, "\nbiya_congestion_ratio 0.0084\n")
}


