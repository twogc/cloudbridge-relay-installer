#!/usr/bin/env python3
"""
Threat Intelligence Integration для CloudBridge Relay Security
Интеграция с внешними сервисами Threat Intelligence
"""

import requests
import json
import time
import logging
from typing import Dict, List, Optional
from dataclasses import dataclass
from datetime import datetime, timedelta

# Настройка логирования
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

@dataclass
class ThreatIntelResult:
    """Результат проверки Threat Intelligence"""
    indicator: str
    indicator_type: str
    is_malicious: bool
    confidence: float
    sources: List[str]
    first_seen: Optional[str]
    last_seen: Optional[str]
    tags: List[str]
    description: Optional[str]

class ThreatIntelligence:
    """Класс для работы с Threat Intelligence сервисами"""
    
    def __init__(self, config: Dict):
        self.config = config
        self.cache = {}
        self.cache_ttl = config.get('cache_ttl', 3600)  # 1 час по умолчанию
        
        # API ключи
        self.virustotal_api_key = config.get('virustotal_api_key')
        self.abuseipdb_api_key = config.get('abuseipdb_api_key')
        self.alienvault_api_key = config.get('alienvault_api_key')
        
        # Настройки запросов
        self.timeout = config.get('timeout', 10)
        self.max_retries = config.get('max_retries', 3)
        
    def check_ip_reputation(self, ip_address: str) -> Optional[ThreatIntelResult]:
        """Проверка репутации IP адреса"""
        try:
            # Проверяем кэш
            cache_key = f"ip_{ip_address}"
            if cache_key in self.cache:
                cached_result = self.cache[cache_key]
                if time.time() - cached_result['timestamp'] < self.cache_ttl:
                    return cached_result['result']
            
            # Проверяем через AbuseIPDB
            abuse_result = self._check_abuseipdb(ip_address)
            if abuse_result:
                self._cache_result(cache_key, abuse_result)
                return abuse_result
            
            # Проверяем через AlienVault OTX
            otx_result = self._check_alienvault(ip_address)
            if otx_result:
                self._cache_result(cache_key, otx_result)
                return otx_result
                
        except Exception as e:
            logger.error(f"Error checking IP reputation for {ip_address}: {e}")
            
        return None
    
    def check_domain_reputation(self, domain: str) -> Optional[ThreatIntelResult]:
        """Проверка репутации домена"""
        try:
            cache_key = f"domain_{domain}"
            if cache_key in self.cache:
                cached_result = self.cache[cache_key]
                if time.time() - cached_result['timestamp'] < self.cache_ttl:
                    return cached_result['result']
            
            # Проверяем через VirusTotal
            vt_result = self._check_virustotal_domain(domain)
            if vt_result:
                self._cache_result(cache_key, vt_result)
                return vt_result
                
        except Exception as e:
            logger.error(f"Error checking domain reputation for {domain}: {e}")
            
        return None
    
    def check_file_hash(self, file_hash: str) -> Optional[ThreatIntelResult]:
        """Проверка хеша файла"""
        try:
            cache_key = f"hash_{file_hash}"
            if cache_key in self.cache:
                cached_result = self.cache[cache_key]
                if time.time() - cached_result['timestamp'] < self.cache_ttl:
                    return cached_result['result']
            
            # Проверяем через VirusTotal
            vt_result = self._check_virustotal_hash(file_hash)
            if vt_result:
                self._cache_result(cache_key, vt_result)
                return vt_result
                
        except Exception as e:
            logger.error(f"Error checking file hash {file_hash}: {e}")
            
        return None
    
    def _check_abuseipdb(self, ip_address: str) -> Optional[ThreatIntelResult]:
        """Проверка через AbuseIPDB"""
        if not self.abuseipdb_api_key:
            return None
            
        try:
            url = "https://api.abuseipdb.com/api/v2/check"
            headers = {
                'Accept': 'application/json',
                'Key': self.abuseipdb_api_key
            }
            params = {
                'ipAddress': ip_address,
                'maxAgeInDays': '90'
            }
            
            response = requests.get(url, headers=headers, params=params, timeout=self.timeout)
            response.raise_for_status()
            
            data = response.json()
            result = data.get('data', {})
            
            return ThreatIntelResult(
                indicator=ip_address,
                indicator_type='ip',
                is_malicious=result.get('abuseConfidenceScore', 0) > 50,
                confidence=result.get('abuseConfidenceScore', 0) / 100.0,
                sources=['AbuseIPDB'],
                first_seen=result.get('firstSeenAt'),
                last_seen=result.get('lastSeenAt'),
                tags=result.get('usageType', '').split(',') if result.get('usageType') else [],
                description=f"AbuseIPDB confidence: {result.get('abuseConfidenceScore', 0)}%"
            )
            
        except Exception as e:
            logger.error(f"Error checking AbuseIPDB for {ip_address}: {e}")
            return None
    
    def _check_alienvault(self, ip_address: str) -> Optional[ThreatIntelResult]:
        """Проверка через AlienVault OTX"""
        if not self.alienvault_api_key:
            return None
            
        try:
            url = f"https://otx.alienvault.com/api/v1/indicators/IPv4/{ip_address}/general"
            headers = {
                'X-OTX-API-KEY': self.alienvault_api_key
            }
            
            response = requests.get(url, headers=headers, timeout=self.timeout)
            response.raise_for_status()
            
            data = response.json()
            
            # Анализируем результаты
            pulse_count = data.get('pulse_info', {}).get('count', 0)
            is_malicious = pulse_count > 0
            confidence = min(pulse_count / 10.0, 1.0)  # Нормализуем до 0-1
            
            tags = []
            if data.get('pulse_info', {}).get('pulses'):
                for pulse in data['pulse_info']['pulses']:
                    tags.extend(pulse.get('tags', []))
            
            return ThreatIntelResult(
                indicator=ip_address,
                indicator_type='ip',
                is_malicious=is_malicious,
                confidence=confidence,
                sources=['AlienVault OTX'],
                first_seen=data.get('base_indicator', {}).get('created'),
                last_seen=data.get('base_indicator', {}).get('updated'),
                tags=list(set(tags)),  # Убираем дубликаты
                description=f"Found in {pulse_count} threat reports"
            )
            
        except Exception as e:
            logger.error(f"Error checking AlienVault for {ip_address}: {e}")
            return None
    
    def _check_virustotal_domain(self, domain: str) -> Optional[ThreatIntelResult]:
        """Проверка домена через VirusTotal"""
        if not self.virustotal_api_key:
            return None
            
        try:
            url = f"https://www.virustotal.com/vtapi/v2/domain/report"
            params = {
                'apikey': self.virustotal_api_key,
                'domain': domain
            }
            
            response = requests.get(url, params=params, timeout=self.timeout)
            response.raise_for_status()
            
            data = response.json()
            
            # Анализируем результаты
            positives = data.get('positives', 0)
            total = data.get('total', 0)
            is_malicious = positives > 0
            confidence = positives / total if total > 0 else 0.0
            
            return ThreatIntelResult(
                indicator=domain,
                indicator_type='domain',
                is_malicious=is_malicious,
                confidence=confidence,
                sources=['VirusTotal'],
                first_seen=None,  # VirusTotal не предоставляет эту информацию
                last_seen=None,
                tags=data.get('categories', {}).values() if data.get('categories') else [],
                description=f"Detected by {positives}/{total} security vendors"
            )
            
        except Exception as e:
            logger.error(f"Error checking VirusTotal domain for {domain}: {e}")
            return None
    
    def _check_virustotal_hash(self, file_hash: str) -> Optional[ThreatIntelResult]:
        """Проверка хеша файла через VirusTotal"""
        if not self.virustotal_api_key:
            return None
            
        try:
            url = f"https://www.virustotal.com/vtapi/v2/file/report"
            params = {
                'apikey': self.virustotal_api_key,
                'resource': file_hash
            }
            
            response = requests.get(url, params=params, timeout=self.timeout)
            response.raise_for_status()
            
            data = response.json()
            
            # Анализируем результаты
            positives = data.get('positives', 0)
            total = data.get('total', 0)
            is_malicious = positives > 0
            confidence = positives / total if total > 0 else 0.0
            
            return ThreatIntelResult(
                indicator=file_hash,
                indicator_type='hash',
                is_malicious=is_malicious,
                confidence=confidence,
                sources=['VirusTotal'],
                first_seen=data.get('scan_date'),
                last_seen=data.get('scan_date'),
                tags=data.get('scans', {}).keys() if data.get('scans') else [],
                description=f"Detected by {positives}/{total} security vendors"
            )
            
        except Exception as e:
            logger.error(f"Error checking VirusTotal hash for {file_hash}: {e}")
            return None
    
    def _cache_result(self, key: str, result: ThreatIntelResult):
        """Кэширование результата"""
        self.cache[key] = {
            'result': result,
            'timestamp': time.time()
        }
    
    def clear_cache(self):
        """Очистка кэша"""
        self.cache.clear()
    
    def get_cache_stats(self) -> Dict:
        """Получение статистики кэша"""
        return {
            'cache_size': len(self.cache),
            'cache_ttl': self.cache_ttl
        }

