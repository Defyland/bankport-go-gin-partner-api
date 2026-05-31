package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/allanflavio/bankport-go-gin-partner-api/internal/config"
	"github.com/allanflavio/bankport-go-gin-partner-api/internal/domain"
	"github.com/allanflavio/bankport-go-gin-partner-api/internal/httpapi/middleware"
	"github.com/allanflavio/bankport-go-gin-partner-api/internal/observability"
	"github.com/allanflavio/bankport-go-gin-partner-api/internal/usecase"
	"github.com/allanflavio/bankport-go-gin-partner-api/internal/webhook"
)

type Repository interface {
	middleware.Authenticator
	usecase.AccountReader
	usecase.FinancialCommandStore
	usecase.WebhookStore
	usecase.PlatformReader
}

type Dependencies struct {
	Config     config.Config
	Logger     *slog.Logger
	Repository Repository
	Metrics    *observability.Metrics
}

type Server struct {
	cfg              config.Config
	logger           *slog.Logger
	usecases         usecase.Service
	idempotencyStore *middleware.IdempotencyStore
	rateLimiter      *middleware.RateLimiter
}

func NewRouter(deps Dependencies) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	logger := deps.Logger
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}

	signer := webhook.NewSigner(deps.Config.WebhookSigningKey)
	server := &Server{
		cfg:    deps.Config,
		logger: logger,
		usecases: usecase.New(usecase.Dependencies{
			Accounts:  deps.Repository,
			Financial: deps.Repository,
			Webhooks:  deps.Repository,
			Platform:  deps.Repository,
			Metrics:   metricsRecorder{metrics: deps.Metrics},
			SignEvent: signer.SignEventForEndpoint,
		}),
		idempotencyStore: middleware.NewIdempotencyStoreWithTTL(deps.Config.IdempotencyTTL),
		rateLimiter:      middleware.NewRateLimiter(),
	}

	router := gin.New()
	router.Use(
		gin.Recovery(),
		middleware.RequestIdentity(),
		middleware.RequestBodyLimit(deps.Config.MaxRequestBodyBytes),
		middleware.Timeout(deps.Config.RequestTimeout),
		middleware.Tracing(deps.Config.ServiceName),
		middleware.StructuredLogger(logger),
		middleware.Metrics(deps.Metrics),
	)

	router.GET("/health/live", server.live)
	router.GET("/health/ready", server.ready)
	router.GET("/metrics", gin.WrapH(promhttp.HandlerFor(deps.Metrics.Registry, promhttp.HandlerOpts{})))
	if deps.Config.PprofEnabled {
		registerPprof(router)
	}

	v1 := router.Group("/v1")
	v1.Use(middleware.Authenticate(deps.Repository, deps.Metrics), middleware.RateLimit(server.rateLimiter, deps.Metrics))
	v1.GET("/accounts/:account_id/balance", middleware.RequireScopes(domain.ScopeAccountsRead), server.getBalance)
	v1.GET("/accounts/:account_id/statements", middleware.RequireScopes(domain.ScopeAccountsRead), server.listStatements)
	v1.POST("/pix/transfers", middleware.RequireScopes(domain.ScopePixWrite), middleware.Idempotency(server.idempotencyStore, deps.Metrics), server.createPixTransfer)
	v1.POST("/payouts", middleware.RequireScopes(domain.ScopePayoutsWrite), middleware.Idempotency(server.idempotencyStore, deps.Metrics), server.createPayout)
	v1.POST("/refunds", middleware.RequireScopes(domain.ScopeRefundsWrite), middleware.Idempotency(server.idempotencyStore, deps.Metrics), server.createRefund)
	v1.POST("/webhooks/endpoints", middleware.RequireScopes(domain.ScopeWebhooksWrite), server.registerWebhookEndpoint)
	v1.GET("/audit-logs", middleware.RequireScopes(domain.ScopeAuditRead), server.listAuditLogs)
	v1.GET("/sandbox/scenarios", middleware.RequireScopes(domain.ScopeSandboxRead), server.listSandboxScenarios)

	return router
}

func registerPprof(router *gin.Engine) {
	pprofRoutes := router.Group("/debug/pprof")
	pprofRoutes.GET("/", gin.WrapF(pprof.Index))
	pprofRoutes.GET("/cmdline", gin.WrapF(pprof.Cmdline))
	pprofRoutes.GET("/profile", gin.WrapF(pprof.Profile))
	pprofRoutes.GET("/symbol", gin.WrapF(pprof.Symbol))
	pprofRoutes.POST("/symbol", gin.WrapF(pprof.Symbol))
	pprofRoutes.GET("/trace", gin.WrapF(pprof.Trace))
	pprofRoutes.GET("/:profile", gin.WrapF(pprof.Index))
}

func (s *Server) live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":         "ok",
		"service":        s.cfg.ServiceName,
		"request_id":     middleware.RequestID(c),
		"correlation_id": middleware.CorrelationID(c),
	})
}

func (s *Server) ready(c *gin.Context) {
	if !s.cfg.ReadinessEnabled {
		middleware.Abort(c, http.StatusServiceUnavailable, "not_ready", "Readiness is disabled by configuration.", nil)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
		"checks": gin.H{
			"repository":  "ready",
			"rate_limit":  "ready",
			"idempotency": "ready",
		},
		"request_id":     middleware.RequestID(c),
		"correlation_id": middleware.CorrelationID(c),
	})
}

