package collectors

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/biya-coin/biya-dex-backend-exporter/internal/adapters/tendermint"
	"github.com/biya-coin/biya-dex-backend-exporter/internal/config"
	"github.com/biya-coin/biya-dex-backend-exporter/internal/metrics"
)

type MinuteChainCollector struct {
	log  *slog.Logger
	m    *metrics.Metrics
	tm   *tendermint.Client
	mock config.MockConfig

	tpsWindow time.Duration
	samples   []tpsSample
}

type tpsSample struct {
	at      time.Time
	txCount int
}

func NewMinuteChainCollector(log *slog.Logger, m *metrics.Metrics, tm *tendermint.Client, mock config.MockConfig) *MinuteChainCollector {
	return &MinuteChainCollector{
		log:       log,
		m:         m,
		tm:        tm,
		mock:      mock,
		tpsWindow: 60 * time.Second,
		samples:   make([]tpsSample, 0, 8),
	}
}

func (c *MinuteChainCollector) Run(ctx context.Context) error {
	chainID := c.m.ChainID()

	// 1) mempool pending
	if v, ok := c.readMempoolPending(ctx); ok {
		c.m.SetGauge("biya_chain_mempool_pending_txs", map[string]string{"chain_id": chainID}, v)
		c.m.SetGauge("biya_mempool_size", nil, v)
	}

	// 2) TPS window
	if v, ok := c.readTPSWindow(ctx); ok {
		c.m.SetGauge("biya_chain_tps_window", map[string]string{"chain_id": chainID}, v)
		c.m.SetGauge("biya_tps_current", nil, v)
	}

	return nil
}

func (c *MinuteChainCollector) readMempoolPending(ctx context.Context) (float64, bool) {
	resp, err := c.tm.NumUnconfirmedTxs(ctx)
	if err != nil {
		c.m.SetGauge("biya_exporter_source_up", map[string]string{"source": "tendermint_mempool"}, 0)
		if c.mock.Enabled {
			c.log.Warn("mempool endpoint unavailable, use mock", "source", "tendermint_mempool", "err", err)
			return c.mock.Values.MempoolPendingTxs, true
		}
		return 0, false
	}
	c.m.SetGauge("biya_exporter_source_up", map[string]string{"source": "tendermint_mempool"}, 1)

	n, err := strconv.ParseFloat(resp.Result.NTxs, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

func (c *MinuteChainCollector) readTPSWindow(ctx context.Context) (float64, bool) {
	// 用 /status 取最新高度，再用 /block 取该高度 txs 数，形成时间序列近似 TPS。
	st, err := c.tm.Status(ctx)
	if err != nil {
		c.m.SetGauge("biya_exporter_source_up", map[string]string{"source": "tendermint_status_for_tps"}, 0)
		if c.mock.Enabled {
			c.log.Warn("status endpoint unavailable for tps, use mock", "source", "tendermint_status_for_tps", "err", err)
			return c.mock.Values.TPSWindow, true
		}
		return 0, false
	}
	c.m.SetGauge("biya_exporter_source_up", map[string]string{"source": "tendermint_status_for_tps"}, 1)

	h, err := strconv.ParseInt(st.Result.SyncInfo.LatestBlockHeight, 10, 64)
	if err != nil {
		return 0, false
	}

	blk, err := c.tm.Block(ctx, h)
	if err != nil {
		c.m.SetGauge("biya_exporter_source_up", map[string]string{"source": "tendermint_block"}, 0)
		if c.mock.Enabled {
			c.log.Warn("block endpoint unavailable for tps, use mock", "source", "tendermint_block", "err", err)
			return c.mock.Values.TPSWindow, true
		}
		return 0, false
	}
	c.m.SetGauge("biya_exporter_source_up", map[string]string{"source": "tendermint_block"}, 1)

	now := time.Now()
	txCount := len(blk.Result.Block.Data.Txs)
	c.samples = append(c.samples, tpsSample{at: now, txCount: txCount})
	c.trimSamples(now)

	if len(c.samples) < 2 {
		return 0, true
	}

	first := c.samples[0]
	last := c.samples[len(c.samples)-1]
	span := last.at.Sub(first.at).Seconds()
	if span <= 0 {
		return 0, true
	}
	sumTx := 0
	for _, s := range c.samples {
		sumTx += s.txCount
	}
	return float64(sumTx) / span, true
}

func (c *MinuteChainCollector) trimSamples(now time.Time) {
	cut := now.Add(-c.tpsWindow)
	i := 0
	for i < len(c.samples) && c.samples[i].at.Before(cut) {
		i++
	}
	if i > 0 {
		c.samples = append([]tpsSample(nil), c.samples[i:]...)
	}
}
