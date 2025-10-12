# CloudBridge LoadBalancer & Tenants Report

**Дата:** Sun Sep 28 12:45:11 MSK 2025  
**Кластер:** minikube  

## LoadBalancer Services

NAME                             TYPE           CLUSTER-IP       EXTERNAL-IP      PORT(S)                                                                                                                    AGE     SELECTOR
cloudbridge-relay-lb-tcp         LoadBalancer   10.104.18.159    10.104.18.159    9090:30325/TCP,8084:30651/TCP,8082:31024/TCP                                                                               3d10h   app.kubernetes.io/component=relay,app.kubernetes.io/name=cloudbridge-relay
cloudbridge-relay-lb-udp         LoadBalancer   10.99.61.221     10.99.61.221     9090:30897/UDP,51820:31921/UDP                                                                                             3d10h   app.kubernetes.io/component=relay,app.kubernetes.io/name=cloudbridge-relay
cloudbridge-relay-loadbalancer   LoadBalancer   10.106.92.52     10.106.92.52     5550:31229/TCP,5551:31295/TCP,5552:31167/TCP,5553:32397/TCP,19302:30975/UDP,9090:31168/UDP,8443:32196/TCP,9092:30952/UDP   3d13h   app=cloudbridge-relay
p2p-api-loadbalancer             LoadBalancer   10.106.227.132   10.106.227.132   8082:30612/TCP                                                                                                             3d10h   app.kubernetes.io/component=relay,app.kubernetes.io/name=cloudbridge-relay-p2p

## Tenant Configuration

### Настроенные тенанты:
- default
- dev-team

### Настроенные кластеры:
- production
- staging

## Pod Status

NAME                                 READY   STATUS    RESTARTS   AGE   IP             NODE       NOMINATED NODE   READINESS GATES
cloudbridge-relay-79c9488f7d-5c786   1/1     Running   0          24m   10.244.1.163   minikube   <none>           <none>
cloudbridge-relay-79c9488f7d-9fhk7   1/1     Running   0          25m   10.244.1.161   minikube   <none>           <none>
cloudbridge-relay-79c9488f7d-rnfwq   1/1     Running   0          25m   10.244.1.162   minikube   <none>           <none>

## Service Endpoints

cloudbridge-relay                <none>                                                               3d11h
cloudbridge-relay-external       10.244.1.161:5550,10.244.1.162:5550,10.244.1.163:5550 + 24 more...   3d13h
cloudbridge-relay-headless       <none>                                                               3d10h
cloudbridge-relay-lb-tcp         <none>                                                               3d10h
cloudbridge-relay-lb-udp         <none>                                                               3d10h
cloudbridge-relay-loadbalancer   10.244.1.161:5550,10.244.1.162:5550,10.244.1.163:5550 + 21 more...   3d13h
cloudbridge-relay-nodeport       <none>                                                               3d10h
cloudbridge-relay-simple         <none>                                                               3d12h
cloudbridge-relay-tls                                                                                 15m

## Рекомендации

1. Проверить External IP для всех LoadBalancer сервисов
2. Настроить мониторинг LoadBalancer метрик
3. Создать секреты для Kubernetes кластеров
4. Протестировать tenant-specific API endpoints
5. Настроить health checks для LoadBalancer'ов

