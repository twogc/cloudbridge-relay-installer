package cadence

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Тесты для структур данных activities
// Поскольку функции activities требуют Cadence activity контекст,
// мы тестируем только структуры данных и их валидацию

func TestConfigValidationResult(t *testing.T) {
	t.Run("valid result", func(t *testing.T) {
		result := ConfigValidationResult{
			Valid: true,
			Error: "",
		}

		assert.True(t, result.Valid)
		assert.Empty(t, result.Error)
	})

	t.Run("invalid result", func(t *testing.T) {
		result := ConfigValidationResult{
			Valid: false,
			Error: "configuration error",
		}

		assert.False(t, result.Valid)
		assert.Equal(t, "configuration error", result.Error)
	})
}

func TestWireGuardSetupResult(t *testing.T) {
	t.Run("successful setup", func(t *testing.T) {
		result := WireGuardSetupResult{
			Success: true,
			Error:   "",
			Config:  "wg0",
		}

		assert.True(t, result.Success)
		assert.Equal(t, "wg0", result.Config)
		assert.Empty(t, result.Error)
	})

	t.Run("failed setup", func(t *testing.T) {
		result := WireGuardSetupResult{
			Success: false,
			Error:   "setup failed",
			Config:  "",
		}

		assert.False(t, result.Success)
		assert.Equal(t, "setup failed", result.Error)
		assert.Empty(t, result.Config)
	})
}

func TestQUICSetupResult(t *testing.T) {
	t.Run("successful setup", func(t *testing.T) {
		result := QUICSetupResult{
			Success: true,
			Error:   "",
			Port:    9090,
		}

		assert.True(t, result.Success)
		assert.Equal(t, 9090, result.Port)
		assert.Empty(t, result.Error)
	})

	t.Run("failed setup", func(t *testing.T) {
		result := QUICSetupResult{
			Success: false,
			Error:   "port already in use",
			Port:    0,
		}

		assert.False(t, result.Success)
		assert.Equal(t, "port already in use", result.Error)
		assert.Equal(t, 0, result.Port)
	})
}

func TestServiceStartResult(t *testing.T) {
	t.Run("successful start", func(t *testing.T) {
		services := []string{"relay", "wireguard", "quic"}
		result := ServiceStartResult{
			Success:  true,
			Error:    "",
			Services: services,
		}

		assert.True(t, result.Success)
		assert.Equal(t, services, result.Services)
		assert.Empty(t, result.Error)
	})

	t.Run("failed start", func(t *testing.T) {
		result := ServiceStartResult{
			Success:  false,
			Error:    "service failed to start",
			Services: []string{},
		}

		assert.False(t, result.Success)
		assert.Equal(t, "service failed to start", result.Error)
		assert.Empty(t, result.Services)
	})
}

func TestHealthCheckInput(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		input := HealthCheckInput{
			Services: []string{"relay", "wireguard"},
			Timeout:  30 * time.Second,
		}

		assert.Len(t, input.Services, 2)
		assert.Contains(t, input.Services, "relay")
		assert.Contains(t, input.Services, "wireguard")
		assert.Equal(t, 30*time.Second, input.Timeout)
	})

	t.Run("empty services", func(t *testing.T) {
		input := HealthCheckInput{
			Services: []string{},
			Timeout:  10 * time.Second,
		}

		assert.Empty(t, input.Services)
		assert.Equal(t, 10*time.Second, input.Timeout)
	})
}

func TestHealthCheckResult(t *testing.T) {
	t.Run("healthy result", func(t *testing.T) {
		serviceStatus := map[string]bool{
			"relay":     true,
			"wireguard": true,
			"quic":      true,
		}

		result := HealthCheckResult{
			Healthy:  true,
			Services: serviceStatus,
			Error:    "",
		}

		assert.True(t, result.Healthy)
		assert.Len(t, result.Services, 3)
		assert.True(t, result.Services["relay"])
		assert.True(t, result.Services["wireguard"])
		assert.True(t, result.Services["quic"])
		assert.Empty(t, result.Error)
	})

	t.Run("unhealthy result", func(t *testing.T) {
		serviceStatus := map[string]bool{
			"relay":     true,
			"wireguard": false,
			"quic":      true,
		}

		result := HealthCheckResult{
			Healthy:  false,
			Services: serviceStatus,
			Error:    "wireguard service is down",
		}

		assert.False(t, result.Healthy)
		assert.False(t, result.Services["wireguard"])
		assert.Equal(t, "wireguard service is down", result.Error)
	})
}

