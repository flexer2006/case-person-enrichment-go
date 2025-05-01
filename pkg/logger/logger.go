// Package logger собирает функциональность логирования, контекстного логирования
// и логирования с идентификаторами запросов.
package logger

import (
	"context"
	"fmt"
	"sync"

	logCtx "github.com/flexer2006/case-person-enrichment-go/pkg/logger/context"
	"github.com/flexer2006/case-person-enrichment-go/pkg/logger/logging"
	"github.com/flexer2006/case-person-enrichment-go/pkg/logger/request"
	"go.uber.org/zap"
)

// Экспортируем типы из подпакетов для удобства использования.
type (
	// Logger - основной тип логгера.
	Logger = logging.Logger

	// LogLevel - тип уровня логирования.
	LogLevel = logging.LogLevel
)

// Экспортируем константы уровней логирования.
const (
	DebugLevel = logging.DebugLevel
	InfoLevel  = logging.InfoLevel
	WarnLevel  = logging.WarnLevel
	ErrorLevel = logging.ErrorLevel
	FatalLevel = logging.FatalLevel
)

// Глобальный логгер и его защита.
var (
	globalLogger     *logging.Logger
	globalLoggerOnce sync.Once
	globalLoggerMu   sync.RWMutex
)

// логгерПровайдер для пакета context.
type loggerProvider struct{}

// Global возвращает глобальный логгер, создавая его при необходимости.
func (p *loggerProvider) Global() logging.LoggerInterface {
	return Global()
}

// contextLoggerProvider для пакета request.
type contextLoggerProvider struct{}

func (p *contextLoggerProvider) FromContext(ctx context.Context) logging.LoggerInterface {
	return GetContextLogger(ctx)
}

// Инициализация зависимостей между пакетами.
func init() {
	request.GetContextLogger = GetContextLogger
	logCtx.GetLoggerFunc = func() logging.LoggerInterface {
		return Global()
	}

	logCtx.SetLoggerProvider(&loggerProvider{})
	request.SetContextLoggerProvider(&contextLoggerProvider{})
}

// GetContextLogger возвращает логгер из контекста или глобальный логгер.
func GetContextLogger(ctx context.Context) logging.LoggerInterface {
	if log, ok := logCtx.Logger(ctx); ok && log != nil {
		return log
	}
	return Global()
}

// Global возвращает глобальный логгер, создавая его при необходимости.
func Global() *logging.Logger {
	globalLoggerMu.RLock()
	logger := globalLogger
	globalLoggerMu.RUnlock()

	if logger != nil {
		return logger
	}

	globalLoggerOnce.Do(func() {
		var log *logging.Logger
		var err error

		log, err = logging.NewProduction()
		if err != nil {
			log = logging.NewConsole(logging.InfoLevel, true)
		}

		globalLoggerMu.Lock()
		globalLogger = log
		globalLoggerMu.Unlock()
	})

	return globalLogger
}

// SetGlobal устанавливает глобальный логгер.
func SetGlobal(logger *logging.Logger) {
	if logger == nil {
		return
	}

	globalLoggerMu.Lock()
	globalLogger = logger
	globalLoggerMu.Unlock()
}

// NewDevelopment создает логгер для разработки.
func NewDevelopment() (*logging.Logger, error) {
	logger, err := logging.NewDevelopment()
	if err != nil {
		return nil, fmt.Errorf("failed to create development logger: %w", err)
	}
	return logger, nil
}

// NewProduction создает логгер для продакшена.
func NewProduction() (*logging.Logger, error) {
	logger, err := logging.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("failed to create production logger: %w", err)
	}
	return logger, nil
}

// NewConsole создает новый логгер с выводом в консоль.
func NewConsole(level LogLevel, useJSON bool) *logging.Logger {
	return logging.NewConsole(level, useJSON)
}

// WithLogger добавляет логгер в контекст.
func WithLogger(ctx context.Context, log *logging.Logger) context.Context {
	return logCtx.WithLogger(ctx, log)
}

// WithFields добавляет поля к логгеру в контексте.
func WithFields(ctx context.Context, fields ...zap.Field) context.Context {
	return logCtx.WithFields(ctx, fields...)
}

// WithRequestID добавляет идентификатор запроса в контекст.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return request.WithRequestID(ctx, requestID)
}

// RequestID извлекает идентификатор запроса из контекста.
func RequestID(ctx context.Context) (string, bool) {
	return request.ID(ctx)
}

// ConfigureRequestID настраивает формат идентификатора запроса.
func ConfigureRequestID(config request.IDConfig) {
	request.ConfigureRequestID(config)
}

// Вспомогательная функция для выбора соответствующего логгера.
func getLoggerForContext(ctx context.Context) logging.LoggerInterface {
	if id, ok := request.ID(ctx); ok && id != "" {
		return request.Logger(ctx)
	}
	return logCtx.GetLogger(ctx)
}

// Log обобщённая функция логирования без дублирования кода.
func Log(ctx context.Context, level LogLevel, msg string, fields ...zap.Field) {
	logger := getLoggerForContext(ctx)
	if logger == nil {
		return
	}

	switch level {
	case DebugLevel:
		logger.Debug(msg, fields...)
	case InfoLevel:
		logger.Info(msg, fields...)
	case WarnLevel:
		logger.Warn(msg, fields...)
	case ErrorLevel:
		logger.Error(msg, fields...)
	case FatalLevel:
		logger.Fatal(msg, fields...)
	}
}

// Debug логирует сообщение с уровнем Debug, автоматически добавляя request_id если есть.
func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	Log(ctx, DebugLevel, msg, fields...)
}

// Info логирует сообщение с уровнем Info, автоматически добавляя request_id если есть.
func Info(ctx context.Context, msg string, fields ...zap.Field) {
	Log(ctx, InfoLevel, msg, fields...)
}

// Warn логирует сообщение с уровнем Warn, автоматически добавляя request_id если есть.
func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	Log(ctx, WarnLevel, msg, fields...)
}

// Error логирует сообщение с уровнем Error, автоматически добавляя request_id если есть.
func Error(ctx context.Context, msg string, fields ...zap.Field) {
	Log(ctx, ErrorLevel, msg, fields...)
}

// Fatal логирует сообщение с уровнем Fatal, автоматически добавляя request_id если есть.
func Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	Log(ctx, FatalLevel, msg, fields...)
}
