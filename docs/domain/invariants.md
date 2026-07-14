# Domain invariants

Invariant IDs are stable references for migrations, tests, application errors,
and review checklists. SQLite is the final authority for structural rules;
cross-row algorithms are enforced in the posting transaction and verified with
real-SQLite integration and replay tests.

## Settings and time

| ID | Rule | Primary enforcement |
|---|---|---|
| SET-001 | Exactly one settings row exists. | SQLite |
| SET-002 | The database has one ISO currency and one IANA business timezone. | SQLite + application |
| SET-003 | Currency cannot change after the first stock document posts. | SQLite |
| SET-004 | UTC instants and date-only business values are stored separately. | SQLite types + application |
| SET-005 | A document may use an earlier business date, but valuation always follows posting sequence. | Application transaction |

## Catalog and units

| ID | Rule | Primary enforcement |
|---|---|---|
| CAT-001 | Every stocked physical thing is one `item`; ingredient and product are views, not tables. | Schema design |
| CAT-002 | An active item has at least one purchasable, producible, or sellable capability. | SQLite |
| CAT-003 | Capability combinations are valid; historical documents do not change when capabilities change. | Application + immutability |
| CAT-004 | Default sale price is optional; when present it is nonnegative and the item is sellable. | SQLite |
| CAT-005 | Item names are trimmed and case-insensitively unique; archived names remain reserved. | SQLite + application |
| CAT-006 | Base unit cannot change while the item has active packaging or after it appears in a recipe revision or stock document. Archived incompatible packaging must be reconfigured before restoration. | SQLite |
| CAT-007 | Archived catalog data is readable historically but unavailable for new posting. | Application transaction |
| CAT-008 | Optional item SKUs use the documented normalized key and remain unique across active and archived items. | SQLite + application |
| CPY-001 | An active counterparty has at least one supplier or customer role; names need not be unique. | Application + SQLite |
| CPY-002 | Removing a role affects only future eligibility and never rewrites historical documents. | Application + immutability |
| UNIT-001 | Quantities and conversion factors are never stored as floating point. | SQLite |
| UNIT-002 | Canonical atomic quantity is milligrams, microlitres, or thousandths of a count item. | ADR + typed application values |
| UNIT-003 | Packaging conversion is a positive reduced rational directly to an item's base quantity. | SQLite + application |
| UNIT-004 | A conversion must produce an exact atomic quantity; stock is never silently rounded. | Application transaction |
| UNIT-005 | A posted line and recipe component preserve the entered unit and conversion snapshot. | SQLite immutability |
| UNIT-006 | Changing units or packaging does not create stock; item transformation is production. | Schema and use-case boundary |

## Documents

| ID | Rule | Primary enforcement |
|---|---|---|
| DOC-001 | The database contains only complete posted documents, not drafts or cancelled placeholders. | Schema design |
| DOC-002 | Every document contains at least one line and every line belongs to exactly one document. | SQLite + posting transaction |
| DOC-003 | A posted document, its lines, lots, allocations, and production metadata are immutable. | SQLite guards |
| DOC-004 | Posting commits document, lines, lots, allocations, and projections atomically. | Application transaction |
| DOC-005 | A posting idempotency key is unique and a retry returns the existing document. | SQLite unique key + application |
| DOC-006 | Posting sequence is monotonic and is the canonical inventory order. | SQLite + transaction serialization |
| DOC-007 | Commercial total and inventory valuation are separate values. | Schema design |
| DOC-008 | Quantities are positive; direction expresses the sign of their stock effect. | SQLite |
| DOC-009 | A document cannot use its own inbound output to satisfy an outbound input. | Application transaction |

## Document kinds

| ID | Rule | Primary enforcement |
|---|---|---|
| PUR-001 | A purchase has one or more `IN` lines for active purchasable items. | Application + SQLite checks |
| PUR-002 | Every purchase line has an explicit commercial total; zero requires a free-stock reason. | Application transaction |
| PUR-003 | A purchase counterparty is optional; when present it must be an active supplier. | Application transaction |
| SAL-001 | A sale has one or more `OUT` lines for active sellable items. | Application + SQLite checks |
| SAL-002 | Every sale line has an explicit commercial total; zero requires a promotion/sample reason. | Application transaction |
| SAL-003 | A sale counterparty is optional; when present it must be an active customer. | Application transaction |
| ADJ-001 | An adjustment has a typed reason and each line direction is valid for that reason. | SQLite + application |
| ADJ-002 | A negative adjustment uses current weighted-average value and complete lot allocation. | Application transaction |
| ADJ-003 | A positive adjustment into zero stock has explicit value; zero value requires `FREE_STOCK`. | Application transaction |
| ADJ-004 | A physical-count line preserves expected and observed quantity and posts only their difference. | SQLite + application transaction |

