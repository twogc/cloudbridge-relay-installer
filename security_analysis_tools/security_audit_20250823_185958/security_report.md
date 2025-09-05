# 🔒 Отчет по безопасности токена

## Общая информация
- **Домен**: example.com
- **Дата проверки**: Sat Aug 23 19:00:44 MSK 2025
- **Версия скрипта**: 1.0

## Резюме
# Анализ структуры токена
Токен: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c
Длина: 155 символов
Количество частей: 3

## Разбор частей:
Часть 1: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9 (длина: 36, энтропия: 36)
Часть 2: eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ (длина: 74, энтропия: 74)
Часть 3: SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c (длина: 43, энтропия: 43)
Неизвестный формат подписи (43 символов)

## SSL/TLS безопасность
# SSL/TLS анализ для example.com
Дата проверки: Sat Aug 23 18:59:58 MSK 2025

Срок действия сертификата: Jan 15 23:59:59 2026 GMT
Дней до истечения: 145

## Поддерживаемые протоколы:
❌ ssl2: не поддерживается
❌ ssl3: не поддерживается
❌ tls1: не поддерживается
❌ tls1_1: не поддерживается
❌ tls1_2: не поддерживается
❌ tls1_3: не поддерживается

## Заголовки безопасности
# Анализ заголовков безопасности для example.com
Дата проверки: Sat Aug 23 19:00:01 MSK 2025

## Полученные заголовки:
HTTP/2 200 
content-type: text/html
etag: "84238dfc8092e5d9c0dac8ef93371a07:1736799080.121134"
last-modified: Mon, 13 Jan 2025 20:11:20 GMT
cache-control: max-age=1655
date: Sat, 23 Aug 2025 16:00:01 GMT
alt-svc: h3=":443"; ma=93600,h3-29=":443"; ma=93600


## Анализ заголовков безопасности:
❌ Strict-Transport-Security: отсутствует
❌ Content-Security-Policy: отсутствует
❌ X-Frame-Options: отсутствует
❌ X-Content-Type-Options: отсутствует
❌ X-XSS-Protection: отсутствует
❌ Referrer-Policy: отсутствует
❌ Permissions-Policy: отсутствует

## DNS безопасность
# DNS анализ для example.com
Дата проверки: Sat Aug 23 19:00:02 MSK 2025

A запись: 23.220.75.232
Обратный DNS: a23-220-75-232.deploy.static.akamaitechnologies.com.
MX записи:
0 .
TXT записи:
"_k2n1y4vw3qtb4skdx9e7dxt97qrmmq9"
"v=spf1 -all"
DMARC запись: "v=DMARC1;p=reject;sp=reject;adkim=s;aspf=s"

## Открытые порты
# Анализ открытых портов для example.com
Дата проверки: Sat Aug 23 19:00:02 MSK 2025

IP адрес: 23.192.228.80

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
# Анализ уязвимостей для example.com
Дата проверки: Sat Aug 23 19:00:39 MSK 2025

## Проверка Heartbleed:
✅ Heartbeat extension не обнаружен

## Проверка BEAST:
✅ RC4 не поддерживается

## Проверка POODLE:
✅ SSL 3.0 не поддерживается

## Утечка информации
# Анализ утечки информации в токене
Дата проверки: Sat Aug 23 19:00:44 MSK 2025

## Анализ частей токена:
User ID: eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ
App ID: SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c
Session ID: SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c
Подпись: SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c
Уникальных символов в подписи: 29

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
- Полный отчет: ./security_audit_20250823_185958/security_report.md
- Анализ токена: ./security_audit_20250823_185958/token_analysis.txt
- SSL анализ: ./security_audit_20250823_185958/ssl_analysis.txt
- Заголовки безопасности: ./security_audit_20250823_185958/security_headers.txt
- DNS анализ: ./security_audit_20250823_185958/dns_analysis.txt
- Анализ портов: ./security_audit_20250823_185958/ports_analysis.txt
- Анализ уязвимостей: ./security_audit_20250823_185958/vulnerabilities.txt
- Анализ утечки информации: ./security_audit_20250823_185958/information_leak_analysis.txt

