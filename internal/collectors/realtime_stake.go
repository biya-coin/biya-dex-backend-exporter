package collectors

import (
	"context"
	"log/slog"

	"github.com/biya-coin/biya-dex-backend-exporter/internal/adapters/stake"
	"github.com/biya-coin/biya-dex-backend-exporter/internal/metrics"
)

type RealtimeStakeCollector struct {
	log *slog.Logger
	m   *metrics.Metrics
	api *stake.Client
}

func NewRealtimeStakeCollector(log *slog.Logger, m *metrics.Metrics, api *stake.Client) *RealtimeStakeCollector {
	return &RealtimeStakeCollector{log: log, m: m, api: api}
}

func (c *RealtimeStakeCollector) Run(ctx context.Context) error {
	chainID := c.m.ChainID()

	resp, err := c.api.GetValidators(ctx, 1, 100)
	if err != nil {
		c.m.SetGauge("biya_exporter_source_up", map[string]string{"source": "stake_validators"}, 0)
		// 按需求：拿不到先返回固定值，不让 exporter 直接失败
		c.m.SetGauge("biya_validators_total", nil, 0)
		c.m.SetGauge("biya_validators_jailed", nil, 0)
		return nil
	}
	c.m.SetGauge("biya_exporter_source_up", map[string]string{"source": "stake_validators"}, 1)

	total := len(resp.Validators)
	var jailed, bonded int
	var uptimeSum float64
	var uptimeN int

	for _, v := range resp.Validators {
		if v.Jailed {
			jailed++
		}
		// Cosmos staking status：1 unbonded, 2 unbonding, 3 bonded
		if v.Status == 3 {
			bonded++
		}
		if v.UptimePercentage > 0 {
			uptimeSum += v.UptimePercentage
			uptimeN++
		}
	}

	c.m.SetGauge("biya_stake_validators_total", map[string]string{"chain_id": chainID}, float64(total))
	c.m.SetGauge("biya_stake_validators_bonded", map[string]string{"chain_id": chainID}, float64(bonded))
	c.m.SetGauge("biya_stake_validators_jailed", map[string]string{"chain_id": chainID}, float64(jailed))
	// METRICS.md（聚合）
	c.m.SetGauge("biya_validators_total", nil, float64(total))
	c.m.SetGauge("biya_validators_jailed", nil, float64(jailed))

	// METRICS.md（单验证人维度）：只对当前返回的 validators 填充；字段不足的先置 0。
	for _, v := range resp.Validators {
		labels := map[string]string{"address": v.OperatorAddress, "moniker": v.Moniker}
		if v.Jailed {
			c.m.SetGauge("biya_validator_status", labels, -1)
			c.m.SetGauge("biya_validator_jailed", labels, 1)
		} else if v.Status == 3 {
			c.m.SetGauge("biya_validator_status", labels, 1)
			c.m.SetGauge("biya_validator_jailed", labels, 0)
		} else {
			c.m.SetGauge("biya_validator_status", labels, 0)
			c.m.SetGauge("biya_validator_jailed", labels, 0)
		}
		// uptime_percentage 为 0-100；指标要求 0-1
		if v.UptimePercentage > 0 {
			c.m.SetGauge("biya_validator_uptime_ratio", labels, v.UptimePercentage/100.0)
		} else {
			c.m.SetGauge("biya_validator_uptime_ratio", labels, 0)
		}
		// 其他字段暂无接口对接，先置 0
		c.m.SetGauge("biya_validator_commission_rate", labels, 0)
		c.m.SetGauge("biya_validator_stake_byb", labels, 0)
		c.m.SetGauge("biya_validator_voting_power", labels, 0)
		c.m.SetGauge("biya_validator_rewards_24h_byb", labels, 0)
		c.m.SetGauge("biya_validator_last_active_timestamp", labels, 0)
		c.m.SetGauge("biya_validator_blocks_proposed_total", labels, 0)
		c.m.SetGauge("biya_validator_blocks_missed_total", labels, 0)
	}
	if uptimeN > 0 {
		c.m.SetGauge("biya_stake_validators_uptime_percentage_avg", map[string]string{"chain_id": chainID}, uptimeSum/float64(uptimeN))
	}
	return nil
}
