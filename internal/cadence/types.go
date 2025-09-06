package cadence

import (
	"time"
)

// RelayDeploymentInput входные данные для развертывания relay
type RelayDeploymentInput struct {
	Config         interface{}     `json:"config"`
	WireGuardConfig interface{}    `json:"wireguard_config"`
	QUICConfig     interface{}     `json:"quic_config"`
	Services       []string        `json:"services"`
	Timeout        time.Duration   `json:"timeout"`
}

// ConfigValidationResult результат валидации конфигурации
type ConfigValidationResult struct {
	Valid bool   `json:"valid"`
	Error string `json:"error,omitempty"`
}

// WireGuardSetupResult результат настройки WireGuard
type WireGuardSetupResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Config  string `json:"config,omitempty"`
}

// QUICSetupResult результат настройки QUIC
type QUICSetupResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Port    int    `json:"port,omitempty"`
}

// ServiceStartResult результат запуска сервисов
type ServiceStartResult struct {
	Success bool     `json:"success"`
	Error   string   `json:"error,omitempty"`
	Services []string `json:"services,omitempty"`
}

// HealthCheckInput входные данные для проверки здоровья
type HealthCheckInput struct {
	Services []string      `json:"services"`
	Timeout  time.Duration `json:"timeout"`
}

// HealthCheckResult результат проверки здоровья
type HealthCheckResult struct {
	Healthy bool              `json:"healthy"`
	Services map[string]bool  `json:"services"`
	Error   string            `json:"error,omitempty"`
}

// CertificateRotationInput входные данные для ротации сертификатов
type CertificateRotationInput struct {
	CertConfig interface{} `json:"cert_config"`
	Services   []string    `json:"services"`
}

// CertificateGenerationResult результат генерации сертификатов
type CertificateGenerationResult struct {
	Success      bool     `json:"success"`
	Error        string   `json:"error,omitempty"`
	Certificates []string `json:"certificates,omitempty"`
}

// CertificateValidationResult результат валидации сертификатов
type CertificateValidationResult struct {
	Valid bool   `json:"valid"`
	Error string `json:"error,omitempty"`
}

// UpdateCertificateConfigInput входные данные для обновления конфигурации сертификатов
type UpdateCertificateConfigInput struct {
	Certificates []string `json:"certificates"`
	Services     []string `json:"services"`
}

// ConfigUpdateResult результат обновления конфигурации
type ConfigUpdateResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// ServiceRestartResult результат перезапуска сервисов
type ServiceRestartResult struct {
	Success bool     `json:"success"`
	Error   string   `json:"error,omitempty"`
	Services []string `json:"services,omitempty"`
}

// NetworkMaintenanceInput входные данные для обслуживания сети
type NetworkMaintenanceInput struct {
	NetworkConfig interface{} `json:"network_config"`
	Services      []string    `json:"services"`
}

// NetworkAnalysisResult результат анализа сети
type NetworkAnalysisResult struct {
	Issues    []string `json:"issues"`
	Metrics   map[string]interface{} `json:"metrics"`
	Recommendations []string `json:"recommendations"`
}

// MaintenancePlanResult результат планирования обслуживания
type MaintenancePlanResult struct {
	Tasks []MaintenanceTask `json:"tasks"`
}

// MaintenanceTask задача обслуживания
type MaintenanceTask struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Config      interface{} `json:"config"`
	Priority    int         `json:"priority"`
}

// MaintenanceTaskResult результат выполнения задачи обслуживания
type MaintenanceTaskResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Result  interface{} `json:"result,omitempty"`
}

// SecurityAuditInput входные данные для аудита безопасности
type SecurityAuditInput struct {
	ScanConfig      interface{} `json:"scan_config"`
	SecurityConfig  interface{} `json:"security_config"`
	ComplianceConfig interface{} `json:"compliance_config"`
}

// VulnerabilityScanResult результат сканирования уязвимостей
type VulnerabilityScanResult struct {
	Vulnerabilities []Vulnerability `json:"vulnerabilities"`
	ScanTime        time.Time       `json:"scan_time"`
}

// Vulnerability уязвимость
type Vulnerability struct {
	ID          string `json:"id"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Component   string `json:"component"`
}

// SecurityConfigAnalysisResult результат анализа конфигурации безопасности
type SecurityConfigAnalysisResult struct {
	Issues []SecurityIssue `json:"issues"`
}

// SecurityIssue проблема безопасности
type SecurityIssue struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Component   string `json:"component"`
}

// ComplianceCheckResult результат проверки соответствия
type ComplianceCheckResult struct {
	Issues []ComplianceIssue `json:"issues"`
}

// ComplianceIssue проблема соответствия
type ComplianceIssue struct {
	Standard   string `json:"standard"`
	Requirement string `json:"requirement"`
	Status     string `json:"status"`
	Description string `json:"description"`
}

// SecurityReportInput входные данные для генерации отчета безопасности
type SecurityReportInput struct {
	Vulnerabilities []Vulnerability   `json:"vulnerabilities"`
	ConfigIssues    []SecurityIssue   `json:"config_issues"`
	ComplianceIssues []ComplianceIssue `json:"compliance_issues"`
}

// SecurityReportResult результат генерации отчета безопасности
type SecurityReportResult struct {
	ReportPath string    `json:"report_path"`
	GeneratedAt time.Time `json:"generated_at"`
}

// BackupInput входные данные для резервного копирования
type BackupInput struct {
	BackupConfig  interface{} `json:"backup_config"`
	ConfigPaths   []string    `json:"config_paths"`
	DataPaths     []string    `json:"data_paths"`
	CleanupConfig interface{} `json:"cleanup_config"`
}

// BackupPreparationResult результат подготовки к резервному копированию
type BackupPreparationResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// ConfigBackupResult результат резервного копирования конфигурации
type ConfigBackupResult struct {
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
	BackupPath string `json:"backup_path,omitempty"`
}

// DataBackupResult результат резервного копирования данных
type DataBackupResult struct {
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
	BackupPath string `json:"backup_path,omitempty"`
}

// BackupValidationInput входные данные для валидации резервных копий
type BackupValidationInput struct {
	ConfigBackup string `json:"config_backup"`
	DataBackup   string `json:"data_backup"`
}

// BackupValidationResult результат валидации резервных копий
type BackupValidationResult struct {
	Valid bool   `json:"valid"`
	Error string `json:"error,omitempty"`
}

// DisasterRecoveryInput входные данные для восстановления после сбоев
type DisasterRecoveryInput struct {
	AssessmentConfig interface{} `json:"assessment_config"`
	ConfigBackup     string      `json:"config_backup"`
	DataBackup       string      `json:"data_backup"`
	Services         []string    `json:"services"`
}

// DamageAssessmentResult результат оценки ущерба
type DamageAssessmentResult struct {
	Severity string   `json:"severity"`
	Affected []string `json:"affected"`
	Issues   []string `json:"issues"`
}

// ConfigRecoveryResult результат восстановления конфигурации
type ConfigRecoveryResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// DataRecoveryResult результат восстановления данных
type DataRecoveryResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}