func TestCertificateGenerationResult(t *testing.T) {
	t.Run("successful generation", func(t *testing.T) {
		certificates := []string{"/etc/ssl/certs/relay.crt", "/etc/ssl/private/relay.key"}
		result := CertificateGenerationResult{
			Success:      true,
			Error:        "",
			Certificates: certificates,
		}

		assert.True(t, result.Success)
		assert.Len(t, result.Certificates, 2)
		assert.Contains(t, result.Certificates, "/etc/ssl/certs/relay.crt")
		assert.Contains(t, result.Certificates, "/etc/ssl/private/relay.key")
		assert.Empty(t, result.Error)
	})

	t.Run("failed generation", func(t *testing.T) {
		result := CertificateGenerationResult{
			Success:      false,
			Error:        "certificate generation failed",
			Certificates: []string{},
		}

		assert.False(t, result.Success)
		assert.Equal(t, "certificate generation failed", result.Error)
		assert.Empty(t, result.Certificates)
	})
}

func TestCertificateValidationResult(t *testing.T) {
	t.Run("valid certificates", func(t *testing.T) {
		result := CertificateValidationResult{
			Valid: true,
			Error: "",
		}

		assert.True(t, result.Valid)
		assert.Empty(t, result.Error)
	})

	t.Run("invalid certificates", func(t *testing.T) {
		result := CertificateValidationResult{
			Valid: false,
			Error: "certificate expired",
		}

		assert.False(t, result.Valid)
		assert.Equal(t, "certificate expired", result.Error)
	})
}

func TestUpdateCertificateConfigInput(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		input := UpdateCertificateConfigInput{
			Certificates: []string{"/etc/ssl/certs/relay.crt"},
			Services:     []string{"relay", "quic"},
		}

		assert.Len(t, input.Certificates, 1)
		assert.Len(t, input.Services, 2)
		assert.Contains(t, input.Services, "relay")
		assert.Contains(t, input.Services, "quic")
	})

	t.Run("empty input", func(t *testing.T) {
		input := UpdateCertificateConfigInput{
			Certificates: []string{},
			Services:     []string{},
		}

		assert.Empty(t, input.Certificates)
		assert.Empty(t, input.Services)
	})
}

func TestConfigUpdateResult(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		result := ConfigUpdateResult{
			Success: true,
			Error:   "",
		}

		assert.True(t, result.Success)
		assert.Empty(t, result.Error)
	})

	t.Run("failed update", func(t *testing.T) {
		result := ConfigUpdateResult{
			Success: false,
			Error:   "update failed",
		}

		assert.False(t, result.Success)
		assert.Equal(t, "update failed", result.Error)
	})
}

func TestServiceRestartResult(t *testing.T) {
	t.Run("successful restart", func(t *testing.T) {
		services := []string{"relay", "wireguard"}
		result := ServiceRestartResult{
			Success:  true,
			Error:    "",
			Services: services,
		}

		assert.True(t, result.Success)
		assert.Equal(t, services, result.Services)
		assert.Empty(t, result.Error)
	})

	t.Run("failed restart", func(t *testing.T) {
		result := ServiceRestartResult{
			Success:  false,
			Error:    "restart failed",
			Services: []string{},
		}

		assert.False(t, result.Success)
		assert.Equal(t, "restart failed", result.Error)
		assert.Empty(t, result.Services)
	})
}

