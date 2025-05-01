// Package nationality содержит интерфейсы для определения национальности по имени.
package nationality

import (
	"context"
)

// Service определяет интерфейс для определения национальности по имени.
type Service interface {
	// GetNationalityByName возвращает вероятную национальность и вероятность по имени.
	// Возвращает: код национальности (например, "RU"), вероятность (0-1), ошибка.
	GetNationalityByName(ctx context.Context, name string) (string, float64, error)
}
