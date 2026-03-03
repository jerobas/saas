package model

import "time"


type Recipe struct {
	ID                      int64           `json:"id"`
	Name                    string          `json:"name"`
	OutputItemId            int64           `json:"output_item_id"`
	PreparationTimeMinutes  int64           `json:"preparation_time_minutes"`
	Instructions            string          `json:"instructions"`
	StandardYieldQuantity   float64         `json:"standard_yield_quantity"`
	CreatedAt               time.Time       `json:"created_at"`
}

type RecipeInsertDTO struct {
	Name                    string          `json:"name"`
	OutputItemId            int64           `json:"output_item_id"`
	PreparationTimeMinutes  int64           `json:"preparation_time_minutes"`
	Instructions            string          `json:"instructions"`
	StandardYieldQuantity   float64         `json:"standard_yield_quantity"`
}