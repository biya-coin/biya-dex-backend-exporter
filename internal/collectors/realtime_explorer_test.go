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

	"github.com/biya-coin/biya-dex-backend-exporter/internal/adapters/explorer"
	"github.com/biya-coin/biya-dex-backend-exporter/internal/config"
	"github.com/biya-coin/biya-dex-backend-exporter/internal/metrics"
)

func TestRealtimeExplorerCollector_ProvideMDMetrics(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/v1/block/latest":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code":    0,
				"message": "success",
				"data": map[string]any{
					"data": []any{
						map[string]any{"height": "123"},
					},
				},
			})
			return
		case "/api/v1/transaction/stats":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code":    0,
				"message": "success",
				"data": map[string]any{
					"count_24h":            5,
					"tps":                 "7.5",
					"avg_block_time":      2.2,
					"active_addresses_24h": "100",
				},
			})
			return
		case "/api/v1/block/gas-utilization":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code":    0,
				"message": "success",
				"data": map[string]any{
					"gas_price": "88.8",
				},
			})
			return
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"code":404,"message":"not found","data":{}}`))
			return
		}
	}))
	defer srv.Close()

	_, m := metrics.New("biya", "dev", "none")
	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, &slog.HandlerOptions{}))
	cli := explorer.NewClient(srv.URL, "k", 2*time.Second)

	c := NewRealtimeExplorerCollector(logger, m, cli, config.MockConfig{Enabled: false})
	if err := c.Run(context.Background()); err != nil {
		t.Fatalf("collector run err: %v", err)
	}

	out := m.RenderText()
	assertContains(t, out, "\nbiya_block_height 123\n")
	assertContains(t, out, "\nbiya_tx_24h_total 5\n")
	assertContains(t, out, "\nbiya_tps_current 7.5\n")
	assertContains(t, out, "\nbiya_block_time_seconds 2.2\n")
	assertContains(t, out, "\nbiya_active_addresses_24h 100\n")
	assertContains(t, out, "\nbiya_gas_price 88.8\n")
}

func assertContains(t *testing.T, s, sub string) {
	t.Helper()
	if !strings.Contains(s, sub) {
		t.Fatalf("expected output to contain %q\n\nfull output:\n%s", sub, s)
	}
}


