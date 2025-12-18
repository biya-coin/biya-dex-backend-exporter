# Biya Chain Prometheus Metrics Definition

## Overview

This document defines all Prometheus metrics required for the BIYA Chain admin backend, based on the product requirements.

## Prometheus Metrics Model

### Metric Types

| Type | Description | Use Case |
|------|-------------|----------|
| **Counter** | Monotonically increasing value | Total transactions, total blocks |
| **Gauge** | Value that can go up or down | Current TPS, block height, mempool size |
| **Histogram** | Distribution of values | Transaction confirmation time, gas prices |
| **Summary** | Similar to histogram with quantiles | Response times |

### Naming Convention

```
biya_<subsystem>_<name>_<unit>
```

Examples:
- `biya_block_height` - Current block height (gauge)
- `biya_tx_total` - Total transactions (counter)
- `biya_gas_price_gwei` - Gas price in Gwei (gauge)

### Labels

Labels add dimensions to metrics:
- `validator` - Validator operator address
- `status` - Status (success/failed, active/inactive)
- `type` - Transaction or proposal type

---

## Module 1: Runtime Status Monitoring (运行状态监控)

### 1.1 Block Metrics

| Metric Name | Type | Labels | Description | Data Source |
|-------------|------|--------|-------------|-------------|
| `biya_block_height` | Gauge | - | Current block height | biya-explorer |
| `biya_block_time_seconds` | Gauge | - | Average block time (last 100 blocks) | biya-explorer |
| `biya_blocks_total` | Counter | - | Total blocks produced | biya-explorer |

### 1.2 Transaction Metrics

| Metric Name | Type | Labels | Description | Data Source |
|-------------|------|--------|-------------|-------------|
| `biya_tx_total` | Counter | `status` | Total transactions (success/failed) | biya-explorer |
| `biya_tx_24h_total` | Gauge | - | Transactions in last 24h | biya-explorer |
| `biya_tps_current` | Gauge | - | Current TPS (transactions per second) | biya-explorer |
| `biya_tps_24h_avg` | Gauge | - | 24h average TPS | biya-explorer |
| `biya_tx_success_rate` | Gauge | - | Transaction success rate (0-1) | biya-explorer |
| `biya_tx_failed_24h_total` | Gauge | - | Failed transactions in last 24h | biya-explorer |
| `biya_tx_confirm_time_seconds` | Histogram | - | Transaction confirmation time distribution | biya-explorer |
| `biya_tx_confirm_time_avg_seconds` | Gauge | - | Average confirmation time | biya-explorer |

### 1.3 Gas Metrics

| Metric Name | Type | Labels | Description | Data Source |
|-------------|------|--------|-------------|-------------|
| `biya_gas_price_gwei` | Gauge | - | Current average gas price (Gwei) | biya-explorer |
| `biya_gas_price_24h_max_gwei` | Gauge | - | 24h maximum gas price | biya-explorer |
| `biya_gas_price_24h_min_gwei` | Gauge | - | 24h minimum gas price | biya-explorer |
| `biya_gas_utilization_ratio` | Gauge | - | Block gas utilization (0-1) | biya-explorer |
| `biya_gas_limit_per_block` | Gauge | - | Block gas limit | biya-explorer |
| `biya_gas_used_per_block` | Gauge | - | Average gas used per block | biya-explorer |

### 1.4 Network Status Metrics

| Metric Name | Type | Labels | Description | Data Source |
|-------------|------|--------|-------------|-------------|
| `biya_mempool_size` | Gauge | - | Pending transactions in mempool | injective-core |
| `biya_mempool_capacity` | Gauge | - | Mempool capacity limit (default 5000) | genesis config |
| `biya_congestion_ratio` | Gauge | - | Network congestion (mempool_size/capacity) | calculated |
| `biya_active_addresses_24h` | Gauge | - | Unique active addresses in 24h | biya-explorer |

### 1.5 Node Sync Status

| Metric Name | Type | Labels | Description | Data Source |
|-------------|------|--------|-------------|-------------|
| `biya_node_sync_status` | Gauge | `node` | Node sync status (1=synced, 0=syncing) | injective-core |
| `biya_node_sync_height` | Gauge | `node` | Current sync height | injective-core |
| `biya_node_behind_blocks` | Gauge | `node` | Blocks behind latest | calculated |

---

## Module 2: Node Management (节点管理)

### 2.1 Validator Aggregate Metrics

| Metric Name | Type | Labels | Description | Data Source |
|-------------|------|--------|-------------|-------------|
| `biya_validators_total` | Gauge | - | Total validators (all created) | biya-stake |
| `biya_validators_consensus` | Gauge | - | Validators participating in consensus (TOP N) | biya-stake |
| `biya_validators_active` | Gauge | - | Active validators (online + participating) | biya-stake |
| `biya_validators_jailed` | Gauge | - | Jailed validators count | biya-stake |
| `biya_validators_max` | Gauge | - | MaxValidators parameter | biya-stake |

