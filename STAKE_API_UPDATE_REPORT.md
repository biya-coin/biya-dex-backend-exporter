# Stake API 更新报告

基于最新的 Postman Collection (`Biya-Stake-API.postman_collection.json`)，以下是需要更新的内容。

## 一、参数命名检查

### 检查结果
经过检查 Postman Collection 中的 `raw` URL 格式，发现：
- **实际 URL 参数使用 camelCase**（如 `operatorAddress`, `delegatorAddress`, `validatorAddress`, `proposalId`）
- Postman Collection 的 `query` 数组中的 `key` 字段使用 snake_case，但这只是 Postman 的显示方式
- **代码中使用的 camelCase 格式是正确的** ✅

### 参数格式对比

| API 端点 | Postman raw URL (实际格式) | 代码中 | 状态 |
|---------|---------------------------|--------|------|
| `/stake/validator` | `operatorAddress` | `operatorAddress` | ✅ 正确 |
| `/stake/governance/proposals/by-id` | `proposalId` | `proposalId` | ✅ 正确 |
| `/stake/delegation` | `delegatorAddress`, `validatorAddress` | `delegatorAddress`, `validatorAddress` | ✅ 正确 |
| `/stake/delegation/reward` | `delegatorAddress`, `validatorAddress` | `delegatorAddress`, `validatorAddress` | ✅ 正确 |
| `/stake/delegation/rewards` | `delegatorAddress` | `delegatorAddress` | ✅ 正确 |
| `/stake/delegator/validators` | `delegatorAddress` | `delegatorAddress` | ✅ 正确 |
| `/stake/delegator/delegations` | `delegatorAddress` | `delegatorAddress` | ✅ 正确 |
| `/stake/delegator/withdraw/address` | `delegatorAddress` | `delegatorAddress` | ✅ 正确 |

**结论**：代码中的参数命名与 Postman Collection 中的实际 URL 格式一致，无需修改。

---

## 二、缺失的 API 端点

### 1. 委托管理相关

#### `/stake/validator/delegators` - 获取验证人的委托人列表
- **方法**: GET
- **参数**: 
  - `validator_address` (必需)
  - `pagination.page` (可选，默认1)
  - `pagination.page_size` (可选，默认20，最大100)
  - `pagination.cursor` (可选，基于游标的分页)
- **状态**: ❌ 未实现
- **优先级**: 中

### 2. 治理管理相关

#### `/stake/governance/statistics` - 获取治理统计信息
- **方法**: GET
- **参数**: 无
- **返回**: 治理统计数据，包括平均参与率和总投票权重
- **状态**: ❌ 未实现
- **优先级**: 高（用于填充 `biya_participation_rate_avg` 和 `biya_voting_power_total` 指标）

### 3. 统计信息相关

#### `/stake/statistics` - 获取质押统计信息
- **方法**: GET
- **参数**: 无
- **返回**: 质押统计数据，包括总质押量、24小时奖励和年化收益率
- **状态**: ❌ 未实现
- **优先级**: 高（用于填充以下指标）：
  - `biya_staked_total_byb` - 总质押金额 (BYB)
  - `biya_rewards_24h_total_byb` - 24h总奖励 (BYB)
  - `biya_apr_annual` - 奖励率 (年化)

#### `/stake/slashing/events` - 获取惩罚事件记录
- **方法**: GET
- **参数**: 
  - `start_time` (可选，RFC3339格式，默认24小时前)
  - `end_time` (可选，RFC3339格式，默认现在)
  - `pagination.page` (可选，默认1)
  - `pagination.page_size` (可选，默认20，最大100)
  - `pagination.cursor` (可选，基于游标的分页)
- **返回**: 指定时间范围内的验证人惩罚事件
- **状态**: ❌ 未实现
- **优先级**: 高（用于填充以下指标）：
  - `biya_slashing_events_24h` - 24h惩罚事件
  - `biya_slashing_events_total` - 总惩罚事件（按类型）

### 4. 回购管理相关（Buyback）

以下 API 端点在 Postman Collection 中存在，但代码中完全未实现。需要评估是否需要在 exporter 中实现：

#### 回购轮次管理
- `GET /stake/buyback/rounds` - 获取所有回购轮次
- `GET /stake/buyback/rounds/by-id` - 根据ID获取回购轮次
- `POST /stake/buyback/rounds` - 创建回购轮次
- `PUT /stake/buyback/rounds` - 更新回购轮次
- `PATCH /stake/buyback/rounds/status` - 更新轮次状态

#### 参与记录管理
- `GET /stake/buyback/participations` - 获取所有参与记录
- `GET /stake/buyback/participations/by-id` - 根据ID获取参与记录
- `POST /stake/buyback/participations` - 预订参与名额
- `POST /stake/buyback/participations/submit` - 提交参与代币

