package database

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"
)

func TestCatalogSchemaEnforcesExactTypesLifecycleAndUnitRules(t *testing.T) {
	db := openSchemaTestDatabase(t)

	var units, baseUnits int
	if err := db.Conn.QueryRow(`
		SELECT COUNT(*), SUM(is_item_base) FROM measurement_units
	`).Scan(&units, &baseUnits); err != nil {
		t.Fatal(err)
	}
	if units != 9 || baseUnits != 3 {
		t.Fatalf("seeded units = %d and base units = %d, want 9 and 3", units, baseUnits)
	}

	itemID := insertTestItem(t, db, "Farinha", "farinha", "g", true, false, false)
	var quantity, value int64
	if err := db.Conn.QueryRow(`
		SELECT quantity_atomic, inventory_value_micro
		FROM inventory_balances WHERE item_id = ?
	`, itemID).Scan(&quantity, &value); err != nil {
		t.Fatal(err)
	}
	if quantity != 0 || value != 0 {
		t.Fatalf("initial balance = (%d, %d), want zero", quantity, value)
	}
	if _, err := db.Conn.Exec(`
		UPDATE items SET sku = 'FAR-01', normalized_sku = 'far-01' WHERE id = ?
	`, itemID); err != nil {
		t.Fatal(err)
	}

	expectExecError(t, db.Conn, `
		INSERT INTO items (
			name, normalized_name, base_unit_code,
			is_purchasable, is_producible, is_sellable,
			created_at_ms, updated_at_ms
		) VALUES ('FARINHA', 'farinha', 'g', 1, 0, 0, 1, 1)
	`)
	expectExecError(t, db.Conn, `
		INSERT INTO items (
			name, normalized_name, base_unit_code,
			sku, normalized_sku, is_purchasable, is_producible, is_sellable,
			created_at_ms, updated_at_ms
		) VALUES ('Other flour', 'other flour', 'g', 'far-01', 'far-01', 1, 0, 0, 1, 1)
	`)
	expectExecError(t, db.Conn, `
		INSERT INTO items (
			name, normalized_name, base_unit_code,
			is_purchasable, is_producible, is_sellable,
			created_at_ms, updated_at_ms
		) VALUES ('Bag', 'bag', 'kg', 1, 0, 0, 1, 1)
	`)
	expectExecError(t, db.Conn, `
		INSERT INTO items (
			name, normalized_name, base_unit_code,
			is_purchasable, is_producible, is_sellable,
			created_at_ms, updated_at_ms
		) VALUES ('Invalid', 'invalid', 'g', 1.5, 0, 0, 1, 1)
	`)
	expectExecError(t, db.Conn, `UPDATE measurement_units SET symbol = 'gram' WHERE code = 'g'`)
	expectExecError(t, db.Conn, `DELETE FROM items WHERE id = ?`, itemID)
	expectExecError(t, db.Conn, `UPDATE schema_migrations SET name = 'changed.sql' WHERE version = 1`)
}

