// Package data содержит настройки для работы с базой данных.
package data

import (
	"github.com/flexer2006/case-person-enrichment-go/pkg/database/postgres"
	"go.uber.org/zap"
)

// PostgresConfig содержит настройки для базы данных PostgreSQL.
type PostgresConfig struct {
	Host                string `env:"POSTGRES_HOST" env-default:"localhost"`
	Port                int    `env:"POSTGRES_PORT" env-default:"5432"`
	User                string `env:"POSTGRES_USER" env-default:"postgres"`
	Password            string `env:"POSTGRES_PASSWORD" env-default:"postgres"`
	Database            string `env:"POSTGRES_DB" env-default:"postgres"`
	SSLMode             string `env:"POSTGRES_SSLMODE" env-default:"disable"`
	PoolMinConns        int    `env:"PGX_POOL_MIN_CONNS" env-default:"2"`
	PoolMaxConns        int    `env:"PGX_POOL_MAX_CONNS" env-default:"10"`
	PoolMaxConnLifetime int    `env:"PGX_POOL_MAX_CONN_LIFETIME" env-default:"3600"`
	PoolMaxConnIdleTime int    `env:"PGX_POOL_MAX_CONN_IDLE_TIME" env-default:"600"`
	ConnectTimeout      string `env:"PGX_CONNECT_TIMEOUT" env-default:"10s"`
	AcquireTimeout      string `env:"PGX_POOL_ACQUIRE_TIMEOUT" env-default:"30s"`
}

// LogFields реализует интерфейс LoggableConfig для PostgresConfig.
func (c *PostgresConfig) LogFields() []zap.Field {
	return []zap.Field{
		zap.String("host", c.Host),
		zap.Int("port", c.Port),
		zap.String("database", c.Database),
		zap.String("user", c.User),
		zap.String("sslmode", c.SSLMode),
		zap.Int("pool_min_conns", c.PoolMinConns),
		zap.Int("pool_max_conns", c.PoolMaxConns),
	}
}

// ToConfig преобразует PostgresConfig в postgres.Config.
func (c *PostgresConfig) ToConfig() postgres.Config {
	return postgres.Config{
		Host:     c.Host,
		Port:     c.Port,
		User:     c.User,
		Password: c.Password,
		Database: c.Database,
		SSLMode:  c.SSLMode,
		MinConns: c.PoolMinConns,
		MaxConns: c.PoolMaxConns,
	}
}
