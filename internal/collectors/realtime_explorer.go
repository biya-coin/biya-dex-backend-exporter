package collectors

import (
	"context"
	"log/slog"

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
	// provide.md 指标口径：
	// - block height:   GET /api/v1/block/latest                -> .data.data[0].height
	// - tx stats:       GET /api/v1/transaction/stats           -> .data.count_24h / .data.tps / .data.avg_block_time / .data.active_addresses_24h
	// - gas price gwei: GET /api/v1/block/gas-utilization       -> .data.gas_price（你已澄清：该字段即“平均 gas 费”）

	if v, ok := c.readLatestBlockHeight(ctx); ok {
		c.m.SetGauge("biya_block_height", nil, v)
	}

	if stats, ok := c.readTransactionStats(ctx); ok {
		if stats.Count24H >= 0 {
			c.m.SetGauge("biya_tx_24h_total", nil, stats.Count24H)
		}
		if stats.TPS >= 0 {
			c.m.SetGauge("biya_tps_current", nil, stats.TPS)
		}
		if stats.AvgBlockTimeSeconds >= 0 {
			c.m.SetGauge("biya_block_time_seconds", nil, stats.AvgBlockTimeSeconds)
		}
		if stats.ActiveAddresses24H >= 0 {
			c.m.SetGauge("biya_active_addresses_24h", nil, stats.ActiveAddresses24H)
		}
	}

	if v, ok := c.readGasPriceGwei(ctx); ok {
		c.m.SetGauge("biya_gas_price", nil, v)
	}

	if v, ok := c.readGasUtilization(ctx); ok {
		c.m.SetGauge("biya_gas_utilization", nil, v)
	}

	return nil
}

func (c *RealtimeExplorerCollector) readLatestBlockHeight(ctx context.Context) (float64, bool) {
	raw, err := c.api.GetLatestBlocks(ctx, explorer.CursorPage{Page: 1, PageSize: 1})
	if err != nil {
		c.m.SetGauge("biya_exporter_source_up", map[string]string{"source": "explorer_latest_block"}, 0)
		return 0, false
	}
	c.m.SetGauge("biya_exporter_source_up", map[string]string{"source": "explorer_latest_block"}, 1)

	// apiclient 已剥离 envelope.data，因此这里的结构一般为：
	// {"data":[{"height":"123"}], ...}
	var resp struct {
		Data []struct {
			Height any `json:"height"`
		} `json:"data"`
	}
	if err := jsonUnmarshal(raw, &resp); err != nil {
		c.log.Warn("explorer latest block parse failed", "collector", "realtime_explorer", "method", "readLatestBlockHeight", "err", err)
		return 0, false
	}
	if len(resp.Data) == 0 {
		return 0, false
	}
	v, ok := toFloat64(resp.Data[0].Height)
	return v, ok
}

type txStats struct {
	Count24H            float64
	TPS                 float64
	AvgBlockTimeSeconds float64
	ActiveAddresses24H  float64
}

func (c *RealtimeExplorerCollector) readTransactionStats(ctx context.Context) (txStats, bool) {
	raw, err := c.api.GetTransactionStats(ctx)
	if err != nil {
		c.m.SetGauge("biya_exporter_source_up", map[string]string{"source": "explorer_transaction_stats"}, 0)
		return txStats{}, false
	}
	c.m.SetGauge("biya_exporter_source_up", map[string]string{"source": "explorer_transaction_stats"}, 1)

	// apiclient 已剥离 envelope.data，因此这里期望结构为：
	// {"count_24h":..., "tps":..., "avg_block_time":..., "active_addresses_24h":...}
	var resp struct {
		Count24H           any `json:"count_24h"`
		TPS                any `json:"tps"`
		AvgBlockTime       any `json:"avg_block_time"`
		ActiveAddresses24H any `json:"active_addresses_24h"`
	}
	if err := jsonUnmarshal(raw, &resp); err != nil {
		c.log.Warn("explorer tx stats parse failed", "collector", "realtime_explorer", "method", "readTransactionStats", "err", err)
		return txStats{}, false
	}

	out := txStats{
		Count24H:            -1,
		TPS:                 -1,
		AvgBlockTimeSeconds: -1,
		ActiveAddresses24H:  -1,
	}
	if v, ok := toFloat64(resp.Count24H); ok {
		out.Count24H = v
	}
	if v, ok := toFloat64(resp.TPS); ok {
		out.TPS = v
	}
	if v, ok := toFloat64(resp.AvgBlockTime); ok {
		out.AvgBlockTimeSeconds = v
	}
	if v, ok := toFloat64(resp.ActiveAddresses24H); ok {
		out.ActiveAddresses24H = v
	}
	return out, true
}

