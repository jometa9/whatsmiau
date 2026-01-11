package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/verbeux-ai/whatsmiau/server/controllers"
)

func Root(group *echo.Group) {
	group.GET("", controllers.Root)
	group.GET("/health", controllers.Health)
	group.POST("/reset", controllers.HardReset)
	group.GET("/swagger", controllers.SwaggerUI)
	group.GET("/swagger.yaml", controllers.SwaggerYAML)
}
