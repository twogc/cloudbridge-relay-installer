package main

import (
	"fmt"
	"log"

	"go.uber.org/cadence/client"
	"go.uber.org/cadence/client/workflowserviceclient"
	"go.uber.org/cadence/client/workflowserviceclient/thrift"
)

func main() {
	// Создаем Cadence service client
	serviceClient := workflowserviceclient.NewThriftClient(thrift.NewClient("localhost:7933"))
	
	// Создаем домен клиент
	domainClient := client.NewDomainClient(serviceClient)

	// Регистрируем домен cloudbridge
	err := domainClient.Register(nil, &client.RegisterDomainRequest{
		Name:        "cloudbridge",
		Description: "CloudBridge Relay Workflow Domain",
		OwnerEmail:  "admin@cloudbridge.local",
		RetentionDays: 7,
		EmitMetric:  true,
	})
	if err != nil {
		// Проверяем, не существует ли уже домен
		if _, describeErr := domainClient.Describe(nil, "cloudbridge"); describeErr == nil {
			fmt.Println("Domain 'cloudbridge' already exists")
		} else {
			log.Printf("Failed to register domain: %v", err)
		}
	} else {
		fmt.Println("Domain 'cloudbridge' created successfully")
	}

	// Проверяем создание домена
	domain, err := domainClient.Describe(nil, "cloudbridge")
	if err != nil {
		log.Fatalf("Failed to describe domain: %v", err)
	}

	fmt.Printf("Domain info: %+v\n", domain.DomainInfo)
}