### 2.2 Staking Metrics

| Metric Name | Type | Labels | Description | Data Source |
|-------------|------|--------|-------------|-------------|
| `biya_staked_total_byb` | Gauge | - | Total staked amount (BYB) | biya-stake |
| `biya_staked_ratio` | Gauge | - | Staking ratio (staked/total supply) | biya-stake |
| `biya_rewards_24h_total_byb` | Gauge | - | 24h total rewards (BYB) | biya-stake |
| `biya_apr_annual` | Gauge | - | Annual percentage rate (0-100) | biya-stake |
| `biya_slashing_events_24h` | Gauge | - | Slashing events in 24h | biya-stake |
| `biya_slashing_events_total` | Counter | `type` | Total slashing events by type | biya-stake |

### 2.3 Individual Validator Metrics

| Metric Name | Type | Labels | Description | Data Source |
|-------------|------|--------|-------------|-------------|
| `biya_validator_status` | Gauge | `address`, `moniker` | Validator status (1=active, 0=inactive, -1=jailed) | biya-stake |
| `biya_validator_stake_byb` | Gauge | `address`, `moniker` | Validator self-stake + delegations | biya-stake |
| `biya_validator_voting_power` | Gauge | `address`, `moniker` | Voting power percentage (0-100) | biya-stake |
| `biya_validator_commission_rate` | Gauge | `address`, `moniker` | Commission rate (0-1) | biya-stake |
| `biya_validator_blocks_proposed_total` | Counter | `address`, `moniker` | Total blocks proposed | biya-stake |
| `biya_validator_blocks_missed_total` | Counter | `address`, `moniker` | Total blocks missed | biya-stake |
| `biya_validator_uptime_ratio` | Gauge | `address`, `moniker` | Uptime ratio (0-1) | biya-stake |
| `biya_validator_last_active_timestamp` | Gauge | `address`, `moniker` | Last activity timestamp | biya-stake |
| `biya_validator_rewards_24h_byb` | Gauge | `address`, `moniker` | 24h rewards for validator | biya-stake |
| `biya_validator_jailed` | Gauge | `address`, `moniker` | Is jailed (1=yes, 0=no) | biya-stake |

---

## Module 3: Network Performance (网络性能监控)

### 3.1 Performance Index Components

> Note: Performance index is **calculated by Gin backend**, not stored in Prometheus.
> Gin backend queries these raw metrics and calculates the 0-100 score.

| Metric Name | Type | Description | Weight |
|-------------|------|-------------|--------|
| `biya_tps_current` | Gauge | Current TPS | 30% |
| `biya_tx_confirm_time_avg_seconds` | Gauge | Avg confirmation time | 30% |
| `biya_gas_utilization_ratio` | Gauge | Gas utilization | 25% |
| `biya_mempool_size` | Gauge | Pending transactions | 15% |

### 3.2 Performance Formula (Calculated in Gin)

```
Performance_Score = TPS_Score(30%) + ConfirmTime_Score(30%) + GasUtil_Score(25%) + Pending_Score(15%)

TPS Score (based on mempool clearance time):
  clearance_time = mempool_size / tps_current
  ≤3s → 30 pts
  3-5s → 27-30 pts
  5-10s → 23-27 pts
  10-20s → 18-23 pts
  20-40s → 12-18 pts
  40-60s → 6-12 pts
  >60s → 0-6 pts

Confirm Time Score:
  ≤30s → 20 pts
  30-60s → 15-20 pts
  60-120s → 10-15 pts
  >120s → 0-10 pts
  + Block Time Score (10 pts)

Gas Utilization Score:
  60-80% → 25 pts (optimal)
  40-60% or 80-95% → 15-25 pts
  <40% or >95% → 0-15 pts

Pending TX Score:
  ≤1000 → 15 pts
  1000-3000 → 10-15 pts
  3000-10000 → 4-10 pts
  >10000 → 0-4 pts
```

---

## Module 4: Network Trends (网络指标趋势)

### 4.1 Historical Data Storage (Prometheus Native)

Prometheus automatically stores time series data. Use `query_range` API for historical queries.

| Time Range | Step Interval | Data Points | PromQL Example |
|------------|--------------|-------------|----------------|
| 1 hour | 1 minute | 60 | `avg_over_time(biya_tps_current[1m])` |
| 6 hours | 5 minutes | 72 | `avg_over_time(biya_tps_current[5m])` |
| 24 hours | 20 minutes | 72 | `avg_over_time(biya_tps_current[20m])` |

### 4.2 Trend Metrics

Same as Module 1 & 3 metrics, queried with range:
- `biya_tps_current`
- `biya_tx_confirm_time_avg_seconds`
- `biya_gas_utilization_ratio`
- `biya_mempool_size`

---

## Module 5: Governance (网络治理)

