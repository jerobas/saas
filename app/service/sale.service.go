package service

import (
	"github.com/google/uuid"
	"github.com/jerobas/saas/repository"
	"github.com/jerobas/saas/database"
	"github.com/jerobas/saas/model"
)

type Database = database.Database

type SaleService struct {
	repo *repository.SaleRepository
}

func NewSaleService(db *Database) *SaleService {
	return &SaleService{repo: repository.NewSaleRepository(db)}
}

func (s *SaleService) CreateSale(productID string, quantity int, unitPrice float64) (*model.SaleDTO, error) {
	sale := &model.Sale{
		ID:           uuid.New().String(),
		ProductID:    productID,
		QuantitySold: quantity,
		UnitPrice:    unitPrice,
		TotalPrice:   float64(quantity) * unitPrice,
	}
	if err := s.repo.Create(sale); err != nil {
		return nil, err
	}
	
	created, err := s.repo.GetByID(sale.ID)
	if err != nil {
		return nil, err
	}
	return model.ToSaleDTO(created), nil
}

func (s *SaleService) GetSale(id string) (*model.SaleDTO, error) {
	sale, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return model.ToSaleDTO(sale), nil
}

func (s *SaleService) GetAllSales() ([]*model.SaleDTO, error) {
	sales, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	return model.ToSaleDTOList(sales), nil
}

func (s *SaleService) DeleteSale(id string) error {
	return s.repo.Delete(id)
}