func TestRecipeSchemaPreservesPublishedRevisionHistory(t *testing.T) {
	db := openSchemaTestDatabase(t)
	ingredientID := insertTestItem(t, db, "Flour", "flour", "g", true, false, false)
	outputID := insertTestItem(t, db, "Cake", "cake", "each", false, true, true)

	if _, err := db.Conn.Exec(`
		INSERT INTO item_packagings (
			item_id, name, normalized_name, entered_unit_code,
			conversion_numerator_atomic, conversion_denominator,
			created_at_ms, updated_at_ms
		) VALUES (?, '5 kg bag', '5 kg bag', 'kg', 5000000, 1, 1, 1)
	`, ingredientID); err != nil {
		t.Fatal(err)
	}
	expectExecError(t, db.Conn, `
		INSERT INTO item_packagings (
			item_id, name, normalized_name, entered_unit_code,
			conversion_numerator_atomic, conversion_denominator,
			created_at_ms, updated_at_ms
		) VALUES (?, 'Bottle', 'bottle', 'ml', 1000, 1, 1, 1)
	`, ingredientID)

	result, err := db.Conn.Exec(`
		INSERT INTO recipes (
			name, normalized_name, output_item_id, created_at_ms, updated_at_ms
		) VALUES ('Cake recipe', 'cake recipe', ?, 1, 1)
	`, outputID)
	if err != nil {
		t.Fatal(err)
	}
	recipeID, _ := result.LastInsertId()
	result, err = db.Conn.Exec(`
		INSERT INTO recipe_revisions (
			recipe_id, revision_number, standard_yield_quantity_atomic,
			instructions, preparation_time_minutes, created_at_ms
		) VALUES (?, 1, 1000, 'Mix and bake', 45, 2)
	`, recipeID)
	if err != nil {
		t.Fatal(err)
	}
	revisionID, _ := result.LastInsertId()
	expectExecError(t, db.Conn, `UPDATE items SET base_unit_code = 'g' WHERE id = ?`, outputID)
	if _, err := db.Conn.Exec(`
		INSERT INTO recipe_revision_components (
			recipe_revision_id, component_order, item_id, quantity_atomic,
			entered_unit_code, conversion_numerator_atomic, conversion_denominator,
			created_at_ms
		) VALUES (?, 1, ?, 500000, 'g', 1000, 1, 2)
	`, revisionID, ingredientID); err != nil {
		t.Fatal(err)
	}

	expectExecError(t, db.Conn, `
		INSERT INTO recipe_revision_components (
			recipe_revision_id, component_order, item_id, quantity_atomic,
			entered_unit_code, conversion_numerator_atomic, conversion_denominator,
			created_at_ms
		) VALUES (?, 2, ?, 1000, 'each', 1000, 1, 2)
	`, revisionID, outputID)
	expectExecError(t, db.Conn, `
		UPDATE recipe_revisions SET instructions = 'Changed' WHERE id = ?
	`, revisionID)
	expectExecError(t, db.Conn, `UPDATE recipes SET output_item_id = ? WHERE id = ?`, ingredientID, recipeID)
	expectExecError(t, db.Conn, `UPDATE items SET base_unit_code = 'ml' WHERE id = ?`, ingredientID)
}