### 5.1 Governance Aggregate Metrics

| Metric Name | Type | Labels | Description | Data Source |
|-------------|------|--------|-------------|-------------|
| `biya_proposals_total` | Counter | - | Total proposals created | biya-stake |
| `biya_proposals_passed` | Gauge | - | Total passed proposals | biya-stake |
| `biya_proposals_rejected` | Gauge | - | Total rejected proposals | biya-stake |
| `biya_proposals_active` | Gauge | - | Currently active (voting) proposals | biya-stake |
| `biya_voting_power_total` | Gauge | - | Total voting power | biya-stake |
| `biya_participation_rate_avg` | Gauge | - | Average participation rate | biya-stake |

### 5.2 Individual Proposal Metrics (Optional)

| Metric Name | Type | Labels | Description | Data Source |
|-------------|------|--------|-------------|-------------|
| `biya_proposal_status` | Gauge | `id`, `title` | Proposal status (0-5 enum) | biya-stake |
| `biya_proposal_votes_yes` | Gauge | `id` | Yes votes | biya-stake |
| `biya_proposal_votes_no` | Gauge | `id` | No votes | biya-stake |
| `biya_proposal_votes_veto` | Gauge | `id` | NoWithVeto votes | biya-stake |
| `biya_proposal_votes_abstain` | Gauge | `id` | Abstain votes | biya-stake |

---

## Alert Rules (Prometheus/Alertmanager)

### 6.1 Critical Alerts

```yaml
groups:
  - name: biya_critical
    rules:
      # Block production stopped
      - alert: BlockProductionStopped
        expr: increase(biya_block_height[5m]) == 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Block production has stopped"
          
      # Node offline
      - alert: ValidatorOffline
        expr: time() - biya_validator_last_active_timestamp > 600
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Validator {{ $labels.moniker }} is offline"
```

### 6.2 Warning Alerts

```yaml
      # Low transaction success rate
      - alert: LowTxSuccessRate
        expr: biya_tx_success_rate < 0.95
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Transaction success rate is below 95%"
          
      # High gas price
      - alert: HighGasPrice
        expr: biya_gas_price_gwei > 100
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Gas price is unusually high"
          
      # Mempool congestion
      - alert: MempoolCongestion
        expr: biya_congestion_ratio > 0.8
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Mempool congestion above 80%"
          
      # Node sync behind
      - alert: NodeSyncBehind
        expr: biya_node_behind_blocks > 100
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Node {{ $labels.node }} is {{ $value }} blocks behind"
```

### 6.3 Performance Alerts

```yaml
      # Low performance (calculated in Gin, can also alert via Prometheus)
      - alert: LowPerformance
        expr: (
          (biya_mempool_size / biya_tps_current) > 60
        )
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Network performance is degraded (mempool clearance > 60s)"
```

---

## Recording Rules (Pre-aggregation)

```yaml
groups:
  - name: biya_recording
    rules:
      # 5-minute average TPS
      - record: biya_tps_5m_avg
        expr: avg_over_time(biya_tps_current[5m])
        
      # 1-hour average TPS
      - record: biya_tps_1h_avg
        expr: avg_over_time(biya_tps_current[1h])
        
      # 24-hour average TPS
      - record: biya_tps_24h_avg
        expr: avg_over_time(biya_tps_current[24h])
        
      # Transaction success rate (1 hour)
      - record: biya_tx_success_rate_1h
        expr: |
          (
            increase(biya_tx_total{status="success"}[1h])
            /
            increase(biya_tx_total[1h])
          )
          
      # Active validators ratio
      - record: biya_validators_active_ratio
        expr: biya_validators_active / biya_validators_consensus
        
      # Network congestion ratio
      - record: biya_congestion_ratio
        expr: biya_mempool_size / biya_mempool_capacity
```

---

## Data Collection Summary

### Exporter Collection Sources

| Source | Endpoint | Metrics Collected | Collection Interval |
|--------|----------|-------------------|---------------------|
| **biya-explorer** | HTTP API | Block, TX, Gas, Address | 10s |
| **biya-stake** | HTTP API | Validator, Staking, Governance | 30s |
| **injective-core** | Tendermint RPC | Mempool, Node Sync | 10s |

### Metric Count Summary

| Module | Metric Count |
|--------|--------------|
| Runtime Status Monitoring | 20 |
| Node Management | 18 |
| Network Performance | 4 (reused from above) |
| Network Trends | 4 (reused from above) |
| Governance | 10 |
| **Total Unique Metrics** | ~48 |

---

## Implementation Notes

1. **Exporter is responsible for collection** - Gin backend does NOT collect metrics
2. **Gin backend queries Prometheus** - Use PromQL for all data access
3. **Performance index calculated in Gin** - Not stored in Prometheus
4. **Redis caches Prometheus results** - Reduce query load
5. **Labels for validators** - Use `address` and `moniker` for identification