func TestNetworkAnalysisResult(t *testing.T) {
	t.Run("healthy network", func(t *testing.T) {
		metrics := map[string]interface{}{
			"latency":   10,
			"bandwidth": 1000,
		}

		result := NetworkAnalysisResult{
			Issues:          []string{},
			Metrics:         metrics,
			Recommendations: []string{"Network is healthy"},
		}

		assert.Empty(t, result.Issues)
		assert.Equal(t, 10, result.Metrics["latency"])
		assert.Equal(t, 1000, result.Metrics["bandwidth"])
		assert.Contains(t, result.Recommendations, "Network is healthy")
	})

	t.Run("network with issues", func(t *testing.T) {
		metrics := map[string]interface{}{
			"latency":   100,
			"bandwidth": 100,
		}

		result := NetworkAnalysisResult{
			Issues:          []string{"High latency", "Low bandwidth"},
			Metrics:         metrics,
			Recommendations: []string{"Optimize network configuration"},
		}

		assert.Len(t, result.Issues, 2)
		assert.Contains(t, result.Issues, "High latency")
		assert.Contains(t, result.Issues, "Low bandwidth")
		assert.Equal(t, 100, result.Metrics["latency"])
		assert.Equal(t, 100, result.Metrics["bandwidth"])
		assert.Contains(t, result.Recommendations, "Optimize network configuration")
	})
}

func TestMaintenancePlanResult(t *testing.T) {
	t.Run("empty plan", func(t *testing.T) {
		result := MaintenancePlanResult{
			Tasks: []MaintenanceTask{},
		}

		assert.Empty(t, result.Tasks)
	})

	t.Run("plan with tasks", func(t *testing.T) {
		tasks := []MaintenanceTask{
			{
				Name:     "update certificates",
				Type:     "security",
				Config:   map[string]interface{}{"auto": true},
				Priority: 1,
			},
			{
				Name:     "cleanup logs",
				Type:     "maintenance",
				Config:   map[string]interface{}{"retention": 30},
				Priority: 2,
			},
		}

		result := MaintenancePlanResult{
			Tasks: tasks,
		}

		assert.Len(t, result.Tasks, 2)
		assert.Equal(t, "update certificates", result.Tasks[0].Name)
		assert.Equal(t, "security", result.Tasks[0].Type)
		assert.Equal(t, 1, result.Tasks[0].Priority)
		assert.Equal(t, "cleanup logs", result.Tasks[1].Name)
		assert.Equal(t, "maintenance", result.Tasks[1].Type)
		assert.Equal(t, 2, result.Tasks[1].Priority)
	})
}

func TestMaintenanceTask(t *testing.T) {
	t.Run("valid task", func(t *testing.T) {
		task := MaintenanceTask{
			Name:     "test task",
			Type:     "network",
			Config:   map[string]interface{}{"timeout": 30},
			Priority: 1,
		}

		assert.Equal(t, "test task", task.Name)
		assert.Equal(t, "network", task.Type)
		if config, ok := task.Config.(map[string]interface{}); ok {
			assert.Equal(t, 30, config["timeout"])
		}
		assert.Equal(t, 1, task.Priority)
	})

	t.Run("empty task", func(t *testing.T) {
		task := MaintenanceTask{
			Name:     "",
			Type:     "",
			Config:   nil,
			Priority: 0,
		}

		assert.Empty(t, task.Name)
		assert.Empty(t, task.Type)
		assert.Nil(t, task.Config)
		assert.Equal(t, 0, task.Priority)
	})
}

func TestMaintenanceTaskResult(t *testing.T) {
	t.Run("successful task", func(t *testing.T) {
		result := MaintenanceTaskResult{
			Success: true,
			Error:   "",
			Result:  "Task completed successfully",
		}

		assert.True(t, result.Success)
		assert.Empty(t, result.Error)
		assert.Equal(t, "Task completed successfully", result.Result)
	})

	t.Run("failed task", func(t *testing.T) {
		result := MaintenanceTaskResult{
			Success: false,
			Error:   "task failed",
			Result:  nil,
		}

		assert.False(t, result.Success)
		assert.Equal(t, "task failed", result.Error)
		assert.Nil(t, result.Result)
	})
}

