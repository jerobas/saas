package model

import "time"

type SaleLine struct {
	ID           int64     `json:"id"`
	EventID      int64     `json:"event_id"`
	ItemID       int64     `json:"item_id"`
	Quantity     float64   `json:"quantity"`
	UnitPrice    int64     `json:"unit_price"`
	CreatedAt    time.Time `json:"created_at"`
}

type SaleLineInsertDTO struct {
	EventID      int64     `json:"event_id"`
	ItemID       int64     `json:"item_id"`
	Quantity     float64   `json:"quantity"`
	UnitPrice    int64     `json:"unit_price"`
}

// type SaleLineUpdateDTO struct {
// 	EventID      *int64    `json:"event_id"`
// 	ItemID       *int64    `json:"item_id"`
// 	Quantity     *float64  `json:"quantity"`
// 	UnitPrice    *int64    `json:"unit_price"`
// }

// func (saleLine *SaleLine) ToInsertDTO() *SaleLineInsertDTO {
// 	if saleLine == nil {
// 		return nil
// 	}
// 	return &saleLineInsertDTO{
// 		ID:           saleLine.ID,
// 		ProductID:    saleLine.ProductID,
// 		QuantitySold: saleLine.QuantitySold,
// 		UnitPrice:    saleLine.UnitPrice,
// 		TotalPrice:   saleLine.TotalPrice,
// 		SoldAt:       saleLine.SoldAt.Format(time.RFC3339),
// 	}
// }

// func (saleLines []*SaleLine) ToInstertDTOList []*saleLineInsertDTO {
// 	dtos := make([]*saleLineInsertDTO, len(saleLines))
// 	for i, saleLine := range saleLines {
// 		dtos[i] = saleLine.t
// 	}
// 	return dtos
// }