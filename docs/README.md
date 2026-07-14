# Sweeters documentation

This directory describes the local desktop product that is being built on the
`refactor/local-v2` branch.

## Authority

The sources below are authoritative, in this order:

1. Accepted product decisions in [`decisions/`](decisions/README.md).
2. Domain invariants in [`domain/invariants.md`](domain/invariants.md).
3. The target data model in
   [`architecture/data-model.md`](architecture/data-model.md).
4. Embedded SQLite migrations once the target model is implemented.
5. Generated queries, repositories, application use cases, Wails handlers, and
   frontend features, in that order.

An implemented higher layer must conform to every implemented lower layer. If a
product decision changes, its ADR and invariants change first, followed by a
new migration and then every dependent layer.

## Current documents

- [`architecture/overview.md`](architecture/overview.md): product boundary,
  dependency direction, and runtime architecture.
- [`architecture/data-model.md`](architecture/data-model.md): proposed V2 ERD
  and table responsibilities.
- [`domain/glossary.md`](domain/glossary.md): canonical vocabulary.
- [`domain/invariants.md`](domain/invariants.md): testable business rules.
- [`domain/inventory-ledger.md`](domain/inventory-ledger.md): posting, costing,
  lot allocation, and replay rules.
- [`domain/use-cases.md`](domain/use-cases.md): supported and deferred workflows.
- [`decisions/`](decisions/README.md): accepted architectural decisions.

The data model is a Phase 1 contract, not a description of the current seven
experimental migrations. Those migrations will be replaced during the database
phase after the toolchain upgrade.

## Historical material

Pre-V2 notes are retained under [`archive/pre-v2/`](archive/pre-v2/README.md)
for context only. They are not specifications and must not be used to resolve a
conflict with the documents above.
