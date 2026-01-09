# 数据来源路径分析

> **Last Updated**: 2026-01-08 
> **Note**: 告警管理通过 Alertmanager API v2 实现
> **Note**: （injective-core/chain → Gin Backend → Frontend）表示数据流向，数据来源是 injective-core/chain，经过 Gin Backend，最终到达 Frontend
> **Note**: injective-core/chain 代表链节点，biya-explorer 代表 Explorer API，biya-stake 代表 Stake API
> **Note**: ? 表示数据来源待确认，可以先放一个placeholder，继续开发其余部分，后续确认数据来源后再更新


## 1. 运行概览

### 1.1 系统概览
- 当前区块高度（injective-core/chain → Gin Backend → Frontend）
- 24h交易数（biya-explorer → Prometheus → Gin Backend → Frontend）
- 当前网络TPS （biya-explorer → Prometheus → Gin Backend → Frontend）
- 平均出块时间 （biya-explorer → Prometheus → Gin Backend → Frontend）
- 活跃节点数(活跃验证者节点数)（biya-stake: Get Validators → Gin Backend → Frontend）
- 24h活跃地址数 （biya-explorer → Prometheus → Gin Backend → Frontend）
- 网络拥堵状态 （injective-core/chain → Gin Backend → Frontend）（池大小默认5000，在genesis中定义）

### 1.2 实时告警 ⭐ Alertmanager集成
**数据流**: Prometheus (Alert Rules) → Alertmanager → Gin Backend → Frontend

| 告警类型 | 数据源 | Prometheus alertname |
|----------|--------|---------------------|
| 节点离线 | biya-stake | `ValidatorOffline` |
| 交易成功率低 | biya-explorer | `LowTxSuccessRate` |
| Gas价格飙升 | biya-explorer | `HighGasPrice` |
| 交易池拥堵 | injective-core | `MempoolCongestion` |
| 节点同步落后 | biya-stake + injective-core | `NodeSyncBehind` |
| Gas价格波动 | biya-explorer | `GasPriceVolatile` |

**UI操作映射**:
| UI操作 | Alertmanager API |
|--------|------------------|
| 确认 | `POST /api/v2/silences` (创建静默) |
| 无需处理 | `POST /api/v2/silences` (带"false_alarm"注释) |
| 解决 | 存储到 Redis (Alertmanager无法强制解决) |
| 取消静默 | `DELETE /api/v2/silence/{id}` |


1.3 gas费用分析
1. 当前Gas价格/24h最高Gas/24h最低Gas （biya-explorer → Prometheus → Gin Backend → Frontend）
2. 交易池状态 
  - 待处理交易（injective–core/chain  → Gin Backend  → Frontend， fetch txs from mempool） 
  – 平均等待时间 (目前不做) 

1.4 节点同步状态  [?]
- 同步进度 （injective–core/chain  → Gin Backend  → Frontend, Gin提供包含该节点最大区块高度数据的节点列表给前端)
- 最新区块 （injective–core/chain  → Gin Backend  → Frontend, fetch block from chain directly）

1.5 交易数据分析
- 24h失败交易数 （biya-explorer: Get Failed Transactions 24H → Gin Backend  → Frontend）
- 失败原因分析 （biya-explorer: Get Failed Transactions 24H → Gin Backend  → Frontend， 根据失败交易列表遍历和统计对应的所有交易）
- 失败率最高的合约（biya-explorer: Get Failed Transactions 24H / List Contract Infos → Gin Backend  → Frontend，根据智能合约列表遍历和统计对应的所有交易）
  - 合约 
  - 失败率


2. 网络性能
2.1 实时监控
- 当前网络状态 （Gin Backend 后台管理自己算,根据公链后台管理系统需求文档）
- 网络性能指数 （Gin Backend 后台管理自己算，根据公链后台管理系统需求文档）
- 24h性能告警 （Gin Backend 后台管理自己算，根据公链后台管理系统需求文档）

