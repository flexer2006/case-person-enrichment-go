// Package gender содержит интерфейсы для определения пола по имени.
package gender

import (
	"context"
)

// Service определяет интерфейс для определения пола по имени.
type Service interface {
	// GetGenderByName возвращает вероятный пол и вероятность по имени.
	// Возвращает: пол (male/female), вероятность (0-1), ошибка.
	GetGenderByName(ctx context.Context, name string) (string, float64, error)
}
