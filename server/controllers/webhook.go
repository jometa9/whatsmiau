package controllers

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/verbeux-ai/whatsmiau/env"
	"go.uber.org/zap"
)

type Webhook struct{}

func NewWebhook() *Webhook {
	return &Webhook{}
}

type UpdateHeartbeatWebhookRequest struct {
	URL string `json:"url" validate:"required,url"`
}

type UpdateHeartbeatWebhookResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

type GetHeartbeatWebhookResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
	URL     string `json:"url,omitempty"`
}

func (s *Webhook) UpdateHeartbeat(ctx echo.Context) error {
	var request UpdateHeartbeatWebhookRequest
	if err := ctx.Bind(&request); err != nil {
		return ctx.JSON(http.StatusBadRequest, UpdateHeartbeatWebhookResponse{
			Success: false,
			Error:   "failed to bind request body",
			Message: err.Error(),
		})
	}

	if err := validator.New().Struct(&request); err != nil {
		return ctx.JSON(http.StatusBadRequest, UpdateHeartbeatWebhookResponse{
			Success: false,
			Error:   "invalid request body",
			Message: err.Error(),
		})
	}

	env.SetWebhookURL(request.URL)
	if env.Env.DebugMode {
		zap.L().Info("heartbeat webhook URL updated", zap.String("url", request.URL))
	}

	return ctx.JSON(http.StatusOK, UpdateHeartbeatWebhookResponse{
		Success: true,
	})
}

func (s *Webhook) GetHeartbeat(ctx echo.Context) error {
	url := env.GetWebhookURL()
	return ctx.JSON(http.StatusOK, GetHeartbeatWebhookResponse{
		Success: true,
		URL:     url,
	})
}