- 当前TPS （biya-explorer:Get Block Gas Utilization → Gin Backend → Frontend）
- 24h平均TPS （biya-explorer: Get Block Gas Utilization → Gin Backend → Frontend）
-  平均确认时间 （目前不做）
- 待处理交易数量 （injective–core/chain  → Gin Backend  → Frontend， fetch txs from mempool）
-  区块Gas利用率 （获取最近生成的10个区块，biya-explorer: Get Block Gas Utilization → Gin Backend → Frontend）

-  历史性能异常记录 （后台管理自己算，根据公链后台管理系统需求文档）

2.2 指标趋势
- 网络指标趋势（biya-explorer:Get Block Gas Utilization → Prometheus → Grafana → Frontend） 

3. 核心管理
3.1 节点管理
- 总验证者数量 （stake: Get Validators → Gin Backend → Frontend）
- 参与共识验证者 （stake: Get Validators → Gin Backend → Frontend）与活跃验证者是一个意思?
- 活跃验证者  （stake: Get Validators → Gin Backend → Frontend）
- 总质押金额 (BYB) （stake: Get Validators → Gin Backend → Frontend, 从活跃validators统计）
- 24h总奖励 (BYB) （stake: ? → Gin Backend → Frontend）
- 奖励率 (年化) （stake: ? → Gin Backend → Frontend）
- 24h惩罚事件 （stake: ? → Gin Backend → Frontend）

- 验证者节点列表 （stake: Get Validators → Gin Backend → Frontend）
  (只有排名前50的验证者参与共识并获得奖励)
  当前MaxValidators: 50 (?, 参与共识验证者)  
  第50名质押量: 2,850,000 BYB (?) 
  未参与共识验证者: 1,184 (?)
  
  详情
  - 基础信息 
    - 节点名称也是链上validator有提供 （不能编辑）待确认
    - 验证者操作地址 （biya-stake: Get Validator: operatorAddress → Gin Backend → Frontend）
    - 钱包地址 （biya-stake: Get Validator → Gin Backend → Frontend）
    - 节点类型 （Gin Backend 管理后台自己标记, 节点编辑并保存到本地pg数据库）
    - 注册时间（ biya–stake: 验证者首次出块的时间戳 → Gin Backend → Frontend）
  - 性能监控
    - 当前状态 （biya-stake: Get Validator: status → Gin Backend → Frontend）
    - 最后活跃（biya-stake: Get Validator: timestamp → Gin Backend → Frontend）
    - 出块成功率 （biya-stake: Get Validator: uptimePercentage → Gin Backend → Frontend）
    - 总出块数 （biya-stake: Get Validator: proposed → Gin Backend → Frontend）
    - 连续在线时长（biya-stake: Get Validator → Gin Backend → Frontend ）
    – 出块成功率趋势图 (Prometheus → Grafana → Frontend, 不需要Gin Backend)
  - 质押&奖励 （biya-stake）
    – 💰 质押详情
      – 质押金额: (biya-stake: Get Validator: tokens → Gin Backend → Frontend)
      – 质押状态: (biya-stake: Get Validator: ? → Gin Backend → Frontend)
      – 质押时间: (biya-stake: Get Validator: ? → Gin Backend → Frontend)
      – 解锁期: (biya-stake: Get Validator: ? → Gin Backend → Frontend)
      – 委托数量: (biya-stake: Get Validator: ? → Gin Backend → Frontend)
      – 佣金费率:  (biya-stake: Get Validator: ? → Gin Backend → Frontend)
    – 💎 奖励统计 (?)
      – 累计出块奖励: (biya-stake: Get Validator: ? → Gin Backend → Frontend)
      – 累计交易费分成: (biya-stake: Get Validator: ? → Gin Backend → Frontend)
      – 总累计奖励: (biya-stake: Get Validator: ? → Gin Backend → Frontend)
      – 本月奖励: (biya-stake: Get Validator: ? → Gin Backend → Frontend)
      – 待领取奖励: (biya-stake: Get Validator: ? → Gin Backend → Frontend)
      – 实际年化收益率: (biya-stake: Get Validator: ? → Gin Backend → Frontend)
    – ⚠️ 惩罚记录
  - 🌐 网络配置
    – P2P地址:
    – RPC端点:
    – RPC状态: （biya–core/chain → Prometheus → Gin Backend → Frontend）
    – REST API端点:
    – API状态: (biya–core/chain → Prometheus → Gin Backend → Frontend )
    – 当前连接数: (biya–core/chain → Prometheus → Gin Backend → Frontend )
    – 24h请求量: (biya–core/chain → Prometheus → Gin Backend → Frontend )
    – 平均响应时间: (biya–core/chain → Prometheus → Gin Backend → Frontend )
    – 限流状态: (biya–core/chain → Prometheus → Gin Backend → Frontend )
    
  - 👤 运营商信息 (Gin Backend后台管理，编辑和维护在本地pg数据库，然后Gin Backend提供接口给前端)
    – 运营商图标:
    – 运营商名称:
    – 联系邮箱:
    – 网站:
    – 简介:
    – 管理员备注:
    
  – 📋 节点日志
    – 节点日志分析 - validator-001 (biya–core/chain → Prometheus → Gin Backend → Frontend )
    –  💾 下载完整日志 (biya–core/chain → Prometheus → Gin Backend → Frontend ) 
  - 📝 编辑验证者节点 (Gin Backend后台管理，编辑和维护在本地pg数据库，然后Gin Backend提供接口给前端)
    – 基础信息
      – 节点名称
      – 节点类型
      – 节点描述
    – 🌐 网络配置 (手动编辑字段)
      – P2P地址 
      – RPC端点 
      – REST API端点 
    – 👤 运营商信息
      – 运营商图标 
      – 运营商名称 
      – 联系邮箱 
      – 网站 
      – 简介 
      – 管理员备注 
