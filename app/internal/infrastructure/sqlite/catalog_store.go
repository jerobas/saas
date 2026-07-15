package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/domain/catalog"
	"github.com/jerobas/saas/internal/infrastructure/sqlite/sqlcgen"
)

type ItemPageSize struct{ value int64 }

func NewItemPageSize(value int) (ItemPageSize, error) {
	if value < 1 || value > 100 {
		return ItemPageSize{}, domain.Invalid("page_size", domain.ViolationOutOfRange, "")
	}
	return ItemPageSize{value: int64(value)}, nil
}

func (s ItemPageSize) Int() int { return int(s.value) }

type ItemCursor struct {
	Name domain.UniqueName
	ID   domain.ItemID
}

type ItemListFilter struct {
	Archive             domain.ArchiveFilter
	RequireCapabilities catalog.Capabilities
	Search              domain.Option[domain.NonEmptyText]
	After               domain.Option[ItemCursor]
	PageSize            ItemPageSize
}

type PackagingAggregate struct {
	packaging   catalog.ItemPackaging
	baseUnit    catalog.MeasurementUnit
	enteredUnit catalog.MeasurementUnit
}

func (a PackagingAggregate) Packaging() catalog.ItemPackaging     { return a.packaging }
func (a PackagingAggregate) BaseUnit() catalog.MeasurementUnit    { return a.baseUnit }
func (a PackagingAggregate) EnteredUnit() catalog.MeasurementUnit { return a.enteredUnit }

type ItemAggregate struct {
	item       catalog.Item
	baseUnit   catalog.MeasurementUnit
	packagings []PackagingAggregate
}

func (a ItemAggregate) Item() catalog.Item                { return a.item }
func (a ItemAggregate) BaseUnit() catalog.MeasurementUnit { return a.baseUnit }
func (a ItemAggregate) Packagings() []PackagingAggregate {
	result := make([]PackagingAggregate, len(a.packagings))
	copy(result, a.packagings)
	return result
}

type ItemPage struct {
	items []catalog.ItemSummary
	next  domain.Option[ItemCursor]
}

func (p ItemPage) Items() []catalog.ItemSummary {
	result := make([]catalog.ItemSummary, len(p.items))
	copy(result, p.items)
	return result
}
func (p ItemPage) Next() domain.Option[ItemCursor] { return p.next }

type CreateItemInput struct {
	Name             domain.UniqueName
	SKU              domain.Option[domain.SKU]
	Description      domain.Option[domain.NonEmptyText]
	BaseUnit         domain.UnitCode
	Capabilities     catalog.Capabilities
	DefaultSalePrice domain.Option[domain.MinorAmount]
	ReorderQuantity  domain.Option[domain.AtomicQuantity]
	CreatedAt        domain.UTCInstant
	UpdatedAt        domain.UTCInstant
}

type UpdateItemInput struct {
	ID                domain.ItemID
	Name              domain.UniqueName
	SKU               domain.Option[domain.SKU]
	Description       domain.Option[domain.NonEmptyText]
	BaseUnit          domain.UnitCode
	Capabilities      catalog.Capabilities
	DefaultSalePrice  domain.Option[domain.MinorAmount]
	ReorderQuantity   domain.Option[domain.AtomicQuantity]
	ExpectedUpdatedAt domain.UTCInstant
	UpdatedAt         domain.UTCInstant
}

type ArchiveItemInput struct {
	ID                domain.ItemID
	ExpectedUpdatedAt domain.UTCInstant
	ArchivedAt        domain.UTCInstant
}

type RestoreItemInput struct {
	ID                domain.ItemID
	ExpectedUpdatedAt domain.UTCInstant
	UpdatedAt         domain.UTCInstant
}

type CreatePackagingInput struct {
	ItemID      domain.ItemID
	Name        domain.UniqueName
	EnteredUnit domain.UnitCode
	Conversion  domain.UnitConversion
	CreatedAt   domain.UTCInstant
	UpdatedAt   domain.UTCInstant
}

type UpdatePackagingInput struct {
	ID                domain.PackagingID
	Name              domain.UniqueName
	EnteredUnit       domain.UnitCode
	Conversion        domain.UnitConversion
	ExpectedUpdatedAt domain.UTCInstant
	UpdatedAt         domain.UTCInstant
}

type ArchivePackagingInput struct {
	ID                domain.PackagingID
	ExpectedUpdatedAt domain.UTCInstant
	ArchivedAt        domain.UTCInstant
}

type ReconfigureArchivedPackagingInput struct {
	ID                domain.PackagingID
	Name              domain.UniqueName
	EnteredUnit       domain.UnitCode
	Conversion        domain.UnitConversion
	ExpectedUpdatedAt domain.UTCInstant
	UpdatedAt         domain.UTCInstant
}

type RestorePackagingInput struct {
	ID                domain.PackagingID
	ExpectedUpdatedAt domain.UTCInstant
	UpdatedAt         domain.UTCInstant
}

func (s *Store) GetItem(ctx context.Context, id domain.ItemID) (ItemAggregate, error) {
	if id.IsZero() {
		return ItemAggregate{}, domain.Invalid("item_id", domain.ViolationNotPositive, "")
	}
	var value ItemAggregate
	err := s.withReadQueries(ctx, "get item", func(queries *sqlcgen.Queries) error {
		var err error
		value, err = loadItemAggregateWithHook(ctx, queries, id.Int64(), s.catalogReadHook)
		return err
	})
	return value, err
}

