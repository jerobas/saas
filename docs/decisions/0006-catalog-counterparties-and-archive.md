# ADR 0006: Catalog, counterparties, and archival

- Status: Accepted
- Date: 2026-07-13

## Context

The existing schema already points toward unified items, but the UI and
compatibility services still separate ingredients, products, and batches. Item
units are free text, default sale price is unnecessarily mandatory, the
minimum-stock input is discarded, and no master data has a complete lifecycle.
Counterparties are stored as ambiguous `entities`.

## Decision

One `items` catalog represents every physical stock item. Ingredient and product
screens are filtered views. An item can be purchasable, producible, sellable, or
any useful combination, and an active item has at least one capability.

Items are created explicitly. A recipe never invents an output item. Optional
catalog fields include SKU, description, default sale price, and reorder level.
Default sale price is allowed only for a sellable item but is not required.
Actual document totals always win.

Base unit becomes immutable after the item is referenced by stock or a recipe
revision. Names are normalized and case-insensitively unique, including archived
names, to keep history unambiguous.

`counterparties` replace `entities`. A counterparty may have supplier and/or
customer roles. Purchases and sales may omit a counterparty for cash/anonymous
activity; when present, its active role must match the document.
An active counterparty has at least one role, while names are allowed to repeat.

Items, counterparties, recipes, and user-defined packaging use one nullable
`archived_at` timestamp. Archived rows remain readable through historical
documents but cannot be used for new work. Unarchive reruns current uniqueness
and validity checks. Seeded units and historical rows are immutable rather than
archivable.

## Consequences

- Separate product and ingredient persistence/services are removed.
- Flags control future eligibility, not historical interpretation.
- A sellable item can have variable price and no misleading default.
- Reorder threshold becomes real persisted catalog data.
- Hard deletion is not a public use case for referenced master data.
- Real suppliers and customers may share names; counterparty name is not a
  unique identity key.
