package metrics

// Metrics 统一收敛指标定义，避免 collectors 内零散拼接 metric 名称。
// 注意：label 严格控制低基数；严禁把 address/tx_hash/contract 放入 label。
type Metrics struct {
	chainID string
	reg     *Registry
}

// New 返回自研 registry 与 metrics facade。
func New(chainID, version, commit string) (*Registry, *Metrics) {
	reg := NewRegistry()

	// declare metrics
	reg.MustDeclare("biya_chain_head_block_height", TypeGauge, "Latest block height observed from the chain node.", []string{"chain_id"})
	reg.MustDeclare("biya_chain_node_catching_up", TypeGauge, "Whether the node is catching up (1) or fully synced (0).", []string{"chain_id"})
	reg.MustDeclare("biya_chain_block_time_seconds_avg", TypeGauge, "Average block time in seconds (EMA).", []string{"chain_id"})
	reg.MustDeclare("biya_chain_tps_window", TypeGauge, "Approximate TPS over a rolling time window. May be mocked until explorer/indexer endpoints are ready.", []string{"chain_id"})
	reg.MustDeclare("biya_chain_tx_confirm_time_seconds_avg", TypeGauge, "Average transaction confirmation time in seconds. May be mocked.", []string{"chain_id"})
	reg.MustDeclare("biya_chain_block_gas_utilization_ratio_avg", TypeGauge, "Average gas utilization ratio (0-1). May be mocked.", []string{"chain_id"})
	reg.MustDeclare("biya_chain_mempool_pending_txs", TypeGauge, "Pending transactions in mempool. May be mocked.", []string{"chain_id"})
	reg.MustDeclare("biya_chain_congestion_ratio", TypeGauge, "Congestion ratio (0-1). May be mocked.", []string{"chain_id"})

	reg.MustDeclare("biya_stake_validators_total", TypeGauge, "Total validators returned by stake API.", []string{"chain_id"})
	reg.MustDeclare("biya_stake_validators_bonded", TypeGauge, "Bonded validators count (status==bonded).", []string{"chain_id"})
	reg.MustDeclare("biya_stake_validators_jailed", TypeGauge, "Jailed validators count.", []string{"chain_id"})
	reg.MustDeclare("biya_stake_validators_uptime_percentage_avg", TypeGauge, "Average uptime percentage across validators (aggregate).", []string{"chain_id"})
	reg.MustDeclare("biya_stake_bonded_tokens", TypeGauge, "Bonded tokens from LCD staking pool (raw units).", []string{"chain_id"})

	reg.MustDeclare("biya_exporter_scrape_success", TypeGauge, "Whether a collector run succeeded (1) or failed (0).", []string{"source"})
	reg.MustDeclare("biya_exporter_scrape_duration_seconds", TypeHistogram, "Collector run duration in seconds.", []string{"source"})
	reg.MustDeclare("biya_exporter_build_info", TypeGauge, "Build info as a gauge with labels version/commit.", []string{"version", "commit"})
	reg.MustDeclare("biya_exporter_source_up", TypeGauge, "Whether a concrete data source call is up (1) or down (0).", []string{"source"})

	// ---- Metrics defined by METRICS.md (admin backend) ----
	// 说明：
	// - 当前 registry 为最小实现（counter 以“可写入的 counter 类型”方式输出，值通过 SetGauge 写入）
	// - 拿不到的数据会在 collectors 侧兜底为固定值（通常为 0），避免 exporter 失效。
	//
	// 来源：仓库内 `METRICS.md`
	reg.MustDeclare("biya_block_height", TypeGauge, "Current block height.", nil)
	reg.MustDeclare("biya_block_time_seconds", TypeGauge, "Average block time (last 100 blocks).", nil)
	reg.MustDeclare("biya_blocks_total", TypeCounter, "Total blocks produced.", nil)

	reg.MustDeclare("biya_tx_total", TypeCounter, "Total transactions (success/failed).", []string{"status"})
	reg.MustDeclare("biya_tx_24h_total", TypeGauge, "Transactions in last 24h.", nil)
	reg.MustDeclare("biya_tps_current", TypeGauge, "Current TPS.", nil)
	reg.MustDeclare("biya_tps_24h_avg", TypeGauge, "24h average TPS.", nil)
	reg.MustDeclare("biya_tx_success_rate", TypeGauge, "Transaction success rate (0-1).", nil)
	reg.MustDeclare("biya_tx_failed_24h_total", TypeGauge, "Failed transactions in last 24h.", nil)
	reg.MustDeclare("biya_tx_confirm_time_seconds", TypeHistogram, "Transaction confirmation time distribution (seconds).", nil)
	reg.MustDeclare("biya_tx_confirm_time_avg_seconds", TypeGauge, "Average confirmation time (seconds).", nil)

	reg.MustDeclare("biya_gas_price", TypeGauge, "Current average gas price.", nil)
	reg.MustDeclare("biya_gas_price_24h_max_gwei", TypeGauge, "24h maximum gas price (Gwei).", nil)
	reg.MustDeclare("biya_gas_price_24h_min_gwei", TypeGauge, "24h minimum gas price (Gwei).", nil)
	reg.MustDeclare("biya_gas_utilization", TypeGauge, "Gas利用率", nil)
	reg.MustDeclare("biya_gas_limit_per_block", TypeGauge, "Block gas limit.", nil)
	reg.MustDeclare("biya_gas_used_per_block", TypeGauge, "Average gas used per block.", nil)

	reg.MustDeclare("biya_mempool_size", TypeGauge, "当前交易池中pending的交易数", nil)
	reg.MustDeclare("biya_mempool_capacity", TypeGauge, "Mempool capacity limit.", nil)
	reg.MustDeclare("biya_congestion_ratio", TypeGauge, "Network congestion ratio (mempool_size/capacity).", nil)
	reg.MustDeclare("biya_active_addresses_24h", TypeGauge, "Unique active addresses in 24h.", nil)

	reg.MustDeclare("biya_node_sync_status", TypeGauge, "Node sync status (1=synced, 0=syncing).", []string{"node"})
	reg.MustDeclare("biya_node_sync_height", TypeGauge, "Current node sync height.", []string{"node"})
	reg.MustDeclare("biya_node_behind_blocks", TypeGauge, "Blocks behind latest.", []string{"node"})

	reg.MustDeclare("biya_validators_total", TypeGauge, "Total validators (all created).", nil)
	reg.MustDeclare("biya_validators_consensus", TypeGauge, "Validators participating in consensus (TOP N).", nil)
	reg.MustDeclare("biya_validators_active", TypeGauge, "Active validators.", nil)
	reg.MustDeclare("biya_validators_jailed", TypeGauge, "Jailed validators count.", nil)
	reg.MustDeclare("biya_validators_max", TypeGauge, "MaxValidators parameter.", nil)

	reg.MustDeclare("biya_staked_total_byb", TypeGauge, "Total staked amount (BYB).", nil)
	reg.MustDeclare("biya_staked_ratio", TypeGauge, "Staking ratio (staked/total supply).", nil)
	reg.MustDeclare("biya_rewards_24h_total_byb", TypeGauge, "24h total rewards (BYB).", nil)
	reg.MustDeclare("biya_apr_annual", TypeGauge, "Annual percentage rate (0-100).", nil)
	reg.MustDeclare("biya_slashing_events_24h", TypeGauge, "Slashing events in 24h.", nil)
	reg.MustDeclare("biya_slashing_events_total", TypeCounter, "Total slashing events by type.", []string{"type"})

	reg.MustDeclare("biya_validator_status", TypeGauge, "Validator status (1=active, 0=inactive, -1=jailed).", []string{"address", "moniker"})
	reg.MustDeclare("biya_validator_stake_byb", TypeGauge, "Validator stake (BYB).", []string{"address", "moniker"})
	reg.MustDeclare("biya_validator_voting_power", TypeGauge, "Validator voting power percentage (0-100).", []string{"address", "moniker"})
	reg.MustDeclare("biya_validator_commission_rate", TypeGauge, "Validator commission rate (0-1).", []string{"address", "moniker"})
	reg.MustDeclare("biya_validator_blocks_proposed_total", TypeCounter, "Total blocks proposed by validator.", []string{"address", "moniker"})
	reg.MustDeclare("biya_validator_blocks_missed_total", TypeCounter, "Total blocks missed by validator.", []string{"address", "moniker"})
	reg.MustDeclare("biya_validator_uptime_ratio", TypeGauge, "Validator uptime ratio (0-1).", []string{"address", "moniker"})
	reg.MustDeclare("biya_validator_last_active_timestamp", TypeGauge, "Validator last active timestamp (unix seconds).", []string{"address", "moniker"})
	reg.MustDeclare("biya_validator_rewards_24h_byb", TypeGauge, "Validator 24h rewards (BYB).", []string{"address", "moniker"})
	reg.MustDeclare("biya_validator_jailed", TypeGauge, "Validator jailed (1=yes, 0=no).", []string{"address", "moniker"})

	reg.MustDeclare("biya_proposals_total", TypeCounter, "Total proposals created.", nil)
	reg.MustDeclare("biya_proposals_passed", TypeGauge, "Total passed proposals.", nil)
	reg.MustDeclare("biya_proposals_rejected", TypeGauge, "Total rejected proposals.", nil)
	reg.MustDeclare("biya_proposals_active", TypeGauge, "Currently active proposals.", nil)
	reg.MustDeclare("biya_voting_power_total", TypeGauge, "Total voting power.", nil)
	reg.MustDeclare("biya_participation_rate_avg", TypeGauge, "Average participation rate.", nil)
	reg.MustDeclare("biya_proposal_status", TypeGauge, "Proposal status (0-5 enum).", []string{"id", "title"})
	reg.MustDeclare("biya_proposal_votes_yes", TypeGauge, "Yes votes.", []string{"id"})
	reg.MustDeclare("biya_proposal_votes_no", TypeGauge, "No votes.", []string{"id"})
	reg.MustDeclare("biya_proposal_votes_veto", TypeGauge, "NoWithVeto votes.", []string{"id"})
	reg.MustDeclare("biya_proposal_votes_abstain", TypeGauge, "Abstain votes.", []string{"id"})

	// 关键指标默认置 0（缺失时也能看到 metric 存在；后续逐步对接接口字段）
	reg.SetGauge("biya_tx_total", map[string]string{"status": "success"}, 0)
	reg.SetGauge("biya_tx_total", map[string]string{"status": "failed"}, 0)
	reg.SetGauge("biya_block_height", nil, 0)
	reg.SetGauge("biya_block_time_seconds", nil, 0)
	reg.SetGauge("biya_blocks_total", nil, 0)
	reg.SetGauge("biya_tx_24h_total", nil, 0)
	reg.SetGauge("biya_tps_current", nil, 0)
	reg.SetGauge("biya_tps_24h_avg", nil, 0)
	reg.SetGauge("biya_tx_success_rate", nil, 0)
	reg.SetGauge("biya_tx_failed_24h_total", nil, 0)
	reg.SetGauge("biya_tx_confirm_time_avg_seconds", nil, 0)
	reg.SetGauge("biya_gas_utilization", nil, 0)
	reg.SetGauge("biya_gas_price", nil, 0)
	reg.SetGauge("biya_mempool_size", nil, 0)
	reg.SetGauge("biya_mempool_capacity", nil, 0)
	reg.SetGauge("biya_congestion_ratio", nil, 0)
	reg.SetGauge("biya_validators_total", nil, 0)
	reg.SetGauge("biya_validators_jailed", nil, 0)

	m := &Metrics{chainID: chainID, reg: reg}

	// build info
	reg.SetGauge("biya_exporter_build_info", map[string]string{"version": version, "commit": commit}, 1)

	return reg, m
}

func (m *Metrics) ChainID() string { return m.chainID }

// ---- helpers for collectors ----

func (m *Metrics) SetGauge(metric string, labels map[string]string, v float64) {
	m.reg.SetGauge(metric, labels, v)
}

func (m *Metrics) ObserveDuration(source string, seconds float64) {
	// Prometheus 默认 buckets；这里硬编码一组常用 buckets
	buckets := []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}
	m.reg.ObserveHistogram("biya_exporter_scrape_duration_seconds", map[string]string{"source": source}, buckets, seconds)
}

// ObserveHistogramMetric 提供给 collectors 使用的通用 histogram 观测封装。
func (m *Metrics) ObserveHistogramMetric(metric string, labels map[string]string, buckets []float64, v float64) {
	m.reg.ObserveHistogram(metric, labels, buckets, v)
}

func (m *Metrics) RenderText() string {
	return m.reg.RenderText()
}
