package recipe

import (
	"unicode/utf8"

	"github.com/jerobas/saas/internal/domain"
)

type ComponentParams struct {
	ID                   domain.RecipeComponentID
	RevisionID           domain.RecipeRevisionID
	Order                domain.ComponentOrder
	ItemID               domain.ItemID
	Quantity             domain.AtomicQuantity
	EnteredUnit          domain.UnitCode
	EnteredPackagingName domain.Option[domain.NonEmptyText]
	Conversion           domain.UnitConversion
	CreatedAt            domain.UTCInstant
}

type Component struct {
	id                   domain.RecipeComponentID
	revisionID           domain.RecipeRevisionID
	order                domain.ComponentOrder
	itemID               domain.ItemID
	quantity             domain.AtomicQuantity
	enteredUnit          domain.UnitCode
	enteredPackagingName domain.Option[domain.NonEmptyText]
	conversion           domain.UnitConversion
	createdAt            domain.UTCInstant
}

func NewComponent(params ComponentParams) (Component, error) {
	violations := make([]domain.Violation, 0, 8)
	if params.ID.IsZero() {
		violations = append(violations, required("component_id"))
	}
	if params.RevisionID.IsZero() {
		violations = append(violations, required("recipe_revision_id"))
	}
	if params.Order.IsZero() {
		violations = append(violations, required("component_order"))
	}
	if params.ItemID.IsZero() {
		violations = append(violations, required("item_id"))
	}
	if params.Quantity.Int64() <= 0 {
		violations = append(violations, domain.Violation{Field: "quantity_atomic", Code: domain.ViolationNotPositive, InvariantID: "REC-003"})
	}
	if params.EnteredUnit.String() == "" {
		violations = append(violations, required("entered_unit_code"))
	}
	if name, ok := params.EnteredPackagingName.Get(); ok && name.String() == "" {
		violations = append(violations, required("entered_packaging_name"))
	}
	if params.Conversion.IsZero() {
		violations = append(violations, required("conversion"))
	}
	if params.CreatedAt.IsZero() {
		violations = append(violations, required("created_at"))
	}
	if err := domain.NewValidationError(violations...); err != nil {
		return Component{}, err
	}
	return Component{
		id: params.ID, revisionID: params.RevisionID, order: params.Order,
		itemID: params.ItemID, quantity: params.Quantity, enteredUnit: params.EnteredUnit,
		enteredPackagingName: params.EnteredPackagingName, conversion: params.Conversion,
		createdAt: params.CreatedAt,
	}, nil
}

func (c Component) ID() domain.RecipeComponentID        { return c.id }
func (c Component) RevisionID() domain.RecipeRevisionID { return c.revisionID }
func (c Component) Order() domain.ComponentOrder        { return c.order }
func (c Component) ItemID() domain.ItemID               { return c.itemID }
func (c Component) Quantity() domain.AtomicQuantity     { return c.quantity }
func (c Component) EnteredUnit() domain.UnitCode        { return c.enteredUnit }
func (c Component) EnteredPackagingName() domain.Option[domain.NonEmptyText] {
	return c.enteredPackagingName
}
func (c Component) Conversion() domain.UnitConversion { return c.conversion }
func (c Component) CreatedAt() domain.UTCInstant      { return c.createdAt }

type RevisionParams struct {
	ID                  domain.RecipeRevisionID
	RecipeID            domain.RecipeID
	Number              domain.RevisionNumber
	StandardYield       domain.AtomicQuantity
	Instructions        string
	PreparationTime     domain.PreparationMinutes
	EstimatedDirectCost domain.Option[domain.InventoryValue]
	CreatedAt           domain.UTCInstant
	Components          []Component
}

// Revision is an immutable published recipe snapshot. Its component slice is
// copied both on construction and access.
type Revision struct {
	id                  domain.RecipeRevisionID
	recipeID            domain.RecipeID
	number              domain.RevisionNumber
	standardYield       domain.AtomicQuantity
	instructions        string
	preparationTime     domain.PreparationMinutes
	estimatedDirectCost domain.Option[domain.InventoryValue]
	createdAt           domain.UTCInstant
	components          []Component
}

