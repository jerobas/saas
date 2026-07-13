package service

import (
	"errors"
	"math"
	"strconv"
)

// BatchService is a temporary adapter from the existing UI's batch vocabulary
// to PURCHASE events and IN inventory movements.
type BatchService struct {
	db       *Database
	purchase *PurchaseService
}

func NewBatchService(db *Database) *BatchService {
	return &BatchService{db: db, purchase: NewPurchaseService(db)}
}

func (s *BatchService) CreateBatch(itemID string, quantity, totalPrice float64) (*InventoryBatchDTO, error) {
	id, err := strconv.ParseInt(itemID, 10, 64)
	if err != nil {
		return nil, err
	}
	if quantity <= 0 {
		return nil, errors.New("quantity must be greater than zero")
	}
	unitCost := int64(math.Round(totalPrice * 100 / quantity))
	purchaseID, err := s.purchase.Buy(CreatePurchaseInput{
		Lines: []PurchaseLineInput{{ItemID: id, Quantity: quantity, UnitCost: unitCost}},
	})
	if err != nil {
		return nil, err
	}
	rows, err := s.getBatches(itemID, purchaseID)
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	return rows[0], nil
}

func (s *BatchService) GetBatch(id string) (*InventoryBatchDTO, error) {
	query := `SELECT m.id, m.item_id, m.quantity, m.unit_cost, m.occurred_at
		FROM inventory_movements m JOIN events e ON e.id = m.event_id
		WHERE m.id = ? AND e.event_type = 'PURCHASE' AND m.direction = 'IN'`
	var movementID, itemID, unitCost int64
	var quantity float64
	var occurredAt string
	if err := s.db.Conn.QueryRow(query, id).Scan(&movementID, &itemID, &quantity, &unitCost, &occurredAt); err != nil {
		return nil, err
	}
	return newBatchDTO(movementID, itemID, quantity, unitCost, occurredAt), nil
}

func (s *BatchService) GetBatchesByItem(itemID string) ([]*InventoryBatchDTO, error) {
	return s.getBatches(itemID, 0)
}

func (s *BatchService) DeleteBatch(_ string) error {
	return errors.New("posted purchase movements are immutable; use a reversal event instead")
}

func (s *BatchService) getBatches(itemID string, eventID int64) ([]*InventoryBatchDTO, error) {
	query := `SELECT m.id, m.item_id, m.quantity, m.unit_cost, m.occurred_at
		FROM inventory_movements m JOIN events e ON e.id = m.event_id
		WHERE m.item_id = ? AND e.event_type = 'PURCHASE' AND m.direction = 'IN'`
	args := []any{itemID}
	if eventID != 0 {
		query += " AND e.id = ?"
		args = append(args, eventID)
	}
	query += " ORDER BY m.occurred_at DESC"
	rows, err := s.db.Conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []*InventoryBatchDTO{}
	for rows.Next() {
		var movementID, parsedItemID, unitCost int64
		var quantity float64
		var occurredAt string
		if err := rows.Scan(&movementID, &parsedItemID, &quantity, &unitCost, &occurredAt); err != nil {
			return nil, err
		}
		result = append(result, newBatchDTO(movementID, parsedItemID, quantity, unitCost, occurredAt))
	}
	return result, rows.Err()
}

func newBatchDTO(id, itemID int64, quantity float64, unitCost int64, occurredAt string) *InventoryBatchDTO {
	unitPrice := float64(unitCost) / 100
	return &InventoryBatchDTO{
		ID:                 strconv.FormatInt(id, 10),
		ItemID:             strconv.FormatInt(itemID, 10),
		QuantityTotal:      quantity,
		QuantityRemaining:  quantity,
		PurchasePriceTotal: unitPrice * quantity,
		UnitPrice:          unitPrice,
		PurchasedAt:        occurredAt,
	}
}
