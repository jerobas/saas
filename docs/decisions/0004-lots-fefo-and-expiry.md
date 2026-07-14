# ADR 0004: Lots, FEFO, and expiry

- Status: Accepted
- Date: 2026-07-13

## Context

The experimental `origin_movement_id` lets an outbound row reference only one
inbound row and does not prevent overconsumption. The batch compatibility
service reports original quantity as remaining quantity. Expiry is attached to
movements rather than traceable stock.

## Decision

Lots are first-class inventory records.

Every normal inbound line creates exactly one same-item lot. A user splits an
inbound item into multiple document lines when lot code or expiry differs.
Outbound lines use immutable many-to-many lot allocations whose quantities sum
exactly to the line quantity.

Default physical allocation is FEFO:

1. earliest non-null expiry;
2. lots without expiry after dated lots;
3. inbound posting sequence;
4. lot ID.

A user may explicitly select a different nonexpired lot. The selection is
frozen in the posted allocation.

Expiry is an inclusive ISO business date. An expired lot remains in physical
quantity and inventory value until a reasoned waste/expiry adjustment removes
it, but it cannot be allocated to a sale or production run. Eligibility uses
the current business date at posting, not a backdated document date.

Lot allocation and financial valuation are independent: FEFO tracks physical
stock while pooled weighted average calculates financial outflow.

An exact reversal of outbound stock restores its original allocations through
new linked restoration entries rather than editing consumption history.

## Consequences

- One sale or production input can consume multiple lots safely.
- Remaining lot quantity is derived and independently replayable.
- Expired inventory stays visible instead of disappearing automatically.
- Physical traceability does not force lot-specific financial costing.
- Dedicated customer/supplier return safety rules are still required later;
  exact reversal is not a general return workflow.