func (s *Store) ListItems(ctx context.Context, filter ItemListFilter) (ItemPage, error) {
	archiveFilter, err := archiveFilterValue(filter.Archive)
	if err != nil {
		return ItemPage{}, err
	}
	if filter.PageSize.value < 1 || filter.PageSize.value > 100 {
		return ItemPage{}, domain.Invalid("page_size", domain.ViolationOutOfRange, "")
	}

	searchKey := ""
	if search, ok := filter.Search.Get(); ok {
		_, searchKey, err = domain.NormalizeDisplayAndKey(search.String())
		if err != nil {
			return ItemPage{}, err
		}
	}
	afterName, afterID := "", int64(0)
	if cursor, ok := filter.After.Get(); ok {
		if cursor.Name.Key() == "" || cursor.ID.IsZero() {
			return ItemPage{}, domain.Invalid("item_cursor", domain.ViolationInvalidFormat, "")
		}
		afterName, afterID = cursor.Name.Key(), cursor.ID.Int64()
	}

	rows, err := s.queries.ListItems(ctx, sqlcgen.ListItemsParams{
		ArchiveFilter:       archiveFilter,
		RequirePurchasable:  boolInteger(filter.RequireCapabilities.Purchasable()),
		RequireProducible:   boolInteger(filter.RequireCapabilities.Producible()),
		RequireSellable:     boolInteger(filter.RequireCapabilities.Sellable()),
		SearchKey:           searchKey,
		AfterNormalizedName: afterName,
		AfterID:             afterID,
		LimitCount:          filter.PageSize.value + 1,
	})
	if err != nil {
		return ItemPage{}, classifyError("list items", err)
	}

	hasMore := int64(len(rows)) > filter.PageSize.value
	if hasMore {
		rows = rows[:filter.PageSize.value]
	}
	items := make([]catalog.ItemSummary, 0, len(rows))
	for _, row := range rows {
		item, err := mapItemSummary(row)
		if err != nil {
			return ItemPage{}, corruptDataError("map listed item", err)
		}
		items = append(items, item)
	}
	next := domain.None[ItemCursor]()
	if hasMore && len(items) > 0 {
		last := items[len(items)-1]
		next = domain.Some(ItemCursor{Name: last.Name(), ID: last.ID()})
	}
	return ItemPage{items: items, next: next}, nil
}

func (s *Store) CreateItem(ctx context.Context, input CreateItemInput) (ItemAggregate, error) {
	var created ItemAggregate
	err := s.withWriteQueries(ctx, "create item", func(queries *sqlcgen.Queries) error {
		baseUnit, err := loadRequiredUnit(ctx, queries, input.BaseUnit)
		if err != nil {
			return err
		}
		if !baseUnit.IsItemBase() {
			return domain.Invalid("base_unit", domain.ViolationInvariant, "CAT-006")
		}
		placeholderID, _ := domain.NewItemID(1)
		if _, err := catalog.NewItem(catalog.ItemParams{
			ID: placeholderID, Name: input.Name, SKU: input.SKU,
			Description: input.Description, BaseUnit: input.BaseUnit,
			Capabilities: input.Capabilities, DefaultSalePrice: input.DefaultSalePrice,
			ReorderQuantity: input.ReorderQuantity, CreatedAt: input.CreatedAt,
			UpdatedAt: input.UpdatedAt, ArchivedAt: domain.None[domain.UTCInstant](),
			Packagings: []catalog.ItemPackaging{},
		}); err != nil {
			return err
		}

		id, err := queries.InsertItem(ctx, insertItemParams(input))
		if err != nil {
			return err
		}
		created, err = loadItemAggregate(ctx, queries, id)
		return err
	})
	return created, err
}

func (s *Store) UpdateItem(ctx context.Context, input UpdateItemInput) (ItemAggregate, error) {
	if input.ID.IsZero() {
		return ItemAggregate{}, domain.Invalid("item_id", domain.ViolationNotPositive, "")
	}
	if err := validateVersionAdvance(input.ExpectedUpdatedAt, input.UpdatedAt); err != nil {
		return ItemAggregate{}, err
	}
	var updated ItemAggregate
	err := s.withWriteQueries(ctx, "update item", func(queries *sqlcgen.Queries) error {
		current, err := loadItemAggregate(ctx, queries, input.ID.Int64())
		if err != nil {
			return err
		}
		if !current.Item().UpdatedAt().Equal(input.ExpectedUpdatedAt) {
			return fmt.Errorf("%w: item version changed", domain.ErrStale)
		}
		if current.Item().IsArchived() {
			return fmt.Errorf("%w: archived item cannot be updated", domain.ErrConflict)
		}
		baseUnit, err := loadRequiredUnit(ctx, queries, input.BaseUnit)
		if err != nil {
			return err
		}
		if !baseUnit.IsItemBase() {
			return domain.Invalid("base_unit", domain.ViolationInvariant, "CAT-006")
		}
		if err := validateActivePackagingDimensions(baseUnit, current.Packagings()); err != nil {
			return err
		}
		if _, err := catalog.NewItem(catalog.ItemParams{
			ID: input.ID, Name: input.Name, SKU: input.SKU,
			Description: input.Description, BaseUnit: input.BaseUnit,
			Capabilities: input.Capabilities, DefaultSalePrice: input.DefaultSalePrice,
			ReorderQuantity: input.ReorderQuantity, CreatedAt: current.Item().CreatedAt(),
			UpdatedAt: input.UpdatedAt, ArchivedAt: domain.None[domain.UTCInstant](),
			Packagings: current.Item().Packagings(),
		}); err != nil {
			return err
		}

		rows, err := queries.UpdateItem(ctx, updateItemParams(input))
		if err != nil {
			return err
		}
		if rows == 0 {
			return classifyItemMutationMiss(ctx, queries, input.ID, input.ExpectedUpdatedAt, false)
		}
		updated, err = loadItemAggregate(ctx, queries, input.ID.Int64())
		return err
	})
	return updated, err
}

