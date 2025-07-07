#!/usr/bin/env python3
"""
SIEM Alert Manager
Управляет уведомлениями о событиях безопасности
"""

import json
import logging
import time
from datetime import datetime, timedelta
from typing import Dict, List, Any, Optional
from dataclasses import dataclass
import smtplib
import requests
from email.mime.text import MimeText
from email.mime.multipart import MimeMultipart
import telegram
from elasticsearch import Elasticsearch
import redis

@dataclass
class Alert:
    """Структура уведомления"""
    id: str
    timestamp: datetime
    severity: str
    title: str
    description: str
    source: str
    event_type: str
    ip_address: Optional[str] = None
    user_id: Optional[str] = None
    indicators: List[str] = None
    tags: List[str] = None
    status: str = 'new'  # new, acknowledged, resolved
    assigned_to: Optional[str] = None
    notes: List[str] = None

class SIEMAlertManager:
    def __init__(self, config: Dict[str, Any]):
        self.config = config
        self.logger = logging.getLogger(__name__)
        
        # Инициализация компонентов
        self.elasticsearch = Elasticsearch(
            hosts=config['elasticsearch']['hosts'],
            basic_auth=(
                config['elasticsearch']['username'],
                config['elasticsearch']['password']
            )
        )
        
        self.redis = redis.Redis(
            host=config['redis']['host'],
            port=config['redis']['port'],
            password=config['redis']['password'],
            decode_responses=True
        )
        
        # Настройки уведомлений
        self.notification_config = config.get('notifications', {})
        
        # Инициализация каналов уведомлений
        self._init_notification_channels()
        
        # Кэш для предотвращения дублирования
        self.alert_cache = {}
        self.alert_cache_ttl = 300  # 5 минут
        
    def _init_notification_channels(self):
        """Инициализация каналов уведомлений"""
        # Email
        if 'email' in self.notification_config:
            self.email_config = self.notification_config['email']
        
        # Telegram
        if 'telegram' in self.notification_config:
            telegram_config = self.notification_config['telegram']
            self.telegram_bot = telegram.Bot(token=telegram_config['bot_token'])
            self.telegram_chat_id = telegram_config['chat_id']
        
        # Slack
        if 'slack' in self.notification_config:
            self.slack_webhook = self.notification_config['slack']['webhook_url']
        
        # Webhook
        if 'webhook' in self.notification_config:
            self.webhook_url = self.notification_config['webhook']['url']
    
    def create_alert(self, security_event: Dict[str, Any]) -> Alert:
        """Создание уведомления из события безопасности"""
        alert_id = f"alert_{int(time.time())}_{hash(str(security_event))}"
        
        alert = Alert(
            id=alert_id,
            timestamp=datetime.utcnow(),
            severity=security_event.get('severity', 'medium'),
            title=self._generate_alert_title(security_event),
            description=security_event.get('description', ''),
            source=security_event.get('source', 'unknown'),
            event_type=security_event.get('event_type', 'unknown'),
            ip_address=security_event.get('ip_address'),
            user_id=security_event.get('user_id'),
            indicators=security_event.get('indicators', []),
            tags=security_event.get('tags', []),
            notes=[]
        )
        
        return alert
    
    def _generate_alert_title(self, security_event: Dict[str, Any]) -> str:
        """Генерация заголовка уведомления"""
        event_type = security_event.get('event_type', 'unknown')
        severity = security_event.get('severity', 'medium')
        
        if event_type == 'brute_force_detection':
            return f"[{severity.upper()}] Brute Force Attack Detected"
        elif event_type == 'suspicious_ip_detection':
            return f"[{severity.upper()}] Suspicious IP Activity"
        elif event_type == 'unusual_user_activity':
            return f"[{severity.upper()}] Unusual User Behavior"
        elif event_type == 'relay_anomaly_detection':
            return f"[{severity.upper()}] Relay System Anomaly"
        elif event_type == 'data_exfiltration_detection':
            return f"[{severity.upper()}] Potential Data Exfiltration"
        else:
            return f"[{severity.upper()}] Security Event: {event_type}"
    
    def should_send_alert(self, alert: Alert) -> bool:
        """Проверка, нужно ли отправлять уведомление"""
        # Проверяем кэш на дублирование
        cache_key = f"alert_sent:{alert.id}"
        if self.redis.exists(cache_key):
            return False
        
        # Проверяем настройки по важности
        severity_levels = {
            'low': 1,
            'medium': 2,
            'high': 3,
            'critical': 4
        }
        
        min_severity = self.notification_config.get('min_severity', 'medium')
        if severity_levels.get(alert.severity, 0) < severity_levels.get(min_severity, 0):
            return False
        
        # Проверяем время тишины
        quiet_hours = self.notification_config.get('quiet_hours', {})
        if quiet_hours:
            current_hour = datetime.utcnow().hour
            start_hour = quiet_hours.get('start', 22)
            end_hour = quiet_hours.get('end', 8)
            
            if start_hour <= end_hour:
                if start_hour <= current_hour <= end_hour:
                    return False
            else:  # Переход через полночь
                if current_hour >= start_hour or current_hour <= end_hour:
                    return False
        
        return True
    
    def send_alert(self, alert: Alert) -> bool:
        """Отправка уведомления"""
        try:
            # Проверяем, нужно ли отправлять
            if not self.should_send_alert(alert):
                return False
            
            # Подготавливаем сообщение
            message = self._format_alert_message(alert)
            
            # Отправляем по всем каналам
            success = True
            
            if hasattr(self, 'email_config'):
                success &= self._send_email_alert(alert, message)
            
            if hasattr(self, 'telegram_bot'):
                success &= self._send_telegram_alert(alert, message)
            
            if hasattr(self, 'slack_webhook'):
                success &= self._send_slack_alert(alert, message)
            
            if hasattr(self, 'webhook_url'):
                success &= self._send_webhook_alert(alert, message)
            
            # Сохраняем в кэш
            if success:
                cache_key = f"alert_sent:{alert.id}"
                self.redis.setex(cache_key, self.alert_cache_ttl, '1')
            
            # Сохраняем в Elasticsearch
            self._save_alert_to_elasticsearch(alert)
            
            return success
            
        except Exception as e:
            self.logger.error(f"Error sending alert: {e}")
            return False
    
    def _format_alert_message(self, alert: Alert) -> str:
        """Форматирование сообщения уведомления"""
        message = f"""
🚨 SECURITY ALERT 🚨

Title: {alert.title}
Severity: {alert.severity.upper()}
Time: {alert.timestamp.strftime('%Y-%m-%d %H:%M:%S UTC')}
Source: {alert.source}
Event Type: {alert.event_type}

Description:
{alert.description}

Details:
"""
        
        if alert.ip_address:
            message += f"IP Address: {alert.ip_address}\n"
        
        if alert.user_id:
            message += f"User ID: {alert.user_id}\n"
        
        if alert.indicators:
            message += f"Threat Indicators: {', '.join(alert.indicators)}\n"
        
        if alert.tags:
            message += f"Tags: {', '.join(alert.tags)}\n"
        
        message += f"""
Alert ID: {alert.id}
Status: {alert.status}

---
CloudBridge SIEM System
Generated at: {datetime.utcnow().strftime('%Y-%m-%d %H:%M:%S UTC')}
"""
        
        return message.strip()
    
    def _send_email_alert(self, alert: Alert, message: str) -> bool:
        """Отправка уведомления по email"""
        try:
            msg = MimeMultipart()
            msg['From'] = self.email_config['from_email']
            msg['To'] = ', '.join(self.email_config['to_emails'])
            msg['Subject'] = alert.title
            
            msg.attach(MimeText(message, 'plain'))
            
            server = smtplib.SMTP(self.email_config['smtp_server'], self.email_config['smtp_port'])
            server.starttls()
            server.login(self.email_config['username'], self.email_config['password'])
            
            text = msg.as_string()
            server.sendmail(self.email_config['from_email'], self.email_config['to_emails'], text)
            server.quit()
            
            self.logger.info(f"Email alert sent: {alert.id}")
            return True
            
        except Exception as e:
            self.logger.error(f"Error sending email alert: {e}")
            return False
    
    def _send_telegram_alert(self, alert: Alert, message: str) -> bool:
        """Отправка уведомления в Telegram"""
        try:
            # Разбиваем длинное сообщение на части
            max_length = 4096
            if len(message) > max_length:
                parts = [message[i:i+max_length] for i in range(0, len(message), max_length)]
                for part in parts:
                    self.telegram_bot.send_message(
                        chat_id=self.telegram_chat_id,
                        text=part,
                        parse_mode='HTML'
                    )
            else:
                self.telegram_bot.send_message(
                    chat_id=self.telegram_chat_id,
                    text=message,
                    parse_mode='HTML'
                )
            
            self.logger.info(f"Telegram alert sent: {alert.id}")
            return True
            
        except Exception as e:
            self.logger.error(f"Error sending Telegram alert: {e}")
            return False
    
    def _send_slack_alert(self, alert: Alert, message: str) -> bool:
        """Отправка уведомления в Slack"""
        try:
            # Определяем цвет в зависимости от важности
            color_map = {
                'low': '#36a64f',      # зеленый
                'medium': '#ff8c00',   # оранжевый
                'high': '#ff0000',     # красный
                'critical': '#8b0000'  # темно-красный
            }
            
            payload = {
                "attachments": [
                    {
                        "color": color_map.get(alert.severity, '#ff8c00'),
                        "title": alert.title,
                        "text": message,
                        "fields": [
                            {
                                "title": "Severity",
                                "value": alert.severity.upper(),
                                "short": True
                            },
                            {
                                "title": "Source",
                                "value": alert.source,
                                "short": True
                            },
                            {
                                "title": "Event Type",
                                "value": alert.event_type,
                                "short": True
                            },
                            {
                                "title": "Alert ID",
                                "value": alert.id,
                                "short": True
                            }
                        ],
                        "footer": "CloudBridge SIEM System",
                        "ts": int(alert.timestamp.timestamp())
                    }
                ]
            }
            
            response = requests.post(self.slack_webhook, json=payload, timeout=10)
            response.raise_for_status()
            
            self.logger.info(f"Slack alert sent: {alert.id}")
            return True
            
        except Exception as e:
            self.logger.error(f"Error sending Slack alert: {e}")
            return False
    
    def _send_webhook_alert(self, alert: Alert, message: str) -> bool:
        """Отправка уведомления через webhook"""
        try:
            payload = {
                'alert_id': alert.id,
                'timestamp': alert.timestamp.isoformat(),
                'severity': alert.severity,
                'title': alert.title,
                'description': alert.description,
                'source': alert.source,
                'event_type': alert.event_type,
                'ip_address': alert.ip_address,
                'user_id': alert.user_id,
                'indicators': alert.indicators,
                'tags': alert.tags,
                'message': message
            }
            
            response = requests.post(self.webhook_url, json=payload, timeout=10)
            response.raise_for_status()
            
            self.logger.info(f"Webhook alert sent: {alert.id}")
            return True
            
        except Exception as e:
            self.logger.error(f"Error sending webhook alert: {e}")
            return False
    
    def _save_alert_to_elasticsearch(self, alert: Alert):
        """Сохранение уведомления в Elasticsearch"""
        try:
            alert_doc = {
                'alert_id': alert.id,
                'timestamp': alert.timestamp.isoformat(),
                'severity': alert.severity,
                'title': alert.title,
                'description': alert.description,
                'source': alert.source,
                'event_type': alert.event_type,
                'ip_address': alert.ip_address,
                'user_id': alert.user_id,
                'indicators': alert.indicators,
                'tags': alert.tags,
                'status': alert.status,
                'assigned_to': alert.assigned_to,
                'notes': alert.notes
            }
            
            self.elasticsearch.index(
                index='siem-alerts',
                document=alert_doc
            )
            
        except Exception as e:
            self.logger.error(f"Error saving alert to Elasticsearch: {e}")
    
    def get_alerts(self, filters: Dict[str, Any] = None) -> List[Alert]:
        """Получение списка уведомлений"""
        try:
            query = {"match_all": {}}
            
            if filters:
                must_conditions = []
                
                if 'severity' in filters:
                    must_conditions.append({"match": {"severity": filters['severity']}})
                
                if 'status' in filters:
                    must_conditions.append({"match": {"status": filters['status']}})
                
                if 'source' in filters:
                    must_conditions.append({"match": {"source": filters['source']}})
                
                if 'time_range' in filters:
                    time_range = filters['time_range']
                    must_conditions.append({
                        "range": {
                            "timestamp": {
                                "gte": time_range['start'],
                                "lte": time_range['end']
                            }
                        }
                    })
                
                if must_conditions:
                    query = {"bool": {"must": must_conditions}}
            
            response = self.elasticsearch.search(
                index='siem-alerts',
                body={
                    "query": query,
                    "sort": [{"timestamp": {"order": "desc"}}],
                    "size": filters.get('limit', 100)
                }
            )
            
            alerts = []
            for hit in response['hits']['hits']:
                source = hit['_source']
                alert = Alert(
                    id=source['alert_id'],
                    timestamp=datetime.fromisoformat(source['timestamp']),
                    severity=source['severity'],
                    title=source['title'],
                    description=source['description'],
                    source=source['source'],
                    event_type=source['event_type'],
                    ip_address=source.get('ip_address'),
                    user_id=source.get('user_id'),
                    indicators=source.get('indicators', []),
                    tags=source.get('tags', []),
                    status=source.get('status', 'new'),
                    assigned_to=source.get('assigned_to'),
                    notes=source.get('notes', [])
                )
                alerts.append(alert)
            
            return alerts
            
        except Exception as e:
            self.logger.error(f"Error getting alerts: {e}")
            return []
    
    def update_alert_status(self, alert_id: str, status: str, assigned_to: str = None, note: str = None):
        """Обновление статуса уведомления"""
        try:
            update_doc = {"status": status}
            
            if assigned_to:
                update_doc["assigned_to"] = assigned_to
            
            if note:
                # Добавляем заметку
                script = {
                    "source": "if (ctx._source.notes == null) { ctx._source.notes = [] } ctx._source.notes.add(params.note)",
                    "lang": "painless",
                    "params": {"note": note}
                }
                update_doc["script"] = script
            
            self.elasticsearch.update_by_query(
                index='siem-alerts',
                body={
                    "query": {"match": {"alert_id": alert_id}},
                    "script": update_doc
                }
            )
            
            self.logger.info(f"Alert {alert_id} status updated to {status}")
            
        except Exception as e:
            self.logger.error(f"Error updating alert status: {e}")

