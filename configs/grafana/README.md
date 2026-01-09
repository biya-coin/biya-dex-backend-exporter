# Grafana 配置说明

## 概述

Grafana 已配置为自动连接到 Prometheus 数据源，用于可视化监控指标。

## 访问信息

- **访问地址**: http://localhost:3000
- **默认用户名**: `admin`
- **默认密码**: `admin`

⚠️ **安全提示**: 生产环境请修改默认密码！

## 配置说明

### 数据源配置

Prometheus 数据源已通过 provisioning 自动配置：
- 配置文件位置: `configs/grafana/provisioning/datasources/prometheus.yml`
- 数据源名称: Prometheus
- Prometheus 地址: `http://prometheus:9090` (容器内网络)
- 默认数据源: 是

### 数据持久化

Grafana 的数据（包括仪表板、用户配置等）存储在 Docker volume `grafana-data` 中，重启容器不会丢失。

## 使用步骤

1. **启动服务**
   ```bash
   docker compose up -d grafana
   ```

2. **访问 Grafana**
   打开浏览器访问 http://localhost:3000

3. **登录**
   使用默认账号 `admin/admin` 登录（首次登录会要求修改密码）

4. **查看数据源**
   - 进入 Configuration → Data Sources
   - 应该能看到已自动配置的 "Prometheus" 数据源

5. **创建仪表板**
   - 进入 Dashboards → New Dashboard
   - 添加 Panel，选择 Prometheus 数据源
   - 输入 PromQL 查询语句，例如：
     - `up` - 检查服务状态
     - `rate(biya_chain_block_height[5m])` - 查看区块高度增长率
     - 更多指标请参考项目根目录的 `METRICS.md`

## 常用 PromQL 查询示例

### 服务健康检查
```
up{job="biya-exporter"}
```

### 区块高度
```
biya_chain_block_height
```

### 区块高度增长率
```
rate(biya_chain_block_height[5m])
```

### 内存池大小
```
biya_chain_mempool_size
```

## 修改配置

### 修改默认密码

可以通过环境变量或修改 `compose.yaml` 中的 `GF_SECURITY_ADMIN_PASSWORD` 来设置初始密码。

### 添加更多数据源

在 `configs/grafana/provisioning/datasources/` 目录下添加新的 YAML 配置文件即可。

### 导入仪表板

可以通过以下方式导入仪表板：
1. 在 Grafana UI 中：Dashboards → Import
2. 通过 provisioning：在 `configs/grafana/provisioning/dashboards/` 目录下添加配置文件

## 故障排查

### 无法连接 Prometheus

1. 检查 Prometheus 服务是否运行：
   ```bash
   docker compose ps prometheus
   ```

2. 检查网络连接：
   ```bash
   docker compose exec grafana wget -qO- http://prometheus:9090/api/v1/status/config
   ```

3. 查看 Grafana 日志：
   ```bash
   docker compose logs grafana
   ```

### 重置 Grafana

如果需要重置 Grafana 配置：
```bash
docker compose down grafana
docker volume rm biya-backend-exporter_grafana-data
docker compose up -d grafana
```

## 相关文档

- [Prometheus 配置文档](../prometheus/README.md)
- [指标说明](../../METRICS.md)
