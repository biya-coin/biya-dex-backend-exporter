# 告警系统架构说明

## 系统架构图

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Biya Chain 告警系统                          │
└─────────────────────────────────────────────────────────────────────┘

┌──────────────┐    HTTP       ┌──────────────┐    Alert      ┌─────────────────┐
│              │  ──────────>  │              │  ──────────>  │                 │
│  Biya Chain  │   Metrics     │  Prometheus  │   Firing      │  Alertmanager   │
│   Exporter   │   Endpoint    │              │   Alerts      │                 │
│              │  <──────────  │              │  <──────────  │                 │
└──────────────┘   /metrics    └──────────────┘   Resolved    └─────────────────┘
                                      │                               │
                                      │ Rule                          │ Notifications
                                      │ Evaluation                    │
                                      ▼                               ▼
                              ┌──────────────┐              ┌─────────────────┐
                              │ Alert Rules  │              │   Receivers     │
                              │              │              │                 │
                              │ - TPS Drop   │              │ - Email (SMTP)  │
                              │ - Latency    │              │ - WeChat        │
                              │ - Success    │              │ - DingTalk      │
                              │   Rate       │              │ - Slack         │
                              │ - ...        │              │ - PagerDuty     │
                              └──────────────┘              │ - Webhook       │
                                                            └─────────────────┘
```

## 数据流向

### 1. 指标采集流程

```
Biya Chain → Exporter → Prometheus → Time Series Database
   (Chain)    (采集)      (存储)          (TSDB)

详细步骤：
1. Exporter 定期调用 Biya Chain API 获取链上数据
2. Exporter 将数据转换为 Prometheus 格式的指标
3. Exporter 暴露 /metrics 端点
4. Prometheus 定期 scrape Exporter 的 /metrics 端点
5. Prometheus 将指标数据存储到 TSDB
```

### 2. 告警评估流程

```
Alert Rules → PromQL Query → Evaluation → Alert State
  (规则)        (查询)        (评估)       (状态)

详细步骤：
1. Prometheus 按照 evaluation_interval (默认30秒) 评估所有规则
2. 对每条规则执行 PromQL 查询
3. 如果查询结果非空，进入 Pending 状态
4. 如果 Pending 持续时间达到 for 指定的时间，进入 Firing 状态
5. 如果查询结果为空，且之前是 Firing 状态，进入 Resolved 状态
```

### 3. 告警通知流程

```
Firing Alert → Alertmanager → Routing → Grouping → Notification
  (触发)         (接收)        (路由)     (分组)      (通知)

详细步骤：
1. Prometheus 将 Firing/Resolved 告警发送到 Alertmanager
2. Alertmanager 根据 route 配置进行路由匹配
3. 根据 group_by 标签对告警进行分组
4. 应用 inhibit_rules 抑制低优先级告警
5. 检查是否存在匹配的 silence（静默规则）
6. 发送通知到配置的 receivers
7. 等待 repeat_interval 后，如果告警仍在，重复发送
```

## 告警状态机

```
                  ┌─────────┐
                  │ Inactive│  (告警规则不满足)
                  └────┬────┘
                       │
        expr 返回结果  │
                       ▼
                  ┌─────────┐
             ┌────│ Pending │────┐
             │    └─────────┘    │
             │                   │
   expr 不满足│                  │ 持续时间达到 for
             │                   │
             ▼                   ▼
        ┌─────────┐         ┌─────────┐
        │ Inactive│         │ Firing  │  (发送告警)
        └─────────┘         └────┬────┘
                                 │
                       expr 不满足│
                                 ▼
                            ┌─────────┐
                            │Resolved │  (发送恢复通知)
                            └─────────┘
```

## 告警规则分层

### 第一层：原始指标（Raw Metrics）

```
biya_tps_current                     ← Exporter 直接导出
biya_tps_24h_avg                     ← Exporter 直接导出
biya_tx_success_rate                 ← Exporter 直接导出
biya_tx_confirm_time_seconds_avg     ← Exporter 直接导出
```

### 第二层：记录规则（Recording Rules）

```
biya_tps_7d_avg                      ← 预计算的 7 天平均值
  = avg_over_time(biya_tps_24h_avg[7d])