– ⚠️ 惩罚记录列表 (biya-stake: ? → Gin Backend → Frontend)
  – ⚠️ 惩罚详情
    
3.2 BYB数据监控 （暂时没有不做，没有服务提供数据来源）

4. 浏览器管理
4.1 地址标签管理 （列表: biya-explorer: ? → Gin Backend → Frontend）
  – 添加地址标签 ( biya-explorer: Add Address Tag → Gin Backend → Frontend)
  – 删除地址标签 ( biya-explorer: Delete Address Tag → Gin Backend → Frontend)
  – 编辑地址标签 ( biya-explorer: Update Address Tag → Gin Backend → Frontend)
  
4.2 智能合约管理 （列表: biya-explorer: List Contract Infos → Gin Backend → Frontend）
  – 管理标签/合约标签管理
    – 📋 已配置标签 ( biya-explorer: Get Contract Tags → Gin Backend → Frontend)
    – 添加标签 ( biya-explorer: Add Contract Tag → Gin Backend → Frontend)
    – 删除标签 ( biya-explorer: Remove Contract Tag → Gin Backend → Frontend)
    – 编辑标签 ( biya-explorer: Update Contract Tag → Gin Backend → Frontend)
  – 添加合约 ( biya-explorer: Add Contract Info → Gin Backend → Frontend)
  – 删除 ( biya-explorer: Remove Contract Info → Gin Backend → Frontend)
  – 编辑 ( biya-explorer: Update Contract Info → Gin Backend → Frontend)
  
5. 网络治理
5.1 网络治理
  - 总提案数  （biya-stake: Get Proposals → Gin Backend → Frontend）
  - 已通过提案 （biya-stake: Get Proposals: status=passed → Gin Backend → Frontend）
  - 进行中投票 （biya-stake: Get Proposals: status=active → Gin Backend → Frontend）
  - 平均参与率 （biya-stake: Get Proposals ? → Gin Backend → Frontend）待确认 所有提案的平均投票参与率 | 计算：总投票权重/总质押量 
  - 总投票权重 （biya-stake: Get Proposals ? → Gin Backend → Frontend）待确认
  
  - 治理提案管理 (列表: biya-stake: Get Proposals → Gin Backend → Frontend)
    – 创建治理提案（biya-core/chain: Create Proposal → Gin Backend → Frontend, 只做文本提案）
    – 提交提案 (biya-core/chain: Create Proposal → Gin Backend → Frontend, 第一版先直接上链，第二版通过审批，审批完成后走KMS签名) 
    
