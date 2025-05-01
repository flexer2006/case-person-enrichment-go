// Package person содержит определение персоны.
package person

import (
	"time"

	"github.com/google/uuid"
)

// Person представляет сущность человека в системе.
type Person struct {
	ID                     uuid.UUID `db:"id" json:"id"`
	Name                   string    `db:"name" json:"name"`
	Surname                string    `db:"surname" json:"surname"`
	Patronymic             *string   `db:"patronymic" json:"patronymic,omitempty"`
	Age                    *int      `db:"age" json:"age,omitempty"`
	Gender                 *string   `db:"gender" json:"gender,omitempty"`
	GenderProbability      *float64  `db:"gender_probability" json:"gender_probability,omitempty"`
	Nationality            *string   `db:"nationality" json:"nationality,omitempty"`
	NationalityProbability *float64  `db:"nationality_probability" json:"nationality_probability,omitempty"`
	CreatedAt              time.Time `db:"created_at" json:"created_at"`
	UpdatedAt              time.Time `db:"updated_at" json:"updated_at"`
}
