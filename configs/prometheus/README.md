# Prometheus 告警配置说明

本目录包含 Biya Chain 监控系统的 Prometheus 和 Alertmanager 配置文件。

## 文件说明

### 1. prometheus.yml
Prometheus 主配置文件，包含：
- 数据采集配置（scrape_configs）
- 告警规则文件路径（rule_files）
- Alertmanager 连接配置（alerting）

### 2. alert_rules.yml
Prometheus 告警规则文件，包含以下告警规则组：

#### chain_performance_alerts（链性能告警）
根据产品需求文档定义的4个核心告警规则：

| 告警名称 | 触发条件 | 持续时间 | 级别 | 说明 |
|---------|---------|---------|------|------|
| TPSAbnormalDrop | TPS < 历史均值的50% | 5分钟 | warning | TPS异常下降 |
| NetworkLatencyAbnormalIncrease | 延迟 > 10秒 | 10分钟 | warning | 网络延迟异常 |
| TransactionSuccessRateDrop | 成功率 < 98% | 5分钟 | critical | 交易成功率下降 |
| PerformanceComprehensiveAbnormal | 多指标同时异常 | 3分钟 | emergency | 性能综合异常 |

#### chain_performance_recording_rules（性能记录规则）
预计算的指标，用于优化查询性能：
- `biya_tps_7d_avg`: 7天TPS平均值
- `biya_tps_drop_percentage`: TPS下降百分比
- `biya_performance_health_score`: 性能健康度评分（0-100）

#### chain_performance_threshold_alerts（阈值告警）
额外的阈值监控告警：
- `TPSZero`: TPS为零告警
- `MempoolCongestion`: 交易池拥堵告警
- `HighGasUtilization`: Gas利用率过高告警
- `BlockTimeAbnormal`: 出块时间异常告警
- `NetworkCongestion`: 网络拥堵告警

### 3. alertmanager.yml
Alertmanager 配置文件，包含：
- 全局配置（SMTP等）
- 告警路由规则（route）
- 告警抑制规则（inhibit_rules）
- 接收者配置（receivers）
- 时间窗口配置（time_intervals）

## 告警级别说明

| 级别 | 标签值 | 响应时间 | 通知方式 | 适用场景 |
|------|--------|---------|---------|---------|
| 紧急 | emergency | 立即 | 电话+短信+IM+邮件 | 多指标异常，需要紧急处理 |
| 严重 | critical | < 5分钟 | 短信+IM+邮件 | 单一关键指标异常 |
| 警告 | warning | < 30分钟 | IM+邮件 | 需要关注但不紧急 |
| 信息 | info | < 2小时 | 邮件 | 一般性提醒 |

## 快速开始

### 方式1：使用 Docker Compose（推荐）

1. **更新 compose.yaml**，添加 Alertmanager 服务：

```yaml
services:
  prometheus:
    image: prom/prometheus:latest
    container_name: biya-prometheus
    volumes:
      - ./configs/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
      - ./configs/prometheus/alert_rules.yml:/etc/prometheus/alert_rules.yml
      - prometheus-data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/usr/share/prometheus/console_libraries'
      - '--web.console.templates=/usr/share/prometheus/consoles'
      - '--web.enable-lifecycle'
    ports:
      - "9090:9090"
    networks:
      - biya-network
    restart: unless-stopped

  alertmanager:
    image: prom/alertmanager:latest
    container_name: biya-alertmanager
    volumes:
      - ./configs/prometheus/alertmanager.yml:/etc/alertmanager/alertmanager.yml
      - alertmanager-data:/alertmanager
    command:
      - '--config.file=/etc/alertmanager/alertmanager.yml'
      - '--storage.path=/alertmanager'
    ports:
      - "9093:9093"
    networks:
      - biya-network
    restart: unless-stopped

volumes:
  prometheus-data:
  alertmanager-data:
```

2. **启动服务**：

```bash
docker-compose up -d prometheus alertmanager
```

3. **验证服务**：

- Prometheus UI: http://localhost:9090
- Prometheus 告警页面: http://localhost:9090/alerts
- Alertmanager UI: http://localhost:9093

### 方式2：本地运行

1. **安装 Prometheus**：

```bash
# macOS
brew install prometheus

# Linux
wget https://github.com/prometheus/prometheus/releases/download/v2.45.0/prometheus-2.45.0.linux-amd64.tar.gz
tar xvfz prometheus-*.tar.gz
cd prometheus-*
```

2. **启动 Prometheus**：

```bash
./prometheus --config.file=/path/to/configs/prometheus/prometheus.yml
```

3. **安装并启动 Alertmanager**：

```bash
# macOS
brew install alertmanager

# Linux
wget https://github.com/prometheus/alertmanager/releases/download/v0.25.0/alertmanager-0.25.0.linux-amd64.tar.gz
tar xvfz alertmanager-*.tar.gz
cd alertmanager-*

# 启动
./alertmanager --config.file=/path/to/configs/prometheus/alertmanager.yml
```

## 配置告警接收者

### 1. 配置邮件通知

编辑 `alertmanager.yml`，修改 SMTP 配置：

```yaml
global:
  smtp_smarthost: 'smtp.gmail.com:587'
  smtp_from: 'alerts@yourdomain.com'
  smtp_auth_username: 'alerts@yourdomain.com'
  smtp_auth_password: 'your-app-password'
  smtp_require_tls: true
```

### 2. 配置企业微信通知

```yaml
receivers:
  - name: 'wechat-receiver'
    wechat_configs:
      - corp_id: 'your-corp-id'
        to_party: 'party-id'
        agent_id: 'agent-id'
        api_secret: 'your-api-secret'
```

### 3. 配置钉钉通知

