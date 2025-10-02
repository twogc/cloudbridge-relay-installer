# Создание сети с Neutron SDN для VPN

## 🔍 Проблема
Все созданные сети и роутеры в проекте используют **Sprut SDN**, а VPN сервис поддерживается только для **Neutron SDN**.

## 🌐 Решение: Создание сети через веб-интерфейс

### Шаг 1: Вход в веб-консоль
1. Откройте: **https://msk.cloud.vk.com/app/en/**
2. Войдите с учетными данными:
   - **Username**: `info@2gc.ru`
   - **Password**: `7Thbngy6`
3. Выберите проект: `8530029fa7b64199b2ad87c18d4620df`

### Шаг 2: Создание новой сети
1. Перейдите в **Virtual networks** → **Networks**
2. Нажмите **Create network**
3. **ВАЖНО**: При создании сети выберите **SDN: Neutron** (не Sprut!)
4. Заполните параметры:
   - **Network name**: `vpn-neutron-network`
   - **SDN**: `Neutron` (обязательно!)
   - **Internet Access**: Включить
   - **Router**: `Create new`

### Шаг 3: Настройка подсети
1. В разделе **Subnet List** настройте:
   - **Name**: `vpn-neutron-subnet`
   - **Address**: `10.20.0.0/24`
   - **Gateway**: `10.20.0.1`
   - **Enable DHCP**: Включить
   - **DHCP IP address pool**: `10.20.0.2 - 10.20.0.254`

### Шаг 4: Создание сети
1. Нажмите **Create network**
2. Дождитесь создания сети и роутера

## 🔧 Проверка создания через CLI

После создания сети в веб-интерфейсе, проверьте через CLI:

```bash
# Загрузите переменные окружения
source vkcloud-openstack-env.sh

# Проверьте созданную сеть
openstack network list

# Проверьте SDN тип
openstack network show vpn-neutron-network --format json | jq '.sdn'

# Проверьте роутер
openstack router list
```

## 🚀 Создание VPN сервиса

Если сеть создана с Neutron SDN, попробуйте создать VPN сервис:

```bash
# Найдите ID роутера для новой сети
ROUTER_ID=$(openstack router list --format value -c ID --name vpn-router)

# Создайте VPN сервис
openstack vpn service create vkcloud-vpn-service --router $ROUTER_ID --enable
```

## 📋 Создание VPN туннеля

Если VPN сервис создан успешно, создайте туннель:

### 1. Создание групп конечных точек
```bash
# Локальная группа (подсеть VK Cloud)
openstack vpn endpoint group create vkcloud-local-endpoint \
  --type subnet \
  --value <SUBNET_ID_VPN_NEUTRON_NETWORK>

# Удаленная группа (клиентская подсеть)
openstack vpn endpoint group create client-remote-endpoint \
  --type cidr \
  --value "172.16.0.0/29"
```

### 2. Создание VPN туннеля
```bash
# Получите ID компонентов
IKE_POLICY_ID=$(openstack vpn ike policy show vkcloud-client-ike -f value -c id)
IPSEC_POLICY_ID=$(openstack vpn ipsec policy show vkcloud-client-ipsec -f value -c id)
VPN_SERVICE_ID=$(openstack vpn service show vkcloud-vpn-service -f value -c id)
LOCAL_ENDPOINT_ID=$(openstack vpn endpoint group show vkcloud-local-endpoint -f value -c id)
REMOTE_ENDPOINT_ID=$(openstack vpn endpoint group show client-remote-endpoint -f value -c id)

# Создайте туннель
openstack vpn ipsec site connection create vkcloud-client-vpn \
  --dpd action=hold,interval=30,timeout=120 \
  --initiator bi-directional \
  --peer-address "192.0.2.200" \
  --peer-id "192.0.2.200" \
  --ikepolicy "$IKE_POLICY_ID" \
  --ipsecpolicy "$IPSEC_POLICY_ID" \
  --vpnservice "$VPN_SERVICE_ID" \
  --local-endpoint-group "$LOCAL_ENDPOINT_ID" \
  --peer-endpoint-group "$REMOTE_ENDPOINT_ID" \
  --psk "VkCloud2024!SecureKey#123" \
  --enable
```

## 🔑 Обновленные параметры для клиента

Если VPN туннель создан успешно, обновите конфигурацию клиента:

### Новые параметры:
- **VK Cloud Router IP**: `<NEW_ROUTER_IP>` (получите из веб-интерфейса)
- **VK Cloud Subnet**: `10.20.0.0/24` (новая подсеть)
- **Client IP**: `192.0.2.200`
- **Client Subnet**: `172.16.0.0/29`
- **PSK**: `VkCloud2024!SecureKey#123`

### Обновленная конфигурация `/etc/ipsec.conf`:
```ini
conn client-vkcloud-vpn
   authby=secret
   left=%defaultroute
   leftid=192.0.2.200
   leftsubnet=172.16.0.0/29
   right=<NEW_ROUTER_IP>
   rightsubnet=10.20.0.0/24
   ike=aes256-sha2_256-modp2048!
   esp=aes256-sha2_256!
   keyingtries=0
   ikelifetime=3600
   lifetime=8h
   dpddelay=30
   dpdtimeout=120
   dpdaction=hold
   auto=start
```

### Обновленная конфигурация `/etc/ipsec.secrets`:
```ini
192.0.2.200 <NEW_ROUTER_IP> : PSK "VkCloud2024!SecureKey#123"
```

## ✅ Проверка

После создания VPN туннеля:

1. **Проверьте статус туннеля**:
```bash
openstack vpn ipsec site connection list
openstack vpn ipsec site connection show vkcloud-client-vpn
```

2. **Настройте клиентскую сторону** с новыми параметрами

3. **Протестируйте подключение**:
```bash
ping <NEW_ROUTER_IP>
ping 10.20.0.1
```

## 🚨 Если не получается создать сеть с Neutron SDN

Если в веб-интерфейсе нет опции выбора Neutron SDN, значит в вашем проекте VPN поддерживается только через веб-интерфейс с существующими сетями Sprut SDN.

В этом случае используйте инструкцию `VPN_WEB_ONLY_SETUP.md` для создания VPN туннеля через веб-интерфейс.






