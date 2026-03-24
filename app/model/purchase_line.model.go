package model

import "time"

type PurchaseLine struct {
	ID           int64     `json:"id"`
	EventID      int64     `json:"event_id"`
	ItemID       int64     `json:"item_id"`
	Quantity     float64   `json:"quantity"`
	UnitCost     int64     `json:"unit_cost"`
	CreatedAt    time.Time `json:"created_at"`
}

type PurchaseLineInsertDTO struct {
	EventID      int64     `json:"event_id"`
	ItemID       int64     `json:"item_id"`
	Quantity     float64   `json:"quantity"`
	UnitCost     int64     `json:"unit_cost"`
}

type CreatePurchaseWithLinesInput struct {
	ItemID       int64     `json:"item_id"`
	Quantity     float64   `json:"quantity"`
	UnitCost     float64   `json:"unit_cost"`
}

// type PurchaseLineUpdateDTO struct {
// 	EventID      *int      `json:"event_id"`
// 	ItemID       *int      `json:"item_id"`
// 	Quantity     *float64  `json:"quantity"`
// 	UnitCost     *int      `json:"unit_cost"`
// }