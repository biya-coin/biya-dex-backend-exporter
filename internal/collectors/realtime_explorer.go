package collectors

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"

	"github.com/biya-coin/biya-dex-backend-exporter/internal/adapters/explorer"
	"github.com/biya-coin/biya-dex-backend-exporter/internal/config"
	"github.com/biya-coin/biya-dex-backend-exporter/internal/metrics"
)

// RealtimeExplorerCollector 负责用 biya-explorer API 填充 METRICS.md 中的 explorer 指标。
// 拿不到的数据（接口缺失/字段不明确/APIKey 未配置）会按 mock 配置兜底为固定值，确保 exporter 可用。
type RealtimeExplorerCollector struct {
	log  *slog.Logger
	m    *metrics.Metrics
	api  *explorer.Client
	mock config.MockConfig
}

func NewRealtimeExplorerCollector(log *slog.Logger, m *metrics.Metrics, api *explorer.Client, mock config.MockConfig) *RealtimeExplorerCollector {
	return &RealtimeExplorerCollector{log: log, m: m, api: api, mock: mock}
}

func (c *RealtimeExplorerCollector) Run(ctx context.Context) error {
	// 1) block height（优先 explorer；失败则不报错，等 node collector 兜底）
	if v, ok := c.readLatestBlockHeight(ctx); ok {
		c.m.SetGauge("biya_block_height", nil, v)
	}

	// 2) tx stats（字段不确定，尽量从响应中提取常见数值；失败则兜底固定值）
	if v, ok := c.readTxFailed24H(ctx); ok {
		c.m.SetGauge("biya_tx_failed_24h_total", nil, v)
	}

	// 3) gas utilization（若无法解析，使用 mock 配置值）
	if c.mock.Enabled {
		c.m.SetGauge("biya_gas_utilization_ratio", nil, c.mock.Values.GasUtilizationRatio)
	}

	return nil
}

func (c *RealtimeExplorerCollector) readLatestBlockHeight(ctx context.Context) (float64, bool) {
	raw, err := c.api.GetLatestBlockHeight(ctx)
	if err != nil {
		c.m.SetGauge("biya_exporter_source_up", map[string]string{"source": "explorer_latest_block_height"}, 0)
		return 0, false
	}
	c.m.SetGauge("biya_exporter_source_up", map[string]string{"source": "explorer_latest_block_height"}, 1)
	// 尝试从 data 中提取 height 字段（常见命名：height / latestBlockHeight）
	if v, ok := findFirstNumber(raw, "height", "latestBlockHeight", "blockHeight"); ok {
		return v, true
	}
	return 0, false
}

func (c *RealtimeExplorerCollector) readTxFailed24H(ctx context.Context) (float64, bool) {
	raw, err := c.api.GetFailedTransactions24H(ctx, explorer.NestedPagination{Page: 1, PageSize: 10})
	if err != nil {
		c.m.SetGauge("biya_exporter_source_up", map[string]string{"source": "explorer_failed_transactions_24h"}, 0)
		if c.mock.Enabled {
			// 没有对应 mock 字段时，先返回 0
			return 0, true
		}
		return 0, false
	}
	c.m.SetGauge("biya_exporter_source_up", map[string]string{"source": "explorer_failed_transactions_24h"}, 1)

	// 常见：data.pagination.total 或 data.total / data.count
	if v, ok := findFirstNumber(raw, "total", "count", "failedTotal"); ok {
		return v, true
	}
	return 0, true
}

// findFirstNumber 在任意 JSON 中按“常见字段名”搜索第一个可解析的数值。
// 这里用启发式而非严格 schema，目的是快速把指标跑通，后续再精确对接字段。
func findFirstNumber(raw json.RawMessage, keys ...string) (float64, bool) {
	var anyv any
	if err := json.Unmarshal(raw, &anyv); err != nil {
		return 0, false
	}
	for _, k := range keys {
		if v, ok := findByKey(anyv, k); ok {
			if f, ok := toFloat64(v); ok {
				return f, true
			}
		}
	}
	return 0, false
}

func findByKey(v any, key string) (any, bool) {
	switch t := v.(type) {
	case map[string]any:
		if vv, ok := t[key]; ok {
			return vv, true
		}
		for _, vv := range t {
			if out, ok := findByKey(vv, key); ok {
				return out, true
			}
		}
	case []any:
		for _, vv := range t {
			if out, ok := findByKey(vv, key); ok {
				return out, true
			}
		}
	}
	return nil, false
}

func toFloat64(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case json.Number:
		f, err := x.Float64()
		return f, err == nil
	case string:
		if x == "" {
			return 0, false
		}
		f, err := strconv.ParseFloat(x, 64)
		return f, err == nil
	default:
		return 0, false
	}
}
