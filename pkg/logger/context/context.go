// Package context предоставляет утилиты для работы с логгерами в контексте.
package context

import (
	"context"

	"github.com/flexer2006/case-person-enrichment-go/pkg/logger/logging"
	"go.uber.org/zap"
)

// Тип для ключей контекста, использующий строгую типизацию.
type ctxKey struct{ name string }

// LoggerKey ключ для логгера в контексте.
var LoggerKey = &ctxKey{"logger"}

// В начале файла, после определения переменных.
var (
	// Функция для получения глобального логгера, будет установлена извне.
	GetLoggerFunc func() logging.LoggerInterface
)

// LoggerProvider интерфейс для получения глобального логгера.
type LoggerProvider interface {
	Global() logging.LoggerInterface
}

// defaultProvider реализация провайдера по умолчанию.
type defaultProvider struct{}

func (p *defaultProvider) Global() logging.LoggerInterface {
	return &logging.Logger{}
}

// провайдер логгеров, будет установлен в основном пакете.
var provider LoggerProvider = &defaultProvider{}

// SetLoggerProvider устанавливает провайдер логгеров.
func SetLoggerProvider(p LoggerProvider) {
	if p != nil {
		provider = p
	}
}

// WithLogger добавляет логгер в контекст.
func WithLogger(ctx context.Context, log *logging.Logger) context.Context {
	if ctx == nil || log == nil {
		return ctx
	}
	return context.WithValue(ctx, LoggerKey, log)
}

// WithFields добавляет поля к логгеру в контексте или к новому логгеру.
func WithFields(ctx context.Context, fields ...zap.Field) context.Context {
	logger := GetLogger(ctx)
	if logger == nil {
		return ctx
	}

	if concreteLogger, ok := logger.(*logging.Logger); ok {
		return WithLogger(ctx, concreteLogger.With(fields...).(*logging.Logger))
	}

	return ctx
}

// Logger безопасно извлекает логгер из контекста.
func Logger(ctx context.Context) (*logging.Logger, bool) {
	if ctx == nil {
		return nil, false
	}

	value := ctx.Value(LoggerKey)
	if value == nil {
		return nil, false
	}

	log, ok := value.(*logging.Logger)
	return log, ok
}

// GetLogger возвращает логгер из контекста или глобальный логгер.
func GetLogger(ctx context.Context) logging.LoggerInterface {
	if log, ok := Logger(ctx); ok && log != nil {
		return log
	}

	return provider.Global()
}

// Log - обобщенная функция логирования для всех уровней.
func Log(ctx context.Context, level logging.LogLevel, msg string, fields ...zap.Field) {
	lgr := GetLogger(ctx)
	if lgr == nil {
		return
	}

	switch level {
	case logging.DebugLevel:
		lgr.Debug(msg, fields...)
	case logging.InfoLevel:
		lgr.Info(msg, fields...)
	case logging.WarnLevel:
		lgr.Warn(msg, fields...)
	case logging.ErrorLevel:
		lgr.Error(msg, fields...)
	case logging.FatalLevel:
		lgr.Fatal(msg, fields...)
	}
}

// Debug логирует сообщение с уровнем Debug, используя логгер из контекста.
func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	Log(ctx, logging.DebugLevel, msg, fields...)
}

// Info логирует сообщение с уровнем Info, используя логгер из контекста.
func Info(ctx context.Context, msg string, fields ...zap.Field) {
	Log(ctx, logging.InfoLevel, msg, fields...)
}

// Warn логирует сообщение с уровнем Warn, используя логгер из контекста.
func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	Log(ctx, logging.WarnLevel, msg, fields...)
}

// Error логирует сообщение с уровнем Error, используя логгер из контекста.
func Error(ctx context.Context, msg string, fields ...zap.Field) {
	Log(ctx, logging.ErrorLevel, msg, fields...)
}

// Fatal логирует сообщение с уровнем Fatal, используя логгер из контекста.
func Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	Log(ctx, logging.FatalLevel, msg, fields...)
}