func (s *Store) ArchiveItem(ctx context.Context, input ArchiveItemInput) (ItemAggregate, error) {
	if input.ID.IsZero() {
		return ItemAggregate{}, domain.Invalid("item_id", domain.ViolationNotPositive, "")
	}
	if err := validateVersionAdvance(input.ExpectedUpdatedAt, input.ArchivedAt); err != nil {
		return ItemAggregate{}, err
	}
	var archived ItemAggregate
	err := s.withWriteQueries(ctx, "archive item", func(queries *sqlcgen.Queries) error {
		current, err := loadItemAggregate(ctx, queries, input.ID.Int64())
		if err != nil {
			return err
		}
		if !current.Item().UpdatedAt().Equal(input.ExpectedUpdatedAt) {
			return fmt.Errorf("%w: item version changed", domain.ErrStale)
		}
		if current.Item().IsArchived() {
			return fmt.Errorf("%w: item is already archived", domain.ErrConflict)
		}
		if _, err := catalog.NewItem(catalog.ItemParams{
			ID: current.Item().ID(), Name: current.Item().Name(), SKU: current.Item().SKU(),
			Description: current.Item().Description(), BaseUnit: current.Item().BaseUnit(),
			Capabilities: current.Item().Capabilities(), DefaultSalePrice: current.Item().DefaultSalePrice(),
			ReorderQuantity: current.Item().ReorderQuantity(), CreatedAt: current.Item().CreatedAt(),
			UpdatedAt: input.ArchivedAt, ArchivedAt: domain.Some(input.ArchivedAt),
			Packagings: current.Item().Packagings(),
		}); err != nil {
			return err
		}
		rows, err := queries.ArchiveItem(ctx, sqlcgen.ArchiveItemParams{
			ArchivedAtMs: input.ArchivedAt.UnixMilli(), UpdatedAtMs: input.ArchivedAt.UnixMilli(),
			ID: input.ID.Int64(), ExpectedUpdatedAtMs: input.ExpectedUpdatedAt.UnixMilli(),
		})
		if err != nil {
			return err
		}
		if rows == 0 {
			return classifyItemMutationMiss(ctx, queries, input.ID, input.ExpectedUpdatedAt, false)
		}
		archived, err = loadItemAggregate(ctx, queries, input.ID.Int64())
		return err
	})
	return archived, err
}

func (s *Store) RestoreItem(ctx context.Context, input RestoreItemInput) (ItemAggregate, error) {
	if input.ID.IsZero() {
		return ItemAggregate{}, domain.Invalid("item_id", domain.ViolationNotPositive, "")
	}
	if err := validateVersionAdvance(input.ExpectedUpdatedAt, input.UpdatedAt); err != nil {
		return ItemAggregate{}, err
	}
	var restored ItemAggregate
	err := s.withWriteQueries(ctx, "restore item", func(queries *sqlcgen.Queries) error {
		current, err := loadItemAggregate(ctx, queries, input.ID.Int64())
		if err != nil {
			return err
		}
		if !current.Item().UpdatedAt().Equal(input.ExpectedUpdatedAt) {
			return fmt.Errorf("%w: item version changed", domain.ErrStale)
		}
		if !current.Item().IsArchived() {
			return fmt.Errorf("%w: item is already active", domain.ErrConflict)
		}
		if !current.BaseUnit().IsItemBase() {
			return corruptDataError("restore item base unit", domain.ErrInvariant)
		}
		if err := validateActivePackagingDimensions(current.BaseUnit(), current.Packagings()); err != nil {
			return err
		}
		if _, err := catalog.NewItem(catalog.ItemParams{
			ID: current.Item().ID(), Name: current.Item().Name(), SKU: current.Item().SKU(),
			Description: current.Item().Description(), BaseUnit: current.Item().BaseUnit(),
			Capabilities: current.Item().Capabilities(), DefaultSalePrice: current.Item().DefaultSalePrice(),
			ReorderQuantity: current.Item().ReorderQuantity(), CreatedAt: current.Item().CreatedAt(),
			UpdatedAt: input.UpdatedAt, ArchivedAt: domain.None[domain.UTCInstant](),
			Packagings: current.Item().Packagings(),
		}); err != nil {
			return err
		}
		rows, err := queries.RestoreItem(ctx, sqlcgen.RestoreItemParams{
			UpdatedAtMs: input.UpdatedAt.UnixMilli(), ID: input.ID.Int64(),
			ExpectedUpdatedAtMs: input.ExpectedUpdatedAt.UnixMilli(),
		})
		if err != nil {
			return err
		}
		if rows == 0 {
			return classifyItemMutationMiss(ctx, queries, input.ID, input.ExpectedUpdatedAt, true)
		}
		restored, err = loadItemAggregate(ctx, queries, input.ID.Int64())
		return err
	})
	return restored, err
}

func (s *Store) GetItemPackaging(ctx context.Context, id domain.PackagingID) (PackagingAggregate, error) {
	if id.IsZero() {
		return PackagingAggregate{}, domain.Invalid("packaging_id", domain.ViolationNotPositive, "")
	}
	var value PackagingAggregate
	err := s.withReadQueries(ctx, "get item packaging", func(queries *sqlcgen.Queries) error {
		var err error
		value, err = loadPackagingAggregateWithHook(ctx, queries, id.Int64(), s.catalogReadHook)
		return err
	})
	return value, err
}

func (s *Store) CreatePackaging(ctx context.Context, input CreatePackagingInput) (PackagingAggregate, error) {
	if input.ItemID.IsZero() {
		return PackagingAggregate{}, domain.Invalid("item_id", domain.ViolationNotPositive, "")
	}
	var created PackagingAggregate
	err := s.withWriteQueries(ctx, "create item packaging", func(queries *sqlcgen.Queries) error {
		item, err := loadItemAggregate(ctx, queries, input.ItemID.Int64())
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("load packaging item: %w", domain.ErrInvalidReference)
		}
		if err != nil {
			return err
		}
		if item.Item().IsArchived() {
			return fmt.Errorf("%w: packaging item is archived", domain.ErrInvalidReference)
		}
		enteredUnit, err := loadRequiredUnit(ctx, queries, input.EnteredUnit)
		if err != nil {
			return err
		}
		if err := catalog.ValidateCompatibleDimensions(item.BaseUnit().Dimension(), enteredUnit.Dimension()); err != nil {
			return err
		}
		placeholderID, _ := domain.NewPackagingID(1)
		if _, err := catalog.NewItemPackaging(catalog.ItemPackagingParams{
			ID: placeholderID, ItemID: input.ItemID, Name: input.Name,
			EnteredUnit: input.EnteredUnit, Conversion: input.Conversion,
			CreatedAt: input.CreatedAt, UpdatedAt: input.UpdatedAt,
			ArchivedAt: domain.None[domain.UTCInstant](),
		}); err != nil {
			return err
		}
		id, err := queries.InsertItemPackaging(ctx, sqlcgen.InsertItemPackagingParams{
			ItemID: input.ItemID.Int64(), Name: input.Name.Display(),
			NormalizedName: input.Name.Key(), EnteredUnitCode: input.EnteredUnit.String(),
			ConversionNumeratorAtomic: input.Conversion.NumeratorAtomic(),
			ConversionDenominator:     input.Conversion.Denominator(),
			CreatedAtMs:               input.CreatedAt.UnixMilli(), UpdatedAtMs: input.UpdatedAt.UnixMilli(),
		})
		if err != nil {
			return err
		}
		created, err = loadPackagingAggregate(ctx, queries, id)
		return err
	})
	return created, err
}

