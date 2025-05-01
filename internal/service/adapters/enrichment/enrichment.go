// Package enrichment предоставляет реализацию сервисов обогащения данных.
package enrichment

import (
	"github.com/flexer2006/case-person-enrichment-go/internal/service/adapters/enrichment/api"
	apiports "github.com/flexer2006/case-person-enrichment-go/internal/service/ports/api"
	peopleapi "github.com/flexer2006/case-person-enrichment-go/internal/service/ports/api/people"
)

// Проверка, что Enrichment реализует интерфейс apiports.Api.
var _ apiports.API = (*Enrichment)(nil)

// Enrichment реализует интерфейс apiports.Api, предоставляя доступ к сервисам обогащения данных.
type Enrichment struct {
	api *api.API
}

// NewEnrichment создает новый экземпляр Enrichment с указанными API сервисами.
func NewEnrichment(api *api.API) *Enrichment {
	return &Enrichment{
		api: api,
	}
}

// NewDefaultEnrichment создает новый экземпляр Enrichment с API сервисами по умолчанию.
func NewDefaultEnrichment() *Enrichment {
	return &Enrichment{
		api: api.NewDefaultAPI(),
	}
}

// People возвращает интерфейсы для работы с данными о людях.
func (e *Enrichment) People() peopleapi.Services {
	return e.api.People()
}