func TestVulnerabilityScanResult(t *testing.T) {
	t.Run("clean scan", func(t *testing.T) {
		scanTime := time.Now()
		result := VulnerabilityScanResult{
			Vulnerabilities: []Vulnerability{},
			ScanTime:        scanTime,
		}

		assert.Empty(t, result.Vulnerabilities)
		assert.Equal(t, scanTime, result.ScanTime)
	})

	t.Run("scan with vulnerabilities", func(t *testing.T) {
		scanTime := time.Now()
		vulnerabilities := []Vulnerability{
			{
				ID:          "CVE-2023-1234",
				Severity:    "high",
				Description: "Remote code execution",
				Component:   "openssl",
			},
			{
				ID:          "CVE-2023-5678",
				Severity:    "medium",
				Description: "Information disclosure",
				Component:   "nginx",
			},
		}

		result := VulnerabilityScanResult{
			Vulnerabilities: vulnerabilities,
			ScanTime:        scanTime,
		}

		assert.Len(t, result.Vulnerabilities, 2)
		assert.Equal(t, "CVE-2023-1234", result.Vulnerabilities[0].ID)
		assert.Equal(t, "high", result.Vulnerabilities[0].Severity)
		assert.Equal(t, "CVE-2023-5678", result.Vulnerabilities[1].ID)
		assert.Equal(t, "medium", result.Vulnerabilities[1].Severity)
		assert.Equal(t, scanTime, result.ScanTime)
	})
}

func TestVulnerability(t *testing.T) {
	t.Run("valid vulnerability", func(t *testing.T) {
		vuln := Vulnerability{
			ID:          "CVE-2023-1234",
			Severity:    "high",
			Description: "Remote code execution vulnerability",
			Component:   "openssl",
		}

		assert.Equal(t, "CVE-2023-1234", vuln.ID)
		assert.Equal(t, "high", vuln.Severity)
		assert.Equal(t, "Remote code execution vulnerability", vuln.Description)
		assert.Equal(t, "openssl", vuln.Component)
	})

	t.Run("empty vulnerability", func(t *testing.T) {
		vuln := Vulnerability{
			ID:          "",
			Severity:    "",
			Description: "",
			Component:   "",
		}

		assert.Empty(t, vuln.ID)
		assert.Empty(t, vuln.Severity)
		assert.Empty(t, vuln.Description)
		assert.Empty(t, vuln.Component)
	})
}

func TestSecurityConfigAnalysisResult(t *testing.T) {
	t.Run("clean analysis", func(t *testing.T) {
		result := SecurityConfigAnalysisResult{
			Issues: []SecurityIssue{},
		}

		assert.Empty(t, result.Issues)
	})

	t.Run("analysis with issues", func(t *testing.T) {
		issues := []SecurityIssue{
			{
				Type:        "firewall",
				Severity:    "high",
				Description: "Firewall is disabled",
				Component:   "iptables",
			},
		}

		result := SecurityConfigAnalysisResult{
			Issues: issues,
		}

		assert.Len(t, result.Issues, 1)
		assert.Equal(t, "firewall", result.Issues[0].Type)
		assert.Equal(t, "high", result.Issues[0].Severity)
		assert.Equal(t, "Firewall is disabled", result.Issues[0].Description)
		assert.Equal(t, "iptables", result.Issues[0].Component)
	})
}

func TestSecurityIssue(t *testing.T) {
	t.Run("valid security issue", func(t *testing.T) {
		issue := SecurityIssue{
			Type:        "authentication",
			Severity:    "medium",
			Description: "Weak password policy",
			Component:   "auth-service",
		}

		assert.Equal(t, "authentication", issue.Type)
		assert.Equal(t, "medium", issue.Severity)
		assert.Equal(t, "Weak password policy", issue.Description)
		assert.Equal(t, "auth-service", issue.Component)
	})

	t.Run("empty security issue", func(t *testing.T) {
		issue := SecurityIssue{
			Type:        "",
			Severity:    "",
			Description: "",
			Component:   "",
		}

		assert.Empty(t, issue.Type)
		assert.Empty(t, issue.Severity)
		assert.Empty(t, issue.Description)
		assert.Empty(t, issue.Component)
	})
}

