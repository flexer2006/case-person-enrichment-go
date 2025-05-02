package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/flexer2006/case-person-enrichment-go/internal/service/adapters/server/handlers"
	"github.com/flexer2006/case-person-enrichment-go/internal/service/domain/entities"
	"github.com/flexer2006/case-person-enrichment-go/internal/service/ports/api/people"
	"github.com/flexer2006/case-person-enrichment-go/internal/service/ports/api/people/age"
	"github.com/flexer2006/case-person-enrichment-go/internal/service/ports/api/people/gender"
	"github.com/flexer2006/case-person-enrichment-go/internal/service/ports/api/people/nationality"
	personapi "github.com/flexer2006/case-person-enrichment-go/internal/service/ports/api/people/person"
	repopeople "github.com/flexer2006/case-person-enrichment-go/internal/service/ports/repo/people"
	personrepo "github.com/flexer2006/case-person-enrichment-go/internal/service/ports/repo/people/person"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockAPI struct {
	mock.Mock
	mockPeopleServices *MockPeopleServices
}

func (m *MockAPI) People() people.Services {
	return m.mockPeopleServices
}

type MockPeopleServices struct {
	mock.Mock
	mockPersonService      *MockPersonService
	mockAgeService         *MockAgeService
	mockGenderService      *MockGenderService
	mockNationalityService *MockNationalityService
}

func (m *MockPeopleServices) Person() personapi.Service {
	return m.mockPersonService
}

func (m *MockPeopleServices) Age() age.Service {
	return m.mockAgeService
}

func (m *MockPeopleServices) Gender() gender.Service {
	return m.mockGenderService
}

func (m *MockPeopleServices) Nationality() nationality.Service {
	return m.mockNationalityService
}

type MockPersonService struct {
	mock.Mock
}

