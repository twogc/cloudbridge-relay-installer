#!/usr/bin/env python3
"""
Скрипт управления CloudBridge SIEM
Предоставляет команды для управления системой
"""

import click
import subprocess
import json
import os
import sys
from datetime import datetime
from typing import Dict, List, Any

@click.group()
def cli():
    """CloudBridge SIEM Management Tool"""
    pass

@cli.command()
@click.option('--config', default='config/siem_config.yaml', help='Путь к конфигурации')
def start(config):
    """Запуск SIEM системы"""
    click.echo("🚀 Запуск CloudBridge SIEM...")
    
    try:
        # Проверка Docker
        result = subprocess.run(['docker', '--version'], capture_output=True, text=True)
        if result.returncode != 0:
            click.echo("❌ Docker не установлен или недоступен")
            sys.exit(1)
        
        # Запуск через docker-compose
        cmd = ['docker-compose', '-f', 'docker-compose.siem.yml', 'up', '-d']
        result = subprocess.run(cmd, cwd=os.getcwd())
        
        if result.returncode == 0:
            click.echo("✅ SIEM система запущена")
            click.echo("📊 Доступные сервисы:")
            click.echo("   - Elasticsearch: http://localhost:9200")
            click.echo("   - Kibana: http://localhost:5601")
            click.echo("   - Grafana: http://localhost:3001")
            click.echo("   - Prometheus: http://localhost:9091")
            click.echo("   - SIEM Web UI: http://localhost:8081")
        else:
            click.echo("❌ Ошибка запуска SIEM")
            sys.exit(1)
            
    except Exception as e:
        click.echo(f"❌ Ошибка: {e}")
        sys.exit(1)

@cli.command()
def stop():
    """Остановка SIEM системы"""
    click.echo("🛑 Остановка CloudBridge SIEM...")
    
    try:
        cmd = ['docker-compose', '-f', 'docker-compose.siem.yml', 'down']
        result = subprocess.run(cmd, cwd=os.getcwd())
        
        if result.returncode == 0:
            click.echo("✅ SIEM система остановлена")
        else:
            click.echo("❌ Ошибка остановки SIEM")
            sys.exit(1)
            
    except Exception as e:
        click.echo(f"❌ Ошибка: {e}")
        sys.exit(1)

@cli.command()
def restart():
    """Перезапуск SIEM системы"""
    click.echo("🔄 Перезапуск CloudBridge SIEM...")
    
    try:
        # Остановка
        stop_cmd = ['docker-compose', '-f', 'docker-compose.siem.yml', 'down']
        subprocess.run(stop_cmd, cwd=os.getcwd())
        
        # Запуск
        start_cmd = ['docker-compose', '-f', 'docker-compose.siem.yml', 'up', '-d']
        result = subprocess.run(start_cmd, cwd=os.getcwd())
        
        if result.returncode == 0:
            click.echo("✅ SIEM система перезапущена")
        else:
            click.echo("❌ Ошибка перезапуска SIEM")
            sys.exit(1)
            
    except Exception as e:
        click.echo(f"❌ Ошибка: {e}")
        sys.exit(1)

@cli.command()
def status():
    """Статус SIEM системы"""
    click.echo("📊 Статус CloudBridge SIEM...")
    
    try:
        cmd = ['docker-compose', '-f', 'docker-compose.siem.yml', 'ps']
        result = subprocess.run(cmd, cwd=os.getcwd(), capture_output=True, text=True)
        
        if result.returncode == 0:
            click.echo(result.stdout)
        else:
            click.echo("❌ Ошибка получения статуса")
            
    except Exception as e:
        click.echo(f"❌ Ошибка: {e}")

