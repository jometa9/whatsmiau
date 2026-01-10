package controllers

import (
	"os"
	"path/filepath"
	"sync"

	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/verbeux-ai/whatsmiau/env"
	"github.com/verbeux-ai/whatsmiau/lib/whatsmiau"
	"github.com/verbeux-ai/whatsmiau/services"
	"github.com/verbeux-ai/whatsmiau/utils"
	"go.mau.fi/whatsmeow"
	"go.uber.org/zap"
)

var (
	swaggerYAML     []byte
	swaggerYAMLLoad sync.Once
)

func Root(ctx echo.Context) error {
	res, err := whatsmeow.GetLatestVersion(ctx.Request().Context(), &http.Client{Timeout: 5 * time.Second})
	if err != nil {
		return utils.HTTPFail(ctx, http.StatusBadRequest, err, "failed to fetch latest version")
	}

	jsonData := map[string]any{
		"status":             200,
		"message":            "Welcome to the PingQR API, it is working!",
		"version":            "0.3.2",
		"clientName":         "pingqr",
		"whatsappWebVersion": res.String(),
	}

	return ctx.JSON(http.StatusOK, jsonData)
}

func Health(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, map[string]any{
		"success": true,
	})
}

func HardReset(ctx echo.Context) error {
	ctxReq := ctx.Request().Context()
	repo := services.GetInstanceRepository()
	whatsmiauInstance := whatsmiau.Get()

	// Get all instances
	instances, err := repo.List(ctxReq, "")
	if err != nil {
		zap.L().Error("failed to list instances for hard reset", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   "failed to list instances",
			"message": err.Error(),
		})
	}

	// Logout and delete all instances
	for _, instance := range instances {
		if whatsmiauInstance != nil {
			if err := whatsmiauInstance.Logout(ctxReq, instance.ID); err != nil {
				zap.L().Warn("failed to logout instance during hard reset", zap.String("instance", instance.ID), zap.Error(err))
			}
		}
		if err := repo.Delete(ctxReq, instance.ID); err != nil {
			zap.L().Warn("failed to delete instance during hard reset", zap.String("instance", instance.ID), zap.Error(err))
		}
	}

	// Clear all devices from storage
	container := services.SQLStore()
	if container != nil {
		devices, err := container.GetAllDevices(ctxReq)
		if err == nil {
			for _, device := range devices {
				if err := container.DeleteDevice(ctxReq, device); err != nil {
					zap.L().Warn("failed to delete device during hard reset", zap.Error(err))
				}
			}
		}
	}

	if env.Env.DebugMode {
		zap.L().Info("hard reset completed", zap.Int("instancesDeleted", len(instances)))
	}

	return ctx.JSON(http.StatusOK, map[string]any{
		"success": true,
	})
}

func loadSwaggerYAML() {
	swaggerYAMLLoad.Do(func() {
		// Try to find swagger.yaml relative to the binary
		execPath, err := os.Executable()
		if err == nil {
			execDir := filepath.Dir(execPath)
			swaggerPath := filepath.Join(execDir, "swagger.yaml")
			if data, err := os.ReadFile(swaggerPath); err == nil {
				swaggerYAML = data
				return
			}
		}
		// Fallback: try current working directory
		if data, err := os.ReadFile("swagger.yaml"); err == nil {
			swaggerYAML = data
			return
		}
		// Fallback: try relative to source
		if data, err := os.ReadFile("../../swagger.yaml"); err == nil {
			swaggerYAML = data
			return
		}
		swaggerYAML = []byte("# Swagger file not found")
	})
}

func SwaggerYAML(ctx echo.Context) error {
	loadSwaggerYAML()
	ctx.Response().Header().Set("Content-Type", "application/yaml")
	return ctx.Blob(http.StatusOK, "application/yaml", swaggerYAML)
}

func SwaggerUI(ctx echo.Context) error {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>PingQR API - Swagger UI</title>
  <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.10.3/swagger-ui.css" />
  <style>
    html {
      box-sizing: border-box;
      overflow: -moz-scrollbars-vertical;
      overflow-y: scroll;
    }
    *, *:before, *:after {
      box-sizing: inherit;
    }
    body {
      margin:0;
      background: #fafafa;
    }
    .models {
      display: none !important;
    }
    .model-box {
      display: none !important;
    }
    .model-container {
      display: none !important;
    }
    [data-name="models"] {
      display: none !important;
    }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5.10.3/swagger-ui-bundle.js"></script>
  <script src="https://unpkg.com/swagger-ui-dist@5.10.3/swagger-ui-standalone-preset.js"></script>
  <script>
    window.onload = function() {
      const ui = SwaggerUIBundle({
        url: "/v1/swagger.yaml",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout",
        defaultModelsExpandDepth: -1,
        docExpansion: "list"
      });
    };
  </script>
</body>
</html>`
	return ctx.HTML(http.StatusOK, html)
}