func (m *MockPersonService) GetByID(ctx context.Context, id uuid.UUID) (*entities.Person, error) {
	args := m.Called(ctx, id)
	if person, ok := args.Get(0).(*entities.Person); ok {
		return person, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPersonService) GetPersons(ctx context.Context, filter map[string]any, offset, limit int) ([]*entities.Person, int, error) {
	args := m.Called(ctx, filter, offset, limit)
	if persons, ok := args.Get(0).([]*entities.Person); ok {
		return persons, args.Int(1), args.Error(2)
	}
	return nil, args.Int(1), args.Error(2)
}

func (m *MockPersonService) CreatePerson(ctx context.Context, person *entities.Person) error {
	args := m.Called(ctx, person)
	return args.Error(0)
}

func (m *MockPersonService) UpdatePerson(ctx context.Context, person *entities.Person) error {
	args := m.Called(ctx, person)
	return args.Error(0)
}

func (m *MockPersonService) DeletePerson(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPersonService) EnrichPerson(ctx context.Context, id uuid.UUID) (*entities.Person, error) {
	args := m.Called(ctx, id)
	if person, ok := args.Get(0).(*entities.Person); ok {
		return person, args.Error(1)
	}
	return nil, args.Error(1)
}

type MockAgeService struct {
	mock.Mock
}

func (m *MockAgeService) GetAgeByName(ctx context.Context, name string) (int, float64, error) {
	args := m.Called(ctx, name)
	return args.Int(0), args.Get(1).(float64), args.Error(2)
}

type MockGenderService struct {
	mock.Mock
}

func (m *MockGenderService) GetGenderByName(ctx context.Context, name string) (string, float64, error) {
	args := m.Called(ctx, name)
	return args.String(0), args.Get(1).(float64), args.Error(2)
}

type MockNationalityService struct {
	mock.Mock
}

func (m *MockNationalityService) GetNationalityByName(ctx context.Context, name string) (string, float64, error) {
	args := m.Called(ctx, name)
	return args.String(0), args.Get(1).(float64), args.Error(2)
}

type MockRepositories struct {
	mock.Mock
	mockPeopleRepositories *MockPeopleRepositories
}

func (m *MockRepositories) People() repopeople.Repositories {
	return m.mockPeopleRepositories
}

type MockPeopleRepositories struct {
	mock.Mock
	mockPersonRepository *MockPersonRepository
}

func (m *MockPeopleRepositories) Person() personrepo.Repository {
	return m.mockPersonRepository
}

type MockPersonRepository struct {
	mock.Mock
}

func (m *MockPersonRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Person, error) {
	args := m.Called(ctx, id)
	if person, ok := args.Get(0).(*entities.Person); ok {
		return person, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPersonRepository) GetPersons(ctx context.Context, filter map[string]any, offset, limit int) ([]*entities.Person, int, error) {
	args := m.Called(ctx, filter, offset, limit)
	if persons, ok := args.Get(0).([]*entities.Person); ok {
		return persons, args.Int(1), args.Error(2)
	}
	return nil, args.Int(1), args.Error(2)
}

func (m *MockPersonRepository) CreatePerson(ctx context.Context, person *entities.Person) error {
	args := m.Called(ctx, person)
	return args.Error(0)
}

func (m *MockPersonRepository) UpdatePerson(ctx context.Context, person *entities.Person) error {
	args := m.Called(ctx, person)
	return args.Error(0)
}

func (m *MockPersonRepository) DeletePerson(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPersonRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func TestNewPersonHandler(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "should create a new person handler with provided dependencies",
			test: func(t *testing.T) {
				// Arrange
				mockPersonService := &MockPersonService{}
				mockAgeService := &MockAgeService{}
				mockGenderService := &MockGenderService{}
				mockNationalityService := &MockNationalityService{}

				mockPeopleServices := &MockPeopleServices{
					mockPersonService:      mockPersonService,
					mockAgeService:         mockAgeService,
					mockGenderService:      mockGenderService,
					mockNationalityService: mockNationalityService,
				}

				mockAPI := &MockAPI{
					mockPeopleServices: mockPeopleServices,
				}

				mockPersonRepository := &MockPersonRepository{}
				mockPeopleRepositories := &MockPeopleRepositories{
					mockPersonRepository: mockPersonRepository,
				}

				mockRepositories := &MockRepositories{
					mockPeopleRepositories: mockPeopleRepositories,
				}

				handler := handlers.NewPersonHandler(mockAPI, mockRepositories)

				assert.NotNil(t, handler, "Handler should not be nil")
			},
		},
		{
			name: "should handle nil dependencies",
			test: func(t *testing.T) {
				handler := handlers.NewPersonHandler(nil, nil)

				assert.NotNil(t, handler, "Handler should not be nil even with nil dependencies")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

func TestGetPersons(t *testing.T) {
	setupApp := func() (*fiber.App, *MockPersonRepository, *handlers.PersonHandler) {
		app := fiber.New()

		mockPersonService := &MockPersonService{}
		mockAgeService := &MockAgeService{}
		mockGenderService := &MockGenderService{}
		mockNationalityService := &MockNationalityService{}

		mockPeopleServices := &MockPeopleServices{
			mockPersonService:      mockPersonService,
			mockAgeService:         mockAgeService,
			mockGenderService:      mockGenderService,
			mockNationalityService: mockNationalityService,
		}

		mockAPI := &MockAPI{
			mockPeopleServices: mockPeopleServices,
		}

		mockPersonRepository := &MockPersonRepository{}
		mockPeopleRepositories := &MockPeopleRepositories{
			mockPersonRepository: mockPersonRepository,
		}

		mockRepositories := &MockRepositories{
			mockPeopleRepositories: mockPeopleRepositories,
		}

		handler := handlers.NewPersonHandler(mockAPI, mockRepositories)
		return app, mockPersonRepository, handler
	}

	createTestPersons := func() []*entities.Person {
		now := time.Now()
		age1 := 25
		age2 := 30
		male := "male"
		female := "female"
		genderProb1 := 0.95
		genderProb2 := 0.98
		ru := "RU"
		us := "US"
		nationProb1 := 0.9
		nationProb2 := 0.85
		patronymic := "Ivanovich"

		return []*entities.Person{
			{
				ID:                     uuid.New(),
				Name:                   "Ivan",
				Surname:                "Petrov",
				Patronymic:             &patronymic,
				Age:                    &age1,
				Gender:                 &male,
				GenderProbability:      &genderProb1,
				Nationality:            &ru,
				NationalityProbability: &nationProb1,
				CreatedAt:              now,
				UpdatedAt:              now,
			},
			{
				ID:                     uuid.New(),
				Name:                   "Maria",
				Surname:                "Smith",
				Age:                    &age2,
				Gender:                 &female,
				GenderProbability:      &genderProb2,
				Nationality:            &us,
				NationalityProbability: &nationProb2,
				CreatedAt:              now,
				UpdatedAt:              now,
			},
		}
	}

	t.Run("should return persons with default pagination", func(t *testing.T) {
		app, mockRepo, handler := setupApp()
		testPersons := createTestPersons()
		totalPersons := len(testPersons)

		mockRepo.On("GetPersons", mock.Anything, mock.Anything, 0, 10).Return(testPersons, totalPersons, nil)

		app.Get("/persons", handler.GetPersons)

		req := httptest.NewRequest(http.MethodGet, "/persons", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should handle custom pagination parameters", func(t *testing.T) {
		app, mockRepo, handler := setupApp()
		testPersons := createTestPersons()
		totalPersons := 100 // Simulate a large dataset

		mockRepo.On("GetPersons", mock.Anything, mock.Anything, 20, 30).Return(testPersons, totalPersons, nil)

		app.Get("/persons", handler.GetPersons)

		req := httptest.NewRequest(http.MethodGet, "/persons?offset=20&limit=30", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should handle invalid pagination parameters", func(t *testing.T) {
		app, mockRepo, handler := setupApp()
		testPersons := createTestPersons()
		totalPersons := len(testPersons)

		mockRepo.On("GetPersons", mock.Anything, mock.Anything, 0, 10).Return(testPersons, totalPersons, nil)

		app.Get("/persons", handler.GetPersons)

		req := httptest.NewRequest(http.MethodGet, "/persons?offset=-5&limit=abc", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should handle zero/negative limit", func(t *testing.T) {
		app, mockRepo, handler := setupApp()
		testPersons := createTestPersons()
		totalPersons := len(testPersons)

		mockRepo.On("GetPersons", mock.Anything, mock.Anything, 0, 10).Return(testPersons, totalPersons, nil)

		app.Get("/persons", handler.GetPersons)

		req := httptest.NewRequest(http.MethodGet, "/persons?limit=0", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should apply name filter", func(t *testing.T) {
		app, mockRepo, handler := setupApp()
		testPersons := []*entities.Person{createTestPersons()[0]} // Only Ivan
		totalPersons := 1

		expectedFilter := map[string]any{"name": "Ivan"}
		mockRepo.On("GetPersons", mock.Anything, expectedFilter, 0, 10).Return(testPersons, totalPersons, nil)

		app.Get("/persons", handler.GetPersons)

		req := httptest.NewRequest(http.MethodGet, "/persons?name=Ivan", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should apply multiple filters including age", func(t *testing.T) {
		app, mockRepo, handler := setupApp()
		testPersons := []*entities.Person{createTestPersons()[0]} // Only Ivan
		totalPersons := 1

		expectedFilter := map[string]any{
			"name":        "Ivan",
			"surname":     "Petrov",
			"gender":      "male",
			"age":         25,
			"nationality": "RU",
		}
		mockRepo.On("GetPersons", mock.Anything, expectedFilter, 0, 10).Return(testPersons, totalPersons, nil)

		app.Get("/persons", handler.GetPersons)

		req := httptest.NewRequest(http.MethodGet, "/persons?name=Ivan&surname=Petrov&gender=male&age=25&nationality=RU", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should handle invalid age parameter", func(t *testing.T) {
		app, mockRepo, handler := setupApp()
		testPersons := createTestPersons()
		totalPersons := len(testPersons)

		expectedFilter := map[string]any{"name": "Ivan"}
		mockRepo.On("GetPersons", mock.Anything, expectedFilter, 0, 10).Return(testPersons, totalPersons, nil)

		app.Get("/persons", handler.GetPersons)

		req := httptest.NewRequest(http.MethodGet, "/persons?name=Ivan&age=notanumber", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should handle repository error", func(t *testing.T) {
		app, mockRepo, handler := setupApp()
		mockRepo.On("GetPersons", mock.Anything, mock.Anything, 0, 10).
			Return([]*entities.Person(nil), 0, errors.New("database error"))

		app.Get("/persons", handler.GetPersons)

		req := httptest.NewRequest(http.MethodGet, "/persons", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return empty array when no persons found", func(t *testing.T) {
		app, mockRepo, handler := setupApp()
		mockRepo.On("GetPersons", mock.Anything, mock.Anything, 0, 10).
			Return([]*entities.Person{}, 0, nil)

		app.Get("/persons", handler.GetPersons)

		req := httptest.NewRequest(http.MethodGet, "/persons", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetPersonsDirectHandler(t *testing.T) {
	t.Run("should handle JSON marshaling error", func(t *testing.T) {
		mockPersonService := &MockPersonService{}
		mockAgeService := &MockAgeService{}
		mockGenderService := &MockGenderService{}
		mockNationalityService := &MockNationalityService{}

		mockPeopleServices := &MockPeopleServices{
			mockPersonService:      mockPersonService,
			mockAgeService:         mockAgeService,
			mockGenderService:      mockGenderService,
			mockNationalityService: mockNationalityService,
		}

		mockAPI := &MockAPI{
			mockPeopleServices: mockPeopleServices,
		}

		mockPersonRepository := &MockPersonRepository{}
		mockPeopleRepositories := &MockPeopleRepositories{
			mockPersonRepository: mockPersonRepository,
		}

		mockRepositories := &MockRepositories{
			mockPeopleRepositories: mockPeopleRepositories,
		}

		testPersons := []*entities.Person{
			{
				ID:      uuid.New(),
				Name:    "Test",
				Surname: "User",
			},
		}

		// Mock repository to return valid data
		mockPersonRepository.On("GetPersons", mock.Anything, mock.Anything, 0, 10).
			Return(testPersons, 1, nil)

		// Create a custom app that forces JSON errors
		app := fiber.New(fiber.Config{
			JSONEncoder: func(v interface{}) ([]byte, error) {
				return nil, errors.New("forced JSON marshal error")
			},
		})

		handler := handlers.NewPersonHandler(mockAPI, mockRepositories)

		var capturedErr error
		app.Get("/persons", func(c fiber.Ctx) error {
			err := handler.GetPersons(c)
			capturedErr = err
			return err
		})

		req := httptest.NewRequest(http.MethodGet, "/persons", nil)
		resp, _ := app.Test(req)

		assert.NotNil(t, capturedErr)
		assert.Contains(t, capturedErr.Error(), "failed to send JSON response")
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		mockPersonRepository.AssertExpectations(t)
	})

	t.Run("should handle JSON response with special characters", func(t *testing.T) {
		mockPersonService := &MockPersonService{}
		mockAgeService := &MockAgeService{}
		mockGenderService := &MockGenderService{}
		mockNationalityService := &MockNationalityService{}

		mockPeopleServices := &MockPeopleServices{
			mockPersonService:      mockPersonService,
			mockAgeService:         mockAgeService,
			mockGenderService:      mockGenderService,
			mockNationalityService: mockNationalityService,
		}

		mockAPI := &MockAPI{
			mockPeopleServices: mockPeopleServices,
		}

		mockPersonRepository := &MockPersonRepository{}
		mockPeopleRepositories := &MockPeopleRepositories{
			mockPersonRepository: mockPersonRepository,
		}

		mockRepositories := &MockRepositories{
			mockPeopleRepositories: mockPeopleRepositories,
		}

		specialChars := "Special\"Characters\n\tWith\"Quotes"
		testPersons := []*entities.Person{
			{
				ID:      uuid.New(),
				Name:    specialChars,
				Surname: specialChars,
			},
		}

		mockPersonRepository.On("GetPersons", mock.Anything, mock.Anything, 0, 10).
			Return(testPersons, 1, nil)

		app := fiber.New()
		handler := handlers.NewPersonHandler(mockAPI, mockRepositories)

		app.Get("/persons", handler.GetPersons)

		req := httptest.NewRequest(http.MethodGet, "/persons", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var responseData map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&responseData)
		require.NoError(t, err)

		persons := responseData["data"].([]interface{})
		assert.Equal(t, 1, len(persons))

		mockPersonRepository.AssertExpectations(t)
	})
}
func TestGetPersonByID(t *testing.T) {
	setupTest := func() (*fiber.App, *MockPersonRepository, *handlers.PersonHandler) {
		app := fiber.New(fiber.Config{
			ErrorHandler: func(c fiber.Ctx, err error) error {
				code := fiber.StatusInternalServerError

				if e, ok := err.(*fiber.Error); ok {
					code = e.Code
				} else {
					if errors.Is(err, handlers.ErrPersonNotFound) {
						code = fiber.StatusNotFound
					}
					if strings.Contains(err.Error(), "invalid UUID") {
						code = fiber.StatusBadRequest
					}
				}

				return c.Status(code).JSON(fiber.Map{
					"error": err.Error(),
				})
			},
		})

		mockPersonService := &MockPersonService{}
		mockAgeService := &MockAgeService{}
		mockGenderService := &MockGenderService{}
		mockNationalityService := &MockNationalityService{}

		mockPeopleServices := &MockPeopleServices{
			mockPersonService:      mockPersonService,
			mockAgeService:         mockAgeService,
			mockGenderService:      mockGenderService,
			mockNationalityService: mockNationalityService,
		}

		mockAPI := &MockAPI{
			mockPeopleServices: mockPeopleServices,
		}

		mockPersonRepository := &MockPersonRepository{}
		mockPeopleRepositories := &MockPeopleRepositories{
			mockPersonRepository: mockPersonRepository,
		}

		mockRepositories := &MockRepositories{
			mockPeopleRepositories: mockPeopleRepositories,
		}

		handler := handlers.NewPersonHandler(mockAPI, mockRepositories)
		return app, mockPersonRepository, handler
	}

	createTestPerson := func() *entities.Person {
		now := time.Now()
		age := 25
		gender := "male"
		genderProb := 0.95
		nationality := "RU"
		nationProb := 0.90
		patronymic := "Ivanovich"

		return &entities.Person{
			ID:                     uuid.New(),
			Name:                   "Ivan",
			Surname:                "Petrov",
			Patronymic:             &patronymic,
			Age:                    &age,
			Gender:                 &gender,
			GenderProbability:      &genderProb,
			Nationality:            &nationality,
			NationalityProbability: &nationProb,
			CreatedAt:              now,
			UpdatedAt:              now,
		}
	}

	t.Run("should return a person when found", func(t *testing.T) {
		app, mockRepo, handler := setupTest()
		testPerson := createTestPerson()

		mockRepo.On("GetByID", mock.Anything, testPerson.ID).Return(testPerson, nil)

		app.Get("/persons/:id", handler.GetPersonByID)

		req := httptest.NewRequest(http.MethodGet, "/persons/"+testPerson.ID.String(), nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		mockRepo.AssertExpectations(t)

		var returnedPerson entities.Person
		err = json.NewDecoder(resp.Body).Decode(&returnedPerson)
		require.NoError(t, err)

		assert.Equal(t, testPerson.ID, returnedPerson.ID)
		assert.Equal(t, testPerson.Name, returnedPerson.Name)
		assert.Equal(t, testPerson.Surname, returnedPerson.Surname)
		assert.Equal(t, *testPerson.Patronymic, *returnedPerson.Patronymic)
		assert.Equal(t, *testPerson.Age, *returnedPerson.Age)
		assert.Equal(t, *testPerson.Gender, *returnedPerson.Gender)
		assert.Equal(t, *testPerson.Nationality, *returnedPerson.Nationality)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		app, _, handler := setupTest()

		app.Get("/persons/:id", func(c fiber.Ctx) error {
			err := handler.GetPersonByID(c)
			if err != nil && strings.Contains(err.Error(), "invalid UUID") {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Invalid UUID format",
				})
			}
			return err
		})

		req := httptest.NewRequest(http.MethodGet, "/persons/not-a-uuid", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var errorResp map[string]string
		err = json.NewDecoder(resp.Body).Decode(&errorResp)
		require.NoError(t, err)
		assert.Equal(t, "Invalid UUID format", errorResp["error"])
	})

	t.Run("should return 404 when person not found", func(t *testing.T) {
		app, mockRepo, handler := setupTest()
		testID := uuid.New()
		notFoundErr := errors.New("person not found: " + testID.String())

		mockRepo.On("GetByID", mock.Anything, testID).Return(nil, notFoundErr)

		app.Get("/persons/:id", func(c fiber.Ctx) error {
			err := handler.GetPersonByID(c)
			if err != nil && strings.Contains(err.Error(), "not found") {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Person not found",
				})
			}
			return err
		})

		req := httptest.NewRequest(http.MethodGet, "/persons/"+testID.String(), nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		mockRepo.AssertExpectations(t)

		var errorResp map[string]string
		err = json.NewDecoder(resp.Body).Decode(&errorResp)
		require.NoError(t, err)
		assert.Equal(t, "Person not found", errorResp["error"])
	})

	t.Run("should return 500 on repository error", func(t *testing.T) {
		app, mockRepo, handler := setupTest()
		testID := uuid.New()
		dbErr := errors.New("database connection error")

		mockRepo.On("GetByID", mock.Anything, testID).Return(nil, dbErr)

		app.Get("/persons/:id", func(c fiber.Ctx) error {
			err := handler.GetPersonByID(c)
			if err != nil && !strings.Contains(err.Error(), "not found") {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to get person",
				})
			}
			return err
		})

		req := httptest.NewRequest(http.MethodGet, "/persons/"+testID.String(), nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		mockRepo.AssertExpectations(t)

		var errorResp map[string]string
		err = json.NewDecoder(resp.Body).Decode(&errorResp)
		require.NoError(t, err)
		assert.Equal(t, "Failed to get person", errorResp["error"])
	})

	t.Run("should handle JSON marshaling error", func(t *testing.T) {
		testID := uuid.New()
		testPerson := createTestPerson()
		testPerson.ID = testID

		mockPersonService := &MockPersonService{}
		mockAgeService := &MockAgeService{}
		mockGenderService := &MockGenderService{}
		mockNationalityService := &MockNationalityService{}

		mockPeopleServices := &MockPeopleServices{
			mockPersonService:      mockPersonService,
			mockAgeService:         mockAgeService,
			mockGenderService:      mockGenderService,
			mockNationalityService: mockNationalityService,
		}

		mockAPI := &MockAPI{
			mockPeopleServices: mockPeopleServices,
		}

		mockPersonRepository := &MockPersonRepository{}
		mockPersonRepository.On("GetByID", mock.Anything, testID).Return(testPerson, nil)

		mockPeopleRepositories := &MockPeopleRepositories{
			mockPersonRepository: mockPersonRepository,
		}

		mockRepositories := &MockRepositories{
			mockPeopleRepositories: mockPeopleRepositories,
		}

		app := fiber.New(fiber.Config{
			JSONEncoder: func(v interface{}) ([]byte, error) {
				return nil, errors.New("forced JSON marshal error")
			},
		})

		handler := handlers.NewPersonHandler(mockAPI, mockRepositories)

		var capturedErr error
		app.Get("/persons/:id", func(c fiber.Ctx) error {
			err := handler.GetPersonByID(c)
			capturedErr = err
			return err
		})

		req := httptest.NewRequest(http.MethodGet, "/persons/"+testID.String(), nil)
		resp, _ := app.Test(req)

		assert.NotNil(t, capturedErr)
		assert.Contains(t, capturedErr.Error(), "failed to send JSON response")
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		mockPersonRepository.AssertExpectations(t)
	})

	t.Run("should handle person with partial data", func(t *testing.T) {
		app, mockRepo, handler := setupTest()

		now := time.Now()
		partialPerson := &entities.Person{
			ID:        uuid.New(),
			Name:      "John",
			Surname:   "Doe",
			CreatedAt: now,
			UpdatedAt: now,
		}

		mockRepo.On("GetByID", mock.Anything, partialPerson.ID).Return(partialPerson, nil)

		app.Get("/persons/:id", handler.GetPersonByID)

		req := httptest.NewRequest(http.MethodGet, "/persons/"+partialPerson.ID.String(), nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		mockRepo.AssertExpectations(t)

		var returnedPerson entities.Person
		err = json.NewDecoder(resp.Body).Decode(&returnedPerson)
		require.NoError(t, err)

		assert.Equal(t, partialPerson.ID, returnedPerson.ID)
		assert.Equal(t, partialPerson.Name, returnedPerson.Name)
		assert.Equal(t, partialPerson.Surname, returnedPerson.Surname)
		assert.Nil(t, returnedPerson.Patronymic)
		assert.Nil(t, returnedPerson.Age)
		assert.Nil(t, returnedPerson.Gender)
		assert.Nil(t, returnedPerson.GenderProbability)
		assert.Nil(t, returnedPerson.Nationality)
		assert.Nil(t, returnedPerson.NationalityProbability)
	})
}

func TestCreatePerson(t *testing.T) {
	setupTest := func() (*fiber.App, *MockPersonRepository, *handlers.PersonHandler) {
		app := fiber.New(fiber.Config{
			ErrorHandler: func(c fiber.Ctx, err error) error {
				code := fiber.StatusInternalServerError
				errMsg := "Internal Server Error"

				if fiber.IsChild() {
					return fiber.DefaultErrorHandler(c, err)
				}

				if err != nil {
					if errors.Is(err, handlers.ErrNameSurnameRequired) {
						code = fiber.StatusBadRequest
						errMsg = "Name and surname are required"
					} else if strings.Contains(err.Error(), "invalid request body") {
						code = fiber.StatusBadRequest
						errMsg = "Invalid request body"
					}
				}

				return c.Status(code).JSON(fiber.Map{
					"error": errMsg,
				})
			},
		})

		mockPersonService := &MockPersonService{}
		mockAgeService := &MockAgeService{}
		mockGenderService := &MockGenderService{}
		mockNationalityService := &MockNationalityService{}

		mockPeopleServices := &MockPeopleServices{
			mockPersonService:      mockPersonService,
			mockAgeService:         mockAgeService,
			mockGenderService:      mockGenderService,
			mockNationalityService: mockNationalityService,
		}

		mockAPI := &MockAPI{
			mockPeopleServices: mockPeopleServices,
		}

		mockPersonRepository := &MockPersonRepository{}
		mockPeopleRepositories := &MockPeopleRepositories{
			mockPersonRepository: mockPersonRepository,
		}

		mockRepositories := &MockRepositories{
			mockPeopleRepositories: mockPeopleRepositories,
		}

		handler := handlers.NewPersonHandler(mockAPI, mockRepositories)
		return app, mockPersonRepository, handler
	}

	t.Run("should successfully create a person", func(t *testing.T) {
		app, mockRepo, handler := setupTest()
		app.Post("/persons", handler.CreatePerson)

		personID := uuid.New()
		personData := map[string]interface{}{
			"id":      personID.String(),
			"name":    "John",
			"surname": "Doe",
		}

		expectedPerson := &entities.Person{
			ID:      personID,
			Name:    "John",
			Surname: "Doe",
		}

		mockRepo.On("CreatePerson", mock.Anything, mock.MatchedBy(func(p *entities.Person) bool {
			return p.ID == personID && p.Name == "John" && p.Surname == "Doe"
		})).Return(nil)

		bodyBytes, err := json.Marshal(personData)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/persons", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var returnedPerson entities.Person
		err = json.NewDecoder(resp.Body).Decode(&returnedPerson)
		require.NoError(t, err)

		assert.Equal(t, expectedPerson.ID, returnedPerson.ID)
		assert.Equal(t, expectedPerson.Name, returnedPerson.Name)
		assert.Equal(t, expectedPerson.Surname, returnedPerson.Surname)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should generate UUID if not provided", func(t *testing.T) {
		app, mockRepo, handler := setupTest()
		app.Post("/persons", handler.CreatePerson)

		personData := map[string]interface{}{
			"name":    "Jane",
			"surname": "Smith",
		}

		var capturedPerson *entities.Person
		mockRepo.On("CreatePerson", mock.Anything, mock.MatchedBy(func(p *entities.Person) bool {
			capturedPerson = p
			return p.ID != uuid.Nil && p.Name == "Jane" && p.Surname == "Smith"
		})).Return(nil)

		bodyBytes, err := json.Marshal(personData)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/persons", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var returnedPerson entities.Person
		err = json.NewDecoder(resp.Body).Decode(&returnedPerson)
		require.NoError(t, err)

		assert.NotEqual(t, uuid.Nil, returnedPerson.ID)
		assert.Equal(t, "Jane", returnedPerson.Name)
		assert.Equal(t, "Smith", returnedPerson.Surname)

		mockRepo.AssertExpectations(t)
		assert.NotEqual(t, uuid.Nil, capturedPerson.ID)
	})

	t.Run("should return 400 for invalid request body", func(t *testing.T) {
		app, _, handler := setupTest()

		app.Post("/persons", func(c fiber.Ctx) error {
			err := handler.CreatePerson(c)
			if err != nil && strings.Contains(err.Error(), "invalid request body") {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Invalid request body",
				})
			}
			return err
		})

		invalidJSON := `{"name": "John", "surname": "Doe", invalid_json}`

		req := httptest.NewRequest(http.MethodPost, "/persons", bytes.NewReader([]byte(invalidJSON)))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), "Invalid request body")
	})

	t.Run("should return 400 when name is missing", func(t *testing.T) {
		app, _, handler := setupTest()

		app.Post("/persons", func(c fiber.Ctx) error {
			err := handler.CreatePerson(c)
			if errors.Is(err, handlers.ErrNameSurnameRequired) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Name and surname are required",
				})
			}
			return err
		})

		personData := map[string]interface{}{
			"surname": "Doe",
		}

		bodyBytes, err := json.Marshal(personData)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/persons", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), "Name and surname are required")
	})

	t.Run("should return 400 when surname is missing", func(t *testing.T) {
		app, _, handler := setupTest()

		app.Post("/persons", func(c fiber.Ctx) error {
			err := handler.CreatePerson(c)
			if errors.Is(err, handlers.ErrNameSurnameRequired) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Name and surname are required",
				})
			}
			return err
		})

		personData := map[string]interface{}{
			"name": "John",
		}

		bodyBytes, err := json.Marshal(personData)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/persons", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), "Name and surname are required")
	})

	t.Run("should return 500 when repository fails", func(t *testing.T) {
		app, mockRepo, handler := setupTest()

		app.Post("/persons", func(c fiber.Ctx) error {
			err := handler.CreatePerson(c)
			if err != nil && strings.Contains(err.Error(), "failed to create person") {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to create person",
				})
			}
			return err
		})

		personData := map[string]interface{}{
			"name":    "John",
			"surname": "Doe",
		}

		mockRepo.On("CreatePerson", mock.Anything, mock.Anything).Return(errors.New("database error"))

		bodyBytes, err := json.Marshal(personData)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/persons", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), "Failed to create person")

		mockRepo.AssertExpectations(t)
	})

	t.Run("should handle JSON marshaling error on response", func(t *testing.T) {
		app := fiber.New(fiber.Config{
			JSONEncoder: func(v interface{}) ([]byte, error) {
				return nil, errors.New("forced JSON marshal error")
			},
		})

		mockPersonService := &MockPersonService{}
		mockAgeService := &MockAgeService{}
		mockGenderService := &MockGenderService{}
		mockNationalityService := &MockNationalityService{}

		mockPeopleServices := &MockPeopleServices{
			mockPersonService:      mockPersonService,
			mockAgeService:         mockAgeService,
			mockGenderService:      mockGenderService,
			mockNationalityService: mockNationalityService,
		}

		mockAPI := &MockAPI{
			mockPeopleServices: mockPeopleServices,
		}

		mockPersonRepository := &MockPersonRepository{}

		mockPersonRepository.On("CreatePerson", mock.Anything, mock.Anything).Return(nil)

		mockPeopleRepositories := &MockPeopleRepositories{
			mockPersonRepository: mockPersonRepository,
		}

		mockRepositories := &MockRepositories{
			mockPeopleRepositories: mockPeopleRepositories,
		}

		handler := handlers.NewPersonHandler(mockAPI, mockRepositories)

		// Create valid person data
		personData := map[string]interface{}{
			"name":    "John",
			"surname": "Doe",
		}

		var capturedErr error
		app.Post("/persons", func(c fiber.Ctx) error {
			err := handler.CreatePerson(c)
			capturedErr = err
			return err
		})

		bodyBytes, err := json.Marshal(personData)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/persons", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req)

		assert.NotNil(t, capturedErr)
		assert.Contains(t, capturedErr.Error(), "failed to send JSON response")
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		mockPersonRepository.AssertExpectations(t)
	})

	t.Run("should create a person with all optional fields", func(t *testing.T) {
		app, mockRepo, handler := setupTest()
		app.Post("/persons", handler.CreatePerson)

		now := time.Now()
		age := 30
		gender := "female"
		genderProb := 0.95
		nationality := "US"
		nationProb := 0.85
		patronymic := "Alex"

		personID := uuid.New()
		personData := map[string]interface{}{
			"id":                      personID.String(),
			"name":                    "Jane",
			"surname":                 "Smith",
			"patronymic":              patronymic,
			"age":                     age,
			"gender":                  gender,
			"gender_probability":      genderProb,
			"nationality":             nationality,
			"nationality_probability": nationProb,
			"created_at":              now,
			"updated_at":              now,
		}

		mockRepo.On("CreatePerson", mock.Anything, mock.MatchedBy(func(p *entities.Person) bool {
			return p.ID == personID &&
				p.Name == "Jane" &&
				p.Surname == "Smith" &&
				*p.Patronymic == patronymic &&
				*p.Age == age &&
				*p.Gender == gender &&
				*p.GenderProbability == genderProb &&
				*p.Nationality == nationality &&
				*p.NationalityProbability == nationProb
		})).Return(nil)

		bodyBytes, err := json.Marshal(personData)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/persons", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var returnedPerson entities.Person
		err = json.NewDecoder(resp.Body).Decode(&returnedPerson)
		require.NoError(t, err)

		assert.Equal(t, personID, returnedPerson.ID)
		assert.Equal(t, "Jane", returnedPerson.Name)
		assert.Equal(t, "Smith", returnedPerson.Surname)
		assert.Equal(t, patronymic, *returnedPerson.Patronymic)
		assert.Equal(t, age, *returnedPerson.Age)
		assert.Equal(t, gender, *returnedPerson.Gender)
		assert.Equal(t, genderProb, *returnedPerson.GenderProbability)
		assert.Equal(t, nationality, *returnedPerson.Nationality)
		assert.Equal(t, nationProb, *returnedPerson.NationalityProbability)

		mockRepo.AssertExpectations(t)
	})
}

func TestUpdatePerson(t *testing.T) {
	setupTest := func() (*fiber.App, *MockPersonRepository, *handlers.PersonHandler) {
		app := fiber.New(fiber.Config{
			ErrorHandler: func(c fiber.Ctx, err error) error {
				code := fiber.StatusInternalServerError
				errMsg := "Internal Server Error"

				if fiber.IsChild() {
					return fiber.DefaultErrorHandler(c, err)
				}

				if err != nil {
					if errors.Is(err, handlers.ErrNameSurnameRequired) {
						code = fiber.StatusBadRequest
						errMsg = "Name and surname are required"
					} else if errors.Is(err, handlers.ErrPersonNotFound) {
						code = fiber.StatusNotFound
						errMsg = "Person not found"
					} else if strings.Contains(err.Error(), "invalid request body") {
						code = fiber.StatusBadRequest
						errMsg = "Invalid request body"
					} else if strings.Contains(err.Error(), "invalid UUID") {
						code = fiber.StatusBadRequest
						errMsg = "Invalid UUID format"
					}
				}

				return c.Status(code).JSON(fiber.Map{
					"error": errMsg,
				})
			},
		})

		mockPersonService := &MockPersonService{}
		mockAgeService := &MockAgeService{}
		mockGenderService := &MockGenderService{}
		mockNationalityService := &MockNationalityService{}

		mockPeopleServices := &MockPeopleServices{
			mockPersonService:      mockPersonService,
			mockAgeService:         mockAgeService,
			mockGenderService:      mockGenderService,
			mockNationalityService: mockNationalityService,
		}

		mockAPI := &MockAPI{
			mockPeopleServices: mockPeopleServices,
		}

		mockPersonRepository := &MockPersonRepository{}
		mockPeopleRepositories := &MockPeopleRepositories{
			mockPersonRepository: mockPersonRepository,
		}

		mockRepositories := &MockRepositories{
			mockPeopleRepositories: mockPeopleRepositories,
		}

		handler := handlers.NewPersonHandler(mockAPI, mockRepositories)
		return app, mockPersonRepository, handler
	}

	createTestPerson := func() *entities.Person {
		now := time.Now()
		age := 25
		gender := "male"
		genderProb := 0.95
		nationality := "RU"
		nationProb := 0.90
		patronymic := "Ivanovich"

		return &entities.Person{
			ID:                     uuid.New(),
			Name:                   "Ivan",
			Surname:                "Petrov",
			Patronymic:             &patronymic,
			Age:                    &age,
			Gender:                 &gender,
			GenderProbability:      &genderProb,
			Nationality:            &nationality,
			NationalityProbability: &nationProb,
			CreatedAt:              now,
			UpdatedAt:              now,
		}
	}

	t.Run("should successfully update a person", func(t *testing.T) {
		app, mockRepo, handler := setupTest()
		personID := uuid.New()

		mockRepo.On("ExistsByID", mock.Anything, personID).Return(true, nil)

		mockRepo.On("UpdatePerson", mock.Anything, mock.MatchedBy(func(p *entities.Person) bool {
			return p.ID == personID &&
				p.Name == "John" &&
				p.Surname == "Smith"
		})).Return(nil)

		app.Put("/persons/:id", handler.UpdatePerson)

		updateData := map[string]interface{}{
			"name":    "John",
			"surname": "Smith",
		}

		bodyBytes, err := json.Marshal(updateData)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPut, "/persons/"+personID.String(), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var returnedPerson entities.Person
		err = json.NewDecoder(resp.Body).Decode(&returnedPerson)
		require.NoError(t, err)

		assert.Equal(t, personID, returnedPerson.ID)
		assert.Equal(t, "John", returnedPerson.Name)
		assert.Equal(t, "Smith", returnedPerson.Surname)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		app, _, handler := setupTest()

		app.Put("/persons/:id", handler.UpdatePerson)

		updateData := map[string]interface{}{
			"name":    "John",
			"surname": "Smith",
		}

		bodyBytes, err := json.Marshal(updateData)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPut, "/persons/not-a-uuid", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var errorResp map[string]string
		err = json.NewDecoder(resp.Body).Decode(&errorResp)
		require.NoError(t, err)
		assert.Equal(t, "Invalid UUID format", errorResp["error"])
	})

	t.Run("should return 404 when person not found", func(t *testing.T) {
		app, mockRepo, handler := setupTest()
		personID := uuid.New()

		mockRepo.On("ExistsByID", mock.Anything, personID).Return(false, nil)

		app.Put("/persons/:id", handler.UpdatePerson)

		updateData := map[string]interface{}{
			"name":    "John",
			"surname": "Smith",
		}

		bodyBytes, err := json.Marshal(updateData)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPut, "/persons/"+personID.String(), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		var errorResp map[string]string
		err = json.NewDecoder(resp.Body).Decode(&errorResp)
		require.NoError(t, err)
		assert.Equal(t, "Person not found", errorResp["error"])

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return 400 for invalid request body", func(t *testing.T) {
		app, mockRepo, handler := setupTest()
		personID := uuid.New()

		mockRepo.On("ExistsByID", mock.Anything, personID).Return(true, nil)

		app.Put("/persons/:id", handler.UpdatePerson)

		invalidJSON := `{"name": "John", "surname": "Smith", invalid_json}`

		req := httptest.NewRequest(http.MethodPut, "/persons/"+personID.String(), bytes.NewReader([]byte(invalidJSON)))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var errorResp map[string]string
		err = json.NewDecoder(resp.Body).Decode(&errorResp)
		require.NoError(t, err)
		assert.Equal(t, "Invalid request body", errorResp["error"])

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return 400 when name is missing", func(t *testing.T) {
		app, mockRepo, handler := setupTest()
		personID := uuid.New()

		mockRepo.On("ExistsByID", mock.Anything, personID).Return(true, nil)

		app.Put("/persons/:id", handler.UpdatePerson)

		updateData := map[string]interface{}{
			"surname": "Smith",
		}

		bodyBytes, err := json.Marshal(updateData)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPut, "/persons/"+personID.String(), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var errorResp map[string]string
		err = json.NewDecoder(resp.Body).Decode(&errorResp)
		require.NoError(t, err)
		assert.Equal(t, "Name and surname are required", errorResp["error"])

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return 400 when surname is missing", func(t *testing.T) {
		app, mockRepo, handler := setupTest()
		personID := uuid.New()

		mockRepo.On("ExistsByID", mock.Anything, personID).Return(true, nil)

		app.Put("/persons/:id", handler.UpdatePerson)

		updateData := map[string]interface{}{
			"name": "John",
		}

		bodyBytes, err := json.Marshal(updateData)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPut, "/persons/"+personID.String(), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var errorResp map[string]string
		err = json.NewDecoder(resp.Body).Decode(&errorResp)
		require.NoError(t, err)
		assert.Equal(t, "Name and surname are required", errorResp["error"])

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return 500 when ExistsByID fails", func(t *testing.T) {
		app, mockRepo, handler := setupTest()
		personID := uuid.New()

		mockRepo.On("ExistsByID", mock.Anything, personID).Return(false, errors.New("database error"))

		app.Put("/persons/:id", func(c fiber.Ctx) error {
			err := handler.UpdatePerson(c)
			if err != nil && strings.Contains(err.Error(), "failed to check if person exists") {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to check if person exists",
				})
			}
			return err
		})

		updateData := map[string]interface{}{
			"name":    "John",
			"surname": "Smith",
		}

		bodyBytes, err := json.Marshal(updateData)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPut, "/persons/"+personID.String(), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		var errorResp map[string]string
		err = json.NewDecoder(resp.Body).Decode(&errorResp)
		require.NoError(t, err)
		assert.Equal(t, "Failed to check if person exists", errorResp["error"])

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return 500 when UpdatePerson fails", func(t *testing.T) {
		app, mockRepo, handler := setupTest()
		personID := uuid.New()

		mockRepo.On("ExistsByID", mock.Anything, personID).Return(true, nil)

		mockRepo.On("UpdatePerson", mock.Anything, mock.MatchedBy(func(p *entities.Person) bool {
			return p.ID == personID && p.Name == "John" && p.Surname == "Smith"
		})).Return(errors.New("database error"))

		app.Put("/persons/:id", func(c fiber.Ctx) error {
			err := handler.UpdatePerson(c)
			if err != nil && strings.Contains(err.Error(), "failed to update person") {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to update person",
				})
			}
			return err
		})

		updateData := map[string]interface{}{
			"name":    "John",
			"surname": "Smith",
		}

		bodyBytes, err := json.Marshal(updateData)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPut, "/persons/"+personID.String(), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		var errorResp map[string]string
		err = json.NewDecoder(resp.Body).Decode(&errorResp)
		require.NoError(t, err)
		assert.Equal(t, "Failed to update person", errorResp["error"])

		mockRepo.AssertExpectations(t)
	})

	t.Run("should handle JSON marshaling error", func(t *testing.T) {
		personID := uuid.New()

		mockPersonService := &MockPersonService{}
		mockAgeService := &MockAgeService{}
		mockGenderService := &MockGenderService{}
		mockNationalityService := &MockNationalityService{}

		mockPeopleServices := &MockPeopleServices{
			mockPersonService:      mockPersonService,
			mockAgeService:         mockAgeService,
			mockGenderService:      mockGenderService,
			mockNationalityService: mockNationalityService,
		}

		mockAPI := &MockAPI{
			mockPeopleServices: mockPeopleServices,
		}

		mockPersonRepository := &MockPersonRepository{}
		mockPersonRepository.On("ExistsByID", mock.Anything, personID).Return(true, nil)
		mockPersonRepository.On("UpdatePerson", mock.Anything, mock.Anything).Return(nil)

		mockPeopleRepositories := &MockPeopleRepositories{
			mockPersonRepository: mockPersonRepository,
		}

		mockRepositories := &MockRepositories{
			mockPeopleRepositories: mockPeopleRepositories,
		}

		app := fiber.New(fiber.Config{
			JSONEncoder: func(v interface{}) ([]byte, error) {
				return nil, errors.New("forced JSON marshal error")
			},
		})

		handler := handlers.NewPersonHandler(mockAPI, mockRepositories)

		var capturedErr error
		app.Put("/persons/:id", func(c fiber.Ctx) error {
			err := handler.UpdatePerson(c)
			capturedErr = err
			return err
		})

		updateData := map[string]interface{}{
			"name":    "John",
			"surname": "Smith",
		}

		bodyBytes, err := json.Marshal(updateData)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPut, "/persons/"+personID.String(), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req)

		assert.NotNil(t, capturedErr)
		assert.Contains(t, capturedErr.Error(), "failed to send JSON response")
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		mockPersonRepository.AssertExpectations(t)
	})

	t.Run("should update person with all optional fields", func(t *testing.T) {
		app, mockRepo, handler := setupTest()
		personID := uuid.New()

		mockRepo.On("ExistsByID", mock.Anything, personID).Return(true, nil)

		now := time.Now()
		age := 30
		gender := "female"
		genderProb := 0.95
		nationality := "US"
		nationProb := 0.85
		patronymic := "Alex"

		updateData := map[string]interface{}{
			"name":                    "Jane",
			"surname":                 "Smith",
			"patronymic":              patronymic,
			"age":                     age,
			"gender":                  gender,
			"gender_probability":      genderProb,
			"nationality":             nationality,
			"nationality_probability": nationProb,
			"created_at":              now,
			"updated_at":              now,
		}

		mockRepo.On("UpdatePerson", mock.Anything, mock.MatchedBy(func(p *entities.Person) bool {
			return p.ID == personID &&
				p.Name == "Jane" &&
				p.Surname == "Smith" &&
				*p.Patronymic == patronymic &&
				*p.Age == age &&
				*p.Gender == gender &&
				*p.GenderProbability == genderProb &&
				*p.Nationality == nationality &&
				*p.NationalityProbability == nationProb
		})).Return(nil)

		app.Put("/persons/:id", handler.UpdatePerson)

		bodyBytes, err := json.Marshal(updateData)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPut, "/persons/"+personID.String(), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var returnedPerson entities.Person
		err = json.NewDecoder(resp.Body).Decode(&returnedPerson)
		require.NoError(t, err)

		assert.Equal(t, personID, returnedPerson.ID)
		assert.Equal(t, "Jane", returnedPerson.Name)
		assert.Equal(t, "Smith", returnedPerson.Surname)
		assert.Equal(t, patronymic, *returnedPerson.Patronymic)
		assert.Equal(t, age, *returnedPerson.Age)
		assert.Equal(t, gender, *returnedPerson.Gender)
		assert.Equal(t, genderProb, *returnedPerson.GenderProbability)
		assert.Equal(t, nationality, *returnedPerson.Nationality)
		assert.Equal(t, nationProb, *returnedPerson.NationalityProbability)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should use existing person as template for update", func(t *testing.T) {
		app, mockRepo, handler := setupTest()
		testPerson := createTestPerson()
		personID := testPerson.ID

		mockRepo.On("ExistsByID", mock.Anything, personID).Return(true, nil)

		updateData := map[string]interface{}{
			"name":    "Updated",
			"surname": testPerson.Surname,
		}

		mockRepo.On("UpdatePerson", mock.Anything, mock.MatchedBy(func(p *entities.Person) bool {
			return p.ID == personID &&
				p.Name == "Updated" &&
				p.Surname == testPerson.Surname
		})).Return(nil)

		app.Put("/persons/:id", handler.UpdatePerson)

		bodyBytes, err := json.Marshal(updateData)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPut, "/persons/"+personID.String(), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var returnedPerson entities.Person
		err = json.NewDecoder(resp.Body).Decode(&returnedPerson)
		require.NoError(t, err)

		assert.Equal(t, personID, returnedPerson.ID)
		assert.Equal(t, "Updated", returnedPerson.Name)
		assert.Equal(t, testPerson.Surname, returnedPerson.Surname)

		mockRepo.AssertExpectations(t)
	})
}