biya_tps_drop_percentage             ← 预计算的 TPS 下降百分比
  = ((7d_avg - current) / 7d_avg) * 100

biya_performance_health_score        ← 预计算的性能健康度
  = 综合计算公式
```

### 第三层：告警规则（Alerting Rules）

```
TPSAbnormalDrop                      ← 使用记录规则
  = biya_tps_current < (biya_tps_7d_avg * 0.5)

NetworkLatencyAbnormalIncrease       ← 直接使用原始指标
  = biya_tx_confirm_time_seconds_avg > 10

PerformanceComprehensiveAbnormal     ← 组合多个条件
  = (TPS下降) AND (高延迟) AND (低成功率)
```

## 告警路由决策树

```
                            告警产生
                               │
                               ▼
                     ┌──────────────────┐
                     │  severity 匹配？  │
                     └──────────────────┘
                               │
              ┌────────────────┼────────────────┐
              │                │                │
              ▼                ▼                ▼
        emergency          critical          warning
              │                │                │
              ▼                ▼                ▼
     ┌─────────────┐   ┌─────────────┐   ┌─────────────┐
     │emergency-team│   │ oncall-team │   │  dev-team   │
     └─────────────┘   └─────────────┘   └─────────────┘
     │                 │                 │
     │ 10s等待         │ 20s等待         │ 30s等待
     │ 1m间隔          │ 2m间隔          │ 5m间隔
     │ 30m重复         │ 1h重复          │ 4h重复
     │                 │                 │
     ▼                 ▼                 ▼
┌─────────────┐   ┌─────────────┐   ┌─────────────┐
│电话+短信+IM │   │ 短信+IM+邮件│   │  IM+邮件    │
└─────────────┘   └─────────────┘   └─────────────┘
```

## 告警抑制逻辑

```
┌─────────────────────────────────────┐
│ PerformanceComprehensiveAbnormal    │ (综合异常)
└─────────────────────────────────────┘
                 │
                 │ 触发后抑制以下告警：
                 │
    ┌────────────┼────────────┐
    │            │            │
    ▼            ▼            ▼
┌────────┐  ┌────────┐  ┌────────┐
│TPS Drop│  │Latency │  │Success │
│        │  │Increase│  │Rate    │
└────────┘  └────────┘  └────────┘
   (被抑制)    (被抑制)    (被抑制)
```

**抑制规则说明**：
- 当综合异常告警触发时，单个指标的告警会被抑制
- 避免同一问题产生多个告警通知
- 减少告警噪音，提高告警质量

## 时间参数说明

### Prometheus 配置

```yaml
global:
  scrape_interval: 15s       # 每 15 秒采集一次指标
  evaluation_interval: 30s   # 每 30 秒评估一次规则

rule_files:
  - alert_rules.yml
```

### 规则组配置

```yaml
groups:
  - name: chain_performance_alerts
    interval: 30s              # 规则组评估间隔（覆盖全局配置）
    rules:
      - alert: Example
        expr: metric > threshold
        for: 5m                # 持续 5 分钟才触发
```

### Alertmanager 配置

```yaml
route:
  group_wait: 30s              # 分组等待时间
  group_interval: 5m           # 分组通知间隔
  repeat_interval: 4h          # 重复通知间隔
```

### 时间线示例

```
时间轴：
0s    15s   30s   45s   60s   ...   5m00s  5m15s
│     │     │     │     │            │      │
├─────┼─────┼─────┼─────┼────...─────┼──────┤
│     │     │     │     │            │      │
Scrape Scrape Eval Scrape Eval ... Scrape  Alert!
                                     Eval    Firing
                                     (5min)

说明：
- Scrape: 采集指标（每15秒）
- Eval: 评估规则（每30秒）
- 当条件持续满足 5 分钟后，告警进入 Firing 状态
```

## 性能考虑

### 1. 查询性能优化

**使用记录规则**：
```yaml
# ❌ 直接在告警规则中使用复杂查询
- alert: TPSDrop
  expr: biya_tps_current < (avg_over_time(biya_tps_24h_avg[7d]) * 0.5)

# ✅ 使用记录规则预计算
- record: biya_tps_7d_avg
  expr: avg_over_time(biya_tps_24h_avg[7d])

