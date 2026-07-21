# ADR 0003: Document ledger, valuation, and projections

- Status: Accepted
- Date: 2026-07-13

## Context

The current schema independently stores purchase/sale lines, inventory
movements, and stock. Nothing guarantees their item, quantity, or value agrees.
Trigger-based posting also hides a large evolving algorithm inside SQL.

## Decision

`stock_documents` and `stock_document_lines` are the only historical stock
ledger. Separate purchase lines, sale lines, and manually authored inventory
movements are removed.

A line has one item, positive canonical quantity, explicit `IN`/`OUT`
direction, inventory value, optional commercial total, and historical unit
snapshot. Document kind constrains its line shape:

- purchase: inbound lines;
- sale: outbound lines;
- production: outbound inputs and exactly one inbound output;
- adjustment: reason-constrained directions;
- reversal: exact inverse of an eligible target.

Inventory uses perpetual pooled weighted-average valuation. The projection
stores total canonical quantity and total inventory value, never a rounded
average-unit-cost column.

For partial outflow, proportional value is rounded half-up to the nearest
microcurrency unit. Full depletion consumes the exact remaining value.
Same-item outflow is calculated in aggregate and distributed by deterministic
largest remainder so line totals reconcile exactly.

`inventory_balances` is a rebuildable projection maintained in the same posting
transaction. It has no general-purpose mutation repository.

Foreign keys, type checks, nonnegative values, uniqueness, and immutable rows
remain SQLite responsibilities. Posting, valuation, and projection updates are
explicit Go transaction responsibilities.

Adjustment direction follows its reason: opening balance and free stock are
inbound; waste, expiry, damage, and internal sample are outbound; physical count
and documented correction may contain both directions across different items.
A physical-count line snapshots expected and observed quantity in dedicated
adjustment metadata while its canonical ledger line stores the resulting delta.

## Consequences

- Commercial revenue/cost and inventory value cannot silently disagree through
  duplicate lines.
- Ledger replay becomes an exact corruption check rather than an approximate
  comparison.
- Changing a projection algorithm cannot rewrite historical line values.
- Posting integration and property tests become critical application tests.
- Dashboard/report tables are read models, not additional stock truth.
