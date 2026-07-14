# Inventory ledger and posting rules

## Representation

Each stock line stores a positive canonical quantity and an explicit `IN` or
`OUT` direction. Signed effects exist only during calculation:

```text
signed quantity = +quantity for IN, -quantity for OUT
signed value    = +value    for IN, -value    for OUT
```

An item's projection is:

```text
balance quantity = sum(signed line quantities in posting order)
balance value    = sum(signed line inventory values in posting order)
```

The projection is cached for fast reads but must equal ledger replay exactly.

## Precision

Canonical quantity uses dimension-specific atomic units:

| Dimension | Display base | Stored atomic unit | Scale |
|---|---:|---:|---:|
| Mass | gram | milligram | 1,000 |
| Volume | millilitre | microlitre | 1,000 |
| Count | unit | thousandth | 1,000 |

Commercial totals use configured currency minor units. Internal inventory
valuation uses 1,000,000 units per whole currency unit. For a two-decimal
currency, one minor unit converts exactly to 10,000 valuation units.

All multiplication uses checked integer arithmetic. A value that would overflow
`int64` is rejected instead of wrapped.

## Weighted-average outbound value

For current quantity `Q`, current value `V`, and aggregate outbound quantity
`q`:

1. Reject when `q > Q`.
2. If `q = Q`, outbound value is exactly `V`.
3. Otherwise calculate `V * q / Q` and round half-up to the nearest inventory
   valuation unit.

All same-item outbound lines in one document are costed as one aggregate.
Their aggregate value is distributed by the largest-remainder method:

1. Assign each line the floor of its proportional share.
2. Rank fractional remainders descending, breaking ties by line order.
3. Assign the remaining valuation units in that order.

The result is deterministic and individual line values sum exactly to the
aggregate outbound value.

## Inbound value

- Purchase: line commercial total converted exactly to valuation units.
- Positive adjustment: existing weighted-average value when stock exists, or
  explicit value when it does not. A deliberate zero uses `FREE_STOCK`.
- Production: sum of actual input values plus explicitly entered direct cost.
- Exact reversal: exact inverse of the target line value.

No independently editable average-unit-cost field exists.

## Physical lot allocation

Financial weighted average and physical allocation are separate:

- weighted average calculates financial outflow;
- FEFO chooses the physical lots consumed.

Eligible lots are ordered by:

1. non-null earliest expiry date;
2. null expiry after every dated lot;
3. inbound posting sequence;
4. lot ID.

The allocation walks that order until the complete outbound quantity is
covered. An optional manual selection can change the physical order but never
the financial valuation and cannot use expired or insufficient lots.

`expires_on` is usable through that business date. Eligibility is checked
against the application's current local business date at posting, not against a
user-supplied backdated occurrence date.

## Posting transaction

Posting executes as one serialized SQLite write transaction:

1. Look up the idempotency key and return the existing result if present.
2. Validate settings, active master data, capabilities, units, and document
   shape.
3. Read current balances and eligible lots inside the transaction.
4. Convert entered quantities exactly to canonical atomic quantities.
5. Validate aggregate and lot availability.
6. Build FEFO or manually selected lot allocations.
7. Calculate and distribute outbound inventory value.
8. Calculate inbound inventory value, including production value transfer.
9. Insert the immutable document, lines, production metadata, lots, and
   allocations.
10. Update balance projections using the same signed quantities and values.
11. Assert no negative balance, zero-quantity value, or overallocated lot.
12. Commit.

A failure at any step rolls back everything.

## Posting time and business date

Posting sequence determines availability and valuation. `occurred_on` is a
reporting date and may be earlier than today, but it never inserts a document
into the middle of valuation history. `posted_at` is the actual UTC commit
instant.

## Exact reversal

Weighted-average documents cannot always be negated after later activity
without rewriting the frozen value of those later documents. Exact reversal is
therefore deliberately narrow.

It is allowed only when the target is the latest stock-affecting document for
every affected item and its lot effects have no non-reversible dependencies.
The reversal copies the target commercial context, creates exact inverse lines,
and restores original lot allocations. It posts at the current sequence.

If those conditions fail, the user must record a current-period compensating
workflow. Historical rows are never edited.

## Reconciliation

A diagnostic rebuild replays documents by posting sequence and independently
recalculates:

- item quantity and inventory value;
- lot initial quantity, consumption, restoration, and remaining quantity;
- links between documents, reversal lines, and allocations.

Any difference from stored projections is a corruption error, not a rounding
tolerance.
