# Инструкция по обновлению nginx для P2P API

## 🔧 Проблема
Текущая конфигурация nginx проксирует на неправильные порты Minikube LoadBalancer.

## ✅ Решение
Обновить конфигурацию nginx для использования правильных портов.

## 📋 Шаги выполнения

### 1. Создать резервную копию текущей конфигурации
```bash
sudo cp /etc/nginx/sites-enabled/cloudbridge-external-proxy.conf /etc/nginx/sites-enabled/cloudbridge-external-proxy.conf.backup
```

### 2. Обновить конфигурацию P2P API
```bash
sudo sed -i 's/192.168.49.2:31466/192.168.49.2:30502/g' /etc/nginx/sites-enabled/cloudbridge-external-proxy.conf
```

### 3. Проверить конфигурацию nginx
```bash
sudo nginx -t
```

### 4. Перезагрузить nginx
```bash
sudo systemctl reload nginx
```

### 5. Проверить работу
```bash
curl http://edge.2gc.ru:8082/health
```

## 🎯 Ожидаемый результат
P2P API должен отвечать на `http://edge.2gc.ru:8082/health` с JSON ответом.
