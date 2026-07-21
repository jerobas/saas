package catalog

import "github.com/jerobas/saas/internal/domain"

type Capabilities struct {
	purchasable bool
	producible  bool
	sellable    bool
}

func NewCapabilities(purchasable, producible, sellable bool) Capabilities {
	return Capabilities{purchasable: purchasable, producible: producible, sellable: sellable}
}

func (c Capabilities) Purchasable() bool { return c.purchasable }
func (c Capabilities) Producible() bool  { return c.producible }
func (c Capabilities) Sellable() bool    { return c.sellable }
func (c Capabilities) Any() bool         { return c.purchasable || c.producible || c.sellable }

type MeasurementUnitParams struct {
	Code       domain.UnitCode
	Name       domain.DisplayName
	Symbol     domain.NonEmptyText
	Dimension  domain.Dimension
	Conversion domain.UnitConversion
	ItemBase   bool
	Seeded     bool
}

type MeasurementUnit struct {
	code       domain.UnitCode
	name       domain.DisplayName
	symbol     domain.NonEmptyText
	dimension  domain.Dimension
	conversion domain.UnitConversion
	itemBase   bool
	seeded     bool
}

func NewMeasurementUnit(params MeasurementUnitParams) (MeasurementUnit, error) {
	violations := make([]domain.Violation, 0, 5)
	if params.Code.String() == "" {
		violations = append(violations, required("code"))
	}
	if params.Name.String() == "" {
		violations = append(violations, required("name"))
	}
	if params.Symbol.String() == "" {
		violations = append(violations, required("symbol"))
	}
	if _, err := domain.ParseDimension(params.Dimension.String()); err != nil {
		violations = append(violations, domain.Violation{Field: "dimension", Code: domain.ViolationInvalidEnum, InvariantID: "UNIT-002"})
	}
	if params.Conversion.IsZero() {
		violations = append(violations, required("conversion"))
	}
	if err := domain.NewValidationError(violations...); err != nil {
		return MeasurementUnit{}, err
	}
	return MeasurementUnit{
		code: params.Code, name: params.Name, symbol: params.Symbol,
		dimension: params.Dimension, conversion: params.Conversion,
		itemBase: params.ItemBase, seeded: params.Seeded,
	}, nil
}

func (u MeasurementUnit) Code() domain.UnitCode             { return u.code }
func (u MeasurementUnit) Name() domain.DisplayName          { return u.name }
func (u MeasurementUnit) Symbol() domain.NonEmptyText       { return u.symbol }
func (u MeasurementUnit) Dimension() domain.Dimension       { return u.dimension }
func (u MeasurementUnit) Conversion() domain.UnitConversion { return u.conversion }
func (u MeasurementUnit) IsItemBase() bool                  { return u.itemBase }
func (u MeasurementUnit) IsSeeded() bool                    { return u.seeded }

type ItemPackagingParams struct {
	ID          domain.PackagingID
	ItemID      domain.ItemID
	Name        domain.UniqueName
	EnteredUnit domain.UnitCode
	Conversion  domain.UnitConversion
	CreatedAt   domain.UTCInstant
	UpdatedAt   domain.UTCInstant
	ArchivedAt  domain.Option[domain.UTCInstant]
}

type ItemPackaging struct {
	id          domain.PackagingID
	itemID      domain.ItemID
	name        domain.UniqueName
	enteredUnit domain.UnitCode
	conversion  domain.UnitConversion
	createdAt   domain.UTCInstant
	updatedAt   domain.UTCInstant
	archivedAt  domain.Option[domain.UTCInstant]
}

