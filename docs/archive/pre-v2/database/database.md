---

# Inventory & Sales Model – Behavioral Specification (V1 – Moving Average Edition)

---

# 1. Core Philosophy

This system follows five foundational principles:

1. **Items are the physical truth**  
   Anything that can exist in stock is an `item`.

2. **Events represent business actions**  
   Sales, purchases, production, conversions, and adjustments are `events`.

3. **Inventory movements are the immutable ledger**  
   Every physical stock change is recorded as an `inventory_movement`.

4. **`item_stock` is a derived projection**  
   It represents current stock state (quantity + valuation) and is updated only at posting time.

5. **Stock valuation follows Moving Average costing**  
   Each item maintains a running weighted average cost.

---

# 2. Tables & Their Meaning

---

## 2.1 `items`

Represents any physical stock unit.

Examples:

* Flour
* Ganache
* Whole cake
* Cake slice

### Capabilities

Each item may be:

* `sellable`
* `purchasable`
* `producible`

### Rules

* If `sellable = 1` → `default_sale_price` must exist (enforced by triggers on INSERT/UPDATE).
* If `sellable = 0` → `default_sale_price` must be NULL.
* Items cannot be deleted (`trg_items_no_delete`).
* Creating an item automatically creates its corresponding `item_stock` row.

---

## 2.2 `item_stock`

Projection of current stock state.

Fields:

* `item_id` (1:1 with items)
* `quantity`
* `average_unit_cost`
* `updated_at`

### Rules

* Row is auto-created when item is created.
* Row cannot be deleted.
* `quantity >= 0` is enforced by table check.
* `average_unit_cost >= 0` when non-NULL.
* `average_unit_cost` being NULL when quantity is zero is **intended** by posting logic, but not fully enforced by a table-level CHECK.

---

 2.3 `events`

Represents a business action.

### event_type

* `SALE`
* `PURCHASE`
* `PRODUCTION`
* `CONVERSION`
* `ADJUSTMENT`

### Status

* `DRAFT` → Editable, no stock impact
* `POSTED` → Final, stock updated
* `CANCELLED`

### Rules (as currently implemented)

* New events must be inserted as `DRAFT` (`trg_events_no_insert_unless_draft`).
* Only rows with `OLD.status = 'DRAFT'` can be updated (`trg_events_no_update_unless_draft`).
* Events **cannot be deleted at all** (`trg_events_no_delete`).
* Transition to `POSTED` requires lines only for:
  * SALE (`sale_lines` must exist)
  * PURCHASE (`purchase_lines` must exist)
* No explicit trigger enforces valid status transitions beyond the general update block above.

---

## 2.4 `sale_lines`

Commercial representation of a sale.

Fields:

* `event_id`
* `item_id`
* `quantity`
* `unit_price`

Rules:

* Enforced: insert/update/delete allowed only while parent event is `DRAFT` (via status triggers).
* Enforced: `event_id` must exist and be `event_type='SALE'`.
* Enforced: item must be `sellable=1`.
* Enforced: `quantity > 0`, `unit_price >= 0`.

---

## 2.5 `purchase_lines`

Commercial representation of a purchase.

Fields:

* `event_id`
* `item_id`
* `quantity`
* `unit_cost`

Rules:

* Enforced: insert/update/delete allowed only while parent event is `DRAFT` (via status triggers).
* Enforced: `event_id` must exist and be `event_type='PURCHASE'`.
* Enforced: item must be `purchasable=1`.
* Enforced: `quantity > 0`, `unit_cost >= 0`, and `unit_cost` NOT NULL.

---

# 2.6 `inventory_movements` (updated)

The immutable physical ledger.

Each row represents a real stock change.

Fields:

* `event_id`
* `item_id`
* `direction` (`IN` or `OUT`)
* `quantity`
* `unit_cost`
* optional `origin_movement_id` (lot/source tracking support)
* optional `expires_at`

### Rules

* Movements can only be inserted/updated/deleted while event is DRAFT.
* `origin_movement_id` (if provided) must:
  * exist,
  * point to an `IN` movement,
  * have same `item_id`,
  * and current movement must be `OUT`.
* Direction restrictions by event type:
  * PURCHASE → only `IN`
  * SALE → only `OUT`
  * PRODUCTION/CONVERSION/ADJUSTMENT → both allowed
