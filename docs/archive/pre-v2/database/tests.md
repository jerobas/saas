# Business Logic Test Plan (Non-bug-focused)

## 1) `items` and `item_stock`

- Insert item with `sellable=1` and `default_sale_price=NULL` → must fail.
- Insert item with `sellable=0` and non-NULL `default_sale_price` → must fail.
- Update existing item to break sellable/price rule → must fail.
- Delete an item → must fail (`trg_items_no_delete`).
- Insert item and assert exactly one `item_stock` row auto-created.
- Delete corresponding `item_stock` row → must fail (`trg_item_stock_no_delete`).
- Attempt to set `item_stock.quantity < 0` directly → must fail table CHECK.
- Attempt to set `item_stock.average_unit_cost < 0` directly → must fail table CHECK.

## 2) `events`

- Insert event with invalid `event_type` → must fail CHECK.
- Insert event with invalid `status` → must fail CHECK.
- Insert event with initial status `POSTED` or `CANCELLED` → must fail (must start in `DRAFT`).
- Update non-DRAFT event (e.g., POSTED) → must fail.
- Delete any event (DRAFT or POSTED) → must fail.
- POST SALE without `sale_lines` → must fail.
- POST PURCHASE without `purchase_lines` → must fail.
- POST PRODUCTION/CONVERSION/ADJUSTMENT without line tables → verify expected behavior (currently allowed by line-check trigger).

## 3) `sale_lines`

- Insert sale_line with missing event_id → must fail.
- Insert sale_line for non-SALE event → must fail.
- Insert sale_line with non-sellable item → must fail.
- Insert sale_line with `quantity <= 0` → must fail.
- Insert sale_line with `unit_price < 0` → must fail.
- Insert/update/delete sale_line when event is POSTED → must fail.
- Insert sale_line in valid DRAFT SALE event → must pass.

## 4) `purchase_lines`

- Insert purchase_line with missing event_id → must fail.
- Insert purchase_line for non-PURCHASE event → must fail.
- Insert purchase_line with non-purchasable item → must fail.
- Insert purchase_line with `quantity <= 0` → must fail.
- Insert purchase_line with `unit_cost < 0` or NULL → must fail.
- Insert/update/delete purchase_line when event is POSTED → must fail.
- Insert purchase_line in valid DRAFT PURCHASE event → must pass.

## 5) `inventory_movements` core validation

- Insert movement for missing event_id → must fail.
- Insert movement when event is POSTED → must fail.
- Update/delete movement when event is POSTED → must fail.
- Insert PURCHASE movement with `direction='OUT'` → must fail.
- Insert SALE movement with `direction='IN'` → must fail.
- Insert movement with `quantity <= 0` → must fail.
- Insert movement with invalid `direction` value → must fail CHECK.
- Insert `IN` movement with `unit_cost=NULL` → must fail.
- Insert both `IN` and `OUT` for same `item_id` in same event → must fail.

## 6) `inventory_movements` origin tracking

- Insert movement with non-existent `origin_movement_id` → must fail.
- Set origin to movement with direction OUT → must fail.
- Set origin with different item_id → must fail.
- Set `origin_movement_id` but new movement direction IN → must fail.
- Valid case: OUT movement consuming from prior IN movement of same item → must pass.

## 7) Posting and stock projection

- POST event containing movement item with no `item_stock` row (manually corrupted DB) → must fail integrity guard.
- POST event with OUT movement where `average_unit_cost IS NULL` → must fail OUT cost guard.
- POST event where resulting stock would be negative → must fail.
- POST event with valid IN movement should update stock qty and average cost correctly.
- POST event with valid OUT movement should stamp movement.unit_cost to current average.
- POST mixed IN/OUT event and verify moving average arithmetic.
- Post event driving resulting quantity to zero and verify `average_unit_cost` becomes NULL.

## 8) Auxiliary tables

### `entities`
- Insert entity with NULL name → must fail.
- Validate index-assisted lookup by name.

### `recipes` / `recipe_components`
- Insert recipe with non-producible output item → must fail.
- Update recipe output to non-producible item → must fail.
- Insert component with duplicate `(recipe_id, item_id)` pair → must fail UNIQUE.
- Insert component with quantity <= 0 → must fail.

### `item_conversions`
- Insert conversion with same from/to item → must fail CHECK.
- Insert conversion with `factor <= 0` → must fail CHECK.
- Insert duplicate from/to pair → must fail UNIQUE.

## 9) End-to-end scenarios

- Purchase -> post -> stock increases and average cost set.
- Sale -> post -> stock decreases and OUT movement stamped with pre-sale average.
- Production event consuming ingredients and producing output item in one posting.
- Conversion event moving quantity across units/items using configured factor.
- Cancellation path: DRAFT -> CANCELLED and ensure no stock impact.
