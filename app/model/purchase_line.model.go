package model

import "time"

type PurchaseLine struct {
	ID           int       `json:"id"`
	EventID      int       `json:"event_id"`
	ItemID       int       `json:"item_id"`
	Quantity     float64   `json:"quantity"`
	UnitCost     int       `json:"unit_cost"`
	CreatedAt    time.Time `json:"created_at"`
}

type PurchaseLineInsertDTO struct {
	EventID      int       `json:"event_id"`
	ItemID       int       `json:"item_id"`
	Quantity     float64   `json:"quantity"`
	UnitCost     int       `json:"unit_price"`
}

// type PurchaseLineUpdateDTO struct {
// 	EventID      *int      `json:"event_id"`
// 	ItemID       *int      `json:"item_id"`
// 	Quantity     *float64  `json:"quantity"`
// 	UnitCost     *int      `json:"unit_cost"`
// }