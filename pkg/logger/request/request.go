// Package request предоставляет функциональность для работы с идентификаторами запросов в логах
package request

import (
	"context"
	"regexp"

	"github.com/flexer2006/case-person-enrichment-go/pkg/logger/logging"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	// RequestIDKey ключ для идентификатора запроса в контексте.
	RequestIDKey = &ctxKey{"request_id"}

	// Глобальная конфигурация.
	requestIDConfig = IDConfig{
		ContextKey: RequestIDKey,
		FieldName:  "request_id",
	}

	// Регулярное выражение для проверки UUID.
	uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

	// GetContextLogger функция для получения логгера из контекста, будет установлена извне.
	GetContextLogger func(ctx context.Context) logging.LoggerInterface
)

// Тип для ключей контекста, использующий строгую типизацию.
type ctxKey struct{ name string }

// IDConfig конфигурация для идентификатора запроса.
type IDConfig struct {
	// Ключ для хранения в контексте
	ContextKey *ctxKey
	// Название поля в логах
	FieldName string
}

// ContextLoggerProvider интерфейс для получения логгера из контекста.
type ContextLoggerProvider interface {
	FromContext(ctx context.Context) logging.LoggerInterface
}

// defaultProvider реализация провайдера по умолчанию.
type defaultProvider struct{}

func (p *defaultProvider) FromContext(_ context.Context) logging.LoggerInterface {
	return &logging.Logger{}
}

// провайдер логгеров из контекста, будет установлен в основном пакете.
var contextProvider ContextLoggerProvider = &defaultProvider{}

// SetContextLoggerProvider устанавливает провайдер логгеров из контекста.
func SetContextLoggerProvider(p ContextLoggerProvider) {
	if p != nil {
		contextProvider = p
	}
}

// ConfigureRequestID настраивает формат идентификатора запроса.
func ConfigureRequestID(config IDConfig) {
	// Проверяем ContextKey на nil, чтобы избежать паники
	if config.ContextKey != nil {
		requestIDConfig.ContextKey = config.ContextKey
	}

	// Проверяем FieldName на пустоту
	if config.FieldName != "" {
		requestIDConfig.FieldName = config.FieldName
	}
}

// WithRequestID добавляет идентификатор запроса в контекст.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	if ctx == nil {
		return context.Background()
	}

	if requestID == "" || !IsValidUUID(requestID) {
		requestID = GenerateRequestID()
	}

	return context.WithValue(ctx, requestIDConfig.ContextKey, requestID)
}

// ID извлекает идентификатор запроса из контекста.
func ID(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}

	value := ctx.Value(requestIDConfig.ContextKey)
	if value == nil {
		return "", false
	}

	id, ok := value.(string)
	return id, ok
}

// GenerateRequestID генерирует новый идентификатор запроса.
func GenerateRequestID() string {
	return uuid.New().String()
}

// IsValidUUID проверяет, соответствует ли строка формату UUID.
func IsValidUUID(id string) bool {
	if id == "" {
		return false
	}
	return uuidRegex.MatchString(id)
}

// WithRequestIDField добавляет поле request_id к логгеру, если оно присутствует в контексте.
func WithRequestIDField(ctx context.Context, log logging.LoggerInterface) logging.LoggerInterface {
	if log == nil || ctx == nil {
		return log
	}

	id, ok := ID(ctx)
	if ok && id != "" {
		return log.With(zap.String(requestIDConfig.FieldName, id))
	}
	return log
}

// Logger получает логгер из контекста и добавляет к нему request_id.
func Logger(ctx context.Context) logging.LoggerInterface {
	log := contextProvider.FromContext(ctx)
	return WithRequestIDField(ctx, log)
}

// log унифицированный метод логирования для всех уровней.
func log(ctx context.Context, level logging.LogLevel, msg string, fields ...zap.Field) {
	lgr := Logger(ctx)
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

// Debug логирует сообщение с уровнем Debug и request_id из контекста.
func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	log(ctx, logging.DebugLevel, msg, fields...)
}

// Info логирует сообщение с уровнем Info и request_id из контекста.
func Info(ctx context.Context, msg string, fields ...zap.Field) {
	log(ctx, logging.InfoLevel, msg, fields...)
}

// Warn логирует сообщение с уровнем Warn и request_id из контекста.
func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	log(ctx, logging.WarnLevel, msg, fields...)
}

// Error логирует сообщение с уровнем Error и request_id из контекста.
func Error(ctx context.Context, msg string, fields ...zap.Field) {
	log(ctx, logging.ErrorLevel, msg, fields...)
}

// Fatal логирует сообщение с уровнем Fatal и request_id из контекста.
func Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	log(ctx, logging.FatalLevel, msg, fields...)
}
