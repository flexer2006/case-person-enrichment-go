// Package people объединяет адаптеры для работы с персонами.
package people

import (
	"github.com/flexer2006/case-person-enrichment-go/internal/service/adapters/enrichment/api/people/age"
	"github.com/flexer2006/case-person-enrichment-go/internal/service/adapters/enrichment/api/people/gender"
	"github.com/flexer2006/case-person-enrichment-go/internal/service/adapters/enrichment/api/people/nationality"
	peopleports "github.com/flexer2006/case-person-enrichment-go/internal/service/ports/api/people"
	ageservice "github.com/flexer2006/case-person-enrichment-go/internal/service/ports/api/people/age"
	genderservice "github.com/flexer2006/case-person-enrichment-go/internal/service/ports/api/people/gender"
	nationalityservice "github.com/flexer2006/case-person-enrichment-go/internal/service/ports/api/people/nationality"
	personservice "github.com/flexer2006/case-person-enrichment-go/internal/service/ports/api/people/person"
)

// Проверка, что Services реализует интерфейс peopleports.Services.
var _ peopleports.Services = (*Services)(nil)

// Services реализует интерфейс peopleports.Services, предоставляя доступ к сервисам обогащения данных о людях.
type Services struct {
	ageService         ageservice.Service
	genderService      genderservice.Service
	nationalityService nationalityservice.Service
	personService      personservice.Service
}

// NewServices создает новый экземпляр Services с указанными адаптерами.
func NewServices(
	personService personservice.Service,
	ageService ageservice.Service,
	genderService genderservice.Service,
	nationalityService nationalityservice.Service,
) *Services {
	return &Services{
		personService:      personService,
		ageService:         ageService,
		genderService:      genderService,
		nationalityService: nationalityService,
	}
}

// NewAPIServices создает новый экземпляр Services с адаптерами внешних API без персон.
func NewAPIServices() *Services {
	return &Services{
		personService:      nil, // Будет добавлен позже в другом месте
		ageService:         age.NewAgeAPIClient(nil),
		genderService:      gender.NewGenderAPIClient(nil),
		nationalityService: nationality.NewNationalityAPIClient(nil),
	}
}

// Person возвращает интерфейс для работы с персонами.
func (s *Services) Person() personservice.Service {
	return s.personService
}

// Age возвращает интерфейс для определения возраста.
func (s *Services) Age() ageservice.Service {
	return s.ageService
}

// Gender возвращает интерфейс для определения пола.
func (s *Services) Gender() genderservice.Service {
	return s.genderService
}

// Nationality возвращает интерфейс для определения национальности.
func (s *Services) Nationality() nationalityservice.Service {
	return s.nationalityService
}
