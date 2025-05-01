// Package handlers содержит обработчики HTTP-запросов.
package handlers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/flexer2006/case-person-enrichment-go/internal/service/domain/entities"
	"github.com/flexer2006/case-person-enrichment-go/internal/service/ports/api"
	"github.com/flexer2006/case-person-enrichment-go/internal/service/ports/repo"
	"github.com/flexer2006/case-person-enrichment-go/pkg/logger"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Ошибки, которые могут возникнуть при работе с персонами.
var (
	ErrNameSurnameRequired = errors.New("name and surname are required")
	ErrPersonNotFound      = errors.New("person not found")
)

// PersonHandler обрабатывает HTTP-запросы для работы с персонами.
type PersonHandler struct {
	api          api.API
	repositories repo.Repositories
}

// NewPersonHandler создает новый обработчик для работы с персонами.
func NewPersonHandler(api api.API, repositories repo.Repositories) *PersonHandler {
	return &PersonHandler{
		api:          api,
		repositories: repositories,
	}
}

// GetPersons обрабатывает запрос на получение списка персон с фильтрами и пагинацией.
func (h *PersonHandler) GetPersons(ctx fiber.Ctx) error {
	requestCtx := ctx.Context()
	logger.Debug(requestCtx, "handling get persons request")

	// Извлечение параметров пагинации
	limitStr := ctx.Query("limit", "10")
	offsetStr := ctx.Query("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	filter := make(map[string]any)
	for _, field := range []string{"name", "surname", "patronymic", "gender", "nationality"} {
		if value := ctx.Query(field); value != "" {
			filter[field] = value
		}
	}

	if ageStr := ctx.Query("age"); ageStr != "" {
		if age, err := strconv.Atoi(ageStr); err == nil {
			filter["age"] = age
		}
	}

	persons, total, err := h.repositories.People().Person().GetPersons(requestCtx, filter, offset, limit)
	if err != nil {
		logger.Error(requestCtx, "failed to get persons", zap.Error(err))
		if err := ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve persons",
		}); err != nil {
			return fmt.Errorf("failed to send JSON response: %w", err)
		}
		return fmt.Errorf("failed to get persons: %w", err)
	}

	if err := ctx.JSON(fiber.Map{
		"data":   persons,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}); err != nil {
		return fmt.Errorf("failed to send JSON response: %w", err)
	}
	return nil
}

// GetPersonByID обрабатывает запрос на получение персоны по идентификатору.
func (h *PersonHandler) GetPersonByID(ctx fiber.Ctx) error {
	requestCtx := ctx.Context()
	idParam := ctx.Params("id")

	logger.Debug(requestCtx, "handling get person by ID request", zap.String("id", idParam))

	personID, err := uuid.Parse(idParam)
	if err != nil {
		if err := ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid UUID format",
		}); err != nil {
			return fmt.Errorf("failed to send JSON response: %w", err)
		}
		return fmt.Errorf("invalid UUID format: %w", err)
	}

	person, err := h.repositories.People().Person().GetByID(requestCtx, personID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			if err := ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Person not found",
			}); err != nil {
				return fmt.Errorf("failed to send JSON response: %w", err)
			}
			return fmt.Errorf("person not found: %w", err)
		}
		if err := ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get person",
		}); err != nil {
			return fmt.Errorf("failed to send JSON response: %w", err)
		}
		return fmt.Errorf("failed to get person: %w", err)
	}

	if err := ctx.JSON(person); err != nil {
		return fmt.Errorf("failed to send JSON response: %w", err)
	}
	return nil
}

// CreatePerson обрабатывает запрос на создание новой персоны.
func (h *PersonHandler) CreatePerson(ctx fiber.Ctx) error {
	requestCtx := ctx.Context()
	logger.Debug(requestCtx, "handling create person request")

	var person entities.Person
	if err := ctx.Bind().Body(&person); err != nil {
		if err := ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		}); err != nil {
			return fmt.Errorf("failed to send JSON response: %w", err)
		}
		return fmt.Errorf("invalid request body: %w", err)
	}

	if person.Name == "" || person.Surname == "" {
		if err := ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name and surname are required",
		}); err != nil {
			return fmt.Errorf("failed to send JSON response: %w", err)
		}
		return fmt.Errorf("%w", ErrNameSurnameRequired)
	}

	if person.ID == uuid.Nil {
		person.ID = uuid.New()
	}

	if err := h.repositories.People().Person().CreatePerson(requestCtx, &person); err != nil {
		logger.Error(requestCtx, "failed to create person", zap.Error(err))
		if err := ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create person",
		}); err != nil {
			return fmt.Errorf("failed to send JSON response: %w", err)
		}
		return fmt.Errorf("failed to create person: %w", err)
	}

	if err := ctx.Status(fiber.StatusCreated).JSON(person); err != nil {
		return fmt.Errorf("failed to send JSON response: %w", err)
	}
	return nil
}

