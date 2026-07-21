// Package sqlite maps the V2 domain to named, generated SQLite queries. Its
// Store exposes aggregate operations and read models while keeping sqlc rows
// and database/sql null values private.
//
// Master-data writes use the database package's serialized BEGIN IMMEDIATE
// coordinator and optimistic updated-at snapshots. Inventory access in this
// package is read-only: ledger posting, valuation, lot allocation, projection
// updates, reversals, and replay belong to the Phase 5 application transaction.
package sqlite
