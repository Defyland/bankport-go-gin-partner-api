package spec

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestReadmeContainsReviewerSections(t *testing.T) {
	content := readFile(t, "README.md")
	required := []string{
		"## 1. What is this product?",
		"## 5. Architecture overview",
		"## 8. API documentation",
		"## 11. Testing strategy",
		"## 13. Observability",
		"## 14. Security considerations",
		"## 16. How to evaluate in 5 minutes",
		"## 17. How to run locally",
		"## 18. How to run tests",
		"## 19. Failure scenarios",
		"## 20. Roadmap",
		"docs/spec-driven/verification-report.md",
		"docs/deployment/railway.md",
	}

	for _, snippet := range required {
		if !strings.Contains(content, snippet) {
			t.Fatalf("README.md is missing %q", snippet)
		}
	}
}

func TestMandatoryPartnerAPIPublicArtifactsExist(t *testing.T) {
	requiredDirs := []string{
		".github/workflows",
		"benchmarks/results",
		"db/migrations",
		"deployments/grafana/dashboards",
		"deployments/prometheus",
		"docs/adr",
		"docs/api",
		"docs/architecture",
		"docs/domain",
		"docs/events",
		"docs/runbooks",
		"docs/security",
		"docs/spec-driven",
		"internal/httpapi",
		"internal/httpapi/middleware",
		"internal/usecase",
		"internal/webhook",
		"spec",
	}
	requiredFiles := []string{
		".github/workflows/ci.yml",
		"Dockerfile",
		"Makefile",
		"README.md",
		"db/migrations/001_init.sql",
		"deployments/grafana/dashboards/bankport-partner-api.json",
		"deployments/prometheus/alerts.yml",
		"deployments/prometheus/prometheus.yml",
		"docs/adr/0003-use-idempotency-and-outbox-style-events.md",
		"docs/adr/0006-use-railway-single-service-demo.md",
		"docs/architecture/testing-strategy.md",
		"docs/deployment/railway.md",
		"docs/production-readiness.md",
		"docs/runbooks/partner-api-contract-drift.md",
		"docs/runbooks/webhook-delivery-backlog.md",
		"docs/security/authorization-matrix.md",
		"docs/security/threat-model.md",
		"internal/httpapi/router_test.go",
		"internal/httpapi/middleware/idempotency_test.go",
		"internal/usecase/service_test.go",
		"internal/webhook/signer_test.go",
		"openapi.yaml",
	}

	for _, rel := range requiredDirs {
		info, err := os.Stat(absPath(t, rel))
		if err != nil {
			t.Fatalf("expected directory %s: %v", rel, err)
		}
		if !info.IsDir() {
			t.Fatalf("%s must be a directory", rel)
		}
	}

	for _, rel := range requiredFiles {
		info, err := os.Stat(absPath(t, rel))
		if err != nil {
			t.Fatalf("expected file %s: %v", rel, err)
		}
		if info.IsDir() {
			t.Fatalf("%s must be a file", rel)
		}
	}
}

func TestOpenAPIContractCoversPublicPartnerSurface(t *testing.T) {
	content := readFile(t, "openapi.yaml")
	required := []string{
		"openapi: 3.1.0",
		"/health/live:",
		"/health/ready:",
		"/metrics:",
		"/v1/accounts/{account_id}/balance:",
		"/v1/pix/transfers:",
		"/v1/payouts:",
		"/v1/refunds:",
		"/v1/webhooks/endpoints:",
		"/v1/audit-logs:",
		"bearerAuth: []",
		"Idempotency-Key",
		"Idempotency-Replayed",
		"insufficient_scope",
		"idempotency_conflict",
		"idempotency_original_failed",
		"idempotency_wait_timeout",
		"queued_webhook_deliveries",
	}

	for _, snippet := range required {
		if !strings.Contains(content, snippet) {
			t.Fatalf("openapi.yaml is missing %q", snippet)
		}
	}
}

