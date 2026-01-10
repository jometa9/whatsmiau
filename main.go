package main

import (
	"log"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/verbeux-ai/whatsmiau/env"
	log_connect "github.com/verbeux-ai/whatsmiau/lib/log-connect"
	"github.com/verbeux-ai/whatsmiau/lib/whatsmiau"
	"github.com/verbeux-ai/whatsmiau/server/routes"
	"github.com/verbeux-ai/whatsmiau/services"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"golang.org/x/net/http2"
)

func main() {
	if err := env.Load(); err != nil {
		panic(err)
	}

	if err := log_connect.StartLogger(); err != nil {
		log.Fatalln(err)
	}

	ctx, c := context.WithTimeout(context.Background(), 10*time.Second)
	defer c()
	whatsmiau.LoadMiau(ctx, services.SQLStore())

	// Set instance repository for monitoring (use the same repository as routes)
	repo := services.GetInstanceRepository()
	services.SetInstanceRepository(repo)

	// Set connected instance checker to avoid import cycle
	// This function checks if an instance is connected using whatsmiau
	services.SetConnectedInstanceChecker(func(instanceID string) bool {
		whatsmiauInstance := whatsmiau.Get()
		if whatsmiauInstance == nil {
			return false
		}
		status, err := whatsmiauInstance.Status(instanceID)
		return err == nil && status == whatsmiau.Connected
	})

	// Start system monitoring service
	services.StartMonitor()

	app := echo.New()
	
	// Si NO estamos en modo debug, deshabilitar el banner y el logger de Echo
	if !env.Env.DebugMode {
		app.HideBanner = true
		app.Logger.SetLevel(0) // Deshabilitar todos los logs de Echo
	}
	
	app.Pre(middleware.Recover())
	app.Pre(middleware.RemoveTrailingSlash())
	app.Pre(middleware.CORS())

	routes.Load(app)

	port := ":" + env.Env.Port
	
	if env.Env.DebugMode {
		zap.L().Info("starting server...", zap.String("port", port))
	}

	s := &http2.Server{}
	if err := app.StartH2CServer(port, s); err != nil {
		zap.L().Fatal("failed to start server", zap.Error(err))
	}
}