func (s *Server) getBalance(c *gin.Context) {
	partner, _ := middleware.Partner(c)
	account, err := s.usecases.GetBalance(c.Request.Context(), partner, c.Param("account_id"))
	if err != nil {
		s.respondDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, success(account, c))
}

func (s *Server) listStatements(c *gin.Context) {
	partner, _ := middleware.Partner(c)
	entries, err := s.usecases.ListStatements(c.Request.Context(), partner, c.Param("account_id"))
	if err != nil {
		s.respondDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, success(entries, c))
}

func (s *Server) createPixTransfer(c *gin.Context) {
	partner, _ := middleware.Partner(c)
	var request domain.PixTransferRequest
	if !bindJSON(c, &request) {
		return
	}

	result, err := s.usecases.CreatePixTransfer(c.Request.Context(), partner, request, middleware.CorrelationID(c), middleware.RequestID(c))
	if err != nil {
		s.respondDomainError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, success(gin.H{
		"transfer":                  result.Transfer,
		"queued_webhook_deliveries": result.QueuedDeliveries,
	}, c))
}

func (s *Server) createPayout(c *gin.Context) {
	partner, _ := middleware.Partner(c)
	var request domain.PayoutRequest
	if !bindJSON(c, &request) {
		return
	}

	result, err := s.usecases.CreatePayout(c.Request.Context(), partner, request, middleware.CorrelationID(c), middleware.RequestID(c))
	if err != nil {
		s.respondDomainError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, success(gin.H{
		"payout":                    result.Payout,
		"queued_webhook_deliveries": result.QueuedDeliveries,
	}, c))
}

func (s *Server) createRefund(c *gin.Context) {
	partner, _ := middleware.Partner(c)
	var request domain.RefundRequest
	if !bindJSON(c, &request) {
		return
	}

	result, err := s.usecases.CreateRefund(c.Request.Context(), partner, request, middleware.CorrelationID(c), middleware.RequestID(c))
	if err != nil {
		s.respondDomainError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, success(gin.H{
		"refund":                    result.Refund,
		"queued_webhook_deliveries": result.QueuedDeliveries,
	}, c))
}

func (s *Server) registerWebhookEndpoint(c *gin.Context) {
	partner, _ := middleware.Partner(c)
	var request domain.WebhookEndpointRequest
	if !bindJSON(c, &request) {
		return
	}

	endpoint, err := s.usecases.RegisterWebhookEndpoint(c.Request.Context(), partner, request, middleware.CorrelationID(c), middleware.RequestID(c))
	if err != nil {
		s.respondDomainError(c, err)
		return
	}
	c.JSON(http.StatusCreated, success(endpoint, c))
}

func (s *Server) listAuditLogs(c *gin.Context) {
	partner, _ := middleware.Partner(c)
	c.JSON(http.StatusOK, success(s.usecases.ListAuditLogs(c.Request.Context(), partner), c))
}

func (s *Server) listSandboxScenarios(c *gin.Context) {
	c.JSON(http.StatusOK, success(s.usecases.ListSandboxScenarios(), c))
}

func (s *Server) respondDomainError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrValidation):
		middleware.Abort(c, http.StatusBadRequest, "validation_failed", err.Error(), nil)
	case errors.Is(err, domain.ErrAccountNotFound):
		middleware.Abort(c, http.StatusNotFound, "account_not_found", "The account was not found for this partner.", nil)
	case errors.Is(err, domain.ErrInsufficientFunds):
		middleware.Abort(c, http.StatusUnprocessableEntity, "insufficient_funds", "The account does not have enough available balance.", nil)
	case errors.Is(err, domain.ErrOriginalTxnNotFound):
		middleware.Abort(c, http.StatusNotFound, "original_transaction_not_found", "The original transaction was not found for this partner and account.", nil)
	case errors.Is(err, domain.ErrRefundExceedsOriginal):
		middleware.Abort(c, http.StatusUnprocessableEntity, "refund_exceeds_original", "The refund amount exceeds the original transaction amount.", nil)
	case errors.Is(err, context.DeadlineExceeded):
		middleware.Abort(c, http.StatusGatewayTimeout, "request_timeout", "The request exceeded the configured timeout.", nil)
	case errors.Is(err, context.Canceled):
		middleware.Abort(c, http.StatusRequestTimeout, "request_canceled", "The request was canceled before it could complete.", nil)
	default:
		s.logger.Error("domain_error_unmapped", "error", err, "request_id", middleware.RequestID(c))
		middleware.Abort(c, http.StatusInternalServerError, "internal_error", "An internal error occurred.", nil)
	}
}

func bindJSON(c *gin.Context, target any) bool {
	if err := c.ShouldBindJSON(target); err != nil {
		middleware.Abort(c, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON matching the endpoint schema.", map[string]any{
			"error": err.Error(),
		})
		return false
	}
	return true
}

func success(data any, c *gin.Context) gin.H {
	return gin.H{
		"data":           data,
		"request_id":     middleware.RequestID(c),
		"correlation_id": middleware.CorrelationID(c),
	}
}

type metricsRecorder struct {
	metrics *observability.Metrics
}

func (r metricsRecorder) RecordFinancialCommand(command, outcome string) {
	if r.metrics != nil {
		r.metrics.FinancialCommands.WithLabelValues(command, outcome).Inc()
	}
}

func (r metricsRecorder) RecordWebhookDeliveries(eventType, status string, count int) {
	if count > 0 && r.metrics != nil {
		r.metrics.WebhookDeliveries.WithLabelValues(eventType, status).Add(float64(count))
	}
}
