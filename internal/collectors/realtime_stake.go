package collectors

import (
	"context"
	"log/slog"
	"time"

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
		// provide.md：biya_validators_active 取 validators 数组长度
		c.m.SetGauge("biya_validators_active", nil, 0)
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
	// provide.md：biya_validators_active 取 validators 数组长度（你已澄清）
	c.m.SetGauge("biya_validators_active", nil, float64(total))
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

	// 获取质押统计信息
	c.readStatistics(ctx)

	// 获取惩罚事件
	c.readSlashingEvents(ctx)

	// 获取治理统计信息
	c.readGovernanceStatistics(ctx)

	return nil
}

func (c *RealtimeStakeCollector) readStatistics(ctx context.Context) {
	raw, err := c.api.GetStatistics(ctx)
	if err != nil {
		c.m.SetGauge("biya_exporter_source_up", map[string]string{"source": "stake_statistics"}, 0)
		c.m.SetGauge("biya_staked_total_byb", nil, 0)
		c.m.SetGauge("biya_rewards_24h_total_byb", nil, 0)
		c.m.SetGauge("biya_apr_annual", nil, 0)
		return
	}
	c.m.SetGauge("biya_exporter_source_up", map[string]string{"source": "stake_statistics"}, 1)

	// 尝试解析常见的字段名（兼容多种可能的命名）
	var resp struct {
		TotalStaked     any `json:"totalStaked"`
		TotalStakedBYB  any `json:"totalStakedByb"`
		StakedTotal     any `json:"stakedTotal"`
		Rewards24H      any `json:"rewards24h"`
		Rewards24HBYB   any `json:"rewards24hByb"`
		Rewards24HTotal any `json:"rewards24hTotal"`
		APR             any `json:"apr"`
		APRAnnual       any `json:"aprAnnual"`
		AnnualAPR       any `json:"annualApr"`
		StakingRatio    any `json:"stakingRatio"`
		StakedRatio     any `json:"stakedRatio"`
	}
	if err := jsonUnmarshal(raw, &resp); err != nil {
		c.log.Warn("stake statistics parse failed", "collector", "realtime_stake", "method", "readStatistics", "err", err)
		return
	}

	// 总质押量 (BYB)
	if v, ok := toFloat64(resp.TotalStakedBYB); ok {
		c.m.SetGauge("biya_staked_total_byb", nil, v)
	} else if v, ok := toFloat64(resp.TotalStaked); ok {
		c.m.SetGauge("biya_staked_total_byb", nil, v)
	} else if v, ok := toFloat64(resp.StakedTotal); ok {
		c.m.SetGauge("biya_staked_total_byb", nil, v)
	}

	// 24h总奖励 (BYB)
	if v, ok := toFloat64(resp.Rewards24HBYB); ok {
		c.m.SetGauge("biya_rewards_24h_total_byb", nil, v)
	} else if v, ok := toFloat64(resp.Rewards24HTotal); ok {
		c.m.SetGauge("biya_rewards_24h_total_byb", nil, v)
	} else if v, ok := toFloat64(resp.Rewards24H); ok {
		c.m.SetGauge("biya_rewards_24h_total_byb", nil, v)
	}

	// 年化收益率 (0-100)
	if v, ok := toFloat64(resp.APRAnnual); ok {
		c.m.SetGauge("biya_apr_annual", nil, v)
	} else if v, ok := toFloat64(resp.AnnualAPR); ok {
		c.m.SetGauge("biya_apr_annual", nil, v)
	} else if v, ok := toFloat64(resp.APR); ok {
		c.m.SetGauge("biya_apr_annual", nil, v)
	}

	// 质押比例
	if v, ok := toFloat64(resp.StakingRatio); ok {
		c.m.SetGauge("biya_staked_ratio", nil, v)
	} else if v, ok := toFloat64(resp.StakedRatio); ok {
		c.m.SetGauge("biya_staked_ratio", nil, v)
	}
}

