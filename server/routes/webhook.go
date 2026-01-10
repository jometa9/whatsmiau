package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/verbeux-ai/whatsmiau/server/controllers"
)

func Webhook(group *echo.Group) {
	controller := controllers.NewWebhook()
	group.PUT("/heartbeat", controller.UpdateHeartbeat)
	group.GET("/heartbeat", controller.GetHeartbeat)
}