func NewRevision(params RevisionParams) (Revision, error) {
	violations := make([]domain.Violation, 0, 8)
	if params.ID.IsZero() {
		violations = append(violations, required("recipe_revision_id"))
	}
	if params.RecipeID.IsZero() {
		violations = append(violations, required("recipe_id"))
	}
	if params.Number.IsZero() {
		violations = append(violations, required("revision_number"))
	}
	if params.StandardYield.Int64() <= 0 {
		violations = append(violations, domain.Violation{Field: "standard_yield_quantity_atomic", Code: domain.ViolationNotPositive, InvariantID: "REC-003"})
	}
	if !utf8.ValidString(params.Instructions) {
		violations = append(violations, domain.Violation{Field: "instructions", Code: domain.ViolationInvalidFormat})
	}
	if params.CreatedAt.IsZero() {
		violations = append(violations, required("created_at"))
	}
	if len(params.Components) == 0 {
		violations = append(violations, domain.Violation{Field: "components", Code: domain.ViolationRequired, InvariantID: "REC-002"})
	}
	seenOrders := make(map[int64]struct{}, len(params.Components))
	seenItems := make(map[int64]struct{}, len(params.Components))
	for _, component := range params.Components {
		if component.ID().IsZero() || component.RevisionID().IsZero() {
			violations = append(violations, domain.Violation{Field: "components", Code: domain.ViolationInvariant, InvariantID: "REC-002"})
			continue
		}
		if component.RevisionID() != params.ID {
			violations = append(violations, domain.Violation{Field: "components.recipe_revision_id", Code: domain.ViolationInvariant, InvariantID: "REC-002"})
		}
		if _, found := seenOrders[component.Order().Int64()]; found {
			violations = append(violations, domain.Violation{Field: "components.order", Code: domain.ViolationDuplicate, InvariantID: "REC-002"})
		}
		seenOrders[component.Order().Int64()] = struct{}{}
		if _, found := seenItems[component.ItemID().Int64()]; found {
			violations = append(violations, domain.Violation{Field: "components.item_id", Code: domain.ViolationDuplicate, InvariantID: "REC-002"})
		}
		seenItems[component.ItemID().Int64()] = struct{}{}
	}
	if err := domain.NewValidationError(violations...); err != nil {
		return Revision{}, err
	}
	return Revision{
		id: params.ID, recipeID: params.RecipeID, number: params.Number,
		standardYield: params.StandardYield, instructions: params.Instructions,
		preparationTime:     params.PreparationTime,
		estimatedDirectCost: params.EstimatedDirectCost, createdAt: params.CreatedAt,
		components: cloneComponents(params.Components),
	}, nil
}

func (r Revision) ID() domain.RecipeRevisionID                { return r.id }
func (r Revision) RecipeID() domain.RecipeID                  { return r.recipeID }
func (r Revision) Number() domain.RevisionNumber              { return r.number }
func (r Revision) StandardYield() domain.AtomicQuantity       { return r.standardYield }
func (r Revision) Instructions() string                       { return r.instructions }
func (r Revision) PreparationTime() domain.PreparationMinutes { return r.preparationTime }
func (r Revision) EstimatedDirectCost() domain.Option[domain.InventoryValue] {
	return r.estimatedDirectCost
}
func (r Revision) CreatedAt() domain.UTCInstant { return r.createdAt }
func (r Revision) Components() []Component      { return cloneComponents(r.components) }

type RevisionSummary struct {
	id                  domain.RecipeRevisionID
	recipeID            domain.RecipeID
	number              domain.RevisionNumber
	standardYield       domain.AtomicQuantity
	estimatedDirectCost domain.Option[domain.InventoryValue]
	createdAt           domain.UTCInstant
}

