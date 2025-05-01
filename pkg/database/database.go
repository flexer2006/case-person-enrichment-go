// Package database объединяет функциональность работы с базой данных и миграциями.
package database

import (
	"context"
	"fmt"

	"github.com/flexer2006/case-person-enrichment-go/pkg/database/migrate"
	"github.com/flexer2006/case-person-enrichment-go/pkg/database/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Config содержит настройки для базы данных и миграций.
type Config struct {
	Postgres        postgres.Config // Конфигурация PostgreSQL.
	Migrate         migrate.Config  // Конфигурация миграций.
	ApplyMigrations bool            // Флаг для применения миграций при инициализации.
}

// Database представляет базу данных с подключением и миграциями.
type Database struct {
	provider postgres.Provider // Провайдер базы данных.
	migrator migrate.Provider  // Провайдер миграций.
}

// New создает новое подключение к базе данных и опционально применяет миграции.
func New(ctx context.Context, cfg Config) (*Database, error) {
	postgresDB, err := postgres.New(ctx, cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("failed to setup database: %w", err)
	}

	migrator := migrate.NewAdapter(cfg.Migrate)

	database := &Database{
		provider: postgresDB,
		migrator: migrator,
	}

	if cfg.ApplyMigrations {
		if err := database.ApplyMigrations(ctx); err != nil {
			database.Close(ctx)
			return nil, err
		}
	}

	return database, nil
}

// NewWithDSN создает новое подключение к базе данных по DSN строке и применяет миграции.
func NewWithDSN(ctx context.Context, dsn string, minConn, maxConn int, migrationsPath string, applyMigrations bool) (*Database, error) {
	postgresDB, err := postgres.NewWithDSN(ctx, dsn, minConn, maxConn)
	if err != nil {
		return nil, fmt.Errorf("failed to setup database: %w", err)
	}

	migrateConfig := migrate.Config{Path: migrationsPath}
	migrator := migrate.NewAdapter(migrateConfig)

	database := &Database{
		provider: postgresDB,
		migrator: migrator,
	}

	if applyMigrations && migrationsPath != "" {
		if err := database.ApplyMigrations(ctx); err != nil {
			database.Close(ctx)
			return nil, err
		}
	}

	return database, nil
}

// Pool возвращает пул соединений с базой данных.
func (d *Database) Pool() *pgxpool.Pool {
	return d.provider.Pool()
}

// Close закрывает соединение с базой данных.
func (d *Database) Close(ctx context.Context) {
	d.provider.Close(ctx)
}

// ApplyMigrations применяет миграции к базе данных.
func (d *Database) ApplyMigrations(ctx context.Context) error {
	dsn := d.provider.GetDSN()
	if err := d.migrator.Up(ctx, dsn); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}
	return nil
}

// RollbackMigrations откатывает все миграции.
func (d *Database) RollbackMigrations(ctx context.Context) error {
	dsn := d.provider.GetDSN()
	if err := d.migrator.Down(ctx, dsn); err != nil {
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}
	return nil
}

// GetMigrationVersion возвращает текущую версию миграции и статус "грязный".
func (d *Database) GetMigrationVersion(ctx context.Context) (uint, bool, error) {
	dsn := d.provider.GetDSN()
	version, dirty, err := d.migrator.Version(ctx, dsn)
	if err != nil {
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}
	return version, dirty, nil
}

// Ping проверяет доступность базы данных.
func (d *Database) Ping(ctx context.Context) error {
	if err := d.provider.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	return nil
}
