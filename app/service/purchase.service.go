package service

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jerobas/saas/model"
	"github.com/jerobas/saas/repository"
)

type PurchaseLineInput struct {
	ItemID   int64   `json:"item_id"`
	Quantity float64 `json:"quantity"`
	UnitCost int64   `json:"unit_cost"`
}

type CreatePurchaseInput struct {
	CounterpartyEntityID *int64              `json:"counterparty_entity_id"`
	Notes                *string             `json:"notes"`
	OccurredAt           time.Time           `json:"occurred_at"`
	Lines                []PurchaseLineInput `json:"lines"`
}

type PurchaseService struct {
	db *Database
}

func NewPurchaseService(db *Database) *PurchaseService {
	return &PurchaseService{db: db}
}

// Buy atomically creates and posts a purchase. Amounts are expressed in the
// database's minor currency unit (cents for BRL).
func (s *PurchaseService) Buy(input CreatePurchaseInput) (int64, error) {
	return s.create(input, true)
}

func (s *PurchaseService) CreateDraftPurchase(input CreatePurchaseInput) (int64, error) {
	return s.create(input, false)
}

func (s *PurchaseService) GetPurchaseByID(eventID int64) (*model.GetPurchaseByIDOutput, error) {
	eventRepo := repository.NewEventRepository(s.db.Conn)
	lineRepo := repository.NewPurchaseLineRepository(s.db.Conn)
	event, err := eventRepo.GetByID(eventID)
	if err != nil {
		return nil, err
	}
	if event.EventType != "PURCHASE" {
		return nil, errors.New("event is not a purchase")
	}
	lines, err := lineRepo.GetAllByEventID(eventID)
	if err != nil {
		return nil, err
	}
	result := &model.GetPurchaseByIDOutput{
		ID:                   event.ID,
		EventType:            event.EventType,
		Status:               event.Status,
		CounterpartyEntityID: event.CounterpartyEntityID,
		Notes:                event.Notes,
		OccurredAt:           event.OccurredAt,
		CreatedAt:            event.CreatedAt,
		PurchaseLines:        make([]model.PurchaseLine, 0, len(lines)),
	}
	for _, line := range lines {
		result.PurchaseLines = append(result.PurchaseLines, *line)
	}
	return result, nil
}

func (s *PurchaseService) create(input CreatePurchaseInput, post bool) (eventID int64, err error) {
	if len(input.Lines) == 0 {
		return 0, errors.New("purchase requires at least one line")
	}
	if input.OccurredAt.IsZero() {
		input.OccurredAt = time.Now()
	}
	tx, err := s.db.Conn.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	eventRepo := repository.NewEventRepository(tx)
	lineRepo := repository.NewPurchaseLineRepository(tx)
	movementRepo := repository.NewInventoryMovementRepository(tx)
	eventID, err = eventRepo.Create(&model.EventInsertDTO{
		EventType:            "PURCHASE",
		Status:               "DRAFT",
		CounterpartyEntityID: input.CounterpartyEntityID,
		Notes:                input.Notes,
		OccurredAt:           input.OccurredAt,
	})
	if err != nil {
		return 0, err
	}

	for _, line := range input.Lines {
		if line.Quantity <= 0 || line.UnitCost < 0 {
			return 0, errors.New("purchase line has invalid quantity or unit cost")
		}
		_, err = lineRepo.Create(&model.PurchaseLineInsertDTO{
			EventID:  eventID,
			ItemID:   line.ItemID,
			Quantity: line.Quantity,
			UnitCost: line.UnitCost,
		})
		if err != nil {
			return 0, err
		}
		_, err = movementRepo.Create(&model.InventoryMovementInsertDTO{
			EventID:    eventID,
			ItemID:     line.ItemID,
			Direction:  "IN",
			Quantity:   line.Quantity,
			UnitCost:   sql.NullInt64{Int64: line.UnitCost, Valid: true},
			OccurredAt: input.OccurredAt,
		})
		if err != nil {
			return 0, err
		}
	}

	if post {
		if err = eventRepo.Post(eventID); err != nil {
			return 0, err
		}
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return eventID, nil
}
