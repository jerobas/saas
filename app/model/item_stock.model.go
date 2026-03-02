package model

import (
	"database/sql"
	"time"
)

type ItemStock struct {
	ItemID           int64          `json:"item_id"`
	Quantity         float64        `json:"quantity"`
	AverageUnitCost  sql.NullInt64  `json:"average_unit_cost"`
	UpdatedAt        time.Time      `json:"updated_at"`
}