#!/bin/bash

# Простой тест P2P маршрутизации
echo "🚀 Тестирование P2P маршрутизации CloudBridge Relay"

# Проверка доступности сервера
echo "Проверка сервера..."
if curl -s -f "http://localhost:8080/health" > /dev/null; then
    echo "✅ HTTP API доступен"
else
    echo "❌ HTTP API недоступен"
    exit 1
fi

if curl -s -f "http://localhost:8082/health" > /dev/null; then
    echo "✅ P2P API доступен"
else
    echo "❌ P2P API недоступен"
    exit 1
fi

# Тест регистрации пира
echo "Тестирование регистрации пира..."
JWT_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6ImZhbGxiYWNrLWtleSJ9.eyJwcm90b2NvbF90eXBlIjoicDJwLW1lc2giLCJzY29wIjoicDJwLW1lc2gtY2xhaW1zIiwib3JnX2lkIjoidGVuYW50LTIxNjQyMDE2NSIsInRlbmFudF9pZCI6InRlbmFudC0yMTY0MjAxNjUiLCJzZXJ2ZXJfaWQiOiJzZXJ2ZXItMTc1ODIxOTE5NTcwMyIsImNvbm5lY3Rpb25fdHlwZSI6InF1aWMtaWNlIiwibWF4X3BlZXJzIjoiMTAiLCJwZXJtaXNzaW9ucyI6WyJtZXNoX2pvaW4iLCJtZXNoX21hbmFnZSJdLCJuZXR3b3JrX2NvbmZpZyI6eyJzdWJuZXQiOiIxMC4wLjAuMC8yNCIsImdhdGV3YXkiOiIxMC4wLjAuMSIsImRucyI6WyI4LjguOC44IiwiMS4xLjEuMSJdLCJtdHUiOjE0MjAsImZpcmV3YWxsX3J1bGVzIjpbImFsbG93X3NzaCIsImFsbG93X2h0dHAiXSwiZW5hYmxlX2lwdjYiOmZhbHNlfSwicXVpY19jb25maWciOnsiaGFuZHNoYWtlX3RpbWVvdXQiOiIxMHMiLCJpZGxlX3RpbWVvdXQiOiIzMHMiLCJtYXhfc3RyZWFtcyI6MTAwLCJjb25uZWN0aW9uX21pZ3JhdGlvbiI6dHJ1ZSwibXVsdGlwbGV4aW5nIjp0cnVlfSwiaWNlX2NvbmZpZyI6eyJzdHVuX3NlcnZlcnMiOlsiZWRnZS4yZ2MucnU6MTkzMDIiXSwidHVybl9zZXJ2ZXJzIjpbXSwiaWNlX3RpbWVvdXQiOiIzMHMiLCJjb25uZWN0aW9uX3RpbWVvdXQiOiIxMHMiLCJnYXRoZXJpbmdfdGltZW91dCI6IjVzIiwiY2FuZGlkYXRlX3R5cGVzIjpbImhvc3QiLCJzcmZseCIsInJlbGF5Il19LCJtYXNxdWVfY29uZmlnIjp7ImVuYWJsZWQiOnRydWUsInByb3h5X3VybCI6Imh0dHBzOi8vZWRnZS4yZ2MucnU6ODQ0MyIsImFscG5fcHJvdG9jb2xzIjpbImgzIiwiaDMtMjkiXSwiZmFsbGJhY2tfZW5hYmxlZCI6dHJ1ZX0sIm1lc2hfY29uZmlnIjp7Im5ldHdvcmtfaWQiOiJtZXNoLW5ldHdvcmstMDAxIiwic3VibmV0IjoiMTAuMC4wLjAvMTYiLCJyZWdpc3RyeV91cmwiOiJodHRwczovL2VkZ2UuMmdjLnJ1OjgwODAiLCJoZWFydGJlYXRfaW50ZXJ2YWwiOiIzMHMiLCJtYXhfcGVlcnMiOjEwLCJyb3V0aW5nX3N0cmF0ZWd5IjoicGVyZm9ybWFuY2Vfb3B0aW1hbCIsImVuYWJsZV9hdXRvX2Rpc2NvdmVyeSI6dHJ1ZSwidHJ1c3RfbGV2ZWwiOiJiYXNpYyJ9LCJwZWVyX3doaXRlbGlzdCI6WyJwZWVyLTAwMSIsInBlZXItMDAyIiwicGVlci0wMDMiXSwicmVsYXlfZW5kcG9pbnRzIjp7Imh0dHBfYXBpIjoiZWRnZS4yZ2MucnU6ODA4MCIsInAycF9hcGkiOiJlZGdlLjJnYy5ydTo4MDgyIiwicXVpY190cmFuc3BvcnQiOiJlZGdlLjJnYy5ydTo5MDkwIiwic3R1bl9zZXJ2ZXIiOiJlZGdlLjJnYy5ydTo4NDQzIiwibWFzcXVlX3Byb3h5IjoiZWRnZS4yZ2MucnU6ODQ0MyIsImVuaGFuY2VkX3F1aWMiOiJlZGdlLjJnYy5ydTo5MDkyIn0sImlhdCI6MTc1ODIxOTE5OCwiaXNzIjoiaHR0cHM6Ly9hdXRoLjJnYy5ydS9yZWFsbXMvY2xvdWRicmlkZ2UiLCJhdWQiOiJhY2NvdW50Iiwic3ViIjoic2VydmVyLWNsaWVudC1zZXJ2ZXItMTc1ODIxOTE5NTcwMyIsImp0aSI6Imp3dF8xNzU4MjE5MTk4MTA5X216MXR4cnNkYSJ9.d-o1yXtZOlnyv53Uhg5zEI3l3IRbQV27VdnEQIn5w_4"

RESPONSE=$(curl -s -X POST "http://localhost:8082/api/v1/tenants/tenant-216420165/peers/register" \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "public_key": "test-peer-public-key",
        "allowed_ips": ["10.100.77.10/32"],
        "peer_info": {
            "name": "Test Peer",
            "location": "Test Location"
        }
    }')

if echo "$RESPONSE" | grep -q "success.*true"; then
    echo "✅ Регистрация пира работает"
else
    echo "❌ Ошибка регистрации пира"
    echo "$RESPONSE"
fi

echo "🎯 P2P маршрутизация готова к тестированию!"