@cli.command()
def logs():
    """Просмотр логов SIEM"""
    click.echo("📋 Логи CloudBridge SIEM...")
    
    try:
        cmd = ['docker-compose', '-f', 'docker-compose.siem.yml', 'logs', '-f']
        subprocess.run(cmd, cwd=os.getcwd())
        
    except KeyboardInterrupt:
        click.echo("\n👋 Выход из просмотра логов")
    except Exception as e:
        click.echo(f"❌ Ошибка: {e}")

@cli.command()
@click.option('--service', help='Конкретный сервис')
def logs_service(service):
    """Просмотр логов конкретного сервиса"""
    if not service:
        click.echo("❌ Укажите сервис: --service <service_name>")
        return
    
    click.echo(f"📋 Логи сервиса {service}...")
    
    try:
        cmd = ['docker-compose', '-f', 'docker-compose.siem.yml', 'logs', '-f', service]
        subprocess.run(cmd, cwd=os.getcwd())
        
    except KeyboardInterrupt:
        click.echo("\n👋 Выход из просмотра логов")
    except Exception as e:
        click.echo(f"❌ Ошибка: {e}")

@cli.command()
def test():
    """Запуск тестов SIEM"""
    click.echo("🧪 Запуск тестов CloudBridge SIEM...")
    
    try:
        result = subprocess.run([sys.executable, 'test_siem.py'], cwd=os.getcwd())
        
        if result.returncode == 0:
            click.echo("✅ Тесты завершены успешно")
        else:
            click.echo("❌ Некоторые тесты не прошли")
            sys.exit(1)
            
    except Exception as e:
        click.echo(f"❌ Ошибка запуска тестов: {e}")
        sys.exit(1)

@cli.command()
def backup():
    """Создание резервной копии данных SIEM"""
    click.echo("💾 Создание резервной копии SIEM...")
    
    try:
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        backup_dir = f"backup_siem_{timestamp}"
        
        # Создание директории
        os.makedirs(backup_dir, exist_ok=True)
        
        # Резервная копия Elasticsearch
        es_backup_cmd = [
            'docker', 'exec', 'siem-elasticsearch',
            'elasticsearch-dump', '--input=http://localhost:9200/siem-*',
            f'--output=/tmp/backup_{timestamp}.json'
        ]
        
        click.echo("📊 Резервная копия Elasticsearch...")
        subprocess.run(es_backup_cmd)
        
        # Копирование файлов конфигурации
        config_files = ['config/siem_config.yaml', 'docker-compose.siem.yml']
        for file in config_files:
            if os.path.exists(file):
                subprocess.run(['cp', file, backup_dir])
        
        click.echo(f"✅ Резервная копия создана: {backup_dir}")
        
    except Exception as e:
        click.echo(f"❌ Ошибка создания резервной копии: {e}")

@cli.command()
@click.option('--backup-dir', required=True, help='Директория с резервной копией')
def restore(backup_dir):
    """Восстановление данных SIEM из резервной копии"""
    if not os.path.exists(backup_dir):
        click.echo(f"❌ Директория {backup_dir} не существует")
        return
    
    click.echo(f"🔄 Восстановление из {backup_dir}...")
    
    try:
        # Восстановление конфигурации
        config_file = os.path.join(backup_dir, 'siem_config.yaml')
        if os.path.exists(config_file):
            subprocess.run(['cp', config_file, 'config/'])
            click.echo("✅ Конфигурация восстановлена")
        
        # Восстановление Elasticsearch
        backup_files = [f for f in os.listdir(backup_dir) if f.endswith('.json')]
        for backup_file in backup_files:
            click.echo(f"📊 Восстановление {backup_file}...")
            restore_cmd = [
                'docker', 'exec', '-i', 'siem-elasticsearch',
                'elasticsearch-dump', '--input=/tmp/restore.json',
                '--output=http://localhost:9200'
            ]
            subprocess.run(restore_cmd, input=open(os.path.join(backup_dir, backup_file)).read())
        
        click.echo("✅ Восстановление завершено")
        
    except Exception as e:
        click.echo(f"❌ Ошибка восстановления: {e}")