func TestDeletePerson(t *testing.T) {
	setupTest := func() (*fiber.App, *MockPersonRepository, *handlers.PersonHandler) {
		app := fiber.New(fiber.Config{
			ErrorHandler: func(c fiber.Ctx, err error) error {
				code := fiber.StatusInternalServerError
				errMsg := "Internal Server Error"

				if fiber.IsChild() {
					return fiber.DefaultErrorHandler(c, err)
				}

				if err != nil {
					if strings.Contains(err.Error(), "not found") {
						code = fiber.StatusNotFound
						errMsg = "Person not found"
					} else if strings.Contains(err.Error(), "invalid UUID") {
						code = fiber.StatusBadRequest
						errMsg = "Invalid UUID format"
					}
				}

				return c.Status(code).JSON(fiber.Map{
					"error": errMsg,
				})
			},
		})

		mockPersonService := &MockPersonService{}
		mockAgeService := &MockAgeService{}
		mockGenderService := &MockGenderService{}
		mockNationalityService := &MockNationalityService{}

		mockPeopleServices := &MockPeopleServices{
			mockPersonService:      mockPersonService,
			mockAgeService:         mockAgeService,
			mockGenderService:      mockGenderService,
			mockNationalityService: mockNationalityService,
		}

		mockAPI := &MockAPI{
			mockPeopleServices: mockPeopleServices,
		}

		mockPersonRepository := &MockPersonRepository{}
		mockPeopleRepositories := &MockPeopleRepositories{
			mockPersonRepository: mockPersonRepository,
		}

		mockRepositories := &MockRepositories{
			mockPeopleRepositories: mockPeopleRepositories,
		}

		handler := handlers.NewPersonHandler(mockAPI, mockRepositories)
		return app, mockPersonRepository, handler
	}

	t.Run("should successfully delete a person", func(t *testing.T) {
		app, mockRepo, handler := setupTest()
		personID := uuid.New()

		mockRepo.On("DeletePerson", mock.Anything, personID).Return(nil)

		app.Delete("/persons/:id", handler.DeletePerson)

		req := httptest.NewRequest(http.MethodDelete, "/persons/"+personID.String(), nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		assert.Equal(t, 0, len(resp.Header.Get("Content-Type")), "Content-Type header should be empty for 204 responses")
		assert.Equal(t, int64(0), resp.ContentLength, "Content length should be 0 for 204 responses")

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		app, _, handler := setupTest()

		app.Delete("/persons/:id", func(c fiber.Ctx) error {
			err := handler.DeletePerson(c)
			if err != nil && strings.Contains(err.Error(), "invalid UUID") {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Invalid UUID format",
				})
			}
			return err
		})

		req := httptest.NewRequest(http.MethodDelete, "/persons/not-a-uuid", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "application/json")

		body := make(map[string]string)
		err = json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "Invalid UUID format", body["error"])
	})

	t.Run("should return 404 when person not found", func(t *testing.T) {
		app, mockRepo, handler := setupTest()
		personID := uuid.New()
		notFoundErr := errors.New("person not found: " + personID.String())

		mockRepo.On("DeletePerson", mock.Anything, personID).Return(notFoundErr)

		app.Delete("/persons/:id", func(c fiber.Ctx) error {
			err := handler.DeletePerson(c)
			if err != nil && strings.Contains(err.Error(), "not found") {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Person not found",
				})
			}
			return err
		})

		req := httptest.NewRequest(http.MethodDelete, "/persons/"+personID.String(), nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		body := make(map[string]string)
		err = json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "Person not found", body["error"])

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return 500 when repository error occurs", func(t *testing.T) {
		app, mockRepo, handler := setupTest()
		personID := uuid.New()
		dbErr := errors.New("database connection error")

		mockRepo.On("DeletePerson", mock.Anything, personID).Return(dbErr)

		app.Delete("/persons/:id", func(c fiber.Ctx) error {
			err := handler.DeletePerson(c)
			if err != nil && !strings.Contains(err.Error(), "not found") {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to delete person",
				})
			}
			return err
		})

		req := httptest.NewRequest(http.MethodDelete, "/persons/"+personID.String(), nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		body := make(map[string]string)
		err = json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "Failed to delete person", body["error"])

		mockRepo.AssertExpectations(t)
	})

	t.Run("should handle response error when person is not found", func(t *testing.T) {
		app := fiber.New(fiber.Config{
			JSONEncoder: func(v interface{}) ([]byte, error) {
				return nil, errors.New("forced JSON marshal error")
			},
		})

		mockPersonRepository := &MockPersonRepository{}
		personID := uuid.New()
		notFoundErr := errors.New("person not found: " + personID.String())

		mockPersonRepository.On("DeletePerson", mock.Anything, personID).Return(notFoundErr)

		mockPeopleRepositories := &MockPeopleRepositories{
			mockPersonRepository: mockPersonRepository,
		}

		mockRepositories := &MockRepositories{
			mockPeopleRepositories: mockPeopleRepositories,
		}

		mockAPI := &MockAPI{
			mockPeopleServices: &MockPeopleServices{},
		}

		handler := handlers.NewPersonHandler(mockAPI, mockRepositories)

		var capturedErr error
		app.Delete("/persons/:id", func(c fiber.Ctx) error {
			err := handler.DeletePerson(c)
			capturedErr = err
			return err
		})

		req := httptest.NewRequest(http.MethodDelete, "/persons/"+personID.String(), nil)
		resp, _ := app.Test(req)

		assert.NotNil(t, capturedErr)
		assert.Contains(t, capturedErr.Error(), "failed to send JSON response")
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		mockPersonRepository.AssertExpectations(t)
	})

	t.Run("should handle response error when repository fails", func(t *testing.T) {
		app := fiber.New(fiber.Config{
			JSONEncoder: func(v interface{}) ([]byte, error) {
				return nil, errors.New("forced JSON marshal error")
			},
		})

		mockPersonRepository := &MockPersonRepository{}
		personID := uuid.New()
		dbErr := errors.New("database error")

		mockPersonRepository.On("DeletePerson", mock.Anything, personID).Return(dbErr)

		mockPeopleRepositories := &MockPeopleRepositories{
			mockPersonRepository: mockPersonRepository,
		}

		mockRepositories := &MockRepositories{
			mockPeopleRepositories: mockPeopleRepositories,
		}

		mockAPI := &MockAPI{
			mockPeopleServices: &MockPeopleServices{},
		}

		handler := handlers.NewPersonHandler(mockAPI, mockRepositories)

		var capturedErr error
		app.Delete("/persons/:id", func(c fiber.Ctx) error {
			err := handler.DeletePerson(c)
			capturedErr = err
			return err
		})

		req := httptest.NewRequest(http.MethodDelete, "/persons/"+personID.String(), nil)
		resp, _ := app.Test(req)

		assert.NotNil(t, capturedErr)
		assert.Contains(t, capturedErr.Error(), "failed to send JSON response")
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		mockPersonRepository.AssertExpectations(t)
	})
}

