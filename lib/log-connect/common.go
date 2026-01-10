package log_connect

import (
	"github.com/verbeux-ai/whatsmiau/env"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func StartLogger() error {
	if env.Env.GCLEnabled {
		c, err := startGCL()
		if err != nil {
			return err
		}
		zap.ReplaceGlobals(zap.New(*c))
		return nil
	}

	var logger *zap.Logger
	var err error

	if env.Env.DebugMode {
		logger, err = zap.NewDevelopment()
	} else {
		// En modo producción, solo mostrar logs Fatal (errores críticos)
		// Todos los demás logs (Info, Debug, Warn, Error) se suprimen
		config := zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zapcore.FatalLevel)
		logger, err = config.Build()
	}

	if err != nil {
		return err
	}

	zap.ReplaceGlobals(logger)

	return nil
}
