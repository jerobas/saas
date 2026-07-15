// Package domain defines persistence-independent values shared by Sweeters V2
// aggregates. Constructors enforce exact integer quantities, currency
// snapshots, Unicode identity normalization, typed identifiers, controlled
// enums, and explicit optional values. Subpackages own validated aggregate and
// query snapshots for settings, catalog, counterparties, recipes, and
// inventory.
//
// This package deliberately has no JSON, Wails, database/sql, or SQLite
// dependency. Presentation and persistence adapters must translate at their
// respective boundaries.
package domain
