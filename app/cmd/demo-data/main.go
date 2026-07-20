package main

import (
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/jerobas/saas/database"
	"github.com/jerobas/saas/internal/application"
	"github.com/jerobas/saas/internal/infrastructure/sqlite"
	presentationwails "github.com/jerobas/saas/internal/presentation/wails"
	"github.com/jerobas/saas/internal/presentation/wails/dto"
)

const defaultSalesPerMonth = 30

type seedSummary struct {
	Counterparties int
	Items          int
	Packagings     int
	Purchases      int
	Sales          int
}

type demoProduct struct {
	ID         int64
	Name       string
	PriceMinor int64
	CostMinor  int64
}

type demoIngredient struct {
	ID   int64
	Name string
}

type demoSeeder struct {
	catalog      *presentationwails.CatalogHandler
	counterparty *presentationwails.CounterpartyHandler
	purchase     *presentationwails.PurchaseHandler
	sale         *presentationwails.SaleHandler
	random       *rand.Rand
	now          time.Time
	scale        int
	summary      seedSummary
}

func main() {
	databasePath := flag.String("database", "", "path to a new demo app.db")
	scale := flag.Int("scale", defaultSalesPerMonth, "sales documents generated per month (1-200)")
	flag.Parse()

	if err := seedDemoDatabase(*databasePath, *scale, time.Now()); err != nil {
		fmt.Fprintf(os.Stderr, "demo data: %v\n", err)
		os.Exit(1)
	}
}

