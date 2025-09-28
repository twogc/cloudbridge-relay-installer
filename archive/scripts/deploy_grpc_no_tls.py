#!/usr/bin/env python3
"""
Скрипт для развертывания gRPC сервера без TLS
"""
import subprocess
import time
import os

def run_command(cmd, description):
    """Выполнить команду и показать результат"""
    print(f"🔍 {description}")
    print(f"Команда: {cmd}")
    try:
        result = subprocess.run(cmd, shell=True, capture_output=True, text=True, cwd="/home/ubuntu/cloudbridge-relay-installer")
        print(f"Код выхода: {result.returncode}")
        if result.stdout:
            print(f"Вывод:\n{result.stdout}")
        if result.stderr:
            print(f"Ошибки:\n{result.stderr}")
        return result.returncode == 0
    except Exception as e:
        print(f"Ошибка выполнения: {e}")
        return False

def main():
    print("🚀 РАЗВЕРТЫВАНИЕ gRPC СЕРВЕРА БЕЗ TLS")
    print("=" * 50)
    
    # Шаг 1: Применяем ConfigMap
    if not run_command("kubectl apply -f k8s/configmap-no-tls.yaml", "Применяем ConfigMap без TLS"):
        print("❌ Ошибка при применении ConfigMap")
        return False
    
    # Шаг 2: Применяем Deployment
    if not run_command("kubectl apply -f k8s/deployment-no-tls.yaml", "Применяем Deployment без TLS"):
        print("❌ Ошибка при применении Deployment")
        return False
    
    # Шаг 3: Применяем Service
    if not run_command("kubectl apply -f k8s/service-no-tls.yaml", "Применяем Service без TLS"):
        print("❌ Ошибка при применении Service")
        return False
    
    # Шаг 4: Ждем запуска пода
    print("⏳ Ждем запуска пода...")
    time.sleep(10)
    
    # Шаг 5: Проверяем статус
    run_command("kubectl get pods -n cloudbridge | grep no-tls", "Проверяем статус подов")
    run_command("kubectl get services -n cloudbridge | grep no-tls", "Проверяем сервисы")
    
    # Шаг 6: Настраиваем port-forward
    print("🔍 Настраиваем port-forward...")
    subprocess.run("pkill -f 'kubectl port-forward.*8444'", shell=True, capture_output=True)
    time.sleep(2)
    
    # Запускаем port-forward в фоне
    subprocess.Popen("kubectl port-forward -n cloudbridge service/cloudbridge-relay-no-tls 8444:8444", 
                    shell=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    time.sleep(3)
    
    # Шаг 7: Проверяем порт
    run_command("netstat -tlnp | grep 8444", "Проверяем доступность порта 8444")
    
    # Шаг 8: Создаем конфигурацию клиента
    client_config = """relay:
  host: "localhost"
  ports:
    grpc: 8444
  timeout: "30s"
  tls:
    enabled: false

auth:
  type: "none"
  skip_validation: true

api:
  base_url: "grpc://localhost:8444"
  timeout: "30s"
  insecure_skip_verify: true

logging:
  level: "debug"
"""
    
    with open("/home/ubuntu/cloudbridge-relay-installer/test/relay-client/config-no-tls.yaml", "w") as f:
        f.write(client_config)
    
    print("✅ Конфигурация клиента создана")
    
    # Шаг 9: Тестируем клиент
    print("🚀 ТЕСТИРУЕМ КЛИЕНТ...")
    os.chdir("/home/ubuntu/cloudbridge-relay-installer/test/relay-client")
    
    client_cmd = ("timeout 15s ./cloudbridge-client tunnel "
                 "--config config-no-tls.yaml "
                 "--tunnel-id test-tunnel "
                 "--transport grpc "
                 "--insecure-skip-tls-verify "
                 "--log-level debug "
                 "--local-port 8080 "
                 "--remote-host 127.0.0.1 "
                 "--remote-port 80")
    
    run_command(client_cmd, "ТЕСТИРУЕМ gRPC ПОДКЛЮЧЕНИЕ")
    
    print("🔍 Проверяем логи сервера...")
    run_command("kubectl logs -n cloudbridge -l app=cloudbridge-relay-no-tls --tail=10", "Логи сервера")
    
    print("✅ РАЗВЕРТЫВАНИЕ ЗАВЕРШЕНО!")

if __name__ == "__main__":
    main()