func TestComplianceCheckResult(t *testing.T) {
	t.Run("compliant", func(t *testing.T) {
		result := ComplianceCheckResult{
			Issues: []ComplianceIssue{},
		}

		assert.Empty(t, result.Issues)
	})

	t.Run("non-compliant", func(t *testing.T) {
		issues := []ComplianceIssue{
			{
				Standard:    "ISO27001",
				Requirement: "Access control",
				Status:      "non-compliant",
				Description: "Missing access control policies",
			},
		}

		result := ComplianceCheckResult{
			Issues: issues,
		}

		assert.Len(t, result.Issues, 1)
		assert.Equal(t, "ISO27001", result.Issues[0].Standard)
		assert.Equal(t, "Access control", result.Issues[0].Requirement)
		assert.Equal(t, "non-compliant", result.Issues[0].Status)
		assert.Equal(t, "Missing access control policies", result.Issues[0].Description)
	})
}

func TestComplianceIssue(t *testing.T) {
	t.Run("valid compliance issue", func(t *testing.T) {
		issue := ComplianceIssue{
			Standard:    "SOC2",
			Requirement: "Data encryption",
			Status:      "non-compliant",
			Description: "Data not encrypted at rest",
		}

		assert.Equal(t, "SOC2", issue.Standard)
		assert.Equal(t, "Data encryption", issue.Requirement)
		assert.Equal(t, "non-compliant", issue.Status)
		assert.Equal(t, "Data not encrypted at rest", issue.Description)
	})

	t.Run("empty compliance issue", func(t *testing.T) {
		issue := ComplianceIssue{
			Standard:    "",
			Requirement: "",
			Status:      "",
			Description: "",
		}

		assert.Empty(t, issue.Standard)
		assert.Empty(t, issue.Requirement)
		assert.Empty(t, issue.Status)
		assert.Empty(t, issue.Description)
	})
}

func TestSecurityReportInput(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		vulnerabilities := []Vulnerability{
			{ID: "CVE-2023-1234", Severity: "high"},
		}
		configIssues := []SecurityIssue{
			{Type: "firewall", Severity: "medium"},
		}
		complianceIssues := []ComplianceIssue{
			{Standard: "ISO27001", Status: "non-compliant"},
		}

		input := SecurityReportInput{
			Vulnerabilities:  vulnerabilities,
			ConfigIssues:     configIssues,
			ComplianceIssues: complianceIssues,
		}

		assert.Len(t, input.Vulnerabilities, 1)
		assert.Len(t, input.ConfigIssues, 1)
		assert.Len(t, input.ComplianceIssues, 1)
		assert.Equal(t, "CVE-2023-1234", input.Vulnerabilities[0].ID)
		assert.Equal(t, "firewall", input.ConfigIssues[0].Type)
		assert.Equal(t, "ISO27001", input.ComplianceIssues[0].Standard)
	})

	t.Run("empty input", func(t *testing.T) {
		input := SecurityReportInput{
			Vulnerabilities:  []Vulnerability{},
			ConfigIssues:     []SecurityIssue{},
			ComplianceIssues: []ComplianceIssue{},
		}

		assert.Empty(t, input.Vulnerabilities)
		assert.Empty(t, input.ConfigIssues)
		assert.Empty(t, input.ComplianceIssues)
	})
}

func TestSecurityReportResult(t *testing.T) {
	t.Run("valid result", func(t *testing.T) {
		generatedAt := time.Now()
		result := SecurityReportResult{
			ReportPath:  "/var/log/security-report.json",
			GeneratedAt: generatedAt,
		}

		assert.Equal(t, "/var/log/security-report.json", result.ReportPath)
		assert.Equal(t, generatedAt, result.GeneratedAt)
	})

	t.Run("empty result", func(t *testing.T) {
		result := SecurityReportResult{
			ReportPath:  "",
			GeneratedAt: time.Time{},
		}

		assert.Empty(t, result.ReportPath)
		assert.Zero(t, result.GeneratedAt)
	})
}

