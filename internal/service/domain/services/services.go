// Package services объединяет все доменные сервисы приложения.
package services

import (
	"github.com/flexer2006/case-person-enrichment-go/internal/service/domain/services/person"
)

// PersonService представляет сервис для работы с персонами.
type PersonService = person.Service
