package recipe

import "github.com/jerobas/saas/internal/domain"

type CurrentRevisionSummaryParams struct {
	ID            domain.RecipeRevisionID
	Number        domain.RevisionNumber
	StandardYield domain.AtomicQuantity
}

// CurrentRevisionSummary is the compact, non-null revision projection used by
// recipe lists. A recipe without one is an invalid persisted aggregate.
type CurrentRevisionSummary struct {
	id            domain.RecipeRevisionID
	number        domain.RevisionNumber
	standardYield domain.AtomicQuantity
}

func NewCurrentRevisionSummary(params CurrentRevisionSummaryParams) (CurrentRevisionSummary, error) {
	violations := make([]domain.Violation, 0, 3)
	if params.ID.IsZero() {
		violations = append(violations, required("current_revision_id"))
	}
	if params.Number.IsZero() {
		violations = append(violations, required("current_revision_number"))
	}
	if params.StandardYield.Int64() <= 0 {
		violations = append(violations, domain.Violation{Field: "current_standard_yield_quantity_atomic", Code: domain.ViolationNotPositive, InvariantID: "REC-003"})
	}
	if err := domain.NewValidationError(violations...); err != nil {
		return CurrentRevisionSummary{}, err
	}
	return CurrentRevisionSummary{
		id: params.ID, number: params.Number, standardYield: params.StandardYield,
	}, nil
}

func (s CurrentRevisionSummary) ID() domain.RecipeRevisionID          { return s.id }
func (s CurrentRevisionSummary) Number() domain.RevisionNumber        { return s.number }
func (s CurrentRevisionSummary) StandardYield() domain.AtomicQuantity { return s.standardYield }

type RecipeSummaryParams struct {
	ID              domain.RecipeID
	Name            domain.UniqueName
	OutputItemID    domain.ItemID
	OutputItemName  domain.DisplayName
	CreatedAt       domain.UTCInstant
	UpdatedAt       domain.UTCInstant
	ArchivedAt      domain.Option[domain.UTCInstant]
	CurrentRevision CurrentRevisionSummary
}

// Summary is the validated ListRecipes projection. It intentionally does not
// pretend to be a full Recipe aggregate with component rows.
type RecipeSummary struct {
	id              domain.RecipeID
	name            domain.UniqueName
	outputItemID    domain.ItemID
	outputItemName  domain.DisplayName
	createdAt       domain.UTCInstant
	updatedAt       domain.UTCInstant
	archivedAt      domain.Option[domain.UTCInstant]
	currentRevision CurrentRevisionSummary
}

func NewRecipeSummary(params RecipeSummaryParams) (RecipeSummary, error) {
	violations := make([]domain.Violation, 0, 6)
	if params.ID.IsZero() {
		violations = append(violations, required("recipe_id"))
	}
	if params.Name.Display() == "" || params.Name.Key() == "" {
		violations = append(violations, required("name"))
	}
	if params.OutputItemID.IsZero() {
		violations = append(violations, required("output_item_id"))
	}
	if params.OutputItemName.String() == "" {
		violations = append(violations, required("output_item_name"))
	}
	if params.CurrentRevision.ID().IsZero() {
		violations = append(violations, domain.Violation{Field: "current_revision", Code: domain.ViolationRequired, InvariantID: "REC-002"})
	}
	if err := domain.ValidateTimestampOrder(params.CreatedAt, params.UpdatedAt, params.ArchivedAt); err != nil {
		if validation, ok := err.(*domain.ValidationError); ok {
			violations = append(violations, validation.Violations()...)
		}
	}
	if err := domain.NewValidationError(violations...); err != nil {
		return RecipeSummary{}, err
	}
	return RecipeSummary{
		id: params.ID, name: params.Name, outputItemID: params.OutputItemID,
		outputItemName: params.OutputItemName, createdAt: params.CreatedAt,
		updatedAt: params.UpdatedAt, archivedAt: params.ArchivedAt,
		currentRevision: params.CurrentRevision,
	}, nil
}

func (s RecipeSummary) ID() domain.RecipeID                          { return s.id }
func (s RecipeSummary) Name() domain.UniqueName                      { return s.name }
func (s RecipeSummary) OutputItemID() domain.ItemID                  { return s.outputItemID }
func (s RecipeSummary) OutputItemName() domain.DisplayName           { return s.outputItemName }
func (s RecipeSummary) CreatedAt() domain.UTCInstant                 { return s.createdAt }
func (s RecipeSummary) UpdatedAt() domain.UTCInstant                 { return s.updatedAt }
func (s RecipeSummary) ArchivedAt() domain.Option[domain.UTCInstant] { return s.archivedAt }
func (s RecipeSummary) IsArchived() bool                             { return s.archivedAt.IsSome() }
func (s RecipeSummary) CurrentRevision() CurrentRevisionSummary      { return s.currentRevision }
