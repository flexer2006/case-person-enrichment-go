// Package routes определяет маршруты HTTP-сервера.
package routes

import (
	"github.com/flexer2006/case-person-enrichment-go/internal/service/adapters/server/handlers"
	"github.com/flexer2006/case-person-enrichment-go/internal/service/ports/api"
	"github.com/flexer2006/case-person-enrichment-go/internal/service/ports/repo"
	"github.com/gofiber/fiber/v3"
)

// Setup настраивает маршруты HTTP-сервера.
func Setup(app *fiber.App, api api.API, repositories repo.Repositories) {
	personHandler := handlers.NewPersonHandler(api, repositories)

	// Группа для API версии 1.
	v1 := app.Group("/api/v1")

	// Маршруты для работы с персонами.
	persons := v1.Group("/persons")
	persons.Get("/", personHandler.GetPersons)         // Получение списка с фильтрами и пагинацией.
	persons.Get("/:id", personHandler.GetPersonByID)   // Получение по ID.
	persons.Post("/", personHandler.CreatePerson)      // Создание новой персоны.
	persons.Put("/:id", personHandler.UpdatePerson)    // Обновление персоны.
	persons.Patch("/:id", personHandler.UpdatePerson)  // Частичное обновление персоны.
	persons.Delete("/:id", personHandler.DeletePerson) // Удаление персоны.

	// Маршрут для обогащения данных персоны.
	persons.Post("/:id/enrich", personHandler.EnrichPerson)
}
