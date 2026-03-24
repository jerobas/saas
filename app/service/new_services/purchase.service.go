package service

import (
	"github.com/google/uuid"
	"github.com/jerobas/saas/repository"
	"github.com/jerobas/saas/model"
)

type PurchaseService struct {
	evt_repo *repository.NewEventRepository
	pll_repo *repository.NewPurchaseLineRepository
	mov_repo *repository.NewInventoryMovementRepository
}

func NewPurchaseService(db *Database) *ItemService {
	return &ItemService{
		evt_repo: repository.NewEventRepository(db)
		pll_repo: repository.NewPurchaseLineRepository(db)
		mov_repo: repository.NewInventoryMovementRepository(db)
	}
}

func (s *PurchaseService) createPurchaseEvent(param model.CreateEventInput) (int64, error) {
	evt := &model.EventInsertDTO{
		EventType: "PURCHASE",
		Status: "DRAFT",
		CounterpartyEntityID: param.CounterpartyEntityID,
		Notes: param.Notes,
		OccurredAt: param.OccurredAt
	}
	
	return s.evt_repo.Create(evt)
}

func (s *PurchaseService) createPurchaseLines(evt_id int64, param []model.CreatePurchaseWithLinesInput) error {
	for _, row := range param {
		pll := &model.PurchaseLineInsertDTO{
			EventID: evt_id,
			ItemID: row.ItemID,
			Quantity: row.Quantity,
			UnitCost: row.UnitCost,
		}

		_, err := s.pll_repo.Create(pll); if err != nil {
			return err
		}
	}

	return nil
}

func (s *PurchaseService) Buy(evt_param model.CreateEventInput, plls_param []model.CreatePurchaseLinesInput) (int64, error) {	
	evt_id, err := s.createPurchaseEvent(evt_param)
	
	if err != nil {
		return (0, nil)
	}

	err := s.createPurchaseLines(evt_id, plls_param); if err != nil {
		return (0, nil)
	}
	
	err := s.evt_repo.Post(evt_id); if err != nil {
		return (0, err)
	}

	return (evt_id, nil)
}

func (s *PurchaseService) CreateDraftPurchase(evt_param model.CreateEventInput) (int64, error) {
	return createPurchaseEvent(evt_param)
}

func (s* PurchaseService) GetPurchaseByID(evt_id int64) (*model.Event, error) {
	return GetPurchaseByID
}

// func (s *ItemService) CreateItem(name, unit string, minStockAlert float64) (*model.ItemDTO, error) {
// 	item := &model.Item{
// 		ID:            uuid.New().String(),
// 		Name:          name,
// 		Unit:          unit,
// 		MinStockAlert: minStockAlert,
// 	}
// 	if err := s.repo.Create(item); err != nil {
// 		return nil, err
// 	}
	
// 	created, err := s.repo.GetByID(item.ID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return model.ToItemDTO(created), nil
// }

// func (s *ItemService) GetItem(id string) (*model.ItemDTO, error) {
// 	item, err := s.repo.GetByID(id)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return model.ToItemDTO(item), nil
// }

// func (s *ItemService) GetAllItems() ([]*model.ItemDTO, error) {
// 	items, err := s.repo.GetAll()
// 	if err != nil {
// 		return nil, err
// 	}
// 	return model.ToItemDTOList(items), nil
// }

// func (s *ItemService) UpdateItem(id, name, unit string, minStockAlert float64) error {
// 	item := &model.Item{
// 		ID:            id,
// 		Name:          name,
// 		Unit:          unit,
// 		MinStockAlert: minStockAlert,
// 	}
// 	return s.repo.Update(item)
// }

// func (s *ItemService) DeleteItem(id string) error {
// 	return s.repo.Delete(id)
// }