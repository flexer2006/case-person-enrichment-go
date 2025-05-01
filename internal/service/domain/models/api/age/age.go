// Package age содержит модели данных для работы с API определения возраста.
package age

// Response представляет ответ от API agify.io.
type Response struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Count int    `json:"count"`
}