- alert: TPSDrop
  expr: biya_tps_current < (biya_tps_7d_avg * 0.5)
```

**优势**：
- 减少重复计算
- 降低查询延迟
- 提高系统稳定性

### 2. 资源消耗

| 组件 | 内存 | CPU | 磁盘 |
|------|------|-----|------|
| Exporter | ~50MB | 低 | - |
| Prometheus | ~1GB | 中 | 15GB (15天) |
| Alertmanager | ~50MB | 低 | ~100MB |

### 3. 扩展性

**横向扩展**：
- Prometheus Federation（联邦）
- Thanos（长期存储）
- Cortex（多租户）

**纵向优化**：
- 调整 retention 时间
- 使用 Recording Rules
- 优化标签基数

## 高可用架构（可选）

```
┌────────────────────────────────────────────────────────────┐
│                      高可用部署                             │
└────────────────────────────────────────────────────────────┘

         ┌─────────────┐          ┌─────────────┐
         │ Prometheus  │          │ Prometheus  │
         │  Instance 1 │          │  Instance 2 │
         └──────┬──────┘          └──────┬──────┘
                │                        │
                └────────┬───────────────┘
                         │ 都发送告警到
                         ▼
                ┌─────────────────┐
                │  Alertmanager   │
                │    Cluster      │
                │  ┌───┬───┬───┐  │
                │  │AM1│AM2│AM3│  │ (去重、分组)
                │  └───┴───┴───┘  │
                └─────────────────┘
                         │
                         ▼
                  [Receivers]
```

**特点**：
- 多个 Prometheus 实例采集相同的目标
- Alertmanager 集群自动去重
- 提供高可用性和容错能力

## 监控告警系统本身

### 元监控告警

```yaml
# 监控 Prometheus 是否在线
- alert: PrometheusDown
  expr: up{job="prometheus"} == 0
  for: 5m

# 监控 Alertmanager 是否在线
- alert: AlertmanagerDown
  expr: up{job="alertmanager"} == 0
  for: 5m

# 监控规则评估失败
- alert: RuleEvaluationFailures
  expr: increase(prometheus_rule_evaluation_failures_total[5m]) > 0
  for: 5m

# 监控告警发送失败
- alert: AlertmanagerNotificationsFailing
  expr: rate(alertmanager_notifications_failed_total[5m]) > 0
  for: 10m
```

## 安全考虑

### 1. 访问控制

- Prometheus 和 Alertmanager 应配置认证
- 使用反向代理（如 Nginx）添加 Basic Auth
- 考虑使用 OAuth2 Proxy

### 2. 通知安全

- SMTP 使用 TLS 加密
- Webhook 使用 HTTPS
- API Key 和密码存储在环境变量或密钥管理系统

### 3. 网络隔离

- 内部服务使用 Docker 网络隔离
- 仅暴露必要的端口
- 使用防火墙规则限制访问

## 故障场景分析

### 场景1：Exporter 宕机

```
Exporter 宕机
    │
    ▼
Prometheus 无法采集指标
    │
    ▼
指标数据缺失 (absent() 返回 true)
    │
    ▼
MetricMissing 告警触发
    │
    ▼
通知运维团队
```

### 场景2：Alertmanager 宕机

```
Alertmanager 宕机
    │
    ▼
Prometheus 无法发送告警
    │
    ▼
/alertmanagers 页面显示 down
    │
    ▼
告警仍然在 Prometheus 中触发
    │
    ▼
需要手动监控或部署 HA
```

### 场景3：网络分区

```
网络分区
    │
    ├─> Prometheus 无法访问 Exporter
    │       │
    │       ▼
    │   指标采集失败
    │
    └─> Alertmanager 无法发送通知
            │
            ▼
        通知失败累积
```

## 参考资料

- [Prometheus 架构](https://prometheus.io/docs/introduction/overview/)
- [Alertmanager 架构](https://prometheus.io/docs/alerting/latest/alertmanager/)
- [高可用方案](https://prometheus.io/docs/introduction/faq/#high-availability)
- [最佳实践](https://prometheus.io/docs/practices/)

---

**最后更新**: 2025-12-24

