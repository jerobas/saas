package main

import "time"

// ============================================
// DTOs para o Frontend (sem time.Time)
// ============================================

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

type RecipeDTO struct {
	ID                  string  `json:"id"`
	Name                string  `json:"name"`
	ProfitMarginPercent float64 `json:"profit_margin_percent"`
	CreatedAt           string  `json:"created_at"`
}

type ProductDTO struct {
	ID               string  `json:"id"`
	RecipeID         string  `json:"recipe_id"`
	QuantityProduced int     `json:"quantity_produced"`
	UnitCost         float64 `json:"unit_cost"`
	SalePrice        float64 `json:"sale_price"`
	ProducedAt       string  `json:"produced_at"`
}

type SaleDTO struct {
	ID           string  `json:"id"`
	ProductID    string  `json:"product_id"`
	QuantitySold int     `json:"quantity_sold"`
	UnitPrice    float64 `json:"unit_price"`
	TotalPrice   float64 `json:"total_price"`
	SoldAt       string  `json:"sold_at"`
}

// ============================================
// Conversores (Entity -> DTO)
// ============================================

func toItemDTO(item *Item) *ItemDTO {
	if item == nil {
		return nil
	}
	return &ItemDTO{
		ID:            item.ID,
		Name:          item.Name,
		Unit:          item.Unit,
		MinStockAlert: item.MinStockAlert,
		CreatedAt:     item.CreatedAt.Format(time.RFC3339),
	}
}

func toItemDTOList(items []*Item) []*ItemDTO {
	dtos := make([]*ItemDTO, len(items))
	for i, item := range items {
		dtos[i] = toItemDTO(item)
	}
	return dtos
}

func toBatchDTO(batch *InventoryBatch) *InventoryBatchDTO {
	if batch == nil {
		return nil
	}
	return &InventoryBatchDTO{
		ID:                 batch.ID,
		ItemID:             batch.ItemID,
		QuantityTotal:      batch.QuantityTotal,
		QuantityRemaining:  batch.QuantityRemaining,
		PurchasePriceTotal: batch.PurchasePriceTotal,
		UnitPrice:          batch.UnitPrice,
		PurchasedAt:        batch.PurchasedAt.Format(time.RFC3339),
	}
}

func toBatchDTOList(batches []*InventoryBatch) []*InventoryBatchDTO {
	dtos := make([]*InventoryBatchDTO, len(batches))
	for i, batch := range batches {
		dtos[i] = toBatchDTO(batch)
	}
	return dtos
}

func toRecipeDTO(recipe *Recipe) *RecipeDTO {
	if recipe == nil {
		return nil
	}
	return &RecipeDTO{
		ID:                  recipe.ID,
		Name:                recipe.Name,
		ProfitMarginPercent: recipe.ProfitMarginPercent,
		CreatedAt:           recipe.CreatedAt.Format(time.RFC3339),
	}
}

func toRecipeDTOList(recipes []*Recipe) []*RecipeDTO {
	dtos := make([]*RecipeDTO, len(recipes))
	for i, recipe := range recipes {
		dtos[i] = toRecipeDTO(recipe)
	}
	return dtos
}

func toProductDTO(product *Product) *ProductDTO {
	if product == nil {
		return nil
	}
	return &ProductDTO{
		ID:               product.ID,
		RecipeID:         product.RecipeID,
		QuantityProduced: product.QuantityProduced,
		UnitCost:         product.UnitCost,
		SalePrice:        product.SalePrice,
		ProducedAt:       product.ProducedAt.Format(time.RFC3339),
	}
}

func toProductDTOList(products []*Product) []*ProductDTO {
	dtos := make([]*ProductDTO, len(products))
	for i, product := range products {
		dtos[i] = toProductDTO(product)
	}
	return dtos
}

func toSaleDTO(sale *Sale) *SaleDTO {
	if sale == nil {
		return nil
	}
	return &SaleDTO{
		ID:           sale.ID,
		ProductID:    sale.ProductID,
		QuantitySold: sale.QuantitySold,
		UnitPrice:    sale.UnitPrice,
		TotalPrice:   sale.TotalPrice,
		SoldAt:       sale.SoldAt.Format(time.RFC3339),
	}
}

func toSaleDTOList(sales []*Sale) []*SaleDTO {
	dtos := make([]*SaleDTO, len(sales))
	for i, sale := range sales {
		dtos[i] = toSaleDTO(sale)
	}
	return dtos
}
