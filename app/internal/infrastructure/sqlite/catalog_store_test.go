package sqlite

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/jerobas/saas/database"
	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/domain/catalog"
)

func TestCatalogStoreCreatesCompleteItemAndPackagingAggregates(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "catalog-create.db"), database.DefaultOpenOptions())
	ctx := context.Background()
	createdAt := mustCatalogInstant(t, 1)

	created := createCatalogItem(t, store, CreateItemInput{
		Name:         mustCatalogName(t, "Flour"),
		BaseUnit:     mustCatalogUnitCode(t, "g"),
		Capabilities: catalog.NewCapabilities(true, false, false),
		CreatedAt:    createdAt,
		UpdatedAt:    createdAt,
	})
	if created.Item().ID().IsZero() || created.Item().Name().Display() != "Flour" {
		t.Fatalf("unexpected created item: id=%d name=%q", created.Item().ID().Int64(), created.Item().Name().Display())
	}
	if created.BaseUnit().Code().String() != "g" || created.BaseUnit().Dimension() != domain.DimensionMass {
		t.Fatalf("unexpected base unit: %s/%s", created.BaseUnit().Code().String(), created.BaseUnit().Dimension())
	}
	if packagings := created.Packagings(); packagings == nil || len(packagings) != 0 {
		t.Fatalf("new item packagings = %#v, want non-nil empty slice", packagings)
	}

	var quantity, inventoryValue, balanceUpdatedAt int64
	err := store.database.QueryRowContext(ctx, `
		SELECT quantity_atomic, inventory_value_micro, updated_at_ms
		FROM inventory_balances WHERE item_id = ?`, created.Item().ID().Int64()).
		Scan(&quantity, &inventoryValue, &balanceUpdatedAt)
	if err != nil {
		t.Fatal(err)
	}
	if quantity != 0 || inventoryValue != 0 || balanceUpdatedAt != createdAt.UnixMilli() {
		t.Fatalf("initial balance = %d/%d/%d, want 0/0/%d", quantity, inventoryValue, balanceUpdatedAt, createdAt.UnixMilli())
	}

	conversion := mustCatalogConversion(t, 2_000_000, 2)
	packaging, err := store.CreatePackaging(ctx, CreatePackagingInput{
		ItemID:      created.Item().ID(),
		Name:        mustCatalogName(t, "Kilogram bag"),
		EnteredUnit: mustCatalogUnitCode(t, "kg"),
		Conversion:  conversion,
		CreatedAt:   mustCatalogInstant(t, 2),
		UpdatedAt:   mustCatalogInstant(t, 2),
	})
	if err != nil {
		t.Fatal(err)
	}
	if packaging.Packaging().Conversion().NumeratorAtomic() != 1_000_000 || packaging.Packaging().Conversion().Denominator() != 1 {
		t.Fatalf("packaging conversion = %d/%d, want reduced 1000000/1",
			packaging.Packaging().Conversion().NumeratorAtomic(), packaging.Packaging().Conversion().Denominator())
	}
	if packaging.BaseUnit().Dimension() != domain.DimensionMass || packaging.EnteredUnit().Dimension() != domain.DimensionMass {
		t.Fatalf("packaging dimensions = %s/%s, want MASS/MASS", packaging.BaseUnit().Dimension(), packaging.EnteredUnit().Dimension())
	}

	reloaded, err := store.GetItem(ctx, created.Item().ID())
	if err != nil {
		t.Fatal(err)
	}
	if len(reloaded.Packagings()) != 1 || len(reloaded.Item().Packagings()) != 1 {
		t.Fatalf("reloaded packaging counts = %d/%d, want 1/1", len(reloaded.Packagings()), len(reloaded.Item().Packagings()))
	}
	if reloaded.Packagings()[0].EnteredUnit().Code().String() != "kg" {
		t.Fatalf("entered unit = %q, want kg", reloaded.Packagings()[0].EnteredUnit().Code().String())
	}

	var numerator, denominator int64
	if err := store.database.QueryRowContext(ctx, `
		SELECT conversion_numerator_atomic, conversion_denominator
		FROM item_packagings WHERE id = ?`, packaging.Packaging().ID().Int64()).Scan(&numerator, &denominator); err != nil {
		t.Fatal(err)
	}
	if numerator != 1_000_000 || denominator != 1 {
		t.Fatalf("persisted conversion = %d/%d, want 1000000/1", numerator, denominator)
	}
}