## Inventory valuation and projection

| ID | Rule | Primary enforcement |
|---|---|---|
| INV-001 | Commercial money uses currency minor units; inventory value uses integer microcurrency units. | SQLite + typed values |
| INV-002 | Balance quantity and inventory value are never negative. | SQLite + application |
| INV-003 | Zero quantity implies zero inventory value. | SQLite |
| INV-004 | Zero value enters inventory only through explicitly reasoned free stock; production and exact reversal may propagate that value without inventing cost. | Application transaction |
| INV-005 | Outbound value is frozen at posting from pooled weighted-average quantity and value. | Application transaction |
| INV-006 | Consuming all remaining quantity consumes all remaining inventory value exactly. | Application transaction |
| INV-007 | Same-item outflow is costed in aggregate; deterministic remainder allocation makes line values sum exactly. | Application transaction |
| INV-008 | Replaying immutable lines in posting order exactly rebuilds every balance. | Integration/replay tests |
| INV-009 | Balance rows have no general-purpose public update operation. | Store boundary |
| INV-010 | Exact integer values cross Wails as decimal strings, never unsafe JavaScript numbers. | Presentation contracts |

## Lots and expiry

| ID | Rule | Primary enforcement |
|---|---|---|
| LOT-001 | Every non-reversal inbound line creates exactly one same-item lot. | SQLite + application |
| LOT-002 | Users create separate inbound lines when lot code or expiry differs. | Use-case validation |
| LOT-003 | Every outbound line is fully allocated across one or more same-item lots. | Application + reconciliation tests |
| LOT-004 | A lot's total consumption minus restoration never exceeds its initial quantity. | Application transaction |
| LOT-005 | Default allocation is FEFO, then inbound posting sequence and lot ID; no-expiry lots sort last. | Application transaction |
| LOT-006 | Expiry is an inclusive date, not an instant. | SQLite representation |
| LOT-007 | Expired lots remain in physical and financial stock until an adjustment removes them. | Query and adjustment policy |
| LOT-008 | Expired lots cannot be allocated to a new sale or production run. | Application transaction |
| LOT-009 | Sale and production overrides use only nonexpired available lots; a reasoned negative adjustment may deliberately consume expired stock. Every selection is frozen at posting. | Application transaction |
| LOT-010 | Lot availability can be rebuilt exactly from lot sources and allocation effects. | Integration/replay tests |

## Reversals

| ID | Rule | Primary enforcement |
|---|---|---|
| REV-001 | A posted document is corrected by a new linked document, never by mutation. | SQLite immutability |
| REV-002 | Exact reversal is allowed only when the target is the latest stock-affecting document for every affected item. | Application transaction |
| REV-003 | Lots created by the target must have no downstream consumption that the reversal cannot restore exactly. | Application transaction |
| REV-004 | Reversal lines exactly invert target quantity and inventory value and reference their target lines. | SQLite + application |
| REV-005 | Outbound allocations are restored to their original lots by linked restoration entries. | Application transaction |
| REV-006 | A document can be exactly reversed at most once; a reversal cannot be reversed. | SQLite unique/check + application |
| REV-007 | When exact reversal is ineligible, a current-period compensating workflow is required. | Use-case boundary |

## Recipes and production

| ID | Rule | Primary enforcement |
|---|---|---|
| REC-001 | A recipe has one fixed active producible output item. Changing output means a new recipe. | Application + SQLite FK |
| REC-002 | A recipe revision is immutable, positively numbered, and contains at least one unique component. | SQLite + application |
| REC-003 | Standard yield and every component quantity are positive exact canonical quantities. | SQLite |
| REC-004 | A revision cannot directly consume its own output item. | SQLite + application |
| REC-005 | Publishing an edit creates revision N+1 atomically; historical revisions are never repointed or edited. | Application transaction |
| PRO-001 | Production references exactly one recipe revision and has one or more `OUT` inputs plus exactly one `IN` output. | SQLite + application |
| PRO-002 | The output item matches the recipe and inputs cannot contain that output item in V2. | Application transaction |
| PRO-003 | Posted actual consumption and actual yield, not the recipe estimate, are stock truth. | Ledger design |
| PRO-004 | Output inventory value equals actual consumed value plus explicitly entered direct production cost. | Application transaction |
| PRO-005 | Forecast labor or overhead is never silently capitalized into stock. | Use-case boundary |

## Archival and deletion

| ID | Rule | Primary enforcement |
|---|---|---|
| ARC-001 | Items, counterparties, recipes, and user-created packaging are archived, not hard-deleted. | Store boundary + FK policy |
| ARC-002 | Unarchive reruns uniqueness and validity checks. | Application transaction |
| ARC-003 | Seeded measurement units and all immutable historical records cannot be archived or deleted. | SQLite + store boundary |
