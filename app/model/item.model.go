package model

import (
	"database/sql"
	"time"
)

type Item struct {
	ID               int64         `json:"id"`
	Name             string        `json:"name"`
	Unit             string        `json:"unit"`
	
	// If these are booleans in spirit, consider bool in Go and INTEGER(0/1) in SQLite,
	// but keeping int64 here to match your current schema style.
	Sellable         int64         `json:"sellable"`
	Purchasable      int64         `json:"purchasable"`
	Producible       int64         `json:"producible"`
	
	DefaultSalePrice sql.NullInt64 `json:"default_sale_price"`

	CreatedAt        time.Time     `json:"created_at"`
}

type ItemInsertDTO struct {
	Name             string        `json:"name"`
	Unit             string        `json:"unit"`
	
	Sellable         int64         `json:"sellable"`
	Purchasable      int64         `json:"purchasable"`
	Producible       int64         `json:"producible"`
	
	DefaultSalePrice sql.NullInt64 `json:"default_sale_price"`
}