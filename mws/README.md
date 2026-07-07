# CloudBridge Relay Private PoP — OVA build for MWS B2B Store

Воспроизводимая сборка OVA из `cloudbridge-relay-installer` для маркетплейса **MWS B2B Store** (вендор `gc2`, среда **stage**).

## Минимальный стек в образе

| Компонент | Образ / сервис |
|-----------|----------------|
| Docker CE | Ubuntu 22.04 + compose plugin |
| PostgreSQL 15 | `postgres:15-alpine` — БД для Zitadel |
| Zitadel | `ghcr.io/zitadel/zitadel:v2.64.1` — OIDC / control plane |
| CloudBridge Relay | `twogc/cloudbridge-relay:latest` — Private PoP |

Секреты **не** запекаются в OVA: при первом запуске генерируются в `/opt/cloudbridge/.env` (см. `templates/env.example`). Параметры маркетплейса подставляются на этапе Terraform bootstrap (Step 2+).

## Минимальные ресурсы VM

| Ресурс | Значение |
|--------|----------|
| vCPU | **4** |
| RAM | **8 GB** |
| Disk | **80 GB** (thin provisioned) |
| OS | Ubuntu 22.04 LTS Server amd64 |
| NIC | vmxnet3 (VMware) |

Эти значения совпадают с дефолтами в `packer/cloudbridge-relay.pkr.hcl` и документируются в бандле `mts-2gc/bundle/cloudbridge-relay`.

## Структура `mws/`

```
mws/
├── README.md
├── cloud-init/
│   └── cloudbridge-relay-minimal.yaml   # user-data: Docker + stack layout
├── templates/
│   ├── docker-compose.minimal.yml       # Postgres + Zitadel + relay
│   └── env.example                      # шаблон секретов (не коммитить .env)
├── packer/
│   ├── cloudbridge-relay.pkr.hcl        # скелет Packer (vSphere)
│   ├── variables.pkrvars.hcl.example
│   └── http/                            # autoinstall для ISO boot
└── scripts/
    └── build-ova.sh                     # packer | manual | upload
```

## Сборка OVA

### Вариант A — Packer + VMware vSphere

1. Установите [Packer](https://www.packer.io/) ≥ 1.9 и плагин `vsphere`.
2. Скопируйте переменные (без секретов в git):

   ```bash
   cp mws/packer/variables.pkrvars.hcl.example mws/packer/variables.pkrvars.hcl
   # отредактируйте vCenter, ISO checksum, SSH password
   ```

3. Соберите VM:

   ```bash
   chmod +x mws/scripts/build-ova.sh
   ./mws/scripts/build-ova.sh packer
   ```

4. Экспортируйте VM в OVA (vCenter UI или `ovftool`):

   ```bash
   mkdir -p dist
   ovftool vi://USER:PASS@vcenter/Datacenter/vm/cloudbridge-relay-dev dist/cloudbridge-relay-dev.ova
   ```

### Вариант B — ручной экспорт из VMware

```bash
./mws/scripts/build-ova.sh manual
```

Скрипт создаёт `dist/ova-staging/` с `docker-compose.yml`, `env.example`, `cloud-init.yaml` и пошаговой инструкцией `MANUAL_STEPS.md`.

На чистой Ubuntu 22.04 VM (4 vCPU / 8 GB / 80 GB):

1. Примените `cloud-init/cloudbridge-relay-minimal.yaml` как user-data.
2. Скопируйте файлы из `dist/ova-staging/opt/cloudbridge/` в `/opt/cloudbridge/`.
3. Проверьте: `sudo systemctl start cloudbridge-stack && cloudbridge-status`.
4. Экспортируйте выключенную VM в `.ova`.

## Загрузка в Artifactory (stage)

Утилита `artifactory_linux_amd64` выдаётся MWS отдельно. Учётные данные вендора — в `mts-2gc/.env` (`Client_Id=gc2`, `Client_Secret=...`). **Не коммитьте `.env`.**

```bash
# из cloudbridge-relay-installer
export B2BSTORE_ENV=/home/lanies/git/2GC/mts-2gc/.env
export ARTIFACTORY_CLI=/path/to/artifactory_linux_amd64

./mws/scripts/build-ova.sh upload dist/cloudbridge-relay-dev.ova
```

Или вручную:

```bash
./artifactory_linux_amd64 \
  -client-id gc2 \
  -client-secret "$Client_Secret" \
  -endpoint https://stage.b2bstore.mws.ru/a \
  -upload-ova ./dist/cloudbridge-relay-dev.ova
```

После загрузки обновите `hash` в `mts-2gc/bundle/cloudbridge-relay/artifacts.yaml` (refID из вывода CLI).

Проверка через API:

```bash
cd /home/lanies/git/2GC/mts-2gc
python3 scripts/b2bstore.py request GET /vendors/gc2/artifacts
```

## Связь с бандлом B2B Store

| Артефакт | Путь |
|----------|------|
| OVA build | `cloudbridge-relay-installer/mws/` |
| Bundle | `mts-2gc/bundle/cloudbridge-relay/` |
| Инструкция бандла | `mts-2gc/bundle/cloudbridge-relay/BUILD.md` |
| API / доступ | `mts-2gc/docs/ACCESS.md` |

Terraform-манифесты маркетплейса используют backend `vcd` и **не** объявляют `provider` blocks — см. Step 2.

## Troubleshooting

| Симптом | Действие |
|---------|----------|
| `relay: down` | `docker compose -f /opt/cloudbridge/docker-compose.yml logs cloudbridge-relay` |
| Zitadel не стартует | Проверьте `ZITADEL_MASTERKEY` (≥32 символов) и `POSTGRES_PASSWORD` в `.env` |
| Нет образа relay | `docker pull twogc/cloudbridge-relay:latest` или укажите `RELAY_IMAGE` в `.env` |
| Upload 401 | Обновите токен: `python3 scripts/b2bstore.py token` (TTL ~5 мин) |

## Безопасность

- Не коммитьте `variables.pkrvars.hcl`, `.env`, пароли vCenter.
- OVA для stage: смените дефолтные пароли Zitadel после первого входа.
- Production: подставляйте TLS для Zitadel и relay через параметры развёртывания.
