#!/bin/bash
# Демон: отправка алертов Suricata из eve.json в Telegram в реальном времени с логированием
EVE_FILE="/var/log/suricata/eve.json"
LOG_FILE="/tmp/suricata_telegram_debug.log"

log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1" >> "$LOG_FILE"
}

log "--- Демон запущен ---"
tail -F "$EVE_FILE" | \
while read -r line; do
    if echo "$line" | grep -q '"event_type":"alert"'; then
        log "Обнаружен алерт: $line"
        TS=$(echo "$line" | jq -r '.timestamp // empty')
        SIG=$(echo "$line" | jq -r '.alert.signature // empty')
        SRC=$(echo "$line" | jq -r '.src_ip // empty')
        SRCPORT=$(echo "$line" | jq -r '.src_port // empty')
        DST=$(echo "$line" | jq -r '.dest_ip // empty')
        DSTPORT=$(echo "$line" | jq -r '.dest_port // empty')
        PROTO=$(echo "$line" | jq -r '.proto // empty')
        SEV=$(echo "$line" | jq -r '.alert.severity // empty')
        MSG="[Suricata] $TS\n$SIG\n$SRC:$SRCPORT → $DST:$DSTPORT ($PROTO)\nSeverity: $SEV"
        log "Отправка в Telegram: $MSG"
        $(dirname "$0")/telegram_notify.sh "$MSG" HTML 2>>"$LOG_FILE"
        if [ $? -eq 0 ]; then
            log "Успешно отправлено в Telegram"
        else
            log "Ошибка отправки в Telegram"
        fi
    fi
done 