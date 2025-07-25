#!/bin/bash
# Демон: отправка алертов Suricata из eve.json в Telegram в реальном времени
EVE_FILE="/var/log/suricata/eve.json"

# tail -F следит за появлением новых строк в eve.json
 tail -F "$EVE_FILE" | \
while read -r line; do
    if echo "$line" | grep -q '"event_type":"alert"'; then
        # Парсим ключевые поля
        TS=$(echo "$line" | jq -r '.timestamp // empty')
        SRC=$(echo "$line" | jq -r '.src_ip // empty')
        DST=$(echo "$line" | jq -r '.dest_ip // empty')
        PROTO=$(echo "$line" | jq -r '.proto // empty')
        SIG=$(echo "$line" | jq -r '.alert.signature // empty')
        SEV=$(echo "$line" | jq -r '.alert.severity // empty')
        MSG="[Suricata] $TS\n$SIG\n$SRC → $DST ($PROTO)\nSeverity: $SEV"
        $(dirname "$0")/telegram_notify.sh "$MSG"
    fi
done 