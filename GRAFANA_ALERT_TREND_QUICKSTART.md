# Grafana 告警趋势图表 - 快速开始

## 实现逻辑

告警趋势图表在 Grafana 中的实现逻辑非常简单：

```
Prometheus (ALERTS 指标) 
    ↓
Grafana Prometheus 数据源
    ↓
告警趋势折线图（3条线：严重告警、警告告警、总告警数）
```

## 为什么直接在 Grafana 中查询？

1. **数据源统一**：Grafana 已经配置了 Prometheus 数据源
2. **性能更好**：直接查询 Prometheus，无需额外的 API 层
3. **功能完整**：支持 Grafana 的所有功能（时间范围选择、刷新、告警等）
4. **实时更新**：数据实时同步，延迟低

## 快速查看

1. **启动服务**（如果还没启动）：
   ```bash
   docker-compose up -d
   ```

2. **访问 Grafana**：
   - 地址：http://localhost:3000
   - 用户名/密码：admin/admin

3. **查看 Dashboard**：
   - 导航到：**Dashboards** → **Biya Chain 概览**
   - 滚动到底部，找到 **"告警趋势（最近7天）"** 面板

4. **调整时间范围**：
   - 使用顶部的时间选择器
   - 可以选择：最近1小时、6小时、24小时、7天、30天等

## 图表说明

图表显示三条折线：
- 🔴 **红色线**：严重告警数（severity=critical）
- 🟠 **橙色线**：警告告警数（severity=warning）
- 🔵 **蓝色线**：总告警数（所有 firing 状态的告警）

## PromQL 查询

图表使用的 PromQL 查询：

```promql
# 严重告警
sum(ALERTS{alertstate="firing",severity="critical"})

# 警告告警
sum(ALERTS{alertstate="firing",severity="warning"})

# 总告警数
count(ALERTS{alertstate="firing"})
```

## 配置位置

Dashboard 配置文件：
- `configs/grafana/provisioning/dashboards/biya-chain-overview.json`
- 面板 ID: 9
- 面板标题: "告警趋势（最近7天）"

## 与 Exporter API 的关系

- **Grafana Dashboard**：用于运维监控，直接在 Grafana UI 中查看
- **Exporter API** (`/api/v1/alerts/trend`)：用于前端应用集成，可以通过 HTTP API 获取数据

两者使用相同的数据源（Prometheus ALERTS 指标），但应用场景不同。

## 故障排查

如果图表显示 "No data"：

1. **检查 Prometheus 是否有告警规则**：
   ```bash
   curl http://localhost:9090/api/v1/rules
   ```

2. **检查是否有告警触发**：
   ```bash
   curl http://localhost:9090/api/v1/alerts
   ```

3. **在 Prometheus UI 中测试查询**：
   - 访问：http://localhost:9090
   - 在查询框中输入：`ALERTS{alertstate="firing"}`

## 详细文档

更多详细信息请参考：
- [Grafana 告警趋势配置说明](./configs/grafana/ALERT_TREND_GRAFANA.md)
- [告警趋势 API 文档](./ALERT_TREND_API.md)
