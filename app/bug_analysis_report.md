# Bug Analysis Report (`app/database/schemas`, `app/model`, `app/repository`)

This report was refreshed against the current code in the requested directories.
Solved items from the previous version were removed (except 1.1, as requested), and newly found issues were added.

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
    evt.OcurredAt,
)
```

---

## 2) Integration syntax bugs

### 2.1 `PurchaseLine` JSON tag mismatch: model uses `unit_price`, SQL/repo use `unit_cost`
**Where:**
- Schema: `app/database/schemas/006_purchases.sql`
- Model: `app/model/purchase_line.model.go`
- Repository: `app/repository/purchase_line.repository.go`

**Problem:** `PurchaseLine` and `PurchaseLineInsertDTO` expose `UnitCost` with JSON tag `unit_price`, while schema/repository use `unit_cost`.

**Impact:** API payloads using `unit_cost` will not bind as expected; inserted `unit_cost` may silently become zero-value.

**Fix snippet:**
```go
// model
UnitCost int `json:"unit_cost"`
```

### 2.2 `Event` JSON tag typo: `ocurred_at` vs `occurred_at`
**Where:**
- Schema: `app/database/schemas/003_events.sql`
- Model: `app/model/event.model.go`
- Repository: `app/repository/event.repository.go`

**Problem:** Model uses `json:"ocurred_at"` (missing `r`), but DB/repository semantics are `occurred_at`.

**Impact:** Inputs using `occurred_at` may not populate the DTO field correctly, leading to incorrect timestamps being stored.

**Fix snippet:**
```go
OcurredAt time.Time `json:"occurred_at"`
```

---

## 3) Solo syntax bugs

No current solo syntax bugs found in the analyzed scope.

---

## 4) Solo functional bugs

No current solo functional bugs found in the analyzed scope.

---

## Suggested execution order for fixes
1. Keep/implement status fallback in `EventRepository.Create` (1.1).
2. Fix model JSON tags to align payload names with DB/repository contracts (2.1, 2.2).
3. Add table/repository contract tests to catch future model/repo/schema drift.
