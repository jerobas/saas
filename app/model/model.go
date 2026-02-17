package model

import "time"

type InventoryBatch struct {
	ID                 string    `json:"id"`
	ItemID             string    `json:"item_id"`
	QuantityTotal      float64   `json:"quantity_total"`
	QuantityRemaining  float64   `json:"quantity_remaining"`
	PurchasePriceTotal float64   `json:"purchase_price_total"`
	UnitPrice          float64   `json:"unit_price"`
	PurchasedAt        time.Time `json:"purchased_at"`
}

type Item struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Unit          string    `json:"unit"`
	MinStockAlert float64   `json:"min_stock_alert"`
	CreatedAt     time.Time `json:"created_at"`
}

type Product struct {
	ID               string    `json:"id"`
	RecipeID         string    `json:"recipe_id"`
	QuantityProduced int       `json:"quantity_produced"`
	UnitCost         float64   `json:"unit_cost"`
	SalePrice        float64   `json:"sale_price"`
	ProducedAt       time.Time `json:"produced_at"`
}

type Recipe struct {
	ID                  string    `json:"id"`
	Name                string    `json:"name"`
	ProfitMarginPercent float64   `json:"profit_margin_percent"`
	CreatedAt           time.Time `json:"created_at"`
}

type Sale struct {
	ID           string    `json:"id"`
	ProductID    string    `json:"product_id"`
	QuantitySold int       `json:"quantity_sold"`
	UnitPrice    float64   `json:"unit_price"`
	TotalPrice   float64   `json:"total_price"`
	SoldAt       time.Time `json:"sold_at"`
}

type RecipeIngredient struct {
	RecipeID       string  `json:"recipe_id"`
	ItemID         string  `json:"item_id"`
	QuantityNeeded float64 `json:"quantity_needed"`
}
