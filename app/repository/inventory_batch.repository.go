package repository

import (
	"database/sql"
	"github.com/jerobas/saas/model"
)

type InventoryBatchRepository struct {
	db *Database
}

func NewInventoryBatchRepository(db *Database) *InventoryBatchRepository {
	return &InventoryBatchRepository{db: db}
}

func (r *InventoryBatchRepository) Create(batch *model.InventoryBatch) error {
	query := `
		INSERT INTO inventory_batches 
		(id, item_id, quantity_total, quantity_remaining, purchase_price_total, unit_price)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.Conn.Exec(query, batch.ID, batch.ItemID, batch.QuantityTotal,
		batch.QuantityRemaining, batch.PurchasePriceTotal, batch.UnitPrice)
	return err
}

func (r *InventoryBatchRepository) GetByID(id string) (*model.InventoryBatch, error) {
	query := `
		SELECT id, item_id, quantity_total, quantity_remaining, 
		       purchase_price_total, unit_price, purchased_at
		FROM inventory_batches WHERE id = ?
	`
	batch := &model.InventoryBatch{}
	err := r.db.Conn.QueryRow(query, id).Scan(
		&batch.ID, &batch.ItemID, &batch.QuantityTotal, &batch.QuantityRemaining,
		&batch.PurchasePriceTotal, &batch.UnitPrice, &batch.PurchasedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return batch, err
}

func (r *InventoryBatchRepository) GetByItemID(itemID string) ([]*model.InventoryBatch, error) {
	query := `
		SELECT id, item_id, quantity_total, quantity_remaining,
		       purchase_price_total, unit_price, purchased_at
		FROM inventory_batches 
		WHERE item_id = ? AND quantity_remaining > 0
		ORDER BY purchased_at
	`
	rows, err := r.db.Conn.Query(query, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	batches := []*model.InventoryBatch{}
	for rows.Next() {
		batch := &model.InventoryBatch{}
		if err := rows.Scan(&batch.ID, &batch.ItemID, &batch.QuantityTotal,
			&batch.QuantityRemaining, &batch.PurchasePriceTotal, &batch.UnitPrice,
			&batch.PurchasedAt); err != nil {
			return nil, err
		}
		batches = append(batches, batch)
	}
	return batches, rows.Err()
}

func (r *InventoryBatchRepository) UpdateQuantity(id string, quantity float64) error {
	_, err := r.db.Conn.Exec(`UPDATE inventory_batches SET quantity_remaining = ? WHERE id = ?`, quantity, id)
	return err
}

func (r *InventoryBatchRepository) Delete(id string) error {
	_, err := r.db.Conn.Exec(`DELETE FROM inventory_batches WHERE id = ?`, id)
	return err
}