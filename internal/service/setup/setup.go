// Package setup содержит основную конфигурацию для приложения.
package setup

import (
	"github.com/flexer2006/case-person-enrichment-go/internal/service/setup/data"
	"github.com/flexer2006/case-person-enrichment-go/internal/service/setup/graceful"
	"github.com/flexer2006/case-person-enrichment-go/internal/service/setup/logs"
	"github.com/flexer2006/case-person-enrichment-go/internal/service/setup/migration"
	"go.uber.org/zap"
)

// Config представляет основную конфигурацию приложения.
type Config struct {
	Logger     logs.Config         `env-prefix:"LOGGER_"`
	Postgres   data.PostgresConfig `env-prefix:""`
	Migrations migration.Config    `env-prefix:""`
	Graceful   graceful.Config
}

// LogFields реализует интерфейс LoggableConfig и возвращает поля конфигурации
// приложения для логирования.
func (c *Config) LogFields() []zap.Field {
	return []zap.Field{
		zap.String("graceful_shutdown_timeout", c.Graceful.ShutdownTimeout),
	}
}
