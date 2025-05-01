// Package api объединяет все адаптеры для взаимодействия с внешними API.
package api

import (
	"github.com/flexer2006/case-person-enrichment-go/internal/service/adapters/enrichment/api/people"
	apiports "github.com/flexer2006/case-person-enrichment-go/internal/service/ports/api"
	peopleapi "github.com/flexer2006/case-person-enrichment-go/internal/service/ports/api/people"
)

// Проверка, что API реализует интерфейс apiports.Api.
var _ apiports.API = (*API)(nil)

// API реализует интерфейс apiports.Api, предоставляя доступ к различным API сервисам.
type API struct {
	peopleServices peopleapi.Services
}

// NewAPI создает новый экземпляр API с указанными сервисами.
func NewAPI(peopleServices peopleapi.Services) *API {
	return &API{
		peopleServices: peopleServices,
	}
}

// NewDefaultAPI создает новый экземпляр API с сервисами по умолчанию.
func NewDefaultAPI() *API {
	return &API{
		peopleServices: people.NewAPIServices(),
	}
}

// People возвращает интерфейсы для работы с данными о людях.
func (a *API) People() peopleapi.Services {
	return a.peopleServices
}
