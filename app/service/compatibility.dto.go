package service

// These DTOs keep the existing React pages usable while the application is
// migrated bottom-up to the event-ledger domain model.
type ItemDTO struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Unit          string  `json:"unit"`
	MinStockAlert float64 `json:"min_stock_alert"`
	CreatedAt     string  `json:"created_at"`
}

type InventoryBatchDTO struct {
	ID                 string  `json:"id"`
	ItemID             string  `json:"item_id"`
	QuantityTotal      float64 `json:"quantity_total"`
	QuantityRemaining  float64 `json:"quantity_remaining"`
	PurchasePriceTotal float64 `json:"purchase_price_total"`
	UnitPrice          float64 `json:"unit_price"`
	PurchasedAt        string  `json:"purchased_at"`
}

type RecipeIngredientInput struct {
	ItemID   string  `json:"item_id"`
	Quantity float64 `json:"quantity"`
}

type RecipeDTO struct {
	ID          string                  `json:"id"`
	Name        string                  `json:"name"`
	Ingredients []RecipeIngredientInput `json:"ingredients"`
	CreatedAt   string                  `json:"created_at"`
}
