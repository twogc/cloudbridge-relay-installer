#!/bin/bash
# 🧹 CloudBridge Relay - Очистка развертывания

echo "🧹 Очистка CloudBridge Relay развертывания..."
echo ""

# Подтверждение
read -p "Вы уверены, что хотите удалить все ресурсы? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Отменено пользователем"
    exit 1
fi

echo "🗑️ Удаление ресурсов..."

# Удаление deployments
echo "Удаление deployments..."
kubectl delete deployment --all -n cloudbridge 2>/dev/null || true

# Удаление services
echo "Удаление services..."
kubectl delete service --all -n cloudbridge 2>/dev/null || true

# Удаление configmaps
echo "⚙️ Удаление configmaps..."
kubectl delete configmap --all -n cloudbridge 2>/dev/null || true

# Удаление secrets
echo "Удаление secrets..."
kubectl delete secret --all -n cloudbridge 2>/dev/null || true

# Удаление PVC
echo "💾 Удаление PVC..."
kubectl delete pvc --all -n cloudbridge 2>/dev/null || true

# Удаление PV
echo "💽 Удаление PV..."
kubectl delete pv --all 2>/dev/null || true

# Удаление namespaces
echo "📁 Удаление namespaces..."
kubectl delete namespace cloudbridge 2>/dev/null || true
kubectl delete namespace cloudbridge-monitoring 2>/dev/null || true
kubectl delete namespace monitoring 2>/dev/null || true

# Очистка временных файлов
echo "🗂️ Очистка временных файлов..."
rm -f /tmp/cassandra-simple.yaml
rm -f /tmp/cloudbridge-relay-turn-pool.yaml
rm -f /tmp/ai-service-simple.yaml

echo ""
echo "Очистка завершена!"
echo ""
echo "Проверка:"
kubectl get namespaces | grep -E "(cloudbridge|monitoring)" || echo "Все namespaces удалены"
