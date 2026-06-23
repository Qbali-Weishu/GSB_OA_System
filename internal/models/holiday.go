package models

import "time"

// Holiday 法定节假日
type Holiday struct {
	ID          int64     `json:"id" db:"id"`
	Date        time.Time `json:"date" db:"date"`
	Name        string    `json:"name" db:"name"`
	// IsWorkday 调班补班日，节假日调整后某些周末需要上班
	IsWorkday   bool      `json:"is_workday" db:"is_workday"`
	Description string    `json:"description" db:"description"`
	Year        int       `json:"year" db:"year"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}
