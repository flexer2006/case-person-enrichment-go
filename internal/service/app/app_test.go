package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/flexer2006/case-person-enrichment-go/internal/service/app"
	"github.com/flexer2006/case-person-enrichment-go/internal/service/domain/entities"
	"github.com/flexer2006/case-person-enrichment-go/internal/service/ports/api/people"
	"github.com/flexer2006/case-person-enrichment-go/internal/service/ports/api/people/age"
	"github.com/flexer2006/case-person-enrichment-go/internal/service/ports/api/people/gender"
	"github.com/flexer2006/case-person-enrichment-go/internal/service/ports/api/people/nationality"
	personapi "github.com/flexer2006/case-person-enrichment-go/internal/service/ports/api/people/person"
	repopeople "github.com/flexer2006/case-person-enrichment-go/internal/service/ports/repo/people"
	personrepo "github.com/flexer2006/case-person-enrichment-go/internal/service/ports/repo/people/person"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockAPIAdapter struct {
	mock.Mock
}

func (m *mockAPIAdapter) People() people.Services {
	args := m.Called()
	return args.Get(0).(people.Services)
}

type mockRepositories struct {
	mock.Mock
}

func (m *mockRepositories) People() repopeople.Repositories {
	args := m.Called()
	return args.Get(0).(repopeople.Repositories)
}

type mockPeopleRepositories struct {
	mock.Mock
}

func (m *mockPeopleRepositories) Person() personrepo.Repository {
	args := m.Called()
	return args.Get(0).(personrepo.Repository)
}

type mockPersonRepository struct {
	mock.Mock
}

func (m *mockPersonRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Person, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Person), args.Error(1)
}

func (m *mockPersonRepository) GetPersons(ctx context.Context, filter map[string]any, offset, limit int) ([]*entities.Person, int, error) {
	args := m.Called(ctx, filter, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*entities.Person), args.Int(1), args.Error(2)
}

func (m *mockPersonRepository) CreatePerson(ctx context.Context, person *entities.Person) error {
	args := m.Called(ctx, person)
	return args.Error(0)
}

func (m *mockPersonRepository) UpdatePerson(ctx context.Context, person *entities.Person) error {
	args := m.Called(ctx, person)
	return args.Error(0)
}

func (m *mockPersonRepository) DeletePerson(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockPersonRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

type mockPeopleServices struct {
	mock.Mock
}

func (m *mockPeopleServices) Person() personapi.Service {
	args := m.Called()
	return args.Get(0).(personapi.Service)
}

func (m *mockPeopleServices) Age() age.Service {
	args := m.Called()
	return args.Get(0).(age.Service)
}

func (m *mockPeopleServices) Gender() gender.Service {
	args := m.Called()
	return args.Get(0).(gender.Service)
}

func (m *mockPeopleServices) Nationality() nationality.Service {
	args := m.Called()
	return args.Get(0).(nationality.Service)
}

type mockAgeService struct {
	mock.Mock
}

func (m *mockAgeService) GetAgeByName(ctx context.Context, name string) (int, float64, error) {
	args := m.Called(ctx, name)
	return args.Int(0), args.Get(1).(float64), args.Error(2)
}

type mockGenderService struct {
	mock.Mock
}

func (m *mockGenderService) GetGenderByName(ctx context.Context, name string) (string, float64, error) {
	args := m.Called(ctx, name)
	return args.String(0), args.Get(1).(float64), args.Error(2)
}

type mockNationalityService struct {
	mock.Mock
}

func (m *mockNationalityService) GetNationalityByName(ctx context.Context, name string) (string, float64, error) {
	args := m.Called(ctx, name)
	return args.String(0), args.Get(1).(float64), args.Error(2)
}

type mockPersonAPIService struct {
	mock.Mock
}

func (m *mockPersonAPIService) GetByName(ctx context.Context, name string) (*entities.Person, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Person), args.Error(1)
}

func TestNewPersonService(t *testing.T) {
	repositories := new(mockRepositories)
	apiAdapter := new(mockAPIAdapter)
	peopleRepo := new(mockPeopleRepositories)
	personRepo := new(mockPersonRepository)

	peopleRepo.On("Person").Return(personRepo)
	repositories.On("People").Return(peopleRepo)

	service := app.NewPersonService(repositories, apiAdapter)
	require.NotNil(t, service)
}

func TestPersonServiceGetByID(t *testing.T) {
	repositories := new(mockRepositories)
	apiAdapter := new(mockAPIAdapter)
	peopleRepo := new(mockPeopleRepositories)
	personRepo := new(mockPersonRepository)
	ctx := context.Background()
	id := uuid.New()
	expectedPerson := &entities.Person{ID: id, Name: "John Doe"}

	peopleRepo.On("Person").Return(personRepo)
	repositories.On("People").Return(peopleRepo)
	personRepo.On("GetByID", mock.Anything, id).Return(expectedPerson, nil)

	service := app.NewPersonService(repositories, apiAdapter)
	person, err := service.GetByID(ctx, id)

	require.NoError(t, err)
	assert.Equal(t, expectedPerson, person)
	personRepo.AssertExpectations(t)
}

