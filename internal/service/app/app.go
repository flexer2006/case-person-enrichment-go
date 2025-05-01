// Package app содержит центральную точку связывания компонентов приложения.
package app

import (
	"context"
	"fmt"
	"time"

	"github.com/flexer2006/case-person-enrichment-go/internal/service/adapters/enrichment"
	"github.com/flexer2006/case-person-enrichment-go/internal/service/adapters/postgres"
	"github.com/flexer2006/case-person-enrichment-go/internal/service/adapters/server"
	"github.com/flexer2006/case-person-enrichment-go/internal/service/domain/entities"
	"github.com/flexer2006/case-person-enrichment-go/internal/service/domain/services/person"
	"github.com/flexer2006/case-person-enrichment-go/internal/service/ports/api"
	"github.com/flexer2006/case-person-enrichment-go/internal/service/ports/repo"
	personrepo "github.com/flexer2006/case-person-enrichment-go/internal/service/ports/repo/people/person"
	"github.com/flexer2006/case-person-enrichment-go/internal/service/setup"
	"github.com/flexer2006/case-person-enrichment-go/pkg/database/migrate"
	pgadapter "github.com/flexer2006/case-person-enrichment-go/pkg/database/postgres"
	"github.com/flexer2006/case-person-enrichment-go/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Application представляет основное приложение, объединяющее все компоненты.
type Application struct {
	config        *setup.Config
	db            *pgadapter.Database
	pgAdapter     *postgres.Adapter
	apiAdapter    api.API
	repositories  repo.Repositories
	httpServer    *server.Server
	personService person.Service
}

// NewApplication создает новый экземпляр приложения с указанной конфигурацией.
func NewApplication(ctx context.Context, config *setup.Config) (*Application, error) {
	logger.Info(ctx, "initializing application")

	dbConfig := pgadapter.Config{
		Host:     config.Postgres.Host,
		Port:     config.Postgres.Port,
		User:     config.Postgres.User,
		Password: config.Postgres.Password,
		Database: config.Postgres.Database,
		SSLMode:  config.Postgres.SSLMode,
		MinConns: config.Postgres.PoolMinConns,
		MaxConns: config.Postgres.PoolMaxConns,
	}

	// Создание базы данных
	database, err := pgadapter.New(ctx, dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	if config.Migrations.Path != "" {
		migrateConfig := migrate.Config{
			Path: config.Migrations.Path,
		}
		migrator := migrate.NewAdapter(migrateConfig)
		if err := migrator.Up(ctx, database.GetDSN()); err != nil {
			return nil, fmt.Errorf("failed to apply migrations: %w", err)
		}
	}

	pgAdapter := postgres.NewPostgresAdapter(database)

	apiAdapter := enrichment.NewDefaultEnrichment()

	personSvc := NewPersonService(pgAdapter.Repositories(), apiAdapter)

	httpServer := server.New(config.Server, apiAdapter, pgAdapter.Repositories())

	app := &Application{
		config:        config,
		db:            database,
		pgAdapter:     pgAdapter,
		apiAdapter:    apiAdapter,
		repositories:  pgAdapter.Repositories(),
		httpServer:    httpServer,
		personService: personSvc,
	}

	logger.Info(ctx, "application initialized successfully")
	return app, nil
}

// Start запускает все сервисы приложения.
func (a *Application) Start(ctx context.Context) error {
	logger.Info(ctx, "starting application")

	if err := a.httpServer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}
	return nil
}

// Stop останавливает все сервисы приложения.
func (a *Application) Stop(ctx context.Context) error {
	logger.Info(ctx, "stopping application")

	shutdownTimeout, err := time.ParseDuration(a.config.Graceful.ShutdownTimeout)
	if err != nil {
		shutdownTimeout = 5 * time.Second
		logger.Warn(ctx, "invalid graceful shutdown timeout, using default",
			zap.String("default", shutdownTimeout.String()))
	}

	ctx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	if err := a.httpServer.Stop(ctx); err != nil {
		logger.Error(ctx, "error stopping HTTP server", zap.Error(err))
	}

	a.pgAdapter.Close(ctx)

	logger.Info(ctx, "application stopped")
	return nil
}

// Repositories возвращает репозитории приложения.
func (a *Application) Repositories() repo.Repositories {
	return a.repositories
}

// API возвращает API-адаптер приложения.
func (a *Application) API() api.API {
	return a.apiAdapter
}

// PersonService возвращает сервис для работы с персонами.
func (a *Application) PersonService() person.Service {
	return a.personService
}

// NewPersonService создает новый сервис для работы с персонами.
func NewPersonService(repositories repo.Repositories, apiAdapter api.API) person.Service {
	return &personServiceImpl{
		repository: repositories.People().Person(),
		apiAdapter: apiAdapter,
	}
}

// personServiceImpl реализует интерфейс PersonService.
type personServiceImpl struct {
	repository personrepo.Repository
	apiAdapter api.API
}

// Реализация методов PersonService...
// GetByID возвращает персону по идентификатору.
func (s *personServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Person, error) {
	person, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get person by ID: %w", err)
	}
	return person, nil
}

