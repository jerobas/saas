# Sweeters documentation

This directory describes the local desktop product that is being built on the
`refactor/local-v2` branch.

## Authority

The sources below are authoritative, in this order:

1. Accepted product decisions in [`decisions/`](decisions/README.md).
2. Domain invariants in [`domain/invariants.md`](domain/invariants.md).
3. The V2 data model in
   [`architecture/data-model.md`](architecture/data-model.md).
4. Embedded, checksummed SQLite migrations.
5. Generated queries, repositories, application use cases, Wails handlers, and
   frontend features, in that order.

An implemented higher layer must conform to every implemented lower layer. If a
product decision changes, its ADR and invariants change first, followed by a
new migration and then every dependent layer.

## Current documents

- [`architecture/overview.md`](architecture/overview.md): product boundary,
  dependency direction, and runtime architecture.
- [`architecture/data-model.md`](architecture/data-model.md): V2 baseline ERD,
  storage representations, and enforcement boundary.
- [`domain/glossary.md`](domain/glossary.md): canonical vocabulary.
- [`domain/invariants.md`](domain/invariants.md): testable business rules.
- [`domain/inventory-ledger.md`](domain/inventory-ledger.md): posting, costing,
  lot allocation, and replay rules.
- [`domain/use-cases.md`](domain/use-cases.md): supported and deferred workflows.
- [`decisions/`](decisions/README.md): accepted architectural decisions.
- [`development/toolchain.md`](development/toolchain.md): pinned language,
  framework, and tool versions plus setup instructions.
- [`development/database.md`](development/database.md): SQLite identity,
  migrations, local data, backup, and restore policy.
- [`development/testing.md`](development/testing.md): local checks, security
  audits, browser smoke tests, and CI expectations.

Phase 3 replaces the seven experimental migrations with the strict,
checksummed V2 baseline. That migration is now the executable persistence
contract. Legacy models, repositories, services, and pages remain outside the
accepted chain until they are rebuilt against it in bottom-up order; their
continued presence in the source tree does not override the baseline.

## Historical material

Pre-V2 notes are retained under [`archive/pre-v2/`](archive/pre-v2/README.md)
for context only. They are not specifications and must not be used to resolve a
conflict with the documents above.