func (c *RealtimeExplorerCollector) readGasPriceGwei(ctx context.Context) (float64, bool) {
	raw, err := c.api.GetBlockGasUtilization(ctx)
	if err != nil {
		c.m.SetGauge("biya_exporter_source_up", map[string]string{"source": "explorer_block_gas_price"}, 0)
		return 0, false
	}

	// apiclient 已剥离 envelope.data，因此这里期望结构为：
	// {"gas_price": ...}
	var resp struct {
		GasPrice any `json:"gas_price"`
	}
	if err := jsonUnmarshal(raw, &resp); err != nil {
		c.log.Warn("explorer gas price parse failed", "collector", "realtime_explorer", "method", "readGasPriceGwei", "err", err)
		c.m.SetGauge("biya_exporter_source_up", map[string]string{"source": "explorer_block_gas_price"}, 0)
		return 0, false
	}
	v, ok := toFloat64(resp.GasPrice)
	if !ok {
		// 上游在部分环境可能不返回 gas_price 字段；此时视为该 source 不可用，避免“source_up=1 但指标为 0”的误导。
		c.log.Warn("explorer gas price field missing", "collector", "realtime_explorer", "method", "readGasPriceGwei")
		c.m.SetGauge("biya_exporter_source_up", map[string]string{"source": "explorer_block_gas_price"}, 0)
		return 0, false
	}
	c.m.SetGauge("biya_exporter_source_up", map[string]string{"source": "explorer_block_gas_price"}, 1)
	return v, true
}

func (c *RealtimeExplorerCollector) readGasUtilization(ctx context.Context) (float64, bool) {
	raw, err := c.api.GetBlockGasUtilization(ctx)
	if err != nil {
		c.m.SetGauge("biya_exporter_source_up", map[string]string{"source": "explorer_block_gas_utilization"}, 0)
		return 0, false
	}

	// apiclient 已剥离 envelope.data，因此这里期望结构为：
	// {"gas_utilization": ...}
	var resp struct {
		GasUtilization any `json:"gas_utilization"`
	}
	if err := jsonUnmarshal(raw, &resp); err != nil {
		c.log.Warn("explorer gas utilization parse failed", "collector", "realtime_explorer", "method", "readGasUtilization", "err", err)
		c.m.SetGauge("biya_exporter_source_up", map[string]string{"source": "explorer_block_gas_utilization"}, 0)
		return 0, false
	}
	v, ok := toFloat64(resp.GasUtilization)
	if !ok {
		// 上游在部分环境可能不返回 gas_utilization 字段；此时视为该 source 不可用，避免"source_up=1 但指标为 0"的误导。
		c.log.Warn("explorer gas utilization field missing", "collector", "realtime_explorer", "method", "readGasUtilization")
		c.m.SetGauge("biya_exporter_source_up", map[string]string{"source": "explorer_block_gas_utilization"}, 0)
		return 0, false
	}
	c.m.SetGauge("biya_exporter_source_up", map[string]string{"source": "explorer_block_gas_utilization"}, 1)
	return v, true
}
