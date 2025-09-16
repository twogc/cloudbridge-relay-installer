#!/bin/bash

echo "🧪 Тестирование AI API endpoints..."

# Базовые endpoints
echo "✅ Тестирование базовых endpoints..."
curl -s http://localhost:8083/health | jq . || echo "❌ Health check failed"
curl -s http://localhost:8083/api/v1/ai/status | jq . || echo "❌ Status endpoint failed"

# Новые endpoints
echo "✅ Тестирование новых endpoints..."
curl -s http://localhost:8083/api/v1/ai/history/anomalies | jq . || echo "❌ History anomalies failed"
curl -s http://localhost:8083/api/v1/ai/history/predictions | jq . || echo "❌ History predictions failed"
curl -s http://localhost:8083/api/v1/ai/history/routes | jq . || echo "❌ History routes failed"
curl -s http://localhost:8083/api/v1/ai/trends/performance | jq . || echo "❌ Performance trends failed"

# Realtime endpoints
echo "✅ Тестирование realtime endpoints..."
curl -s http://localhost:8083/api/v1/ai/realtime/connections | jq . || echo "❌ Realtime connections failed"
curl -s http://localhost:8083/api/v1/ai/realtime/performance | jq . || echo "❌ Realtime performance failed"

# Alert endpoints
echo "✅ Тестирование alert endpoints..."
curl -s http://localhost:8083/api/v1/ai/alerts | jq . || echo "❌ Alerts failed"

# Health endpoints
echo "✅ Тестирование health endpoints..."
curl -s http://localhost:8083/api/v1/ai/health/detailed | jq . || echo "❌ Detailed health failed"
curl -s http://localhost:8083/api/v1/ai/health/components | jq . || echo "❌ Component health failed"

echo "🎉 Тестирование завершено!"
