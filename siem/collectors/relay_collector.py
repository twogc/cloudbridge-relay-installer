#!/usr/bin/env python3
"""
CloudBridge Relay SIEM Collector
Собирает логи, метрики и события из Relay для SIEM
"""

import json
import logging
import time
from datetime import datetime
from typing import Dict, List, Any
import requests
from prometheus_client import CollectorRegistry, Gauge, Counter, Histogram
from kafka import KafkaProducer
from elasticsearch import Elasticsearch

class RelaySIEMCollector:
    def __init__(self, config: Dict[str, Any]):
        self.config = config
        self.logger = logging.getLogger(__name__)
        
        # Инициализация компонентов
        self.kafka_producer = KafkaProducer(
            bootstrap_servers=config['kafka']['bootstrap_servers'],
            value_serializer=lambda v: json.dumps(v).encode('utf-8')
        )
        
        self.elasticsearch = Elasticsearch(
            hosts=config['elasticsearch']['hosts'],
            basic_auth=(
                config['elasticsearch']['username'],
                config['elasticsearch']['password']
            )
        )
        
        # Prometheus метрики
        self.registry = CollectorRegistry()
        self.events_collected = Counter(
            'siem_events_collected_total',
            'Total events collected',
            ['source', 'event_type'],
            registry=self.registry
        )
        self.collection_duration = Histogram(
            'siem_collection_duration_seconds',
            'Time spent collecting events',
            ['source'],
            registry=self.registry
        )
        
    def collect_relay_logs(self) -> List[Dict[str, Any]]:
        """Сбор логов из CloudBridge Relay"""
        events = []
        
        try:
            # Чтение логов Relay
            log_file = self.config['relay']['log_file']
            with open(log_file, 'r') as f:
                for line in f:
                    try:
                        log_entry = json.loads(line.strip())
                        event = self._enrich_log_event(log_entry)
                        events.append(event)
                        self.events_collected.labels(
                            source='relay', 
                            event_type='log'
                        ).inc()
                    except json.JSONDecodeError:
                        continue
                        
        except Exception as e:
            self.logger.error(f"Error collecting Relay logs: {e}")
            
        return events
    
    def collect_relay_metrics(self) -> List[Dict[str, Any]]:
        """Сбор метрик из Prometheus endpoint"""
        events = []
        
        try:
            metrics_url = self.config['relay']['metrics_url']
            response = requests.get(metrics_url, timeout=10)
            
            if response.status_code == 200:
                metrics_data = response.text
                event = {
                    'timestamp': datetime.utcnow().isoformat(),
                    'source': 'relay_metrics',
                    'event_type': 'metrics_snapshot',
                    'data': metrics_data,
                    'metadata': {
                        'collection_method': 'prometheus_scrape',
                        'url': metrics_url
                    }
                }
                events.append(event)
                self.events_collected.labels(
                    source='relay', 
                    event_type='metrics'
                ).inc()
                
        except Exception as e:
            self.logger.error(f"Error collecting Relay metrics: {e}")
            
        return events
    
    def collect_keycloak_events(self) -> List[Dict[str, Any]]:
        """Сбор событий из Keycloak"""
        events = []
        
        try:
            keycloak_url = self.config['keycloak']['admin_url']
            headers = {
                'Authorization': f"Bearer {self._get_keycloak_token()}",
                'Content-Type': 'application/json'
            }
            
            # Получение событий аутентификации
            auth_events_url = f"{keycloak_url}/events"
            response = requests.get(auth_events_url, headers=headers, timeout=10)
            
            if response.status_code == 200:
                auth_events = response.json()
                for event in auth_events:
                    enriched_event = self._enrich_keycloak_event(event)
                    events.append(enriched_event)
                    self.events_collected.labels(
                        source='keycloak', 
                        event_type='auth'
                    ).inc()
                    
        except Exception as e:
            self.logger.error(f"Error collecting Keycloak events: {e}")
            
        return events
    
    def collect_system_events(self) -> List[Dict[str, Any]]:
        """Сбор системных событий"""
        events = []
        
        try:
            # Сбор событий из systemd journal
            import subprocess
            
            cmd = [
                'journalctl', 
                '--since', '1 hour ago',
                '--output', 'json',
                '--no-pager'
            ]
            
            result = subprocess.run(cmd, capture_output=True, text=True)
            
            if result.returncode == 0:
                for line in result.stdout.strip().split('\n'):
                    if line:
                        try:
                            journal_entry = json.loads(line)
                            event = self._enrich_system_event(journal_entry)
                            events.append(event)
                            self.events_collected.labels(
                                source='system', 
                                event_type='journal'
                            ).inc()
                        except json.JSONDecodeError:
                            continue
                            
        except Exception as e:
            self.logger.error(f"Error collecting system events: {e}")
            
        return events
    
    def _enrich_log_event(self, log_entry: Dict[str, Any]) -> Dict[str, Any]:
        """Обогащение логов Relay"""
        return {
            'timestamp': log_entry.get('timestamp', datetime.utcnow().isoformat()),
            'source': 'relay_logs',
            'event_type': 'relay_operation',
            'level': log_entry.get('level', 'info'),
            'message': log_entry.get('message', ''),
            'fields': log_entry.get('fields', {}),
            'metadata': {
                'enrichment': 'relay_log_parser',
                'version': '1.0'
            }
        }
    
    def _enrich_keycloak_event(self, event: Dict[str, Any]) -> Dict[str, Any]:
        """Обогащение событий Keycloak"""
        return {
            'timestamp': event.get('time', datetime.utcnow().isoformat()),
            'source': 'keycloak_events',
            'event_type': 'keycloak_operation',
            'user_id': event.get('userId', ''),
            'session_id': event.get('sessionId', ''),
            'ip_address': event.get('ipAddress', ''),
            'event_type_keycloak': event.get('type', ''),
            'details': event.get('details', {}),
            'metadata': {
                'enrichment': 'keycloak_event_parser',
                'version': '1.0'
            }
        }
    
    def _enrich_system_event(self, journal_entry: Dict[str, Any]) -> Dict[str, Any]:
        """Обогащение системных событий"""
        return {
            'timestamp': journal_entry.get('__REALTIME_TIMESTAMP', datetime.utcnow().isoformat()),
            'source': 'system_journal',
            'event_type': 'system_operation',
            'unit': journal_entry.get('_SYSTEMD_UNIT', ''),
            'message': journal_entry.get('MESSAGE', ''),
            'priority': journal_entry.get('PRIORITY', ''),
            'hostname': journal_entry.get('_HOSTNAME', ''),
            'metadata': {
                'enrichment': 'system_journal_parser',
                'version': '1.0'
            }
        }
    
    def _get_keycloak_token(self) -> str:
        """Получение токена администратора Keycloak"""
        token_url = f"{self.config['keycloak']['auth_url']}/realms/master/protocol/openid-connect/token"
        
        data = {
            'grant_type': 'password',
            'client_id': 'admin-cli',
            'username': self.config['keycloak']['admin_username'],
            'password': self.config['keycloak']['admin_password']
        }
        
        response = requests.post(token_url, data=data, timeout=10)
        if response.status_code == 200:
            return response.json()['access_token']
        else:
            raise Exception(f"Failed to get Keycloak token: {response.status_code}")
    
    def send_to_kafka(self, events: List[Dict[str, Any]], topic: str = 'siem-events'):
        """Отправка событий в Kafka"""
        for event in events:
            try:
                self.kafka_producer.send(topic, event)
                self.logger.debug(f"Sent event to Kafka topic {topic}")
            except Exception as e:
                self.logger.error(f"Error sending event to Kafka: {e}")
    
    def send_to_elasticsearch(self, events: List[Dict[str, Any]], index: str = 'siem-events'):
        """Отправка событий в Elasticsearch"""
        for event in events:
            try:
                self.elasticsearch.index(
                    index=index,
                    document=event
                )
                self.logger.debug(f"Sent event to Elasticsearch index {index}")
            except Exception as e:
                self.logger.error(f"Error sending event to Elasticsearch: {e}")
    
    def run_collection_cycle(self):
        """Выполнение цикла сбора данных"""
        start_time = time.time()
        
        # Сбор всех типов событий
        all_events = []
        
        with self.collection_duration.labels(source='relay').time():
            relay_logs = self.collect_relay_logs()
            all_events.extend(relay_logs)
        
        with self.collection_duration.labels(source='metrics').time():
            relay_metrics = self.collect_relay_metrics()
            all_events.extend(relay_metrics)
        
        with self.collection_duration.labels(source='keycloak').time():
            keycloak_events = self.collect_keycloak_events()
            all_events.extend(keycloak_events)
        
        with self.collection_duration.labels(source='system').time():
            system_events = self.collect_system_events()
            all_events.extend(system_events)
        
        # Отправка в системы обработки
        if all_events:
            self.send_to_kafka(all_events)
            self.send_to_elasticsearch(all_events)
            
        self.logger.info(f"Collected {len(all_events)} events in {time.time() - start_time:.2f}s")

if __name__ == "__main__":
    # Конфигурация
    config = {
        'relay': {
            'log_file': '/var/log/relay/relay.log',
            'metrics_url': 'http://localhost:8081/metrics'
        },
        'keycloak': {
            'admin_url': 'http://localhost:8080/admin',
            'auth_url': 'http://localhost:8080',
            'admin_username': 'admin',
            'admin_password': 'admin'
        },
        'kafka': {
            'bootstrap_servers': ['localhost:9092']
        },
        'elasticsearch': {
            'hosts': ['http://localhost:9200'],
            'username': 'elastic',
            'password': 'changeme'
        }
    }
    
    # Настройка логирования
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )
    
    # Запуск коллектора
    collector = RelaySIEMCollector(config)
    
    while True:
        try:
            collector.run_collection_cycle()
            time.sleep(60)  # Сбор каждую минуту
        except KeyboardInterrupt:
            break
        except Exception as e:
            logging.error(f"Collection cycle failed: {e}")
            time.sleep(30) 