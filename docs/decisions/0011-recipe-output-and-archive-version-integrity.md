# ADR 0011: Recipe output and archive-version integrity

- Status: Accepted
- Date: 2026-07-14

## Context

The V2 baseline validated a recipe's output when the recipe was inserted, but
later catalog writes could archive that item or remove its producible
capability while the recipe remained active. The baseline also allowed recipe
revision numbers to be inserted with gaps and allowed an archive timestamp to
be later than the `updated_at` value used as the aggregate's optimistic
version.

Those states cannot be produced by the aggregate stores and make a database
file disagree with the domain contract. Because SQLite is the bottom-up
authority, store-only checks are insufficient.

## Decision

Forward migration `0002_recipe_output_and_archive_versions.sql` establishes
the following structural rules:

- an item referenced as the output of an active recipe cannot be archived or
  lose its producible capability;
- restoring a recipe requires its fixed output item to be active and
  producible;
- each inserted recipe revision is exactly the next number after the current
  maximum, beginning with revision one; and
- for items, item packaging, counterparties, and recipes, a non-null
  `archived_at_ms` equals `updated_at_ms` because archival is itself the
  optimistic-version mutation.

The migration re-evaluates version-one rows under these rules and fails
atomically if an existing active output, revision chain, or archive timestamp
is invalid. It does not guess at an automatic repair. Recipe aggregate and list
queries also validate revision count and minimum number as defense in depth,
before a publish transaction can mutate a corrupt chain.

## Consequences

- Archiving or disabling a recipe output requires archiving every active recipe
  that references it first, in the same serialized workflow when appropriate.
- Recipe revision numbering is contiguous in SQLite as well as in the publish
  transaction.
- An invalid version-one development file remains at migration version one and
  must be repaired explicitly before it can open with the newer application.
- Archive and restore chronology cannot move behind an archive event hidden
  from the optimistic version.
