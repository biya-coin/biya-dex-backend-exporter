# Grafana 面板 iframe 嵌入指南

## 配置说明

已配置 Grafana 支持 iframe 嵌入，相关配置在 `compose.yaml` 中：
- `GF_SECURITY_ALLOW_EMBEDDING=true` - 允许 iframe 嵌入
- `GF_AUTH_ANONYMOUS_ENABLED=true` - 允许匿名访问（Viewer 角色）

## iframe URL 格式

### 1. 完整仪表板嵌入

```html
<iframe 
  src="http://45.249.245.183:3000/d/biya-chain-overview/biya-chain-e6a682-e8a788?orgId=1&from=now-1h&to=now&refresh=10s&kiosk=tv" 
  width="100%" 
  height="800" 
  frameborder="0">
</iframe>
```

### 2. 单个面板嵌入（推荐 - 使用 d-solo 格式）

**重要：** Grafana 推荐使用 `d-solo` 路径来嵌入单个面板，这是专门为嵌入设计的格式。

#### 面板列表和对应的 URL（使用 d-solo 格式）：

| 面板标题 | 配置文件ID | 实际面板ID | iframe URL |
|---------|-----------|-----------|-----------|
| 当前区块高度 | 1 | 需确认 | `http://45.249.245.183:3000/d-solo/biya-chain-overview/biya-chain-e6a682-e8a788?from=now-1h&to=now&refresh=10s&orgId=1&panelId=1&__feature.dashboardSceneSolo` |
| TPS (每秒交易数) | 2 | 需确认 | `http://45.249.245.183:3000/d-solo/biya-chain-overview/biya-chain-e6a682-e8a788?from=now-1h&to=now&refresh=10s&orgId=1&panelId=2&__feature.dashboardSceneSolo` |
| 交易成功率 | 3 | 需确认 | `http://45.249.245.183:3000/d-solo/biya-chain-overview/biya-chain-e6a682-e8a788?from=now-1h&to=now&refresh=10s&orgId=1&panelId=3&__feature.dashboardSceneSolo` |
| 内存池状态 | 4 | 需确认 | `http://45.249.245.183:3000/d-solo/biya-chain-overview/biya-chain-e6a682-e8a788?from=now-1h&to=now&refresh=10s&orgId=1&panelId=4&__feature.dashboardSceneSolo` |
| 验证者总数 | 5 | 需确认 | `http://45.249.245.183:3000/d-solo/biya-chain-overview/biya-chain-e6a682-e8a788?from=now-1h&to=now&refresh=10s&orgId=1&panelId=5&__feature.dashboardSceneSolo` |
| 平均区块时间 | 6 | 需确认 | `http://45.249.245.183:3000/d-solo/biya-chain-overview/biya-chain-e6a682-e8a788?from=now-1h&to=now&refresh=10s&orgId=1&panelId=6&__feature.dashboardSceneSolo` |
| 24小时交易统计 | 7 | 需确认 | `http://45.249.245.183:3000/d-solo/biya-chain-overview/biya-chain-e6a682-e8a788?from=now-1h&to=now&refresh=10s&orgId=1&panelId=7&__feature.dashboardSceneSolo` |
| Gas 价格 | 8 | **11** | `http://45.249.245.183:3000/d-solo/biya-chain-overview/biya-chain-e6a682-e8a788?from=now-1h&to=now&refresh=10s&orgId=1&panelId=11&__feature.dashboardSceneSolo` |

**重要提示：** 
- Grafana 导入仪表板时可能会重新分配面板 ID，所以配置文件中的 ID（1-8）可能与实际 ID 不同
- 请通过 Grafana 的 "Share" → "Embed" 功能获取每个面板的实际 ID
- 使用 `d-solo` 路径和 `panelId` 参数是 Grafana 推荐的单个面板嵌入方式

### 3. URL 参数说明

#### d-solo 格式参数（推荐）：
- `d-solo` - Grafana 的 solo 模式路径，专门用于嵌入单个面板
- `orgId=1` - 组织 ID（默认 1）
- `from=now-1h&to=now` - 时间范围（过去1小时到现在）
- `refresh=10s` - 自动刷新间隔（10秒）
- `panelId=11` - 面板 ID（使用 `panelId` 而不是 `viewPanel`）
- `__feature.dashboardSceneSolo` - 启用新的 dashboard scene solo 功能

