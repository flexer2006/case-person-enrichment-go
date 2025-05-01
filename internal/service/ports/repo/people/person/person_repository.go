// Package person содержит интерфейсы для работы с хранилищем данных о персонах.
package person

import (
	"context"

	"github.com/flexer2006/case-person-enrichment-go/internal/service/domain/entities"
	"github.com/google/uuid"
)

// Repository определяет интерфейс для работы с хранилищем персон.
type Repository interface {
	// GetByID получает персону по идентификатору.
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Person, error)

	// GetPersons получает список персон с фильтрацией и пагинацией.
	// Возвращает: список персон, общее количество записей, ошибка.
	GetPersons(ctx context.Context, filter map[string]any, offset, limit int) ([]*entities.Person, int, error)

	// CreatePerson создает новую персону.
	CreatePerson(ctx context.Context, person *entities.Person) error

	// UpdatePerson обновляет существующую персону.
	UpdatePerson(ctx context.Context, person *entities.Person) error

	// DeletePerson удаляет персону по идентификатору.
	DeletePerson(ctx context.Context, id uuid.UUID) error

	// ExistsByID проверяет существование персоны по идентификатору.
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
}
