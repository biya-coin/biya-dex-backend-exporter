#!/bin/bash

# Biya Chain 告警规则测试脚本
# 用途：验证 Prometheus 告警规则和 Alertmanager 配置

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

PROMETHEUS_URL="${PROMETHEUS_URL:-http://localhost:9090}"
ALERTMANAGER_URL="${ALERTMANAGER_URL:-http://localhost:9093}"

echo -e "${BLUE}=== Biya Chain 告警配置测试 ===${NC}\n"

# 1. 检查 Prometheus 服务状态
echo -e "${YELLOW}[1/6] 检查 Prometheus 服务状态...${NC}"
if curl -s "${PROMETHEUS_URL}/-/healthy" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Prometheus 服务运行正常${NC}"
else
    echo -e "${RED}✗ Prometheus 服务不可用，请检查服务是否启动${NC}"
    echo -e "${YELLOW}  提示：运行 'docker-compose up -d prometheus' 启动服务${NC}"
    exit 1
fi

# 2. 检查 Alertmanager 服务状态
echo -e "\n${YELLOW}[2/6] 检查 Alertmanager 服务状态...${NC}"
if curl -s "${ALERTMANAGER_URL}/-/healthy" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Alertmanager 服务运行正常${NC}"
else
    echo -e "${RED}✗ Alertmanager 服务不可用，请检查服务是否启动${NC}"
    echo -e "${YELLOW}  提示：运行 'docker-compose up -d alertmanager' 启动服务${NC}"
    exit 1
fi

# 3. 验证告警规则语法
echo -e "\n${YELLOW}[3/6] 验证告警规则语法...${NC}"
if command -v promtool &> /dev/null; then
    if promtool check rules configs/prometheus/alert_rules.yml 2>&1 | grep -q "SUCCESS"; then
        echo -e "${GREEN}✓ 告警规则语法正确${NC}"
    else
        echo -e "${RED}✗ 告警规则语法错误${NC}"
        promtool check rules configs/prometheus/alert_rules.yml
        exit 1
    fi
else
    echo -e "${YELLOW}⚠ promtool 未安装，跳过本地语法检查${NC}"
fi

# 4. 验证 Alertmanager 配置
echo -e "\n${YELLOW}[4/6] 验证 Alertmanager 配置...${NC}"
if command -v amtool &> /dev/null; then
    if amtool check-config configs/prometheus/alertmanager.yml > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Alertmanager 配置正确${NC}"
    else
        echo -e "${RED}✗ Alertmanager 配置错误${NC}"
        amtool check-config configs/prometheus/alertmanager.yml
        exit 1
    fi
else
    echo -e "${YELLOW}⚠ amtool 未安装，跳过本地配置检查${NC}"
fi

# 5. 检查已加载的告警规则
echo -e "\n${YELLOW}[5/6] 检查已加载的告警规则...${NC}"
RULES_RESPONSE=$(curl -s "${PROMETHEUS_URL}/api/v1/rules")
RULES_COUNT=$(echo "$RULES_RESPONSE" | jq -r '.data.groups[].rules | length' 2>/dev/null | awk '{s+=$1} END {print s}')

if [ -n "$RULES_COUNT" ] && [ "$RULES_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✓ 已加载 ${RULES_COUNT} 条告警规则${NC}"
    
    # 显示告警规则名称
    echo -e "\n${BLUE}已加载的告警规则：${NC}"
    echo "$RULES_RESPONSE" | jq -r '.data.groups[].rules[] | select(.type=="alerting") | "  - \(.name) (\(.labels.severity))"' 2>/dev/null || echo "  无法解析规则列表"
else
    echo -e "${RED}✗ 未找到已加载的告警规则${NC}"
    echo -e "${YELLOW}  提示：检查 prometheus.yml 中的 rule_files 配置${NC}"
fi

# 6. 检查当前活跃告警
echo -e "\n${YELLOW}[6/6] 检查当前活跃告警...${NC}"
ALERTS_RESPONSE=$(curl -s "${PROMETHEUS_URL}/api/v1/alerts")
ACTIVE_ALERTS=$(echo "$ALERTS_RESPONSE" | jq -r '.data.alerts[] | select(.state=="firing") | .labels.alertname' 2>/dev/null | wc -l)

if [ "$ACTIVE_ALERTS" -gt 0 ]; then
    echo -e "${YELLOW}⚠ 发现 ${ACTIVE_ALERTS} 个活跃告警${NC}"
    echo -e "\n${BLUE}活跃告警列表：${NC}"
    echo "$ALERTS_RESPONSE" | jq -r '.data.alerts[] | select(.state=="firing") | "  - \(.labels.alertname) (\(.labels.severity)) - \(.annotations.summary)"' 2>/dev/null
else
    echo -e "${GREEN}✓ 当前没有活跃告警${NC}"
fi

# 7. 显示 Alertmanager 连接状态
echo -e "\n${YELLOW}[额外] 检查 Alertmanager 连接状态...${NC}"
AM_STATUS=$(curl -s "${PROMETHEUS_URL}/api/v1/alertmanagers")
AM_ACTIVE=$(echo "$AM_STATUS" | jq -r '.data.activeAlertmanagers | length' 2>/dev/null)

if [ "$AM_ACTIVE" -gt 0 ]; then
    echo -e "${GREEN}✓ Prometheus 已连接到 ${AM_ACTIVE} 个 Alertmanager${NC}"
    echo "$AM_STATUS" | jq -r '.data.activeAlertmanagers[] | "  - \(.url)"' 2>/dev/null
else
    echo -e "${YELLOW}⚠ Prometheus 未连接到 Alertmanager${NC}"
    echo -e "${YELLOW}  提示：检查 prometheus.yml 中的 alerting 配置${NC}"
fi

# 总结
echo -e "\n${BLUE}=== 测试完成 ===${NC}"
echo -e "${GREEN}✓ 所有关键检查已完成${NC}"
echo -e "\n${BLUE}访问地址：${NC}"
echo -e "  - Prometheus UI:  ${PROMETHEUS_URL}"
echo -e "  - 告警规则页面:   ${PROMETHEUS_URL}/alerts"
echo -e "  - Alertmanager:   ${ALERTMANAGER_URL}"
echo -e "\n${BLUE}提示：${NC}"
echo -e "  - 使用 'docker-compose logs prometheus' 查看 Prometheus 日志"
echo -e "  - 使用 'docker-compose logs alertmanager' 查看 Alertmanager 日志"
echo -e "  - 访问 ${PROMETHEUS_URL}/alerts 查看所有告警规则和状态"
echo -e "  - 访问 ${ALERTMANAGER_URL} 管理告警和静默规则"

exit 0

