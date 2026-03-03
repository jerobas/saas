```go
package dto

type SaleLineUpdateDTO struct {
	// Updatable fields (nil = no change)
	EventID   *int64   `json:"event_id"`
	ItemID    *int64   `json:"item_id"`
	Quantity  *float64 `json:"quantity"`
	UnitPrice *int64   `json:"unit_price"`
}
```

```go
package dto

import "time"

type InventoryMovementUpdateDTO struct {
	EventID    *int64   `json:"event_id"`
	ItemID     *int64   `json:"item_id"`
	Direction  *string  `json:"direction"`   // "IN" | "OUT"
	Quantity   *float64 `json:"quantity"`    // > 0
	UnitCost   *int64   `json:"unit_cost"`   // nullable in DB, but pointer-style can only SET a value (see note below)
	ExpiresAt  *time.Time `json:"expires_at"` // nullable in DB, same note
	OriginID   *int64   `json:"origin_movement_id"` // nullable in DB, same note
	OccurredAt *time.Time `json:"occurred_at"`
}
```

```go
package repository

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jerobas/saas/dto"
)

type SaleLineRepository struct {
	db *Database
}

func NewSaleLineRepository(db *Database) *SaleLineRepository {
	return &SaleLineRepository{db: db}
}

func (r *SaleLineRepository) UpdateByID(id int64, upd dto.SaleLineUpdateDTO) error {
	sets := []string{}
	args := []any{}

	if upd.EventID != nil {
		sets = append(sets, "event_id = ?")
		args = append(args, *upd.EventID)
	}
	if upd.ItemID != nil {
		sets = append(sets, "item_id = ?")
		args = append(args, *upd.ItemID)
	}
	if upd.Quantity != nil {
		sets = append(sets, "quantity = ?")
		args = append(args, *upd.Quantity)
	}
	if upd.UnitPrice != nil {
		sets = append(sets, "unit_price = ?")
		args = append(args, *upd.UnitPrice)
	}

	if len(sets) == 0 {
		return errors.New("no fields provided to update")
	}

	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE sale_lines
		SET %s
		WHERE id = ?
	`, strings.Join(sets, ", "))

	_, err := r.db.Conn.Exec(query, args...)
	return err
}
```

```go
package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jerobas/saas/dto"
)

type InventoryMovementRepository struct {
	db *Database
}

func NewInventoryMovementRepository(db *Database) *InventoryMovementRepository {
	return &InventoryMovementRepository{db: db}
}

func (r *InventoryMovementRepository) UpdateByID(id int64, upd dto.InventoryMovementUpdateDTO) error {
	sets := []string{}
	args := []any{}

	if upd.EventID != nil {
		sets = append(sets, "event_id = ?")
		args = append(args, *upd.EventID)
	}
	if upd.ItemID != nil {
		sets = append(sets, "item_id = ?")
		args = append(args, *upd.ItemID)
	}
	if upd.Direction != nil {
		sets = append(sets, "direction = ?")
		args = append(args, *upd.Direction)
	}
	if upd.Quantity != nil {
		sets = append(sets, "quantity = ?")
		args = append(args, *upd.Quantity)
	}

	// Nullable DB columns (pointer-style => only “set to value”)
	if upd.UnitCost != nil {
		sets = append(sets, "unit_cost = ?")
		args = append(args, sql.NullInt64{Int64: *upd.UnitCost, Valid: true})
	}
	if upd.ExpiresAt != nil {
		sets = append(sets, "expires_at = ?")
		args = append(args, sql.NullTime{Time: *upd.ExpiresAt, Valid: true})
	}
	if upd.OriginID != nil {
		sets = append(sets, "origin_movement_id = ?")
		args = append(args, sql.NullInt64{Int64: *upd.OriginID, Valid: true})
	}

	if upd.OccurredAt != nil {
		sets = append(sets, "occurred_at = ?")
		args = append(args, *upd.OccurredAt)
	}

	if len(sets) == 0 {
		return errors.New("no fields provided to update")
	}

	// optional: bump updated_at if you have it (inventory_movements does not)
	// sets = append(sets, "updated_at = CURRENT_TIMESTAMP")

	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE inventory_movements
		SET %s
		WHERE id = ?
	`, strings.Join(sets, ", "))

	_, err := r.db.Conn.Exec(query, args...)
	_ = time.Time{} // remove if you don’t need time import elsewhere
	return err
}
```