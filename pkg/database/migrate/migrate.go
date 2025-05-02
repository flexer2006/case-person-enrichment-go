// Package migrate предоставляет функциональность для выполнения миграций базы данных.
package migrate

import (
	"context"
	"errors"
	"fmt"

	"github.com/flexer2006/case-person-enrichment-go/pkg/logger"
	"github.com/golang-migrate/migrate/v4"

	// Импортируем драйвер для работы с Postgres.
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	// Импортируем драйвер для чтения миграций из файлов.
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

// ErrMigrationPathNotSpecified ошибка, возникающая, когда путь к миграциям не указан.
var (
	ErrMigrationPathNotSpecified = errors.New("migration path not specified")
)

type MigrateInstance interface {
	Up() error
	Down() error
	Version() (uint, bool, error)
	Force(version int) error
	Close() (source error, database error)
}

// Config содержит настройки для миграций.
type Config struct {
	// Путь к файлам миграций.
	Path string
}

// Migrator представляет сервис для выполнения миграций базы данных.
type Migrator struct{}

// NewMigrator создает новый сервис миграций.
func NewMigrator() *Migrator {
	return &Migrator{}
}

// Up выполняет все доступные миграции.
func (m *Migrator) Up(ctx context.Context, dsn string, cfg ...Config) error {
	var path string
	if len(cfg) > 0 && cfg[0].Path != "" {
		path = fmt.Sprintf("file://%s", cfg[0].Path)
	} else {
		return ErrMigrationPathNotSpecified
	}

	migrator, err := m.createMigrator(ctx, dsn, path)
	if err != nil {
		return err
	}

	defer m.closeMigrator(ctx, migrator)

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logger.Error(ctx, "failed to apply migrations", zap.Error(err))
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	logger.Info(ctx, "database migrations applied")
	return nil
}

// Down откатывает все миграции.
func (m *Migrator) Down(ctx context.Context, dsn string, cfg ...Config) error {
	var path string
	if len(cfg) > 0 && cfg[0].Path != "" {
		path = fmt.Sprintf("file://%s", cfg[0].Path)
	} else {
		return ErrMigrationPathNotSpecified
	}

	migrator, err := m.createMigrator(ctx, dsn, path)
	if err != nil {
		return err
	}

	defer m.closeMigrator(ctx, migrator)

	if err := migrator.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logger.Error(ctx, "failed to rollback migrations", zap.Error(err))
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}

	logger.Info(ctx, "database migrations rolled back")
	return nil
}

// Version возвращает текущую версию миграции и статус "грязный".
func (m *Migrator) Version(ctx context.Context, dsn string, cfg ...Config) (uint, bool, error) {
	var path string
	if len(cfg) > 0 && cfg[0].Path != "" {
		path = fmt.Sprintf("file://%s", cfg[0].Path)
	} else {
		return 0, false, ErrMigrationPathNotSpecified
	}

	migrator, err := m.createMigrator(ctx, dsn, path)
	if err != nil {
		return 0, false, err
	}

	defer m.closeMigrator(ctx, migrator)

	version, dirty, err := migrator.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		logger.Error(ctx, "failed to get migration version", zap.Error(err))
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}

	if errors.Is(err, migrate.ErrNilVersion) {
		return 0, false, nil
	}

	return version, dirty, nil
}

// Force устанавливает определенную версию миграции принудительно.
func (m *Migrator) Force(ctx context.Context, dsn string, version int, cfg ...Config) error {
	var path string
	if len(cfg) > 0 && cfg[0].Path != "" {
		path = fmt.Sprintf("file://%s", cfg[0].Path)
	} else {
		return ErrMigrationPathNotSpecified
	}

	migrator, err := m.createMigrator(ctx, dsn, path)
	if err != nil {
		return err
	}

	defer m.closeMigrator(ctx, migrator)

	if err := migrator.Force(version); err != nil {
		logger.Error(ctx, "failed to force migration version", zap.Error(err), zap.Int("version", version))
		return fmt.Errorf("failed to force migration version %d: %w", version, err)
	}

	return nil
}

// createMigrator создает новый экземпляр мигратора.
func (m *Migrator) createMigrator(ctx context.Context, dsn string, path string) (*migrate.Migrate, error) {
	migrator, err := migrate.New(path, dsn)
	if err != nil {
		logger.Error(ctx, "failed to create migration instance", zap.Error(err), zap.String("path", path))
		return nil, fmt.Errorf("failed to create migration instance: %w", err)
	}
	return migrator, nil
}

// closeMigrator безопасно закрывает мигратор.
func (m *Migrator) closeMigrator(ctx context.Context, migrator *migrate.Migrate) {
	sourceErr, dbErr := migrator.Close()
	if sourceErr != nil {
		logger.Error(ctx, "failed to close migration source", zap.Error(sourceErr))
	}
	if dbErr != nil {
		logger.Error(ctx, "failed to close migration database", zap.Error(dbErr))
	}
}
