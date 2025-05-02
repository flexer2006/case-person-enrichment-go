// Package server предоставляет HTTP-сервер приложения с использованием Fiber.
package server

import (
	"context"
	"fmt"

	"github.com/flexer2006/case-person-enrichment-go/internal/service/adapters/server/routes"
	"github.com/flexer2006/case-person-enrichment-go/internal/service/ports/api"
	"github.com/flexer2006/case-person-enrichment-go/internal/service/ports/repo"
	"github.com/flexer2006/case-person-enrichment-go/internal/service/setup/server"
	"github.com/flexer2006/case-person-enrichment-go/pkg/logger"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

	_ "github.com/flexer2006/case-person-enrichment-go/docs/swagger"
)

// @title Person Enrichment API
// @version 1.0
// @description API for managing and enriching person data with external services
// @contact.name API Support
// @contact.email andrewgo1133official@gmail.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @BasePath /api/v1

// Server представляет HTTP-сервер приложения.
type Server struct {
	app    *fiber.App
	config server.Config
}

// New создает новый экземпляр HTTP-сервера.
func New(config server.Config, api api.API, repositories repo.Repositories) *Server {
	app := fiber.New(fiber.Config{
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		AppName:      "Person Enrichment Service",
	})

	app.Get("/swagger", func(c fiber.Ctx) error {
		r := c.Redirect()
		r.Status(fiber.StatusFound)
		return r.To("/swagger/swagger.html")
	})

	app.Get("/swagger/swagger.html", func(c fiber.Ctx) error {
		return c.SendFile("./docs/swagger/swagger.html")
	})

	app.Get("/swagger/swagger.json", func(c fiber.Ctx) error {
		return c.SendFile("./docs/swagger/swagger.json")
	})

	routes.Setup(app, api, repositories)

	return &Server{
		app:    app,
		config: config,
	}
}

// Start запускает HTTP-сервер.
func (s *Server) Start(ctx context.Context) error {
	address := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	logger.Info(ctx, "starting HTTP server", zap.String("address", address))

	go func() {
		if err := s.app.Listen(address); err != nil {
			logger.Error(ctx, "failed to start HTTP server", zap.Error(err))
		}
	}()

	<-ctx.Done()
	return nil
}

// Stop останавливает HTTP-сервер.
func (s *Server) Stop(ctx context.Context) error {
	logger.Info(ctx, "stopping HTTP server")

	if err := s.app.ShutdownWithContext(ctx); err != nil {
		logger.Error(ctx, "failed to shutdown HTTP server gracefully", zap.Error(err))
		return fmt.Errorf("failed to shutdown HTTP server gracefully: %w", err)
	}

	return nil
}

// GetConfig возвращает конфигурацию сервера.
func (s *Server) GetConfig() server.Config {
	return s.config
}
