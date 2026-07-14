# Domain glossary

These terms are canonical in code, schema, UI copy, and tests.

## Catalog

**Item**
A physical thing that can exist in stock. Flour, packaging, ganache, a whole
cake, and a cake slice are all items.

**Capability**
An item's permission to participate in a new workflow: `purchasable`,
`producible`, or `sellable`. An item may have any useful combination.

**Base unit**
The canonical measurement in which an item's stock is stored. V2 supports mass
in grams, volume in millilitres, and count in units.

**Packaging**
An item-specific entry/display measure with a direct exact conversion to the
base unit, such as one 5 kg flour bag. Changing representation is not a stock
movement.

**Archived**
Hidden from normal selection for new work but retained for historical reads.
Archiving is not deletion.

**Counterparty**
A person or organization participating as a supplier, customer, or both.

## Quantities and money

**Canonical quantity**
An integer count of the item's atomic stock unit: milligrams for mass,
microlitres for volume, and thousandths for count. Business quantities are
never floating point.

**Commercial total**
The exact amount charged or paid for a document line, stored in currency minor
units such as centavos. It is authoritative; unit price is derived for display.

**Inventory value**
The financial value carried by stock, stored in microcurrency units so small
ingredient consumption can retain sub-cent value. For BRL, one real is
1,000,000 valuation units and one centavo is 10,000 valuation units.

**Weighted-average valuation**
The financial policy that values an outbound quantity as its proportion of the
item's current total quantity and total inventory value.

## Documents and stock

**Stock document**
One completed, immutable business action: purchase, sale, production,
adjustment, or exact reversal.

**Document line**
The canonical record of a physical quantity and its commercial or inventory
value. A line has one item and an `IN` or `OUT` direction.

**Posting**
Atomically validating and committing a complete stock document. V2 does not
persist incomplete drafts.

**Posting sequence**
The monotonic order in which documents affect availability and valuation. A
backdated business date does not change this order.

**Idempotency key**
A unique client-generated command identifier. Retrying a completed posting
command with the same key returns its existing result.

**Ledger**
All immutable posted document lines in posting order. The ledger is the source
from which current stock can be rebuilt.

**Inventory balance**
A rebuildable projection of an item's current canonical quantity and total
inventory value.

**Lot**
The traceable stock introduced by one normal inbound line, with its own source,
code, and optional expiry date.

**Lot allocation**
An immutable record connecting an outbound line to the lots it physically
consumed, or an exact reversal line to the original allocations it restored.

**FEFO**
First-expire, first-out physical allocation. Nonexpired lots with the earliest
expiry are consumed first; lots without expiry follow dated lots.

**Expiry date**
A business date through which a lot is usable, inclusive. Expired stock remains
in inventory until explicitly removed, but is unavailable to sales and
production.

## Corrections

**Exact reversal**
A new linked document that exactly negates an eligible latest document. It is
allowed only while doing so can restore the prior quantity, value, and lot
state without rewriting later history.

**Compensating document**
A current-period purchase return, customer return, waste, or adjustment used
when later stock activity makes exact reversal impossible. Initial V2 supports
adjustments; dedicated return workflows are deferred.

**Adjustment**
A reasoned stock correction such as opening balance, physical count, waste,
expiry, damage, sample, free stock, or data correction. It is never an unnamed
shortcut for production.

## Recipes and production

**Recipe**
The stable identity and output item of a formula.

**Recipe revision**
An immutable numbered snapshot of standard yield, instructions, preparation
time, and components.

**Production run**
A posted stock document that references one recipe revision, records actual
input consumption, and creates exactly one output line and lot.

**Standard yield**
The expected output quantity of a recipe revision. Actual production yield is
recorded separately and remains authoritative for stock.

## Time

**Business date**
The date the user associates with a document for reporting.

**Posted at**
The UTC instant at which the application committed a document.
