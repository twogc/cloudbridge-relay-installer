# 🔒 Отчет по безопасности токена

## Общая информация
- **Домен**: app60526.pushgrow.io
- **Дата проверки**: Sat Aug 23 18:59:44 MSK 2025
- **Версия скрипта**: 1.0

## Резюме
# Анализ структуры токена
Токен: user.2045597475330327839.60526-0d7eb462c42a221300af346a7a.1efa83c73e0ac992c0cf22b196cf6541bd268d402051e1e2
Длина: 106 символов
Количество частей: 4

## Разбор частей:
Часть 1: user (длина: 4, энтропия: 4)
Часть 2: 2045597475330327839 (длина: 19, энтропия: 19)
Часть 3: 60526-0d7eb462c42a221300af346a7a (длина: 32, энтропия: 32)
Часть 4: 1efa83c73e0ac992c0cf22b196cf6541bd268d402051e1e2 (длина: 48, энтропия: 48)
Неизвестный формат подписи (48 символов)

## SSL/TLS безопасность
# SSL/TLS анализ для app60526.pushgrow.io
Дата проверки: Sat Aug 23 18:59:02 MSK 2025

Срок действия сертификата: Sep  5 11:01:23 2025 GMT
Дней до истечения: 12

## Поддерживаемые протоколы:
❌ ssl2: не поддерживается
❌ ssl3: не поддерживается
❌ tls1: не поддерживается
❌ tls1_1: не поддерживается
❌ tls1_2: не поддерживается
❌ tls1_3: не поддерживается

## Заголовки безопасности
# Анализ заголовков безопасности для app60526.pushgrow.io
Дата проверки: Sat Aug 23 18:59:02 MSK 2025

## Полученные заголовки:
HTTP/2 200 
server: nginx
date: Sat, 23 Aug 2025 15:57:11 GMT
content-type: text/html
content-length: 1262
last-modified: Wed, 06 Mar 2019 06:33:42 GMT
etag: "5c7f69c6-4ee"
strict-transport-security: max-age=31536000; includeSubDomains
content-security-policy-report-only: default-src https:; script-src https: 'unsafe-eval' 'unsafe-inline'; style-src https: 'unsafe-inline'; img-src https: data:; font-src https: data:; report-uri /csp-report
cache-control: no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0
x-frame-options: SAMEORIGIN
x-xss-protection: 1; mode=block
accept-ranges: bytes


## Анализ заголовков безопасности:
✅ strict-transport-security: max-age=31536000; includeSubDomains
✅ content-security-policy-report-only: default-src https:; script-src https: 'unsafe-eval' 'unsafe-inline'; style-src https: 'unsafe-inline'; img-src https: data:; font-src https: data:; report-uri /csp-report
✅ x-frame-options: SAMEORIGIN
❌ X-Content-Type-Options: отсутствует
✅ x-xss-protection: 1; mode=block
❌ Referrer-Policy: отсутствует
❌ Permissions-Policy: отсутствует

## DNS безопасность
# DNS анализ для app60526.pushgrow.io
Дата проверки: Sat Aug 23 18:59:03 MSK 2025

A запись: pushgrow.io.
MX записи:
pushgrow.io.
TXT записи:
pushgrow.io.
DMARC запись: pushgrow.io.

## Открытые порты
# Анализ открытых портов для app60526.pushgrow.io
Дата проверки: Sat Aug 23 18:59:03 MSK 2025

IP адрес: pushgrow.io.

## Сканирование портов:
❌ Порт 21: закрыт
❌ Порт 22: закрыт
❌ Порт 23: закрыт
❌ Порт 25: закрыт
❌ Порт 53: закрыт
✅ Порт 80: открыт
❌ Порт 110: закрыт
❌ Порт 143: закрыт
✅ Порт 443: открыт
❌ Порт 993: закрыт
❌ Порт 995: закрыт
❌ Порт 1433: закрыт
❌ Порт 1521: закрыт
❌ Порт 3306: закрыт
❌ Порт 3389: закрыт
❌ Порт 5432: закрыт
❌ Порт 5900: закрыт
❌ Порт 6379: закрыт
❌ Порт 8080: закрыт
❌ Порт 8443: закрыт

## Уязвимости
# Анализ уязвимостей для app60526.pushgrow.io
Дата проверки: Sat Aug 23 18:59:39 MSK 2025

## Проверка Heartbleed:
✅ Heartbeat extension не обнаружен

## Проверка BEAST:
✅ RC4 не поддерживается

## Проверка POODLE:
✅ SSL 3.0 не поддерживается

## Утечка информации
# Анализ утечки информации в токене
Дата проверки: Sat Aug 23 18:59:44 MSK 2025

## Анализ частей токена:
User ID: 2045597475330327839
⚠️  User ID содержит только цифры (может быть последовательным)
App ID: 60526
Session ID: 0d7eb462c42a221300af346a7a
⚠️  App ID содержит только цифры
Подпись: 1efa83c73e0ac992c0cf22b196cf6541bd268d402051e1e2
Уникальных символов в подписи: 16

## Рекомендации

### Критические
1. Убедитесь, что все заголовки безопасности настроены
2. Отключите устаревшие протоколы (SSL 2.0, SSL 3.0)
3. Используйте только современные шифры

### Важные
1. Настройте SPF, DKIM и DMARC записи
2. Регулярно обновляйте SSL сертификаты
3. Мониторьте открытые порты

### Рекомендуемые
1. Используйте HSTS заголовок
2. Настройте Content Security Policy
3. Регулярно проводите аудит безопасности

## Файлы отчета
- Полный отчет: ./security_audit_20250823_185902/security_report.md
- Анализ токена: ./security_audit_20250823_185902/token_analysis.txt
- SSL анализ: ./security_audit_20250823_185902/ssl_analysis.txt
- Заголовки безопасности: ./security_audit_20250823_185902/security_headers.txt
- DNS анализ: ./security_audit_20250823_185902/dns_analysis.txt
- Анализ портов: ./security_audit_20250823_185902/ports_analysis.txt
- Анализ уязвимостей: ./security_audit_20250823_185902/vulnerabilities.txt
- Анализ утечки информации: ./security_audit_20250823_185902/information_leak_analysis.txt

