// Package postgres предоставляет функциональность для работы с PostgreSQL.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/flexer2006/case-person-enrichment-go/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// ErrInvalidConfiguration ошибка, возникающая при неверной конфигурации базы данных.
var (
	ErrInvalidConfiguration = errors.New("invalid database configuration: required fields missing")
)

// Config содержит настройки для подключения к базе данных.
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
	MinConns int
	MaxConns int
}

// Validate проверяет конфигурацию на валидность.
func (c Config) Validate() error {
	if c.Host == "" || c.Port == 0 || c.User == "" || c.Database == "" {
		return ErrInvalidConfiguration
	}
	return nil
}

// DSN возвращает строку подключения к базе данных.
func (c Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Database, c.SSLMode)
}

// Database представляет соединение с PostgreSQL.
type Database struct {
	pool   *pgxpool.Pool
	config Config
}

// New создает новое соединение с базой данных PostgreSQL.
func New(ctx context.Context, config Config) (*Database, error) {
	if err := config.Validate(); err != nil {
		logger.Error(ctx, "invalid database configuration", zap.Error(err))
		return nil, err
	}

	dsn := config.DSN()

	logger.Info(ctx, "connecting to postgres database",
		zap.String("host", config.Host),
		zap.Int("port", config.Port),
		zap.String("database", config.Database),
		zap.String("user", config.User))

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		logger.Error(ctx, "failed to parse database configuration", zap.Error(err))
		return nil, fmt.Errorf("failed to parse database configuration: %w", err)
	}

	if config.MinConns > 0 {
		if config.MinConns > math.MaxInt32 {
			logger.Warn(ctx, "MinConns value exceeds maximum allowed value, setting to max int32")
			poolCfg.MinConns = math.MaxInt32
		} else {
			poolCfg.MinConns = int32(config.MinConns)
		}
	}

	if config.MaxConns > 0 {
		if config.MaxConns > math.MaxInt32 {
			logger.Warn(ctx, "MaxConns value exceeds maximum allowed value, setting to max int32")
			poolCfg.MaxConns = math.MaxInt32
		} else {
			poolCfg.MaxConns = int32(config.MaxConns)
		}
	}

	poolCfg.ConnConfig.ConnectTimeout = 5 * time.Second

	poolCfg.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		logger.Error(ctx, "failed to create connection pool", zap.Error(err))
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		logger.Error(ctx, "failed to ping database", zap.Error(err))
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info(ctx, "connected to postgres database",
		zap.String("host", config.Host),
		zap.Int("port", config.Port),
		zap.String("database", config.Database))

	return &Database{
		pool:   pool,
		config: config,
	}, nil
}

// NewWithDSN создает новое соединение с базой данных по DSN.
func NewWithDSN(ctx context.Context, dsn string, minConn, maxConn int) (*Database, error) {
	logger.Info(ctx, "connecting to postgres database")

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		logger.Error(ctx, "failed to parse database configuration", zap.Error(err))
		return nil, fmt.Errorf("failed to parse database configuration: %w", err)
	}

	if minConn > 0 {
		if minConn > math.MaxInt32 {
			logger.Warn(ctx, "minConn value exceeds maximum allowed value, setting to max int32")
			poolCfg.MinConns = math.MaxInt32
		} else {
			poolCfg.MinConns = int32(minConn)
		}
	}

	if maxConn > 0 {
		if maxConn > math.MaxInt32 {
			logger.Warn(ctx, "maxConn value exceeds maximum allowed value, setting to max int32")
			poolCfg.MaxConns = math.MaxInt32
		} else {
			poolCfg.MaxConns = int32(maxConn)
		}
	}

	poolCfg.ConnConfig.ConnectTimeout = 5 * time.Second

	poolCfg.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		logger.Error(ctx, "failed to create connection pool", zap.Error(err))
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		logger.Error(ctx, "failed to ping database", zap.Error(err))
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info(ctx, "connected to postgres database")

	config := Config{
		MinConns: minConn,
		MaxConns: maxConn,
	}

	return &Database{
		pool:   pool,
		config: config,
	}, nil
}

// Pool возвращает пул соединений с базой данных.
func (db *Database) Pool() *pgxpool.Pool {
	return db.pool
}

// Close закрывает соединение с базой данных.
func (db *Database) Close(ctx context.Context) {
	logger.Info(ctx, "closing postgres database connection")
	db.pool.Close()
}

// Ping проверяет доступность базы данных.
func (db *Database) Ping(ctx context.Context) error {
	if db.pool == nil {
		return fmt.Errorf("failed to ping database: connection pool is nil")
	}

	if err := db.pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	return nil
}

// Config возвращает конфигурацию базы данных.
func (db *Database) Config() Config {
	return db.config
}

// GetDSN возвращает строку подключения к базе данных.
func (db *Database) GetDSN() string {
	return db.config.DSN()
}