func TestSecurityDocsAndRuntimeEvidenceStayAligned(t *testing.T) {
	matrix := readFile(t, "docs/security/authorization-matrix.md")
	requiredMatrixSnippets := []string{
		"`GET /v1/accounts/{id}/balance`",
		"`POST /v1/pix/transfers`",
		"`POST /v1/webhooks/endpoints`",
		"`GET /v1/audit-logs`",
		"`pix:write`",
		"`webhooks:write`",
		"Requires `Idempotency-Key`.",
		"Foreign accounts return 404.",
	}
	for _, snippet := range requiredMatrixSnippets {
		if !strings.Contains(matrix, snippet) {
			t.Fatalf("authorization matrix is missing %q", snippet)
		}
	}

	threatModel := readFile(t, "docs/security/threat-model.md")
	requiredThreatSnippets := []string{
		"Request replay",
		"Rate-limit abuse",
		"Webhook tampering",
		"Overbroad partner access",
		"structured log fields",
		"bankport_partner_api_rate_limit_exceeded_total",
		"bankport_partner_api_idempotency_conflicts_total",
	}
	for _, snippet := range requiredThreatSnippets {
		if !strings.Contains(threatModel, snippet) {
			t.Fatalf("threat model is missing %q", snippet)
		}
	}

	runtimeEvidence := []struct {
		file     string
		snippets []string
	}{
		{
			file: "internal/httpapi/router_test.go",
			snippets: []string{
				"TestHealthReadyAndMetrics",
				"TestRequiresAuthentication",
				"TestRejectsInsufficientScope",
				"TestTenantIsolationHidesForeignAccount",
				"TestIdempotentFinancialWriteReplaysCachedResponse",
				"TestIdempotencyConflict",
				"TestRateLimitExceeded",
				"TestWebhookRegistrationAndDeliveryQueue",
				"TestRequestIdentityRejectsUnsafeCallerSuppliedIDs",
				"TestMetricsUseRoutePatternForAccountIDs",
			},
		},
		{
			file: "internal/httpapi/middleware/idempotency_test.go",
			snippets: []string{
				"TestIdempotencyConcurrentSameKeyRunsHandlerOnce",
				"TestIdempotencyRejectsMalformedKey",
				"TestIdempotencyDoesNotCacheRequestTimeout",
			},
		},
		{
			file: "internal/webhook/signer_test.go",
			snippets: []string{
				"TestSignerCreatesVersionedHMACSignature",
				"TestSignerDerivesEndpointSpecificSignatures",
			},
		},
		{
			file: "internal/usecase/service_test.go",
			snippets: []string{
				"TestCreatePixTransferAuditsAndRecordsMetrics",
				"TestCreatePixTransferAuditsRejectedDomainError",
				"TestRegisterWebhookEndpointAuditsResult",
			},
		},
	}

	for _, item := range runtimeEvidence {
		content := readFile(t, item.file)
		for _, snippet := range item.snippets {
			if !strings.Contains(content, snippet) {
				t.Fatalf("%s is missing runtime evidence %q", item.file, snippet)
			}
		}
	}
}

func TestOperationalDocsAndRunbooksStayAligned(t *testing.T) {
	content := strings.Join([]string{
		readFile(t, "docs/architecture/testing-strategy.md"),
		readFile(t, "docs/production-readiness.md"),
		readFile(t, "docs/runbooks/webhook-delivery-backlog.md"),
		readFile(t, "docs/runbooks/partner-api-contract-drift.md"),
	}, "\n")

	required := []string{
		"Redocly validates OpenAPI contract drift.",
		"`/health/live` and `/health/ready` with dependency checks.",
		"Prometheus metrics",
		"Grafana dashboard",
		"Prometheus alerts",
		"runbooks",
		"A durable worker, retry queue, and",
		"dead-letter queue are next-phase production work.",
		"Verify idempotency behavior if the failing route is a financial write operation.",
		"Keep request and audit evidence intact; do not rewrite prior partner responses.",
		"Compose runtime smoke checks validate liveness, readiness, authenticated",
		"balance read, and Prometheus readiness.",
	}

	for _, snippet := range required {
		if !strings.Contains(content, snippet) {
			t.Fatalf("operational documentation is missing %q", snippet)
		}
	}
}

func TestWorkflowMatchesReviewerVerificationContract(t *testing.T) {
	content := readFile(t, ".github/workflows/ci.yml")
	required := []string{
		`GO_VERSION: "1.26.4"`,
		`NODE_VERSION: "22"`,
		`GOVULNCHECK_VERSION: "v1.3.0"`,
		`REDOCLY_VERSION: "2.31.5"`,
		"go mod tidy",
		"go test -race -coverpkg=./... -coverprofile=coverage.out ./...",
		"go build ./cmd/bankport-api ./cmd/bankportctl",
		"go tool cover -func=coverage.out",
		"govulncheck ./...",
		"npx --yes @redocly/cli@${REDOCLY_VERSION} lint openapi.yaml",
		"docker compose config",
		"docker build -t bankport-go-gin-partner-api:ci .",
	}

	for _, snippet := range required {
		if !strings.Contains(content, snippet) {
			t.Fatalf("ci workflow is missing %q", snippet)
		}
	}
}

func TestRecentCommitHistoryUsesConventionalCommits(t *testing.T) {
	cmd := exec.Command("git", "log", "--format=%s", "-n", "8")
	cmd.Dir = repoRoot(t)
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("git log failed: %v", err)
	}

	lines := strings.FieldsFunc(strings.TrimSpace(string(output)), func(r rune) bool {
		return r == '\n' || r == '\r'
	})
	if len(lines) == 0 {
		t.Fatal("expected recent commits")
	}

	pattern := regexp.MustCompile(`^(feat|fix|docs|test|chore|refactor|perf|ci|build|deploy)(\([^)]+\))?!?: .+`)
	for _, line := range lines {
		if !pattern.MatchString(line) {
			t.Fatalf("commit message does not follow Conventional Commits: %q", line)
		}
	}
}

func repoRoot(t testing.TB) string {
	t.Helper()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	for dir := cwd; ; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not locate repository root from go.mod")
		}
	}
}

func absPath(t testing.TB, rel string) string {
	t.Helper()
	return filepath.Join(repoRoot(t), rel)
}

func readFile(t testing.TB, rel string) string {
	t.Helper()

	content, err := os.ReadFile(absPath(t, rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(content)
}
