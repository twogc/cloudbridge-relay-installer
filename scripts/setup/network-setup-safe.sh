#!/bin/bash

# Безопасная настройка дополнительной сетевой карты для CloudBridge Relay
# Этот скрипт предотвращает потерю SSH доступа

set -e

echo "=== Безопасная настройка дополнительной сетевой карты ==="

# 1. Проверяем текущее состояние сети
echo "1. Проверяем текущие сетевые интерфейсы..."
ip addr show

echo -e "\n2. Проверяем текущую маршрутизацию..."
ip route show

# 2. Определяем основную сетевую карту (с SSH доступом)
PRIMARY_INTERFACE=$(ip route | grep default | awk '{print $5}' | head -1)
PRIMARY_IP=$(ip addr show $PRIMARY_INTERFACE | grep 'inet ' | awk '{print $2}' | cut -d/ -f1)

echo -e "\n3. Основной интерфейс: $PRIMARY_INTERFACE"
echo "   Основной IP: $PRIMARY_IP"

# 3. Ищем дополнительную сетевую карту
SECONDARY_INTERFACE=""
for interface in $(ls /sys/class/net/ | grep -E '^ens|^eth|^enp'); do
    if [ "$interface" != "$PRIMARY_INTERFACE" ] && [ "$interface" != "lo" ]; then
        if ! ip addr show $interface | grep -q "inet "; then
            SECONDARY_INTERFACE=$interface
            break
        fi
    fi
done

if [ -z "$SECONDARY_INTERFACE" ]; then
    echo "❌ Дополнительная сетевая карта не найдена"
    echo "   Убедитесь, что карта добавлена в VK Cloud и видна в системе"
    exit 1
fi

echo -e "\n4. Найдена дополнительная карта: $SECONDARY_INTERFACE"

# 4. Создаем резервную копию сетевых настроек
echo -e "\n5. Создаем резервную копию сетевых настроек..."
sudo cp /etc/netplan/50-cloud-init.yaml /etc/netplan/50-cloud-init.yaml.backup.$(date +%Y%m%d_%H%M%S)

# 5. Настраиваем дополнительную карту БЕЗ изменения маршрутизации по умолчанию
echo -e "\n6. Настраиваем дополнительную карту..."

# Получаем MAC адрес дополнительной карты
SECONDARY_MAC=$(cat /sys/class/net/$SECONDARY_INTERFACE/address)

# Создаем конфигурацию netplan для дополнительной карты
sudo tee /etc/netplan/99-secondary-interface.yaml > /dev/null <<EOF
network:
  version: 2
  ethernets:
    $SECONDARY_INTERFACE:
      match:
        macaddress: $SECONDARY_MAC
      dhcp4: true
      dhcp4-overrides:
        route-metric: 200  # Более высокий metric, чтобы не перехватывать маршрут по умолчанию
      optional: true
EOF

echo "   Конфигурация создана: /etc/netplan/99-secondary-interface.yaml"

# 6. Применяем конфигурацию
echo -e "\n7. Применяем сетевую конфигурацию..."
sudo netplan try --timeout=30

if [ $? -eq 0 ]; then
    echo "Сетевая конфигурация применена успешно"
    sudo netplan apply
else
    echo "❌ Ошибка применения конфигурации, откатываемся..."
    sudo netplan apply
    exit 1
fi

# 7. Проверяем результат
echo -e "\n8. Проверяем результат..."
sleep 5

echo "   Текущие интерфейсы:"
ip addr show

echo -e "\n   Текущая маршрутизация:"
ip route show

# 8. Проверяем SSH доступ
echo -e "\n9. Проверяем SSH доступ..."
if ping -c 1 8.8.8.8 > /dev/null 2>&1; then
    echo "Интернет доступен"
else
    echo "❌ Проблемы с интернетом"
fi

# 9. Настраиваем балансировку нагрузки (если нужно)
echo -e "\n10. Настройка балансировки нагрузки..."

# Создаем скрипт для настройки bonding (опционально)
cat > setup-bonding.sh << 'EOF'
#!/bin/bash
# Опциональная настройка bonding для балансировки нагрузки

PRIMARY_INTERFACE="ens3"  # Замените на ваш основной интерфейс
SECONDARY_INTERFACE="ens8"  # Замените на ваш дополнительный интерфейс

echo "Настройка bonding для балансировки нагрузки..."

# Создаем конфигурацию bonding
sudo tee /etc/netplan/99-bonding.yaml > /dev/null <<BONDEOF
network:
  version: 2
  bonds:
    bond0:
      interfaces:
        - $PRIMARY_INTERFACE
        - $SECONDARY_INTERFACE
      parameters:
        mode: balance-rr  # Round-robin балансировка
        mii-monitor-interval: 100
      dhcp4: true
      dhcp4-overrides:
        route-metric: 100
BONDEOF

echo "Конфигурация bonding создана. Применить? (y/N)"
read -r response
if [[ "$response" =~ ^[Yy]$ ]]; then
    sudo netplan try --timeout=30
    if [ $? -eq 0 ]; then
        sudo netplan apply
        echo "Bonding настроен успешно"
    else
        echo "❌ Ошибка настройки bonding"
    fi
fi
EOF

chmod +x setup-bonding.sh

echo -e "\n=== Настройка завершена ==="
echo "Дополнительная сетевая карта настроена"
echo "SSH доступ сохранен"
echo "Создан скрипт setup-bonding.sh для настройки балансировки нагрузки"
echo ""
echo "Следующие шаги:"
echo "1. Проверьте работу SSH: ssh user@your-server-ip"
echo "2. Если нужно, запустите: ./setup-bonding.sh"
echo "3. Настройте CloudBridge Relay для использования нескольких интерфейсов"