func (s *Store) UpdatePackaging(ctx context.Context, input UpdatePackagingInput) (PackagingAggregate, error) {
	if input.ID.IsZero() {
		return PackagingAggregate{}, domain.Invalid("packaging_id", domain.ViolationNotPositive, "")
	}
	if err := validateVersionAdvance(input.ExpectedUpdatedAt, input.UpdatedAt); err != nil {
		return PackagingAggregate{}, err
	}
	var updated PackagingAggregate
	err := s.withWriteQueries(ctx, "update item packaging", func(queries *sqlcgen.Queries) error {
		current, err := loadPackagingAggregate(ctx, queries, input.ID.Int64())
		if err != nil {
			return err
		}
		if !current.Packaging().UpdatedAt().Equal(input.ExpectedUpdatedAt) {
			return fmt.Errorf("%w: packaging version changed", domain.ErrStale)
		}
		if current.Packaging().IsArchived() {
			return fmt.Errorf("%w: archived packaging cannot be updated", domain.ErrConflict)
		}
		enteredUnit, err := loadRequiredUnit(ctx, queries, input.EnteredUnit)
		if err != nil {
			return err
		}
		if err := catalog.ValidateCompatibleDimensions(current.BaseUnit().Dimension(), enteredUnit.Dimension()); err != nil {
			return err
		}
		if _, err := catalog.NewItemPackaging(catalog.ItemPackagingParams{
			ID: input.ID, ItemID: current.Packaging().ItemID(), Name: input.Name,
			EnteredUnit: input.EnteredUnit, Conversion: input.Conversion,
			CreatedAt: current.Packaging().CreatedAt(), UpdatedAt: input.UpdatedAt,
			ArchivedAt: domain.None[domain.UTCInstant](),
		}); err != nil {
			return err
		}
		rows, err := queries.UpdateItemPackaging(ctx, sqlcgen.UpdateItemPackagingParams{
			Name: input.Name.Display(), NormalizedName: input.Name.Key(),
			EnteredUnitCode:           input.EnteredUnit.String(),
			ConversionNumeratorAtomic: input.Conversion.NumeratorAtomic(),
			ConversionDenominator:     input.Conversion.Denominator(),
			UpdatedAtMs:               input.UpdatedAt.UnixMilli(), ID: input.ID.Int64(),
			ExpectedUpdatedAtMs: input.ExpectedUpdatedAt.UnixMilli(),
		})
		if err != nil {
			return err
		}
		if rows == 0 {
			return classifyPackagingMutationMiss(ctx, queries, input.ID, input.ExpectedUpdatedAt, false)
		}
		updated, err = loadPackagingAggregate(ctx, queries, input.ID.Int64())
		return err
	})
	return updated, err
}

func (s *Store) ArchivePackaging(ctx context.Context, input ArchivePackagingInput) (PackagingAggregate, error) {
	if input.ID.IsZero() {
		return PackagingAggregate{}, domain.Invalid("packaging_id", domain.ViolationNotPositive, "")
	}
	if err := validateVersionAdvance(input.ExpectedUpdatedAt, input.ArchivedAt); err != nil {
		return PackagingAggregate{}, err
	}
	var archived PackagingAggregate
	err := s.withWriteQueries(ctx, "archive item packaging", func(queries *sqlcgen.Queries) error {
		current, err := loadPackagingAggregate(ctx, queries, input.ID.Int64())
		if err != nil {
			return err
		}
		if !current.Packaging().UpdatedAt().Equal(input.ExpectedUpdatedAt) {
			return fmt.Errorf("%w: packaging version changed", domain.ErrStale)
		}
		if current.Packaging().IsArchived() {
			return fmt.Errorf("%w: packaging is already archived", domain.ErrConflict)
		}
		if _, err := catalog.NewItemPackaging(catalog.ItemPackagingParams{
			ID: current.Packaging().ID(), ItemID: current.Packaging().ItemID(),
			Name: current.Packaging().Name(), EnteredUnit: current.Packaging().EnteredUnit(),
			Conversion: current.Packaging().Conversion(), CreatedAt: current.Packaging().CreatedAt(),
			UpdatedAt: input.ArchivedAt, ArchivedAt: domain.Some(input.ArchivedAt),
		}); err != nil {
			return err
		}
		rows, err := queries.ArchiveItemPackaging(ctx, sqlcgen.ArchiveItemPackagingParams{
			ArchivedAtMs: input.ArchivedAt.UnixMilli(), UpdatedAtMs: input.ArchivedAt.UnixMilli(),
			ID: input.ID.Int64(), ExpectedUpdatedAtMs: input.ExpectedUpdatedAt.UnixMilli(),
		})
		if err != nil {
			return err
		}
		if rows == 0 {
			return classifyPackagingMutationMiss(ctx, queries, input.ID, input.ExpectedUpdatedAt, false)
		}
		archived, err = loadPackagingAggregate(ctx, queries, input.ID.Int64())
		return err
	})
	return archived, err
}

