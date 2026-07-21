# ADR 0002: Exact quantities, money, valuation, and units

- Status: Accepted
- Date: 2026-07-13

## Context

The experimental model stores quantities and conversion factors as SQLite
`REAL` while storing rounded integer unit cost. Fractional ingredients and
repeated conversion therefore make stock and value non-reproducible. Storing
only centavos for inventory value would also round small recipe consumption to
zero.

## Decision

Business arithmetic uses checked `int64` values only.

Canonical stock quantity is:

- mass: milligrams, displayed from grams;
- volume: microlitres, displayed from millilitres;
- count: thousandths of an item.

This gives a scale of 1,000 atomic units per displayed base unit. A quantity
that cannot convert exactly to an atomic unit is rejected.

Commercial totals are stored in the configured currency's minor units. BRL
therefore stores prices and totals in centavos.

Inventory valuation is stored separately in microcurrency: 1,000,000 valuation
units per whole currency unit. With BRL, one centavo converts exactly to 10,000
valuation units. The higher precision prevents sub-cent ingredient value from
being lost while preserving exact commercial totals.

Items have one canonical base dimension/unit. Entry and display packaging use
positive reduced rational conversion directly to canonical quantity. There are
no floating factors or conversion chains. A posted line and recipe revision
snapshot the conversion used.

Unit or packaging conversion changes representation only. Converting one
physical item into another is production.

All multiplication and conversion check overflow before writing.

Exact integers cross the Wails boundary as base-10 strings. TypeScript uses
branded string/`bigint` helpers for arithmetic and formatting; authoritative
`int64` values are never routed through an unsafe JavaScript `number`.

## Consequences

- `REAL` disappears from business quantities and valuation.
- Unit price and average cost become derived display values, not authoritative
  stored inputs.
- Purchase and sale lines preserve exact totals even for fractional quantities.
- The current item-to-item conversion table and `CONVERSION` event disappear.
- Wails and TypeScript contracts use decimal strings for exact integer values.
- Changing the atomic precision later requires an ADR and data migration.
