// Package age определяет интерфейс для сервиса, который определяет возраст по имени.
package age

import (
	"context"
)

// Service определяет интерфейс для определения возраста по имени.
type Service interface {
	// GetAgeByName возвращает вероятный возраст и вероятность по имени.
	// Возвращает: возраст, вероятность (0-1), ошибка.
	GetAgeByName(ctx context.Context, name string) (int, float64, error)
}
