// Package config содержит утилиты для загрузки конфигурации из переменных окружения.
package config

import (
	"context"
	"fmt"
	"os"

	"github.com/flexer2006/case-person-enrichment-go/pkg/logger"
	"github.com/ilyakaznacheev/cleanenv"
	"go.uber.org/zap"
)

// LoggableConfig интерфейс для типов конфигурации, которые могут предоставить поля для логирования.
type LoggableConfig interface {
	LogFields() []zap.Field
}

// LoadOptions содержит опции для загрузки конфигурации.
type LoadOptions struct {
	ConfigPath string
}

// Load загружает конфигурацию для любого типа T.
// Если указан путь к файлу конфигурации, сначала загружается из него.
// Затем загружаются переменные окружения, которые могут переопределить значения из файла.
// Если тип T реализует интерфейс LoggableConfig, его поля будут залогированы.
func Load[T any](ctx context.Context, opts ...LoadOptions) (*T, error) {
	log := logger.GetContextLogger(ctx)
	log.Info("loading configuration")

	var cfg T

	var options LoadOptions
	if len(opts) > 0 {
		options = opts[0]
	}

	if options.ConfigPath != "" {
		if _, err := os.Stat(options.ConfigPath); err == nil {
			if err := cleanenv.ReadConfig(options.ConfigPath, &cfg); err != nil {
				log.Error("failed to load configuration", zap.Error(err), zap.String("path", options.ConfigPath))
				return nil, fmt.Errorf("%s from file %s: %w", "failed to load configuration", options.ConfigPath, err)
			}
		}
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Error("failed to load configuration", zap.Error(err))
		return nil, fmt.Errorf("%s from environment: %w", "failed to load configuration", err)
	}

	if loggable, ok := any(&cfg).(LoggableConfig); ok {
		log.Info("configuration loaded successfully", loggable.LogFields()...)
	} else {
		log.Info("configuration loaded successfully")
	}

	return &cfg, nil
}
