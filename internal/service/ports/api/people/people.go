// Package people объединяет интерфейсы для работы с персонами и их обогащением.
package people

import (
	"github.com/flexer2006/case-person-enrichment-go/internal/service/ports/api/people/age"
	"github.com/flexer2006/case-person-enrichment-go/internal/service/ports/api/people/gender"
	"github.com/flexer2006/case-person-enrichment-go/internal/service/ports/api/people/nationality"
	"github.com/flexer2006/case-person-enrichment-go/internal/service/ports/api/people/person"
)

// Services объединяет все сервисные интерфейсы для работы с персонами.
type Services interface {
	// Person возвращает интерфейс для работы с персонами.
	Person() person.Service

	// Age возвращает интерфейс для определения возраста.
	Age() age.Service

	// Gender возвращает интерфейс для определения пола.
	Gender() gender.Service

	// Nationality возвращает интерфейс для определения национальности.
	Nationality() nationality.Service
}
