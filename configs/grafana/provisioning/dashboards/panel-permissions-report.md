# Grafana 面板权限检查报告

## 仪表板级别设置
- **editable**: `true` ✅ (第18行)
- **权限**: Editor/Admin 可以编辑

## Provisioning 配置
- **allowUiUpdates**: `true` ✅
- **updateIntervalSeconds**: `10` ⚠️ (可能每10秒覆盖用户编辑)

## 面板权限详情

| 面板ID | 面板标题 | 类型 | editable字段 | pluginVersion | 实际权限 | 状态 |
|--------|---------|------|--------------|--------------|---------|------|
| 1 | 当前区块高度 | gauge | ❌ 无（继承仪表板） | ✅ 10.0.0 | Editor/Admin | ✅ 可编辑 |
| 2 | TPS (每秒交易数) | timeseries | ❌ 无（继承仪表板） | ✅ 10.0.0 | Editor/Admin | ✅ 可编辑 |
| 3 | 交易成功率 | gauge | ❌ 无（继承仪表板） | ✅ 10.0.0 | Editor/Admin | ✅ 可编辑 |
| 4 | 内存池状态 | timeseries | ❌ 无（继承仪表板） | ✅ 10.0.0 | Editor/Admin | ✅ 可编辑 |
| 5 | 验证者总数 | gauge | ❌ 无（继承仪表板） | ✅ 10.0.0 | Editor/Admin | ✅ 可编辑 |
| 6 | 平均区块时间 | gauge | ❌ 无（继承仪表板） | ✅ 10.0.0 | Editor/Admin | ✅ 可编辑 |
| 7 | 24小时交易统计 | timeseries | ❌ 无（继承仪表板） | ✅ 10.0.0 | Editor/Admin | ✅ 可编辑 |
| 8 | Gas 价格 | timeseries | ❌ 无（继承仪表板） | ✅ 10.0.0 | Editor/Admin | ✅ 可编辑 |

## 结论

**所有面板的权限状态：Editor/Admin（可编辑）**

### 注意事项：
1. ✅ 所有面板都没有单独的 `editable: false` 设置，因此都继承仪表板级别的 `editable: true`
2. ⚠️ 由于 `updateIntervalSeconds: 10`，provisioning 每10秒会重新加载配置，可能会覆盖用户在UI中的编辑
3. ✅ Provisioning 配置中 `allowUiUpdates: true`，允许通过UI进行更新

### 修复记录：
✅ **已修复**：为所有timeseries面板（面板2、4、7、8）添加了 `pluginVersion: "10.0.0"` 字段
- 问题：大的timeseries面板缺少 `pluginVersion` 字段，导致无法编辑
- 解决：所有8个面板现在都包含 `pluginVersion` 字段

### 如果某个面板无法编辑，可能的原因：
1. 用户权限不足（需要 Editor 或 Admin 角色）
2. Provisioning 更新间隔太短，编辑被覆盖
3. 面板配置不完整（如缺少 `pluginVersion` 字段）- ✅ 已全部修复

### 建议：
- 如果需要保护某些面板不被编辑，可以为特定面板添加 `"editable": false`
- 如果编辑经常被覆盖，可以增加 `updateIntervalSeconds` 的值（如改为60秒或更长）
