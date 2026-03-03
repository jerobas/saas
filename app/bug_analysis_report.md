# Bug Analysis Report (`app/database/schemas`, `app/model`, `app/repository`)

This report lists bugs found without modifying the analyzed files.  
Each item includes **impact** and a **proposed fix snippet**.

---

## 1) Integration functional bugs

### 1.1 `events` insert flow can fail if caller omits explicit `DRAFT`
**Where:** `app/database/schemas/003_events.sql`, `app/model/event.model.go`, `app/repository/event.repository.go`

**Problem:** The DB trigger allows inserts only when `NEW.status = 'DRAFT'`. Repository `Create` always sends `evt.Status`; if caller sends empty string (Go zero-value), insert fails despite DB default.

**Impact:** Event creation fails unexpectedly when service assumes DB default status.

**Fix snippet (repository-side fallback):**
```go
status := evt.Status
if status == "" {
    status = "DRAFT"
}

res, err := r.db.Conn.Exec(query,
    evt.EventType,
    status,
    evt.CounterpartyEntityId,
    evt.Notes,
    evt.OccurredAt,
)
```

---

## 2) Integration syntax bugs

### 2.1 `purchase_lines` column mismatch: schema uses `unit_cost`, model/repo use `unit_price`
**Where:**
- Schema: `app/database/schemas/006_purchases.sql`
- Model: `app/model/purchase_line.model.go`
- Repository: `app/repository/purchase_line.repository.go`

**Problem:** SQL table defines `unit_cost`, but code inserts/selects `unit_price`.

**Impact:** All purchase-line CRUD paths fail at runtime with SQL column errors.

**Fix snippet (align to `unit_cost`):**
```go
// model
UnitCost int `json:"unit_cost"`

// repository INSERT/SELECT
INSERT INTO purchase_lines (event_id, item_id, quantity, unit_cost) VALUES (?, ?, ?, ?)
SELECT id, event_id, item_id, quantity, unit_cost, created_at FROM purchase_lines
```

### 2.2 Wrong events filter column in repository
**Where:** `app/repository/event.repository.go`

**Problem:** `GetAllByCounterpartyID` uses `WHERE counterparty_id = ?`, but schema column is `counterparty_entity_id`.

**Impact:** Query always fails (`no such column: counterparty_id`).

**Fix snippet:**
```sql
WHERE counterparty_entity_id = ?
```

### 2.3 Wrong key column in `item_stock` repository lookup
**Where:** `app/repository/item_stock.repository.go`

**Problem:** `GetByID` filters `WHERE id = ?`, but `item_stock` PK is `item_id`.

**Impact:** Lookup fails with `no such column: id`.

**Fix snippet:**
```sql
WHERE item_id = ?
```

### 2.4 `recipe_components` list orders by nonexistent `occurred_at`
**Where:** `app/repository/recipe_component.repository.go`

**Problem:** `GetAll` uses `ORDER BY occurred_at`, but table has `created_at`.

**Impact:** Query fails at runtime.

**Fix snippet:**
```sql
ORDER BY created_at DESC
```

---

## 3) Solo syntax bugs

### 3.1 Returning `&id` where function return type is `int64`
**Where:**
- `app/repository/event.repository.go`
- `app/repository/inventory_movement.repository.go`
- `app/repository/recipe.repository.go`
- `app/repository/recipe_component.repository.go`

**Problem:** `return (&id, nil)` returns `*int64` for methods declared `(int64, error)`.

**Impact:** Code does not compile.

**Fix snippet:**
```go
return id, nil
```

### 3.2 Undefined variable `eventID` used in `GetAll` methods
**Where:**
- `app/repository/event.repository.go`
- `app/repository/inventory_movement.repository.go`
- `app/repository/recipe.repository.go`
- `app/repository/recipe_component.repository.go`

**Problem:** `Query(query, eventID)` is used in methods that do not declare `eventID` and do not need args.

**Impact:** Code does not compile.

**Fix snippet:**
```go
rows, err := r.db.Conn.Query(query)
```

### 3.3 `PurchaseLineRepository.GetAllByItemID` uses undefined `itemID`
**Where:** `app/repository/purchase_line.repository.go`

**Problem:** Method argument is `itemId`, but query uses `itemID`.

**Impact:** Code does not compile.

**Fix snippet:**
```go
rows, err := r.db.Conn.Query(query, itemId)
```

### 3.4 `ItemStockRepository.GetAll` has invalid receiver/signature/body
**Where:** `app/repository/item_stock.repository.go`

**Problems:**
- Receiver is `*ItemRepository`, not `*ItemStockRepository`
- Return type is `[]*model.Item` but function builds `[]*model.ItemStock`
- Loop uses `QueryRow(..., itemID)` with undefined `itemID` instead of scanning `rows`

**Impact:** Code does not compile.

**Fix snippet:**
```go
func (r *ItemStockRepository) GetAll() ([]*model.ItemStock, error) {
    query := `
        SELECT item_id, quantity, average_unit_cost, updated_at
        FROM item_stock
        ORDER BY updated_at DESC
    `

    rows, err := r.db.Conn.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    stos := []*model.ItemStock{}
    for rows.Next() {
        sto := &model.ItemStock{}
        if err := rows.Scan(&sto.ItemID, &sto.Quantity, &sto.AverageUnitCost, &sto.UpdatedAt); err != nil {
            return nil, err
        }
        stos = append(stos, sto)
    }
    return stos, rows.Err()
}
```

### 3.5 `RecipeComponentRepository.Create` has invalid `INSERT` column/value count
**Where:** `app/repository/recipe_component.repository.go`

**Problem:** INSERT lists 5 columns `(id, recipe_id, item_id, quantity, created_at)` but only 3 placeholders.

**Impact:** SQL statement fails at runtime.

**Fix snippet:**
```go
query := `
    INSERT INTO recipe_components (recipe_id, item_id, quantity)
    VALUES (?, ?, ?)
`
```

---

## 4) Solo functional bugs

### 4.1 `ItemRepository` scan target count mismatch
**Where:** `app/repository/item.repository.go`

**Problem:** `SELECT` includes `created_at`, but `Scan` in `GetByID` and `GetAll` does not read `CreatedAt`.

**Impact:** Runtime error (`sql: expected N destination arguments in Scan`).

**Fix snippet:**
```go
err := r.db.Conn.QueryRow(query, id).Scan(
    &itm.ID,
    &itm.Name,
    &itm.Unit,
    &itm.Sellable,
    &itm.Purchasable,
    &itm.Producible,
    &itm.DefaultSalePrice,
    &itm.CreatedAt,
)
```

### 4.2 `ItemConversionRepository` scan target count mismatch
**Where:** `app/repository/item_conversion.model.go`

**Problem:** Queries select `id, from_item_id, to_item_id, created_at`, but scan attempts `Factor` and `CreatedAt` (5 targets).

**Impact:** Runtime scan errors in `GetByID`, `GetAll`, `GetAllByFromID`, `GetAllByToID`.

**Fix snippet:**
```sql
SELECT id, from_item_id, to_item_id, factor, created_at
FROM item_conversions
```

---

## Suggested execution order for fixes
1. Fix compile blockers (Section 3).
2. Fix column-name mismatches (Section 2).
3. Fix scan mismatches and event defaulting behavior (Sections 4 and 1).
4. Add repository-level tests for each table/repo pair to prevent SQL/model drift.
