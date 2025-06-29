# CloudBridge Relay — Срочные меры по безопасности

## 1. Установка и настройка антивируса ClamAV

```bash
sudo apt-get update
sudo apt-get install clamav clamav-daemon -y
sudo systemctl enable --now clamav-freshclam
```
- После установки дождитесь, пока freshclam скачает вирусные базы (2-5 минут).
- Проверьте статус:
  ```bash
  sudo systemctl status clamav-freshclam --no-pager
  sudo systemctl status clamav-daemon --no-pager
  ```
- Если daemon не запускается, подождите, пока появятся файлы `/var/lib/clamav/daily.cvd` и повторите запуск:
  ```bash
  sudo systemctl start clamav-daemon
  sudo systemctl status clamav-daemon --no-pager
  ```

---

## 2. Проверка и включение AppArmor

```bash
sudo aa-status
```
- Если AppArmor не активен:
  ```bash
  sudo systemctl enable --now apparmor
  ```

---

## 3. Защита SSH

1. Откройте конфиг:
   ```bash
   sudo nano /etc/ssh/sshd_config
   ```
2. Убедитесь, что есть строки:
   ```
   PermitRootLogin no
   PasswordAuthentication no
   ```
   - Если используете только ключи для входа — оставьте `PasswordAuthentication no`.
   - Если пока используете пароли — не меняйте этот параметр, пока не настроите ключи.
3. Перезапустите SSH:
   ```bash
   sudo systemctl restart ssh
   ```

---

## 4. Резервное копирование важных данных

Создайте архив с важными директориями:
```bash
tar czvf ~/cloudbridge-backup-$(date +%F).tar.gz certs/ config/ scripts/
```
Проверьте, что архив создан:
```bash
ls -lh ~/cloudbridge-backup-*.tar.gz
```
**Рекомендация:** Скопируйте архив на внешний носитель или в облако.

---

## 5. Проверка пользователей сервисов

Проверьте, от кого запущен relay и другие сервисы:
```bash
ps aux | grep relay
```
- Сервис не должен работать от root.
- Если работает от root — создайте отдельного пользователя и настройте запуск от него.

---

## 6. Проверка статуса firewall и Fail2Ban

```bash
sudo ufw status
sudo fail2ban-client status
```

---

## 7. Рекомендации
- Регулярно обновляйте систему:
  ```bash
  sudo apt-get update && sudo apt-get upgrade
  ```
- Настройте ротацию логов (обычно logrotate уже настроен в Ubuntu).
- Настройте резервное копирование на регулярной основе.
- Проводите периодические аудиты (lynis, security_check.sh).
- Используйте monitoring/alerting для логов и сервисов.

---

**Эти шаги обеспечивают базовую и срочную защиту сервера CloudBridge Relay.** 