// Package nationality содержит модели данных для работы с API определения национальности.
package nationality

// Country представляет информацию о стране в ответе API.
type Country struct {
	CountryID   string  `json:"country_id"`
	Probability float64 `json:"probability"`
}

// Response представляет ответ от API nationalize.io.
type Response struct {
	Name      string    `json:"name"`
	Countries []Country `json:"country"`
}