* Same event cannot contain both `IN` and `OUT` for the same `item_id` (`trg_inventory_movements_no_insert_in_out_same_item` / `trg_inventory_movements_no_update_in_out_same_item`).

### Costing Rules

* IN movements must carry non-NULL `unit_cost`.
* OUT movements may have NULL `unit_cost` while DRAFT.
* At POST time, OUT movements are stamped from current `item_stock.average_unit_cost`.

---

## 2.7 Additional domain tables missing in README

### `entities`

Counterparties (suppliers/customers) referenced by `events.counterparty_entity_id`.

### `recipes` and `recipe_components`

Define production formulas:

* `recipes.output_item_id` must be `producible=1` (INSERT/UPDATE triggers).
* `recipe_components` defines ingredient quantities per recipe.

### `item_conversions`

Defines conversion factors between two items (`from_item_id`, `to_item_id`, `factor > 0`) with uniqueness and anti-self-conversion checks.

---

# 3. Moving Average Costing Model

For each item:

* `quantity = Q`
* `average_unit_cost = A`
* Implicit value = `V = Q × A`

At posting time:

* `delta_qty = Σ(IN qty) − Σ(OUT qty)`
* `delta_value = Σ(IN value) − Σ(OUT value)`

New state:

* `new_qty = old_qty + delta_qty`
* `new_value = old_qty × old_avg + delta_value`
* `new_avg = new_value / new_qty` (if new_qty > 0)
* If new_qty = 0 → `average_unit_cost = NULL`

OUT valuation:

* OUT `unit_cost` is stamped at POST time using the current `average_unit_cost`.

---

# 4. Posting Pipeline

When event transitions `DRAFT → POSTED`:

---

## Step 1 – Integrity Check

Fail if any movement references a missing `item_stock`.

---

## Step 2 – OUT Cost Guard

Fail if:

* An OUT movement exists AND
* The corresponding `item_stock.average_unit_cost` is NULL

---

## Step 3 – Negative Stock Validation

For each affected item:

* Compute `delta_qty`
* Fail if `item_stock.quantity + delta_qty < 0`

---

## Step 4 – Stamp OUT Costs

For all OUT movements in the event:

* Set `unit_cost = current item_stock.average_unit_cost`

This freezes valuation snapshot for history.

---

## Step 5 – Update `item_stock` Projection

Using aggregated deltas:

* Update quantity
* Recalculate average cost using moving average formula
* Update timestamp

---

# 5. Draft Philosophy

DRAFT is workspace:

* Movements editable
* Lines editable
* No stock impact
* No cost finalization

Only POSTED changes stock.

---

# 6. Cancellation Model (V1)

* DRAFT → may transition to CANCELLED.
* POSTED → immutable (cannot update due event trigger).
* Events cannot be deleted (regardless of status).

---

# 7. Truth Hierarchy

| Layer                       | Purpose           | Mutable?               |
| --------------------------- | ----------------- | ---------------------- |
| items                       | Catalog           | Yes (except delete)    |
| events (DRAFT)              | Intent            | Yes                    |
| events (POSTED)             | Historical record | No                     |
| sale_lines / purchase_lines | Commercial truth  | Only while DRAFT (intended) |
| inventory_movements         | Physical ledger   | Immutable after POSTED |
| item_stock                  | Projection        | System-managed         |

---

# 8. Guarantees of V1

Intended guarantees:

* No negative stock.
* No double-posting.
* Immutable historical ledger.
* Deterministic moving average valuation.
* OUT cost frozen at posting time.
* Stock derivable from ledger.
* 1:1 mapping between items and stock rows.

---

# 9. Explicit V1 Invariants

1. An event should not IN and OUT the same item (**enforced, but current trigger is over-broad and ignores `item_id`**).
2. IN movements should have unit_cost (**enforced in SQL**).
3. OUT unit_cost is set at POST (enforced in posting trigger).
4. `item_stock.average_unit_cost` is intended to be NULL iff quantity = 0 (partially enforced by logic, not table CHECK).
5. `item_stock` row must exist for every item (auto-created trigger).
6. Only `DRAFT → POSTED` transition should change stock (intended).

---