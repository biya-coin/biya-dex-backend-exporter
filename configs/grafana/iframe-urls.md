# Grafana 面板 iframe URL 快速参考

## 基础 URL 模板（推荐使用 d-solo 格式）

```
http://45.249.245.183:3000/d-solo/biya-chain-overview/biya-chain-e6a682-e8a788?from=now-1h&to=now&refresh=10s&orgId=1&panelId={PANEL_ID}&__feature.dashboardSceneSolo
```

**注意：** 使用 `d-solo` 路径和 `panelId` 参数是 Grafana 推荐的单个面板嵌入方式。

## 所有面板的 iframe URL

### 面板 1 - 当前区块高度
```
http://45.249.245.183:3000/d-solo/biya-chain-overview/biya-chain-e6a682-e8a788?from=now-1h&to=now&refresh=10s&orgId=1&panelId=1&__feature.dashboardSceneSolo
```

### 面板 2 - TPS (每秒交易数)
```
http://45.249.245.183:3000/d-solo/biya-chain-overview/biya-chain-e6a682-e8a788?from=now-1h&to=now&refresh=10s&orgId=1&panelId=2&__feature.dashboardSceneSolo
```

### 面板 3 - 交易成功率
```
http://45.249.245.183:3000/d-solo/biya-chain-overview/biya-chain-e6a682-e8a788?from=now-1h&to=now&refresh=10s&orgId=1&panelId=3&__feature.dashboardSceneSolo
```

### 面板 4 - 内存池状态
```
http://45.249.245.183:3000/d-solo/biya-chain-overview/biya-chain-e6a682-e8a788?from=now-1h&to=now&refresh=10s&orgId=1&panelId=4&__feature.dashboardSceneSolo
```

### 面板 5 - 验证者总数
```
http://45.249.245.183:3000/d-solo/biya-chain-overview/biya-chain-e6a682-e8a788?from=now-1h&to=now&refresh=10s&orgId=1&panelId=5&__feature.dashboardSceneSolo
```

### 面板 6 - 平均区块时间
```
http://45.249.245.183:3000/d-solo/biya-chain-overview/biya-chain-e6a682-e8a788?from=now-1h&to=now&refresh=10s&orgId=1&panelId=6&__feature.dashboardSceneSolo
```

### 面板 7 - 24小时交易统计
```
http://45.249.245.183:3000/d-solo/biya-chain-overview/biya-chain-e6a682-e8a788?from=now-1h&to=now&refresh=10s&orgId=1&panelId=7&__feature.dashboardSceneSolo
```

### 面板 11 - Gas 价格（实际面板 ID）
```
http://45.249.245.183:3000/d-solo/biya-chain-overview/biya-chain-e6a682-e8a788?from=now-1h&to=now&refresh=10s&orgId=1&panelId=11&__feature.dashboardSceneSolo
```

**重要提示：** Grafana 导入仪表板时可能会重新分配面板 ID，所以配置文件中的 ID（1-8）可能与实际 ID 不同。请通过 Grafana 的 "Share" → "Embed" 功能获取正确的面板 ID。

## 完整仪表板 URL（所有面板）

```
http://45.249.245.183:3000/d/biya-chain-overview/biya-chain-e6a682-e8a788?orgId=1&from=now-1h&to=now&refresh=10s&kiosk=tv
```

## HTML iframe 代码示例

### 单个面板（Gas 价格 - 使用 d-solo 格式，推荐）

```html
<iframe 
  src="http://45.249.245.183:3000/d-solo/biya-chain-overview/biya-chain-e6a682-e8a788?from=now-1h&to=now&refresh=10s&orgId=1&panelId=11&__feature.dashboardSceneSolo" 
  width="100%" 
  height="600" 
  frameborder="0"
  allowfullscreen>
</iframe>
```

### 其他面板示例（当前区块高度）

```html
<iframe 
  src="http://45.249.245.183:3000/d-solo/biya-chain-overview/biya-chain-e6a682-e8a788?from=now-1h&to=now&refresh=10s&orgId=1&panelId=1&__feature.dashboardSceneSolo" 
  width="100%" 
  height="600" 
  frameborder="0"
  allowfullscreen>
</iframe>
```

## 参数说明

- `d-solo` - Grafana 的 solo 模式路径，专门用于嵌入单个面板（推荐）
- `orgId=1` - 组织 ID
- `from=now-1h&to=now` - 时间范围（可调整）
- `refresh=10s` - 自动刷新间隔
- `panelId=11` - 面板 ID（使用 `panelId` 而不是 `viewPanel`）
- `__feature.dashboardSceneSolo` - 启用新的 dashboard scene solo 功能

## 时间范围选项

- 过去1小时：`from=now-1h&to=now`
- 过去6小时：`from=now-6h&to=now`
- 过去24小时：`from=now-24h&to=now`
- 过去7天：`from=now-7d&to=now`

## 如何获取正确的面板 ID

**重要：** Grafana 导入仪表板时可能会重新分配面板 ID，所以配置文件中的 ID（1-8）可能与实际 ID 不同。

### 方法 1：通过 Grafana UI（推荐）

1. 在 Grafana 中打开仪表板
2. 点击要嵌入的面板右上角的菜单（三个点）
3. 选择 "Share" 或 "分享"
4. 在分享对话框中，选择 "Embed" 或 "嵌入"
5. 复制 iframe URL，其中包含正确的 `panelId` 参数
6. **注意：** 将 URL 中的 `localhost:3000` 替换为 `45.249.245.183:3000`
7. **注意：** 将绝对时间戳（如 `from=1767938828442`）替换为相对时间（如 `from=now-1h`）

### 方法 2：查看 URL

在 Grafana 中打开单个面板时，URL 中会包含 `viewPanel=11` 或 `panelId=11`，使用这个 ID。

## 使用步骤

1. **重启 Grafana 服务**（应用 iframe 嵌入配置）：
   ```bash
   docker-compose restart grafana
   ```

2. **获取正确的面板 ID**：使用上面的方法获取每个面板的实际 ID

3. **测试 URL**：在浏览器中直接访问 iframe URL，确认可以正常显示

4. **嵌入到网页**：将 iframe 代码复制到你的 HTML 页面中

5. **修正 URL**：
   - 将 `localhost:3000` 替换为 `45.249.245.183:3000`
   - 将绝对时间戳替换为相对时间（如 `from=now-1h&to=now`）
   - 确保使用 `d-solo` 路径和 `panelId` 参数
