# ADR 0001: Local desktop and bottom-up authority

- Status: Accepted
- Date: 2026-07-13

## Context

The repository contains a Wails desktop application whose useful business data
already fits a local SQLite model. Previous work mixed historical proposals,
database triggers, repository DTOs, compatibility services, fake frontend data,
and remote product flows. Those layers disagree about the domain.

## Decision

Sweeters V2 is a local-first, single-user desktop modular monolith using the
existing Go, Wails, SQLite, React, Vite, and Tailwind stack.

Initial inventory scope is one business and one stock location. Remote
synchronization, multi-user collaboration, fiscal accounting, and multi-location
stock are excluded.

Authority is:

```text
accepted ADRs and product invariants
    -> SQLite migrations
        -> generated queries and aggregate stores
            -> application commands and queries
                -> Wails handlers and DTOs
                    -> React features
```

The existing database is not assumed correct merely because it is lower in the
current code. Phase 1 corrects the intended contract first. Once a migration
implements an accepted decision, higher layers conform to it. A later domain
change starts with another ADR and migration rather than a compatibility fix in
a service or page.

Complex cross-row workflows live in explicit Go transactions. SQLite retains
structural constraints and immutability guards.

## Consequences

- Existing service and frontend compatibility APIs may be deleted instead of
  preserved.
- The seven experimental migrations can be rebaselined before production.
- Repositories are aggregate-oriented and cannot expose unrestricted table CRUD.
- Every vertical feature is developed schema/query first and UI last.
- Tests and documentation are completion requirements for every layer.
