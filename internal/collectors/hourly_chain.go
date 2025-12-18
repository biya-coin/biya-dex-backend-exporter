package collectors

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/biya-coin/biya-dex-backend-exporter/internal/adapters/lcd"
	"github.com/biya-coin/biya-dex-backend-exporter/internal/metrics"
)

type HourlyChainCollector struct {
	log *slog.Logger
	m   *metrics.Metrics
	lcd *lcd.Client
}

func NewHourlyChainCollector(log *slog.Logger, m *metrics.Metrics, lcdCli *lcd.Client) *HourlyChainCollector {
	return &HourlyChainCollector{log: log, m: m, lcd: lcdCli}
}

func (c *HourlyChainCollector) Run(ctx context.Context) error {
	chainID := c.m.ChainID()

	resp, err := c.lcd.StakingPool(ctx)
	if err != nil {
		c.m.SetGauge("biya_exporter_source_up", map[string]string{"source": "lcd_staking_pool"}, 0)
		// bonded tokens 若缺失也可以不影响联调（先不 mock，避免误导）
		return err
	}
	c.m.SetGauge("biya_exporter_source_up", map[string]string{"source": "lcd_staking_pool"}, 1)

	// bonded_tokens 通常是大整数（字符串）
	v, err := strconv.ParseFloat(resp.Pool.BondedTokens, 64)
	if err != nil {
		return err
	}
	c.m.SetGauge("biya_stake_bonded_tokens", map[string]string{"chain_id": chainID}, v)
	return nil
}
