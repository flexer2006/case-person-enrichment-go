// Package repo содержит реализацию репозиториев с использованием PostgreSQL.
package repo

import (
	"github.com/flexer2006/case-person-enrichment-go/internal/service/adapters/postgres/repo/people"
	"github.com/flexer2006/case-person-enrichment-go/internal/service/ports/repo"
	peoplerepo "github.com/flexer2006/case-person-enrichment-go/internal/service/ports/repo/people"
	"github.com/flexer2006/case-person-enrichment-go/pkg/database/postgres"
)

// Проверка реализации интерфейса.
var _ repo.Repositories = (*Repositories)(nil)

// Repositories реализует интерфейс repo.Repositories для PostgreSQL.
type Repositories struct {
	peopleRepo peoplerepo.Repositories
}

// NewRepositories создает новый экземпляр репозиториев с PostgreSQL.
func NewRepositories(db postgres.Provider) *Repositories {
	return &Repositories{
		peopleRepo: people.NewRepositories(db),
	}
}

// People возвращает репозитории для работы с данными о людях.
func (r *Repositories) People() peoplerepo.Repositories {
	return r.peopleRepo
}