func TestEnrichPerson(t *testing.T) {
	setupTest := func() (*fiber.App, *MockPersonRepository, *MockAgeService, *MockGenderService, *MockNationalityService, *handlers.PersonHandler) {
		app := fiber.New(fiber.Config{
			ErrorHandler: func(c fiber.Ctx, err error) error {
				code := fiber.StatusInternalServerError
				errMsg := "Internal Server Error"

				if fiber.IsChild() {
					return fiber.DefaultErrorHandler(c, err)
				}

				if err != nil {
					if strings.Contains(err.Error(), "invalid UUID format") {
						code = fiber.StatusBadRequest
						errMsg = "Invalid UUID format"
					} else if strings.Contains(err.Error(), "person not found") {
						code = fiber.StatusNotFound
						errMsg = "Person not found"
					} else if strings.Contains(err.Error(), "failed to get person") {
						code = fiber.StatusInternalServerError
						errMsg = "Failed to get person"
					} else if strings.Contains(err.Error(), "failed to save enriched data") {
						code = fiber.StatusInternalServerError
						errMsg = "Failed to save enriched data"
					}
				}

				return c.Status(code).JSON(fiber.Map{
					"error": errMsg,
				})
			},
		})

		mockPersonService := &MockPersonService{}
		mockAgeService := &MockAgeService{}
		mockGenderService := &MockGenderService{}
		mockNationalityService := &MockNationalityService{}

		mockPeopleServices := &MockPeopleServices{
			mockPersonService:      mockPersonService,
			mockAgeService:         mockAgeService,
			mockGenderService:      mockGenderService,
			mockNationalityService: mockNationalityService,
		}

		mockAPI := &MockAPI{
			mockPeopleServices: mockPeopleServices,
		}

		mockPersonRepository := &MockPersonRepository{}
		mockPeopleRepositories := &MockPeopleRepositories{
			mockPersonRepository: mockPersonRepository,
		}

		mockRepositories := &MockRepositories{
			mockPeopleRepositories: mockPeopleRepositories,
		}

		handler := handlers.NewPersonHandler(mockAPI, mockRepositories)

		return app, mockPersonRepository, mockAgeService, mockGenderService, mockNationalityService, handler
	}

	createPersonWithoutEnrichment := func() *entities.Person {
		now := time.Now()
		return &entities.Person{
			ID:        uuid.New(),
			Name:      "Ivan",
			Surname:   "Petrov",
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		app, _, _, _, _, handler := setupTest()

		app.Put("/persons/:id/enrich", func(c fiber.Ctx) error {
			err := handler.EnrichPerson(c)
			if err != nil && strings.Contains(err.Error(), "invalid UUID format") {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Invalid UUID format",
				})
			}
			return err
		})

		req := httptest.NewRequest(http.MethodPut, "/persons/not-a-valid-uuid/enrich", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var errorResp map[string]string
		err = json.NewDecoder(resp.Body).Decode(&errorResp)
		require.NoError(t, err)
		assert.Equal(t, "Invalid UUID format", errorResp["error"])
	})

	t.Run("should return 404 when person not found", func(t *testing.T) {
		app, mockRepo, _, _, _, handler := setupTest()
		personID := uuid.New()

		mockRepo.On("GetByID", mock.Anything, personID).Return(nil, errors.New("person not found"))

		app.Put("/persons/:id/enrich", func(c fiber.Ctx) error {
			err := handler.EnrichPerson(c)
			if err != nil && strings.Contains(err.Error(), "person not found") {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Person not found",
				})
			}
			return err
		})

		req := httptest.NewRequest(http.MethodPut, "/persons/"+personID.String()+"/enrich", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		var errorResp map[string]string
		err = json.NewDecoder(resp.Body).Decode(&errorResp)
		require.NoError(t, err)
		assert.Equal(t, "Person not found", errorResp["error"])

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return 500 when GetByID fails with generic error", func(t *testing.T) {
		app, mockRepo, _, _, _, handler := setupTest()
		personID := uuid.New()

		mockRepo.On("GetByID", mock.Anything, personID).Return(nil, errors.New("database error"))

		app.Put("/persons/:id/enrich", func(c fiber.Ctx) error {
			err := handler.EnrichPerson(c)
			if err != nil && strings.Contains(err.Error(), "failed to get person") {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to get person",
				})
			}
			return err
		})

		req := httptest.NewRequest(http.MethodPut, "/persons/"+personID.String()+"/enrich", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		var errorResp map[string]string
		err = json.NewDecoder(resp.Body).Decode(&errorResp)
		require.NoError(t, err)
		assert.Equal(t, "Failed to get person", errorResp["error"])

		mockRepo.AssertExpectations(t)
	})

	t.Run("should enrich person with all three services", func(t *testing.T) {
		app, mockRepo, mockAgeService, mockGenderService, mockNationalityService, handler := setupTest()
		person := createPersonWithoutEnrichment()
		personID := person.ID

		expectedAge := 30
		expectedGender := "male"
		expectedGenderProb := 0.95
		expectedNationality := "RU"
		expectedNationalityProb := 0.90

		mockRepo.On("GetByID", mock.Anything, personID).Return(person, nil)
		mockRepo.On("UpdatePerson", mock.Anything, mock.MatchedBy(func(p *entities.Person) bool {
			return p.ID == personID &&
				*p.Age == expectedAge &&
				*p.Gender == expectedGender &&
				*p.GenderProbability == expectedGenderProb &&
				*p.Nationality == expectedNationality &&
				*p.NationalityProbability == expectedNationalityProb
		})).Return(nil)

		mockAgeService.On("GetAgeByName", mock.Anything, person.Name).Return(expectedAge, 0.85, nil)
		mockGenderService.On("GetGenderByName", mock.Anything, person.Name).Return(expectedGender, expectedGenderProb, nil)
		mockNationalityService.On("GetNationalityByName", mock.Anything, person.Name).Return(expectedNationality, expectedNationalityProb, nil)

		app.Put("/persons/:id/enrich", handler.EnrichPerson)

		req := httptest.NewRequest(http.MethodPut, "/persons/"+personID.String()+"/enrich", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var respPerson entities.Person
		err = json.NewDecoder(resp.Body).Decode(&respPerson)
		require.NoError(t, err)

		assert.Equal(t, personID, respPerson.ID)
		assert.Equal(t, expectedAge, *respPerson.Age)
		assert.Equal(t, expectedGender, *respPerson.Gender)
		assert.Equal(t, expectedGenderProb, *respPerson.GenderProbability)
		assert.Equal(t, expectedNationality, *respPerson.Nationality)
		assert.Equal(t, expectedNationalityProb, *respPerson.NationalityProbability)

		mockRepo.AssertExpectations(t)
		mockAgeService.AssertExpectations(t)
		mockGenderService.AssertExpectations(t)
		mockNationalityService.AssertExpectations(t)
	})

	t.Run("should skip enrichment for fields that already have values", func(t *testing.T) {
		app, mockRepo, mockAgeService, mockGenderService, mockNationalityService, handler := setupTest()

		person := createPersonWithoutEnrichment()
		personID := person.ID

		existingAge := 25
		person.Age = &existingAge

		existingNationality := "US"
		existingNationalityProb := 0.88
		person.Nationality = &existingNationality
		person.NationalityProbability = &existingNationalityProb

		expectedGender := "male"
		expectedGenderProb := 0.95

		mockRepo.On("GetByID", mock.Anything, personID).Return(person, nil)
		mockRepo.On("UpdatePerson", mock.Anything, mock.MatchedBy(func(p *entities.Person) bool {
			return p.ID == personID &&
				*p.Age == existingAge &&
				*p.Gender == expectedGender &&
				*p.GenderProbability == expectedGenderProb &&
				*p.Nationality == existingNationality
		})).Return(nil)

		mockGenderService.On("GetGenderByName", mock.Anything, person.Name).Return(expectedGender, expectedGenderProb, nil)

		app.Put("/persons/:id/enrich", handler.EnrichPerson)

		req := httptest.NewRequest(http.MethodPut, "/persons/"+personID.String()+"/enrich", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var respPerson entities.Person
		err = json.NewDecoder(resp.Body).Decode(&respPerson)
		require.NoError(t, err)

		assert.Equal(t, personID, respPerson.ID)
		assert.Equal(t, existingAge, *respPerson.Age)
		assert.Equal(t, expectedGender, *respPerson.Gender)
		assert.Equal(t, expectedGenderProb, *respPerson.GenderProbability)
		assert.Equal(t, existingNationality, *respPerson.Nationality)
		assert.Equal(t, existingNationalityProb, *respPerson.NationalityProbability)

		mockAgeService.AssertNotCalled(t, "GetAgeByName", mock.Anything, mock.Anything)
		mockNationalityService.AssertNotCalled(t, "GetNationalityByName", mock.Anything, mock.Anything)

		mockRepo.AssertExpectations(t)
		mockGenderService.AssertExpectations(t)
	})

	t.Run("should handle service errors gracefully", func(t *testing.T) {
		app, mockRepo, mockAgeService, mockGenderService, mockNationalityService, handler := setupTest()
		person := createPersonWithoutEnrichment()
		personID := person.ID

		expectedNationality := "RU"
		expectedNationalityProb := 0.90

		mockRepo.On("GetByID", mock.Anything, personID).Return(person, nil)
		mockRepo.On("UpdatePerson", mock.Anything, mock.MatchedBy(func(p *entities.Person) bool {
			return p.ID == personID &&
				p.Age == nil &&
				p.Gender == nil &&
				*p.Nationality == expectedNationality
		})).Return(nil)

		mockAgeService.On("GetAgeByName", mock.Anything, person.Name).Return(0, 0.0, errors.New("age service error"))
		mockGenderService.On("GetGenderByName", mock.Anything, person.Name).Return("", 0.0, errors.New("gender service error"))
		mockNationalityService.On("GetNationalityByName", mock.Anything, person.Name).Return(expectedNationality, expectedNationalityProb, nil)

		app.Put("/persons/:id/enrich", handler.EnrichPerson)

		req := httptest.NewRequest(http.MethodPut, "/persons/"+personID.String()+"/enrich", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var respPerson entities.Person
		err = json.NewDecoder(resp.Body).Decode(&respPerson)
		require.NoError(t, err)

		assert.Equal(t, personID, respPerson.ID)
		assert.Nil(t, respPerson.Age)
		assert.Nil(t, respPerson.Gender)
		assert.Equal(t, expectedNationality, *respPerson.Nationality)
		assert.Equal(t, expectedNationalityProb, *respPerson.NationalityProbability)

		mockRepo.AssertExpectations(t)
		mockAgeService.AssertExpectations(t)
		mockGenderService.AssertExpectations(t)
		mockNationalityService.AssertExpectations(t)
	})

	t.Run("should return 500 when UpdatePerson fails", func(t *testing.T) {
		app, mockRepo, mockAgeService, mockGenderService, mockNationalityService, handler := setupTest()
		person := createPersonWithoutEnrichment()
		personID := person.ID

		mockRepo.On("GetByID", mock.Anything, personID).Return(person, nil)
		mockRepo.On("UpdatePerson", mock.Anything, mock.Anything).Return(errors.New("database error"))

		expectedAge := 30
		expectedGender := "male"
		expectedNationality := "RU"
		mockAgeService.On("GetAgeByName", mock.Anything, person.Name).Return(expectedAge, 0.85, nil)
		mockGenderService.On("GetGenderByName", mock.Anything, person.Name).Return(expectedGender, 0.95, nil)
		mockNationalityService.On("GetNationalityByName", mock.Anything, person.Name).Return(expectedNationality, 0.90, nil)

		app.Put("/persons/:id/enrich", func(c fiber.Ctx) error {
			err := handler.EnrichPerson(c)
			if err != nil && strings.Contains(err.Error(), "failed to save enriched data") {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to save enriched data",
				})
			}
			return err
		})

		req := httptest.NewRequest(http.MethodPut, "/persons/"+personID.String()+"/enrich", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		var errorResp map[string]string
		err = json.NewDecoder(resp.Body).Decode(&errorResp)
		require.NoError(t, err)
		assert.Equal(t, "Failed to save enriched data", errorResp["error"])

		mockRepo.AssertExpectations(t)
		mockAgeService.AssertExpectations(t)
		mockGenderService.AssertExpectations(t)
		mockNationalityService.AssertExpectations(t)
	})

	t.Run("should handle JSON encoding error", func(t *testing.T) {
		app := fiber.New(fiber.Config{
			JSONEncoder: func(v interface{}) ([]byte, error) {
				return nil, errors.New("forced JSON marshal error")
			},
		})

		mockPersonRepository := &MockPersonRepository{}
		mockAgeService := &MockAgeService{}
		mockGenderService := &MockGenderService{}
		mockNationalityService := &MockNationalityService{}

		person := createPersonWithoutEnrichment()
		personID := person.ID

		mockPersonRepository.On("GetByID", mock.Anything, personID).Return(person, nil)
		mockPersonRepository.On("UpdatePerson", mock.Anything, mock.Anything).Return(nil)

		mockAgeService.On("GetAgeByName", mock.Anything, person.Name).Return(30, 0.85, nil)
		mockGenderService.On("GetGenderByName", mock.Anything, person.Name).Return("male", 0.95, nil)
		mockNationalityService.On("GetNationalityByName", mock.Anything, person.Name).Return("RU", 0.90, nil)

		mockPeopleRepositories := &MockPeopleRepositories{
			mockPersonRepository: mockPersonRepository,
		}
		mockRepositories := &MockRepositories{
			mockPeopleRepositories: mockPeopleRepositories,
		}

		mockPeopleServices := &MockPeopleServices{
			mockAgeService:         mockAgeService,
			mockGenderService:      mockGenderService,
			mockNationalityService: mockNationalityService,
		}
		mockAPI := &MockAPI{
			mockPeopleServices: mockPeopleServices,
		}

		handler := handlers.NewPersonHandler(mockAPI, mockRepositories)

		var capturedErr error
		app.Put("/persons/:id/enrich", func(c fiber.Ctx) error {
			err := handler.EnrichPerson(c)
			capturedErr = err
			return err
		})

		req := httptest.NewRequest(http.MethodPut, "/persons/"+personID.String()+"/enrich", nil)
		resp, _ := app.Test(req)

		assert.NotNil(t, capturedErr)
		assert.Contains(t, capturedErr.Error(), "failed to send JSON response")
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		mockPersonRepository.AssertExpectations(t)
		mockAgeService.AssertExpectations(t)
		mockGenderService.AssertExpectations(t)
		mockNationalityService.AssertExpectations(t)
	})
}