func TestPersonServiceGetByIDError(t *testing.T) {
	repositories := new(mockRepositories)
	apiAdapter := new(mockAPIAdapter)
	peopleRepo := new(mockPeopleRepositories)
	personRepo := new(mockPersonRepository)
	ctx := context.Background()
	id := uuid.New()
	expectedError := errors.New("person not found")

	peopleRepo.On("Person").Return(personRepo)
	repositories.On("People").Return(peopleRepo)
	personRepo.On("GetByID", mock.Anything, id).Return(nil, expectedError)

	service := app.NewPersonService(repositories, apiAdapter)
	person, err := service.GetByID(ctx, id)

	require.Error(t, err)
	assert.Nil(t, person)
	personRepo.AssertExpectations(t)
}

func TestPersonServiceGetPersons(t *testing.T) {
	repositories := new(mockRepositories)
	apiAdapter := new(mockAPIAdapter)
	peopleRepo := new(mockPeopleRepositories)
	personRepo := new(mockPersonRepository)
	ctx := context.Background()
	filter := map[string]any{"name": "John"}
	expectedPersons := []*entities.Person{
		{ID: uuid.New(), Name: "John Doe"},
		{ID: uuid.New(), Name: "John Smith"},
	}
	expectedCount := 2

	peopleRepo.On("Person").Return(personRepo)
	repositories.On("People").Return(peopleRepo)
	personRepo.On("GetPersons", mock.Anything, filter, 0, 10).Return(expectedPersons, expectedCount, nil)

	service := app.NewPersonService(repositories, apiAdapter)
	persons, count, err := service.GetPersons(ctx, filter, 0, 10)

	require.NoError(t, err)
	assert.Equal(t, expectedPersons, persons)
	assert.Equal(t, expectedCount, count)
	personRepo.AssertExpectations(t)
}

func TestPersonServiceCreatePerson(t *testing.T) {
	repositories := new(mockRepositories)
	apiAdapter := new(mockAPIAdapter)
	peopleRepo := new(mockPeopleRepositories)
	personRepo := new(mockPersonRepository)
	ctx := context.Background()
	person := &entities.Person{Name: "John Doe"}

	peopleRepo.On("Person").Return(personRepo)
	repositories.On("People").Return(peopleRepo)
	personRepo.On("CreatePerson", mock.Anything, person).Return(nil)

	service := app.NewPersonService(repositories, apiAdapter)
	err := service.CreatePerson(ctx, person)

	require.NoError(t, err)
	personRepo.AssertExpectations(t)
}

func TestPersonServiceUpdatePerson(t *testing.T) {
	repositories := new(mockRepositories)
	apiAdapter := new(mockAPIAdapter)
	peopleRepo := new(mockPeopleRepositories)
	personRepo := new(mockPersonRepository)
	ctx := context.Background()
	person := &entities.Person{ID: uuid.New(), Name: "John Doe"}

	peopleRepo.On("Person").Return(personRepo)
	repositories.On("People").Return(peopleRepo)
	personRepo.On("UpdatePerson", mock.Anything, person).Return(nil)

	service := app.NewPersonService(repositories, apiAdapter)
	err := service.UpdatePerson(ctx, person)

	require.NoError(t, err)
	personRepo.AssertExpectations(t)
}

func TestPersonServiceDeletePerson(t *testing.T) {
	repositories := new(mockRepositories)
	apiAdapter := new(mockAPIAdapter)
	peopleRepo := new(mockPeopleRepositories)
	personRepo := new(mockPersonRepository)
	ctx := context.Background()
	id := uuid.New()

	peopleRepo.On("Person").Return(personRepo)
	repositories.On("People").Return(peopleRepo)
	personRepo.On("DeletePerson", mock.Anything, id).Return(nil)

	service := app.NewPersonService(repositories, apiAdapter)
	err := service.DeletePerson(ctx, id)

	require.NoError(t, err)
	personRepo.AssertExpectations(t)
}

