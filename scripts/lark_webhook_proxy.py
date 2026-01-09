#!/usr/bin/env python3
# -*- coding: utf-8 -*-

"""
飞书 Webhook 转换代理服务
将 Alertmanager 的 webhook 消息格式转换为飞书机器人所需的格式
"""

import json
import logging
import requests
from datetime import datetime
from flask import Flask, request, jsonify

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

app = Flask(__name__)

# 飞书机器人 Webhook URL
LARK_WEBHOOK_URL = "https://open.larksuite.com/open-apis/bot/v2/hook/020ec13d-fd66-4910-8636-5fd213c903e3"
# 告警级别对应的颜色和图标
SEVERITY_CONFIG = {
    'emergency': {'color': 'red', 'icon': '🔴', 'name': '紧急'},
    'critical': {'color': 'orange', 'icon': '🟠', 'name': '严重'},
    'warning': {'color': 'yellow', 'icon': '🟡', 'name': '警告'},
    'info': {'color': 'blue', 'icon': '🔵', 'name': '信息'}
}

def format_alert_message(alert_data):
    """
    格式化告警消息为飞书富文本格式
    """
    alerts = alert_data.get('alerts', [])
    status = alert_data.get('status', 'firing')
    group_labels = alert_data.get('groupLabels', {})
    common_labels = alert_data.get('commonLabels', {})
    common_annotations = alert_data.get('commonAnnotations', {})
    
    # 获取告警级别
    severity = common_labels.get('severity', 'info')
    severity_config = SEVERITY_CONFIG.get(severity, SEVERITY_CONFIG['info'])
    
    # 构建标题
    alertname = group_labels.get('alertname', '未知告警')
    status_text = '已恢复' if status == 'resolved' else '触发'
    title = f"{severity_config['icon']} [{severity_config['name']}] {alertname} - {status_text}"
    
    # 构建消息内容
    content_lines = []
    
    # 告警摘要
    if 'summary' in common_annotations:
        content_lines.append(f"**告警摘要**: {common_annotations['summary']}")
    
    # 告警描述
    if 'description' in common_annotations:
        content_lines.append(f"**详细描述**: {common_annotations['description']}")
    
    # 告警数量
    firing_count = len([a for a in alerts if a.get('status') == 'firing'])
    resolved_count = len([a for a in alerts if a.get('status') == 'resolved'])
    if firing_count > 0:
        content_lines.append(f"**触发数量**: {firing_count} 个")
    if resolved_count > 0:
        content_lines.append(f"**恢复数量**: {resolved_count} 个")
    
    # 告警时间
    if alerts:
        first_alert = alerts[0]
        starts_at = first_alert.get('startsAt', '')
        if starts_at:
            try:
                dt = datetime.fromisoformat(starts_at.replace('Z', '+00:00'))
                content_lines.append(f"**开始时间**: {dt.strftime('%Y-%m-%d %H:%M:%S')}")
            except:
                pass
    
    # 告警标签
    if common_labels:
        labels_text = ', '.join([f"{k}={v}" for k, v in common_labels.items() if k != 'severity'])
        if labels_text:
            content_lines.append(f"**标签**: {labels_text}")
    
    # 处理建议
    if '处理建议' in common_annotations:
        content_lines.append(f"\n**处理建议**:\n{common_annotations['处理建议']}")
    
    # Runbook 链接
    if 'runbook_url' in common_annotations:
        content_lines.append(f"\n**处理手册**: {common_annotations['runbook_url']}")
    
    # Dashboard 链接
    if 'dashboard' in common_annotations:
        content_lines.append(f"**监控面板**: {common_annotations['dashboard']}")
    
    content = '\n'.join(content_lines)
    
    return title, content

def send_to_lark_text(title, content):
    """
    发送文本消息到飞书
    """
    message = {
        "msg_type": "text",
        "content": {
            "text": f"{title}\n\n{content}"
        }
    }
    
    try:
        response = requests.post(
            LARK_WEBHOOK_URL,
            json=message,
            headers={'Content-Type': 'application/json'},
            timeout=10
        )
        response.raise_for_status()
        result = response.json()
        
        if result.get('code') == 0:
            logger.info(f"消息发送成功: {title}")
            return True
        else:
            logger.error(f"消息发送失败: {result}")
            return False
    except Exception as e:
        logger.error(f"发送消息到飞书失败: {e}")
        return False

