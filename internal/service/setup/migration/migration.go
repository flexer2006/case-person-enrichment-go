// Package migration содержит настройки для миграций базы данных.
package migration

import (
	"go.uber.org/zap"
)

// Config содержит настройки для миграций базы данных.
type Config struct {
	Path string `env:"MIGRATIONS_DIR" env-default:"./migrations"`
}

// LogFields реализует интерфейс LoggableConfig для Config.
func (c *Config) LogFields() []zap.Field {
	return []zap.Field{
		zap.String("path", c.Path),
	}
}