func seedDemoDatabase(databasePath string, scale int, now time.Time) error {
	if databasePath == "" {
		return errors.New("database path is required")
	}
	if scale < 1 || scale > 200 {
		return fmt.Errorf("scale must be between 1 and 200, got %d", scale)
	}
	absPath, err := filepath.Abs(databasePath)
	if err != nil {
		return fmt.Errorf("resolve database path: %w", err)
	}
	if _, err := os.Stat(absPath); err == nil {
		return fmt.Errorf("database already exists: %s (clean the demo directory first)", absPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect database path: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(absPath), 0o700); err != nil {
		return fmt.Errorf("create database directory: %w", err)
	}

	db, err := database.NewDatabase(absPath)
	if err != nil {
		return fmt.Errorf("open demo database: %w", err)
	}
	defer db.Close()

	store := sqlite.NewStore(db)
	clock := application.SystemClock{}
	seeder := &demoSeeder{
		catalog: presentationwails.NewCatalogHandler(application.NewCatalogService(
			application.NewSQLiteCatalogStore(store), clock,
		)),
		counterparty: presentationwails.NewCounterpartyHandler(application.NewCounterpartyService(
			application.NewSQLiteCounterpartyStore(store), clock,
		)),
		purchase: presentationwails.NewPurchaseHandler(application.NewPurchaseService(
			application.NewSQLitePurchaseStore(store), clock,
		)),
		sale: presentationwails.NewSaleHandler(application.NewSaleService(
			application.NewSQLiteSaleStore(store), clock,
		)),
		random: rand.New(rand.NewSource(42)),
		now:    now,
		scale:  scale,
	}

	summary, err := seeder.seed()
	if err != nil {
		return err
	}
	fmt.Printf(
		"Demo database ready at %s\n  counterparties: %d\n  items: %d\n  packagings: %d\n  purchases: %d\n  sales: %d\n",
		absPath,
		summary.Counterparties,
		summary.Items,
		summary.Packagings,
		summary.Purchases,
		summary.Sales,
	)
	return nil
}

func (s *demoSeeder) seed() (seedSummary, error) {
	suppliers, customers, err := s.seedCounterparties()
	if err != nil {
		return seedSummary{}, err
	}
	ingredients, products, err := s.seedCatalog()
	if err != nil {
		return seedSummary{}, err
	}
	if err := s.seedDocuments(suppliers, customers, ingredients, products); err != nil {
		return seedSummary{}, err
	}
	return s.summary, nil
}

func (s *demoSeeder) seedCounterparties() ([]int64, []int64, error) {
	supplierNames := []string{
		"Demo Atacado Doce", "Demo Distribuidora Central", "Demo Fazenda Leiteira", "Demo Embalagens Sul",
	}
	suppliers := make([]int64, 0, len(supplierNames))
	for index, name := range supplierNames {
		email := fmt.Sprintf("fornecedor%02d@demo.local", index+1)
		value, err := s.counterparty.CreateCounterparty(dto.CounterpartyWriteRequest{
			Name: name, Email: &email, Roles: []string{"SUPPLIER"},
		})
		if err != nil {
			return nil, nil, fmt.Errorf("create supplier %q: %w", name, err)
		}
		suppliers = append(suppliers, value.ID)
		s.summary.Counterparties++
	}

	customerCount := max(12, s.scale)
	customers := make([]int64, 0, customerCount)
	for index := 0; index < customerCount; index++ {
		name := fmt.Sprintf("Cliente Demo %03d", index+1)
		email := fmt.Sprintf("cliente%03d@demo.local", index+1)
		value, err := s.counterparty.CreateCounterparty(dto.CounterpartyWriteRequest{
			Name: name, Email: &email, Roles: []string{"CUSTOMER"},
		})
		if err != nil {
			return nil, nil, fmt.Errorf("create customer %q: %w", name, err)
		}
		customers = append(customers, value.ID)
		s.summary.Counterparties++
	}
	return suppliers, customers, nil
}

func (s *demoSeeder) seedCatalog() ([]demoIngredient, []demoProduct, error) {
	ingredientNames := []string{
		"Chocolate em pó", "Farinha de trigo", "Açúcar refinado", "Leite condensado",
		"Creme de leite", "Manteiga", "Morango", "Coco ralado",
	}
	ingredients := make([]demoIngredient, 0, len(ingredientNames))
	for index, name := range ingredientNames {
		sku := fmt.Sprintf("DEMO-I%02d", index+1)
		description := "Ingrediente criado pelo gerador de dados demonstrativos."
		reorder := int64(1_000_000)
		item, err := s.catalog.CreateItem(dto.ItemWriteRequest{
			Name: name, SKU: &sku, Description: &description, BaseUnitCode: "g",
			Capabilities: dto.CapabilitiesRequest{Purchasable: true}, ReorderQuantity: &reorder,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("create ingredient %q: %w", name, err)
		}
		if _, err := s.catalog.CreateItemPackaging(dto.PackagingCreateRequest{
			ItemID: item.ID,
			PackagingWriteRequest: dto.PackagingWriteRequest{
				Name: "Pacote de 1 kg", EnteredUnitCode: "kg", ConversionNumerator: 1_000_000, ConversionDenominator: 1,
			},
		}); err != nil {
			return nil, nil, fmt.Errorf("create ingredient packaging %q: %w", name, err)
		}
		ingredients = append(ingredients, demoIngredient{ID: item.ID, Name: name})
		s.summary.Items++
		s.summary.Packagings++
	}

	productNames := []string{
		"Bolo de chocolate", "Bolo de cenoura", "Brigadeiro gourmet", "Beijinho de coco",
		"Torta de morango", "Torta de limão", "Brownie", "Cookie recheado",
		"Mousse de chocolate", "Pudim", "Cupcake de baunilha", "Caixa de docinhos",
	}
	products := make([]demoProduct, 0, len(productNames))
	for index, name := range productNames {
		sku := fmt.Sprintf("DEMO-P%02d", index+1)
		description := "Produto criado pelo gerador de dados demonstrativos."
		price := int64(450 + index*175)
		cost := price * 42 / 100
		reorder := int64(100)
		item, err := s.catalog.CreateItem(dto.ItemWriteRequest{
			Name: name, SKU: &sku, Description: &description, BaseUnitCode: "each",
			Capabilities:     dto.CapabilitiesRequest{Purchasable: true, Sellable: true},
			DefaultSalePrice: &price, ReorderQuantity: &reorder,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("create product %q: %w", name, err)
		}
		if _, err := s.catalog.CreateItemPackaging(dto.PackagingCreateRequest{
			ItemID: item.ID,
			PackagingWriteRequest: dto.PackagingWriteRequest{
				Name: "Caixa com 6", EnteredUnitCode: "each", ConversionNumerator: 6, ConversionDenominator: 1,
			},
		}); err != nil {
			return nil, nil, fmt.Errorf("create product packaging %q: %w", name, err)
		}
		products = append(products, demoProduct{ID: item.ID, Name: name, PriceMinor: price, CostMinor: cost})
		s.summary.Items++
		s.summary.Packagings++
	}
	return ingredients, products, nil
}

func (s *demoSeeder) seedDocuments(
	suppliers []int64,
	customers []int64,
	ingredients []demoIngredient,
	products []demoProduct,
) error {
	currentMonth := time.Date(s.now.Year(), s.now.Month(), 1, 0, 0, 0, 0, s.now.Location())
	firstMonth := currentMonth.AddDate(0, -5, 0)
	for monthIndex := 0; monthIndex < 6; monthIndex++ {
		month := firstMonth.AddDate(0, monthIndex, 0)
		if err := s.postMonthlyPurchase(monthIndex, month, suppliers, ingredients, products); err != nil {
			return err
		}
		if err := s.postMonthlySales(monthIndex, month, currentMonth, customers, products); err != nil {
			return err
		}
	}
	return nil
}

func (s *demoSeeder) postMonthlyPurchase(
	monthIndex int,
	month time.Time,
	suppliers []int64,
	ingredients []demoIngredient,
	products []demoProduct,
) error {
	lines := make([]dto.PurchaseLineRequest, 0, len(ingredients)+len(products))
	expiresIngredients := month.AddDate(0, 10, 0).Format("2006-01-02")
	for index, ingredient := range ingredients {
		lotCode := fmt.Sprintf("DEMO-%s-I%02d", month.Format("200601"), index+1)
		lines = append(lines, dto.PurchaseLineRequest{
			ItemID: ingredient.ID, QuantityAtomic: 5_000_000, EnteredUnitCode: "kg",
			EnteredPackagingName: stringPointer("Pacote de 1 kg"), ConversionNumeratorAtomic: 1_000_000,
			ConversionDenominator: 1, CommercialTotalMinor: int64(18_000 + index*2_500),
			LotCode: &lotCode, ExpiresOn: &expiresIngredients,
		})
	}
	expiresProducts := month.AddDate(0, 4, 0).Format("2006-01-02")
	for index, product := range products {
		lotCode := fmt.Sprintf("DEMO-%s-P%02d", month.Format("200601"), index+1)
		quantity := int64(2_000)
		lines = append(lines, dto.PurchaseLineRequest{
			ItemID: product.ID, QuantityAtomic: quantity, EnteredUnitCode: "each",
			ConversionNumeratorAtomic: 1, ConversionDenominator: 1,
			CommercialTotalMinor: quantity * product.CostMinor, LotCode: &lotCode, ExpiresOn: &expiresProducts,
		})
	}
	notes := "Reposição mensal criada pelo gerador de dados demonstrativos."
	_, err := s.purchase.PostPurchase(dto.PurchasePostRequest{
		IdempotencyKey: fmt.Sprintf("demo-purchase-%s", month.Format("2006-01")),
		CounterpartyID: int64Pointer(suppliers[monthIndex%len(suppliers)]),
		OccurredOn:     month.Format("2006-01-02"), Notes: &notes, Lines: lines,
	})
	if err != nil {
		return fmt.Errorf("post purchase for %s: %w", month.Format("2006-01"), err)
	}
	s.summary.Purchases++
	return nil
}

func (s *demoSeeder) postMonthlySales(
	monthIndex int,
	month time.Time,
	currentMonth time.Time,
	customers []int64,
	products []demoProduct,
) error {
	maxDay := daysInMonth(month)
	if month.Equal(currentMonth) && s.now.Day() < maxDay {
		maxDay = s.now.Day()
	}
	for saleIndex := 0; saleIndex < s.scale; saleIndex++ {
		lineCount := 1 + s.random.Intn(3)
		permutation := s.random.Perm(len(products))
		isSample := saleIndex == 0
		lines := make([]dto.SaleLineRequest, 0, lineCount)
		for lineIndex := 0; lineIndex < lineCount; lineIndex++ {
			product := products[permutation[lineIndex]]
			quantity := int64(1 + s.random.Intn(6))
			price := product.PriceMinor * int64(100+monthIndex*4) / 100
			total := quantity * price
			if isSample {
				total = 0
			}
			lines = append(lines, dto.SaleLineRequest{
				ItemID: product.ID, QuantityAtomic: quantity, EnteredUnitCode: "each",
				ConversionNumeratorAtomic: 1, ConversionDenominator: 1, CommercialTotalMinor: total,
			})
		}
		occurredOn := time.Date(month.Year(), month.Month(), 1+saleIndex%maxDay, 0, 0, 0, 0, month.Location())
		request := dto.SalePostRequest{
			IdempotencyKey: fmt.Sprintf("demo-sale-%s-%03d", month.Format("2006-01"), saleIndex+1),
			OccurredOn:     occurredOn.Format("2006-01-02"), Lines: lines,
		}
		if saleIndex%7 != 0 {
			request.CounterpartyID = int64Pointer(customers[(monthIndex*s.scale+saleIndex)%len(customers)])
		}
		if isSample {
			request.ReasonCode = stringPointer("SAMPLE")
		}
		if _, err := s.sale.PostSale(request); err != nil {
			return fmt.Errorf("post sale %d for %s: %w", saleIndex+1, month.Format("2006-01"), err)
		}
		s.summary.Sales++
	}
	return nil
}

func daysInMonth(value time.Time) int {
	return time.Date(value.Year(), value.Month()+1, 0, 0, 0, 0, 0, value.Location()).Day()
}

func int64Pointer(value int64) *int64 { return &value }

func stringPointer(value string) *string { return &value }
