// Package server содержит конфигурацию HTTP-сервера.
package server

import (
	"time"

	"go.uber.org/zap"
)

// Config представляет конфигурацию HTTP-сервера.
type Config struct {
	Host         string        `env:"HTTP_HOST" env-default:"0.0.0.0"`
	Port         int           `env:"HTTP_PORT" env-default:"8080"`
	ReadTimeout  time.Duration `env:"HTTP_READ_TIMEOUT" env-default:"5s"`
	WriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT" env-default:"10s"`
}

// LogFields реализует интерфейс для логирования и возвращает поля конфигурации
// HTTP-сервера в формате zap.Field.
func (c *Config) LogFields() []zap.Field {
	return []zap.Field{
		zap.String("host", c.Host),
		zap.Int("port", c.Port),
		zap.Duration("read_timeout", c.ReadTimeout),
		zap.Duration("write_timeout", c.WriteTimeout),
	}
}