func TestCatalogStoreAggregateReadsStayOnSnapshotAcrossExternalCommit(t *testing.T) {
	t.Run("item aggregate", func(t *testing.T) {
		reader, writer, item, packaging := newCatalogSnapshotFixture(t)
		ctx := context.Background()
		hookCalled := false
		reader.catalogReadHook = func(stage catalogReadStage) error {
			if stage != catalogItemRowLoaded {
				return errors.New("unexpected catalog read stage")
			}
			hookCalled = true
			reader.catalogReadHook = nil
			return transitionCatalogAggregateToVolume(ctx, writer, item.Item().ID(), packaging.Packaging().ID())
		}

		snapshot, err := reader.GetItem(ctx, item.Item().ID())
		if err != nil {
			t.Fatalf("GetItem snapshot: %v", err)
		}
		if !hookCalled {
			t.Fatal("external writer hook was not called")
		}
		if got := snapshot.BaseUnit().Code().String(); got != "g" {
			t.Fatalf("snapshot base unit = %q, want g", got)
		}
		if got := snapshot.Packagings()[0].EnteredUnit().Code().String(); got != "kg" {
			t.Fatalf("snapshot entered unit = %q, want kg", got)
		}

		latest, err := reader.GetItem(ctx, item.Item().ID())
		if err != nil {
			t.Fatalf("GetItem latest: %v", err)
		}
		if got := latest.BaseUnit().Code().String(); got != "ml" {
			t.Fatalf("latest base unit = %q, want ml", got)
		}
		if got := latest.Packagings()[0].EnteredUnit().Code().String(); got != "l" {
			t.Fatalf("latest entered unit = %q, want l", got)
		}
	})

	t.Run("packaging aggregate", func(t *testing.T) {
		reader, writer, item, packaging := newCatalogSnapshotFixture(t)
		ctx := context.Background()
		hookCalled := false
		reader.catalogReadHook = func(stage catalogReadStage) error {
			if stage != catalogPackagingRowLoaded {
				return errors.New("unexpected catalog read stage")
			}
			hookCalled = true
			reader.catalogReadHook = nil
			return transitionCatalogAggregateToVolume(ctx, writer, item.Item().ID(), packaging.Packaging().ID())
		}

		snapshot, err := reader.GetItemPackaging(ctx, packaging.Packaging().ID())
		if err != nil {
			t.Fatalf("GetItemPackaging snapshot: %v", err)
		}
		if !hookCalled {
			t.Fatal("external writer hook was not called")
		}
		if got := snapshot.BaseUnit().Code().String(); got != "g" {
			t.Fatalf("snapshot base unit = %q, want g", got)
		}
		if got := snapshot.EnteredUnit().Code().String(); got != "kg" {
			t.Fatalf("snapshot entered unit = %q, want kg", got)
		}

		latest, err := reader.GetItemPackaging(ctx, packaging.Packaging().ID())
		if err != nil {
			t.Fatalf("GetItemPackaging latest: %v", err)
		}
		if got := latest.BaseUnit().Code().String(); got != "ml" {
			t.Fatalf("latest base unit = %q, want ml", got)
		}
		if got := latest.EnteredUnit().Code().String(); got != "l" {
			t.Fatalf("latest entered unit = %q, want l", got)
		}
	})
}

func TestCatalogStoreNormalizesUnicodeKeysAndRollsBackConflicts(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "catalog-normalization.db"), database.DefaultOpenOptions())
	firstAt := mustCatalogInstant(t, 1)
	createCatalogItem(t, store, CreateItemInput{
		Name:         mustCatalogName(t, "Café"),
		BaseUnit:     mustCatalogUnitCode(t, "g"),
		Capabilities: catalog.NewCapabilities(true, false, false),
		CreatedAt:    firstAt,
		UpdatedAt:    firstAt,
	})

	secondAt := mustCatalogInstant(t, 2)
	_, err := store.CreateItem(context.Background(), CreateItemInput{
		Name:         mustCatalogName(t, "  CAFE\u0301  "),
		BaseUnit:     mustCatalogUnitCode(t, "g"),
		Capabilities: catalog.NewCapabilities(true, false, false),
		CreatedAt:    secondAt,
		UpdatedAt:    secondAt,
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("equivalent Unicode name error = %v, want domain.ErrConflict", err)
	}

	var items, balances int
	if err := store.database.QueryRowContext(context.Background(), "SELECT count(*) FROM items").Scan(&items); err != nil {
		t.Fatal(err)
	}
	if err := store.database.QueryRowContext(context.Background(), "SELECT count(*) FROM inventory_balances").Scan(&balances); err != nil {
		t.Fatal(err)
	}
	if items != 1 || balances != 1 {
		t.Fatalf("row counts after conflict = items:%d balances:%d, want 1/1", items, balances)
	}
}

