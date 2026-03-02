package model

import (
	"database/sql"
	"time"
)

type Entity struct {
	ID         int64           `json:"id"`
	Name       string          `json:"name"`
	Phone      sql.NullString  `json:"phone"`
	Email      sql.NullString  `json:"email"`
	CreatedAt  time.Time       `json:"created_at"`
}

type EntityInsertDTO struct {
	Name       string          `json:"name"`
	Phone      sql.NullString  `json:"phone"`
	Email      sql.NullString  `json:"email"`
}