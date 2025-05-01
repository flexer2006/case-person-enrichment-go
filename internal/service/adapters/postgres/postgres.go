// Package postgres предоставляет адаптеры PostgreSQL для приложения.
package postgres

import (
	"context"
	"fmt"

	"github.com/flexer2006/case-person-enrichment-go/internal/service/adapters/postgres/repo"
	repoports "github.com/flexer2006/case-person-enrichment-go/internal/service/ports/repo"
	"github.com/flexer2006/case-person-enrichment-go/pkg/database/postgres"
	"github.com/flexer2006/case-person-enrichment-go/pkg/logger"
	"go.uber.org/zap"
)

// Adapter представляет основной адаптер PostgreSQL для приложения.
type Adapter struct {
	db           postgres.Provider
	repositories repoports.Repositories
}

// NewPostgresAdapter создает новый адаптер PostgreSQL со всеми репозиториями.
func NewPostgresAdapter(db postgres.Provider) *Adapter {
	return &Adapter{
		db:           db,
		repositories: repo.NewRepositories(db),
	}
}

// Repositories возвращает все репозитории для приложения.
func (p *Adapter) Repositories() repoports.Repositories {
	return p.repositories
}

// DB возвращает провайдер базы данных PostgreSQL.
func (p *Adapter) DB() postgres.Provider {
	return p.db
}

// Close закрывает соединение с PostgreSQL.
func (p *Adapter) Close(ctx context.Context) {
	logger.Info(ctx, "closing PostgreSQL adapter")
	p.db.Close(ctx)
}

// Ping проверяет доступность базы данных.
func (p *Adapter) Ping(ctx context.Context) error {
	logger.Debug(ctx, "pinging PostgreSQL database")
	err := p.db.Ping(ctx)
	if err != nil {
		logger.Error(ctx, "failed to ping PostgreSQL database", zap.Error(err))
		return fmt.Errorf("failed to ping database: %w", err)
	}
	return nil
}

// DSN возвращает строку подключения к базе данных.
func (p *Adapter) DSN() string {
	return p.db.GetDSN()
}
