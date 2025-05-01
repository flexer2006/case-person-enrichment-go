// Package graceful содержит настройки для корректного завершения работы приложения.
package graceful

import (
	"go.uber.org/zap"
)

// Config содержит настройки для graceful shutdown.
type Config struct {
	ShutdownTimeout string `env:"GRACEFUL_SHUTDOWN_TIMEOUT" env-default:"5s"`
}

// LogFields реализует интерфейс LoggableConfig для Config.
func (c *Config) LogFields() []zap.Field {
	return []zap.Field{
		zap.String("shutdown_timeout", c.ShutdownTimeout),
	}
}