if __name__ == "__main__":
    # Конфигурация
    config = {
        'elasticsearch': {
            'hosts': ['http://localhost:9200'],
            'username': 'elastic',
            'password': 'changeme'
        },
        'redis': {
            'host': 'localhost',
            'port': 6379,
            'password': None
        },
        'notifications': {
            'min_severity': 'medium',
            'quiet_hours': {
                'start': 22,
                'end': 8
            },
            'email': {
                'smtp_server': 'smtp.gmail.com',
                'smtp_port': 587,
                'username': 'your-email@gmail.com',
                'password': 'your-app-password',
                'from_email': 'siem@yourcompany.com',
                'to_emails': ['admin@yourcompany.com', 'security@yourcompany.com']
            },
            'telegram': {
                'bot_token': 'your-bot-token',
                'chat_id': 'your-chat-id'
            },
            'slack': {
                'webhook_url': 'https://hooks.slack.com/services/your/webhook/url'
            },
            'webhook': {
                'url': 'https://your-webhook-endpoint.com/alerts'
            }
        }
    }
    
    # Настройка логирования
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )
    
    # Тестирование
    alert_manager = SIEMAlertManager(config)
    
    # Создание тестового уведомления
    test_event = {
        'severity': 'high',
        'description': 'Multiple failed login attempts detected',
        'source': 'keycloak_events',
        'event_type': 'brute_force_detection',
        'ip_address': '192.168.1.100',
        'user_id': 'test_user',
        'indicators': ['brute_force', 'authentication'],
        'tags': ['security', 'authentication']
    }
    
    alert = alert_manager.create_alert(test_event)
    alert_manager.send_alert(alert) 