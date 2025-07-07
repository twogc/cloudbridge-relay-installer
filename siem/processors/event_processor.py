#!/usr/bin/env python3
"""
SIEM Event Processor
Обрабатывает и анализирует события безопасности
"""

import json
import logging
import re
from datetime import datetime, timedelta
from typing import Dict, List, Any, Optional
from collections import defaultdict, deque
import ipaddress
from dataclasses import dataclass
from kafka import KafkaConsumer
from elasticsearch import Elasticsearch
import redis

@dataclass
class SecurityEvent:
    """Структура события безопасности"""
    timestamp: datetime
    source: str
    event_type: str
    severity: str
    description: str
    ip_address: Optional[str] = None
    user_id: Optional[str] = None
    session_id: Optional[str] = None
    raw_data: Dict[str, Any] = None
    indicators: List[str] = None
    tags: List[str] = None

class SIEMEventProcessor:
    def __init__(self, config: Dict[str, Any]):
        self.config = config
        self.logger = logging.getLogger(__name__)
        
        # Инициализация компонентов
        self.kafka_consumer = KafkaConsumer(
            'siem-events',
            bootstrap_servers=config['kafka']['bootstrap_servers'],
            value_deserializer=lambda m: json.loads(m.decode('utf-8')),
            group_id='siem-processor-group',
            auto_offset_reset='latest'
        )
        
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
        
        # Кэши для анализа
        self.failed_auth_cache = defaultdict(deque)  # IP -> deque of timestamps
        self.suspicious_ips = set()
        self.user_activity_cache = defaultdict(deque)  # user_id -> deque of activities
        
        # Правила детекции
        self.detection_rules = self._load_detection_rules()
        
    def _load_detection_rules(self) -> List[Dict[str, Any]]:
        """Загрузка правил детекции"""
        return [
            {
                'id': 'brute_force_detection',
                'name': 'Brute Force Attack Detection',
                'description': 'Detect multiple failed authentication attempts',
                'conditions': {
                    'event_type': 'keycloak_operation',
                    'event_type_keycloak': 'LOGIN_ERROR',
                    'threshold': 5,
                    'time_window': 300  # 5 minutes
                },
                'severity': 'high',
                'tags': ['brute_force', 'authentication']
            },
            {
                'id': 'suspicious_ip_detection',
                'name': 'Suspicious IP Detection',
                'description': 'Detect connections from suspicious IP addresses',
                'conditions': {
                    'ip_blacklist': [
                        '192.168.1.100',
                        '10.0.0.50'
                    ]
                },
                'severity': 'medium',
                'tags': ['suspicious_ip', 'network']
            },
            {
                'id': 'unusual_user_activity',
                'name': 'Unusual User Activity',
                'description': 'Detect unusual patterns in user activity',
                'conditions': {
                    'max_events_per_hour': 100,
                    'time_window': 3600  # 1 hour
                },
                'severity': 'medium',
                'tags': ['user_behavior', 'anomaly']
            },
            {
                'id': 'relay_anomaly_detection',
                'name': 'Relay Anomaly Detection',
                'description': 'Detect anomalies in Relay operations',
                'conditions': {
                    'event_type': 'relay_operation',
                    'error_threshold': 10,
                    'time_window': 600  # 10 minutes
                },
                'severity': 'high',
                'tags': ['relay', 'anomaly', 'errors']
            },
            {
                'id': 'data_exfiltration_detection',
                'name': 'Data Exfiltration Detection',
                'description': 'Detect potential data exfiltration',
                'conditions': {
                    'event_type': 'relay_operation',
                    'large_data_transfer': 100 * 1024 * 1024,  # 100MB
                    'time_window': 3600  # 1 hour
                },
                'severity': 'critical',
                'tags': ['data_exfiltration', 'data_loss']
            }
        ]
    
    def process_event(self, event: Dict[str, Any]) -> Optional[SecurityEvent]:
        """Обработка одного события"""
        try:
            # Парсинг события
            parsed_event = self._parse_event(event)
            if not parsed_event:
                return None
            
            # Обогащение события
            enriched_event = self._enrich_event(parsed_event)
            
            # Применение правил детекции
            security_events = self._apply_detection_rules(enriched_event)
            
            # Сохранение в Elasticsearch
            self._save_to_elasticsearch(enriched_event)
            
            return security_events[0] if security_events else None
            
        except Exception as e:
            self.logger.error(f"Error processing event: {e}")
            return None
    
    def _parse_event(self, event: Dict[str, Any]) -> Optional[Dict[str, Any]]:
        """Парсинг события"""
        try:
            timestamp = datetime.fromisoformat(event.get('timestamp', datetime.utcnow().isoformat()))
            
            return {
                'timestamp': timestamp,
                'source': event.get('source', 'unknown'),
                'event_type': event.get('event_type', 'unknown'),
                'data': event.get('data', {}),
                'fields': event.get('fields', {}),
                'metadata': event.get('metadata', {}),
                'raw_data': event
            }
        except Exception as e:
            self.logger.error(f"Error parsing event: {e}")
            return None
    
    def _enrich_event(self, event: Dict[str, Any]) -> Dict[str, Any]:
        """Обогащение события дополнительной информацией"""
        enriched = event.copy()
        
        # Извлечение IP адреса
        ip_address = self._extract_ip_address(event)
        if ip_address:
            enriched['ip_address'] = ip_address
            enriched['ip_info'] = self._get_ip_info(ip_address)
        
        # Извлечение пользователя
        user_id = self._extract_user_id(event)
        if user_id:
            enriched['user_id'] = user_id
            enriched['user_info'] = self._get_user_info(user_id)
        
        # Извлечение сессии
        session_id = self._extract_session_id(event)
        if session_id:
            enriched['session_id'] = session_id
        
        # Геолокация
        if ip_address:
            enriched['geolocation'] = self._get_geolocation(ip_address)
        
        # Угрозы
        enriched['threat_indicators'] = self._detect_threat_indicators(event)
        
        # Поведенческий анализ
        enriched['behavior_score'] = self._calculate_behavior_score(event)
        
        return enriched
    
    def _extract_ip_address(self, event: Dict[str, Any]) -> Optional[str]:
        """Извлечение IP адреса из события"""
        # Проверяем различные поля
        ip_fields = ['ip_address', 'client_ip', 'source_ip', 'remote_addr']
        
        for field in ip_fields:
            if field in event.get('data', {}):
                ip = event['data'][field]
                if self._is_valid_ip(ip):
                    return ip
        
        # Проверяем в полях
        for field in ip_fields:
            if field in event.get('fields', {}):
                ip = event['fields'][field]
                if self._is_valid_ip(ip):
                    return ip
        
        return None
    
    def _extract_user_id(self, event: Dict[str, Any]) -> Optional[str]:
        """Извлечение ID пользователя"""
        user_fields = ['user_id', 'username', 'user', 'subject']
        
        for field in user_fields:
            if field in event.get('data', {}):
                return event['data'][field]
        
        for field in user_fields:
            if field in event.get('fields', {}):
                return event['fields'][field]
        
        return None
    
    def _extract_session_id(self, event: Dict[str, Any]) -> Optional[str]:
        """Извлечение ID сессии"""
        session_fields = ['session_id', 'session', 'sid']
        
        for field in session_fields:
            if field in event.get('data', {}):
                return event['data'][field]
        
        return None
    
    def _is_valid_ip(self, ip: str) -> bool:
        """Проверка валидности IP адреса"""
        try:
            ipaddress.ip_address(ip)
            return True
        except ValueError:
            return False
    
    def _get_ip_info(self, ip: str) -> Dict[str, Any]:
        """Получение информации об IP адресе"""
        # Проверяем кэш
        cache_key = f"ip_info:{ip}"
        cached_info = self.redis.get(cache_key)
        
        if cached_info:
            return json.loads(cached_info)
        
        # Получаем информацию
        ip_info = {
            'is_private': ipaddress.ip_address(ip).is_private,
            'is_loopback': ipaddress.ip_address(ip).is_loopback,
            'is_multicast': ipaddress.ip_address(ip).is_multicast,
            'is_reserved': ipaddress.ip_address(ip).is_reserved,
            'version': ipaddress.ip_address(ip).version
        }
        
        # Кэшируем на 1 час
        self.redis.setex(cache_key, 3600, json.dumps(ip_info))
        
        return ip_info
    
    def _get_user_info(self, user_id: str) -> Dict[str, Any]:
        """Получение информации о пользователе"""
        # Проверяем кэш
        cache_key = f"user_info:{user_id}"
        cached_info = self.redis.get(cache_key)
        
        if cached_info:
            return json.loads(cached_info)
        
        # Получаем информацию из Keycloak или базы данных
        user_info = {
            'id': user_id,
            'last_seen': datetime.utcnow().isoformat(),
            'activity_count': 0
        }
        
        # Кэшируем на 30 минут
        self.redis.setex(cache_key, 1800, json.dumps(user_info))
        
        return user_info
    
    def _get_geolocation(self, ip: str) -> Dict[str, Any]:
        """Получение геолокации IP адреса"""
        if ipaddress.ip_address(ip).is_private:
            return {'country': 'Private', 'city': 'Private'}
        
        # Здесь можно интегрировать с сервисами геолокации
        # Например, MaxMind GeoIP2, ipapi.co и т.д.
        return {'country': 'Unknown', 'city': 'Unknown'}
    
    def _detect_threat_indicators(self, event: Dict[str, Any]) -> List[str]:
        """Детекция индикаторов угроз"""
        indicators = []
        
        # Проверяем на подозрительные паттерны
        message = str(event.get('data', {}).get('message', '')).lower()
        
        # SQL Injection
        sql_patterns = ['union select', 'drop table', 'insert into', 'delete from']
        for pattern in sql_patterns:
            if pattern in message:
                indicators.append('sql_injection')
        
        # XSS
        xss_patterns = ['<script>', 'javascript:', 'onload=', 'onerror=']
        for pattern in xss_patterns:
            if pattern in message:
                indicators.append('xss')
        
        # Path Traversal
        path_patterns = ['../', '..\\', '/etc/passwd', 'c:\\windows']
        for pattern in path_patterns:
            if pattern in message:
                indicators.append('path_traversal')
        
        # Command Injection
        cmd_patterns = ['; rm', '| cat', '&& rm', '`whoami`']
        for pattern in cmd_patterns:
            if pattern in message:
                indicators.append('command_injection')
        
        return indicators
    
    def _calculate_behavior_score(self, event: Dict[str, Any]) -> float:
        """Расчет поведенческого скора"""
        score = 0.0
        
        # Базовый скор
        if event.get('source') == 'keycloak_events':
            if event.get('data', {}).get('event_type_keycloak') == 'LOGIN_ERROR':
                score += 10.0
            elif event.get('data', {}).get('event_type_keycloak') == 'LOGIN':
                score += 1.0
        
        # Учитываем время
        hour = event['timestamp'].hour
        if hour < 6 or hour > 22:  # Ночное время
            score += 5.0
        
        # Учитываем IP
        ip_address = event.get('ip_address')
        if ip_address:
            ip_info = event.get('ip_info', {})
            if ip_info.get('is_private', False):
                score -= 5.0  # Снижаем скор для внутренних IP
            else:
                score += 10.0  # Повышаем скор для внешних IP
        
        # Учитываем индикаторы угроз
        threat_indicators = event.get('threat_indicators', [])
        score += len(threat_indicators) * 20.0
        
        return max(0.0, min(100.0, score))  # Ограничиваем от 0 до 100
    
    def _apply_detection_rules(self, event: Dict[str, Any]) -> List[SecurityEvent]:
        """Применение правил детекции"""
        security_events = []
        
        for rule in self.detection_rules:
            if self._evaluate_rule(rule, event):
                security_event = self._create_security_event(rule, event)
                security_events.append(security_event)
        
        return security_events
    
    def _evaluate_rule(self, rule: Dict[str, Any], event: Dict[str, Any]) -> bool:
        """Оценка правила детекции"""
        rule_id = rule['id']
        conditions = rule['conditions']
        
        if rule_id == 'brute_force_detection':
            return self._evaluate_brute_force_rule(conditions, event)
        elif rule_id == 'suspicious_ip_detection':
            return self._evaluate_suspicious_ip_rule(conditions, event)
        elif rule_id == 'unusual_user_activity':
            return self._evaluate_unusual_activity_rule(conditions, event)
        elif rule_id == 'relay_anomaly_detection':
            return self._evaluate_relay_anomaly_rule(conditions, event)
        elif rule_id == 'data_exfiltration_detection':
            return self._evaluate_data_exfiltration_rule(conditions, event)
        
        return False
    
    def _evaluate_brute_force_rule(self, conditions: Dict[str, Any], event: Dict[str, Any]) -> bool:
        """Оценка правила brute force"""
        if (event.get('event_type') != conditions['event_type'] or
            event.get('data', {}).get('event_type_keycloak') != conditions['event_type_keycloak']):
            return False
        
        ip_address = event.get('ip_address')
        if not ip_address:
            return False
        
        # Проверяем количество неудачных попыток
        threshold = conditions['threshold']
        time_window = conditions['time_window']
        
        now = datetime.utcnow()
        cutoff_time = now - timedelta(seconds=time_window)
        
        # Очищаем старые записи
        while (self.failed_auth_cache[ip_address] and 
               self.failed_auth_cache[ip_address][0] < cutoff_time):
            self.failed_auth_cache[ip_address].popleft()
        
        # Добавляем текущую попытку
        self.failed_auth_cache[ip_address].append(now)
        
        # Проверяем превышение порога
        return len(self.failed_auth_cache[ip_address]) >= threshold
    
    def _evaluate_suspicious_ip_rule(self, conditions: Dict[str, Any], event: Dict[str, Any]) -> bool:
        """Оценка правила подозрительных IP"""
        ip_address = event.get('ip_address')
        if not ip_address:
            return False
        
        blacklist = conditions.get('ip_blacklist', [])
        return ip_address in blacklist
    
    def _evaluate_unusual_activity_rule(self, conditions: Dict[str, Any], event: Dict[str, Any]) -> bool:
        """Оценка правила необычной активности пользователя"""
        user_id = event.get('user_id')
        if not user_id:
            return False
        
        max_events = conditions['max_events_per_hour']
        time_window = conditions['time_window']
        
        now = datetime.utcnow()
        cutoff_time = now - timedelta(seconds=time_window)
        
        # Очищаем старые записи
        while (self.user_activity_cache[user_id] and 
               self.user_activity_cache[user_id][0] < cutoff_time):
            self.user_activity_cache[user_id].popleft()
        
        # Добавляем текущую активность
        self.user_activity_cache[user_id].append(now)
        
        # Проверяем превышение лимита
        return len(self.user_activity_cache[user_id]) > max_events
    
    def _evaluate_relay_anomaly_rule(self, conditions: Dict[str, Any], event: Dict[str, Any]) -> bool:
        """Оценка правила аномалий Relay"""
        if event.get('event_type') != conditions['event_type']:
            return False
        
        # Здесь можно добавить более сложную логику
        # Например, анализ трендов ошибок
        return False
    
    def _evaluate_data_exfiltration_rule(self, conditions: Dict[str, Any], event: Dict[str, Any]) -> bool:
        """Оценка правила эксфильтрации данных"""
        if event.get('event_type') != conditions['event_type']:
            return False
        
        # Проверяем размер передаваемых данных
        data_size = event.get('data', {}).get('bytes_transferred', 0)
        threshold = conditions['large_data_transfer']
        
        return data_size > threshold
    
    def _create_security_event(self, rule: Dict[str, Any], event: Dict[str, Any]) -> SecurityEvent:
        """Создание события безопасности"""
        return SecurityEvent(
            timestamp=event['timestamp'],
            source=event['source'],
            event_type=rule['id'],
            severity=rule['severity'],
            description=rule['description'],
            ip_address=event.get('ip_address'),
            user_id=event.get('user_id'),
            session_id=event.get('session_id'),
            raw_data=event,
            indicators=event.get('threat_indicators', []),
            tags=rule.get('tags', [])
        )
    
    def _save_to_elasticsearch(self, event: Dict[str, Any]):
        """Сохранение события в Elasticsearch"""
        try:
            self.elasticsearch.index(
                index='siem-events',
                document=event
            )
        except Exception as e:
            self.logger.error(f"Error saving to Elasticsearch: {e}")
    
    def run_processing_loop(self):
        """Основной цикл обработки событий"""
        self.logger.info("Starting SIEM event processing loop")
        
        for message in self.kafka_consumer:
            try:
                event = message.value
                security_event = self.process_event(event)
                
                if security_event:
                    self.logger.warning(f"Security event detected: {security_event.event_type} - {security_event.description}")
                    # Здесь можно добавить отправку уведомлений
                    
            except Exception as e:
                self.logger.error(f"Error processing message: {e}")

if __name__ == "__main__":
    # Конфигурация
    config = {
        'kafka': {
            'bootstrap_servers': ['localhost:9092']
        },
        'elasticsearch': {
            'hosts': ['http://localhost:9200'],
            'username': 'elastic',
            'password': 'changeme'
        },
        'redis': {
            'host': 'localhost',
            'port': 6379,
            'password': None
        }
    }
    
    # Настройка логирования
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )
    
    # Запуск процессора
    processor = SIEMEventProcessor(config)
    processor.run_processing_loop() 