func TestPersonServiceEnrichPerson(t *testing.T) {
	repositories := new(mockRepositories)
	apiAdapter := new(mockAPIAdapter)
	peopleRepo := new(mockPeopleRepositories)
	personRepo := new(mockPersonRepository)
	peopleServices := new(mockPeopleServices)
	ageService := new(mockAgeService)
	genderService := new(mockGenderService)
	nationalityService := new(mockNationalityService)
	personAPIService := new(mockPersonAPIService)

	ctx := context.Background()
	id := uuid.New()
	person := &entities.Person{ID: id, Name: "John Doe"}
	expectedAge := 30
	expectedGender := "male"
	expectedNationality := "US"
	genderProb := 0.95
	nationalityProb := 0.85

	enrichedPerson := &entities.Person{
		ID:                     id,
		Name:                   "John Doe",
		Age:                    &expectedAge,
		Gender:                 &expectedGender,
		GenderProbability:      &genderProb,
		Nationality:            &expectedNationality,
		NationalityProbability: &nationalityProb,
	}

	peopleRepo.On("Person").Return(personRepo)
	repositories.On("People").Return(peopleRepo)
	apiAdapter.On("People").Return(peopleServices)

	peopleServices.On("Person").Return(personAPIService)
	peopleServices.On("Age").Return(ageService)
	peopleServices.On("Gender").Return(genderService)
	peopleServices.On("Nationality").Return(nationalityService)

	personRepo.On("GetByID", mock.Anything, id).Return(person, nil)
	ageService.On("GetAgeByName", mock.Anything, "John Doe").Return(expectedAge, 0.9, nil)
	genderService.On("GetGenderByName", mock.Anything, "John Doe").Return(expectedGender, genderProb, nil)
	nationalityService.On("GetNationalityByName", mock.Anything, "John Doe").Return(expectedNationality, nationalityProb, nil)
	personRepo.On("UpdatePerson", mock.Anything, mock.MatchedBy(func(p *entities.Person) bool {
		return p.ID == id &&
			p.Name == "John Doe" &&
			*p.Age == expectedAge &&
			*p.Gender == expectedGender &&
			*p.GenderProbability == genderProb &&
			*p.Nationality == expectedNationality &&
			*p.NationalityProbability == nationalityProb
	})).Return(nil)

	service := app.NewPersonService(repositories, apiAdapter)
	result, err := service.EnrichPerson(ctx, id)

	require.NoError(t, err)
	assert.Equal(t, enrichedPerson.ID, result.ID)
	assert.Equal(t, enrichedPerson.Name, result.Name)
	assert.Equal(t, *enrichedPerson.Age, *result.Age)
	assert.Equal(t, *enrichedPerson.Gender, *result.Gender)
	assert.Equal(t, *enrichedPerson.GenderProbability, *result.GenderProbability)
	assert.Equal(t, *enrichedPerson.Nationality, *result.Nationality)
	assert.Equal(t, *enrichedPerson.NationalityProbability, *result.NationalityProbability)

	personRepo.AssertExpectations(t)
	ageService.AssertExpectations(t)
	genderService.AssertExpectations(t)
	nationalityService.AssertExpectations(t)
}

func TestPersonServiceEnrichPersonPartialFailure(t *testing.T) {
	repositories := new(mockRepositories)
	apiAdapter := new(mockAPIAdapter)
	peopleRepo := new(mockPeopleRepositories)
	personRepo := new(mockPersonRepository)
	peopleServices := new(mockPeopleServices)
	ageService := new(mockAgeService)
	genderService := new(mockGenderService)
	nationalityService := new(mockNationalityService)
	personAPIService := new(mockPersonAPIService)

	ctx := context.Background()
	id := uuid.New()
	person := &entities.Person{ID: id, Name: "John Doe"}
	expectedGender := "male"
	genderProb := 0.95

	peopleRepo.On("Person").Return(personRepo)
	repositories.On("People").Return(peopleRepo)
	apiAdapter.On("People").Return(peopleServices)

	peopleServices.On("Person").Return(personAPIService)
	peopleServices.On("Age").Return(ageService)
	peopleServices.On("Gender").Return(genderService)
	peopleServices.On("Nationality").Return(nationalityService)

	personRepo.On("GetByID", mock.Anything, id).Return(person, nil)
	ageService.On("GetAgeByName", mock.Anything, "John Doe").Return(0, 0.0, errors.New("age service error"))
	genderService.On("GetGenderByName", mock.Anything, "John Doe").Return(expectedGender, genderProb, nil)
	nationalityService.On("GetNationalityByName", mock.Anything, "John Doe").Return("", 0.0, errors.New("nationality service error"))

	personRepo.On("UpdatePerson", mock.Anything, mock.MatchedBy(func(p *entities.Person) bool {
		return p.ID == id &&
			p.Name == "John Doe" &&
			p.Age == nil &&
			*p.Gender == expectedGender &&
			*p.GenderProbability == genderProb &&
			p.Nationality == nil &&
			p.NationalityProbability == nil
	})).Return(nil)

	service := app.NewPersonService(repositories, apiAdapter)
	result, err := service.EnrichPerson(ctx, id)

	require.NoError(t, err)
	assert.Equal(t, person.ID, result.ID)
	assert.Equal(t, person.Name, result.Name)
	assert.Nil(t, result.Age)
	assert.Equal(t, expectedGender, *result.Gender)
	assert.Equal(t, genderProb, *result.GenderProbability)
	assert.Nil(t, result.Nationality)
	assert.Nil(t, result.NationalityProbability)

	personRepo.AssertExpectations(t)
	ageService.AssertExpectations(t)
	genderService.AssertExpectations(t)
	nationalityService.AssertExpectations(t)
}
