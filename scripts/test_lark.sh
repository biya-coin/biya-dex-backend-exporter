#!/bin/bash

# 飞书机器人消息测试脚本

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

PROXY_URL="${LARK_PROXY_URL:-http://localhost:5001}"

echo -e "${BLUE}=== 飞书机器人消息测试 ===${NC}\n"

# 1. 检查代理服务状态
echo -e "${YELLOW}[1/4] 检查飞书 Webhook 代理服务...${NC}"
if curl -s "${PROXY_URL}/health" > /dev/null 2>&1; then
    HEALTH_RESPONSE=$(curl -s "${PROXY_URL}/health")
    echo -e "${GREEN}✓ 代理服务运行正常${NC}"
    echo "$HEALTH_RESPONSE" | jq '.' 2>/dev/null || echo "$HEALTH_RESPONSE"
else
    echo -e "${RED}✗ 代理服务不可用${NC}"
    echo -e "${YELLOW}  提示：运行 'docker-compose up -d lark-webhook-proxy' 启动服务${NC}"
    exit 1
fi

# 2. 发送测试消息 - 信息级别
echo -e "\n${YELLOW}[2/4] 发送测试消息（信息级别）...${NC}"
TEST_INFO=$(cat <<EOF
{
  "title": "🔵 [信息] Biya Chain 告警系统测试",
  "content": "这是一条信息级别的测试消息\\n\\n**测试项目**:\\n- 连接正常\\n- 格式正确\\n- 消息发送成功\\n\\n**测试时间**: $(date '+%Y-%m-%d %H:%M:%S')",
  "severity": "info"
}
EOF
)

RESPONSE=$(curl -s -X POST "${PROXY_URL}/test" \
  -H "Content-Type: application/json" \
  -d "$TEST_INFO")

if echo "$RESPONSE" | jq -e '.status == "success"' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ 信息级别消息发送成功${NC}"
else
    echo -e "${RED}✗ 消息发送失败${NC}"
    echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"
fi

sleep 2

# 3. 发送测试消息 - 警告级别
echo -e "\n${YELLOW}[3/4] 发送测试消息（警告级别）...${NC}"
TEST_WARNING=$(cat <<EOF
{
  "title": "🟡 [警告] TPS 下降预警",
  "content": "**告警摘要**: TPS 出现下降趋势\\n**当前 TPS**: 150\\n**正常范围**: > 200\\n**持续时间**: 3 分钟\\n\\n**处理建议**:\\n1. 检查节点运行状态\\n2. 查看网络连接\\n3. 分析交易来源\\n\\n**监控面板**: http://localhost:9090",
  "severity": "warning"
}
EOF
)

RESPONSE=$(curl -s -X POST "${PROXY_URL}/test" \
  -H "Content-Type: application/json" \
  -d "$TEST_WARNING")

if echo "$RESPONSE" | jq -e '.status == "success"' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ 警告级别消息发送成功${NC}"
else
    echo -e "${RED}✗ 消息发送失败${NC}"
    echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"
fi

sleep 2

# 4. 发送测试消息 - 严重级别
echo -e "\n${YELLOW}[4/4] 发送测试消息（严重级别）...${NC}"
TEST_CRITICAL=$(cat <<EOF
{
  "title": "🟠 [严重] 交易成功率下降",
  "content": "**告警摘要**: 交易成功率低于 98%\\n**当前成功率**: 95.5%\\n**阈值**: 98%\\n**影响范围**: 全链\\n\\n**处理建议**:\\n1. 立即查看失败交易原因\\n2. 检查是否有合约攻击\\n3. 分析失败交易的共同特征\\n4. 紧急处理高危问题\\n\\n**紧急联系**: oncall@biya.chain",
  "severity": "critical"
}
EOF
)

RESPONSE=$(curl -s -X POST "${PROXY_URL}/test" \
  -H "Content-Type: application/json" \
  -d "$TEST_CRITICAL")

if echo "$RESPONSE" | jq -e '.status == "success"' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ 严重级别消息发送成功${NC}"
else
    echo -e "${RED}✗ 消息发送失败${NC}"
    echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"
fi

# 总结
echo -e "\n${BLUE}=== 测试完成 ===${NC}"
echo -e "${GREEN}✓ 所有测试消息已发送${NC}"
echo -e "\n${BLUE}请检查飞书群组，确认是否收到以下消息：${NC}"
echo -e "  1. 🔵 信息级别测试消息"
echo -e "  2. 🟡 警告级别测试消息"
echo -e "  3. 🟠 严重级别测试消息"
echo -e "\n${YELLOW}提示：${NC}"
echo -e "  - 如果未收到消息，请检查飞书机器人 webhook URL 是否正确"
echo -e "  - 查看代理服务日志: docker-compose logs lark-webhook-proxy"
echo -e "  - 访问代理服务健康检查: ${PROXY_URL}/health"

# 可选：发送模拟的 Alertmanager webhook
if [ "$1" = "--alertmanager-test" ]; then
    echo -e "\n${YELLOW}[额外] 发送模拟 Alertmanager Webhook...${NC}"
    
    ALERTMANAGER_WEBHOOK=$(cat <<'EOF'
{
  "status": "firing",
  "groupLabels": {
    "alertname": "TPSAbnormalDrop"
  },
  "commonLabels": {
    "severity": "warning",
    "category": "performance",
    "subsystem": "transaction"
  },
  "commonAnnotations": {
    "summary": "TPS异常下降",
    "description": "当前TPS (45.2) 低于历史7天平均值的50%，已持续超过5分钟",
    "处理建议": "1. 检查节点运行状态\n2. 查看是否有网络分区\n3. 分析交易来源是否异常\n4. 必要时通知技术团队",
    "runbook_url": "https://docs.biya.chain/runbooks/tps-drop",
    "dashboard": "https://grafana.biya.chain/d/chain-performance"
  },
  "alerts": [
    {
      "status": "firing",
      "startsAt": "2025-12-24T10:30:00Z",
      "labels": {
        "alertname": "TPSAbnormalDrop",
        "severity": "warning"
      }
    }
  ]
}
EOF
)
    
    RESPONSE=$(curl -s -X POST "${PROXY_URL}/webhook/lark" \
      -H "Content-Type: application/json" \
      -d "$ALERTMANAGER_WEBHOOK")
    
    if echo "$RESPONSE" | jq -e '.status == "success"' > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Alertmanager webhook 测试成功${NC}"
    else
        echo -e "${RED}✗ Alertmanager webhook 测试失败${NC}"
        echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"
    fi
fi

exit 0