func NewItemPackaging(params ItemPackagingParams) (ItemPackaging, error) {
	violations := make([]domain.Violation, 0, 6)
	if params.ID.IsZero() {
		violations = append(violations, required("packaging_id"))
	}
	if params.ItemID.IsZero() {
		violations = append(violations, required("item_id"))
	}
	if params.Name.Display() == "" || params.Name.Key() == "" {
		violations = append(violations, required("name"))
	}
	if params.EnteredUnit.String() == "" {
		violations = append(violations, required("entered_unit_code"))
	}
	if params.Conversion.IsZero() {
		violations = append(violations, required("conversion"))
	}
	if err := domain.ValidateTimestampOrder(params.CreatedAt, params.UpdatedAt, params.ArchivedAt); err != nil {
		violations = append(violations, validationViolations(err)...)
	}
	if err := domain.NewValidationError(violations...); err != nil {
		return ItemPackaging{}, err
	}
	return ItemPackaging{
		id: params.ID, itemID: params.ItemID, name: params.Name,
		enteredUnit: params.EnteredUnit, conversion: params.Conversion, createdAt: params.CreatedAt,
		updatedAt: params.UpdatedAt, archivedAt: params.ArchivedAt,
	}, nil
}

func (p ItemPackaging) ID() domain.PackagingID                       { return p.id }
func (p ItemPackaging) ItemID() domain.ItemID                        { return p.itemID }
func (p ItemPackaging) Name() domain.UniqueName                      { return p.name }
func (p ItemPackaging) EnteredUnit() domain.UnitCode                 { return p.enteredUnit }
func (p ItemPackaging) Conversion() domain.UnitConversion            { return p.conversion }
func (p ItemPackaging) CreatedAt() domain.UTCInstant                 { return p.createdAt }
func (p ItemPackaging) UpdatedAt() domain.UTCInstant                 { return p.updatedAt }
func (p ItemPackaging) ArchivedAt() domain.Option[domain.UTCInstant] { return p.archivedAt }
func (p ItemPackaging) IsArchived() bool                             { return p.archivedAt.IsSome() }

type ItemParams struct {
	ID               domain.ItemID
	Name             domain.UniqueName
	SKU              domain.Option[domain.SKU]
	Description      domain.Option[domain.NonEmptyText]
	BaseUnit         domain.UnitCode
	Capabilities     Capabilities
	DefaultSalePrice domain.Option[domain.MinorAmount]
	ReorderQuantity  domain.Option[domain.AtomicQuantity]
	CreatedAt        domain.UTCInstant
	UpdatedAt        domain.UTCInstant
	ArchivedAt       domain.Option[domain.UTCInstant]
	Packagings       []ItemPackaging
}

// Item is the catalog aggregate returned by SQLite adapters. Packagings are
// immutable snapshots and are always copied at the aggregate boundary.
type Item struct {
	id               domain.ItemID
	name             domain.UniqueName
	sku              domain.Option[domain.SKU]
	description      domain.Option[domain.NonEmptyText]
	baseUnit         domain.UnitCode
	capabilities     Capabilities
	defaultSalePrice domain.Option[domain.MinorAmount]
	reorderQuantity  domain.Option[domain.AtomicQuantity]
	createdAt        domain.UTCInstant
	updatedAt        domain.UTCInstant
	archivedAt       domain.Option[domain.UTCInstant]
	packagings       []ItemPackaging
}

