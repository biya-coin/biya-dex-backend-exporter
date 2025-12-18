package collectors

import (
	"context"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/biya-coin/biya-dex-backend-exporter/internal/adapters/tendermint"
	"github.com/biya-coin/biya-dex-backend-exporter/internal/config"
	"github.com/biya-coin/biya-dex-backend-exporter/internal/metrics"
)

type RealtimeChainCollector struct {
	log  *slog.Logger
	m    *metrics.Metrics
	tm   *tendermint.Client
	mock config.MockConfig

	mu         sync.Mutex
	lastHeight int64
	lastTime   time.Time
	lastAvgBT  float64
}

func NewRealtimeChainCollector(log *slog.Logger, m *metrics.Metrics, tm *tendermint.Client, mock config.MockConfig) *RealtimeChainCollector {
	return &RealtimeChainCollector{log: log, m: m, tm: tm, mock: mock}
}

func (c *RealtimeChainCollector) Run(ctx context.Context) error {
	chainID := c.m.ChainID()

	st, err := c.tm.Status(ctx)
	if err != nil {
		c.m.SetGauge("biya_exporter_source_up", map[string]string{"source": "tendermint_status"}, 0)
		return err
	}
	c.m.SetGauge("biya_exporter_source_up", map[string]string{"source": "tendermint_status"}, 1)

	h, err := strconv.ParseInt(st.Result.SyncInfo.LatestBlockHeight, 10, 64)
	if err != nil {
		return err
	}
	c.m.SetGauge("biya_chain_head_block_height", map[string]string{"chain_id": chainID}, float64(h))
	// METRICS.md
	c.m.SetGauge("biya_block_height", nil, float64(h))
	c.m.SetGauge("biya_blocks_total", nil, float64(h)) // 以高度近似 total blocks（真实 total 以 explorer 为准，后续再对接）
	if st.Result.SyncInfo.CatchingUp {
		c.m.SetGauge("biya_chain_node_catching_up", map[string]string{"chain_id": chainID}, 1)
		c.m.SetGauge("biya_node_sync_status", map[string]string{"node": "default"}, 0)
	} else {
		c.m.SetGauge("biya_chain_node_catching_up", map[string]string{"chain_id": chainID}, 0)
		c.m.SetGauge("biya_node_sync_status", map[string]string{"node": "default"}, 1)
	}
	c.m.SetGauge("biya_node_sync_height", map[string]string{"node": "default"}, float64(h))
	c.m.SetGauge("biya_node_behind_blocks", map[string]string{"node": "default"}, 0)

	avgBT := c.updateBlockTimeAvg(h, st.Result.SyncInfo.LatestBlockTime)
	if avgBT > 0 {
		c.m.SetGauge("biya_chain_block_time_seconds_avg", map[string]string{"chain_id": chainID}, avgBT)
		c.m.SetGauge("biya_block_time_seconds", nil, avgBT)
		// BFT 下确认时间可先近似为出块时间
		c.m.SetGauge("biya_chain_tx_confirm_time_seconds_avg", map[string]string{"chain_id": chainID}, avgBT)
		c.m.SetGauge("biya_tx_confirm_time_avg_seconds", nil, avgBT)
		// histogram 先用近似值打一条样本，后续接 explorer 的真实确认时间分布
		c.m.ObserveHistogramMetric("biya_tx_confirm_time_seconds", nil, []float64{1, 2, 3, 5, 10, 20, 30, 60, 120}, avgBT)
	} else if c.mock.Enabled {
		v := c.mock.Values.TxConfirmTimeSeconds
		c.m.SetGauge("biya_chain_tx_confirm_time_seconds_avg", map[string]string{"chain_id": chainID}, v)
		c.m.SetGauge("biya_tx_confirm_time_avg_seconds", nil, v)
		c.m.ObserveHistogramMetric("biya_tx_confirm_time_seconds", nil, []float64{1, 2, 3, 5, 10, 20, 30, 60, 120}, v)
	}

	// gas utilization / congestion 目前按约定先 Mock
	if c.mock.Enabled {
		c.m.SetGauge("biya_chain_block_gas_utilization_ratio_avg", map[string]string{"chain_id": chainID}, c.mock.Values.GasUtilizationRatio)
		c.m.SetGauge("biya_chain_congestion_ratio", map[string]string{"chain_id": chainID}, c.mock.Values.CongestionRatio)
		c.m.SetGauge("biya_gas_utilization_ratio", nil, c.mock.Values.GasUtilizationRatio)
	}

	return nil
}

func (c *RealtimeChainCollector) updateBlockTimeAvg(latestHeight int64, latestTime time.Time) float64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 首次无法计算平均值，仅记录基线
	if c.lastHeight == 0 || c.lastTime.IsZero() {
		c.lastHeight = latestHeight
		c.lastTime = latestTime
		return c.lastAvgBT
	}

	dh := latestHeight - c.lastHeight
	if dh <= 0 {
		// 高度未增长或回退，直接返回上一次值
		return c.lastAvgBT
	}

	dt := latestTime.Sub(c.lastTime).Seconds()
	if dt <= 0 {
		return c.lastAvgBT
	}

	bt := dt / float64(dh)
	// 简单 EMA 平滑，避免瞬时抖动；alpha 固定 0.3（低成本、够用）
	if c.lastAvgBT <= 0 {
		c.lastAvgBT = bt
	} else {
		const alpha = 0.3
		c.lastAvgBT = alpha*bt + (1-alpha)*c.lastAvgBT
	}

	c.lastHeight = latestHeight
	c.lastTime = latestTime

	// 极端值保护：出块时间异常大时不让它把指标打爆（仍保留在日志里）
	if c.lastAvgBT > 3600 {
		return 0
	}
	return c.lastAvgBT
}