func TestBackupPreparationResult(t *testing.T) {
	t.Run("successful preparation", func(t *testing.T) {
		result := BackupPreparationResult{
			Success: true,
			Error:   "",
		}

		assert.True(t, result.Success)
		assert.Empty(t, result.Error)
	})

	t.Run("failed preparation", func(t *testing.T) {
		result := BackupPreparationResult{
			Success: false,
			Error:   "preparation failed",
		}

		assert.False(t, result.Success)
		assert.Equal(t, "preparation failed", result.Error)
	})
}

func TestConfigBackupResult(t *testing.T) {
	t.Run("successful backup", func(t *testing.T) {
		result := ConfigBackupResult{
			Success:    true,
			Error:      "",
			BackupPath: "/backup/config-backup.tar.gz",
		}

		assert.True(t, result.Success)
		assert.Empty(t, result.Error)
		assert.Equal(t, "/backup/config-backup.tar.gz", result.BackupPath)
	})

	t.Run("failed backup", func(t *testing.T) {
		result := ConfigBackupResult{
			Success:    false,
			Error:      "backup failed",
			BackupPath: "",
		}

		assert.False(t, result.Success)
		assert.Equal(t, "backup failed", result.Error)
		assert.Empty(t, result.BackupPath)
	})
}

func TestDataBackupResult(t *testing.T) {
	t.Run("successful backup", func(t *testing.T) {
		result := DataBackupResult{
			Success:    true,
			Error:      "",
			BackupPath: "/backup/data-backup.tar.gz",
		}

		assert.True(t, result.Success)
		assert.Empty(t, result.Error)
		assert.Equal(t, "/backup/data-backup.tar.gz", result.BackupPath)
	})

	t.Run("failed backup", func(t *testing.T) {
		result := DataBackupResult{
			Success:    false,
			Error:      "backup failed",
			BackupPath: "",
		}

		assert.False(t, result.Success)
		assert.Equal(t, "backup failed", result.Error)
		assert.Empty(t, result.BackupPath)
	})
}

func TestBackupValidationInput(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		input := BackupValidationInput{
			ConfigBackup: "/backup/config-backup.tar.gz",
			DataBackup:   "/backup/data-backup.tar.gz",
		}

		assert.Equal(t, "/backup/config-backup.tar.gz", input.ConfigBackup)
		assert.Equal(t, "/backup/data-backup.tar.gz", input.DataBackup)
	})

	t.Run("empty input", func(t *testing.T) {
		input := BackupValidationInput{
			ConfigBackup: "",
			DataBackup:   "",
		}

		assert.Empty(t, input.ConfigBackup)
		assert.Empty(t, input.DataBackup)
	})
}

func TestBackupValidationResult(t *testing.T) {
	t.Run("valid backup", func(t *testing.T) {
		result := BackupValidationResult{
			Valid: true,
			Error: "",
		}

		assert.True(t, result.Valid)
		assert.Empty(t, result.Error)
	})

	t.Run("invalid backup", func(t *testing.T) {
		result := BackupValidationResult{
			Valid: false,
			Error: "backup corrupted",
		}

		assert.False(t, result.Valid)
		assert.Equal(t, "backup corrupted", result.Error)
	})
}

func TestDamageAssessmentResult(t *testing.T) {
	t.Run("low damage", func(t *testing.T) {
		result := DamageAssessmentResult{
			Severity: "low",
			Affected: []string{},
			Issues:   []string{},
		}

		assert.Equal(t, "low", result.Severity)
		assert.Empty(t, result.Affected)
		assert.Empty(t, result.Issues)
	})

	t.Run("high damage", func(t *testing.T) {
		result := DamageAssessmentResult{
			Severity: "high",
			Affected: []string{"database", "config"},
			Issues:   []string{"data corruption", "config loss"},
		}

		assert.Equal(t, "high", result.Severity)
		assert.Len(t, result.Affected, 2)
		assert.Contains(t, result.Affected, "database")
		assert.Contains(t, result.Affected, "config")
		assert.Len(t, result.Issues, 2)
		assert.Contains(t, result.Issues, "data corruption")
		assert.Contains(t, result.Issues, "config loss")
	})
}

