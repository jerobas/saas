# V2 use cases

Use cases are task-oriented application commands and queries. The frontend
calls one use case for one user action; it does not assemble store operations.

## Settings and units

- Initialize local business settings on first run.
- Read and update business name, locale, timezone, and planning defaults.
- Select currency before the first stock posting.
- List seeded measurement units.
- Create, update, archive, and restore item-specific packaging definitions.

## Catalog

- Create an item with base unit and capabilities.
- Read an item and list items by capability, stock state, or archive state.
- Update catalog metadata, optional default price, and reorder level.
- Change base unit only while the item has no active packaging, recipe-revision,
  or ledger references; reconfigure incompatible archived packaging before
  restoring it.
- Archive and restore an item.

“Ingredients” and “Products” remain UI views over these catalog queries.

## Counterparties

- Create and update a supplier, customer, or dual-role counterparty.
- List by role and archive state.
- Archive and restore a counterparty.
- Read historical documents for a counterparty.

## Purchases

- Preview exact unit conversion, line totals, and inbound lots.
- Post a purchase atomically.
- Read purchase detail and list/filter purchases.
- Exactly reverse an eligible latest purchase.

Each inbound line represents one lot. The user splits lines when supplier lot
or expiry differs.

## Inventory

- Read current quantity, value, derived average, and reorder state by item.
- List available, expired, and depleted lots.
- Read an item's immutable ledger history.
- Preview automatic FEFO allocation.
- Post an opening balance.
- Post positive or negative reasoned adjustments.
- Record a physical count and post its calculated difference.
- Reconcile projections against ledger replay.

## Recipes

- Create a recipe and revision 1 atomically for an existing output item.
- Read the current revision and revision history.
- Publish a new immutable revision.
- Copy an old revision into a new current revision.
- Estimate material availability and current weighted-average cost.
- Archive and restore a recipe.

## Production

- Preview a production run for a target yield.
- Show expected inputs, shortages, proposed FEFO lots, and estimated value.
- Adjust actual inputs, actual yield, and explicit direct cost before posting.
- Post production atomically, consuming input lots and creating one output lot.
- Read production detail.
- Exactly reverse an eligible latest production run.

V2 production has exactly one output item and no by-products.

## Sales

- Preview availability, FEFO allocations, revenue, and estimated cost of goods.
- Post a sale atomically.
- Read sale detail and list/filter sales.
- Exactly reverse an eligible latest data-entry sale.

A physical customer return is not the same as correcting a data-entry error and
does not automatically restore food to usable stock.

## Backup and recovery

- Export a consistent local snapshot.
- Validate a candidate backup without modifying the active database.
- Make an automatic safety backup and atomically replace the active database.
- Restart the application after import and run integrity/reconciliation checks.

## Reports built after operational flows

- Current inventory quantity and valuation.
- Expiring and expired lots.
- Purchase history and spend.
- Production yield and material variance.
- Sales revenue, cost of goods, and gross margin.
- Ledger and correction audit trail.

## Explicitly deferred

- Durable/autosaved drafts.
- Multiple stock locations and transfers.
- Multiple concurrent users or remote synchronization.
- Multi-currency and foreign exchange.
- Fiscal/tax invoices or general-ledger accounting.
- Partial supplier and customer return workflows.
- Automatic use of expired stock.
- Multiple production outputs/by-products.
- Cross-dimensional density conversions.
- Automatic labor, energy, or overhead capitalization.
- Made-to-order negative stock; production must post before sale.
