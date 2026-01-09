# 告警趋势功能实现说明

## 概述

已实现获取最近7天告警趋势折线图的功能，支持：
- **严重告警数**（severity=critical）
- **警告告警数**（severity=warning）
- **总告警数**

## 实现方案

### 方案选择

**主要方案：使用 Prometheus ALERTS 指标**
- Prometheus 会自动为每个告警规则生成 `ALERTS` 指标
- 通过 Prometheus Query Range API 查询历史数据
- 支持查询任意时间范围的历史告警数据

**备选方案：Alertmanager API**
- 当 Prometheus 查询失败时，回退到 Alertmanager API
- 注意：Alertmanager 不存储历史数据，只能返回当前状态

### 为什么选择 Prometheus 而不是直接使用 Alertmanager？

1. **历史数据存储**：Prometheus 会持久化存储所有指标数据，包括告警状态变化
2. **时间序列查询**：Prometheus 提供强大的时间序列查询能力（Query Range API）
3. **数据完整性**：可以查询任意时间点的告警状态，生成完整的历史趋势

Alertmanager 主要用于告警路由和通知，不存储历史数据，只管理当前活跃的告警。

## 实现细节

### 1. 配置支持

在 `internal/config/config.go` 中添加了 `MonitoringConfig`：

```go
type MonitoringConfig struct {
    PrometheusBaseURL   string // Prometheus 地址
    AlertmanagerBaseURL string // Alertmanager 地址（备选）
}
```

### 2. 客户端实现

- **Prometheus 客户端** (`internal/adapters/prometheus/client.go`)
  - 实现 `QueryRange` 方法，支持时间范围查询
  
- **Alertmanager 客户端** (`internal/adapters/alertmanager/client.go`)
  - 实现 `GetAlerts` 方法，查询当前活跃告警

### 3. 告警趋势服务

`internal/server/alerts.go` 实现了告警趋势查询逻辑：

- `GetAlertTrend()`: 主要查询方法
  - 使用 PromQL 查询 `ALERTS` 指标
  - 按 `severity` 标签分类统计
  - 合并多个查询结果生成趋势数据点

- PromQL 查询语句：
  ```promql
  sum(ALERTS{alertstate="firing",severity="critical"})  # 严重告警
  sum(ALERTS{alertstate="firing",severity="warning"})    # 警告告警
  count(ALERTS{alertstate="firing"})                     # 总告警数
  ```

### 4. HTTP API 接口

- **路径**: `GET /api/v1/alerts/trend`
- **参数**: `days` (可选，默认7天，范围1-30)
- **响应格式**: JSON，包含时间戳和告警数量

## 使用步骤

### 1. 配置监控系统地址

在配置文件中添加：

```yaml
monitoring:
  prometheus_base_url: "http://localhost:9090"      # 本地环境
  # 或
  prometheus_base_url: "http://prometheus:9090"      # Docker Compose 环境
  
  alertmanager_base_url: "http://localhost:9093"     # 本地环境
  # 或
  alertmanager_base_url: "http://alertmanager:9093" # Docker Compose 环境
```

### 2. 启动服务

服务启动时会自动初始化监控客户端（如果配置了监控地址）。

### 3. 调用 API

```bash
# 获取最近7天的告警趋势（默认）
curl http://localhost:18080/api/v1/alerts/trend

# 获取最近14天的告警趋势
curl http://localhost:18080/api/v1/alerts/trend?days=14
```

### 4. 前端集成

参考 `ALERT_TREND_API.md` 中的前端集成示例，使用 ECharts 或其他图表库绘制折线图。

## 数据流

```
前端请求
  ↓
GET /api/v1/alerts/trend?days=7
  ↓
AlertTrendService.GetAlertTrend()
  ↓
Prometheus QueryRange API
  ↓
查询 ALERTS 指标（按 severity 分类）
  ↓
合并数据点（按时间戳）
  ↓
返回 JSON 响应
  ↓
前端绘制折线图
```

## 注意事项

1. **Prometheus 数据保留期**：确保 Prometheus 的数据保留期至少覆盖查询的时间范围
2. **告警规则配置**：确保告警规则中正确设置了 `severity` 标签
3. **数据精度**：当前使用1小时作为步长，每个数据点代表1小时内的告警数量

## 文件清单

新增/修改的文件：

1. `internal/config/config.go` - 添加 MonitoringConfig
2. `internal/adapters/prometheus/client.go` - Prometheus 客户端（新建）
3. `internal/adapters/alertmanager/client.go` - Alertmanager 客户端（新建）
4. `internal/server/alerts.go` - 告警趋势服务（新建）
5. `internal/server/http.go` - 添加告警趋势 API 路由
6. `cmd/exporter/main.go` - 初始化监控客户端和服务
7. `configs/config.example.yaml` - 添加监控配置示例
8. `configs/config.docker.yaml` - 添加 Docker 环境监控配置
9. `ALERT_TREND_API.md` - API 使用文档（新建）
10. `ALERT_TREND_IMPLEMENTATION.md` - 实现说明文档（本文件）

## 测试建议

1. **单元测试**：测试 Prometheus 和 Alertmanager 客户端的查询逻辑
2. **集成测试**：启动完整的 Docker Compose 环境，测试 API 接口
3. **数据验证**：在 Prometheus UI 中手动执行 PromQL 查询，验证结果一致性

## 后续优化建议

1. **缓存机制**：对查询结果进行缓存，减少 Prometheus 查询压力
2. **数据聚合**：支持按天/按小时等不同粒度聚合
3. **告警详情**：扩展 API 返回每个时间点的告警详情列表
4. **实时更新**：支持 WebSocket 推送实时告警趋势更新