func TestArchivedPackagingAllowsSafeBaseUnitCorrection(t *testing.T) {
	db := openSchemaTestDatabase(t)
	itemID := insertTestItem(t, db, "Mistyped item", "mistyped item", "g", true, false, false)
	result, err := db.Conn.Exec(`
		INSERT INTO item_packagings (
			item_id, name, normalized_name, entered_unit_code,
			conversion_numerator_atomic, conversion_denominator,
			created_at_ms, updated_at_ms
		) VALUES (?, 'Bag', 'bag', 'kg', 5000000, 1, 1, 1)
	`, itemID)
	if err != nil {
		t.Fatal(err)
	}
	packagingID, _ := result.LastInsertId()

	expectExecError(t, db.Conn, `
		UPDATE items SET base_unit_code = 'ml', updated_at_ms = 2 WHERE id = ?
	`, itemID)
	if _, err := db.Conn.Exec(`
		UPDATE item_packagings
		SET archived_at_ms = 2, updated_at_ms = 2
		WHERE id = ?
	`, packagingID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Conn.Exec(`
		UPDATE items SET base_unit_code = 'ml', updated_at_ms = 3 WHERE id = ?
	`, itemID); err != nil {
		t.Fatal(err)
	}
	expectExecError(t, db.Conn, `
		UPDATE item_packagings SET archived_at_ms = NULL, updated_at_ms = 4 WHERE id = ?
	`, packagingID)
	if _, err := db.Conn.Exec(`
		UPDATE item_packagings
		SET entered_unit_code = 'ml',
			conversion_numerator_atomic = 1000,
			conversion_denominator = 1,
			archived_at_ms = 4,
			updated_at_ms = 4
		WHERE id = ?
	`, packagingID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Conn.Exec(`
		UPDATE item_packagings SET archived_at_ms = NULL, updated_at_ms = 5 WHERE id = ?
	`, packagingID); err != nil {
		t.Fatal(err)
	}
}

func TestDocumentSchemaEnforcesPostingShapeAndImmutability(t *testing.T) {
	db := openSchemaTestDatabase(t)
	itemID := insertTestItem(t, db, "Sugar", "sugar", "g", true, false, true)

	result, err := db.Conn.Exec(`
		INSERT INTO counterparties (
			name, created_at_ms, updated_at_ms
		) VALUES ('Supplier', 1, 1)
	`)
	if err != nil {
		t.Fatal(err)
	}
	supplierID, _ := result.LastInsertId()
	if _, err := db.Conn.Exec(`
		INSERT INTO counterparty_roles (counterparty_id, role, created_at_ms)
		VALUES (?, 'SUPPLIER', 1)
	`, supplierID); err != nil {
		t.Fatal(err)
	}

	purchaseID := insertTestDocument(
		t, db, "PURCHASE", 1, nil, nil, supplierID, "purchase-1",
	)
	lineID := insertTestLine(t, db, purchaseID, 1, itemID, "IN", 1000, "g", 500000, 500, nil)

	expectExecError(t, db.Conn, `
		INSERT INTO stock_document_lines (
			document_id, line_order, item_id, direction, quantity_atomic,
			entered_unit_code, conversion_numerator_atomic, conversion_denominator,
			inventory_value_micro, commercial_total_minor
		) VALUES (?, 2, ?, 'OUT', 1000, 'g', 1000, 1, 500000, 500)
	`, purchaseID, itemID)
	expectExecError(t, db.Conn, `UPDATE stock_documents SET notes = 'changed' WHERE id = ?`, purchaseID)
	expectExecError(t, db.Conn, `UPDATE stock_document_lines SET quantity_atomic = 2 WHERE id = ?`, lineID)
	expectExecError(t, db.Conn, `UPDATE app_settings SET currency_code = 'USD' WHERE id = 1`)

	freeCandidateID := insertTestDocument(
		t, db, "PURCHASE", 2, nil, nil, nil, "purchase-zero",
	)
	expectExecError(t, db.Conn, `
		INSERT INTO stock_document_lines (
			document_id, line_order, item_id, direction, quantity_atomic,
			entered_unit_code, conversion_numerator_atomic, conversion_denominator,
			inventory_value_micro, commercial_total_minor
		) VALUES (?, 1, ?, 'IN', 1000, 'g', 1000, 1, 0, 0)
	`, freeCandidateID, itemID)
	expectExecError(t, db.Conn, `
		INSERT INTO stock_documents (
			kind, idempotency_key, posting_sequence, occurred_on, posted_at_ms,
			currency_code, currency_minor_digits
		) VALUES ('SALE', 'bad-sequence', 1, '2026-07-14', 3, 'BRL', 2)
	`)
	expectExecError(t, db.Conn, `
		INSERT INTO stock_documents (
			kind, idempotency_key, posting_sequence, occurred_on, posted_at_ms,
			currency_code, currency_minor_digits
		) VALUES ('SALE', 'bad-date', 3, '14/07/2026', 3, 'BRL', 2)
	`)
}

func TestReversalSchemaRequiresExactLinkedInverse(t *testing.T) {
	db := openSchemaTestDatabase(t)
	itemID := insertTestItem(t, db, "Milk", "milk", "ml", true, false, true)
	purchaseID := insertTestDocument(t, db, "PURCHASE", 1, nil, nil, nil, "purchase")
	targetLineID := insertTestLine(t, db, purchaseID, 1, itemID, "IN", 1000, "ml", 500000, 250, nil)
	reversalID := insertTestDocument(
		t, db, "REVERSAL", 2, "EXACT_REVERSAL", purchaseID, nil, "reverse-purchase",
	)

	expectExecError(t, db.Conn, `
		INSERT INTO stock_document_lines (
			document_id, line_order, item_id, direction, quantity_atomic,
			entered_unit_code, conversion_numerator_atomic, conversion_denominator,
			inventory_value_micro, commercial_total_minor, reverses_line_id
		) VALUES (?, 1, ?, 'OUT', 999, 'ml', 1000, 1, 500000, 250, ?)
	`, reversalID, itemID, targetLineID)
	reversalLineID := insertTestLine(
		t, db, reversalID, 1, itemID, "OUT", 1000, "ml", 500000, 250, targetLineID,
	)
	if reversalLineID == 0 {
		t.Fatal("expected reversal line id")
	}
	expectExecError(t, db.Conn, `
		INSERT INTO stock_documents (
			kind, idempotency_key, posting_sequence, occurred_on, posted_at_ms,
			currency_code, currency_minor_digits, reason_code, reverses_document_id
		) VALUES ('REVERSAL', 'reverse-again', 3, '2026-07-14', 3,
			'BRL', 2, 'EXACT_REVERSAL', ?)
	`, purchaseID)
}

func TestAdjustmentAndProductionMetadataMatchCanonicalLines(t *testing.T) {
	db := openSchemaTestDatabase(t)
	ingredientID := insertTestItem(t, db, "Cocoa", "cocoa", "g", true, false, false)
	outputID := insertTestItem(t, db, "Truffle", "truffle", "each", false, true, true)

	adjustmentID := insertTestDocument(
		t, db, "ADJUSTMENT", 1, "PHYSICAL_COUNT", nil, nil, "count",
	)
	countLineID := insertTestLine(t, db, adjustmentID, 1, ingredientID, "IN", 200, "g", 0, nil, nil)
	if _, err := db.Conn.Exec(`
		INSERT INTO adjustment_line_details (
			line_id, expected_quantity_atomic, observed_quantity_atomic
		) VALUES (?, 500, 700)
	`, countLineID); err != nil {
		t.Fatal(err)
	}
	expectExecError(t, db.Conn, `
		INSERT INTO adjustment_line_details (
			line_id, expected_quantity_atomic, observed_quantity_atomic
		) VALUES (?, 500, 650)
	`, countLineID)

	result, err := db.Conn.Exec(`
		INSERT INTO recipes (
			name, normalized_name, output_item_id, created_at_ms, updated_at_ms
		) VALUES ('Truffle recipe', 'truffle recipe', ?, 1, 1)
	`, outputID)
	if err != nil {
		t.Fatal(err)
	}
	recipeID, _ := result.LastInsertId()
	result, err = db.Conn.Exec(`
		INSERT INTO recipe_revisions (
			recipe_id, revision_number, standard_yield_quantity_atomic,
			instructions, preparation_time_minutes, created_at_ms
		) VALUES (?, 1, 1000, '', 10, 1)
	`, recipeID)
	if err != nil {
		t.Fatal(err)
	}
	revisionID, _ := result.LastInsertId()
	productionID := insertTestDocument(t, db, "PRODUCTION", 2, nil, nil, nil, "production")
	inputLineID := insertTestLine(t, db, productionID, 1, ingredientID, "OUT", 100, "g", 10000, nil, nil)
	outputLineID := insertTestLine(t, db, productionID, 2, outputID, "IN", 1000, "each", 15000, nil, nil)
	if _, err := db.Conn.Exec(`
		INSERT INTO production_runs (
			document_id, recipe_revision_id, output_line_id, direct_production_cost_micro
		) VALUES (?, ?, ?, 5000)
	`, productionID, revisionID, outputLineID); err != nil {
		t.Fatal(err)
	}
	expectExecError(t, db.Conn, `
		INSERT INTO production_runs (
			document_id, recipe_revision_id, output_line_id, direct_production_cost_micro
		) VALUES (?, ?, ?, 0)
	`, productionID, revisionID, inputLineID)
}

func TestLotsAndAllocationsPreventOverconsumptionAndSupportExactRestore(t *testing.T) {
	db := openSchemaTestDatabase(t)
	itemID := insertTestItem(t, db, "Butter", "butter", "g", true, false, true)
	purchaseID := insertTestDocument(t, db, "PURCHASE", 1, nil, nil, nil, "purchase-lot")
	purchaseLineID := insertTestLine(t, db, purchaseID, 1, itemID, "IN", 100, "g", 100000, 100, nil)
	result, err := db.Conn.Exec(`
		INSERT INTO inventory_lots (
			item_id, source_line_id, initial_quantity_atomic,
			lot_code, originated_on, expires_on, created_at_ms
		) VALUES (?, ?, 100, 'LOT-1', '2026-07-14', '2026-07-01', 1)
	`, itemID, purchaseLineID)
	if err != nil {
		t.Fatal(err)
	}
	lotID, _ := result.LastInsertId()

	saleID := insertTestDocument(t, db, "SALE", 2, nil, nil, nil, "sale-1")
	saleLineID := insertTestLine(t, db, saleID, 1, itemID, "OUT", 60, "g", 60000, 150, nil)
	result, err = db.Conn.Exec(`
		INSERT INTO lot_allocations (line_id, lot_id, quantity_atomic, created_at_ms)
		VALUES (?, ?, 60, 2)
	`, saleLineID, lotID)
	if err != nil {
		t.Fatal(err)
	}
	allocationID, _ := result.LastInsertId()

	secondSaleID := insertTestDocument(t, db, "SALE", 3, nil, nil, nil, "sale-2")
	secondSaleLineID := insertTestLine(t, db, secondSaleID, 1, itemID, "OUT", 50, "g", 50000, 120, nil)
	expectExecError(t, db.Conn, `
		INSERT INTO lot_allocations (line_id, lot_id, quantity_atomic, created_at_ms)
		VALUES (?, ?, 50, 3)
	`, secondSaleLineID, lotID)

	reversalID := insertTestDocument(
		t, db, "REVERSAL", 4, "EXACT_REVERSAL", saleID, nil, "reverse-sale",
	)
	restoringLineID := insertTestLine(
		t, db, reversalID, 1, itemID, "IN", 60, "g", 60000, 150, saleLineID,
	)
	if _, err := db.Conn.Exec(`
		INSERT INTO lot_allocations (
			line_id, lot_id, quantity_atomic, restores_allocation_id, created_at_ms
		) VALUES (?, ?, 60, ?, 4)
	`, restoringLineID, lotID, allocationID); err != nil {
		t.Fatal(err)
	}

	var netConsumed int64
	if err := db.Conn.QueryRow(`
		SELECT SUM(CASE WHEN restores_allocation_id IS NULL
			THEN quantity_atomic ELSE -quantity_atomic END)
		FROM lot_allocations WHERE lot_id = ?
	`, lotID).Scan(&netConsumed); err != nil {
		t.Fatal(err)
	}
	if netConsumed != 0 {
		t.Fatalf("net lot consumption = %d, want 0 after exact restore", netConsumed)
	}
	expectExecError(t, db.Conn, `DELETE FROM inventory_lots WHERE id = ?`, lotID)
}

func TestLotAllocationCannotConsumeALaterPostingLot(t *testing.T) {
	db := openSchemaTestDatabase(t)
	itemID := insertTestItem(t, db, "Cream", "cream", "ml", true, false, true)
	earlierSaleID := insertTestDocument(t, db, "SALE", 1, nil, nil, nil, "early-sale")
	laterPurchaseID := insertTestDocument(t, db, "PURCHASE", 2, nil, nil, nil, "late-purchase")
	sourceLineID := insertTestLine(
		t, db, laterPurchaseID, 1, itemID, "IN", 1000, "ml", 100000, 100, nil,
	)
	result, err := db.Conn.Exec(`
		INSERT INTO inventory_lots (
			item_id, source_line_id, initial_quantity_atomic, originated_on, created_at_ms
		) VALUES (?, ?, 1000, '2026-07-14', 2)
	`, itemID, sourceLineID)
	if err != nil {
		t.Fatal(err)
	}
	laterLotID, _ := result.LastInsertId()
	earlierSaleLineID := insertTestLine(
		t, db, earlierSaleID, 1, itemID, "OUT", 1000, "ml", 100000, 150, nil,
	)

	expectExecError(t, db.Conn, `
		INSERT INTO lot_allocations (line_id, lot_id, quantity_atomic, created_at_ms)
		VALUES (?, ?, 1000, 3)
	`, earlierSaleLineID, laterLotID)
}

func TestInboundReversalMustConsumeItsTargetLineLot(t *testing.T) {
	db := openSchemaTestDatabase(t)
	itemID := insertTestItem(t, db, "Oil", "oil", "ml", true, false, false)

	firstPurchaseID := insertTestDocument(t, db, "PURCHASE", 1, nil, nil, nil, "purchase-first")
	firstLineID := insertTestLine(t, db, firstPurchaseID, 1, itemID, "IN", 100, "ml", 10000, 10, nil)
	firstLotID := insertTestLot(t, db, itemID, firstLineID, 100, 1)

	secondPurchaseID := insertTestDocument(t, db, "PURCHASE", 2, nil, nil, nil, "purchase-second")
	secondLineID := insertTestLine(t, db, secondPurchaseID, 1, itemID, "IN", 100, "ml", 10000, 10, nil)
	secondLotID := insertTestLot(t, db, itemID, secondLineID, 100, 2)

	reversalID := insertTestDocument(
		t, db, "REVERSAL", 3, "EXACT_REVERSAL", secondPurchaseID, nil, "reverse-second",
	)
	reversalLineID := insertTestLine(
		t, db, reversalID, 1, itemID, "OUT", 100, "ml", 10000, 10, secondLineID,
	)
	expectExecError(t, db.Conn, `
		INSERT INTO lot_allocations (line_id, lot_id, quantity_atomic, created_at_ms)
		VALUES (?, ?, 100, 3)
	`, reversalLineID, firstLotID)
	if _, err := db.Conn.Exec(`
		INSERT INTO lot_allocations (line_id, lot_id, quantity_atomic, created_at_ms)
		VALUES (?, ?, 100, 3)
	`, reversalLineID, secondLotID); err != nil {
		t.Fatal(err)
	}
}

func TestInventoryBalanceRejectsInvalidProjectionStates(t *testing.T) {
	db := openSchemaTestDatabase(t)
	itemID := insertTestItem(t, db, "Egg", "egg", "each", true, false, true)
	expectExecError(t, db.Conn, `
		UPDATE inventory_balances SET quantity_atomic = -1 WHERE item_id = ?
	`, itemID)
	expectExecError(t, db.Conn, `
		UPDATE inventory_balances SET inventory_value_micro = 1 WHERE item_id = ?
	`, itemID)
	expectExecError(t, db.Conn, `DELETE FROM inventory_balances WHERE item_id = ?`, itemID)
}

func openSchemaTestDatabase(t *testing.T) *Database {
	t.Helper()
	db, err := NewDatabase(filepath.Join(t.TempDir(), "schema-test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close test database: %v", err)
		}
	})
	return db
}

