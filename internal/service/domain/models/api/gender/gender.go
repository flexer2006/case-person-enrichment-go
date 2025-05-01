// Package gender содержит модели данных для работы с API определения пола.
package gender

// Response представляет ответ от API genderize.io.
type Response struct {
	Name        string  `json:"name"`
	Gender      string  `json:"gender"`
	Probability float64 `json:"probability"`
	Count       int     `json:"count"`
}
