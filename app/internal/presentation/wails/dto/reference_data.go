package dto

type MeasurementUnitResponse struct {
	Code            string `json:"code"`
	Name            string `json:"name"`
	Symbol          string `json:"symbol"`
	Dimension       string `json:"dimension"`
	NumeratorAtomic int64  `json:"numeratorAtomic"`
	Denominator     int64  `json:"denominator"`
	IsItemBase      bool   `json:"isItemBase"`
	IsSeeded        bool   `json:"isSeeded"`
}