func NewItem(params ItemParams) (Item, error) {
	violations := make([]domain.Violation, 0, 10)
	if params.ID.IsZero() {
		violations = append(violations, required("item_id"))
	}
	if params.Name.Display() == "" || params.Name.Key() == "" {
		violations = append(violations, required("name"))
	}
	if sku, ok := params.SKU.Get(); ok && (sku.Display() == "" || sku.Key() == "") {
		violations = append(violations, required("sku"))
	}
	if description, ok := params.Description.Get(); ok && description.String() == "" {
		violations = append(violations, required("description"))
	}
	if params.BaseUnit.String() == "" {
		violations = append(violations, required("base_unit_code"))
	}
	if params.ArchivedAt.IsNone() && !params.Capabilities.Any() {
		violations = append(violations, domain.Violation{Field: "capabilities", Code: domain.ViolationRequired, InvariantID: "CAT-002"})
	}
	if params.DefaultSalePrice.IsSome() && !params.Capabilities.Sellable() {
		violations = append(violations, domain.Violation{Field: "default_sale_price", Code: domain.ViolationInvariant, InvariantID: "CAT-004"})
	}
	if err := domain.ValidateTimestampOrder(params.CreatedAt, params.UpdatedAt, params.ArchivedAt); err != nil {
		violations = append(violations, validationViolations(err)...)
	}
	seenIDs := make(map[int64]struct{}, len(params.Packagings))
	seenNames := make(map[string]struct{}, len(params.Packagings))
	for _, packaging := range params.Packagings {
		if packaging.ID().IsZero() || packaging.ItemID().IsZero() {
			violations = append(violations, domain.Violation{Field: "packagings", Code: domain.ViolationInvariant, InvariantID: "CAT-006"})
			continue
		}
		if packaging.ItemID() != params.ID {
			violations = append(violations, domain.Violation{Field: "packagings.item_id", Code: domain.ViolationInvariant, InvariantID: "CAT-006"})
		}
		if _, found := seenIDs[packaging.ID().Int64()]; found {
			violations = append(violations, domain.Violation{Field: "packagings.id", Code: domain.ViolationDuplicate, InvariantID: "CAT-006"})
		}
		seenIDs[packaging.ID().Int64()] = struct{}{}
		if _, found := seenNames[packaging.Name().Key()]; found {
			violations = append(violations, domain.Violation{Field: "packagings.name", Code: domain.ViolationDuplicate, InvariantID: "CAT-006"})
		}
		seenNames[packaging.Name().Key()] = struct{}{}
	}
	if err := domain.NewValidationError(violations...); err != nil {
		return Item{}, err
	}
	return Item{
		id: params.ID, name: params.Name, sku: params.SKU,
		description: params.Description, baseUnit: params.BaseUnit,
		capabilities:     params.Capabilities,
		defaultSalePrice: params.DefaultSalePrice, reorderQuantity: params.ReorderQuantity,
		createdAt: params.CreatedAt, updatedAt: params.UpdatedAt,
		archivedAt: params.ArchivedAt,
		packagings: clonePackagings(params.Packagings),
	}, nil
}

func (i Item) ID() domain.ItemID                                     { return i.id }
func (i Item) Name() domain.UniqueName                               { return i.name }
func (i Item) SKU() domain.Option[domain.SKU]                        { return i.sku }
func (i Item) Description() domain.Option[domain.NonEmptyText]       { return i.description }
func (i Item) BaseUnit() domain.UnitCode                             { return i.baseUnit }
func (i Item) Capabilities() Capabilities                            { return i.capabilities }
func (i Item) DefaultSalePrice() domain.Option[domain.MinorAmount]   { return i.defaultSalePrice }
func (i Item) ReorderQuantity() domain.Option[domain.AtomicQuantity] { return i.reorderQuantity }
func (i Item) CreatedAt() domain.UTCInstant                          { return i.createdAt }
func (i Item) UpdatedAt() domain.UTCInstant                          { return i.updatedAt }
func (i Item) ArchivedAt() domain.Option[domain.UTCInstant]          { return i.archivedAt }
func (i Item) IsArchived() bool                                      { return i.archivedAt.IsSome() }
func (i Item) Packagings() []ItemPackaging                           { return clonePackagings(i.packagings) }

func required(field string) domain.Violation {
	return domain.Violation{Field: field, Code: domain.ViolationRequired}
}

func validationViolations(err error) []domain.Violation {
	if validation, ok := err.(*domain.ValidationError); ok {
		return validation.Violations()
	}
	return []domain.Violation{{Field: "aggregate", Code: domain.ViolationInvariant}}
}

func clonePackagings(source []ItemPackaging) []ItemPackaging {
	result := make([]ItemPackaging, len(source))
	copy(result, source)
	return result
}

// ValidateCompatibleDimensions is used before packaging writes once both
// controlled units have been loaded. Read snapshots do not repeat dimensions
// that are absent from the item_packagings row.
func ValidateCompatibleDimensions(base, entered domain.Dimension) error {
	if _, err := domain.ParseDimension(base.String()); err != nil {
		return err
	}
	if _, err := domain.ParseDimension(entered.String()); err != nil {
		return err
	}
	if base != entered {
		return domain.Invalid("entered_unit_code", domain.ViolationIncompatibleDimension, "CAT-006")
	}
	return nil
}
