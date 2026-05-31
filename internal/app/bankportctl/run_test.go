package bankportctl

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/allanflavio/bankport-go-gin-partner-api/internal/config"
	"github.com/allanflavio/bankport-go-gin-partner-api/internal/domain"
	"github.com/allanflavio/bankport-go-gin-partner-api/internal/store"
)

func TestAppsListJSON(t *testing.T) {
	var stdout bytes.Buffer
	app := newTestApp(&stdout)

	if err := app.Run(context.Background(), []string{"apps", "list", "--format", "json"}); err != nil {
		t.Fatalf("apps list: %v", err)
	}

	var apps []domain.PartnerApp
	if err := json.Unmarshal(stdout.Bytes(), &apps); err != nil {
		t.Fatalf("parse apps json: %v", err)
	}
	if len(apps) != 3 {
		t.Fatalf("expected three apps, got %d", len(apps))
	}
	if apps[0].DeveloperAppID != "app_other_partner" {
		t.Fatalf("expected stable app ordering, got %+v", apps)
	}
}

func TestRateLimitsInspectTable(t *testing.T) {
	var stdout bytes.Buffer
	app := newTestApp(&stdout)

	if err := app.Run(context.Background(), []string{"rate-limits", "inspect"}); err != nil {
		t.Fatalf("rate limits inspect: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "DEVELOPER_APP_ID") || !strings.Contains(output, "in_memory_sandbox") {
		t.Fatalf("expected rate limit table, got %s", output)
	}
}

func TestUsageReportJSON(t *testing.T) {
	var stdout bytes.Buffer
	app := newTestApp(&stdout)

	if err := app.Run(context.Background(), []string{"usage", "report", "--format", "json"}); err != nil {
		t.Fatalf("usage report: %v", err)
	}

	var report domain.UsageReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("parse usage report json: %v", err)
	}
	if report.PartnerCount != 2 || report.DeveloperAppCount != 3 || report.AccountCount != 3 {
		t.Fatalf("unexpected report counts: %+v", report)
	}
}

func TestUnknownCommandReturnsUsageError(t *testing.T) {
	var stdout bytes.Buffer
	app := newTestApp(&stdout)

	err := app.Run(context.Background(), []string{"tokens", "rotate"})

	if !errors.Is(err, ErrUsage) {
		t.Fatalf("expected usage error, got %v", err)
	}
}

func newTestApp(stdout *bytes.Buffer) App {
	cfg := config.Load()
	return App{
		Config:     cfg,
		Repository: store.NewSeededRepository(cfg),
		Stdout:     stdout,
		Stderr:     &bytes.Buffer{},
	}
}
