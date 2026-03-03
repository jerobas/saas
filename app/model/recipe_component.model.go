package model

import "time"

type RecipeComponent struct {
	ID         int64      `json:"id"`
	RecipeID   int64      `json:"recipe_id"`
	ItemID     int64      `json:"item_id"`
	Quantity   float64    `json:"quantity"`
	CreatedAt  time.Time  `json:"created_at"`
}

type RecipeComponentInsertDTO struct {
	RecipeID   int64      `json:"recipe_id"`
	ItemID     int64      `json:"item_id"`
	Quantity   float64    `json:"quantity"`
}