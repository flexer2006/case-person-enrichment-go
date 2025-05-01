// Package repo определяет порты для взаимодействия с хранилищами данных.
package repo

import (
	"github.com/flexer2006/case-person-enrichment-go/internal/service/ports/repo/people"
)

// Repositories объединяет все репозитории приложения.
type Repositories interface {
	// People возвращает репозитории для работы с данными о людях.
	People() people.Repositories
}
