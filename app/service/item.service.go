package service

import (
	"github.com/google/uuid"
	"github.com/jerobas/saas/repository"
	"github.com/jerobas/saas/model"
)

type ItemService struct {
	repo *repository.ItemRepository
}

func NewItemService(db *Database) *ItemService {
	return &ItemService{repo: repository.NewItemRepository(db)}
}

func (s *ItemService) CreateItem(name, unit string, minStockAlert float64) (*model.ItemDTO, error) {
	item := &model.Item{
		ID:            uuid.New().String(),
		Name:          name,
		Unit:          unit,
		MinStockAlert: minStockAlert,
	}
	if err := s.repo.Create(item); err != nil {
		return nil, err
	}
	
	created, err := s.repo.GetByID(item.ID)
	if err != nil {
		return nil, err
	}
	return model.ToItemDTO(created), nil
}

func (s *ItemService) GetItem(id string) (*model.ItemDTO, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return model.ToItemDTO(item), nil
}

func (s *ItemService) GetAllItems() ([]*model.ItemDTO, error) {
	items, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	return model.ToItemDTOList(items), nil
}

func (s *ItemService) UpdateItem(id, name, unit string, minStockAlert float64) error {
	item := &model.Item{
		ID:            id,
		Name:          name,
		Unit:          unit,
		MinStockAlert: minStockAlert,
	}
	return s.repo.Update(item)
}

func (s *ItemService) DeleteItem(id string) error {
	return s.repo.Delete(id)
}