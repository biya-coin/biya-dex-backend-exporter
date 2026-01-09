# Grafana 告警趋势图表配置说明

## 概述

告警趋势图表已集成到 Grafana Dashboard 中，可以直接在 Grafana 中查看最近7天的告警趋势。

## 实现方式

### 方案：直接查询 Prometheus ALERTS 指标

**为什么选择这个方案？**
- ✅ Grafana 已经配置了 Prometheus 数据源
- ✅ 无需额外的 API 调用，性能更好
- ✅ 数据实时更新，延迟低
- ✅ 支持 Grafana 的所有时间范围选择功能
- ✅ 可以利用 Grafana 的告警和注释功能

**数据流：**
```
Prometheus (ALERTS 指标) → Grafana (Prometheus 数据源) → 告警趋势图表
```

## Dashboard 配置

### 面板位置

告警趋势图表已添加到 `biya-chain-overview` Dashboard 中：
- **面板 ID**: 9
- **位置**: 底部（y=20）
- **大小**: 全宽（24 列），高度 8 行

### PromQL 查询

图表使用三个 PromQL 查询：

1. **严重告警数**（红色线条）
   ```promql
   sum(ALERTS{alertstate="firing",severity="critical"})
   ```

2. **警告告警数**（橙色线条）
   ```promql
   sum(ALERTS{alertstate="firing",severity="warning"})
   ```

3. **总告警数**（蓝色线条）
   ```promql
   count(ALERTS{alertstate="firing"})
   ```

### 图表配置

- **图表类型**: Time Series（时间序列折线图）
- **颜色方案**:
  - 严重告警：红色（red）
  - 警告告警：橙色（orange）
  - 总告警数：蓝色（blue）
- **图例**: 显示在底部，包含最后值、最大值、平均值
- **工具提示**: 多系列模式，显示所有系列的数据

## 使用方法

### 1. 查看 Dashboard

1. 访问 Grafana: http://localhost:3000
2. 登录（默认用户名/密码: admin/admin）
3. 导航到 **Dashboards** → **Biya Chain 概览**
4. 滚动到底部查看"告警趋势（最近7天）"图表

### 2. 调整时间范围

- 使用 Grafana 顶部的时间选择器
- 默认显示最近7天
- 可以调整为：
  - 最近1小时
  - 最近6小时
  - 最近24小时
  - 最近7天（默认）
  - 最近30天
  - 自定义时间范围

### 3. 刷新数据

- Dashboard 默认每 10 秒自动刷新
- 也可以手动点击刷新按钮

## 数据说明

### ALERTS 指标

Prometheus 会自动为每个告警规则生成 `ALERTS` 指标：

- **指标格式**: `ALERTS{alertname="告警名称",alertstate="firing|pending",severity="critical|warning",...}`
- **值**: 1 表示告警触发，0 表示告警未触发
- **状态**:
  - `firing`: 告警正在触发
  - `pending`: 告警待触发（满足条件但未达到持续时间）

### 告警级别

根据告警规则中的 `severity` 标签分类：
- `critical`: 严重告警（红色）
- `warning`: 警告告警（橙色）
- `info`: 信息告警（如果有）
- `emergency`: 紧急告警（如果有）

## 故障排查

### 问题：图表显示 "No data"

**可能原因：**
1. Prometheus 中没有告警规则配置
2. 告警规则未触发（没有满足条件的告警）
3. Prometheus 数据源配置错误

**解决方法：**
1. 检查 Prometheus 配置：http://localhost:9090/config
2. 检查告警规则：http://localhost:9090/alerts
3. 在 Prometheus UI 中手动执行查询：
   ```promql
   ALERTS{alertstate="firing"}
   ```

### 问题：图表只显示部分数据

**可能原因：**
1. Prometheus 数据保留期不足
2. 告警规则中的 `severity` 标签未正确设置

**解决方法：**
1. 检查 Prometheus 存储保留期设置
2. 检查告警规则文件中的 `labels.severity` 配置

### 问题：颜色显示不正确

**解决方法：**
1. 编辑 Dashboard
2. 选择告警趋势面板
3. 在 "Field" 标签页中检查 "Overrides" 配置
4. 确保颜色映射正确

## 自定义配置

### 修改查询

如果需要修改查询逻辑，编辑 Dashboard JSON 文件：
`configs/grafana/provisioning/dashboards/biya-chain-overview.json`

找到面板 ID 9，修改 `targets` 中的 `expr` 字段。

### 添加更多告警类型

如果需要按告警名称或其他标签分组，可以添加更多查询：

```json
{
  "expr": "sum(ALERTS{alertstate=\"firing\",alertname=\"验证者节点离线\"})",
  "legendFormat": "节点离线告警",
  "refId": "D"
}
```

### 修改时间范围默认值

编辑 Dashboard JSON，修改 `time.from` 字段：

```json
"time": {
  "from": "now-7d",  // 改为 now-1h, now-24h, now-30d 等
  "to": "now"
}
```

## 与 API 方案对比

### Grafana 直接查询（当前方案）

**优点：**
- ✅ 无需额外 API 调用
- ✅ 性能更好，延迟更低
- ✅ 支持 Grafana 的所有功能（告警、注释、变量等）
- ✅ 数据实时更新

**缺点：**
- ❌ 需要访问 Grafana UI
- ❌ 不能直接在前端应用中嵌入（需要使用 iframe）

### 使用 Exporter API

**优点：**
- ✅ 可以在前端应用中直接调用
- ✅ 统一的 API 接口
- ✅ 可以添加缓存和聚合逻辑

**缺点：**
- ❌ 需要额外的 HTTP 请求
- ❌ 增加系统复杂度

**建议：**
- **Grafana Dashboard**: 用于运维监控和告警分析
- **Exporter API**: 用于前端应用集成和自定义展示

## 相关文档

- [告警趋势 API 文档](./ALERT_TREND_API.md) - Exporter API 使用说明
- [告警规则配置](../prometheus/alert_rules.yml) - Prometheus 告警规则
- [Grafana 配置](./README.md) - Grafana 整体配置说明