@cli.command()
def health():
    """Проверка здоровья SIEM системы"""
    click.echo("🏥 Проверка здоровья CloudBridge SIEM...")
    
    try:
        # Проверка сервисов
        services = ['elasticsearch', 'kibana', 'grafana', 'prometheus']
        healthy_services = 0
        
        for service in services:
            try:
                if service == 'elasticsearch':
                    response = subprocess.run([
                        'curl', '-s', 'http://localhost:9200/_cluster/health'
                    ], capture_output=True, text=True)
                elif service == 'kibana':
                    response = subprocess.run([
                        'curl', '-s', 'http://localhost:5601/api/status'
                    ], capture_output=True, text=True)
                elif service == 'grafana':
                    response = subprocess.run([
                        'curl', '-s', 'http://localhost:3001/api/health'
                    ], capture_output=True, text=True)
                elif service == 'prometheus':
                    response = subprocess.run([
                        'curl', '-s', 'http://localhost:9091/api/v1/status/targets'
                    ], capture_output=True, text=True)
                
                if response.returncode == 0:
                    click.echo(f"✅ {service}: здоров")
                    healthy_services += 1
                else:
                    click.echo(f"❌ {service}: недоступен")
                    
            except Exception as e:
                click.echo(f"❌ {service}: ошибка - {e}")
        
        click.echo(f"\n📊 Результат: {healthy_services}/{len(services)} сервисов здоровы")
        
        if healthy_services == len(services):
            click.echo("🎉 Все сервисы работают нормально!")
        else:
            click.echo("⚠️  Некоторые сервисы требуют внимания")
            
    except Exception as e:
        click.echo(f"❌ Ошибка проверки здоровья: {e}")

@cli.command()
def clean():
    """Очистка данных SIEM"""
    click.echo("🧹 Очистка данных CloudBridge SIEM...")
    
    if click.confirm("⚠️  Это удалит все данные SIEM. Продолжить?"):
        try:
            # Остановка сервисов
            subprocess.run(['docker-compose', '-f', 'docker-compose.siem.yml', 'down'])
            
            # Удаление томов
            volumes = ['siem_elasticsearch_data', 'siem_kafka_data', 'siem_redis_data']
            for volume in volumes:
                subprocess.run(['docker', 'volume', 'rm', volume])
            
            click.echo("✅ Данные очищены")
            
        except Exception as e:
            click.echo(f"❌ Ошибка очистки: {e}")

@cli.command()
def info():
    """Информация о SIEM системе"""
    click.echo("ℹ️  Информация о CloudBridge SIEM")
    click.echo("=" * 40)
    
    info_data = {
        "Версия": "2.0.0",
        "Описание": "Система информационной безопасности и мониторинга событий",
        "Компоненты": [
            "Elasticsearch - хранение и поиск данных",
            "Kibana - визуализация и анализ",
            "Grafana - мониторинг метрик",
            "Prometheus - сбор метрик",
            "Kafka - потоковая обработка",
            "Redis - кэширование",
            "SIEM Collector - сбор данных",
            "SIEM Processor - обработка событий",
            "SIEM Alert Manager - уведомления"
        ],
        "Порты": {
            "Elasticsearch": "9200",
            "Kibana": "5601",
            "Grafana": "3001",
            "Prometheus": "9091",
            "SIEM Web UI": "8081",
            "Kafka": "9092",
            "Redis": "6379"
        },
        "Документация": "README.md",
        "Конфигурация": "config/siem_config.yaml"
    }
    
    for key, value in info_data.items():
        if isinstance(value, list):
            click.echo(f"{key}:")
            for item in value:
                click.echo(f"  - {item}")
        elif isinstance(value, dict):
            click.echo(f"{key}:")
            for k, v in value.items():
                click.echo(f"  - {k}: {v}")
        else:
            click.echo(f"{key}: {value}")

if __name__ == '__main__':
    cli() 