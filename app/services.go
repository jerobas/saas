package main

import "github.com/google/uuid"

// ============================================
// ITEM SERVICE
// ============================================

type ItemService struct {
	repo *ItemRepository
}

func NewItemService(db *Database) *ItemService {
	return &ItemService{repo: NewItemRepository(db)}
}

func (s *ItemService) CreateItem(name, unit string, minStockAlert float64) (*ItemDTO, error) {
	item := &Item{
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
	return toItemDTO(created), nil
}

func (s *ItemService) GetItem(id string) (*ItemDTO, error) {
	item, err := s.repo.GetByID(id)
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
	return toItemDTOList(items), nil
}

func (s *ItemService) UpdateItem(id, name, unit string, minStockAlert float64) error {
	item := &Item{
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

// ============================================
// BATCH SERVICE
// ============================================

type BatchService struct {
	repo *InventoryBatchRepository
}

func NewBatchService(db *Database) *BatchService {
	return &BatchService{repo: NewInventoryBatchRepository(db)}
}

func (s *BatchService) CreateBatch(itemID string, quantity, totalPrice float64) (*InventoryBatchDTO, error) {
	batch := &InventoryBatch{
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
	return toBatchDTO(created), nil
}

func (s *BatchService) GetBatch(id string) (*InventoryBatchDTO, error) {
	batch, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return toBatchDTO(batch), nil
}

func (s *BatchService) GetBatchesByItem(itemID string) ([]*InventoryBatchDTO, error) {
	batches, err := s.repo.GetByItemID(itemID)
	if err != nil {
		return nil, err
	}
	return toBatchDTOList(batches), nil
}

func (s *BatchService) UpdateBatchQuantity(id string, quantity float64) error {
	return s.repo.UpdateQuantity(id, quantity)
}

func (s *BatchService) DeleteBatch(id string) error {
	return s.repo.Delete(id)
}

// ============================================
// RECIPE SERVICE
// ============================================

type RecipeService struct {
	repo *RecipeRepository
}

func NewRecipeService(db *Database) *RecipeService {
	return &RecipeService{repo: NewRecipeRepository(db)}
}

func (s *RecipeService) CreateRecipe(name string, profitMargin float64) (*RecipeDTO, error) {
	recipe := &Recipe{
		ID:                  uuid.New().String(),
		Name:                name,
		ProfitMarginPercent: profitMargin,
	}
	if err := s.repo.Create(recipe); err != nil {
		return nil, err
	}
	
	created, err := s.repo.GetByID(recipe.ID)
	if err != nil {
		return nil, err
	}
	return toRecipeDTO(created), nil
}

func (s *RecipeService) GetRecipe(id string) (*RecipeDTO, error) {
	recipe, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return toRecipeDTO(recipe), nil
}

func (s *RecipeService) GetAllRecipes() ([]*RecipeDTO, error) {
	recipes, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	return toRecipeDTOList(recipes), nil
}

func (s *RecipeService) UpdateRecipe(id, name string, profitMargin float64) error {
	recipe := &Recipe{
		ID:                  id,
		Name:                name,
		ProfitMarginPercent: profitMargin,
	}
	return s.repo.Update(recipe)
}

func (s *RecipeService) DeleteRecipe(id string) error {
	return s.repo.Delete(id)
}

func (s *RecipeService) AddIngredient(recipeID, itemID string, quantity float64) error {
	ingredient := &RecipeIngredient{
		RecipeID:       recipeID,
		ItemID:         itemID,
		QuantityNeeded: quantity,
	}
	return s.repo.AddIngredient(ingredient)
}

func (s *RecipeService) GetIngredients(recipeID string) ([]*RecipeIngredient, error) {
	return s.repo.GetIngredients(recipeID)
}

func (s *RecipeService) RemoveIngredient(recipeID, itemID string) error {
	return s.repo.RemoveIngredient(recipeID, itemID)
}

// ============================================
// PRODUCT SERVICE
// ============================================

type ProductService struct {
	repo *ProductRepository
}

func NewProductService(db *Database) *ProductService {
	return &ProductService{repo: NewProductRepository(db)}
}

func (s *ProductService) CreateProduct(recipeID string, quantity int, unitCost, salePrice float64) (*ProductDTO, error) {
	product := &Product{
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
	return toProductDTO(created), nil
}

func (s *ProductService) GetProduct(id string) (*ProductDTO, error) {
	product, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return toProductDTO(product), nil
}

func (s *ProductService) GetAllProducts() ([]*ProductDTO, error) {
	products, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	return toProductDTOList(products), nil
}

func (s *ProductService) DeleteProduct(id string) error {
	return s.repo.Delete(id)
}

// ============================================
// SALE SERVICE
// ============================================

type SaleService struct {
	repo *SaleRepository
}

func NewSaleService(db *Database) *SaleService {
	return &SaleService{repo: NewSaleRepository(db)}
}

func (s *SaleService) CreateSale(productID string, quantity int, unitPrice float64) (*SaleDTO, error) {
	sale := &Sale{
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
	return toSaleDTO(created), nil
}

func (s *SaleService) GetSale(id string) (*SaleDTO, error) {
	sale, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return toSaleDTO(sale), nil
}

func (s *SaleService) GetAllSales() ([]*SaleDTO, error) {
	sales, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	return toSaleDTOList(sales), nil
}

func (s *SaleService) DeleteSale(id string) error {
	return s.repo.Delete(id)
}