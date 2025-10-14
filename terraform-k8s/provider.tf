terraform {
    required_providers {
        vkcs = {
            source = "vk-cs/vkcs"
            version = "~> 0.8.0"
        }
    }
}

provider "vkcs" {
    # Your user account
    username = "info@2gc.ru"
    
    # The password of the account
    password = "7Thbngy6"
    
    # The tenant token can be taken from the project Settings tab -> API keys
    # Project ID will be our token
    project_id = "8530029fa7b64199b2ad87c18d4620df"
    
    # Region name
    region = "RegionOne"
    
    auth_url = "https://infra.mail.ru:35357/v3/"
}
