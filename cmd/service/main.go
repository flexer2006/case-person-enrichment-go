// Package main реализует точку входа для сервиса обогащения данных о персонах.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/flexer2006/case-person-enrichment-go/internal/service/app"
	"github.com/flexer2006/case-person-enrichment-go/internal/service/setup"
	"github.com/flexer2006/case-person-enrichment-go/pkg/config"
	"github.com/flexer2006/case-person-enrichment-go/pkg/database"
	"github.com/flexer2006/case-person-enrichment-go/pkg/database/migrate"
	"github.com/flexer2006/case-person-enrichment-go/pkg/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	initialLogger := logger.NewConsole(logger.InfoLevel, true)
	logger.SetGlobal(initialLogger)

	ctx := context.Background()
	var exitCode int

	func() {
		defer func() {
			if err := initialLogger.Sync(); err != nil {
				errMsg := err.Error()
				if strings.Contains(errMsg, "sync /dev/stderr: invalid argument") ||
					strings.Contains(errMsg, "sync /dev/stdout: invalid argument") {
					return
				}
				if n, writeErr := fmt.Fprintf(os.Stderr, "failed to sync logger: %v\n", err); writeErr != nil {
					panic(fmt.Sprintf("failed to write error message to stderr: %v", writeErr))
				} else if n == 0 {
					panic("failed to write error message to stderr: zero bytes written")
				}
			}
		}()

		cfg, err := config.Load[setup.Config](ctx, config.LoadOptions{
			ConfigPath: "./deploy/.env",
		})
		if err != nil {
			logger.Error(ctx, "failed to load configuration", zap.Error(err))
			exitCode = 1
			return
		}

		var finalLogger *logger.Logger
		switch cfg.Logger.Model {
		case "development":
			finalLogger, err = logger.NewDevelopment()
		case "production":
			finalLogger, err = logger.NewProduction()
		default:
			logger.Warn(ctx, "unknown logger model, using development", zap.String("model", cfg.Logger.Model))
			finalLogger, err = logger.NewDevelopment()
		}

		if err != nil {
			logger.Error(ctx, "failed to initialize logger with config", zap.Error(err))
			exitCode = 1
			return
		}

		logger.SetGlobal(finalLogger)

		shutdownTimeout, err := time.ParseDuration(cfg.Graceful.ShutdownTimeout)
		if err != nil {
			logger.Error(ctx, "invalid graceful shutdown timeout", zap.Error(err))
			shutdownTimeout = 5 * time.Second
		}

		// Конфигурация базы данных
		dbConfig := database.Config{
			Postgres: cfg.Postgres.ToConfig(),
			Migrate: migrate.Config{
				Path: cfg.Migrations.Path,
			},
			ApplyMigrations: true, // Автоматически применяем миграции при запуске
		}

		logger.Info(ctx, "initializing database")
		data, err := database.New(ctx, dbConfig)
		if err != nil {
			logger.Error(ctx, "failed to initialize database", zap.Error(err))
			exitCode = 1
			return
		}

		if err := data.Ping(ctx); err != nil {
			logger.Error(ctx, "database ping failed", zap.Error(err))
			exitCode = 1
			return
		}

		version, dirty, err := data.GetMigrationVersion(ctx)
		if err != nil {
			logger.Warn(ctx, "failed to get migration version", zap.Error(err))
		} else {
			if dirty {
				logger.Warn(ctx, "database has dirty migration", zap.Uint("version", version))
			} else {
				logger.Info(ctx, "current migration version", zap.Uint("version", version))
			}
		}

		logger.Info(ctx, "database initialized successfully")

		application, err := app.NewApplication(ctx, cfg)
		if err != nil {
			logger.Error(ctx, "failed to initialize application", zap.Error(err))
			exitCode = 1
			return
		}

		// Create a context with cancellation for graceful shutdown
		appCtx, appCancel := context.WithCancel(ctx)
		defer appCancel()

		// Set up signal handling
		stopCh := make(chan os.Signal, 1)
		signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

		// Start the application
		errChan := make(chan error, 1)
		go func() {
			errChan <- application.Start(appCtx)
		}()

		logger.Info(ctx, "service started",
			zap.String("environment", cfg.Logger.Model),
			zap.String("log_level", cfg.Logger.Level),
			zap.String("startup_time", time.Now().Format(time.RFC3339)),
			zap.Object("server_config", zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
				for _, field := range cfg.Server.LogFields() {
					field.AddTo(enc)
				}
				return nil
			})),
		)

		// Wait for either an application error or shutdown signal
		var sig os.Signal
		select {
		case err := <-errChan:
			if err != nil {
				logger.Error(ctx, "application stopped with error", zap.Error(err))
				exitCode = 1
			}
		case sig = <-stopCh:
			logger.Info(ctx, "received shutdown signal", zap.String("signal", sig.String()))
		}

		// Create a context with timeout for graceful shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer shutdownCancel()

		// Cancel the application context to signal all components to stop
		appCancel()

		// Stop the application
		if err := application.Stop(shutdownCtx); err != nil {
			logger.Error(ctx, "error stopping application", zap.Error(err))
			exitCode = 1
		}

		// Close the database connection
		logger.Info(ctx, "closing database connection")
		data.Close(shutdownCtx)

		logger.Info(ctx, "service shutdown complete")
	}()

	if exitCode != 0 {
		os.Exit(exitCode)
	}
}
