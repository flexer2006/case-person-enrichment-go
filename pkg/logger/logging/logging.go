// Package logging предоставляет базовую функциональность логирования.
package logging

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogLevel представляет уровень логирования.
type LogLevel int

// LoggerInterface определяет интерфейс для логгеров.
// Это позволяет создавать мок-имплементации для тестирования.
type LoggerInterface interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	With(fields ...zap.Field) LoggerInterface
	SetLevel(level LogLevel)
	GetLevel() LogLevel
	Sync() error
}

// Убедимся, что Logger реализует LoggerInterface.
var _ LoggerInterface = (*Logger)(nil)

// Уровни логирования с использованием iota.
const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// String возвращает строковое представление уровня логирования.
func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	default:
		return "unknown"
	}
}

// ToZapLevel преобразует LogLevel в zapcore.Level.
func (l LogLevel) ToZapLevel() zapcore.Level {
	switch l {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// ParseLevel конвертирует строку в LogLevel.
func ParseLevel(level string) LogLevel {
	switch level {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "fatal":
		return FatalLevel
	default:
		return InfoLevel
	}
}

// LevelFromZap преобразует zapcore.Level в LogLevel.
func LevelFromZap(level zapcore.Level) LogLevel {
	switch level {
	case zapcore.DebugLevel:
		return DebugLevel
	case zapcore.InfoLevel:
		return InfoLevel
	case zapcore.WarnLevel:
		return WarnLevel
	case zapcore.ErrorLevel:
		return ErrorLevel
	case zapcore.FatalLevel:
		return FatalLevel
	default:
		return InfoLevel
	}
}

// Logger предоставляет интерфейс для логирования.
type Logger struct {
	zapLogger *zap.Logger
	level     zap.AtomicLevel
}

// NewCore создает новое ядро логирования.
func NewCore(encoder zapcore.Encoder, writer zapcore.WriteSyncer, level LogLevel) zapcore.Core {
	return zapcore.NewCore(
		encoder,
		writer,
		zap.NewAtomicLevelAt(level.ToZapLevel()),
	)
}

// New создает новый логгер с указанными настройками ядра.
func New(core zapcore.Core) *Logger {
	zapLogger := zap.New(
		core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	var level zap.AtomicLevel

	if le, ok := core.(interface{ Level() zapcore.Level }); ok {
		level = zap.NewAtomicLevelAt(le.Level())
	} else {
		level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	return &Logger{
		zapLogger: zapLogger,
		level:     level,
	}
}

// NewConsole создает логгер с выводом в консоль.
func NewConsole(level LogLevel, json bool) *Logger {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var encoder zapcore.Encoder
	if json {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	atomicLevel := zap.NewAtomicLevelAt(level.ToZapLevel())
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		atomicLevel,
	)

	zapLogger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &Logger{
		zapLogger: zapLogger,
		level:     atomicLevel,
	}
}

// NewDevelopment создает логгер для разработки.
func NewDevelopment() (*Logger, error) {
	cfg := zap.NewDevelopmentConfig()
	zapLogger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build development logger: %w", err)
	}

	return &Logger{
		zapLogger: zapLogger,
		level:     zap.NewAtomicLevelAt(zapcore.DebugLevel),
	}, nil
}

// NewProduction создает логгер для продакшена.
func NewProduction() (*Logger, error) {
	cfg := zap.NewProductionConfig()
	zapLogger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build production logger: %w", err)
	}

	return &Logger{
		zapLogger: zapLogger,
		level:     zap.NewAtomicLevelAt(zapcore.InfoLevel),
	}, nil
}

// With создаёт новый логгер с дополнительными полями.
func (l *Logger) With(fields ...zap.Field) LoggerInterface {
	return &Logger{
		zapLogger: l.zapLogger.With(fields...),
		level:     l.level,
	}
}

// SetLevel изменяет уровень логирования.
func (l *Logger) SetLevel(level LogLevel) {
	l.level.SetLevel(level.ToZapLevel())
}

// GetLevel возвращает текущий уровень логирования.
func (l *Logger) GetLevel() LogLevel {
	return LevelFromZap(l.level.Level())
}

// log унифицированный метод логирования.
func (l *Logger) log(level zapcore.Level, msg string, fields ...zap.Field) {
	// Проверка на Nop-логгер
	if l == nil || l.zapLogger == nil {
		return
	}

	switch level {
	case zapcore.DebugLevel:
		l.zapLogger.Debug(msg, fields...)
	case zapcore.InfoLevel:
		l.zapLogger.Info(msg, fields...)
	case zapcore.WarnLevel:
		l.zapLogger.Warn(msg, fields...)
	case zapcore.ErrorLevel:
		l.zapLogger.Error(msg, fields...)
	case zapcore.FatalLevel:
		l.zapLogger.Fatal(msg, fields...)
	default:
		l.zapLogger.Info(msg, fields...)
	}
}

// Debug логирует сообщение с уровнем Debug.
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.log(zapcore.DebugLevel, msg, fields...)
}

// Info логирует сообщение с уровнем Info.
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.log(zapcore.InfoLevel, msg, fields...)
}

// Warn логирует сообщение с уровнем Warn.
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.log(zapcore.WarnLevel, msg, fields...)
}

// Error логирует сообщение с уровнем Error.
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.log(zapcore.ErrorLevel, msg, fields...)
}

// Fatal логирует сообщение с уровнем Fatal и завершает работу программы.
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.log(zapcore.FatalLevel, msg, fields...)
}

// Sync сбрасывает буферизованные записи логгера.
func (l *Logger) Sync() error {
	if l == nil || l.zapLogger == nil {
		return nil
	}
	return fmt.Errorf("failed to sync logger: %w", l.zapLogger.Sync())
}

// RawLogger возвращает нижележащий zap.Logger для расширенной функциональности.
func (l *Logger) RawLogger() *zap.Logger {
	return l.zapLogger
}
