package dto

type RecipeCursorRequest struct {
	Name string `json:"name"`
	ID   int64  `json:"id"`
}

type RecipeCursorResponse struct {
	Name string `json:"name"`
	ID   int64  `json:"id"`
}

type RecipeListRequest struct {
	ArchiveFilter string               `json:"archiveFilter,omitempty"`
	Search        *string              `json:"search,omitempty"`
	After         *RecipeCursorRequest `json:"after,omitempty"`
	PageSize      int                  `json:"pageSize,omitempty"`
}

type RecipePageResponse struct {
	Items []RecipeSummaryResponse `json:"items"`
	Next  *RecipeCursorResponse   `json:"next,omitempty"`
}

type CurrentRecipeRevisionSummaryResponse struct {
	ID                    int64 `json:"id"`
	Number                int64 `json:"number"`
	StandardYieldQuantity int64 `json:"standardYieldQuantityAtomic"`
}

type RecipeSummaryResponse struct {
	ID              int64                                `json:"id"`
	Name            string                               `json:"name"`
	OutputItemID    int64                                `json:"outputItemId"`
	OutputItemName  string                               `json:"outputItemName"`
	CreatedAtMs     int64                                `json:"createdAtMs"`
	UpdatedAtMs     int64                                `json:"updatedAtMs"`
	ArchivedAtMs    *int64                               `json:"archivedAtMs,omitempty"`
	CurrentRevision CurrentRecipeRevisionSummaryResponse `json:"currentRevision"`
}

type RecipeResponse struct {
	ID              int64                  `json:"id"`
	Name            string                 `json:"name"`
	OutputItemID    int64                  `json:"outputItemId"`
	CreatedAtMs     int64                  `json:"createdAtMs"`
	UpdatedAtMs     int64                  `json:"updatedAtMs"`
	ArchivedAtMs    *int64                 `json:"archivedAtMs,omitempty"`
	CurrentRevision RecipeRevisionResponse `json:"currentRevision"`
}

type RecipeCreateRequest struct {
	Name         string                     `json:"name"`
	OutputItemID int64                      `json:"outputItemId"`
	Revision     RecipeRevisionWriteRequest `json:"revision"`
}

type RecipePublishRevisionRequest struct {
	ExpectedLatestRevision int64                      `json:"expectedLatestRevision"`
	ExpectedUpdatedAtMs    int64                      `json:"expectedUpdatedAtMs"`
	Revision               RecipeRevisionWriteRequest `json:"revision"`
}

type RecipeRenameRequest struct {
	Name                string `json:"name"`
	ExpectedUpdatedAtMs int64  `json:"expectedUpdatedAtMs"`
}

type RecipeRevisionWriteRequest struct {
	StandardYieldQuantity    int64                    `json:"standardYieldQuantityAtomic"`
	Instructions             string                   `json:"instructions"`
	PreparationTimeMinutes   int64                    `json:"preparationTimeMinutes"`
	EstimatedDirectCostMicro *int64                   `json:"estimatedDirectCostMicro,omitempty"`
	Components               []RecipeComponentRequest `json:"components"`
}

type RecipeComponentRequest struct {
	Order          int64   `json:"order"`
	ItemID         int64   `json:"itemId"`
	QuantityAtomic int64   `json:"quantityAtomic"`
	SourceType     string  `json:"sourceType"`
	UnitCode       *string `json:"unitCode,omitempty"`
	PackagingID    *int64  `json:"packagingId,omitempty"`
}

type RecipeRevisionResponse struct {
	ID                       int64                     `json:"id"`
	RecipeID                 int64                     `json:"recipeId"`
	Number                   int64                     `json:"number"`
	StandardYieldQuantity    int64                     `json:"standardYieldQuantityAtomic"`
	Instructions             string                    `json:"instructions"`
	PreparationTimeMinutes   int64                     `json:"preparationTimeMinutes"`
	EstimatedDirectCostMicro *int64                    `json:"estimatedDirectCostMicro,omitempty"`
	CreatedAtMs              int64                     `json:"createdAtMs"`
	Components               []RecipeComponentResponse `json:"components"`
}

type RecipeComponentResponse struct {
	ID                        int64   `json:"id"`
	RevisionID                int64   `json:"revisionId"`
	Order                     int64   `json:"order"`
	ItemID                    int64   `json:"itemId"`
	QuantityAtomic            int64   `json:"quantityAtomic"`
	EnteredUnitCode           string  `json:"enteredUnitCode"`
	EnteredPackagingName      *string `json:"enteredPackagingName,omitempty"`
	ConversionNumeratorAtomic int64   `json:"conversionNumeratorAtomic"`
	ConversionDenominator     int64   `json:"conversionDenominator"`
	CreatedAtMs               int64   `json:"createdAtMs"`
}
