package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/zitadel/oidc/v3/pkg/oidc"

	"github.com/zitadel/zitadel-go/v3/pkg/client"
	"github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/app"
	"github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/management"
	"github.com/zitadel/zitadel-go/v3/pkg/zitadel"
)

var (
	domain       = flag.String("domain", "zitadel.local", "your ZITADEL instance domain")
	clientID     = flag.String("clientID", "", "your service user's client ID")
	clientSecret = flag.String("clientSecret", "", "your service user's client secret")
	projectName  = flag.String("project", "cloudbridge-project", "project name to create")
	appName      = flag.String("app", "cloudbridge-relay", "application name to create")
)

func main() {
	flag.Parse()
	ctx := context.Background()

	// Если не указаны credentials, используем админские
	if *clientID == "" || *clientSecret == "" {
		fmt.Println("Using admin credentials for initial setup...")
		*clientID = "admin@local.test"
		*clientSecret = "jrcCaigzmqzHpRcxo61RDiS+xejxX4/K"
	}

	authOption := client.PasswordAuthentication(
		*clientID,
		*clientSecret,
		oidc.ScopeOpenID,
		client.ScopeZitadelAPI(),
	)

	api, err := client.New(ctx, zitadel.New(*domain), client.WithAuth(authOption))
	if err != nil {
		slog.Error("could not create api client", "error", err)
		os.Exit(1)
	}

	// Получаем информацию об организации
	orgResp, err := api.ManagementService().GetMyOrg(ctx, &management.GetMyOrgRequest{})
	if err != nil {
		slog.Error("failed to get organization", "error", err)
		os.Exit(1)
	}

	orgID := orgResp.GetOrg().GetId()
	fmt.Printf("Organization: %s (ID: %s)\n", orgResp.GetOrg().GetName(), orgID)

	// Создаем проект
	fmt.Printf("Creating project: %s\n", *projectName)
	projectResp, err := api.ManagementService().AddProject(ctx, &management.AddProjectRequest{
		Name: *projectName,
	})
	if err != nil {
		slog.Error("failed to create project", "error", err)
		os.Exit(1)
	}

	projectID := projectResp.GetId()
	fmt.Printf("Project created successfully! ID: %s\n", projectID)

	// Создаем приложение
	fmt.Printf("Creating application: %s\n", *appName)
	appResp, err := api.ManagementService().AddOIDCApp(ctx, &management.AddOIDCAppRequest{
		ProjectId:       projectID,
		Name:            *appName,
		AppType:         app.OIDCAppType_OIDC_APP_TYPE_WEB,
		AuthMethodType:  app.OIDCAuthMethodType_OIDC_AUTH_METHOD_TYPE_BASIC,
		AccessTokenType: app.OIDCTokenType_OIDC_TOKEN_TYPE_BEARER,
		GrantTypes: []app.OIDCGrantType{
			app.OIDCGrantType_OIDC_GRANT_TYPE_AUTHORIZATION_CODE,
		},
		ResponseTypes: []app.OIDCResponseType{
			app.OIDCResponseType_OIDC_RESPONSE_TYPE_CODE,
		},
		RedirectUris: []string{
			"http://localhost:8080/callback",
			"https://localhost:8443/callback",
		},
		PostLogoutRedirectUris: []string{
			"http://localhost:8080/",
			"https://localhost:8443/",
		},
	})
	if err != nil {
		slog.Error("failed to create application", "error", err)
		os.Exit(1)
	}

	appID := appResp.GetAppId()
	clientIDValue := appResp.GetClientId()
	clientSecretValue := appResp.GetClientSecret()

	fmt.Printf("Application created successfully! ID: %s\n", appID)
	fmt.Printf("Client ID: %s\n", clientIDValue)
	fmt.Printf("Client Secret: %s\n", clientSecretValue)

	// Выводим конфигурацию для обновления
	fmt.Println("\n=== CONFIGURATION TO UPDATE ===")
	fmt.Printf("ZITADEL_PROJECT_ID=%s\n", projectID)
	fmt.Printf("ZITADEL_CLIENT_ID=%s\n", clientIDValue)
	fmt.Printf("ZITADEL_CLIENT_SECRET=%s\n", clientSecretValue)
	fmt.Printf("ZITADEL_DOMAIN=https://%s\n", *domain)

	fmt.Println("\n=== JSON CONFIG ===")
	fmt.Printf(`{
  "zitadel": {
    "enabled": true,
    "domain": "https://%s",
    "project_id": "%s",
    "client_id": "%s",
    "client_secret": "%s",
    "fallback_secret": "85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE=",
    "token_validation": {
      "audience": "cloudbridge-relay",
      "clock_skew": "30s"
    },
    "cache_ttl": "5m",
    "http_timeout": "30s",
    "tenant_management": {
      "enabled": true,
      "auto_create": false,
      "default_quotas": {
        "max_peers": 10,
        "max_connections": 100,
        "max_tunnels": 20,
        "bandwidth_limit": "1GB",
        "storage_limit": "10GB"
      }
    }
  }
}
`, *domain, projectID, clientIDValue, clientSecretValue)

	fmt.Println("\nZitadel project and application created successfully!")
}
