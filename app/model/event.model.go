package model

import (
	"database/sql"
	"time"
)

type Event struct {
	ID                    int64           `json:"id"`
	EventType             string          `json:"event_type"`
	Status                string          `json:"status"`
	CounterpartyEntityID  sql.NullInt64   `json:"counterparty_entity_id"`
	Notes                 sql.NullString  `json:"notes"`
	OccurredAt            time.Time       `json:"occurred_at"`
	CreatedAt             time.Time       `json:"created_at"`
}

type EventInsertDTO struct {
	EventType             string
	Status                string
	CounterpartyEntityID  *int64
	Notes                 *string
	OccurredAt            time.Time
}

type CreateEventInput struct {
	CounterpartyEntityID  *int64          `json:"counterparty_entity_id"`
	Notes                 *string         `json:"notes"`
	OccurredAt            time.Time       `json:"occurred_at"`
}

type GetPurchaseByIDOutput struct {
	ID                    int64                  `json:"id"`
	EventType             string                 `json:"event_type"`
	Status                string                 `json:"status"`
	CounterpartyEntityID  sql.NullInt64          `json:"counterparty_entity_id"`
	Notes                 sql.NullString         `json:"notes"`
	OccurredAt            time.Time              `json:"occurred_at"`
	CreatedAt             time.Time              `json:"created_at"`
	PurchaseLines         *[]model.PurchaseLine  `json:"purchase_lines"`
}