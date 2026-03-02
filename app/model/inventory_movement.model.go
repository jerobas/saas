package model

import (
	"database/sql"
	"time"
)

type InventoryMovement struct {
	ID               int64
	EventID          int64
	ItemID           int64
	Quantity         float64 // > 0

	// Nullable (but required when direction='IN' by DB constraint)
	UnitCost         sql.NullInt64
	CreatedAt        time.Time

	//

	Direction        string  // 'IN' | 'OUT'

	// Nullable timestamps / foreign key
	ExpiresAt        sql.NullTime
	OriginMovementID sql.NullInt64

	OccurredAt       time.Time
}

type InventoryMovementInsertDTO struct {
	EventID          int64
	ItemID           int64
	Quantity         float64 // > 0

	// Nullable (but required when direction='IN' by DB constraint)
	UnitCost         sql.NullInt64

	//

	Direction        string  // 'IN' | 'OUT'

	// Nullable timestamps / foreign key
	ExpiresAt        sql.NullTime
	OriginMovementID sql.NullInt64

	OccurredAt       time.Time
}