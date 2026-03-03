```
chat

you'll only consider the folllowing files for this next task:
"app/database/schemas/*"
"app/database/model/*"
"app/repository/*"
"docs/app/database/database.md"
"docs/app/database/tests.md"
"docs/app/service/service_proposal.txt"

take database.md as the specifications for the architecture of this app's db

analyze service_proposal.txt and, by reading the repos, models, schema definitions and database.md, do the following:

1 - create a file inside docs directory called "codex_analysis.md"
2 - create section 1, and add there any functional mistakes my proposal for services may have (eg: "you tried to create A before B but this is impossible")
3 - create section 2, and answer any "question" I did in my proposal (eg: lines starting with ?maybe, the "comment" asking where should conversions be
4 - create section 3, and add there any feature suggestions you may have based on the purpose you presume this app may achieve (eg: "service A requires repository B to have get_all_by_C") 
5 - create section 4, and add there any structural suggestions you may have (eg: "service A and B should actually be the same service")
6 - create section 5, add anything else you want to suggest that may not fit in the previous sections
```

# Codex Analysis of `service_proposal.txt`

This analysis uses `docs/app/database/database.md` as the source of truth for the DB architecture, and cross-checks with current schemas/repositories.

## 1) Functional mistakes / feasibility gaps in the proposed services

1. **`eventRepository.Delete` exists, but `events` cannot be deleted in DB.**
   Any service flow that depends on deleting events will fail at runtime because `trg_events_no_delete` always raises an error.

2. **`recipe service -> delete recipe (future)` as hard delete is currently unsafe/impossible in many cases.**
   There is no soft-delete column in `recipes`, and hard delete can fail if components exist (FK). The proposal says "soft delete", but schema/repository do not support it yet.

3. **`item service -> delete item (future)` is incompatible with current schema.**
   Items are protected by trigger (`trg_items_no_delete`) and there is no soft-delete field.

4. **All "get all X by counterparty" flows need an event-type filter in the repository call, but no such method exists yet.**
   `EventRepository.GetAllByCounterpartyID` does not filter by `event_type`, so purchase/sale services must add filtering in service code or create dedicated repo methods.

5. **`produce`/`adjust`/`conversion` without explicit posting strategy can leave stock unchanged forever.**
   Stock is updated only on `DRAFT -> POSTED`; creating movements alone does not affect stock.

6. **Proposal assumes update flows for lines/movements; repositories currently expose no `Update(...)` methods.**
   Future update operations are conceptually valid while DRAFT, but require new repo methods.

7. **Production event modeling has a hard DB limitation per item: same event cannot have IN and OUT for the same item.**
   If a production recipe ever consumes and produces the same item ID in one event (rework/return patterns), insert/update of movements fails.

## 2) Answers to explicit proposal questions/comments

1. **`?maybe post event` (purchase/sale/production/adjust):**
   Recommended default: **post inside the same application use-case** (transactional "create draft data + post").
   Optional design: expose two endpoints/methods (`create_draft`, `post_event`) for workflows that need manual review.

2. **`?maybe get inventory movements by event id` (purchase/sale get by id):**
   Yes—recommended. These are important for auditability and troubleshooting; repository support already exists via `InventoryMovementRepository.GetAllByEventID`.

3. **`?maybe get all purchases by item id`:**
   Yes—valid and useful. You already have `PurchaseLineRepository.GetAllByItemID`; pair it with event lookup and ensure final results are filtered to `event_type='PURCHASE'`.

4. **`# should conversion be here, in its own service, or inside inventory service`:**
   Best fit: **own `ConversionService`** (domain-specific rules: factors, from/to item semantics, optional rounding policy), but it can share lower-level posting utilities with Inventory/Event services.

5. **`?maybe get all components` (recipe service):**
   Keep as internal/admin endpoint only. For business flows, prefer scoped queries (`by recipe`, `by item`) to avoid high-volume reads.

## 3) Feature suggestions based on current architecture

1. Add `EventRepository.GetAllByType(...)` and `GetAllByTypeAndCounterparty(...)` to eliminate in-memory filtering.
2. Add `EventRepository.Post(eventID)` and `Cancel(eventID)` methods to make lifecycle transitions explicit and centralized.
3. Add `GetByEventIDs(...)` helpers for purchase/sale line repositories to avoid N+1 joins in service code.
4. Add an `Adjustment` read path (`get adjustment by id`) even if list endpoints are postponed; debugging adjustments usually needs full movement details.
5. Add an "event details" aggregate query (event + commercial lines + movements) for purchase/sale/production auditing.
6. Add explicit production input source support (e.g., optional origin movements) if traceability/expiry is relevant.

## 4) Structural suggestions

1. Introduce an **Event Application Service layer** to centralize shared workflow:
   `create_event_draft -> add lines/movements -> validate -> post/cancel`.
   Then Purchase/Sale/Production/Adjustment/Conversion become thin orchestrators.

2. Keep **RecipeService** separate from **ProductionService**:
   recipe = formula master data; production = stock-changing transaction. This separation matches current schema responsibilities.

3. Split **ItemService** and **CatalogQueryService** if read volume grows:
   command side (create/update flags) vs query side (filtered listings).

4. Treat conversion as a first-class service, not just a sub-case of inventory adjustments, because it has dedicated master data (`item_conversions`) and potential rounding/business rules.

## 5) Additional suggestions

1. Define a clear cancellation policy now:
   - Is `DRAFT -> CANCELLED` allowed? (currently yes)
   - Is `POSTED -> CANCELLED` allowed? (currently blocked by trigger because only DRAFT rows are mutable)
   - If posted cancellation is needed, implement reversal events instead of status mutation.

2. Add service-level idempotency for posting (`post event` called twice) and for command retries.

3. Standardize monetary/cost precision policy (INTEGER suggests minor units like cents) and document conversion/rounding for factors and moving-average calculations.

4. Consider filtering `get all stocks` by `quantity > 0` directly in repository (today `ItemStockRepository.GetAll` returns all rows).

5. Add tests for cross-table invariants in service flows (not only schema triggers), especially for production/conversion orchestration.