// ReconfigureArchivedPackaging replaces an archived definition without making
// it selectable. Advancing archived_at together with updated_at preserves the
// schema's timestamp ordering while keeping restoration as an explicit action.
func (s *Store) ReconfigureArchivedPackaging(ctx context.Context, input ReconfigureArchivedPackagingInput) (PackagingAggregate, error) {
	if input.ID.IsZero() {
		return PackagingAggregate{}, domain.Invalid("packaging_id", domain.ViolationNotPositive, "")
	}
	if err := validateVersionAdvance(input.ExpectedUpdatedAt, input.UpdatedAt); err != nil {
		return PackagingAggregate{}, err
	}
	var reconfigured PackagingAggregate
	err := s.withWriteQueries(ctx, "reconfigure archived item packaging", func(queries *sqlcgen.Queries) error {
		current, err := loadPackagingAggregate(ctx, queries, input.ID.Int64())
		if err != nil {
			return err
		}
		if !current.Packaging().UpdatedAt().Equal(input.ExpectedUpdatedAt) {
			return fmt.Errorf("%w: packaging version changed", domain.ErrStale)
		}
		if !current.Packaging().IsArchived() {
			return fmt.Errorf("%w: active packaging cannot be reconfigured as archived", domain.ErrConflict)
		}

		item, err := loadItemAggregate(ctx, queries, current.Packaging().ItemID().Int64())
		if err != nil {
			return err
		}
		if item.Item().IsArchived() {
			return fmt.Errorf("%w: packaging item is archived", domain.ErrInvalidReference)
		}
		enteredUnit, err := loadRequiredUnit(ctx, queries, input.EnteredUnit)
		if err != nil {
			return err
		}
		if err := catalog.ValidateCompatibleDimensions(item.BaseUnit().Dimension(), enteredUnit.Dimension()); err != nil {
			return err
		}
		if _, err := catalog.NewItemPackaging(catalog.ItemPackagingParams{
			ID: input.ID, ItemID: current.Packaging().ItemID(), Name: input.Name,
			EnteredUnit: input.EnteredUnit, Conversion: input.Conversion,
			CreatedAt: current.Packaging().CreatedAt(), UpdatedAt: input.UpdatedAt,
			ArchivedAt: domain.Some(input.UpdatedAt),
		}); err != nil {
			return err
		}

		rows, err := queries.ReconfigureArchivedItemPackaging(ctx, sqlcgen.ReconfigureArchivedItemPackagingParams{
			Name: input.Name.Display(), NormalizedName: input.Name.Key(),
			EnteredUnitCode:           input.EnteredUnit.String(),
			ConversionNumeratorAtomic: input.Conversion.NumeratorAtomic(),
			ConversionDenominator:     input.Conversion.Denominator(),
			UpdatedAtMs:               input.UpdatedAt.UnixMilli(), ID: input.ID.Int64(),
			ExpectedUpdatedAtMs: input.ExpectedUpdatedAt.UnixMilli(),
		})
		if err != nil {
			return err
		}
		if rows == 0 {
			return classifyPackagingMutationMiss(ctx, queries, input.ID, input.ExpectedUpdatedAt, true)
		}
		reconfigured, err = loadPackagingAggregate(ctx, queries, input.ID.Int64())
		return err
	})
	return reconfigured, err
}

func (s *Store) RestorePackaging(ctx context.Context, input RestorePackagingInput) (PackagingAggregate, error) {
	if input.ID.IsZero() {
		return PackagingAggregate{}, domain.Invalid("packaging_id", domain.ViolationNotPositive, "")
	}
	if err := validateVersionAdvance(input.ExpectedUpdatedAt, input.UpdatedAt); err != nil {
		return PackagingAggregate{}, err
	}
	var restored PackagingAggregate
	err := s.withWriteQueries(ctx, "restore item packaging", func(queries *sqlcgen.Queries) error {
		current, err := loadPackagingAggregate(ctx, queries, input.ID.Int64())
		if err != nil {
			return err
		}
		if !current.Packaging().UpdatedAt().Equal(input.ExpectedUpdatedAt) {
			return fmt.Errorf("%w: packaging version changed", domain.ErrStale)
		}
		if !current.Packaging().IsArchived() {
			return fmt.Errorf("%w: packaging is already active", domain.ErrConflict)
		}
		item, err := loadItemAggregate(ctx, queries, current.Packaging().ItemID().Int64())
		if err != nil {
			return err
		}
		if item.Item().IsArchived() {
			return fmt.Errorf("%w: packaging item is archived", domain.ErrInvalidReference)
		}
		if err := catalog.ValidateCompatibleDimensions(current.BaseUnit().Dimension(), current.EnteredUnit().Dimension()); err != nil {
			return err
		}
		if _, err := catalog.NewItemPackaging(catalog.ItemPackagingParams{
			ID: current.Packaging().ID(), ItemID: current.Packaging().ItemID(),
			Name: current.Packaging().Name(), EnteredUnit: current.Packaging().EnteredUnit(),
			Conversion: current.Packaging().Conversion(), CreatedAt: current.Packaging().CreatedAt(),
			UpdatedAt: input.UpdatedAt, ArchivedAt: domain.None[domain.UTCInstant](),
		}); err != nil {
			return err
		}
		rows, err := queries.RestoreItemPackaging(ctx, sqlcgen.RestoreItemPackagingParams{
			UpdatedAtMs: input.UpdatedAt.UnixMilli(), ID: input.ID.Int64(),
			ExpectedUpdatedAtMs: input.ExpectedUpdatedAt.UnixMilli(),
		})
		if err != nil {
			return err
		}
		if rows == 0 {
			return classifyPackagingMutationMiss(ctx, queries, input.ID, input.ExpectedUpdatedAt, true)
		}
		restored, err = loadPackagingAggregate(ctx, queries, input.ID.Int64())
		return err
	})
	return restored, err
}

type catalogReadStage uint8

const (
	catalogItemRowLoaded catalogReadStage = iota + 1
	catalogPackagingRowLoaded
)

func loadItemAggregate(ctx context.Context, queries *sqlcgen.Queries, id int64) (ItemAggregate, error) {
	return loadItemAggregateWithHook(ctx, queries, id, nil)
}

func loadItemAggregateWithHook(
	ctx context.Context,
	queries *sqlcgen.Queries,
	id int64,
	hook func(catalogReadStage) error,
) (ItemAggregate, error) {
	row, err := queries.GetItem(ctx, id)
	if err != nil {
		return ItemAggregate{}, err
	}
	if hook != nil {
		if err := hook(catalogItemRowLoaded); err != nil {
			return ItemAggregate{}, err
		}
	}
	return loadItemAggregateFromRow(ctx, queries, row)
}

