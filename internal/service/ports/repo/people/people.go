// Package people объединяет интерфейсы для работы с хранилищем данных о персонах.
package people

import (
	"github.com/flexer2006/case-person-enrichment-go/internal/service/ports/repo/people/person"
)

// Repositories объединяет все репозитории для работы с данными о людях.
type Repositories interface {
	// Person возвращает репозиторий для работы с персонами.
	Person() person.Repository
}
