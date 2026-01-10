package controllers

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/verbeux-ai/whatsmiau/env"
	"github.com/verbeux-ai/whatsmiau/utils"
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
	URL string `json:"url"`
}

type GetHeartbeatWebhookResponse struct {
	URL string `json:"url"`
}

func (s *Webhook) UpdateHeartbeat(ctx echo.Context) error {
	var request UpdateHeartbeatWebhookRequest
	if err := ctx.Bind(&request); err != nil {
		return utils.HTTPFail(ctx, http.StatusUnprocessableEntity, err, "failed to bind request body")
	}

	if err := validator.New().Struct(&request); err != nil {
		return utils.HTTPFail(ctx, http.StatusBadRequest, err, "invalid request body")
	}

	env.SetWebhookURL(request.URL)
	if env.Env.DebugMode {
		zap.L().Info("heartbeat webhook URL updated", zap.String("url", request.URL))
	}

	return ctx.JSON(http.StatusOK, UpdateHeartbeatWebhookResponse{
		URL: request.URL,
	})
}

func (s *Webhook) GetHeartbeat(ctx echo.Context) error {
	url := env.GetWebhookURL()
	return ctx.JSON(http.StatusOK, GetHeartbeatWebhookResponse{
		URL: url,
	})
}