func (c *RealtimeStakeCollector) readSlashingEvents(ctx context.Context) {
	// 获取过去24小时的惩罚事件
	endTime := time.Now().UTC()
	startTime := endTime.Add(-24 * time.Hour)
	p := stake.NestedPagination{Page: 1, PageSize: 100}

	raw, err := c.api.GetSlashingEvents(ctx, startTime.Format(time.RFC3339), endTime.Format(time.RFC3339), p)
	if err != nil {
		c.m.SetGauge("biya_exporter_source_up", map[string]string{"source": "stake_slashing_events"}, 0)
		c.m.SetGauge("biya_slashing_events_24h", nil, 0)
		return
	}
	c.m.SetGauge("biya_exporter_source_up", map[string]string{"source": "stake_slashing_events"}, 1)

	// 尝试解析事件列表
	var resp struct {
		Events []struct {
			Type string `json:"type"`
		} `json:"events"`
		Data []struct {
			Type string `json:"type"`
		} `json:"data"`
		Count int `json:"count"`
		Total int `json:"total"`
	}
	if err := jsonUnmarshal(raw, &resp); err != nil {
		c.log.Warn("stake slashing events parse failed", "collector", "realtime_stake", "method", "readSlashingEvents", "err", err)
		return
	}

	// 统计24小时内的惩罚事件数量
	var events []struct {
		Type string `json:"type"`
	}
	if len(resp.Events) > 0 {
		events = resp.Events
	} else if len(resp.Data) > 0 {
		events = resp.Data
	}

	count24H := 0
	typeCount := make(map[string]int)
	for _, event := range events {
		count24H++
		if event.Type != "" {
			typeCount[event.Type]++
		}
	}

	// 如果响应中有 count 或 total 字段，优先使用
	if resp.Count > 0 {
		count24H = resp.Count
	} else if resp.Total > 0 {
		count24H = resp.Total
	}

	c.m.SetGauge("biya_slashing_events_24h", nil, float64(count24H))

	// 按类型统计总惩罚事件（使用 SetGauge，因为当前 registry 的 counter 通过 SetGauge 写入）
	for eventType, count := range typeCount {
		c.m.SetGauge("biya_slashing_events_total", map[string]string{"type": eventType}, float64(count))
	}
}

func (c *RealtimeStakeCollector) readGovernanceStatistics(ctx context.Context) {
	raw, err := c.api.GetGovernanceStatistics(ctx)
	if err != nil {
		c.m.SetGauge("biya_exporter_source_up", map[string]string{"source": "stake_governance_statistics"}, 0)
		c.m.SetGauge("biya_voting_power_total", nil, 0)
		c.m.SetGauge("biya_participation_rate_avg", nil, 0)
		return
	}
	c.m.SetGauge("biya_exporter_source_up", map[string]string{"source": "stake_governance_statistics"}, 1)

	// 尝试解析常见的字段名
	var resp struct {
		VotingPowerTotal      any `json:"votingPowerTotal"`
		TotalVotingPower      any `json:"totalVotingPower"`
		ParticipationRateAvg  any `json:"participationRateAvg"`
		AvgParticipationRate  any `json:"avgParticipationRate"`
		AverageParticipation  any `json:"averageParticipation"`
		ParticipationRate     any `json:"participationRate"`
	}
	if err := jsonUnmarshal(raw, &resp); err != nil {
		c.log.Warn("stake governance statistics parse failed", "collector", "realtime_stake", "method", "readGovernanceStatistics", "err", err)
		return
	}

	// 总投票权重
	if v, ok := toFloat64(resp.VotingPowerTotal); ok {
		c.m.SetGauge("biya_voting_power_total", nil, v)
	} else if v, ok := toFloat64(resp.TotalVotingPower); ok {
		c.m.SetGauge("biya_voting_power_total", nil, v)
	}

	// 平均参与率
	if v, ok := toFloat64(resp.ParticipationRateAvg); ok {
		c.m.SetGauge("biya_participation_rate_avg", nil, v)
	} else if v, ok := toFloat64(resp.AvgParticipationRate); ok {
		c.m.SetGauge("biya_participation_rate_avg", nil, v)
	} else if v, ok := toFloat64(resp.AverageParticipation); ok {
		c.m.SetGauge("biya_participation_rate_avg", nil, v)
	} else if v, ok := toFloat64(resp.ParticipationRate); ok {
		c.m.SetGauge("biya_participation_rate_avg", nil, v)
	}
}
