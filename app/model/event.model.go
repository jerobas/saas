package model

import (
	"database/sql"
	"time"
)

type Event struct {
	ID                    int64           `json:"id"`
	EventType             string          `json:"event_type"`
	Status                string          `json:"status"`
	CounterpartyEntityId  sql.NullInt64   `json:"counterparty_entity_id"`
	Notes                 sql.NullString  `json:"notes"`
	OccurredAt             time.Time      `json:"occurred_at"`
	CreatedAt             time.Time       `json:"created_at"`
}

type EventInsertDTO struct {
	EventType             string          `json:"event_type"`
	Status                string          `json:"status"`
	CounterpartyEntityId  sql.NullInt64   `json:"counterparty_entity_id"`
	Notes                 sql.NullString  `json:"notes"`
	OccurredAt             time.Time      `json:"occurred_at"`
}