def create_threat_intelligence_config():
    """Создание конфигурации Threat Intelligence"""
    return {
        'virustotal_api_key': '21e01e440e9848c64231a9825468a923fcc35bee38fc0c1ea5357dfde4fff4b4',
        'abuseipdb_api_key': '598bcbbbe859dc055668d06101c980c3ea18a259e754e793e26d38db413e09035a31168ed6f93a88',
        'alienvault_api_key': '76ac5ccfe2c5d36591d99642575db97c8305d4bc94b2506b6db0ebbf37bb4b86',
        'cache_ttl': 3600,  # 1 час
        'timeout': 10,
        'max_retries': 3
    }

def main():
    """Основная функция с поддержкой аргументов командной строки"""
    import sys
    
    config = create_threat_intelligence_config()
    ti = ThreatIntelligence(config)
    
    # Проверяем аргументы командной строки
    if len(sys.argv) > 2:
        command = sys.argv[1]
        indicator = sys.argv[2]
        
        if command == '--check-ip':
            result = ti.check_ip_reputation(indicator)
            if result:
                print(f"IP: {result.indicator}")
                print(f"Злонамеренный: {result.is_malicious}")
                print(f"Уверенность: {result.confidence:.2f}")
                print(f"Источники: {', '.join(result.sources)}")
                print(f"Описание: {result.description}")
            else:
                print(f"IP {indicator} не найден в базах угроз")
                
        elif command == '--check-domain':
            result = ti.check_domain_reputation(indicator)
            if result:
                print(f"Домен: {result.indicator}")
                print(f"Злонамеренный: {result.is_malicious}")
                print(f"Уверенность: {result.confidence:.2f}")
                print(f"Источники: {', '.join(result.sources)}")
                print(f"Описание: {result.description}")
            else:
                print(f"Домен {indicator} не найден в базах угроз")
                
        elif command == '--check-hash':
            result = ti.check_file_hash(indicator)
            if result:
                print(f"Хеш: {result.indicator}")
                print(f"Злонамеренный: {result.is_malicious}")
                print(f"Уверенность: {result.confidence:.2f}")
                print(f"Источники: {', '.join(result.sources)}")
                print(f"Описание: {result.description}")
            else:
                print(f"Хеш {indicator} не найден в базах угроз")
        else:
            print("Использование:")
            print("  python3 threat_intelligence.py --check-ip <IP>")
            print("  python3 threat_intelligence.py --check-domain <DOMAIN>")
            print("  python3 threat_intelligence.py --check-hash <HASH>")
    else:
        # Примеры проверок по умолчанию
        test_indicators = [
            ('8.8.8.8', 'ip'),
            ('google.com', 'domain'),
            ('d41d8cd98f00b204e9800998ecf8427e', 'hash')  # MD5 хеш пустой строки
        ]
        
        for indicator, indicator_type in test_indicators:
            print(f"\nПроверка {indicator_type}: {indicator}")
            
            if indicator_type == 'ip':
                result = ti.check_ip_reputation(indicator)
            elif indicator_type == 'domain':
                result = ti.check_domain_reputation(indicator)
            elif indicator_type == 'hash':
                result = ti.check_file_hash(indicator)
            else:
                continue
                
            if result:
                print(f"  Злонамеренный: {result.is_malicious}")
                print(f"  Уверенность: {result.confidence:.2f}")
                print(f"  Источники: {', '.join(result.sources)}")
                print(f"  Теги: {', '.join(result.tags)}")
                print(f"  Описание: {result.description}")
            else:
                print("  Результат не найден")

if __name__ == "__main__":
    main() 