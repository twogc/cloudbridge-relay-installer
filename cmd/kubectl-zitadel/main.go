package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientauthv1beta1 "k8s.io/client-go/pkg/apis/clientauthentication/v1beta1"
)

const (
	pluginName    = "kubectl-zitadel"
	pluginVersion = "v1.0.0"
)

// ZitadelConfig конфигурация для подключения к ZITADEL
type ZitadelConfig struct {
	Domain       string   `json:"domain"`
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	Scopes       []string `json:"scopes"`

	// TLS Configuration
	InsecureSkipVerify bool `json:"insecure_skip_verify"`

	// mTLS Configuration
	MTLSEnabled    bool   `json:"mtls_enabled"`
	CACertPath     string `json:"ca_cert_path"`
	ClientCertPath string `json:"client_cert_path"`
	ClientKeyPath  string `json:"client_key_path"`

	// Token Exchange
	TokenExchange bool   `json:"token_exchange"`
	Audience      string `json:"audience"`
}

func main() {
	var configPath string
	var debug bool

	rootCmd := &cobra.Command{
		Use:     pluginName,
		Short:   "kubectl exec credential plugin for ZITADEL authentication",
		Long:    `A kubectl exec credential plugin that uses ZITADEL for Kubernetes authentication with optional mTLS support`,
		Version: pluginVersion,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCredentialPlugin(configPath, debug)
		},
	}

	rootCmd.Flags().StringVar(&configPath, "config", "", "Path to ZITADEL configuration file")
	rootCmd.Flags().BoolVar(&debug, "debug", false, "Enable debug logging")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runCredentialPlugin(configPath string, debug bool) error {
	// Загрузка конфигурации
	config, err := loadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if debug {
		fmt.Fprintf(os.Stderr, "Debug: loaded config for domain %s\n", config.Domain)
	}

	// Создание HTTP клиента с mTLS поддержкой
	httpClient, err := createHTTPClient(config)
	if err != nil {
		return fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Получение токена
	token, err := getAccessToken(httpClient, config)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	if debug {
		fmt.Fprintf(os.Stderr, "Debug: obtained access token (expires: %v)\n", token.Expiry)
	}

	// Создание ExecCredential ответа
	execCred := &clientauthv1beta1.ExecCredential{
		TypeMeta: metav1.TypeMeta{
			APIVersion: clientauthv1beta1.SchemeGroupVersion.String(),
			Kind:       "ExecCredential",
		},
		Status: &clientauthv1beta1.ExecCredentialStatus{
			ExpirationTimestamp: &metav1.Time{Time: token.Expiry},
			Token:               token.AccessToken,
		},
	}

	// Вывод JSON ответа
	output, err := json.Marshal(execCred)
	if err != nil {
		return fmt.Errorf("failed to marshal exec credential: %w", err)
	}

	fmt.Print(string(output))
	return nil
}

func loadConfig(configPath string) (*ZitadelConfig, error) {
	// Если путь не указан, попробовать найти конфигурацию
	if configPath == "" {
		// Попробовать стандартные места
		candidates := []string{
			os.Getenv("KUBECTL_ZITADEL_CONFIG"),
			"~/.kubectl-zitadel.json",
			"/etc/kubectl-zitadel/config.json",
			"/home/ubuntu/cloudbridge-relay-installer/configs/zitadel_config.json",
		}

		for _, candidate := range candidates {
			if candidate == "" {
				continue
			}
			if _, err := os.Stat(candidate); err == nil {
				configPath = candidate
				break
			}
		}

		if configPath == "" {
			return nil, fmt.Errorf("no configuration file found, please specify --config")
		}
	}

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config ZitadelConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Валидация конфигурации
	if config.Domain == "" {
		return nil, fmt.Errorf("domain is required in config")
	}
	if config.ClientID == "" {
		return nil, fmt.Errorf("client_id is required in config")
	}
	if config.ClientSecret == "" {
		return nil, fmt.Errorf("client_secret is required in config")
	}

	// Установка дефолтных scopes если не указаны
	if len(config.Scopes) == 0 {
		config.Scopes = []string{"openid", "profile", "relay.api", "relay.k8s"}
	}

	return &config, nil
}

func createHTTPClient(config *ZitadelConfig) (*http.Client, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Настроить TLS конфигурацию (включая insecure режим)
	tlsConfig, err := createTLSConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS config: %w", err)
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	client.Transport = transport

	return client, nil
}

func createTLSConfig(config *ZitadelConfig) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: config.InsecureSkipVerify,
	}

	// Загрузка CA сертификата если указан
	if config.CACertPath != "" {
		caCert, err := ioutil.ReadFile(config.CACertPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA cert: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		tlsConfig.RootCAs = caCertPool
	}

	// Загрузка клиентского сертификата и ключа
	if config.ClientCertPath != "" && config.ClientKeyPath != "" {
		clientCert, err := tls.LoadX509KeyPair(config.ClientCertPath, config.ClientKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{clientCert}
	}

	return tlsConfig, nil
}

func getAccessToken(httpClient *http.Client, config *ZitadelConfig) (*oauth2.Token, error) {
	ctx := context.Background()

	// Настройка OAuth2 client credentials flow
	oauth2Config := clientcredentials.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		TokenURL:     fmt.Sprintf("https://%s/oauth/v2/token", config.Domain),
		Scopes:       config.Scopes,
	}

	// Контекст с кастомным HTTP клиентом
	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)

	// Получение токена
	token, err := oauth2Config.Token(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	// Если включен Token Exchange, выполнить обмен токенов
	if config.TokenExchange {
		exchangedToken, err := performTokenExchange(httpClient, config, token.AccessToken)
		if err != nil {
			return nil, fmt.Errorf("token exchange failed: %w", err)
		}
		return exchangedToken, nil
	}

	return token, nil
}

func performTokenExchange(httpClient *http.Client, config *ZitadelConfig, subjectToken string) (*oauth2.Token, error) {
	// RFC 8693 Token Exchange
	data := url.Values{
		"grant_type":         {"urn:ietf:params:oauth:grant-type:token-exchange"},
		"subject_token":      {subjectToken},
		"subject_token_type": {"urn:ietf:params:oauth:token-type:access_token"},
		"scope":              {strings.Join(config.Scopes, " ")},
	}

	if config.Audience != "" {
		data.Set("audience", config.Audience)
	}

	tokenURL := fmt.Sprintf("https://%s/oauth/v2/token", config.Domain)
	resp, err := httpClient.PostForm(tokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("token exchange request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int64  `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	token := &oauth2.Token{
		AccessToken: tokenResp.AccessToken,
		TokenType:   tokenResp.TokenType,
	}

	if tokenResp.ExpiresIn > 0 {
		token.Expiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	}

	return token, nil
}