#### 传统格式参数（仍可用）：
- `viewPanel=8` - 只显示指定面板（面板 ID）
- `kiosk=tv` - 电视模式，隐藏顶部导航栏和工具栏（推荐用于嵌入）

### 4. 时间范围参数

- **过去1小时**：`from=now-1h&to=now`
- **过去6小时**：`from=now-6h&to=now`
- **过去24小时**：`from=now-24h&to=now`
- **过去7天**：`from=now-7d&to=now`
- **自定义时间**：`from=2026-01-08T06:58:50.400Z&to=2026-01-09T06:58:50.400Z`

## HTML 示例

### 示例 1：单个面板嵌入（Gas 价格 - 使用 d-solo 格式）

```html
<!DOCTYPE html>
<html>
<head>
    <title>Biya Chain - Gas 价格</title>
    <style>
        body {
            margin: 0;
            padding: 0;
            font-family: Arial, sans-serif;
        }
        .panel-container {
            width: 100%;
            height: 600px;
            border: 1px solid #ddd;
        }
    </style>
</head>
<body>
    <h1>Gas 价格监控</h1>
    <div class="panel-container">
        <iframe 
            src="http://45.249.245.183:3000/d-solo/biya-chain-overview/biya-chain-e6a682-e8a788?from=now-1h&to=now&refresh=10s&orgId=1&panelId=11&__feature.dashboardSceneSolo" 
            width="100%" 
            height="100%" 
            frameborder="0"
            allowfullscreen>
        </iframe>
    </div>
</body>
</html>
```

### 示例 2：多个面板并排显示

```html
<!DOCTYPE html>
<html>
<head>
    <title>Biya Chain 监控面板</title>
    <style>
        body {
            margin: 0;
            padding: 20px;
            font-family: Arial, sans-serif;
        }
        .dashboard-grid {
            display: grid;
            grid-template-columns: repeat(2, 1fr);
            gap: 20px;
        }
        .panel-container {
            width: 100%;
            height: 400px;
            border: 1px solid #ddd;
            border-radius: 4px;
            overflow: hidden;
        }
        .panel-title {
            background: #f5f5f5;
            padding: 10px;
            margin: 0;
            font-size: 14px;
            font-weight: bold;
        }
    </style>
</head>
<body>
    <h1>Biya Chain 监控面板</h1>
    <div class="dashboard-grid">
        <div>
            <div class="panel-title">当前区块高度</div>
            <div class="panel-container">
                <iframe 
                    src="http://45.249.245.183:3000/d/biya-chain-overview/biya-chain-e6a682-e8a788?orgId=1&from=now-1h&to=now&refresh=10s&viewPanel=1&kiosk=tv" 
                    width="100%" 
                    height="100%" 
                    frameborder="0">
                </iframe>
            </div>
        </div>
        <div>
            <div class="panel-title">TPS (每秒交易数)</div>
            <div class="panel-container">
                <iframe 
                    src="http://45.249.245.183:3000/d/biya-chain-overview/biya-chain-e6a682-e8a788?orgId=1&from=now-1h&to=now&refresh=10s&viewPanel=2&kiosk=tv" 
                    width="100%" 
                    height="100%" 
                    frameborder="0">
                </iframe>
            </div>
        </div>
        <div>
            <div class="panel-title">交易成功率</div>
            <div class="panel-container">
                <iframe 
                    src="http://45.249.245.183:3000/d/biya-chain-overview/biya-chain-e6a682-e8a788?orgId=1&from=now-1h&to=now&refresh=10s&viewPanel=3&kiosk=tv" 
                    width="100%" 
                    height="100%" 
                    frameborder="0">
                </iframe>
            </div>
        </div>
        <div>
            <div class="panel-title">Gas 价格</div>
            <div class="panel-container">
                <iframe 
                    src="http://45.249.245.183:3000/d/biya-chain-overview/biya-chain-e6a682-e8a788?orgId=1&from=now-1h&to=now&refresh=10s&viewPanel=8&kiosk=tv" 
                    width="100%" 
                    height="100%" 
                    frameborder="0">
                </iframe>
            </div>
        </div>
    </div>
</body>
</html>
```

