// Package api определяет порты для взаимодействия с внешними сервисами.
package api

import (
	"github.com/flexer2006/case-person-enrichment-go/internal/service/ports/api/people"
)

// API объединяет все сервисные интерфейсы приложения.
type API interface {
	// People возвращает интерфейсы для работы с данными о людях.
	People() people.Services
}
