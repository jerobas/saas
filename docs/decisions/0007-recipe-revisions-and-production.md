# ADR 0007: Recipe revisions and production

- Status: Accepted
- Date: 2026-07-13

## Context

Recipes are currently mutable rows whose components are deleted and recreated.
Production history cannot identify which formula it used, the service silently
creates an output product, and recipe cost is not implemented.

## Decision

A recipe is a stable name and fixed output item. Changing the output item means
creating another recipe.

Recipe content lives in immutable numbered revisions. Creating a recipe writes
revision 1 and its components atomically. Editing publishes N+1; restoring old
content copies it into another new revision. Historical revisions are never
edited or made current by pointer manipulation.

A revision contains positive standard yield, instructions, preparation time,
optional estimated direct cost, and at least one unique positive component.
The output item must be active and producible when publishing. Components must
be active and cannot directly include the output item in V2.

A production run references exactly one revision. The selected target yield
scales expected inputs for preview, but the posted document records actual input
quantities and actual output yield. Actual lines are inventory truth.

V2 production consumes one or more input lots and creates exactly one output
line and lot matching the recipe output. Multiple outputs and by-products are
deferred.

Output inventory value is the exact sum of consumed input valuation plus an
explicitly entered direct production cost. Forecast preparation time, labor,
energy, margin, or overhead is not silently capitalized.

Exact production reversal is subject to ADR 0005, including availability of the
created output lot and restoration of source allocations.

## Consequences

- Production history remains reproducible after recipe edits.
- Recipe cost estimates can use current inventory value without changing
  historical runs.
- Yield variance and material variance become reportable.
- The user, not a naming convention, controls catalog outputs.
- Automatic recursive recipe expansion and by-product accounting remain future
  decisions.