func TestCatalogStoreRejectsBadReferencesAndIncompatibleDimensionsBeforeWrites(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "catalog-validation.db"), database.DefaultOpenOptions())
	ctx := context.Background()
	at := mustCatalogInstant(t, 1)

	_, err := store.CreateItem(ctx, CreateItemInput{
		Name:         mustCatalogName(t, "Unknown unit item"),
		BaseUnit:     mustCatalogUnitCode(t, "not_a_unit"),
		Capabilities: catalog.NewCapabilities(true, false, false),
		CreatedAt:    at,
		UpdatedAt:    at,
	})
	if !errors.Is(err, domain.ErrInvalidReference) {
		t.Fatalf("unknown base unit error = %v, want domain.ErrInvalidReference", err)
	}

	item := createCatalogItem(t, store, CreateItemInput{
		Name:         mustCatalogName(t, "Sugar"),
		BaseUnit:     mustCatalogUnitCode(t, "g"),
		Capabilities: catalog.NewCapabilities(true, false, false),
		CreatedAt:    at,
		UpdatedAt:    at,
	})
	_, err = store.CreatePackaging(ctx, CreatePackagingInput{
		ItemID:      item.Item().ID(),
		Name:        mustCatalogName(t, "Litre"),
		EnteredUnit: mustCatalogUnitCode(t, "ml"),
		Conversion:  mustCatalogConversion(t, 1_000, 1),
		CreatedAt:   mustCatalogInstant(t, 2),
		UpdatedAt:   mustCatalogInstant(t, 2),
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("incompatible dimension error = %v, want domain.ErrValidation", err)
	}

	missingItemID := mustCatalogItemID(t, 999)
	_, err = store.CreatePackaging(ctx, CreatePackagingInput{
		ItemID:      missingItemID,
		Name:        mustCatalogName(t, "Missing item package"),
		EnteredUnit: mustCatalogUnitCode(t, "kg"),
		Conversion:  mustCatalogConversion(t, 1_000_000, 1),
		CreatedAt:   mustCatalogInstant(t, 2),
		UpdatedAt:   mustCatalogInstant(t, 2),
	})
	if !errors.Is(err, domain.ErrInvalidReference) {
		t.Fatalf("missing packaging item error = %v, want domain.ErrInvalidReference", err)
	}

	var itemCount, packagingCount int
	if err := store.database.QueryRowContext(ctx, "SELECT count(*) FROM items").Scan(&itemCount); err != nil {
		t.Fatal(err)
	}
	if err := store.database.QueryRowContext(ctx, "SELECT count(*) FROM item_packagings").Scan(&packagingCount); err != nil {
		t.Fatal(err)
	}
	if itemCount != 1 || packagingCount != 0 {
		t.Fatalf("row counts after rejected writes = items:%d packagings:%d, want 1/0", itemCount, packagingCount)
	}
}

