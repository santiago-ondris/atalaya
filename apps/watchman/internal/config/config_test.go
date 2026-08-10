package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santiago-ondris/atalaya/apps/watchman/internal/domain"
)

func TestLoadSentryCatalogMergesPolicyOverrides(t *testing.T) {
	path := writeCatalog(t, `{"applications":[{"slug":"farmami","alert_policy":{"rate_limit_count":3},"integrations":[
		{"component":"frontend","display_name":"Frontend","project":"farmami-frontend","environments":["vercel-production"]}
	]}]}`)
	defaults := domain.AlertPolicy{Enabled: true, AlwaysAlertSeverities: []string{"critical", "high"},
		ActionableAlertSeverities: []string{"medium"}, DeduplicationWindowSeconds: 900,
		RateLimitWindowSeconds: 600, RateLimitCount: 10}
	applications, err := loadSentryCatalog(path, defaults)
	if err != nil {
		t.Fatal(err)
	}
	if len(applications) != 1 || applications[0].AlertPolicy.RateLimitCount != 3 || applications[0].AlertPolicy.DeduplicationWindowSeconds != 900 {
		t.Fatalf("unexpected catalog: %#v", applications)
	}
	if applications[0].Integrations[0].Environments[0] != "vercel-production" {
		t.Fatalf("unexpected environments: %#v", applications[0].Integrations[0].Environments)
	}
}

func TestLoadSentryCatalogRejectsDuplicateComponents(t *testing.T) {
	path := writeCatalog(t, `{"applications":[{"slug":"prensap","integrations":[
		{"component":"backend","display_name":"Backend","project":"one","environments":["production"]},
		{"component":"backend","display_name":"Backend","project":"two","environments":["production"]}
	]}]}`)
	_, err := loadSentryCatalog(path, domain.AlertPolicy{Enabled: true, DeduplicationWindowSeconds: 1, RateLimitWindowSeconds: 1, RateLimitCount: 1})
	if err == nil || !strings.Contains(err.Error(), "duplicate component") {
		t.Fatalf("expected duplicate component error, got %v", err)
	}
}

func TestLoadRejectsPartialAzureConfiguration(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("ATALAYA_ADMIN_PASSWORD_HASH", "test-only-hash")
	t.Setenv("SENTRY_CATALOG_PATH", writeCatalog(t, `{"applications":[{"slug":"prensap","integrations":[{"component":"backend","display_name":"Backend","project":"prensap","environments":["production"]}]}]}`))
	t.Setenv("AZURE_CLIENT_ID", "configured-without-the-other-required-values")
	t.Setenv("AZURE_CLIENT_SECRET", "")
	t.Setenv("AZURE_TENANT_ID", "")
	t.Setenv("AZURE_LOG_ANALYTICS_WORKSPACE_ID", "")
	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "must be configured together") {
		t.Fatalf("expected partial Azure configuration error, got %v", err)
	}
}

func writeCatalog(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "catalog.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
