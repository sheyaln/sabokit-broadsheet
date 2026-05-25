package service

import (
	"context"
	"net/http"

	"github.com/sheyaln/sabokit-broadside/internal/domain"
	"github.com/sheyaln/sabokit-broadside/pkg/logger"
)

// TelemetryService is a no-op in Broadside. Upstream Notifuse posts
// workspace metrics to a hosted GCP function (`telemetry/`); the fork
// does not. The type is retained so app wiring in internal/app/app.go
// continues to compile, but every method returns immediately.
type TelemetryService struct{}

// TelemetryServiceConfig is kept for call-site compatibility. Fields
// are accepted and discarded.
type TelemetryServiceConfig struct {
	Enabled       bool
	APIEndpoint   string
	WorkspaceRepo domain.WorkspaceRepository
	TelemetryRepo domain.TelemetryRepository
	Logger        logger.Logger
	HTTPClient    *http.Client
}

func NewTelemetryService(_ TelemetryServiceConfig) *TelemetryService {
	return &TelemetryService{}
}

func (t *TelemetryService) SendMetricsForAllWorkspaces(_ context.Context) error {
	return nil
}

func (t *TelemetryService) StartDailyScheduler(_ context.Context) {}
