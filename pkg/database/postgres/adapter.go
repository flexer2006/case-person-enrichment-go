package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Provider определяет интерфейс для провайдера базы данных PostgreSQL.
type Provider interface {
	// Pool возвращает пул соединений с базой данных.
	Pool() *pgxpool.Pool
	// Close закрывает соединение с базой данных.
	Close(ctx context.Context)
	// Ping проверяет доступность базы данных.
	Ping(ctx context.Context) error
	// GetDSN возвращает строку подключения к базе данных.
	GetDSN() string
}
