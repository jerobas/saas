package model

import "time"

type ItemConversion struct {
	ID               int64         `json:"id"`
	
	FromItemID       int64         `json:"from_item_id"`
	ToItemID         int64         `json:"to_item_id"`
	Factor           float64       `json:"factor"` 

	CreatedAt        time.Time     `json:"created_at"`
}

type ItemConversionInsertDTO struct {
	FromItemID       int64         `json:"from_item_id"`
	ToItemID         int64         `json:"to_item_id"`
	Factor           float64       `json:"factor"`
}