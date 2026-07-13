package service

import (
	"database/sql"
	"errors"
	"strconv"

	"github.com/jerobas/saas/model"
	"github.com/jerobas/saas/repository"
)

type ItemService struct {
	repo *repository.ItemRepository
}

func NewItemService(db *Database) *ItemService {
	return &ItemService{repo: repository.NewItemRepository(db.Conn)}
}

// CreateItem is the compatibility entry point used by the current ingredient
// UI. New domain code should prefer CreateCatalogItem.
func (s *ItemService) CreateItem(name, unit string, _ float64) (*ItemDTO, error) {
	id, err := s.repo.Create(&model.ItemInsertDTO{
		Name:             name,
		Unit:             unit,
		Purchasable:      1,
		DefaultSalePrice: sql.NullInt64{},
	})
	if err != nil {
		return nil, err
	}
	return s.GetItem(strconv.FormatInt(id, 10))
}

func (s *ItemService) CreateCatalogItem(input model.ItemInsertDTO) (*model.Item, error) {
	id, err := s.repo.Create(&input)
	if err != nil {
		return nil, err
	}
	return s.repo.GetByID(id)
}

func (s *ItemService) GetItem(id string) (*ItemDTO, error) {
	itemID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, err
	}
	item, err := s.repo.GetByID(itemID)
	if err != nil {
		return nil, err
	}
	return toItemDTO(item), nil
}

func (s *ItemService) GetAllItems() ([]*ItemDTO, error) {
	items, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	result := make([]*ItemDTO, 0, len(items))
	for _, item := range items {
		result = append(result, toItemDTO(item))
	}
	return result, nil
}

func (s *ItemService) GetCatalogItems() ([]*model.Item, error) {
	return s.repo.GetAll()
}

func (s *ItemService) UpdateItem(_, _, _ string, _ float64) error {
	return errors.New("item updates are not implemented for the event-ledger model yet")
}

func (s *ItemService) DeleteItem(_ string) error {
	return errors.New("items cannot be deleted; soft-delete support is the next domain step")
}

func toItemDTO(item *model.Item) *ItemDTO {
	return &ItemDTO{
		ID:        strconv.FormatInt(item.ID, 10),
		Name:      item.Name,
		Unit:      item.Unit,
		CreatedAt: item.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