func TestConfigRecoveryResult(t *testing.T) {
	t.Run("successful recovery", func(t *testing.T) {
		result := ConfigRecoveryResult{
			Success: true,
			Error:   "",
		}

		assert.True(t, result.Success)
		assert.Empty(t, result.Error)
	})

	t.Run("failed recovery", func(t *testing.T) {
		result := ConfigRecoveryResult{
			Success: false,
			Error:   "recovery failed",
		}

		assert.False(t, result.Success)
		assert.Equal(t, "recovery failed", result.Error)
	})
}

func TestDataRecoveryResult(t *testing.T) {
	t.Run("successful recovery", func(t *testing.T) {
		result := DataRecoveryResult{
			Success: true,
			Error:   "",
		}

		assert.True(t, result.Success)
		assert.Empty(t, result.Error)
	})

	t.Run("failed recovery", func(t *testing.T) {
		result := DataRecoveryResult{
			Success: false,
			Error:   "recovery failed",
		}

		assert.False(t, result.Success)
		assert.Equal(t, "recovery failed", result.Error)
	})
}

// Тесты для проверки корректности структур данных
func TestActivityDataStructures(t *testing.T) {
	t.Run("all result types have success/valid fields", func(t *testing.T) {
		// Проверяем, что все структуры результатов имеют поля для статуса
		configResult := ConfigValidationResult{Valid: true}
		assert.True(t, configResult.Valid)

		wireguardResult := WireGuardSetupResult{Success: true}
		assert.True(t, wireguardResult.Success)

		quicResult := QUICSetupResult{Success: true}
		assert.True(t, quicResult.Success)

		serviceResult := ServiceStartResult{Success: true}
		assert.True(t, serviceResult.Success)

		healthResult := HealthCheckResult{Healthy: true}
		assert.True(t, healthResult.Healthy)

		certResult := CertificateGenerationResult{Success: true}
		assert.True(t, certResult.Success)

		certValidResult := CertificateValidationResult{Valid: true}
		assert.True(t, certValidResult.Valid)

		updateResult := ConfigUpdateResult{Success: true}
		assert.True(t, updateResult.Success)

		restartResult := ServiceRestartResult{Success: true}
		assert.True(t, restartResult.Success)

		taskResult := MaintenanceTaskResult{Success: true}
		assert.True(t, taskResult.Success)

		prepResult := BackupPreparationResult{Success: true}
		assert.True(t, prepResult.Success)

		configBackupResult := ConfigBackupResult{Success: true}
		assert.True(t, configBackupResult.Success)

		dataBackupResult := DataBackupResult{Success: true}
		assert.True(t, dataBackupResult.Success)

		backupValidResult := BackupValidationResult{Valid: true}
		assert.True(t, backupValidResult.Valid)

		configRecoveryResult := ConfigRecoveryResult{Success: true}
		assert.True(t, configRecoveryResult.Success)

		dataRecoveryResult := DataRecoveryResult{Success: true}
		assert.True(t, dataRecoveryResult.Success)
	})

	t.Run("all input types have required fields", func(t *testing.T) {
		// Проверяем, что все структуры входных данных имеют необходимые поля
		healthInput := HealthCheckInput{
			Services: []string{"test"},
			Timeout:  time.Second,
		}
		assert.NotEmpty(t, healthInput.Services)
		assert.NotZero(t, healthInput.Timeout)

		updateInput := UpdateCertificateConfigInput{
			Certificates: []string{"cert.pem"},
			Services:     []string{"service"},
		}
		assert.NotEmpty(t, updateInput.Certificates)
		assert.NotEmpty(t, updateInput.Services)

		reportInput := SecurityReportInput{
			Vulnerabilities:  []Vulnerability{},
			ConfigIssues:     []SecurityIssue{},
			ComplianceIssues: []ComplianceIssue{},
		}
		assert.NotNil(t, reportInput.Vulnerabilities)
		assert.NotNil(t, reportInput.ConfigIssues)
		assert.NotNil(t, reportInput.ComplianceIssues)

		validationInput := BackupValidationInput{
			ConfigBackup: "config.tar.gz",
			DataBackup:   "data.tar.gz",
		}
		assert.NotEmpty(t, validationInput.ConfigBackup)
		assert.NotEmpty(t, validationInput.DataBackup)
	})
}