func loadItemAggregateFromRow(ctx context.Context, queries *sqlcgen.Queries, row sqlcgen.Item) (ItemAggregate, error) {
	baseUnitCode, err := domain.NewUnitCode(row.BaseUnitCode)
	if err != nil {
		return ItemAggregate{}, err
	}
	baseUnit, err := loadUnitForPersistedAggregate(ctx, queries, baseUnitCode)
	if err != nil {
		return ItemAggregate{}, err
	}
	packagingRows, err := queries.ListItemPackagings(ctx, sqlcgen.ListItemPackagingsParams{
		ItemID: row.ID, IncludeArchived: 1,
	})
	if err != nil {
		return ItemAggregate{}, err
	}
	packagings := make([]catalog.ItemPackaging, 0, len(packagingRows))
	packagingAggregates := make([]PackagingAggregate, 0, len(packagingRows))
	unitCache := map[string]catalog.MeasurementUnit{baseUnit.Code().String(): baseUnit}
	for _, packagingRow := range packagingRows {
		packaging, err := mapItemPackaging(packagingRow)
		if err != nil {
			return ItemAggregate{}, err
		}
		enteredUnit, found := unitCache[packaging.EnteredUnit().String()]
		if !found {
			enteredUnit, err = loadUnitForPersistedAggregate(ctx, queries, packaging.EnteredUnit())
			if err != nil {
				return ItemAggregate{}, err
			}
			unitCache[packaging.EnteredUnit().String()] = enteredUnit
		}
		if !packaging.IsArchived() {
			if err := catalog.ValidateCompatibleDimensions(baseUnit.Dimension(), enteredUnit.Dimension()); err != nil {
				return ItemAggregate{}, domain.Corrupt(err)
			}
		}
		packagings = append(packagings, packaging)
		packagingAggregates = append(packagingAggregates, PackagingAggregate{
			packaging: packaging, baseUnit: baseUnit, enteredUnit: enteredUnit,
		})
	}
	item, err := mapItem(row, packagings)
	if err != nil {
		return ItemAggregate{}, domain.Corrupt(err)
	}
	return ItemAggregate{item: item, baseUnit: baseUnit, packagings: packagingAggregates}, nil
}

func loadPackagingAggregate(ctx context.Context, queries *sqlcgen.Queries, id int64) (PackagingAggregate, error) {
	return loadPackagingAggregateWithHook(ctx, queries, id, nil)
}

func loadPackagingAggregateWithHook(
	ctx context.Context,
	queries *sqlcgen.Queries,
	id int64,
	hook func(catalogReadStage) error,
) (PackagingAggregate, error) {
	row, err := queries.GetItemPackaging(ctx, id)
	if err != nil {
		return PackagingAggregate{}, err
	}
	if hook != nil {
		if err := hook(catalogPackagingRowLoaded); err != nil {
			return PackagingAggregate{}, err
		}
	}
	packaging, err := mapItemPackaging(row)
	if err != nil {
		return PackagingAggregate{}, domain.Corrupt(err)
	}
	itemRow, err := queries.GetItem(ctx, packaging.ItemID().Int64())
	if errors.Is(err, sql.ErrNoRows) {
		return PackagingAggregate{}, domain.Corrupt(fmt.Errorf("packaging item is missing: %w", domain.ErrInvariant))
	}
	if err != nil {
		return PackagingAggregate{}, err
	}
	baseCode, err := domain.NewUnitCode(itemRow.BaseUnitCode)
	if err != nil {
		return PackagingAggregate{}, err
	}
	baseUnit, err := loadUnitForPersistedAggregate(ctx, queries, baseCode)
	if err != nil {
		return PackagingAggregate{}, err
	}
	enteredUnit, err := loadUnitForPersistedAggregate(ctx, queries, packaging.EnteredUnit())
	if err != nil {
		return PackagingAggregate{}, err
	}
	if !packaging.IsArchived() {
		if err := catalog.ValidateCompatibleDimensions(baseUnit.Dimension(), enteredUnit.Dimension()); err != nil {
			return PackagingAggregate{}, domain.Corrupt(err)
		}
	}
	return PackagingAggregate{packaging: packaging, baseUnit: baseUnit, enteredUnit: enteredUnit}, nil
}

func mapItem(row sqlcgen.Item, packagings []catalog.ItemPackaging) (catalog.Item, error) {
	id, err := domain.NewItemID(row.ID)
	if err != nil {
		return catalog.Item{}, domain.Corrupt(err)
	}
	name, err := domain.RestoreUniqueName(row.Name, row.NormalizedName)
	if err != nil || name.Display() != row.Name {
		if err == nil {
			err = domain.Corrupt(domain.Invalid("name", domain.ViolationInvariant, "CAT-005"))
		}
		return catalog.Item{}, err
	}
	sku, err := restoreOptionalSKU(row.SKU, row.NormalizedSKU)
	if err != nil {
		return catalog.Item{}, err
	}
	description, err := restoreOptionalText(row.Description)
	if err != nil {
		return catalog.Item{}, err
	}
	baseUnit, err := domain.NewUnitCode(row.BaseUnitCode)
	if err != nil || baseUnit.String() != row.BaseUnitCode {
		if err == nil {
			err = domain.Invalid("base_unit_code", domain.ViolationInvariant, "CAT-006")
		}
		return catalog.Item{}, domain.Corrupt(err)
	}
	purchasable, err := restoreBoolean("is_purchasable", row.IsPurchasable)
	if err != nil {
		return catalog.Item{}, domain.Corrupt(err)
	}
	producible, err := restoreBoolean("is_producible", row.IsProducible)
	if err != nil {
		return catalog.Item{}, domain.Corrupt(err)
	}
	sellable, err := restoreBoolean("is_sellable", row.IsSellable)
	if err != nil {
		return catalog.Item{}, domain.Corrupt(err)
	}
	defaultPrice, err := restoreOptionalMinorAmount(row.DefaultSalePriceMinor)
	if err != nil {
		return catalog.Item{}, domain.Corrupt(err)
	}
	reorderQuantity, err := restoreOptionalQuantity(row.ReorderQuantityAtomic)
	if err != nil {
		return catalog.Item{}, domain.Corrupt(err)
	}
	createdAt, err := domain.UTCInstantFromUnixMilli(row.CreatedAtMs)
	if err != nil {
		return catalog.Item{}, domain.Corrupt(err)
	}
	updatedAt, err := domain.UTCInstantFromUnixMilli(row.UpdatedAtMs)
	if err != nil {
		return catalog.Item{}, domain.Corrupt(err)
	}
	archivedAt, err := restoreOptionalInstant(row.ArchivedAtMs)
	if err != nil {
		return catalog.Item{}, domain.Corrupt(err)
	}
	item, err := catalog.NewItem(catalog.ItemParams{
		ID: id, Name: name, SKU: sku, Description: description, BaseUnit: baseUnit,
		Capabilities:     catalog.NewCapabilities(purchasable, producible, sellable),
		DefaultSalePrice: defaultPrice, ReorderQuantity: reorderQuantity,
		CreatedAt: createdAt, UpdatedAt: updatedAt, ArchivedAt: archivedAt,
		Packagings: packagings,
	})
	if err != nil {
		return catalog.Item{}, domain.Corrupt(err)
	}
	return item, nil
}