func (r Revision) Summary() RevisionSummary {
	return RevisionSummary{
		id: r.id, recipeID: r.recipeID, number: r.number,
		standardYield: r.standardYield, estimatedDirectCost: r.estimatedDirectCost,
		createdAt: r.createdAt,
	}
}
func (s RevisionSummary) ID() domain.RecipeRevisionID          { return s.id }
func (s RevisionSummary) RecipeID() domain.RecipeID            { return s.recipeID }
func (s RevisionSummary) Number() domain.RevisionNumber        { return s.number }
func (s RevisionSummary) StandardYield() domain.AtomicQuantity { return s.standardYield }
func (s RevisionSummary) EstimatedDirectCost() domain.Option[domain.InventoryValue] {
	return s.estimatedDirectCost
}
func (s RevisionSummary) CreatedAt() domain.UTCInstant { return s.createdAt }

type Params struct {
	ID              domain.RecipeID
	Name            domain.UniqueName
	OutputItemID    domain.ItemID
	CreatedAt       domain.UTCInstant
	UpdatedAt       domain.UTCInstant
	ArchivedAt      domain.Option[domain.UTCInstant]
	CurrentRevision Revision
}

type Recipe struct {
	id              domain.RecipeID
	name            domain.UniqueName
	outputItemID    domain.ItemID
	createdAt       domain.UTCInstant
	updatedAt       domain.UTCInstant
	archivedAt      domain.Option[domain.UTCInstant]
	currentRevision Revision
}

func New(params Params) (Recipe, error) {
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
	if params.CurrentRevision.ID().IsZero() || params.CurrentRevision.RecipeID() != params.ID {
		violations = append(violations, domain.Violation{Field: "current_revision", Code: domain.ViolationInvariant, InvariantID: "REC-002"})
	} else {
		for _, component := range params.CurrentRevision.Components() {
			if component.ItemID() == params.OutputItemID {
				violations = append(violations, domain.Violation{Field: "components.item_id", Code: domain.ViolationInvariant, InvariantID: "REC-004"})
			}
		}
		if params.CurrentRevision.CreatedAt().Before(params.CreatedAt) {
			violations = append(violations, domain.Violation{Field: "current_revision.created_at", Code: domain.ViolationInvariant, InvariantID: "REC-002"})
		}
		if params.UpdatedAt.Before(params.CurrentRevision.CreatedAt()) {
			violations = append(violations, domain.Violation{Field: "updated_at", Code: domain.ViolationInvariant, InvariantID: "REC-005"})
		}
	}
	if err := domain.ValidateTimestampOrder(params.CreatedAt, params.UpdatedAt, params.ArchivedAt); err != nil {
		if validation, ok := err.(*domain.ValidationError); ok {
			violations = append(violations, validation.Violations()...)
		}
	}
	if err := domain.NewValidationError(violations...); err != nil {
		return Recipe{}, err
	}
	return Recipe{
		id: params.ID, name: params.Name, outputItemID: params.OutputItemID,
		createdAt: params.CreatedAt, updatedAt: params.UpdatedAt,
		archivedAt: params.ArchivedAt, currentRevision: params.CurrentRevision,
	}, nil
}

func (r Recipe) ID() domain.RecipeID                          { return r.id }
func (r Recipe) Name() domain.UniqueName                      { return r.name }
func (r Recipe) OutputItemID() domain.ItemID                  { return r.outputItemID }
func (r Recipe) CreatedAt() domain.UTCInstant                 { return r.createdAt }
func (r Recipe) UpdatedAt() domain.UTCInstant                 { return r.updatedAt }
func (r Recipe) ArchivedAt() domain.Option[domain.UTCInstant] { return r.archivedAt }
func (r Recipe) IsArchived() bool                             { return r.archivedAt.IsSome() }
func (r Recipe) CurrentRevision() Revision                    { return r.currentRevision }

func required(field string) domain.Violation {
	return domain.Violation{Field: field, Code: domain.ViolationRequired}
}

func cloneComponents(source []Component) []Component {
	result := make([]Component, len(source))
	copy(result, source)
	return result
}
