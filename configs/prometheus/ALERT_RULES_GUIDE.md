# Prometheus 告警规则编写指南

本文档提供了编写 Biya Chain 告警规则的详细指南和最佳实践。

## 目录

- [告警规则基础](#告警规则基础)
- [PromQL 常用表达式](#promql-常用表达式)
- [告警规则示例](#告警规则示例)
- [最佳实践](#最佳实践)
- [常见问题](#常见问题)

## 告警规则基础

### 告警规则结构

```yaml
groups:
  - name: 规则组名称
    interval: 评估间隔（可选，默认使用全局配置）
    rules:
      - alert: 告警名称
        expr: PromQL 表达式
        for: 持续时间
        labels:
          severity: 告警级别
          category: 告警分类
        annotations:
          summary: 告警摘要
          description: 告警详细描述
          处理建议: 处理步骤
```

### 关键字段说明

| 字段 | 必需 | 说明 |
|------|------|------|
| alert | 是 | 告警规则名称，应该清晰描述告警内容 |
| expr | 是 | PromQL 表达式，当结果非空时触发告警 |
| for | 否 | 持续时间，表达式需要持续满足多久才触发告警 |
| labels | 否 | 附加标签，用于路由和分组 |
| annotations | 否 | 注释信息，提供告警的详细描述 |

## PromQL 常用表达式

### 1. 基础操作符

```promql
# 比较操作符
metric > 100      # 大于
metric < 100      # 小于
metric >= 100     # 大于等于
metric <= 100     # 小于等于
metric == 100     # 等于
metric != 100     # 不等于

# 逻辑操作符
expr1 and expr2   # 与
expr1 or expr2    # 或
expr1 unless expr2 # 差集
```

### 2. 聚合函数

```promql
# 求和
sum(metric)
sum by (label) (metric)

# 平均值
avg(metric)
avg by (label) (metric)

# 最小值/最大值
min(metric)
max(metric)

# 计数
count(metric)
```

### 3. 时间序列函数

```promql
# 变化率
rate(counter[5m])           # 每秒平均增长率
irate(counter[5m])          # 瞬时增长率
increase(counter[5m])       # 时间段内的总增量

# 时间窗口聚合
avg_over_time(metric[1h])   # 1小时内的平均值
max_over_time(metric[1h])   # 1小时内的最大值
min_over_time(metric[1h])   # 1小时内的最小值
sum_over_time(metric[1h])   # 1小时内的总和

# 预测
predict_linear(metric[1h], 3600)  # 基于1小时数据预测1小时后的值
```

### 4. 数学函数

```promql
abs(metric)                 # 绝对值
ceil(metric)                # 向上取整
floor(metric)               # 向下取整
round(metric)               # 四舍五入

clamp_min(metric, 0)        # 设置最小值
clamp_max(metric, 100)      # 设置最大值
```

### 5. 标签操作

```promql
# 过滤标签
metric{label="value"}
metric{label=~"regex"}      # 正则匹配
metric{label!="value"}      # 不等于
metric{label!~"regex"}      # 正则不匹配

# 向量匹配
metric1 + on(label) metric2
metric1 / ignoring(label) metric2
```

## 告警规则示例

### 示例1：简单阈值告警

```yaml
- alert: HighTPS
  expr: biya_tps_current > 1000
  for: 5m
  labels:
    severity: info
  annotations:
    summary: "TPS 持续高位"
    description: "当前 TPS ({{ $value }}) 持续超过 1000"
```

### 示例2：百分比变化告警

```yaml
- alert: TPSDropped50Percent
  expr: |
    (
      (biya_tps_24h_avg - biya_tps_current) 
      / 
      biya_tps_24h_avg
    ) * 100 > 50
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "TPS 下降超过 50%"
    description: "TPS 从 {{ with query \"biya_tps_24h_avg\" }}{{ . | first | value | humanize }}{{ end }} 下降到 {{ $value | humanize }}"
```

### 示例3：时间窗口比较告警

```yaml
- alert: TPSBelowHistoricalAverage
  expr: |
    biya_tps_current < (avg_over_time(biya_tps_24h_avg[7d]) * 0.5)
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "TPS 低于历史平均值"
    description: "当前 TPS ({{ $value }}) 低于 7 天平均值的 50%"
```

### 示例4：多条件组合告警

```yaml
- alert: SystemDegraded
  expr: |
    (
      biya_tps_current < 100
      and
      biya_tx_success_rate < 0.95
      and
      biya_tx_confirm_time_seconds_avg > 30
    )
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "系统性能严重下降"
    description: |
      检测到多个性能指标异常：
      - TPS: {{ with query "biya_tps_current" }}{{ . | first | value }}{{ end }}
      - 成功率: {{ with query "biya_tx_success_rate" }}{{ . | first | value | humanizePercentage }}{{ end }}
      - 延迟: {{ with query "biya_tx_confirm_time_seconds_avg" }}{{ . | first | value }}{{ end }}秒
```

### 示例5：缺失数据告警

```yaml
- alert: MetricMissing
  expr: |
    absent(biya_tps_current)
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "TPS 指标缺失"
    description: "超过 5 分钟未收到 TPS 指标数据"
```

### 示例6：增长率告警

```yaml
- alert: RapidTxGrowth
  expr: |
    rate(biya_tx_total[5m]) > 1000
  for: 10m
  labels:
    severity: info
  annotations:
    summary: "交易量快速增长"
    description: "过去 5 分钟交易增长率为 {{ $value | humanize }}/秒"
```

### 示例7：预测性告警

```yaml
- alert: MemPoolWillBeFull
  expr: |
    predict_linear(biya_chain_mempool_pending_txs[1h], 3600) > 50000
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "预测交易池将在 1 小时内饱和"
    description: "根据当前增长趋势，交易池预计在 1 小时后达到 {{ $value }}"
```

### 示例8：标签分组告警

```yaml
- alert: ValidatorDowntime
  expr: |
    biya_stake_validators_uptime_percentage_avg < 0.99
  for: 15m
  labels:
    severity: critical
    category: validator
  annotations:
    summary: "验证者在线率下降"
    description: "验证者平均在线率 ({{ $value | humanizePercentage }}) 低于 99%"
```

## 最佳实践

### 1. 告警命名规范

- ✅ 使用描述性名称：`TPSAbnormalDrop` 而不是 `Alert1`
- ✅ 使用驼峰命名：`NetworkLatencyHigh`
- ✅ 避免使用缩写（除非是通用缩写如 TPS）
- ✅ 按层次组织：`<Component><Metric><Condition>`

示例：
```
✅ 好的命名：
- TPSAbnormalDrop
- TransactionSuccessRateLow
- MempoolCongestion

❌ 不好的命名：
- alert1
- problem
- tps_alert
```

### 2. 设置合理的持续时间（for）

```yaml
# ❌ 太短，容易误报
- alert: TPSLow
  expr: biya_tps_current < 100
  for: 30s

# ✅ 合理的持续时间
- alert: TPSLow
  expr: biya_tps_current < 100
  for: 5m

# ✅ 严重问题可以更短
- alert: SystemDown
  expr: up == 0
  for: 1m
```

### 3. 避免告警疲劳

```yaml
# ❌ 阈值设置太敏感
- alert: LatencySlightlyHigh
  expr: latency > 1
  for: 1m

# ✅ 设置有意义的阈值
- alert: LatencyHigh
  expr: latency > 10
  for: 5m

# ✅ 使用分层告警
- alert: LatencyWarning
  expr: latency > 5
  for: 10m
  labels:
    severity: warning

- alert: LatencyCritical
  expr: latency > 20
  for: 5m
  labels:
    severity: critical
```

### 4. 提供有用的注释

```yaml
# ❌ 信息不足
annotations:
  summary: "出问题了"

# ✅ 提供详细信息
annotations:
  summary: "TPS 异常下降"
  description: "当前 TPS ({{ $value }}) 低于历史平均值的 50%，已持续 {{ $labels.alertstate }} 分钟"
  处理建议: |
    1. 检查节点运行状态
    2. 查看网络连接
    3. 分析交易来源
  runbook_url: "https://docs.biya.chain/runbooks/tps-drop"
  dashboard: "https://grafana.biya.chain/d/performance"
```

### 5. 使用录制规则优化性能

```yaml
# ❌ 在告警规则中使用复杂计算
- alert: PerformanceDegraded
  expr: |
    (biya_tps_current / avg_over_time(biya_tps_24h_avg[7d])) < 0.5
  for: 5m

# ✅ 使用录制规则预计算
# 录制规则
- record: biya_tps_7d_avg
  expr: avg_over_time(biya_tps_24h_avg[7d])

# 告警规则
- alert: PerformanceDegraded
  expr: biya_tps_current < (biya_tps_7d_avg * 0.5)
  for: 5m
```

### 6. 告警分级

```yaml
# 级别 1: 信息（info）- 无需立即处理
- alert: HighTrafficDetected
  labels:
    severity: info

# 级别 2: 警告（warning）- 需要关注
- alert: TPSSlowDown
  labels:
    severity: warning

# 级别 3: 严重（critical）- 需要立即处理
- alert: TransactionFailureRateHigh
  labels:
    severity: critical

# 级别 4: 紧急（emergency）- 需要紧急响应
- alert: SystemCompletelyDown
  labels:
    severity: emergency
```

### 7. 使用标签组织告警

```yaml
labels:
  severity: warning           # 告警级别
  category: performance      # 告警类别
  subsystem: transaction     # 子系统
  team: backend             # 负责团队
  environment: production   # 环境
```

## 常见问题

### Q1: 如何避免告警抖动？

使用 `for` 字段设置持续时间：

```yaml
- alert: TPSFluctuation
  expr: biya_tps_current < 100
  for: 5m  # 需要持续 5 分钟才触发
```

### Q2: 如何在告警中使用变量？

使用模板语法：

```yaml
annotations:
  description: "当前值: {{ $value }}"
  # 查询其他指标
  additional_info: "平均值: {{ with query \"avg(biya_tps_current)\" }}{{ . | first | value }}{{ end }}"
```

### Q3: 如何测试告警规则？

```bash
# 1. 验证语法
promtool check rules alert_rules.yml

# 2. 测试表达式
curl "http://localhost:9090/api/v1/query?query=YOUR_EXPRESSION"

# 3. 查看告警状态
curl http://localhost:9090/api/v1/alerts
```

### Q4: 告警规则不触发怎么办？

检查清单：
1. ✅ 表达式是否正确？在 Prometheus UI 中测试
2. ✅ 数据是否存在？查询基础指标
3. ✅ 持续时间是否太长？检查 `for` 字段
4. ✅ 规则是否已加载？访问 `/rules` 页面
5. ✅ Alertmanager 是否连接？检查 `/alertmanagers` 页面

### Q5: 如何临时禁用告警？

方法 1：在 Alertmanager 中添加静默规则
```bash
amtool silence add alertname=TPSAbnormalDrop --duration=2h
```

方法 2：注释掉规则文件中的规则
```yaml
# - alert: TPSAbnormalDrop
#   expr: ...
```

方法 3：使用路由静默
```yaml
# alertmanager.yml
route:
  routes:
    - match:
        alertname: TPSAbnormalDrop
      receiver: 'null'
```

## PromQL 函数速查表

### 时间函数

| 函数 | 说明 | 示例 |
|------|------|------|
| `rate()` | 计算每秒平均增长率 | `rate(requests[5m])` |
| `irate()` | 计算瞬时增长率 | `irate(requests[5m])` |
| `increase()` | 计算时间段内增量 | `increase(requests[1h])` |
| `avg_over_time()` | 时间窗口平均值 | `avg_over_time(metric[1h])` |
| `sum_over_time()` | 时间窗口总和 | `sum_over_time(metric[1h])` |
| `min_over_time()` | 时间窗口最小值 | `min_over_time(metric[1h])` |
| `max_over_time()` | 时间窗口最大值 | `max_over_time(metric[1h])` |
| `delta()` | 时间窗口差值 | `delta(metric[1h])` |
| `deriv()` | 时间窗口导数 | `deriv(metric[5m])` |

### 聚合函数

| 函数 | 说明 | 示例 |
|------|------|------|
| `sum()` | 求和 | `sum(metric)` |
| `avg()` | 平均值 | `avg(metric)` |
| `min()` | 最小值 | `min(metric)` |
| `max()` | 最大值 | `max(metric)` |
| `count()` | 计数 | `count(metric)` |
| `stddev()` | 标准差 | `stddev(metric)` |
| `topk()` | 前 K 个最大值 | `topk(5, metric)` |
| `bottomk()` | 前 K 个最小值 | `bottomk(5, metric)` |
| `quantile()` | 分位数 | `quantile(0.95, metric)` |

### 数学函数

| 函数 | 说明 | 示例 |
|------|------|------|
| `abs()` | 绝对值 | `abs(metric)` |
| `ceil()` | 向上取整 | `ceil(metric)` |
| `floor()` | 向下取整 | `floor(metric)` |
| `round()` | 四舍五入 | `round(metric)` |
| `clamp_min()` | 最小值限制 | `clamp_min(metric, 0)` |
| `clamp_max()` | 最大值限制 | `clamp_max(metric, 100)` |
| `ln()` | 自然对数 | `ln(metric)` |
| `log2()` | 以 2 为底的对数 | `log2(metric)` |
| `log10()` | 以 10 为底的对数 | `log10(metric)` |

### 预测和趋势

| 函数 | 说明 | 示例 |
|------|------|------|
| `predict_linear()` | 线性预测 | `predict_linear(metric[1h], 3600)` |
| `holt_winters()` | Holt-Winters 平滑 | `holt_winters(metric[1d], 0.3, 0.1)` |

## 相关资源

- [Prometheus 查询语言](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [告警规则最佳实践](https://prometheus.io/docs/practices/alerting/)
- [PromQL 函数参考](https://prometheus.io/docs/prometheus/latest/querying/functions/)
- [Recording Rules](https://prometheus.io/docs/prometheus/latest/configuration/recording_rules/)

---

**提示**: 编写告警规则时，始终记住：**好的告警应该是可操作的、有意义的，并且不会造成告警疲劳。**

