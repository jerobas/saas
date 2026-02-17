package service

import (
	"github.com/google/uuid"
	"github.com/jerobas/saas/repository"
	"github.com/jerobas/saas/model"
)

type BatchService struct {
	repo *repository.InventoryBatchRepository
}

func NewBatchService(db *Database) *BatchService {
	return &BatchService{repo: repository.NewInventoryBatchRepository(db)}
}

func (s *BatchService) CreateBatch(itemID string, quantity, totalPrice float64) (*model.InventoryBatchDTO, error) {
	batch := &model.InventoryBatch{
		ID:                 uuid.New().String(),
		ItemID:             itemID,
		QuantityTotal:      quantity,
		QuantityRemaining:  quantity,
		PurchasePriceTotal: totalPrice,
		UnitPrice:          totalPrice / quantity,
	}
	if err := s.repo.Create(batch); err != nil {
		return nil, err
	}
	
	created, err := s.repo.GetByID(batch.ID)
	if err != nil {
		return nil, err
	}
	return model.ToBatchDTO(created), nil
}

func (s *BatchService) GetBatch(id string) (*model.InventoryBatchDTO, error) {
	batch, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return model.ToBatchDTO(batch), nil
}

func (s *BatchService) GetBatchesByItem(itemID string) ([]*model.InventoryBatchDTO, error) {
	batches, err := s.repo.GetByItemID(itemID)
	if err != nil {
		return nil, err
	}
	return model.ToBatchDTOList(batches), nil
}

func (s *BatchService) UpdateBatchQuantity(id string, quantity float64) error {
	return s.repo.UpdateQuantity(id, quantity)
}

func (s *BatchService) DeleteBatch(id string) error {
	return s.repo.Delete(id)
}