## 如何获取正确的面板 ID

**重要：** Grafana 导入仪表板时可能会重新分配面板 ID，所以配置文件中的 ID（1-8）可能与实际 ID 不同。

### 方法 1：通过 Grafana UI（推荐）

1. 在 Grafana 中打开仪表板
2. 点击要嵌入的面板右上角的菜单（三个点）
3. 选择 "Share" 或 "分享"
4. 在分享对话框中，选择 "Embed" 或 "嵌入"
5. 复制 iframe URL，其中包含正确的 `panelId` 参数
6. **修正 URL：**
   - 将 `localhost:3000` 替换为 `45.249.245.183:3000`
   - 将绝对时间戳（如 `from=1767938828442`）替换为相对时间（如 `from=now-1h`）
   - 确保使用 `d-solo` 路径和 `panelId` 参数

### 方法 2：直接查看 URL

在 Grafana 中打开单个面板时，URL 中会包含 `viewPanel=11` 或 `panelId=11`，使用这个 ID。

**示例：**
- 完整仪表板：`http://45.249.245.183:3000/d/biya-chain-overview/...`
- 单个面板（传统格式）：`http://45.249.245.183:3000/d/biya-chain-overview/...&viewPanel=11`
- 单个面板（d-solo 格式，推荐）：`http://45.249.245.183:3000/d-solo/biya-chain-overview/...&panelId=11`

## 使用 API Key 认证（生产环境推荐）

如果需要更安全的访问控制，可以使用 API Key：

1. 在 Grafana 中创建 API Key：
   - 登录 Grafana
   - 进入 Configuration → API Keys
   - 创建新的 API Key（Role: Viewer）

2. 在 iframe URL 中添加认证：
```html
<iframe 
  src="http://45.249.245.183:3000/d/biya-chain-overview/biya-chain-e6a682-e8a788?orgId=1&from=now-1h&to=now&refresh=10s&viewPanel=8&kiosk=tv&auth=YOUR_API_KEY" 
  width="100%" 
  height="600" 
  frameborder="0">
</iframe>
```

或者使用 HTTP Basic Auth：
```html
<iframe 
  src="http://admin:admin@45.249.245.183:3000/d/biya-chain-overview/biya-chain-e6a682-e8a788?orgId=1&from=now-1h&to=now&refresh=10s&viewPanel=8&kiosk=tv" 
  width="100%" 
  height="600" 
  frameborder="0">
</iframe>
```

## 注意事项

1. **CORS 问题**：如果从不同域名嵌入，可能需要配置 CORS
2. **安全性**：生产环境建议使用 API Key 而不是匿名访问
3. **性能**：`refresh=10s` 会每10秒刷新，可根据需要调整
4. **面板 ID**：如果面板 ID 不是 1-8，请使用实际的面板 ID（如 panel-11）

## 快速测试

访问以下 URL 测试单个面板（Gas 价格，使用 d-solo 格式）：
```
http://45.249.245.183:3000/d-solo/biya-chain-overview/biya-chain-e6a682-e8a788?from=now-1h&to=now&refresh=10s&orgId=1&panelId=11&__feature.dashboardSceneSolo
```

如果显示正常，就可以在 iframe 中使用了。

## 常见问题

### Q: Grafana 提供的 embed URL 中显示 `localhost:3000`，怎么办？
A: 将 URL 中的 `localhost:3000` 替换为实际服务器地址 `45.249.245.183:3000`

### Q: URL 中使用的是绝对时间戳（如 `from=1767938828442`），怎么办？
A: 将绝对时间戳替换为相对时间，例如：
- `from=1767938828442&to=1767942428442` → `from=now-1h&to=now`

### Q: 应该使用 `d-solo` 还是 `d` 路径？
A: 推荐使用 `d-solo` 路径，这是 Grafana 专门为嵌入单个面板设计的格式，性能更好。

### Q: 应该使用 `panelId` 还是 `viewPanel` 参数？
A: 使用 `d-solo` 路径时，使用 `panelId` 参数；使用传统 `d` 路径时，使用 `viewPanel` 参数。