func TestCatalogStoreItemAndPackagingLifecyclesClassifyStaleAndStateConflicts(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "catalog-lifecycle.db"), database.DefaultOpenOptions())
	ctx := context.Background()
	initialAt := mustCatalogInstant(t, 10)
	item := createCatalogItem(t, store, CreateItemInput{
		Name:         mustCatalogName(t, "Butter"),
		BaseUnit:     mustCatalogUnitCode(t, "g"),
		Capabilities: catalog.NewCapabilities(true, false, false),
		CreatedAt:    initialAt,
		UpdatedAt:    initialAt,
	})

	itemUpdatedAt := mustCatalogInstant(t, 20)
	updatedItem, err := store.UpdateItem(ctx, UpdateItemInput{
		ID:                item.Item().ID(),
		Name:              mustCatalogName(t, "Unsalted butter"),
		BaseUnit:          mustCatalogUnitCode(t, "g"),
		Capabilities:      catalog.NewCapabilities(true, false, false),
		ExpectedUpdatedAt: initialAt,
		UpdatedAt:         itemUpdatedAt,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updatedItem.Item().Name().Display() != "Unsalted butter" || !updatedItem.Item().UpdatedAt().Equal(itemUpdatedAt) {
		t.Fatalf("unexpected updated item: %q/%d", updatedItem.Item().Name().Display(), updatedItem.Item().UpdatedAt().UnixMilli())
	}

	_, err = store.UpdateItem(ctx, UpdateItemInput{
		ID:                item.Item().ID(),
		Name:              mustCatalogName(t, "Stale butter"),
		BaseUnit:          mustCatalogUnitCode(t, "g"),
		Capabilities:      catalog.NewCapabilities(true, false, false),
		ExpectedUpdatedAt: initialAt,
		UpdatedAt:         mustCatalogInstant(t, 30),
	})
	if !errors.Is(err, domain.ErrStale) {
		t.Fatalf("stale item update error = %v, want domain.ErrStale", err)
	}

	packagingInitialAt := mustCatalogInstant(t, 21)
	packaging, err := store.CreatePackaging(ctx, CreatePackagingInput{
		ItemID:      item.Item().ID(),
		Name:        mustCatalogName(t, "Block"),
		EnteredUnit: mustCatalogUnitCode(t, "g"),
		Conversion:  mustCatalogConversion(t, 250_000, 1),
		CreatedAt:   packagingInitialAt,
		UpdatedAt:   packagingInitialAt,
	})
	if err != nil {
		t.Fatal(err)
	}
	packagingUpdatedAt := mustCatalogInstant(t, 22)
	packaging, err = store.UpdatePackaging(ctx, UpdatePackagingInput{
		ID:                packaging.Packaging().ID(),
		Name:              mustCatalogName(t, "Large block"),
		EnteredUnit:       mustCatalogUnitCode(t, "g"),
		Conversion:        mustCatalogConversion(t, 500_000, 1),
		ExpectedUpdatedAt: packagingInitialAt,
		UpdatedAt:         packagingUpdatedAt,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.UpdatePackaging(ctx, UpdatePackagingInput{
		ID:                packaging.Packaging().ID(),
		Name:              mustCatalogName(t, "Stale block"),
		EnteredUnit:       mustCatalogUnitCode(t, "g"),
		Conversion:        mustCatalogConversion(t, 500_000, 1),
		ExpectedUpdatedAt: packagingInitialAt,
		UpdatedAt:         mustCatalogInstant(t, 23),
	})
	if !errors.Is(err, domain.ErrStale) {
		t.Fatalf("stale packaging update error = %v, want domain.ErrStale", err)
	}

	packagingArchivedAt := mustCatalogInstant(t, 23)
	packaging, err = store.ArchivePackaging(ctx, ArchivePackagingInput{
		ID: packaging.Packaging().ID(), ExpectedUpdatedAt: packagingUpdatedAt, ArchivedAt: packagingArchivedAt,
	})
	if err != nil || !packaging.Packaging().IsArchived() {
		t.Fatalf("archive packaging = %#v, %v", packaging.Packaging(), err)
	}
	_, err = store.ArchivePackaging(ctx, ArchivePackagingInput{
		ID: packaging.Packaging().ID(), ExpectedUpdatedAt: packagingArchivedAt, ArchivedAt: mustCatalogInstant(t, 24),
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("second packaging archive error = %v, want domain.ErrConflict", err)
	}
	packagingRestoredAt := mustCatalogInstant(t, 24)
	packaging, err = store.RestorePackaging(ctx, RestorePackagingInput{
		ID: packaging.Packaging().ID(), ExpectedUpdatedAt: packagingArchivedAt, UpdatedAt: packagingRestoredAt,
	})
	if err != nil || packaging.Packaging().IsArchived() {
		t.Fatalf("restore packaging = %#v, %v", packaging.Packaging(), err)
	}

	itemArchivedAt := mustCatalogInstant(t, 30)
	archivedItem, err := store.ArchiveItem(ctx, ArchiveItemInput{
		ID: item.Item().ID(), ExpectedUpdatedAt: itemUpdatedAt, ArchivedAt: itemArchivedAt,
	})
	if err != nil || !archivedItem.Item().IsArchived() {
		t.Fatalf("archive item = %#v, %v", archivedItem.Item(), err)
	}
	_, err = store.ArchiveItem(ctx, ArchiveItemInput{
		ID: item.Item().ID(), ExpectedUpdatedAt: itemArchivedAt, ArchivedAt: mustCatalogInstant(t, 40),
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("second item archive error = %v, want domain.ErrConflict", err)
	}
	restoredItem, err := store.RestoreItem(ctx, RestoreItemInput{
		ID: item.Item().ID(), ExpectedUpdatedAt: itemArchivedAt, UpdatedAt: mustCatalogInstant(t, 40),
	})
	if err != nil || restoredItem.Item().IsArchived() {
		t.Fatalf("restore item = %#v, %v", restoredItem.Item(), err)
	}
}

func TestCatalogStoreReconfiguresArchivedPackagingAfterBaseDimensionChange(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "catalog-packaging-recovery.db"), database.DefaultOpenOptions())
	ctx := context.Background()
	itemAt := mustCatalogInstant(t, 1)
	item := createCatalogItem(t, store, CreateItemInput{
		Name:         mustCatalogName(t, "Liquid ingredient"),
		BaseUnit:     mustCatalogUnitCode(t, "g"),
		Capabilities: catalog.NewCapabilities(true, false, false),
		CreatedAt:    itemAt,
		UpdatedAt:    itemAt,
	})
	packagingAt := mustCatalogInstant(t, 2)
	packaging, err := store.CreatePackaging(ctx, CreatePackagingInput{
		ItemID:      item.Item().ID(),
		Name:        mustCatalogName(t, "Mass container"),
		EnteredUnit: mustCatalogUnitCode(t, "kg"),
		Conversion:  mustCatalogConversion(t, 1_000_000, 1),
		CreatedAt:   packagingAt,
		UpdatedAt:   packagingAt,
	})
	if err != nil {
		t.Fatal(err)
	}

	archivedAt := mustCatalogInstant(t, 3)
	packaging, err = store.ArchivePackaging(ctx, ArchivePackagingInput{
		ID: packaging.Packaging().ID(), ExpectedUpdatedAt: packagingAt, ArchivedAt: archivedAt,
	})
	if err != nil {
		t.Fatal(err)
	}
	itemUpdatedAt := mustCatalogInstant(t, 4)
	item, err = store.UpdateItem(ctx, UpdateItemInput{
		ID: item.Item().ID(), Name: item.Item().Name(), SKU: item.Item().SKU(),
		Description: item.Item().Description(), BaseUnit: mustCatalogUnitCode(t, "ml"),
		Capabilities: item.Item().Capabilities(), DefaultSalePrice: item.Item().DefaultSalePrice(),
		ReorderQuantity: item.Item().ReorderQuantity(), ExpectedUpdatedAt: itemAt,
		UpdatedAt: itemUpdatedAt,
	})
	if err != nil {
		t.Fatal(err)
	}
	if item.BaseUnit().Dimension() != domain.DimensionVolume {
		t.Fatalf("updated item base dimension = %s, want VOLUME", item.BaseUnit().Dimension())
	}

	_, err = store.RestorePackaging(ctx, RestorePackagingInput{
		ID: packaging.Packaging().ID(), ExpectedUpdatedAt: archivedAt, UpdatedAt: mustCatalogInstant(t, 5),
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("restore incompatible archived packaging error = %v, want domain.ErrValidation", err)
	}
	_, err = store.ReconfigureArchivedPackaging(ctx, ReconfigureArchivedPackagingInput{
		ID: packaging.Packaging().ID(), Name: mustCatalogName(t, "Stale volume container"),
		EnteredUnit: mustCatalogUnitCode(t, "l"), Conversion: mustCatalogConversion(t, 2_000_000, 1),
		ExpectedUpdatedAt: packagingAt, UpdatedAt: mustCatalogInstant(t, 5),
	})
	if !errors.Is(err, domain.ErrStale) {
		t.Fatalf("stale archived packaging reconfiguration error = %v, want domain.ErrStale", err)
	}
	_, err = store.ReconfigureArchivedPackaging(ctx, ReconfigureArchivedPackagingInput{
		ID: packaging.Packaging().ID(), Name: mustCatalogName(t, "Still-mass container"),
		EnteredUnit: mustCatalogUnitCode(t, "kg"), Conversion: mustCatalogConversion(t, 1_000_000, 1),
		ExpectedUpdatedAt: archivedAt, UpdatedAt: mustCatalogInstant(t, 5),
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("incompatible reconfiguration unit error = %v, want domain.ErrValidation", err)
	}
	_, err = store.ReconfigureArchivedPackaging(ctx, ReconfigureArchivedPackagingInput{
		ID: packaging.Packaging().ID(), Name: mustCatalogName(t, "Unknown-unit container"),
		EnteredUnit: mustCatalogUnitCode(t, "missing-unit"), Conversion: mustCatalogConversion(t, 2_000_000, 1),
		ExpectedUpdatedAt: archivedAt, UpdatedAt: mustCatalogInstant(t, 5),
	})
	if !errors.Is(err, domain.ErrInvalidReference) {
		t.Fatalf("unknown reconfiguration unit error = %v, want domain.ErrInvalidReference", err)
	}

	reconfiguredAt := mustCatalogInstant(t, 5)
	packaging, err = store.ReconfigureArchivedPackaging(ctx, ReconfigureArchivedPackagingInput{
		ID: packaging.Packaging().ID(), Name: mustCatalogName(t, "Two litre container"),
		EnteredUnit: mustCatalogUnitCode(t, "l"), Conversion: mustCatalogConversion(t, 2_000_000, 1),
		ExpectedUpdatedAt: archivedAt, UpdatedAt: reconfiguredAt,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !packaging.Packaging().IsArchived() || packaging.Packaging().Name().Display() != "Two litre container" {
		t.Fatalf("reconfigured packaging state/name = %#v", packaging.Packaging())
	}
	if !packaging.Packaging().UpdatedAt().Equal(reconfiguredAt) || packaging.EnteredUnit().Dimension() != domain.DimensionVolume {
		t.Fatalf("reconfigured packaging version/dimension = %d/%s, want 5/VOLUME",
			packaging.Packaging().UpdatedAt().UnixMilli(), packaging.EnteredUnit().Dimension())
	}
	archivedVersion, ok := packaging.Packaging().ArchivedAt().Get()
	if !ok || !archivedVersion.Equal(reconfiguredAt) {
		t.Fatalf("reconfigured archived timestamp = %#v/%t, want %d", archivedVersion, ok, reconfiguredAt.UnixMilli())
	}

	restoredAt := mustCatalogInstant(t, 6)
	packaging, err = store.RestorePackaging(ctx, RestorePackagingInput{
		ID: packaging.Packaging().ID(), ExpectedUpdatedAt: reconfiguredAt, UpdatedAt: restoredAt,
	})
	if err != nil || packaging.Packaging().IsArchived() {
		t.Fatalf("restore reconfigured packaging = %#v, %v", packaging.Packaging(), err)
	}
	_, err = store.ReconfigureArchivedPackaging(ctx, ReconfigureArchivedPackagingInput{
		ID: packaging.Packaging().ID(), Name: mustCatalogName(t, "Active container"),
		EnteredUnit: mustCatalogUnitCode(t, "l"), Conversion: mustCatalogConversion(t, 2_000_000, 1),
		ExpectedUpdatedAt: restoredAt, UpdatedAt: mustCatalogInstant(t, 7),
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("active packaging reconfiguration error = %v, want domain.ErrConflict", err)
	}
}

func TestCatalogStoreListsSummaryPagesWithFiltersAndStableCursor(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "catalog-list.db"), database.DefaultOpenOptions())
	ctx := context.Background()
	alpha := createNamedCatalogItem(t, store, "Alpha", 1, catalog.NewCapabilities(true, false, false))
	beta := createNamedCatalogItem(t, store, "Beta", 2, catalog.NewCapabilities(false, false, true))
	createNamedCatalogItem(t, store, "Gamma", 3, catalog.NewCapabilities(true, false, true))
	if _, err := store.ArchiveItem(ctx, ArchiveItemInput{
		ID: beta.Item().ID(), ExpectedUpdatedAt: mustCatalogInstant(t, 2), ArchivedAt: mustCatalogInstant(t, 4),
	}); err != nil {
		t.Fatal(err)
	}

	pageSizeTwo := mustCatalogPageSize(t, 2)
	first, err := store.ListItems(ctx, ItemListFilter{Archive: domain.ArchiveAll, PageSize: pageSizeTwo})
	if err != nil {
		t.Fatal(err)
	}
	if got := summaryNames(first.Items()); len(got) != 2 || got[0] != "Alpha" || got[1] != "Beta" {
		t.Fatalf("first page = %v, want [Alpha Beta]", got)
	}
	cursor, ok := first.Next().Get()
	if !ok || cursor.Name.Display() != "Beta" || cursor.ID != beta.Item().ID() {
		t.Fatalf("first page cursor = %#v/%t, want Beta/%d", cursor, ok, beta.Item().ID().Int64())
	}
	second, err := store.ListItems(ctx, ItemListFilter{
		Archive: domain.ArchiveAll, After: domain.Some(cursor), PageSize: pageSizeTwo,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := summaryNames(second.Items()); len(got) != 1 || got[0] != "Gamma" {
		t.Fatalf("second page = %v, want [Gamma]", got)
	}
	if second.Next().IsSome() {
		t.Fatal("final page unexpectedly has a cursor")
	}

	pageSizeTen := mustCatalogPageSize(t, 10)
	active, err := store.ListItems(ctx, ItemListFilter{Archive: domain.ArchiveActive, PageSize: pageSizeTen})
	if err != nil {
		t.Fatal(err)
	}
	if got := summaryNames(active.Items()); len(got) != 2 || got[0] != "Alpha" || got[1] != "Gamma" {
		t.Fatalf("active items = %v, want [Alpha Gamma]", got)
	}
	defaultActive, err := store.ListItems(ctx, ItemListFilter{PageSize: pageSizeTen})
	if err != nil {
		t.Fatal(err)
	}
	if got := summaryNames(defaultActive.Items()); len(got) != 2 || got[0] != "Alpha" || got[1] != "Gamma" {
		t.Fatalf("default active items = %v, want [Alpha Gamma]", got)
	}
	archived, err := store.ListItems(ctx, ItemListFilter{Archive: domain.ArchiveArchived, PageSize: pageSizeTen})
	if err != nil {
		t.Fatal(err)
	}
	if got := summaryNames(archived.Items()); len(got) != 1 || got[0] != "Beta" {
		t.Fatalf("archived items = %v, want [Beta]", got)
	}
	sellable, err := store.ListItems(ctx, ItemListFilter{
		Archive: domain.ArchiveAll, RequireCapabilities: catalog.NewCapabilities(false, false, true), PageSize: pageSizeTen,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := summaryNames(sellable.Items()); len(got) != 2 || got[0] != "Beta" || got[1] != "Gamma" {
		t.Fatalf("sellable items = %v, want [Beta Gamma]", got)
	}
	searched, err := store.ListItems(ctx, ItemListFilter{
		Archive: domain.ArchiveAll, Search: domain.Some(mustCatalogText(t, "  BETA  ")), PageSize: pageSizeTen,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := summaryNames(searched.Items()); len(got) != 1 || got[0] != "Beta" {
		t.Fatalf("searched items = %v, want [Beta]", got)
	}
	empty, err := store.ListItems(ctx, ItemListFilter{
		Archive: domain.ArchiveAll, Search: domain.Some(mustCatalogText(t, "missing")), PageSize: pageSizeTen,
	})
	if err != nil {
		t.Fatal(err)
	}
	if empty.Items() == nil || len(empty.Items()) != 0 {
		t.Fatalf("empty page items = %#v, want non-nil empty slice", empty.Items())
	}
	if alpha.Item().ID().IsZero() {
		t.Fatal("unexpected zero Alpha ID")
	}
}

func TestCatalogStoreClassifiesMissingAndCorruptSnapshots(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "catalog-corrupt.db"), database.DefaultOpenOptions())
	ctx := context.Background()
	missing := mustCatalogItemID(t, 999)
	if _, err := store.GetItem(ctx, missing); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("missing item error = %v, want domain.ErrNotFound", err)
	}

	item := createNamedCatalogItem(t, store, "Vanilla", 1, catalog.NewCapabilities(true, false, false))
	if _, err := store.database.ExecContext(ctx, "UPDATE items SET normalized_name = 'not-vanilla' WHERE id = ?", item.Item().ID().Int64()); err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetItem(ctx, item.Item().ID()); !errors.Is(err, domain.ErrCorruptData) {
		t.Fatalf("corrupt item error = %v, want domain.ErrCorruptData", err)
	}
	if _, err := store.ListItems(ctx, ItemListFilter{Archive: domain.ArchiveAll, PageSize: mustCatalogPageSize(t, 10)}); !errors.Is(err, domain.ErrCorruptData) {
		t.Fatalf("corrupt list error = %v, want domain.ErrCorruptData", err)
	}
}

func createNamedCatalogItem(t *testing.T, store *Store, name string, milliseconds int64, capabilities catalog.Capabilities) ItemAggregate {
	t.Helper()
	at := mustCatalogInstant(t, milliseconds)
	return createCatalogItem(t, store, CreateItemInput{
		Name: mustCatalogName(t, name), BaseUnit: mustCatalogUnitCode(t, "g"),
		Capabilities: capabilities, CreatedAt: at, UpdatedAt: at,
	})
}

func newCatalogSnapshotFixture(t *testing.T) (*Store, *Store, ItemAggregate, PackagingAggregate) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "catalog-snapshot.db")
	reader := newAdapterTestStore(t, path, database.DefaultOpenOptions())
	writer := newAdapterTestStore(t, path, database.DefaultOpenOptions())
	item := createCatalogItem(t, reader, CreateItemInput{
		Name:         mustCatalogName(t, "Snapshot ingredient"),
		BaseUnit:     mustCatalogUnitCode(t, "g"),
		Capabilities: catalog.NewCapabilities(true, false, false),
		CreatedAt:    mustCatalogInstant(t, 1),
		UpdatedAt:    mustCatalogInstant(t, 1),
	})
	packaging, err := reader.CreatePackaging(context.Background(), CreatePackagingInput{
		ItemID:      item.Item().ID(),
		Name:        mustCatalogName(t, "Snapshot package"),
		EnteredUnit: mustCatalogUnitCode(t, "kg"),
		Conversion:  mustCatalogConversion(t, 1_000_000, 1),
		CreatedAt:   mustCatalogInstant(t, 2),
		UpdatedAt:   mustCatalogInstant(t, 2),
	})
	if err != nil {
		t.Fatal(err)
	}
	return reader, writer, item, packaging
}

func transitionCatalogAggregateToVolume(
	ctx context.Context,
	writer *Store,
	itemID domain.ItemID,
	packagingID domain.PackagingID,
) error {
	return writer.database.Write(ctx, func(tx *database.WriteTx) error {
		statements := []struct {
			query string
			args  []any
		}{
			{
				query: `UPDATE item_packagings SET updated_at_ms = 3, archived_at_ms = 3 WHERE id = ?`,
				args:  []any{packagingID.Int64()},
			},
			{
				query: `UPDATE items SET base_unit_code = 'ml', updated_at_ms = 3 WHERE id = ?`,
				args:  []any{itemID.Int64()},
			},
			{
				query: `
					UPDATE item_packagings
					SET entered_unit_code = 'l', conversion_numerator_atomic = 1000000,
						conversion_denominator = 1, updated_at_ms = 4, archived_at_ms = 4
					WHERE id = ?
				`,
				args: []any{packagingID.Int64()},
			},
			{
				query: `UPDATE item_packagings SET updated_at_ms = 5, archived_at_ms = NULL WHERE id = ?`,
				args:  []any{packagingID.Int64()},
			},
		}
		for _, statement := range statements {
			if _, err := tx.ExecContext(ctx, statement.query, statement.args...); err != nil {
				return err
			}
		}
		return nil
	})
}

func createCatalogItem(t *testing.T, store *Store, input CreateItemInput) ItemAggregate {
	t.Helper()
	item, err := store.CreateItem(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	return item
}

func summaryNames(items []catalog.ItemSummary) []string {
	result := make([]string, len(items))
	for index, item := range items {
		result[index] = item.Name().Display()
	}
	return result
}

func mustCatalogName(t *testing.T, raw string) domain.UniqueName {
	t.Helper()
	value, err := domain.NewUniqueName(raw)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func mustCatalogText(t *testing.T, raw string) domain.NonEmptyText {
	t.Helper()
	value, err := domain.NewNonEmptyText(raw)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func mustCatalogUnitCode(t *testing.T, raw string) domain.UnitCode {
	t.Helper()
	value, err := domain.NewUnitCode(raw)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func mustCatalogInstant(t *testing.T, milliseconds int64) domain.UTCInstant {
	t.Helper()
	value, err := domain.UTCInstantFromUnixMilli(milliseconds)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func mustCatalogConversion(t *testing.T, numerator, denominator int64) domain.UnitConversion {
	t.Helper()
	value, err := domain.NewUnitConversion(numerator, denominator)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func mustCatalogItemID(t *testing.T, raw int64) domain.ItemID {
	t.Helper()
	value, err := domain.NewItemID(raw)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func mustCatalogPageSize(t *testing.T, raw int) ItemPageSize {
	t.Helper()
	value, err := NewItemPageSize(raw)
	if err != nil {
		t.Fatal(err)
	}
	return value
}