func mapItemSummary(row sqlcgen.Item) (catalog.ItemSummary, error) {
	item, err := mapItem(row, []catalog.ItemPackaging{})
	if err != nil {
		return catalog.ItemSummary{}, err
	}
	summary, err := catalog.NewItemSummary(catalog.ItemSummaryParams{
		ID: item.ID(), Name: item.Name(), SKU: item.SKU(), Description: item.Description(),
		BaseUnit: item.BaseUnit(), Capabilities: item.Capabilities(),
		DefaultSalePrice: item.DefaultSalePrice(), ReorderQuantity: item.ReorderQuantity(),
		CreatedAt: item.CreatedAt(), UpdatedAt: item.UpdatedAt(), ArchivedAt: item.ArchivedAt(),
	})
	if err != nil {
		return catalog.ItemSummary{}, domain.Corrupt(err)
	}
	return summary, nil
}

func mapItemPackaging(row sqlcgen.ItemPackaging) (catalog.ItemPackaging, error) {
	id, err := domain.NewPackagingID(row.ID)
	if err != nil {
		return catalog.ItemPackaging{}, domain.Corrupt(err)
	}
	itemID, err := domain.NewItemID(row.ItemID)
	if err != nil {
		return catalog.ItemPackaging{}, domain.Corrupt(err)
	}
	name, err := domain.RestoreUniqueName(row.Name, row.NormalizedName)
	if err != nil || name.Display() != row.Name {
		if err == nil {
			err = domain.Corrupt(domain.Invalid("name", domain.ViolationInvariant, "CAT-005"))
		}
		return catalog.ItemPackaging{}, err
	}
	enteredUnit, err := domain.NewUnitCode(row.EnteredUnitCode)
	if err != nil || enteredUnit.String() != row.EnteredUnitCode {
		if err == nil {
			err = domain.Invalid("entered_unit_code", domain.ViolationInvariant, "CAT-006")
		}
		return catalog.ItemPackaging{}, domain.Corrupt(err)
	}
	conversion, err := domain.NewUnitConversion(row.ConversionNumeratorAtomic, row.ConversionDenominator)
	if err != nil {
		return catalog.ItemPackaging{}, domain.Corrupt(err)
	}
	if conversion.NumeratorAtomic() != row.ConversionNumeratorAtomic || conversion.Denominator() != row.ConversionDenominator {
		return catalog.ItemPackaging{}, domain.Corrupt(domain.Invalid("packaging_conversion", domain.ViolationInvariant, "UNIT-003"))
	}
	createdAt, err := domain.UTCInstantFromUnixMilli(row.CreatedAtMs)
	if err != nil {
		return catalog.ItemPackaging{}, domain.Corrupt(err)
	}
	updatedAt, err := domain.UTCInstantFromUnixMilli(row.UpdatedAtMs)
	if err != nil {
		return catalog.ItemPackaging{}, domain.Corrupt(err)
	}
	archivedAt, err := restoreOptionalInstant(row.ArchivedAtMs)
	if err != nil {
		return catalog.ItemPackaging{}, domain.Corrupt(err)
	}
	packaging, err := catalog.NewItemPackaging(catalog.ItemPackagingParams{
		ID: id, ItemID: itemID, Name: name, EnteredUnit: enteredUnit,
		Conversion: conversion, CreatedAt: createdAt, UpdatedAt: updatedAt,
		ArchivedAt: archivedAt,
	})
	if err != nil {
		return catalog.ItemPackaging{}, domain.Corrupt(err)
	}
	return packaging, nil
}

func loadRequiredUnit(ctx context.Context, queries *sqlcgen.Queries, code domain.UnitCode) (catalog.MeasurementUnit, error) {
	if code.String() == "" {
		return catalog.MeasurementUnit{}, domain.Invalid("unit_code", domain.ViolationRequired, "")
	}
	row, err := queries.GetMeasurementUnit(ctx, code.String())
	if errors.Is(err, sql.ErrNoRows) {
		return catalog.MeasurementUnit{}, fmt.Errorf("load referenced unit %q: %w", code.String(), domain.ErrInvalidReference)
	}
	if err != nil {
		return catalog.MeasurementUnit{}, err
	}
	unit, err := mapMeasurementUnit(row)
	if err != nil {
		return catalog.MeasurementUnit{}, corruptDataError("map referenced unit", err)
	}
	return unit, nil
}

func loadUnitForPersistedAggregate(ctx context.Context, queries *sqlcgen.Queries, code domain.UnitCode) (catalog.MeasurementUnit, error) {
	row, err := queries.GetMeasurementUnit(ctx, code.String())
	if errors.Is(err, sql.ErrNoRows) {
		return catalog.MeasurementUnit{}, domain.Corrupt(fmt.Errorf("persisted unit %q is missing: %w", code.String(), domain.ErrInvariant))
	}
	if err != nil {
		return catalog.MeasurementUnit{}, err
	}
	unit, err := mapMeasurementUnit(row)
	if err != nil {
		return catalog.MeasurementUnit{}, domain.Corrupt(err)
	}
	return unit, nil
}

func validateActivePackagingDimensions(base catalog.MeasurementUnit, packagings []PackagingAggregate) error {
	for _, packaging := range packagings {
		if packaging.Packaging().IsArchived() {
			continue
		}
		if err := catalog.ValidateCompatibleDimensions(base.Dimension(), packaging.EnteredUnit().Dimension()); err != nil {
			return err
		}
	}
	return nil
}

func classifyItemMutationMiss(ctx context.Context, queries *sqlcgen.Queries, id domain.ItemID, expected domain.UTCInstant, expectedArchived bool) error {
	row, err := queries.GetItem(ctx, id.Int64())
	if err != nil {
		return classifyError("reload item after missed update", err)
	}
	if row.UpdatedAtMs != expected.UnixMilli() {
		return fmt.Errorf("%w: item version changed", domain.ErrStale)
	}
	if row.ArchivedAtMs.Valid != expectedArchived {
		return fmt.Errorf("%w: item archive state changed", domain.ErrConflict)
	}
	return fmt.Errorf("%w: item update matched no row", domain.ErrConflict)
}