def send_to_lark_card(title, content, severity='info'):
    """
    发送卡片消息到飞书（更美观）
    """
    severity_config = SEVERITY_CONFIG.get(severity, SEVERITY_CONFIG['info'])
    
    # 构建富文本内容
    card_content = []
    
    for line in content.split('\n'):
        if line.strip():
            if line.startswith('**') and line.endswith('**'):
                # 标题行
                card_content.append({
                    "tag": "div",
                    "text": {
                        "tag": "lark_md",
                        "content": line
                    }
                })
            else:
                # 普通行
                card_content.append({
                    "tag": "div",
                    "text": {
                        "tag": "plain_text",
                        "content": line.replace('**', '')
                    }
                })
    
    message = {
        "msg_type": "interactive",
        "card": {
            "header": {
                "title": {
                    "tag": "plain_text",
                    "content": title
                },
                "template": severity_config['color']
            },
            "elements": card_content
        }
    }
    
    try:
        response = requests.post(
            LARK_WEBHOOK_URL,
            json=message,
            headers={'Content-Type': 'application/json'},
            timeout=10
        )
        response.raise_for_status()
        result = response.json()
        
        if result.get('code') == 0:
            logger.info(f"卡片消息发送成功: {title}")
            return True
        else:
            logger.error(f"卡片消息发送失败: {result}")
            # 如果卡片格式失败，尝试发送文本消息
            return send_to_lark_text(title, content)
    except Exception as e:
        logger.error(f"发送卡片消息到飞书失败: {e}")
        # 如果卡片格式失败，尝试发送文本消息
        return send_to_lark_text(title, content)

@app.route('/webhook/lark', methods=['POST'])
def webhook_handler():
    """
    接收 Alertmanager 的 webhook 并转发到飞书
    """
    try:
        # 获取 Alertmanager 发送的数据
        alert_data = request.json
        
        logger.info(f"收到告警通知: {json.dumps(alert_data, ensure_ascii=False, indent=2)}")
        
        # 格式化消息
        title, content = format_alert_message(alert_data)
        
        # 获取告警级别
        severity = alert_data.get('commonLabels', {}).get('severity', 'info')
        
        # 发送到飞书（优先使用卡片格式）
        success = send_to_lark_card(title, content, severity)
        
        if success:
            return jsonify({
                'status': 'success',
                'message': '消息已发送到飞书'
            }), 200
        else:
            return jsonify({
                'status': 'error',
                'message': '消息发送失败'
            }), 500
            
    except Exception as e:
        logger.error(f"处理 webhook 请求失败: {e}", exc_info=True)
        return jsonify({
            'status': 'error',
            'message': str(e)
        }), 500

@app.route('/health', methods=['GET'])
def health_check():
    """
    健康检查端点
    """
    return jsonify({
        'status': 'healthy',
        'service': 'lark-webhook-proxy',
        'lark_webhook': LARK_WEBHOOK_URL
    }), 200

@app.route('/test', methods=['POST'])
def test_message():
    """
    测试端点：直接发送测试消息到飞书
    """
    try:
        test_data = request.json or {}
        title = test_data.get('title', '测试消息')
        content = test_data.get('content', '这是一条测试消息')
        severity = test_data.get('severity', 'info')
        
        success = send_to_lark_card(title, content, severity)
        
        if success:
            return jsonify({
                'status': 'success',
                'message': '测试消息已发送'
            }), 200
        else:
            return jsonify({
                'status': 'error',
                'message': '测试消息发送失败'
            }), 500
    except Exception as e:
        logger.error(f"发送测试消息失败: {e}")
        return jsonify({
            'status': 'error',
            'message': str(e)
        }), 500

if __name__ == '__main__':
    logger.info("启动飞书 Webhook 转换代理服务...")
    logger.info(f"飞书 Webhook URL: {LARK_WEBHOOK_URL}")
    logger.info("监听端口: 5001")
    
    # 启动 Flask 服务
    app.run(host='0.0.0.0', port=5001, debug=False)

