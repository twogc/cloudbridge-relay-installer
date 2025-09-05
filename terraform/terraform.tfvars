# CloudBridge Relay - Terraform Variables Example
# Скопируйте этот файл в terraform.tfvars и настройте под свои нужды

# Название проекта
project_name = "cloudbridge-relay"

# Конфигурация виртуальной машины
# Доступные варианты:
# - STD2-2-4: 2 vCPU, 4 GB RAM (для тестирования)
# - STD2-4-8: 4 vCPU, 8 GB RAM (рекомендуется)
# - STD2-8-16: 8 vCPU, 16 GB RAM (для высокой нагрузки)
compute_flavor = "STD2-4-8"

# Имя SSH ключа (создайте в VK Cloud консоли)
key_pair_name = "cloudbridge-key"

# Зона доступности
# Доступные зоны:
# - MS1: Москва
# - SPB1: Санкт-Петербург
availability_zone = "MS1"

# Размер диска в GB
disk_size = 80

# Публичный SSH ключ для доступа к VM
# Сгенерируйте командой: ssh-keygen -t rsa -b 4096 -C "your-email@example.com"
ssh_public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCPo+PYSDf3VmdtIm+lYRSH6+FaGTO1eCtHcyo6JXeb5F4MEVsKCrnIVytFe/h8Fvo25Xd5bgbmUHHvaCtQYDTpV6QovVu3itg6VEMYDse1wl31AaNi4NDQryaR5wa5MD7HM3EQUVWl6pXvUvHUDlHVCih7k8cAnmybLeXA9+BrZxgoohzOts/IjImutP0SI0VLFg2ZZZ9giLcWrRC8SaiUITcsAVtjbtgDOVHbvMGBmiJ4v96ISXZTt3RJGus2QvR+nGOAH9d1nPc4HFEHPZTDENqgmcQmIPJih7gROkjofgZb0mSkN9EUzdRHIiFq1bVPFVHrZ7fbF/FqqmOYz4AExIXObwyQeadJ75ZYr5tzk+NPCo8oYSNvb03bb5ldwjqn9wzzUiQ5nFnZ2fcnjqogLBuaVPyGjnCNnBYxdad8SLkYeGMCscZ2EecxxCbz2V5Dn0iwTOUjwzxVCsUBNU0giNub7V+JawZF1DE77R3tJsC+bxH4X/z9HG2WapIpS0hqguldX905+CQJbzj1TfXRM/F5Wqoh+ft1pfa39Nh+s8HC/nHh1OxYfpqN7Kgg8B0D1OC4//xKUEdyIrCIsaVuGbtaXScfeg1meMv10M+y15ckl5jwpF9fXtusHqflRie0+BdPugfARpRRMIRilgSU9L+YHaYX7n5BTt6NoFWnXQ=="



