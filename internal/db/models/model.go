package models

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// CompanyModel represents the company model
type CompanyModel struct {
	ID              uuid.UUID `json:"id" db:"id"`
	Name            string    `json:"name" db:"name"`
	Description     string    `json:"description" db:"description"`
	AmountEmployees int       `json:"amount_employees" db:"amount_employees"`
	Registered      bool      `json:"registered" db:"registered"`
	Type            string    `json:"type" db:"type"`
}

// UserModel represents the user model
type UserModel struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Email       string    `json:"email" db:"email"`
	EncPassword string    `json:"enc_password" db:"enc_password"`
}

// EventModel represents the event model
type EventModel struct {
	Type      string             `json:"type" db:"type"`
	Timestamp pgtype.Timestamptz `json:"timestamp" db:"timestamp"`
	ID        uuid.UUID          `json:"id" db:"id"`
	EntityID  uuid.UUID          `json:"entity_id" db:"entity_id"`
}