```yaml
receivers:
  - name: 'dingtalk-receiver'
    webhook_configs:
      - url: 'https://oapi.dingtalk.com/robot/send?access_token=your-token'
```

### 4. 配置 Slack 通知

```yaml
receivers:
  - name: 'slack-receiver'
    slack_configs:
      - api_url: 'https://hooks.slack.com/services/YOUR/WEBHOOK/URL'
        channel: '#alerts'
```

## 告警规则调整

### 调整告警阈值

编辑 `alert_rules.yml`，修改表达式中的阈值：

```yaml
# 例如：将 TPS 下降阈值从 50% 调整为 30%
expr: |
  biya_tps_current < (avg_over_time(biya_tps_24h_avg[7d]) * 0.3)
```

### 调整持续时间

修改 `for` 字段：

```yaml
# 从 5 分钟改为 3 分钟
for: 3m
```

### 添加新的告警规则

```yaml
- alert: YourNewAlert
  expr: your_metric > threshold
  for: 5m
  labels:
    severity: warning
    category: custom
  annotations:
    summary: "告警摘要"
    description: "详细描述"
    处理建议: "处理步骤"
```

## 告警测试

### 1. 验证规则语法

```bash
# 使用 promtool 验证规则文件
promtool check rules configs/prometheus/alert_rules.yml

# 验证 Alertmanager 配置
amtool check-config configs/prometheus/alertmanager.yml
```

### 2. 触发测试告警

```bash
# 使用 amtool 发送测试告警
amtool alert add test_alert alertname=TestAlert severity=warning
```

### 3. 查看活跃告警

```bash
# 通过 API 查看
curl http://localhost:9090/api/v1/alerts

# 或访问 Web UI
# http://localhost:9090/alerts
```

## 告警静默

### 通过 Web UI 静默

访问 http://localhost:9093 -> Silences -> New Silence

### 通过命令行静默

```bash
# 静默特定告警
amtool silence add alertname=TPSAbnormalDrop \
  --duration=2h \
  --comment="Maintenance window"

# 查看静默列表
amtool silence query

# 删除静默
amtool silence expire <silence-id>
```

## 监控指标说明

### 核心指标

| 指标名称 | 类型 | 说明 |
|---------|------|------|
| `biya_tps_current` | Gauge | 当前 TPS |
| `biya_tps_24h_avg` | Gauge | 24小时平均 TPS |
| `biya_tx_confirm_time_seconds_avg` | Gauge | 平均交易确认时间（秒） |
| `biya_tx_success_rate` | Gauge | 交易成功率（0-1） |
| `biya_chain_mempool_pending_txs` | Gauge | 待处理交易数 |
| `biya_chain_block_gas_utilization_ratio_avg` | Gauge | 平均 Gas 利用率（0-1） |
| `biya_chain_congestion_ratio` | Gauge | 拥堵指数（0-1） |

### 计算指标

| 指标名称 | 计算公式 | 说明 |
|---------|---------|------|
| `biya_tps_7d_avg` | `avg_over_time(biya_tps_24h_avg[7d])` | 7天TPS平均值 |
| `biya_tps_drop_percentage` | `((7d_avg - current) / 7d_avg) * 100` | TPS下降百分比 |
| `biya_performance_health_score` | 综合计算公式 | 性能健康度评分 |

## 告警处理流程

### 1. 收到告警通知
- 查看告警级别和描述
- 访问 Dashboard 查看详细信息
- 根据 runbook_url 查看处理手册

### 2. 初步诊断
- 检查相关指标趋势
- 查看日志和错误信息
- 确认问题范围和影响

### 3. 执行处理
- 按照"处理建议"执行操作
- 记录处理过程
- 监控指标变化

### 4. 问题解决
- 确认告警已解除
- 总结经验教训
- 更新 runbook 文档

## 最佳实践

### 1. 告警规则设计
- ✅ 设置合理的阈值和持续时间，避免误报
- ✅ 使用分层告警（warning -> critical -> emergency）
- ✅ 添加详细的注释和处理建议
- ✅ 定期审查和优化告警规则

### 2. 通知管理
- ✅ 根据告警级别配置不同的通知方式
- ✅ 避免告警疲劳，控制通知频率
- ✅ 使用告警分组和抑制减少重复通知
- ✅ 设置合理的静默时间窗口

### 3. 响应流程
- ✅ 建立明确的告警响应流程
- ✅ 编写详细的 runbook 文档
- ✅ 定期进行告警演练
- ✅ 持续改进响应效率

## 故障排查

### Prometheus 无法加载规则

```bash
# 检查规则文件语法
promtool check rules alert_rules.yml

# 查看 Prometheus 日志
docker logs biya-prometheus

# 重新加载配置
curl -X POST http://localhost:9090/-/reload
```

### Alertmanager 未收到告警

```bash
# 检查 Prometheus 到 Alertmanager 的连接
curl http://localhost:9090/api/v1/alertmanagers

# 查看 Alertmanager 日志
docker logs biya-alertmanager

# 测试告警发送
curl -X POST http://localhost:9090/api/v1/alerts
```

### 告警通知未发送

```bash
# 检查 Alertmanager 配置
amtool config show

# 查看告警状态
amtool alert query

# 检查静默规则
amtool silence query

# 测试接收者配置
amtool config routes test --config.file=alertmanager.yml
```

## 相关文档

- [Prometheus 官方文档](https://prometheus.io/docs/)
- [Alertmanager 官方文档](https://prometheus.io/docs/alerting/latest/alertmanager/)
- [PromQL 查询语言](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [告警规则最佳实践](https://prometheus.io/docs/practices/alerting/)

## 联系方式

如有问题，请联系：
- 运维团队：ops@biya.chain
- 技术支持：support@biya.chain
- 紧急联系：oncall@biya.chain

