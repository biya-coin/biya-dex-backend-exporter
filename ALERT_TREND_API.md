# 告警趋势查询 API

本文档说明如何使用告警趋势查询 API 获取最近7天的告警趋势折线图数据。

## 功能说明

告警趋势查询 API 提供了获取历史告警数据的功能，支持：
- **严重告警数**（severity=critical）
- **警告告警数**（severity=warning）
- **总告警数**

## 数据来源

系统优先使用 **Prometheus** 的 `ALERTS` 指标来查询历史告警数据。如果 Prometheus 查询失败，会回退到 **Alertmanager API**（但 Alertmanager 不存储历史数据，只能返回当前状态）。

### Prometheus ALERTS 指标

Prometheus 会自动为每个告警规则生成 `ALERTS` 指标：
- `ALERTS{alertstate="firing"}` - 正在触发的告警
- `ALERTS{alertstate="pending"}` - 待触发的告警

告警规则中的 `labels` 会被保留在指标中，包括 `severity` 标签。

## 配置

在配置文件中添加监控系统配置：

```yaml
monitoring:
  # Prometheus 地址，用于查询告警历史数据
  prometheus_base_url: "http://localhost:9090"
  # 如果使用 Docker Compose，可以使用服务名：
  # prometheus_base_url: "http://prometheus:9090"
  
  # Alertmanager 地址，用于查询当前活跃告警（备选方案）
  alertmanager_base_url: "http://localhost:9093"
  # 如果使用 Docker Compose，可以使用服务名：
  # alertmanager_base_url: "http://alertmanager:9093"
```

## API 接口

### GET /api/v1/alerts/trend

获取告警趋势数据。

#### 请求参数

| 参数 | 类型 | 必需 | 默认值 | 说明 |
|------|------|------|--------|------|
| days | int | 否 | 7 | 查询最近 N 天的数据（1-30） |

#### 响应格式

```json
{
  "success": true,
  "data": [
    {
      "timestamp": 1704067200,
      "critical": 2,
      "warning": 5,
      "total": 7
    },
    {
      "timestamp": 1704070800,
      "critical": 1,
      "warning": 3,
      "total": 4
    }
  ],
  "message": ""
}
```

#### 响应字段说明

- `success`: 查询是否成功
- `data`: 数据点数组，按时间戳升序排列
  - `timestamp`: Unix 时间戳（秒）
  - `critical`: 严重告警数量
  - `warning`: 警告告警数量
  - `total`: 总告警数量
- `message`: 可选的提示信息（例如：当使用 Alertmanager 备选方案时的说明）

#### 示例请求

```bash
# 获取最近7天的告警趋势（默认）
curl http://localhost:18080/api/v1/alerts/trend

# 获取最近14天的告警趋势
curl http://localhost:18080/api/v1/alerts/trend?days=14
```

## 前端集成示例

### JavaScript/TypeScript

```typescript
interface AlertTrendPoint {
  timestamp: number;
  critical: number;
  warning: number;
  total: number;
}

interface AlertTrendResponse {
  success: boolean;
  data: AlertTrendPoint[];
  message?: string;
}

async function fetchAlertTrend(days: number = 7): Promise<AlertTrendPoint[]> {
  const response = await fetch(`http://localhost:18080/api/v1/alerts/trend?days=${days}`);
  const result: AlertTrendResponse = await response.json();
  
  if (!result.success) {
    throw new Error(result.message || 'Failed to fetch alert trend');
  }
  
  return result.data;
}

// 使用示例
fetchAlertTrend(7).then(data => {
  // 使用数据绘制折线图
  // data 格式: [{timestamp: 1704067200, critical: 2, warning: 5, total: 7}, ...]
  console.log('Alert trend data:', data);
});
```

### 使用 ECharts 绘制折线图

```javascript
async function renderAlertTrendChart() {
  const data = await fetchAlertTrend(7);
  
  // 转换时间戳为日期字符串
  const times = data.map(point => 
    new Date(point.timestamp * 1000).toLocaleString('zh-CN')
  );
  
  const criticalData = data.map(point => point.critical);
  const warningData = data.map(point => point.warning);
  const totalData = data.map(point => point.total);
  
  const option = {
    title: {
      text: '最近7天告警趋势'
    },
    tooltip: {
      trigger: 'axis'
    },
    legend: {
      data: ['严重告警', '警告告警', '总告警数']
    },
    xAxis: {
      type: 'category',
      data: times
    },
    yAxis: {
      type: 'value'
    },
    series: [
      {
        name: '严重告警',
        type: 'line',
        data: criticalData,
        itemStyle: { color: '#f56c6c' }
      },
      {
        name: '警告告警',
        type: 'line',
        data: warningData,
        itemStyle: { color: '#e6a23c' }
      },
      {
        name: '总告警数',
        type: 'line',
        data: totalData,
        itemStyle: { color: '#409eff' }
      }
    ]
  };
  
  // 使用 ECharts 渲染图表
  const chart = echarts.init(document.getElementById('alert-trend-chart'));
  chart.setOption(option);
}
```

## 注意事项

1. **Prometheus 存储保留期**：确保 Prometheus 的数据保留期至少覆盖查询的时间范围（默认保留15天）。如果查询7天前的数据，但 Prometheus 只保留3天，则无法获取完整数据。

2. **数据精度**：API 使用1小时作为步长（step），这意味着每个数据点代表1小时内的告警数量。如果需要更细粒度的数据，可以修改代码中的 `step` 参数。

3. **Alertmanager 备选方案**：如果 Prometheus 查询失败，系统会回退到 Alertmanager API。但 Alertmanager 不存储历史数据，只能返回当前时间点的告警状态。

4. **告警规则配置**：确保告警规则中正确设置了 `severity` 标签，否则无法按严重程度分类统计。

## 故障排查

### 问题：返回的数据为空

1. 检查 Prometheus 是否正常运行
2. 检查配置中的 `prometheus_base_url` 是否正确
3. 检查 Prometheus 是否有告警规则配置
4. 检查告警规则是否已触发（`ALERTS` 指标是否存在）

### 问题：只能获取当前告警，没有历史数据

1. 检查 Prometheus 的数据保留期设置
2. 检查 Prometheus 的存储是否正常工作
3. 查看 Prometheus 日志确认是否有错误

### 问题：告警数量统计不准确

1. 检查告警规则中的 `severity` 标签是否正确设置
2. 检查 PromQL 查询是否正确（可以在 Prometheus UI 中测试）

## 相关文档

- [Prometheus 告警规则配置](./configs/prometheus/alert_rules.yml)
- [Alertmanager 配置](./configs/prometheus/alertmanager.yml)
- [数据流文档](./data_flow.md)
