package repository

import (
	"database/sql"
	"github.com/jerobas/saas/model"
)

type InventoryMovementRepository struct {
	db *Database
}

func NewInventoryMovementRepository(db *Database) *InventoryMovementRepository {
	return &InventoryMovementRepository{db: db}
}

func (r *InventoryMovementRepository) Create(mov *model.InventoryMovementInsertDTO) (int64, error) {
	query := `
		INSERT INTO inventory_movements
			(event_id, item_id, quantity, unit_cost, direction, expires_at, origin_movement_id, occurred_at)
		VALUES
			(?, ?, ?, ?, ?, ?, ?, ?)
	`

	res, err := r.db.Conn.Exec(
		query,
		mov.EventID,
		mov.ItemID,
		mov.Quantity,
		mov.UnitCost,
		mov.Direction,
		mov.ExpiresAt,
		mov.OriginMovementID,
		mov.OccurredAt,
	)

	if err != nil {
		return (-1, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return (-1, err)
	}

	return (id, nil)
}

func (r *InventoryMovementRepository) GetByID(id int64) (*model.InventoryMovement, error) {
	query := `
		SELECT
			id,
			event_id,
			item_id,
			quantity,
			unit_cost,
			direction,
			expires_at,
			origin_movement_id,
			occurred_at,
			created_at
		FROM inventory_movements
		WHERE id = ?
	`

	mov := &model.InventoryMovement{}
	err := r.db.Conn.QueryRow(query, id).Scan(
		&mov.ID,
		&mov.EventID,
		&mov.ItemID,
		&mov.Quantity,
		&mov.UnitCost,
		&mov.Direction,
		&mov.ExpiresAt,
		&mov.OriginMovementID,
		&mov.OccurredAt,
		&mov.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	return mov, nil
}

func (r *InventoryMovementRepository) GetAll() ([]*model.InventoryMovement, error) {
	query := `
		SELECT
			id,
			event_id,
			item_id,
			quantity,
			unit_cost,
			direction,
			expires_at,
			origin_movement_id,
			occurred_at,
			created_at
		FROM inventory_movements
		ORDER BY occurred_at DESC
	`

	rows, err := r.db.Conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	movs := []*model.InventoryMovement{}
	for rows.Next() {
		mov := &model.InventoryMovement{}
		if err := rows.Scan(
			&mov.ID,
			&mov.EventID,
			&mov.ItemID,
			&mov.Quantity,
			&mov.UnitCost,
			&mov.Direction,
			&mov.ExpiresAt,
			&mov.OriginMovementID,
			&mov.OccurredAt,
			&mov.CreatedAt,
		); err != nil {
			return nil, err
		}
		movs = append(movs, mov)
	}

	return movs, rows.Err()
}

func (r *InventoryMovementRepository) GetAllByEventID(eventID int64) ([]*model.InventoryMovement, error) {
	query := `
		SELECT
			id,
			event_id,
			item_id,
			quantity,
			unit_cost,
			direction,
			expires_at,
			origin_movement_id,
			occurred_at,
			created_at
		FROM inventory_movements
		WHERE event_id = ?
		ORDER BY occurred_at DESC
	`

	rows, err := r.db.Conn.Query(query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	movs := []*model.InventoryMovement{}
	for rows.Next() {
		mov := &model.InventoryMovement{}
		if err := rows.Scan(
			&mov.ID,
			&mov.EventID,
			&mov.ItemID,
			&mov.Quantity,
			&mov.UnitCost,
			&mov.Direction,
			&mov.ExpiresAt,
			&mov.OriginMovementID,
			&mov.OccurredAt,
			&mov.CreatedAt,
		); err != nil {
			return nil, err
		}
		movs = append(movs, mov)
	}

	return movs, rows.Err()
}

func (r *InventoryMovementRepository) GetAllByItemID(itemID int64) ([]*model.InventoryMovement, error) {
	query := `
		SELECT
			id,
			event_id,
			item_id,
			direction,
			quantity,
			unit_cost,
			expires_at,
			origin_movement_id,
			occurred_at,
			created_at
		FROM inventory_movements
		WHERE item_id = ?
		ORDER BY occurred_at DESC
	`

	rows, err := r.db.Conn.Query(query, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	movs := []*model.InventoryMovement{}
	for rows.Next() {
		mov := &model.InventoryMovement{}
		if err := rows.Scan(
			&mov.ID,
			&mov.EventID,
			&mov.ItemID,
			&mov.Direction,
			&mov.Quantity,
			&mov.UnitCost,
			&mov.ExpiresAt,
			&mov.OriginMovementID,
			&mov.OccurredAt,
			&mov.CreatedAt,
		); err != nil {
			return nil, err
		}
		movs = append(movs, mov)
	}

	return movs, rows.Err()
}

func (r *InventoryMovementRepository) Delete(id int64) error {
	query := `DELETE FROM inventory_movements WHERE id = ?`
	_, err := r.db.Conn.Exec(query, id)
	return err
}