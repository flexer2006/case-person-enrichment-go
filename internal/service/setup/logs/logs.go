// Package logs содержит настройки конфигурации логгера для приложения.
package logs

import (
	"go.uber.org/zap"
)

// Config содержит настройки для конфигурации логгера.
type Config struct {
	Level        string `env:"LOGGER_LEVEL" env-default:"info"`
	Format       string `env:"LOGGER_FORMAT" env-default:"json"`
	Output       string `env:"LOGGER_OUTPUT" env-default:"stdout"`
	TimeEncoding string `env:"LOGGER_TIME_ENCODING" env-default:"iso8601"`
	Caller       bool   `env:"LOGGER_CALLER" env-default:"true"`
	Stacktrace   bool   `env:"LOGGER_STACKTRACE" env-default:"true"`
	Model        string `env:"LOGGER_MODEL" env-default:"development"`
}

// LogFields реализует интерфейс LoggableConfig и возвращает поля конфигурации
// логгера для логирования.
func (c *Config) LogFields() []zap.Field {
	return []zap.Field{
		zap.String("level", c.Level),
		zap.String("format", c.Format),
		zap.String("output", c.Output),
		zap.String("time_encoding", c.TimeEncoding),
		zap.Bool("caller", c.Caller),
		zap.Bool("stacktrace", c.Stacktrace),
		zap.String("model", c.Model),
	}
}