#### 收益分配管理
- `GET /stake/buyback/revenue/records` - 获取收益记录
- `POST /stake/buyback/revenue/calculate` - 计算收益
- `POST /stake/buyback/revenue/distribute` - 分配收益
- `POST /stake/buyback/revenue/claim` - 领取收益

#### 销毁记录管理
- `GET /stake/buyback/burn/records` - 获取销毁记录
- `GET /stake/buyback/burn/statistics` - 获取销毁统计
- `POST /stake/buyback/burn/execute` - 执行销毁

#### 统计数据
- `GET /stake/buyback/statistics/participation` - 获取参与统计
- `GET /stake/buyback/statistics/revenue` - 获取收益统计

#### 报告生成
- `POST /stake/buyback/reports/generate` - 生成回购报告

**状态**: ❌ 全部未实现
**优先级**: 低（需要确认 exporter 是否需要收集 buyback 相关指标）

---

## 三、需要实现的 API 方法（按优先级）

### 高优先级（影响现有指标收集）

1. **GetStatistics** - 获取质押统计信息
   ```go
   func (c *Client) GetStatistics(ctx context.Context) (json.RawMessage, error)
   ```

2. **GetSlashingEvents** - 获取惩罚事件记录
   ```go
   func (c *Client) GetSlashingEvents(ctx context.Context, startTime, endTime string, p NestedPagination) (json.RawMessage, error)
   ```

3. **GetGovernanceStatistics** - 获取治理统计信息
   ```go
   func (c *Client) GetGovernanceStatistics(ctx context.Context) (json.RawMessage, error)
   ```

### 中优先级（功能完整性）

4. **GetValidatorDelegators** - 获取验证人的委托人列表
   ```go
   func (c *Client) GetValidatorDelegators(ctx context.Context, validatorAddress string, p NestedPagination) (json.RawMessage, error)
   ```

### 低优先级（需要确认需求）

5. Buyback 相关 API（如果 exporter 需要收集 buyback 指标）

---

## 四、代码更新建议

### 1. 参数命名
✅ **无需修改** - 代码中使用的 camelCase 格式与 Postman Collection 中的实际 URL 格式一致。

### 2. 添加新方法

在 `internal/adapters/stake/client.go` 中添加上述缺失的方法。

### 3. 更新 Collector

在 `internal/collectors/realtime_stake.go` 中：
- 使用 `GetStatistics` 填充质押统计指标
- 使用 `GetSlashingEvents` 填充惩罚事件指标
- 使用 `GetGovernanceStatistics` 填充治理统计指标

---

## 五、实现状态

### ✅ 已完成

1. **参数命名检查** - 确认代码中的 camelCase 格式与 Postman Collection 一致
2. **添加缺失的 API 方法**：
   - ✅ `GetValidatorDelegators` - 获取验证人的委托人列表
   - ✅ `GetGovernanceStatistics` - 获取治理统计信息
   - ✅ `GetStatistics` - 获取质押统计信息
   - ✅ `GetSlashingEvents` - 获取惩罚事件记录
3. **更新 Collector**：
   - ✅ 添加 `readStatistics` 方法填充质押统计指标
   - ✅ 添加 `readSlashingEvents` 方法填充惩罚事件指标
   - ✅ 添加 `readGovernanceStatistics` 方法填充治理统计指标

### 📝 实现细节

- 所有新 API 方法已添加到 `internal/adapters/stake/client.go`
- Collector 已更新以调用新 API 并填充以下指标：
  - `biya_staked_total_byb` - 总质押金额 (BYB)
  - `biya_staked_ratio` - 质押比例
  - `biya_rewards_24h_total_byb` - 24h总奖励 (BYB)
  - `biya_apr_annual` - 年化收益率 (0-100)
  - `biya_slashing_events_24h` - 24h惩罚事件数量
  - `biya_slashing_events_total` - 总惩罚事件（按类型）
  - `biya_voting_power_total` - 总投票权重
  - `biya_participation_rate_avg` - 平均参与率

- 使用灵活的 JSON 解析，兼容多种可能的字段命名方式
- 所有方法都包含错误处理和 source_up 指标设置

## 六、测试建议

1. **参数格式测试**：测试 API 是否同时支持 camelCase 和 snake_case（已确认代码格式正确）
2. **新端点测试**：测试所有新添加的 API 端点
3. **指标验证**：验证新收集的指标是否正确填充到 Prometheus
4. **字段映射验证**：由于使用了灵活的字段名匹配，需要验证实际 API 返回的字段名是否被正确识别

---

## 六、参考文档

- Postman Collection: `Biya-Stake-API.postman_collection.json`
- 现有实现: `internal/adapters/stake/client.go`
- 指标定义: `internal/metrics/metrics.go`
- 数据流文档: `data_flow.md`
