package dto

type SettingsResponse struct {
	BusinessName        string `json:"businessName"`
	Locale              string `json:"locale"`
	Timezone            string `json:"timezone"`
	CurrencyCode        string `json:"currencyCode"`
	CurrencyMinorDigits int64  `json:"currencyMinorDigits"`
	HourlyLaborCost     *int64 `json:"hourlyLaborCost,omitempty"`
	DefaultGrossMargin  *int64 `json:"defaultGrossMargin,omitempty"`
	CreatedAtMs         int64  `json:"createdAtMs"`
	UpdatedAtMs         int64  `json:"updatedAtMs"`
}

type SettingsUpdateRequest struct {
	BusinessName        string `json:"businessName"`
	Locale              string `json:"locale"`
	Timezone            string `json:"timezone"`
	CurrencyCode        string `json:"currencyCode"`
	CurrencyMinorDigits int64  `json:"currencyMinorDigits"`
	HourlyLaborCost     *int64 `json:"hourlyLaborCost,omitempty"`
	DefaultGrossMargin  *int64 `json:"defaultGrossMargin,omitempty"`
	ExpectedUpdatedAtMs int64  `json:"expectedUpdatedAtMs"`
}
