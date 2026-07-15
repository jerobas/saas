package dto

type CounterpartyResponse struct {
	ID           int64    `json:"id"`
	Name         string   `json:"name"`
	Phone        *string  `json:"phone,omitempty"`
	Email        *string  `json:"email,omitempty"`
	Notes        *string  `json:"notes,omitempty"`
	Roles        []string `json:"roles"`
	CreatedAtMs  int64    `json:"createdAtMs"`
	UpdatedAtMs  int64    `json:"updatedAtMs"`
	ArchivedAtMs *int64   `json:"archivedAtMs,omitempty"`
}

type CounterpartyCursorRequest struct {
	Name string `json:"name"`
	ID   int64  `json:"id"`
}

type CounterpartyCursorResponse struct {
	Name string `json:"name"`
	ID   int64  `json:"id"`
}

type CounterpartyListRequest struct {
	ArchiveFilter string                     `json:"archiveFilter,omitempty"`
	Role          *string                    `json:"role,omitempty"`
	Search        *string                    `json:"search,omitempty"`
	After         *CounterpartyCursorRequest `json:"after,omitempty"`
	PageSize      int                        `json:"pageSize,omitempty"`
}

type CounterpartyPageResponse struct {
	Items []CounterpartyResponse      `json:"items"`
	Next  *CounterpartyCursorResponse `json:"next,omitempty"`
}

type CounterpartyWriteRequest struct {
	Name  string   `json:"name"`
	Phone *string  `json:"phone,omitempty"`
	Email *string  `json:"email,omitempty"`
	Notes *string  `json:"notes,omitempty"`
	Roles []string `json:"roles"`
}

type CounterpartyUpdateRequest struct {
	CounterpartyWriteRequest
	ExpectedUpdatedAtMs int64 `json:"expectedUpdatedAtMs"`
}

type VersionedCounterpartyRequest struct {
	ExpectedUpdatedAtMs int64 `json:"expectedUpdatedAtMs"`
}
