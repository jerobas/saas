package service

import (
	"github.com/google/uuid"
	"github.com/jerobas/saas/repository"
	"github.com/jerobas/saas/model"
)

type ProductService struct {
	repo *repository.ProductRepository
}

func NewProductService(db *Database) *ProductService {
	return &ProductService{repo: repository.NewProductRepository(db)}
}

func (s *ProductService) CreateProduct(recipeID string, quantity int, unitCost, salePrice float64) (*model.ProductDTO, error) {
	product := &model.Product{
		ID:               uuid.New().String(),
		RecipeID:         recipeID,
		QuantityProduced: quantity,
		UnitCost:         unitCost,
		SalePrice:        salePrice,
	}
	if err := s.repo.Create(product); err != nil {
		return nil, err
	}
	
	created, err := s.repo.GetByID(product.ID)
	if err != nil {
		return nil, err
	}
	return model.ToProductDTO(created), nil
}

func (s *ProductService) GetProduct(id string) (*model.ProductDTO, error) {
	product, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return model.ToProductDTO(product), nil
}

func (s *ProductService) GetAllProducts() ([]*model.ProductDTO, error) {
	products, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	return model.ToProductDTOList(products), nil
}

func (s *ProductService) DeleteProduct(id string) error {
	return s.repo.Delete(id)
}