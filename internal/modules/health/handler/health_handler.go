package handler

import (
	"context"
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/internal/util/response"
)

// HealthHandler handles health check and welcome endpoints.
type HealthHandler struct {
	appVersion string
	startTime  time.Time
}

// NewHealthHandler initializes and returns a new HealthHandler instance.
func NewHealthHandler(appVersion string) *HealthHandler {
	return &HealthHandler{
		appVersion: appVersion,
		startTime:  time.Now(),
	}
}

// WelcomeResponse represents the welcome message response.
type WelcomeResponse struct {
	Message string `json:"message"`
}

// HealthCheckResponse represents the health check response.
type HealthCheckResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
	Uptime  string `json:"uptime"`
}

// HandleWelcome returns a welcome message for the root endpoint.
func (h *HealthHandler) HandleWelcome(ctx context.Context, input *struct{}) (*response.Response, error) {
	data := &WelcomeResponse{
		Message: "Welcome to API CRM Handai",
	}

	return response.BuildSuccess(data, "Welcome"), nil
}

// HandleHealthCheck returns the health status of the API.
func (h *HealthHandler) HandleHealthCheck(ctx context.Context, input *struct{}) (*response.Response, error) {
	uptime := time.Since(h.startTime)

	data := &HealthCheckResponse{
		Status:  "healthy",
		Version: h.appVersion,
		Uptime:  uptime.String(),
	}

	return response.BuildSuccess(data, "API is running"), nil
}