func classifyPackagingMutationMiss(ctx context.Context, queries *sqlcgen.Queries, id domain.PackagingID, expected domain.UTCInstant, expectedArchived bool) error {
	row, err := queries.GetItemPackaging(ctx, id.Int64())
	if err != nil {
		return classifyError("reload packaging after missed update", err)
	}
	if row.UpdatedAtMs != expected.UnixMilli() {
		return fmt.Errorf("%w: packaging version changed", domain.ErrStale)
	}
	if row.ArchivedAtMs.Valid != expectedArchived {
		return fmt.Errorf("%w: packaging archive state changed", domain.ErrConflict)
	}
	return fmt.Errorf("%w: packaging update matched no row", domain.ErrConflict)
}

func insertItemParams(input CreateItemInput) sqlcgen.InsertItemParams {
	return sqlcgen.InsertItemParams{
		Name: input.Name.Display(), NormalizedName: input.Name.Key(),
		SKU: nullableSKU(input.SKU), NormalizedSKU: nullableNormalizedSKU(input.SKU),
		Description: nullableText(input.Description), BaseUnitCode: input.BaseUnit.String(),
		IsPurchasable:         boolInteger(input.Capabilities.Purchasable()),
		IsProducible:          boolInteger(input.Capabilities.Producible()),
		IsSellable:            boolInteger(input.Capabilities.Sellable()),
		DefaultSalePriceMinor: nullableMinorAmount(input.DefaultSalePrice),
		ReorderQuantityAtomic: nullableQuantity(input.ReorderQuantity),
		CreatedAtMs:           input.CreatedAt.UnixMilli(), UpdatedAtMs: input.UpdatedAt.UnixMilli(),
	}
}

func updateItemParams(input UpdateItemInput) sqlcgen.UpdateItemParams {
	return sqlcgen.UpdateItemParams{
		Name: input.Name.Display(), NormalizedName: input.Name.Key(),
		SKU: nullableSKU(input.SKU), NormalizedSKU: nullableNormalizedSKU(input.SKU),
		Description: nullableText(input.Description), BaseUnitCode: input.BaseUnit.String(),
		IsPurchasable:         boolInteger(input.Capabilities.Purchasable()),
		IsProducible:          boolInteger(input.Capabilities.Producible()),
		IsSellable:            boolInteger(input.Capabilities.Sellable()),
		DefaultSalePriceMinor: nullableMinorAmount(input.DefaultSalePrice),
		ReorderQuantityAtomic: nullableQuantity(input.ReorderQuantity),
		UpdatedAtMs:           input.UpdatedAt.UnixMilli(), ID: input.ID.Int64(),
		ExpectedUpdatedAtMs: input.ExpectedUpdatedAt.UnixMilli(),
	}
}

func archiveFilterValue(filter domain.ArchiveFilter) (int64, error) {
	switch filter {
	case "", domain.ArchiveActive:
		return 0, nil
	case domain.ArchiveArchived:
		return 1, nil
	case domain.ArchiveAll:
		return 2, nil
	default:
		return 0, domain.Invalid("archive_filter", domain.ViolationInvalidEnum, "")
	}
}

func boolInteger(value bool) int64 {
	if value {
		return 1
	}
	return 0
}

func nullableSKU(value domain.Option[domain.SKU]) sql.NullString {
	sku, ok := value.Get()
	return sql.NullString{String: sku.Display(), Valid: ok}
}

func nullableNormalizedSKU(value domain.Option[domain.SKU]) sql.NullString {
	sku, ok := value.Get()
	return sql.NullString{String: sku.Key(), Valid: ok}
}

func nullableText(value domain.Option[domain.NonEmptyText]) sql.NullString {
	text, ok := value.Get()
	return sql.NullString{String: text.String(), Valid: ok}
}

func nullableQuantity(value domain.Option[domain.AtomicQuantity]) sql.NullInt64 {
	quantity, ok := value.Get()
	return sql.NullInt64{Int64: quantity.Int64(), Valid: ok}
}

func restoreOptionalSKU(display, key sql.NullString) (domain.Option[domain.SKU], error) {
	if display.Valid != key.Valid {
		return domain.None[domain.SKU](), domain.Corrupt(domain.ErrInvariant)
	}
	if !display.Valid {
		return domain.None[domain.SKU](), nil
	}
	sku, err := domain.RestoreSKU(display.String, key.String)
	if err != nil || sku.Display() != display.String {
		if err == nil {
			err = domain.Corrupt(domain.Invalid("sku", domain.ViolationInvariant, "CAT-008"))
		}
		return domain.None[domain.SKU](), err
	}
	return domain.Some(sku), nil
}

func restoreOptionalText(value sql.NullString) (domain.Option[domain.NonEmptyText], error) {
	if !value.Valid {
		return domain.None[domain.NonEmptyText](), nil
	}
	text, err := domain.NewNonEmptyText(value.String)
	if err != nil {
		return domain.None[domain.NonEmptyText](), domain.Corrupt(err)
	}
	if text.String() != value.String {
		return domain.None[domain.NonEmptyText](), domain.Corrupt(domain.ErrInvariant)
	}
	return domain.Some(text), nil
}

func restoreOptionalQuantity(value sql.NullInt64) (domain.Option[domain.AtomicQuantity], error) {
	if !value.Valid {
		return domain.None[domain.AtomicQuantity](), nil
	}
	quantity, err := domain.NewAtomicQuantity(value.Int64)
	if err != nil {
		return domain.None[domain.AtomicQuantity](), err
	}
	return domain.Some(quantity), nil
}

func restoreOptionalInstant(value sql.NullInt64) (domain.Option[domain.UTCInstant], error) {
	if !value.Valid {
		return domain.None[domain.UTCInstant](), nil
	}
	instant, err := domain.UTCInstantFromUnixMilli(value.Int64)
	if err != nil {
		return domain.None[domain.UTCInstant](), err
	}
	return domain.Some(instant), nil
}