func insertTestItem(
	t *testing.T,
	db *Database,
	name string,
	normalizedName string,
	baseUnit string,
	purchasable bool,
	producible bool,
	sellable bool,
) int64 {
	t.Helper()
	result, err := db.Conn.Exec(`
		INSERT INTO items (
			name, normalized_name, base_unit_code,
			is_purchasable, is_producible, is_sellable,
			created_at_ms, updated_at_ms
		) VALUES (?, ?, ?, ?, ?, ?, 1, 1)
	`, name, normalizedName, baseUnit, purchasable, producible, sellable)
	if err != nil {
		t.Fatal(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func insertTestDocument(
	t *testing.T,
	db *Database,
	kind string,
	sequence int64,
	reason any,
	reversesDocument any,
	counterparty any,
	idempotencyKey string,
) int64 {
	t.Helper()
	result, err := db.Conn.Exec(`
		INSERT INTO stock_documents (
			kind, idempotency_key, posting_sequence, counterparty_id,
			occurred_on, posted_at_ms, currency_code, currency_minor_digits,
			reason_code, reverses_document_id
		) VALUES (?, ?, ?, ?, '2026-07-14', ?, 'BRL', 2, ?, ?)
	`, kind, idempotencyKey, sequence, counterparty, sequence, reason, reversesDocument)
	if err != nil {
		t.Fatal(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func insertTestLine(
	t *testing.T,
	db *Database,
	documentID int64,
	lineOrder int,
	itemID int64,
	direction string,
	quantity int64,
	unitCode string,
	inventoryValue int64,
	commercialTotal any,
	reversesLine any,
) int64 {
	t.Helper()
	result, err := db.Conn.Exec(`
		INSERT INTO stock_document_lines (
			document_id, line_order, item_id, direction, quantity_atomic,
			entered_unit_code, conversion_numerator_atomic, conversion_denominator,
			inventory_value_micro, commercial_total_minor, reverses_line_id
		) VALUES (?, ?, ?, ?, ?, ?, 1000, 1, ?, ?, ?)
	`, documentID, lineOrder, itemID, direction, quantity, unitCode,
		inventoryValue, commercialTotal, reversesLine)
	if err != nil {
		t.Fatal(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func insertTestLot(
	t *testing.T,
	db *Database,
	itemID int64,
	sourceLineID int64,
	quantity int64,
	createdAt int64,
) int64 {
	t.Helper()
	result, err := db.Conn.Exec(`
		INSERT INTO inventory_lots (
			item_id, source_line_id, initial_quantity_atomic, originated_on, created_at_ms
		) VALUES (?, ?, ?, '2026-07-14', ?)
	`, itemID, sourceLineID, quantity, createdAt)
	if err != nil {
		t.Fatal(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func expectExecError(t *testing.T, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.Exec(query, args...); err == nil {
		t.Fatalf("expected statement to fail: %s; args=%s", query, fmt.Sprint(args...))
	}
}