// GetPersons возвращает список персон с фильтрацией и пагинацией.
func (s *personServiceImpl) GetPersons(ctx context.Context, filter map[string]any, offset, limit int) ([]*entities.Person, int, error) {
	persons, count, err := s.repository.GetPersons(ctx, filter, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get persons: %w", err)
	}
	return persons, count, nil
}

// CreatePerson создает новую персону.
func (s *personServiceImpl) CreatePerson(ctx context.Context, person *entities.Person) error {
	if err := s.repository.CreatePerson(ctx, person); err != nil {
		return fmt.Errorf("failed to create person: %w", err)
	}
	return nil
}

// UpdatePerson обновляет существующую персону.
func (s *personServiceImpl) UpdatePerson(ctx context.Context, person *entities.Person) error {
	if err := s.repository.UpdatePerson(ctx, person); err != nil {
		return fmt.Errorf("failed to update person: %w", err)
	}
	return nil
}

// DeletePerson удаляет персону по идентификатору.
func (s *personServiceImpl) DeletePerson(ctx context.Context, id uuid.UUID) error {
	if err := s.repository.DeletePerson(ctx, id); err != nil {
		return fmt.Errorf("failed to delete person: %w", err)
	}
	return nil
}

// EnrichPerson обогащает данные персоны (возраст, пол, национальность).
func (s *personServiceImpl) EnrichPerson(ctx context.Context, id uuid.UUID) (*entities.Person, error) {
	logger.Debug(ctx, "enriching person data", zap.String("id", id.String()))

	person, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get person: %w", err)
	}

	if person.Age == nil {
		ageService := s.apiAdapter.People().Age()
		age, _, err := ageService.GetAgeByName(ctx, person.Name)
		if err == nil {
			person.Age = &age
		} else {
			logger.Warn(ctx, "failed to enrich with age data", zap.Error(err))
		}
	}

	// Обогащение данными о поле
	if person.Gender == nil {
		genderService := s.apiAdapter.People().Gender()
		gender, probability, err := genderService.GetGenderByName(ctx, person.Name)
		if err == nil {
			person.Gender = &gender
			person.GenderProbability = &probability
		} else {
			logger.Warn(ctx, "failed to enrich with gender data", zap.Error(err))
		}
	}

	if person.Nationality == nil {
		nationalityService := s.apiAdapter.People().Nationality()
		nationality, probability, err := nationalityService.GetNationalityByName(ctx, person.Name)
		if err == nil {
			person.Nationality = &nationality
			person.NationalityProbability = &probability
		} else {
			logger.Warn(ctx, "failed to enrich with nationality data", zap.Error(err))
		}
	}

	err = s.repository.UpdatePerson(ctx, person)
	if err != nil {
		return nil, fmt.Errorf("failed to save enriched person data: %w", err)
	}

	return person, nil
}
