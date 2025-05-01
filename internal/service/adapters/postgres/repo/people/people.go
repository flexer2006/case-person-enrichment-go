// Package people содержит реализацию репозиториев для работы с данными о людях.
package people

import (
	"github.com/flexer2006/case-person-enrichment-go/internal/service/adapters/postgres/repo/people/person"
	"github.com/flexer2006/case-person-enrichment-go/internal/service/ports/repo/people"
	personrepo "github.com/flexer2006/case-person-enrichment-go/internal/service/ports/repo/people/person"
	"github.com/flexer2006/case-person-enrichment-go/pkg/database/postgres"
)

// Проверка реализации интерфейса.
var _ people.Repositories = (*Repositories)(nil)

// Repositories реализует интерфейс people.Repositories для PostgreSQL.
type Repositories struct {
	personRepo personrepo.Repository
}

// NewRepositories создает новый экземпляр репозиториев для работы с данными о людях.
func NewRepositories(db postgres.Provider) *Repositories {
	return &Repositories{
		personRepo: person.NewRepository(db),
	}
}

// Person возвращает репозиторий для работы с персонами.
func (r *Repositories) Person() personrepo.Repository {
	return r.personRepo
}
