// Package person содержит интерфейсы для работы с персонами.
package person

import (
	"context"

	"github.com/flexer2006/case-person-enrichment-go/internal/service/domain/entities"
	"github.com/google/uuid"
)

// Service определяет интерфейс для работы с персонами.
type Service interface {
	// GetByID возвращает персону по идентификатору.
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Person, error)

	// GetPersons возвращает список персон с фильтрацией и пагинацией.
	GetPersons(ctx context.Context, filter map[string]any, offset, limit int) ([]*entities.Person, int, error)

	// CreatePerson создает новую персону.
	CreatePerson(ctx context.Context, person *entities.Person) error

	// UpdatePerson обновляет существующую персону.
	UpdatePerson(ctx context.Context, person *entities.Person) error

	// DeletePerson удаляет персону по идентификатору.
	DeletePerson(ctx context.Context, id uuid.UUID) error

	// EnrichPerson обогащает данные персоны (возраст, пол, национальность).
	EnrichPerson(ctx context.Context, id uuid.UUID) (*entities.Person, error)
}
