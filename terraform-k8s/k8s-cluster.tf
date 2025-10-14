# Источники данных для получения информации о доступных ресурсах

# Получение списка доступных шаблонов виртуальных машин
data "vkcs_compute_flavor" "k8s-master-flavor" {
    name = "STD3-4-8"  # 4 CPU, 8GB RAM - минимальные требования для master-узла
}

# Получение списка доступных шаблонов кластеров Kubernetes
data "vkcs_kubernetes_clustertemplates" "k8s-template-list" {}

# Получение конкретного шаблона кластера
data "vkcs_kubernetes_clustertemplate" "k8s-template" {
    version = "1.28"  # Самая новая доступная версия в VK Cloud
}

# Получение информации о сети (используем cloudbridge-network)
data "vkcs_networking_network" "k8s-network" {
    name = "cloudbridge-network"
}

# Получение информации о подсети
data "vkcs_networking_subnet" "k8s-subnet" {
    name = "cloudbridge-subnet"
    network_id = data.vkcs_networking_network.k8s-network.id
}

# Создание кластера Kubernetes
resource "vkcs_kubernetes_cluster" "k8s-cluster" {
    name                     = "k8s-cluster"
    cluster_type             = "standard"  # или "regional" для регионального кластера
    cluster_template_id      = data.vkcs_kubernetes_clustertemplate.k8s-template.id
    master_flavor            = data.vkcs_compute_flavor.k8s-master-flavor.id
    master_count             = 1  # Для стандартного кластера: 1, 3 или 5
    # cluster_node_volume_type = "ceph-ssd"  # или "high-iops" - параметр не поддерживается
    network_id               = data.vkcs_networking_network.k8s-network.id
    subnet_id                = data.vkcs_networking_subnet.k8s-subnet.id
    availability_zone        = "ME1"  # ME1, MS1 или GZ1 для региона Москва
    floating_ip_enabled      = true
}

# Создание node group для worker nodes
resource "vkcs_kubernetes_node_group" "worker-nodes" {
    cluster_id    = vkcs_kubernetes_cluster.k8s-cluster.id
    name          = "worker-nodes"
    flavor_id     = data.vkcs_compute_flavor.k8s-master-flavor.id
    node_count    = 2
    availability_zones = ["ME1"]
}

# Вывод информации о созданном кластере
output "cluster_id" {
    value = vkcs_kubernetes_cluster.k8s-cluster.id
}

output "cluster_name" {
    value = vkcs_kubernetes_cluster.k8s-cluster.name
}

output "cluster_status" {
    value = vkcs_kubernetes_cluster.k8s-cluster.status
}

output "k8s_config" {
    value = vkcs_kubernetes_cluster.k8s-cluster.k8s_config
    sensitive = true
}

# Вывод списка доступных версий Kubernetes
output "available_k8s_versions" {
    value = data.vkcs_kubernetes_clustertemplates.k8s-template-list.cluster_templates.*.version
}
