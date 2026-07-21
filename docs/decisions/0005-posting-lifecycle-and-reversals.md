# ADR 0005: Posting lifecycle, idempotency, and reversals

- Status: Accepted
- Date: 2026-07-13

## Context

The current event model persists drafts and a `CANCELLED` state, but the active
UI does not provide a meaningful durable-draft workflow. Posted weighted-average
history also cannot always be cancelled later without changing the valuation of
subsequent documents.

## Decision

V2 initially persists only complete posted documents. Form drafts remain local
frontend state and closing the application may discard them. Durable drafts can
be introduced later as a separate workflow.

Posting is one serialized transaction. It writes all rows and projections or no
rows. Each command includes a unique client-generated idempotency key; a retry
returns the document already associated with that key.

Posted documents and descendants are immutable. There is no status mutation to
`CANCELLED`.

An exact full reversal is a new linked immutable document and is permitted only
when:

- the target is the latest stock-affecting document for every affected item;
- target-created lots have no downstream consumption that cannot be restored;
- all inverse line, value, and lot effects can recreate the immediately prior
  state exactly;
- the target has not already been reversed and is not itself a reversal.

Reversal posts at the current sequence. When those requirements fail, the user
must record a current-period compensating adjustment or a future dedicated
return workflow. History is never rewritten.

The document's business date is for reporting. Posting sequence is always the
availability and valuation order, including for backdated entries.

## Consequences

- The database cannot contain abandoned partial events.
- Retried Wails commands cannot double-post stock.
- “Undo” is honest about weighted-average and lot dependencies.
- Data-entry correction and a physical customer/supplier return remain distinct
  product concepts.
- Autosave, partial returns, and reversing nonlatest history are explicitly
  deferred rather than approximated.