// UpdatePerson обрабатывает запрос на обновление персоны.
func (h *PersonHandler) UpdatePerson(ctx fiber.Ctx) error {
	requestCtx := ctx.Context()
	idParam := ctx.Params("id")

	logger.Debug(requestCtx, "handling update person request", zap.String("id", idParam))

	personID, err := uuid.Parse(idParam)
	if err != nil {
		if err := ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid UUID format",
		}); err != nil {
			return fmt.Errorf("failed to send JSON response: %w", err)
		}
		return fmt.Errorf("invalid UUID format: %w", err)
	}

	exists, err := h.repositories.People().Person().ExistsByID(requestCtx, personID)
	if err != nil {
		logger.Error(requestCtx, "failed to check if person exists", zap.Error(err))
		if err := ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check if person exists",
		}); err != nil {
			return fmt.Errorf("failed to send JSON response: %w", err)
		}
		return fmt.Errorf("failed to check if person exists: %w", err)
	}

	if !exists {
		if err := ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Person not found",
		}); err != nil {
			return fmt.Errorf("failed to send JSON response: %w", err)
		}
		return fmt.Errorf("%w", ErrPersonNotFound)
	}

	var person entities.Person
	if err := ctx.Bind().Body(&person); err != nil {
		if err := ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		}); err != nil {
			return fmt.Errorf("failed to send JSON response: %w", err)
		}
		return fmt.Errorf("invalid request body: %w", err)
	}

	person.ID = personID

	if person.Name == "" || person.Surname == "" {
		if err := ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name and surname are required",
		}); err != nil {
			return fmt.Errorf("failed to send JSON response: %w", err)
		}
		return fmt.Errorf("%w", ErrNameSurnameRequired)
	}

	if err := h.repositories.People().Person().UpdatePerson(requestCtx, &person); err != nil {
		logger.Error(requestCtx, "failed to update person", zap.Error(err))
		if err := ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update person",
		}); err != nil {
			return fmt.Errorf("failed to send JSON response: %w", err)
		}
		return fmt.Errorf("failed to update person: %w", err)
	}

	if err := ctx.JSON(person); err != nil {
		return fmt.Errorf("failed to send JSON response: %w", err)
	}
	return nil
}

// DeletePerson обрабатывает запрос на удаление персоны.
func (h *PersonHandler) DeletePerson(ctx fiber.Ctx) error {
	requestCtx := ctx.Context()
	idParam := ctx.Params("id")

	logger.Debug(requestCtx, "handling delete person request", zap.String("id", idParam))

	personID, err := uuid.Parse(idParam)
	if err != nil {
		if err := ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid UUID format",
		}); err != nil {
			return fmt.Errorf("failed to send JSON response: %w", err)
		}
		return fmt.Errorf("invalid UUID format: %w", err)
	}

	if err := h.repositories.People().Person().DeletePerson(requestCtx, personID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			if err := ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Person not found",
			}); err != nil {
				return fmt.Errorf("failed to send JSON response: %w", err)
			}
			return fmt.Errorf("person not found: %w", err)
		}
		logger.Error(requestCtx, "failed to delete person", zap.Error(err))
		if err := ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete person",
		}); err != nil {
			return fmt.Errorf("failed to send JSON response: %w", err)
		}
		return fmt.Errorf("failed to delete person: %w", err)
	}

	if err := ctx.SendStatus(fiber.StatusNoContent); err != nil {
		return fmt.Errorf("failed to send status: %w", err)
	}
	return nil
}

// EnrichPerson обрабатывает запрос на обогащение данных персоны.
func (h *PersonHandler) EnrichPerson(ctx fiber.Ctx) error {
	requestCtx := ctx.Context()
	idParam := ctx.Params("id")

	logger.Debug(requestCtx, "handling enrich person request", zap.String("id", idParam))

	personID, err := uuid.Parse(idParam)
	if err != nil {
		if err := ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid UUID format",
		}); err != nil {
			return fmt.Errorf("failed to send JSON response: %w", err)
		}
		return fmt.Errorf("invalid UUID format: %w", err)
	}

	person, err := h.repositories.People().Person().GetByID(requestCtx, personID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			if err := ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Person not found",
			}); err != nil {
				return fmt.Errorf("failed to send JSON response: %w", err)
			}
			return fmt.Errorf("person not found: %w", err)
		}
		logger.Error(requestCtx, "failed to get person", zap.Error(err))
		if err := ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get person",
		}); err != nil {
			return fmt.Errorf("failed to send JSON response: %w", err)
		}
		return fmt.Errorf("failed to get person: %w", err)
	}

	if person.Age == nil {
		ageService := h.api.People().Age()
		age, probability, err := ageService.GetAgeByName(requestCtx, person.Name)
		if err == nil {
			person.Age = &age
			logger.Debug(requestCtx, "enriched with age data",
				zap.Int("age", age),
				zap.Float64("probability", probability))
		} else {
			logger.Warn(requestCtx, "failed to get age data", zap.Error(err))
		}
	}

	if person.Gender == nil {
		genderService := h.api.People().Gender()
		gender, probability, err := genderService.GetGenderByName(requestCtx, person.Name)
		if err == nil {
			person.Gender = &gender
			person.GenderProbability = &probability
			logger.Debug(requestCtx, "enriched with gender data",
				zap.String("gender", gender),
				zap.Float64("probability", probability))
		} else {
			logger.Warn(requestCtx, "failed to get gender data", zap.Error(err))
		}
	}

	if person.Nationality == nil {
		nationalityService := h.api.People().Nationality()
		nationality, probability, err := nationalityService.GetNationalityByName(requestCtx, person.Name)
		if err == nil {
			person.Nationality = &nationality
			person.NationalityProbability = &probability
			logger.Debug(requestCtx, "enriched with nationality data",
				zap.String("nationality", nationality),
				zap.Float64("probability", probability))
		} else {
			logger.Warn(requestCtx, "failed to get nationality data", zap.Error(err))
		}
	}

	err = h.repositories.People().Person().UpdatePerson(requestCtx, person)
	if err != nil {
		logger.Error(requestCtx, "failed to save enriched data", zap.Error(err))
		if err := ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save enriched data",
		}); err != nil {
			return fmt.Errorf("failed to send JSON response: %w", err)
		}
		return fmt.Errorf("failed to save enriched data: %w", err)
	}

	if err := ctx.JSON(person); err != nil {
		return fmt.Errorf("failed to send JSON response: %w", err)
	}
	return nil
}
